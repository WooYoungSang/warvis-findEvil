package hunt

import (
	"context"
	"testing"
)

func TestFSMNew(t *testing.T) {
	caseID := "test-case-123"
	fsm := New(caseID, nil, nil)

	if fsm == nil {
		t.Fatal("FSM should not be nil")
	}

	state := fsm.CurrentState()
	if state == nil {
		t.Fatal("initial state should not be nil")
	}

	if state.Name() != "INITIALIZE" {
		t.Errorf("expected INITIALIZE state, got %s", state.Name())
	}
}

func TestCurrentState(t *testing.T) {
	fsm := New("test-case", nil, nil)

	state := fsm.CurrentState()
	if state.Name() != "INITIALIZE" {
		t.Errorf("expected INITIALIZE, got %s", state.Name())
	}
}

func TestIsToolAllowed(t *testing.T) {
	fsm := New("test-case", nil, nil)

	tests := []struct {
		toolName string
		expected bool
	}{
		{"case.open", true},
		{"timeline.build", false},
		{"iocs.scan", false},
	}

	for _, test := range tests {
		if allowed := fsm.IsToolAllowed(test.toolName); allowed != test.expected {
			t.Errorf("IsToolAllowed(%s) = %v, expected %v", test.toolName, allowed, test.expected)
		}
	}
}

func TestToolAllowedAfterTransition(t *testing.T) {
	fsm := New("test-case", nil, nil)

	// INITIALIZE allows case.open
	if !fsm.IsToolAllowed("case.open") {
		t.Fatal("case.open should be allowed in INITIALIZE")
	}

	// Transition to TRACE
	if err := fsm.Transition("success"); err != nil {
		t.Fatalf("transition failed: %v", err)
	}

	// TRACE allows timeline.build
	if !fsm.IsToolAllowed("timeline.build") {
		t.Fatal("timeline.build should be allowed in TRACE")
	}

	// TRACE does not allow case.open
	if fsm.IsToolAllowed("case.open") {
		t.Fatal("case.open should not be allowed in TRACE")
	}
}

func TestCallTool(t *testing.T) {
	fsm := New("test-case", nil, nil)

	result, err := fsm.CallTool(context.Background(), "case.open", map[string]interface{}{
		"evidence_path": "/evidence/test.img",
	})

	if err != nil {
		t.Fatalf("CallTool failed: %v", err)
	}

	if result == "" {
		t.Fatal("CallTool should return non-empty result")
	}
}

func TestCallToolNotAllowed(t *testing.T) {
	fsm := New("test-case", nil, nil)

	// case.open is in INITIALIZE, but timeline.build is not
	_, err := fsm.CallTool(context.Background(), "timeline.build", map[string]interface{}{})

	if err == nil {
		t.Fatal("expected error for disallowed tool")
	}
}

func TestTransitionINITIALIZE_TO_TRACE(t *testing.T) {
	fsm := New("test-case", nil, nil)

	if fsm.CurrentState().Name() != "INITIALIZE" {
		t.Fatal("initial state should be INITIALIZE")
	}

	if err := fsm.Transition("success"); err != nil {
		t.Fatalf("transition failed: %v", err)
	}

	if fsm.CurrentState().Name() != "TRACE" {
		t.Errorf("expected TRACE, got %s", fsm.CurrentState().Name())
	}
}

func TestTransitionTRACE_TO_SCAN(t *testing.T) {
	fsm := New("test-case", nil, nil)

	// Transition to TRACE
	_ = fsm.Transition("to trace")

	// Transition to SCAN
	if err := fsm.Transition("events found"); err != nil {
		t.Fatalf("transition to SCAN failed: %v", err)
	}

	if fsm.CurrentState().Name() != "SCAN" {
		t.Errorf("expected SCAN, got %s", fsm.CurrentState().Name())
	}
}

func TestTransitionSCAN_TO_EXPOSE(t *testing.T) {
	fsm := New("test-case", nil, nil)

	_ = fsm.Transition("to trace")
	_ = fsm.Transition("to scan")

	if err := fsm.Transition("findings detected"); err != nil {
		t.Fatalf("transition to EXPOSE failed: %v", err)
	}

	if fsm.CurrentState().Name() != "EXPOSE" {
		t.Errorf("expected EXPOSE, got %s", fsm.CurrentState().Name())
	}
}

func TestTransitionEXPOSE_TO_LOCK(t *testing.T) {
	fsm := New("test-case", nil, nil)

	_ = fsm.Transition("to trace")
	_ = fsm.Transition("to scan")
	_ = fsm.Transition("to expose")

	if err := fsm.Transition("all verified"); err != nil {
		t.Fatalf("transition to LOCK failed: %v", err)
	}

	if fsm.CurrentState().Name() != "LOCK" {
		t.Errorf("expected LOCK, got %s", fsm.CurrentState().Name())
	}
}

func TestTransitionFromLockFails(t *testing.T) {
	fsm := New("test-case", nil, nil)

	// Transition through all states to LOCK
	_ = fsm.Transition("")
	_ = fsm.Transition("")
	_ = fsm.Transition("")
	_ = fsm.Transition("")

	// Try to transition from LOCK (should fail)
	lockErr := fsm.Transition("invalid")

	if lockErr == nil {
		t.Fatal("expected error transitioning from LOCK state")
	}
}

func TestPause(t *testing.T) {
	fsm := New("test-case", nil, nil)

	if err := fsm.Pause("user requested pause"); err != nil {
		t.Fatalf("pause failed: %v", err)
	}

	// Should not be able to transition while paused
	pauseErr := fsm.Transition("try to continue")
	if pauseErr == nil {
		t.Fatal("expected error transitioning while paused")
	}
}

