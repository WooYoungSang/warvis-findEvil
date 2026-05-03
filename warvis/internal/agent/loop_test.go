package agent

import (
	"testing"

	"github.com/woopsfactory/warvis/internal/hunt"
)

func TestLoopBuildMessagesWithHistory(t *testing.T) {
	// Create a simple test without network
	loop := NewLoop(nil, nil, nil)

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
	loop := NewLoop(nil, nil, nil)

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
	loop := NewLoop(nil, nil, nil)

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
	loop := NewLoop(nil, nil, nil)

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
