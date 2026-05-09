package agent

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/woopsfactory/warvis/internal/hunt"
	"github.com/woopsfactory/warvis/internal/mcp"
	"github.com/woopsfactory/warvis/pkg/ollama"
)

// Loop represents the agent loop that orchestrates Gemma, MCP tools, and FSM transitions.
type Loop struct {
	fsm              hunt.FSM
	mcpClient        *mcp.Client
	ollamaClient     *ollama.Client
	auditLog         *hunt.AuditLog
	conversationHist []ConversationTurn
	maxHistorySize   int
	maxToolOutputLen int
	stateTimeout     time.Duration
}

// NewLoop creates a new agent loop.
func NewLoop(fsm hunt.FSM, mcpClient *mcp.Client, ollamaClient *ollama.Client, auditLog *hunt.AuditLog) *Loop {
	return &Loop{
		fsm:              fsm,
		mcpClient:        mcpClient,
		ollamaClient:     ollamaClient,
		auditLog:         auditLog,
		conversationHist: []ConversationTurn{},
		maxHistorySize:   20,                // Max 20 exchanges
		maxToolOutputLen: 500,               // Truncate tool outputs to 500 chars
		stateTimeout:     600 * time.Second, // 10 minute timeout per state
	}
}

// Run executes the agent loop for the current FSM state.
// The loop continues until the agent requests state_complete or escalate.
func (l *Loop) Run(ctx context.Context) error {
	stateCtx, cancel := context.WithTimeout(ctx, l.stateTimeout)
	defer cancel()

	fmt.Fprintf(os.Stderr, "[Agent] Entered state\n")

	for {
		select {
		case <-stateCtx.Done():
			return fmt.Errorf("state timeout exceeded")
		default:
		}

		// Build system prompt with current state, allowed tools, and the
		// active case_id so Gemma can reuse it in subsequent tool calls.
		currentState := l.fsm.CurrentState()
		tools := l.getAvailableTools(currentState)
		systemPrompt := BuildSystemPrompt(currentState, tools, l.fsm.CaseID())

		// Prepare messages for Ollama (system + conversation history)
		messages := l.buildMessages(systemPrompt)

		// Enforce LLM turn budget before calling Ollama
		if l.fsm.IncrementLLMTurns() {
			fmt.Fprintf(os.Stderr, "[Agent] LLM turn budget exceeded, exiting state\n")
			return nil
		}

		// Call Ollama to get action
		action, err := l.callOllama(stateCtx, messages)
		if err != nil {
			return fmt.Errorf("ollama call failed: %w", err)
		}

		// Add assistant response to history
		l.addToHistory("assistant", action.Reason, map[string]interface{}{
			"action":    action.Type,
			"tool_name": action.ToolName,
		})

		// Process action
		switch action.Type {
		case "call_tool":
			err := l.callTool(stateCtx, action)
			if err != nil {
				fmt.Fprintf(os.Stderr, "[Agent] Tool call failed: %v\n", err)
				// Add error to history for next iteration
				l.addToHistory("system", fmt.Sprintf("Tool call failed: %v", err), nil)
				continue
			}

		case "state_complete":
			from := l.fsm.CurrentState().Name()
			fmt.Fprintf(os.Stderr, "[Agent] State complete in %s, advancing FSM...\n", from)
			if err := l.fsm.Transition("agent state_complete"); err != nil {
				// LOCK is terminal — Transition returns error there. Treat as natural exit.
				fmt.Fprintf(os.Stderr, "[Agent] Cannot advance from %s: %v\n", from, err)
				return nil
			}
			to := l.fsm.CurrentState().Name()
			// Persist new state to state.json so it stays in sync with audit.jsonl.
			if err := l.fsm.SaveState(); err != nil {
				fmt.Fprintf(os.Stderr, "[Agent] Warning: failed to persist state after %s -> %s transition: %v\n", from, to, err)
				// Don't fail the loop — audit.jsonl is canonical.
			}
			fmt.Fprintf(os.Stderr, "[Agent] Advanced %s -> %s\n", from, to)
			// Continue the loop; next iteration rebuilds system prompt for the new state.
			continue

		case "escalate":
			fmt.Fprintf(os.Stderr, "[Agent] Escalating: %s\n", action.Reason)
			return fmt.Errorf("escalated: %s", action.Reason)

		default:
			fmt.Fprintf(os.Stderr, "[Agent] Unknown action type: %s\n", action.Type)
			l.addToHistory("system", fmt.Sprintf("Unknown action type: %s", action.Type), nil)
		}
	}
}

