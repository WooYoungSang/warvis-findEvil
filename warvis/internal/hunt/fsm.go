package hunt

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// BudgetStatus tracks resource consumption per state.
// Budgets persist across resume and are NOT reset.
type BudgetStatus struct {
	MaxLLMTurns                int
	CurrentLLMTurns            int
	MaxInvalidJSONAttempts     int
	CurrentInvalidJSONAttempts int
	MaxDuplicateToolCalls      int
	CurrentDuplicateToolCalls  int
	MaxToolCallsTotal          int
	CurrentToolCallsTotal      int
	MaxStateDurationSeconds    int
	StateStartedAt             string // ISO 8601
}

// FSM is the Hunt State Machine interface.
type FSM interface {
	// CurrentState returns the current Hunt state.
	CurrentState() State
	// IsToolAllowed checks if a tool is available in the current state.
	IsToolAllowed(toolName string) bool
	// CallTool invokes a tool via MCP and returns result as JSON string.
	// This is a stub for M2a; actual MCP integration happens later.
	CallTool(ctx context.Context, toolName string, args map[string]interface{}) (string, error)
	// Transition moves the FSM to the next state based on provided reason.
	Transition(reason string) error
	// Pause temporarily halts the hunt for later resume.
	Pause(reason string) error
	// Resume continues a paused hunt from prior state.
	Resume() error
	// GetBudgetStatus returns current resource consumption.
	GetBudgetStatus() BudgetStatus
	// SetBudgets overrides the current budget values (e.g. for testing).
	SetBudgets(budgets BudgetStatus)
	// IncrementLLMTurns increments the LLM turn counter and returns true if the budget is exceeded.
	IncrementLLMTurns() bool
}

// HuntFSM is the concrete implementation of FSM.
type HuntFSM struct {
	mu sync.RWMutex

	// Current state (one of 5 state types)
	currentState State

	// Case context
	caseID string

	// Budget tracking (persisted in state.json)
	budgets BudgetStatus

	// Pause state
	paused bool
	pauseReason string

	// Audit log
	auditLog *AuditLog

	// Registry for persistence
	registry *Registry

	// Tool registry (from MCP)
	toolRegistry map[string]bool // tool name -> allowed in current state
}

// New creates a new Hunt FSM.
func New(caseID string, auditLog *AuditLog, registry *Registry) *HuntFSM {
	return &HuntFSM{
		caseID:       caseID,
		currentState: &InitializeState{CaseID: caseID},
		auditLog:     auditLog,
		registry:     registry,
		toolRegistry: make(map[string]bool),
		budgets: BudgetStatus{
			MaxLLMTurns:             50,
			MaxInvalidJSONAttempts:  10,
			MaxDuplicateToolCalls:   5,
			MaxToolCallsTotal:       100,
			MaxStateDurationSeconds: 600,
			StateStartedAt:          time.Now().UTC().Format(time.RFC3339),
		},
	}
}

// CurrentState returns the current FSM state.
func (f *HuntFSM) CurrentState() State {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.currentState
}

// IsToolAllowed checks if a tool is available in the current state.
func (f *HuntFSM) IsToolAllowed(toolName string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if f.currentState == nil {
		return false
	}

	allowedTools := f.currentState.AllowedTools()
	for _, tool := range allowedTools {
		if tool == toolName {
			return true
		}
	}
	return false
}

// CallTool invokes a tool via MCP.
// For M2a, this is a stub; actual MCP integration happens in later phase.
func (f *HuntFSM) CallTool(ctx context.Context, toolName string, args map[string]interface{}) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Validate tool is allowed
	allowed := false
	for _, tool := range f.currentState.AllowedTools() {
		if tool == toolName {
			allowed = true
			break
		}
	}
	if !allowed {
		return "", fmt.Errorf("tool %s not allowed in %s state", toolName, f.currentState.Name())
	}

	// Increment tool call counter
	f.budgets.CurrentToolCallsTotal++

	// Stub: return placeholder result
	return fmt.Sprintf(`{"tool":"%s","status":"success","message":"stub result"}`, toolName), nil
}

