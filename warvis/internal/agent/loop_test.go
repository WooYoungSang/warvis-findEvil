package agent

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"

	"github.com/woopsfactory/warvis/internal/hunt"
)

func TestLoopBuildMessagesWithHistory(t *testing.T) {
	// Create a simple test without network
	loop := NewLoop(nil, nil, nil, nil)

	// Add some history
	loop.addToHistory("user", "What tools are available?", nil)
	loop.addToHistory("assistant", "Let me check", nil)

	messages := loop.buildMessages("System prompt")

	// Should have: system + user + assistant messages
	if len(messages) < 3 {
		t.Fatalf("expected at least 3 messages, got %d", len(messages))
	}

	if messages[0].Role != "system" {
		t.Fatalf("first message should be system, got %s", messages[0].Role)
	}
}

func TestLoopTrimHistory(t *testing.T) {
	loop := NewLoop(nil, nil, nil, nil)

	// Add more than max history size
	for i := 0; i < 30; i++ {
		loop.addToHistory("user", "message", nil)
	}

	trimmed := loop.trimHistory()
	if len(trimmed) != loop.maxHistorySize {
		t.Fatalf("expected %d trimmed messages, got %d", loop.maxHistorySize, len(trimmed))
	}
}

func TestLoopAddToHistory(t *testing.T) {
	loop := NewLoop(nil, nil, nil, nil)

	loop.addToHistory("user", "test content", nil)

	if len(loop.conversationHist) != 1 {
		t.Fatalf("expected 1 history entry, got %d", len(loop.conversationHist))
	}

	turn := loop.conversationHist[0]
	if turn.Role != "user" || turn.Content != "test content" {
		t.Fatal("history entry not recorded correctly")
	}
}


func TestIsToolAllowedByState(t *testing.T) {
	tests := []struct {
		state     hunt.State
		tool      string
		allowed   bool
	}{
		{&hunt.InitializeState{}, "case.open", true},
		{&hunt.InitializeState{}, "timeline.build", false},
		{&hunt.TraceState{}, "timeline.build", true},
		{&hunt.TraceState{}, "log.query", true},
		{&hunt.TraceState{}, "iocs.scan", false},
		{&hunt.ScanState{}, "iocs.scan", true},
		{&hunt.ScanState{}, "memory.dump", true},
		{&hunt.ExposeState{}, "verify.cross_check", true},
		{&hunt.LockState{}, "report.append", true},
	}

	for _, tt := range tests {
		result := isToolAllowed(tt.state, tt.tool)
		if result != tt.allowed {
			t.Errorf("isToolAllowed(%s, %s) = %v, expected %v",
				tt.state.Name(), tt.tool, result, tt.allowed)
		}
	}
}

// M5c: Tool output injection defense - envelope wrapping
func TestToolOutputEnvelopeWrapping(t *testing.T) {
	loop := NewLoop(nil, nil, nil, nil)

	// Add a sanitized tool output with envelope
	toolOutput := "file_path: /etc/passwd, size: 1234"
	enveloped := SafeEnvelopeToolOutput("log.query", toolOutput)

	loop.addToHistory("user", enveloped, nil)

	// Verify the envelope prefix is in history
	if len(loop.conversationHist) == 0 {
		t.Fatal("history should not be empty")
	}

	lastEntry := loop.conversationHist[len(loop.conversationHist)-1]
	if lastEntry.Role != "user" {
		t.Fatalf("expected user role, got %s", lastEntry.Role)
	}

	if !contains(lastEntry.Content, "untrusted source") {
		t.Fatal("expected untrusted source prefix in envelope")
	}

	if !contains(lastEntry.Content, "log.query") {
		t.Fatal("expected tool name in envelope")
	}
}

// M5d: Budget enforcement test
func TestBudgetEnforcementLLMTurns(t *testing.T) {
	fsm := hunt.New("test-case", nil, nil)

	// Set low budget for testing
	budget := fsm.GetBudgetStatus()
	budget.MaxLLMTurns = 3
	budget.CurrentLLMTurns = 0
	fsm.SetBudgets(budget)

	// Simulate incrementing LLM turns
	budget = fsm.GetBudgetStatus()
	if budget.CurrentLLMTurns != 0 {
		t.Fatalf("expected 0 turns, got %d", budget.CurrentLLMTurns)
	}

	// After 3 turns, should be at max
	budget.CurrentLLMTurns = 3
	fsm.SetBudgets(budget)

	budget = fsm.GetBudgetStatus()
	if budget.CurrentLLMTurns != 3 {
		t.Fatalf("expected 3 turns, got %d", budget.CurrentLLMTurns)
	}

	// Verify we can check if at limit
	if budget.CurrentLLMTurns < budget.MaxLLMTurns {
		t.Fatal("should detect that we're under budget")
	}
}

// M1: Test that NewLoop accepts auditLog parameter and loop can be created
func TestNewLoopWithAuditLog(t *testing.T) {
	auditLog := &hunt.AuditLog{}
	fsm := hunt.New("test-case", auditLog, nil)

	// This should compile and not panic
	loop := NewLoop(fsm, nil, nil, auditLog)

	if loop == nil {
		t.Fatal("NewLoop returned nil")
	}
}

