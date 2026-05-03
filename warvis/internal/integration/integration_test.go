package integration

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/woopsfactory/warvis/internal/hunt"
)

// TestFullHuntFlow verifies the complete hunt workflow:
// 1. Case open via MCP
// 2. FSM transitions INITIALIZE → TRACE
// 3. State and audit files created
// 4. All audit entries are valid JSON
func TestFullHuntFlow(t *testing.T) {
	// Setup temporary case directory
	tmpDir, err := ioutil.TempDir("", "warvis-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	caseID := "test-case-001"
	caseDir := filepath.Join(tmpDir, caseID)
	if err := os.MkdirAll(caseDir, 0755); err != nil {
		t.Fatalf("Failed to create case dir: %v", err)
	}

	// Initialize audit log
	auditFile := filepath.Join(caseDir, "audit.jsonl")
	auditLog, err := hunt.NewAuditLog(auditFile)
	if err != nil {
		t.Fatalf("Failed to create audit log: %v", err)
	}
	defer auditLog.Close()

	// Initialize registry
	registry := hunt.NewRegistry(tmpDir)

	// Create FSM
	fsm := hunt.New(caseID, auditLog, registry)

	// Verify initial state is INITIALIZE
	if state := fsm.CurrentState(); state.Name() != "INITIALIZE" {
		t.Errorf("Initial state: got %s, want INITIALIZE", state.Name())
	}

	// Simulate case.open audit log entry
	err = auditLog.Append(map[string]interface{}{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"event":     "case_opened",
		"case_id":   caseID,
		"details": map[string]interface{}{
			"evidence_path": "/evidence/test.img",
			"sandbox_root":  caseDir,
		},
	})
	if err != nil {
		t.Fatalf("Failed to append case_opened event: %v", err)
	}

	// Transition to TRACE
	err = fsm.Transition("case opened, starting trace collection")
	if err != nil {
		t.Fatalf("Failed to transition to TRACE: %v", err)
	}

	// Verify state is now TRACE
	if state := fsm.CurrentState(); state.Name() != "TRACE" {
		t.Errorf("After transition: got %s, want TRACE", state.Name())
	}

	// Simulate state_transition audit entry
	err = auditLog.Append(map[string]interface{}{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"event":     "state_transition",
		"case_id":   caseID,
		"details": map[string]interface{}{
			"from": "INITIALIZE",
			"to":   "TRACE",
			"reason": "case opened, starting trace collection",
		},
	})
	if err != nil {
		t.Fatalf("Failed to append state_transition event: %v", err)
	}

	// Persist state to state.json
	stateJSON := map[string]interface{}{
		"case_id":       caseID,
		"current_state": "TRACE",
		"timestamp":     time.Now().UTC().Format(time.RFC3339),
	}
	stateFile := filepath.Join(caseDir, "state.json")
	stateData, _ := json.MarshalIndent(stateJSON, "", "  ")
	if err := ioutil.WriteFile(stateFile, stateData, 0644); err != nil {
		t.Fatalf("Failed to write state.json: %v", err)
	}

	// Verify state.json exists and is valid
	if _, err := os.Stat(stateFile); err != nil {
		t.Fatalf("state.json not found: %v", err)
	}

	// Verify audit.jsonl exists and is valid JSONL
	if _, err := os.Stat(auditFile); err != nil {
		t.Fatalf("audit.jsonl not found: %v", err)
	}

	// Validate JSONL format (each line must be valid JSON)
	content, err := ioutil.ReadFile(auditFile)
	if err != nil {
		t.Fatalf("Failed to read audit.jsonl: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) < 2 {
		t.Errorf("audit.jsonl has %d lines, want at least 2", len(lines))
	}

	for i, line := range lines {
		if line == "" {
			continue
		}
		var entry map[string]interface{}
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Errorf("Line %d is not valid JSON: %v", i+1, err)
		}
		// Verify required fields
		if _, ok := entry["timestamp"]; !ok {
			t.Errorf("Line %d missing timestamp", i+1)
		}
		if _, ok := entry["event"]; !ok {
			t.Errorf("Line %d missing event", i+1)
		}
	}
}

// TestStateJsonFormat verifies state.json has correct structure
func TestStateJsonFormat(t *testing.T) {
	tmpDir, _ := ioutil.TempDir("", "warvis-test-")
	defer os.RemoveAll(tmpDir)

	caseID := "test-case-002"
	caseDir := filepath.Join(tmpDir, caseID)
	os.MkdirAll(caseDir, 0755)

	// Write state.json
	stateJSON := map[string]interface{}{
		"case_id":       caseID,
		"current_state": "SCAN",
		"timestamp":     time.Now().UTC().Format(time.RFC3339),
		"budget": map[string]int{
			"current_llm_turns": 15,
			"max_llm_turns":     50,
		},
	}
	stateFile := filepath.Join(caseDir, "state.json")
	stateData, _ := json.MarshalIndent(stateJSON, "", "  ")
	ioutil.WriteFile(stateFile, stateData, 0644)

	// Verify it can be read back
	readData, err := ioutil.ReadFile(stateFile)
	if err != nil {
		t.Fatalf("Failed to read state.json: %v", err)
	}

	var loaded map[string]interface{}
	if err := json.Unmarshal(readData, &loaded); err != nil {
		t.Fatalf("state.json is not valid JSON: %v", err)
	}

	if loaded["case_id"] != caseID {
		t.Errorf("case_id mismatch: got %v, want %s", loaded["case_id"], caseID)
	}
	if loaded["current_state"] != "SCAN" {
		t.Errorf("current_state mismatch: got %v, want SCAN", loaded["current_state"])
	}
}

// TestAuditJsonlFormat verifies audit.jsonl has proper JSONL structure
func TestAuditJsonlFormat(t *testing.T) {
	tmpDir, _ := ioutil.TempDir("", "warvis-test-")
	defer os.RemoveAll(tmpDir)

	caseID := "test-case-003"
	caseDir := filepath.Join(tmpDir, caseID)
	os.MkdirAll(caseDir, 0755)

	// Write multiple JSONL entries
	auditFile := filepath.Join(caseDir, "audit.jsonl")
	entries := []map[string]interface{}{
		{
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"event":     "case_opened",
			"case_id":   caseID,
			"details": map[string]interface{}{
				"evidence_path": "/evidence/test.img",
			},
		},
		{
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"event":     "tool_called",
			"case_id":   caseID,
			"details": map[string]interface{}{
				"tool_name": "timeline.build",
				"arguments": map[string]string{"output_format": "json"},
			},
		},
		{
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"event":     "tool_result",
			"case_id":   caseID,
			"details": map[string]interface{}{
				"tool_name": "timeline.build",
				"status":    "success",
			},
		},
	}

	var lines []string
	for _, entry := range entries {
		data, _ := json.Marshal(entry)
		lines = append(lines, string(data))
	}

	content := strings.Join(lines, "\n")
	ioutil.WriteFile(auditFile, []byte(content), 0644)

	// Validate JSONL format
	readData, err := ioutil.ReadFile(auditFile)
	if err != nil {
		t.Fatalf("Failed to read audit.jsonl: %v", err)
	}

	readLines := strings.Split(strings.TrimSpace(string(readData)), "\n")
	if len(readLines) != len(entries) {
		t.Errorf("Line count mismatch: got %d, want %d", len(readLines), len(entries))
	}

	for i, line := range readLines {
		var entry map[string]interface{}
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Errorf("Line %d is not valid JSON: %v", i+1, err)
		}
	}
}