func TestResume(t *testing.T) {
	fsm := New("test-case", nil, nil)

	_ = fsm.Pause("user pause")

	if err := fsm.Resume(); err != nil {
		t.Fatalf("resume failed: %v", err)
	}

	// Should be able to transition again
	if err := fsm.Transition("continue hunt"); err != nil {
		t.Fatalf("transition after resume failed: %v", err)
	}
}

func TestGetBudgetStatus(t *testing.T) {
	fsm := New("test-case", nil, nil)

	budget := fsm.GetBudgetStatus()

	if budget.MaxLLMTurns != 50 {
		t.Errorf("expected MaxLLMTurns=50, got %d", budget.MaxLLMTurns)
	}

	if budget.MaxInvalidJSONAttempts != 10 {
		t.Errorf("expected MaxInvalidJSONAttempts=10, got %d", budget.MaxInvalidJSONAttempts)
	}
}

func TestBudgetIncrement(t *testing.T) {
	fsm := New("test-case", nil, nil)

	// Call tool to increment counter
	fsm.CallTool(context.Background(), "case.open", map[string]interface{}{})

	budget := fsm.GetBudgetStatus()
	if budget.CurrentToolCallsTotal != 1 {
		t.Errorf("expected CurrentToolCallsTotal=1, got %d", budget.CurrentToolCallsTotal)
	}
}

func TestSetBudgets(t *testing.T) {
	fsm := New("test-case", nil, nil)

	newBudgets := BudgetStatus{
		MaxLLMTurns:             40,
		CurrentLLMTurns:         25,
		MaxInvalidJSONAttempts:  8,
		CurrentInvalidJSONAttempts: 3,
	}

	fsm.SetBudgets(newBudgets)

	budget := fsm.GetBudgetStatus()
	if budget.CurrentLLMTurns != 25 {
		t.Errorf("expected CurrentLLMTurns=25, got %d", budget.CurrentLLMTurns)
	}
}

func TestSetCurrentState(t *testing.T) {
	fsm := New("test-case", nil, nil)

	newState := &TraceState{CaseID: "test-case"}
	fsm.SetCurrentState(newState)

	if fsm.CurrentState().Name() != "TRACE" {
		t.Errorf("expected TRACE, got %s", fsm.CurrentState().Name())
	}
}

// M5a: TRACE state agent loop entry test
func TestTraceStateRunMethod(t *testing.T) {
	fsm := New("test-case", nil, nil)

	// Transition to TRACE state
	if err := fsm.Transition("to trace"); err != nil {
		t.Fatalf("failed to transition to TRACE: %v", err)
	}

	if fsm.CurrentState().Name() != "TRACE" {
		t.Fatalf("expected TRACE state, got %s", fsm.CurrentState().Name())
	}

	// TRACE state should have Run method (defined by State interface extension)
	// This test verifies the state has the correct allowed tools
	traceState := fsm.CurrentState()
	allowedTools := traceState.AllowedTools()

	expectedTools := []string{"timeline.build", "log.query"}
	if len(allowedTools) != len(expectedTools) {
		t.Errorf("expected %d tools, got %d", len(expectedTools), len(allowedTools))
	}

	for _, tool := range expectedTools {
		found := false
		for _, allowed := range allowedTools {
			if allowed == tool {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected tool %s not found in allowed tools", tool)
		}
	}
}

// M5d: Budget enforcement test - LLM turns
func TestIncrementLLMTurns(t *testing.T) {
	fsm := New("test-case", nil, nil)

	// Set low max for testing
	budget := fsm.GetBudgetStatus()
	budget.MaxLLMTurns = 3
	fsm.SetBudgets(budget)

	// First increment should succeed (within budget)
	exceeded := fsm.IncrementLLMTurns()
	if exceeded {
		t.Fatal("first increment should not exceed budget")
	}

	budget = fsm.GetBudgetStatus()
	if budget.CurrentLLMTurns != 1 {
		t.Errorf("expected 1 turn, got %d", budget.CurrentLLMTurns)
	}

	// Increment to max
	fsm.IncrementLLMTurns()
	fsm.IncrementLLMTurns()

	// Now at max
	budget = fsm.GetBudgetStatus()
	if budget.CurrentLLMTurns != 3 {
		t.Errorf("expected 3 turns, got %d", budget.CurrentLLMTurns)
	}

	// Next increment should exceed budget
	exceeded = fsm.IncrementLLMTurns()
	if !exceeded {
		t.Fatal("fourth increment should exceed budget")
	}
}

// M5d: Budget enforcement test - Invalid JSON attempts
func TestIncrementInvalidJSONAttempts(t *testing.T) {
	fsm := New("test-case", nil, nil)

	// Set low max for testing
	budget := fsm.GetBudgetStatus()
	budget.MaxInvalidJSONAttempts = 2
	fsm.SetBudgets(budget)

	// First attempt should succeed
	exceeded := fsm.IncrementInvalidJSONAttempts()
	if exceeded {
		t.Fatal("first attempt should not exceed budget")
	}

	// Second attempt should still be within budget
	exceeded = fsm.IncrementInvalidJSONAttempts()
	if exceeded {
		t.Fatal("second attempt should not exceed budget")
	}

	// Third attempt should exceed budget
	exceeded = fsm.IncrementInvalidJSONAttempts()
	if !exceeded {
		t.Fatal("third attempt should exceed budget")
	}
}