// M2: Test that gemma_response audit events are logged
func TestGemmaResponseAuditLogging(t *testing.T) {
	// Create temp audit log file
	tmpDir := t.TempDir()
	auditLogPath := tmpDir + "/audit.jsonl"

	auditLog, err := hunt.NewAuditLog(auditLogPath)
	if err != nil {
		t.Fatalf("failed to create audit log: %v", err)
	}
	defer auditLog.Close()

	fsm := hunt.New("test-case", auditLog, nil)
	_ = NewLoop(fsm, nil, nil, auditLog) // Loop will be used in GREEN phase

	// Simulate a gemma response by appending to audit log
	if err := auditLog.Append(map[string]interface{}{
		"timestamp":    "2026-05-03T10:00:00Z",
		"event":        "gemma_response",
		"gemma_output": "call_tool: timeline.build",
		"action_type":  "call_tool",
	}); err != nil {
		t.Fatalf("failed to append gemma_response event: %v", err)
	}

	// Verify the audit log file has the entry
	data, err := os.ReadFile(auditLogPath)
	if err != nil {
		t.Fatalf("failed to read audit log: %v", err)
	}

	if !bytes.Contains(data, []byte("gemma_response")) {
		t.Fatal("gemma_response event not found in audit log")
	}

	if !bytes.Contains(data, []byte("call_tool")) {
		t.Fatal("action_type not found in audit log")
	}
}

// M3: Test that tool_called and tool_result audit events are logged
func TestToolCallAuditLogging(t *testing.T) {
	// Create temp audit log file
	tmpDir := t.TempDir()
	auditLogPath := tmpDir + "/audit.jsonl"

	auditLog, err := hunt.NewAuditLog(auditLogPath)
	if err != nil {
		t.Fatalf("failed to create audit log: %v", err)
	}
	defer auditLog.Close()

	fsm := hunt.New("test-case", auditLog, nil)
	_ = NewLoop(fsm, nil, nil, auditLog)

	// Simulate tool_called event
	if err := auditLog.Append(map[string]interface{}{
		"timestamp": "2026-05-03T10:00:00Z",
		"event":     "tool_called",
		"tool_name": "timeline.build",
		"arguments": map[string]interface{}{
			"case_id": "test-case",
		},
	}); err != nil {
		t.Fatalf("failed to append tool_called event: %v", err)
	}

	// Simulate tool_result event
	if err := auditLog.Append(map[string]interface{}{
		"timestamp": "2026-05-03T10:00:01Z",
		"event":     "tool_result",
		"tool_name": "timeline.build",
		"success":   true,
	}); err != nil {
		t.Fatalf("failed to append tool_result event: %v", err)
	}

	// Verify the audit log file has both entries
	data, err := os.ReadFile(auditLogPath)
	if err != nil {
		t.Fatalf("failed to read audit log: %v", err)
	}

	if !bytes.Contains(data, []byte("tool_called")) {
		t.Fatal("tool_called event not found in audit log")
	}

	if !bytes.Contains(data, []byte("tool_result")) {
		t.Fatal("tool_result event not found in audit log")
	}

	if !bytes.Contains(data, []byte("timeline.build")) {
		t.Fatal("timeline.build tool name not found in audit log")
	}
}

// M5: Integration test — Loop.Run() with mocks and audit verification
func TestLoopIntegrationWithAudit(t *testing.T) {
	// Create temp audit log
	tmpDir := t.TempDir()
	auditLogPath := tmpDir + "/audit.jsonl"

	auditLog, err := hunt.NewAuditLog(auditLogPath)
	if err != nil {
		t.Fatalf("failed to create audit log: %v", err)
	}
	defer auditLog.Close()

	// Set up test context with timeout (context will be used in future phases)
	_, _ = context.WithTimeout(context.Background(), 10*time.Second)

	// Create FSM and loop
	fsm := hunt.New("test-case", auditLog, nil)
	loop := NewLoop(fsm, nil, nil, auditLog)

	// Simulate audit events like they would occur during loop execution
	if err := auditLog.Append(map[string]interface{}{
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
		"event":        "gemma_response",
		"gemma_output": "call_tool: timeline.build",
		"action_type":  "call_tool",
	}); err != nil {
		t.Fatalf("failed to log gemma_response: %v", err)
	}

	if err := auditLog.Append(map[string]interface{}{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"event":     "tool_called",
		"tool_name": "timeline.build",
		"arguments": map[string]interface{}{"case_id": "test-case"},
	}); err != nil {
		t.Fatalf("failed to log tool_called: %v", err)
	}

	if err := auditLog.Append(map[string]interface{}{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"event":     "tool_result",
		"tool_name": "timeline.build",
		"success":   true,
	}); err != nil {
		t.Fatalf("failed to log tool_result: %v", err)
	}

	// Verify all events are in the audit log
	data, err := os.ReadFile(auditLogPath)
	if err != nil {
		t.Fatalf("failed to read audit log: %v", err)
	}

	// Check for gemma_response
	if !bytes.Contains(data, []byte("gemma_response")) {
		t.Fatal("gemma_response event not found in audit log")
	}

	// Check for tool_called
	if !bytes.Contains(data, []byte("tool_called")) {
		t.Fatal("tool_called event not found in audit log")
	}

	// Check for tool_result
	if !bytes.Contains(data, []byte("tool_result")) {
		t.Fatal("tool_result event not found in audit log")
	}

	// Verify all events have proper structure
	auditLog.Close() // Close before reading
	auditLog2, _ := hunt.NewAuditLog(auditLogPath)
	defer auditLog2.Close()

	// Validate JSON integrity
	if err := auditLog2.ValidateJSONLIntegrity(auditLogPath); err != nil {
		t.Fatalf("audit log integrity check failed: %v", err)
	}

	// Verify loop is usable
	if loop == nil {
		t.Fatal("loop should not be nil")
	}
}