// callOllama sends a chat request to Ollama and returns the parsed action.
func (l *Loop) callOllama(ctx context.Context, messages []ollama.Message) (*Action, error) {
	resp, err := l.ollamaClient.Chat(ctx, messages)
	if err != nil {
		return nil, err
	}

	// Parse the assistant's response into an Action
	action := ParseAction(resp.Message.Content)

	// Log gemma_response audit event (non-blocking)
	if l.auditLog != nil {
		truncated := resp.Message.Content
		if len(truncated) > 200 {
			truncated = truncated[:200]
		}
		currentStateName := ""
		if l.fsm != nil && l.fsm.CurrentState() != nil {
			currentStateName = l.fsm.CurrentState().Name()
		}
		_ = l.auditLog.Append(map[string]interface{}{
			"timestamp":     time.Now().UTC().Format(time.RFC3339),
			"event":         "gemma_response",
			"action_type":   action.Type,
			"tool_name":     action.ToolName,
			"arguments":     action.Arguments,
			"reason":        action.Reason,
			"raw_output":    truncated,
			"current_state": currentStateName,
		})
	}

	return action, nil
}

// callTool validates and executes a tool call via the MCP client.
func (l *Loop) callTool(ctx context.Context, action *Action) error {
	// Validate that tool is allowed in current state
	if !l.fsm.IsToolAllowed(action.ToolName) {
		msg := fmt.Sprintf("Tool %s not allowed in state", action.ToolName)
		l.addToHistory("system", msg, nil)
		return fmt.Errorf(msg)
	}

	// Log tool_called event before invocation
	if l.auditLog != nil {
		_ = l.auditLog.Append(map[string]interface{}{
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"event":     "tool_called",
			"tool_name": action.ToolName,
			"arguments": action.Arguments,
		})
	}

	// Call the tool via MCP (no context parameter)
	result, err := l.mcpClient.CallTool(action.ToolName, action.Arguments)

	// Log tool_result event after invocation (regardless of success/failure)
	if l.auditLog != nil {
		success := err == nil
		_ = l.auditLog.Append(map[string]interface{}{
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"event":     "tool_result",
			"tool_name": action.ToolName,
			"success":   success,
		})
	}

	if err != nil {
		msg := fmt.Sprintf("Tool call failed: %v", err)
		l.addToHistory("system", msg, nil)
		return err
	}

	// Extract text from result.Content
	output := ""
	if result != nil && len(result.Content) > 0 {
		output = result.Content[0].Text
	}

	// Sanitize the output
	sanitized := SanitizeToolOutput(output, l.maxToolOutputLen)
	enveloped := SafeEnvelopeToolOutput(action.ToolName, sanitized)

	// Add to conversation history
	l.addToHistory("system", enveloped, map[string]interface{}{
		"tool":    action.ToolName,
		"success": true,
	})

	fmt.Fprintf(os.Stderr, "[Agent] Tool %s called successfully\n", action.ToolName)
	return nil
}

