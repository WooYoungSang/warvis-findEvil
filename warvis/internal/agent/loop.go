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
		maxHistorySize:   20, // Max 20 exchanges
		maxToolOutputLen: 500, // Truncate tool outputs to 500 chars
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

		// Build system prompt with current state and allowed tools
		currentState := l.fsm.CurrentState()
		tools := l.getAvailableTools(currentState)
		systemPrompt := BuildSystemPrompt(currentState, tools)

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
			fmt.Fprintf(os.Stderr, "[Agent] State complete, transitioning...\n")
			return nil

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
		_ = l.auditLog.Append(map[string]interface{}{
			"timestamp":    time.Now().UTC().Format(time.RFC3339),
			"event":        "gemma_response",
			"gemma_output": truncated,
			"action_type":  action.Type,
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
	// Map FSM state to available tools
	// In a real implementation, this would query the MCP client and filter by state
	tools := []ToolInfo{}

	// For now, return empty; in M4e, this will call mcpClient.ListTools() and filter
	return tools
}

// isToolAllowed checks if a tool is allowed in a given FSM state.
func isToolAllowed(state hunt.State, toolName string) bool {
	// Simple mapping of state -> allowed tools
	// In a real implementation, this might be more sophisticated
	switch state.Name() {
	case "INITIALIZE":
		return toolName == "case.open"
	case "TRACE":
		return toolName == "timeline.build" || toolName == "log.query"
	case "SCAN":
		return toolName == "iocs.scan" || toolName == "memory.dump" || toolName == "memory.list"
	case "EXPOSE":
		return toolName == "verify.cross_check"
	case "LOCK":
		return toolName == "report.append"
	default:
		return false
	}
}
