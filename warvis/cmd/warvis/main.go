package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/woopsfactory/warvis/internal/agent"
	"github.com/woopsfactory/warvis/internal/config"
	"github.com/woopsfactory/warvis/internal/hunt"
	"github.com/woopsfactory/warvis/internal/mcp"
	"github.com/woopsfactory/warvis/pkg/ollama"
)

// serverCmd returns the command to spawn the MCP server.
// Reads FIND_EVIL_SERVER_CMD (space-separated); defaults to "python -m find_evil_mcp.server".
func serverCmd() []string {
	if v := os.Getenv("FIND_EVIL_SERVER_CMD"); v != "" {
		return strings.Fields(v)
	}
	return []string{"python", "-m", "find_evil_mcp.server"}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: warvis <command> [args]")
		fmt.Fprintln(os.Stderr, "  hunt <evidence-path> [--case-id <id>] [--phase <STATE>]")
		fmt.Fprintln(os.Stderr, "      — new hunt: open case and advance to TRACE")
		fmt.Fprintln(os.Stderr, "      — resume: hunt --case-id <id> --phase <STATE> (loads prior state + budgets)")
		fmt.Fprintln(os.Stderr, "  status <case-id>       — read case state and output JSON")
		fmt.Fprintln(os.Stderr, "  report <case-id>       — read audit log and generate report")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "hunt":
		if err := runHunt(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "status":
		if err := runStatus(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func runHunt(args []string) error {
	// Parse flags: hunt <evidence-path> [--case-id <id>] [--phase <STATE>]
	var (
		evidencePath string
		caseID       string
		phase        string
	)

	if len(args) < 1 {
		return fmt.Errorf("usage: warvis hunt <evidence-path> [--case-id <id>] [--phase <STATE>]")
	}

	evidencePath = args[0]
	// Parse optional flags
	for i := 1; i < len(args); i++ {
		if args[i] == "--case-id" && i+1 < len(args) {
			caseID = args[i+1]
			i++
		} else if args[i] == "--phase" && i+1 < len(args) {
			phase = args[i+1]
			i++
		}
	}

	casesRoot := os.Getenv("FIND_EVIL_CASES_ROOT")
	if casesRoot == "" {
		casesRoot = "./.cases"
	}

	registry := hunt.NewRegistry(casesRoot)

	// Resume mode: load prior case state
	if caseID != "" {
		record, err := registry.Load(caseID)
		if err != nil {
			return fmt.Errorf("failed to load case state: %w", err)
		}
		if record == nil {
			return fmt.Errorf("case not found: %s", caseID)
		}

		caseDir := filepath.Join(casesRoot, caseID)
		auditLogPath := filepath.Join(caseDir, "audit.jsonl")
		auditLog, err := hunt.NewAuditLog(auditLogPath)
		if err != nil {
			return fmt.Errorf("failed to open audit log: %w", err)
		}
		defer auditLog.Close()

		// Reconstruct FSM from prior state
		fsm := hunt.New(caseID, auditLog, registry)
		fsm.SetBudgets(record.Budgets)
		restoredState := registry.LoadState(record)
		if restoredState == nil {
			return fmt.Errorf("failed to restore state from record")
		}
		fsm.SetCurrentState(restoredState)

		// Resume: clear paused flag if set
		if err := fsm.Resume(); err != nil {
			// Resume may fail if not paused; that's OK
			_ = err
		}

		if err := auditLog.Append(map[string]interface{}{
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"event":     "hunt_resumed",
			"case_id":   caseID,
			"prior_phase": record.CurrentState,
			"target_phase": phase,
		}); err != nil {
			return fmt.Errorf("failed to log hunt_resumed: %w", err)
		}

		if err := registry.Save(caseID, fsm.CurrentState(), fsm.GetBudgetStatus()); err != nil {
			return fmt.Errorf("failed to save state after resume: %w", err)
		}

		out, _ := json.Marshal(map[string]string{
			"case_id":       caseID,
			"current_state": fsm.CurrentState().Name(),
			"state_file":    filepath.Join(caseDir, "state.json"),
			"audit_log":     auditLogPath,
			"mode":          "resumed",
		})
		fmt.Println(string(out))
		return nil
	}

	// New hunt mode: validate evidence file exists
	if _, err := os.Stat(evidencePath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("evidence file not found: %s", evidencePath)
		}
		return fmt.Errorf("cannot access evidence file: %w", err)
	}

	// Load timeout configuration
	timeoutCfg := config.LoadTimeoutConfig()

	ctx, cancel := context.WithTimeout(context.Background(), timeoutCfg.HuntTimeout)
	defer cancel()

	mcpClient, err := mcp.NewClient(ctx, serverCmd())
	if err != nil {
		return fmt.Errorf("failed to start MCP server (is Python MCP installed?): %w", err)
	}
	defer mcpClient.Close()

	result, err := mcpClient.CallTool("case.open", evidenceArgs(evidencePath))
	if err != nil {
		return fmt.Errorf("case.open failed: %w", err)
	}

	caseInfo, err := parseCaseOpenResult(result)
	if err != nil {
		return fmt.Errorf("failed to parse case.open result: %w", err)
	}

	newCaseID, ok := caseInfo["case_id"].(string)
	if !ok || newCaseID == "" {
		return fmt.Errorf("case.open returned empty case_id")
	}
	sandboxRoot, _ := caseInfo["sandbox_root"].(string)

	caseDir := filepath.Join(casesRoot, newCaseID)
	if err := os.MkdirAll(caseDir, 0755); err != nil {
		return fmt.Errorf("failed to create case dir: %w", err)
	}

	auditLogPath := filepath.Join(caseDir, "audit.jsonl")
	auditLog, err := hunt.NewAuditLog(auditLogPath)
	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}
	defer auditLog.Close()

	if err := auditLog.Append(map[string]interface{}{
		"timestamp":     time.Now().UTC().Format(time.RFC3339),
		"event":         "case_opened",
		"case_id":       newCaseID,
		"evidence_path": evidencePath,
		"sandbox_root":  sandboxRoot,
	}); err != nil {
		return fmt.Errorf("failed to log case_opened: %w", err)
	}

	fsm := hunt.New(newCaseID, auditLog, registry)

	if err := fsm.Transition("case opened via case.open"); err != nil {
		return fmt.Errorf("failed to transition to TRACE: %w", err)
	}

	if err := registry.Save(newCaseID, fsm.CurrentState(), fsm.GetBudgetStatus()); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	// Initialize Ollama client
	ollamaURL := os.Getenv("OLLAMA_URL")
	if ollamaURL == "" {
		ollamaURL = "http://localhost:29134"
	}
	ollamaModel := os.Getenv("OLLAMA_MODEL")
	if ollamaModel == "" {
		ollamaModel = "gemma4:26b-a4b-it-q4_K_M"
	}

	ollamaClient := ollama.NewClient(ollamaURL, ollamaModel)

	// Create and run agent loop
	loop := agent.NewLoop(fsm, mcpClient, ollamaClient, auditLog)
	if err := loop.Run(ctx); err != nil {
		// Log error but don't fail the hunt
		fmt.Printf("[Agent] Loop exited with error: %v\n", err)
		if err := auditLog.Append(map[string]interface{}{
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"event":     "loop_error",
			"error":     err.Error(),
		}); err != nil {
			fmt.Printf("[Audit] Failed to log loop error: %v\n", err)
		}
	} else {
		if err := auditLog.Append(map[string]interface{}{
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"event":     "loop_completed",
		}); err != nil {
			fmt.Printf("[Audit] Failed to log loop completion: %v\n", err)
		}
	}

	out, _ := json.Marshal(map[string]string{
		"case_id":       newCaseID,
		"current_state": fsm.CurrentState().Name(),
		"state_file":    filepath.Join(caseDir, "state.json"),
		"audit_log":     auditLogPath,
		"mode":          "new",
	})
	fmt.Println(string(out))
	return nil
}

func evidenceArgs(path string) map[string]string {
	lower := strings.ToLower(path)
	switch {
	case strings.HasSuffix(lower, ".mem") || strings.HasSuffix(lower, ".dmp") || strings.HasSuffix(lower, ".vmem"):
		return map[string]string{"mem_path": path}
	case strings.HasSuffix(lower, ".pcap") || strings.HasSuffix(lower, ".pcapng"):
		return map[string]string{"pcap_path": path}
	default:
		return map[string]string{"image_path": path}
	}
}

func parseCaseOpenResult(result *mcp.CallToolResult) (map[string]interface{}, error) {
	if result == nil || len(result.Content) == 0 {
		return nil, fmt.Errorf("empty result from case.open")
	}
	for _, item := range result.Content {
		if item.Type == "text" {
			var m map[string]interface{}
			if err := json.Unmarshal([]byte(item.Text), &m); err != nil {
				return nil, fmt.Errorf("failed to parse case.open response: %w", err)
			}
			return m, nil
		}
	}
	return nil, fmt.Errorf("no text content in case.open result")
}

func runStatus(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: warvis status <case-id>")
	}
	caseID := args[0]

	casesRoot := os.Getenv("FIND_EVIL_CASES_ROOT")
	if casesRoot == "" {
		casesRoot = "./.cases"
	}

	stateFile := filepath.Join(casesRoot, caseID, "state.json")

	// Read state.json
	data, err := os.ReadFile(stateFile)
	if err != nil {
		return fmt.Errorf("failed to read state file: %w", err)
	}

	var stateData map[string]interface{}
	if err := json.Unmarshal(data, &stateData); err != nil {
		return fmt.Errorf("failed to parse state file: %w", err)
	}

	// Extract relevant fields and output JSON
	output := map[string]interface{}{
		"case_id":       caseID,
		"current_state": stateData["current_state"],
		"started_at":    stateData["started_at"],
		"updated_at":    stateData["updated_at"],
		"llm_turns":     stateData["llm_turns"],
		"tool_calls":    stateData["tool_calls"],
	}

	out, _ := json.Marshal(output)
	fmt.Println(string(out))
	return nil
}