// Transition moves the FSM to the next state.
func (f *HuntFSM) Transition(reason string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.paused {
		return fmt.Errorf("cannot transition while paused: %s", f.pauseReason)
	}

	currentStateName := f.currentState.Name()

	// Dispatch to transition rules
	switch currentStateName {
	case "INITIALIZE":
		// INITIALIZE → TRACE
		if f.currentState == nil {
			return fmt.Errorf("invalid state")
		}
		if initState, ok := f.currentState.(*InitializeState); ok {
			traceState := &TraceState{
				CaseID:       initState.CaseID,
				TimelinePath: fmt.Sprintf("/cases/%s/timeline.jsonl", initState.CaseID),
				SourcesFound: []string{},
				QueryResults: []interface{}{},
			}
			f.currentState = traceState
			f.budgets.StateStartedAt = time.Now().UTC().Format(time.RFC3339)
			if f.auditLog != nil {
				_ = f.auditLog.Append(map[string]interface{}{
					"timestamp": time.Now().UTC().Format(time.RFC3339),
					"event":     "state_transition",
					"from":      "INITIALIZE",
					"to":        "TRACE",
					"reason":    reason,
				})
			}
			return nil
		}
		return fmt.Errorf("invalid INITIALIZE state type")

	case "TRACE":
		// TRACE → SCAN
		if traceState, ok := f.currentState.(*TraceState); ok {
			scanState := &ScanState{
				CaseID:   traceState.CaseID,
				Matches:  []Finding{},
				Processes: []Process{},
				MalfindsVAD: []VADRegion{},
				NetFlows: []NetworkFlow{},
				ToolsRun: []string{},
			}
			f.currentState = scanState
			f.budgets.StateStartedAt = time.Now().UTC().Format(time.RFC3339)
			if f.auditLog != nil {
				_ = f.auditLog.Append(map[string]interface{}{
					"timestamp": time.Now().UTC().Format(time.RFC3339),
					"event":     "state_transition",
					"from":      "TRACE",
					"to":        "SCAN",
					"reason":    reason,
				})
			}
			return nil
		}
		return fmt.Errorf("invalid TRACE state type")

	case "SCAN":
		// SCAN → EXPOSE (primary) or SCAN → TRACE (optional deepening)
		if scanState, ok := f.currentState.(*ScanState); ok {
			exposeState := &ExposeState{
				CaseID:   scanState.CaseID,
				Findings: []VerifiedFinding{},
				ReportPath: fmt.Sprintf("/cases/%s/report/findings.jsonl", scanState.CaseID),
			}
			f.currentState = exposeState
			f.budgets.StateStartedAt = time.Now().UTC().Format(time.RFC3339)
			if f.auditLog != nil {
				_ = f.auditLog.Append(map[string]interface{}{
					"timestamp": time.Now().UTC().Format(time.RFC3339),
					"event":     "state_transition",
					"from":      "SCAN",
					"to":        "EXPOSE",
					"reason":    reason,
				})
			}
			return nil
		}
		return fmt.Errorf("invalid SCAN state type")

	case "EXPOSE":
		// EXPOSE → LOCK (final)
		if exposeState, ok := f.currentState.(*ExposeState); ok {
			lockState := &LockState{
				CaseID:     exposeState.CaseID,
				Status:     "CLOSED",
				ReportPath: exposeState.ReportPath,
				ClosedAt:   time.Now().UTC().Format(time.RFC3339),
			}
			f.currentState = lockState
			if f.auditLog != nil {
				_ = f.auditLog.Append(map[string]interface{}{
					"timestamp": time.Now().UTC().Format(time.RFC3339),
					"event":     "state_transition",
					"from":      "EXPOSE",
					"to":        "LOCK",
					"reason":    reason,
				})
			}
			return nil
		}
		return fmt.Errorf("invalid EXPOSE state type")

	case "LOCK":
		return fmt.Errorf("cannot transition from LOCK state (terminal)")

	default:
		return fmt.Errorf("unknown state: %s", currentStateName)
	}
}

// Pause temporarily halts the hunt.
func (f *HuntFSM) Pause(reason string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.paused = true
	f.pauseReason = reason

	if f.auditLog != nil {
		_ = f.auditLog.Append(map[string]interface{}{
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"event":     "hunt_paused",
			"reason":    reason,
			"state":     f.currentState.Name(),
		})
	}

	return nil
}

// Resume continues a paused hunt.
func (f *HuntFSM) Resume() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if !f.paused {
		return fmt.Errorf("hunt is not paused")
	}

	f.paused = false
	prior := f.pauseReason
	f.pauseReason = ""

	if f.auditLog != nil {
		_ = f.auditLog.Append(map[string]interface{}{
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"event":     "hunt_resumed",
			"prior_pause_reason": prior,
			"state":     f.currentState.Name(),
		})
	}

	return nil
}

// GetBudgetStatus returns current resource consumption.
func (f *HuntFSM) GetBudgetStatus() BudgetStatus {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.budgets
}

// SetBudgets allows loading budgets from persistence (for resume).
func (f *HuntFSM) SetBudgets(budgets BudgetStatus) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.budgets = budgets
}

// SetCurrentState allows loading state from persistence (for resume).
func (f *HuntFSM) SetCurrentState(state State) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.currentState = state
}

// IncrementLLMTurns increments the LLM turn counter.
// Returns true if budget exceeded, false otherwise.
func (f *HuntFSM) IncrementLLMTurns() bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.budgets.CurrentLLMTurns++
	if f.budgets.CurrentLLMTurns > f.budgets.MaxLLMTurns {
		if f.auditLog != nil {
			_ = f.auditLog.Append(map[string]interface{}{
				"timestamp":        time.Now().UTC().Format(time.RFC3339),
				"event":            "budget_exceeded",
				"budget_type":      "llm_turns",
				"current":          f.budgets.CurrentLLMTurns,
				"max":              f.budgets.MaxLLMTurns,
				"state":            f.currentState.Name(),
			})
		}
		return true
	}
	return false
}

// IncrementInvalidJSONAttempts increments the invalid JSON counter.
// Returns true if budget exceeded, false otherwise.
func (f *HuntFSM) IncrementInvalidJSONAttempts() bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.budgets.CurrentInvalidJSONAttempts++
	if f.budgets.CurrentInvalidJSONAttempts > f.budgets.MaxInvalidJSONAttempts {
		if f.auditLog != nil {
			_ = f.auditLog.Append(map[string]interface{}{
				"timestamp":        time.Now().UTC().Format(time.RFC3339),
				"event":            "budget_exceeded",
				"budget_type":      "invalid_json_attempts",
				"current":          f.budgets.CurrentInvalidJSONAttempts,
				"max":              f.budgets.MaxInvalidJSONAttempts,
				"state":            f.currentState.Name(),
			})
		}
		return true
	}
	return false
}