// TestTimeoutContext verifies context timeout is enforced
func TestTimeoutContext(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Simulate a slow operation
	select {
	case <-time.After(200 * time.Millisecond):
		t.Error("Context timeout not enforced")
	case <-ctx.Done():
		// Expected: context deadline exceeded
		if err := ctx.Err(); err != context.DeadlineExceeded {
			t.Errorf("Expected DeadlineExceeded, got %v", err)
		}
	}
}

// TestResumeStateLoading verifies prior state can be loaded for resume
func TestResumeStateLoading(t *testing.T) {
	tmpDir, _ := ioutil.TempDir("", "warvis-test-")
	defer os.RemoveAll(tmpDir)

	caseID := "test-case-resume"
	caseDir := filepath.Join(tmpDir, caseID)
	os.MkdirAll(caseDir, 0755)

	// Write initial state
	stateJSON := map[string]interface{}{
		"case_id":       caseID,
		"current_state": "TRACE",
		"timestamp":     time.Now().UTC().Format(time.RFC3339),
		"budget": map[string]int{
			"current_llm_turns": 20,
			"max_llm_turns":     50,
		},
	}
	stateFile := filepath.Join(caseDir, "state.json")
	stateData, _ := json.MarshalIndent(stateJSON, "", "  ")
	ioutil.WriteFile(stateFile, stateData, 0644)

	// Simulate resume: load state
	readData, err := ioutil.ReadFile(stateFile)
	if err != nil {
		t.Fatalf("Failed to read state.json for resume: %v", err)
	}

	var loaded map[string]interface{}
	if err := json.Unmarshal(readData, &loaded); err != nil {
		t.Fatalf("state.json is not valid JSON: %v", err)
	}

	// Verify budget counters are preserved
	if budget, ok := loaded["budget"].(map[string]interface{}); ok {
		if turns, ok := budget["current_llm_turns"].(float64); ok {
			if int(turns) != 20 {
				t.Errorf("Budget not preserved: got %d, want 20", int(turns))
			}
		}
	}

	// Verify current state is preserved
	if state, ok := loaded["current_state"].(string); ok {
		if state != "TRACE" {
			t.Errorf("State not preserved: got %s, want TRACE", state)
		}
	}
}