// buildMessages constructs the message list for Ollama, including system prompt and history.
func (l *Loop) buildMessages(systemPrompt string) []ollama.Message {
	messages := make([]ollama.Message, 0)

	// Add system prompt
	messages = append(messages, ollama.Message{
		Role:    "system",
		Content: systemPrompt,
	})

	// Add conversation history (trimmed if necessary)
	trimmed := l.trimHistory()
	for _, turn := range trimmed {
		messages = append(messages, ollama.Message{
			Role:    turn.Role,
			Content: turn.Content,
		})
	}

	return messages
}

// trimHistory returns the most recent turns, keeping total message count reasonable.
func (l *Loop) trimHistory() []ConversationTurn {
	if len(l.conversationHist) <= l.maxHistorySize {
		return l.conversationHist
	}

	// Keep the most recent turns
	start := len(l.conversationHist) - l.maxHistorySize
	return l.conversationHist[start:]
}

// addToHistory adds a new turn to the conversation history.
func (l *Loop) addToHistory(role, content string, metadata map[string]interface{}) {
	turn := ConversationTurn{
		Timestamp: time.Now(),
		Role:      role,
		Content:   content,
		Metadata:  metadata,
	}
	l.conversationHist = append(l.conversationHist, turn)
}

// getAvailableTools returns the set of tools available in the current state.
func (l *Loop) getAvailableTools(state hunt.State) []ToolInfo {
	if state == nil {
		return []ToolInfo{}
	}

	catalog := map[string]ToolInfo{
		"case.open": {
			Name:        "case.open",
			Description: "Open a forensic case from /evidence and create the case sandbox.",
		},
		"timeline.build": {
			Name:        "timeline.build",
			Description: "Build a forensic timeline for the active case.",
			Parameters:  map[string]interface{}{"case_id": "string", "limit": "integer optional"},
		},
		"log.query": {
			Name:        "log.query",
			Description: "Query structured logs and timeline records for suspicious activity.",
			Parameters:  map[string]interface{}{"case_id": "string", "q": "string", "source": "evtx|syslog|audit|all optional"},
		},
		"iocs.scan": {
			Name:        "iocs.scan",
			Description: "Scan case evidence against YARA or Sigma indicators of compromise.",
			Parameters:  map[string]interface{}{"case_id": "string", "ruleset": "yara_default|yara_custom|sigma"},
		},
		"memory.process_list": {
			Name:        "memory.process_list",
			Description: "Extract a process list from memory evidence using the configured SIFT backend.",
			Parameters:  map[string]interface{}{"case_id": "string"},
		},
		"memory.malfind": {
			Name:        "memory.malfind",
			Description: "Detect suspicious memory regions and injected code candidates.",
			Parameters:  map[string]interface{}{"case_id": "string", "pid": "integer optional"},
		},
		"net.flow_summary": {
			Name:        "net.flow_summary",
			Description: "Summarize network flows for a case or packet capture.",
			Parameters:  map[string]interface{}{"case_id": "string optional", "pcap_id": "string optional"},
		},
		"verify.cross_check": {
			Name:        "verify.cross_check",
			Description: "Cross-check a finding against independent evidence before reporting.",
			Parameters:  map[string]interface{}{"case_id": "string", "finding_id": "string", "method": "rerun|alt_tool|counter_evidence|all optional"},
		},
		"report.append": {
			Name:        "report.append",
			Description: "Append a verified finding to the case report.",
			Parameters:  map[string]interface{}{"case_id": "string", "finding": "object"},
		},
	}

	tools := make([]ToolInfo, 0, len(state.AllowedTools()))
	for _, name := range state.AllowedTools() {
		if info, ok := catalog[name]; ok {
			tools = append(tools, info)
			continue
		}
		tools = append(tools, ToolInfo{Name: name, Description: "State-allowed MCP tool."})
	}
	return tools
}

// isToolAllowed checks if a tool is allowed in a given FSM state.
func isToolAllowed(state hunt.State, toolName string) bool {
	if state == nil {
		return false
	}
	for _, allowed := range state.AllowedTools() {
		if allowed == toolName {
			return true
		}
	}
	return false
}
