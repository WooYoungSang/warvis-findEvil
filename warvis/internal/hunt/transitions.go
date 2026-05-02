package hunt

// TransitionRule defines a typed state transition with guards, actions, and constraints.
// Based on architecture spec section 2.2 (typed transition rules).
type TransitionRule struct {
	Name           string
	From           string
	To             string
	Guard          string // human-readable guard condition
	Action         string // action taken on transition
	MaxCount       int    // max times this rule can fire (0 = unlimited)
	TerminalStatus string // terminal status if this leads to case closure
}

// GetTransitionRules returns all 8 transition rules as per architecture spec.
func GetTransitionRules() []TransitionRule {
	return []TransitionRule{
		{
			Name:           "INITIALIZE_TO_TRACE",
			From:           "INITIALIZE",
			To:             "TRACE",
			Guard:          "case_sandbox_created AND evidence_registered",
			Action:         "load_timeline_sources",
			MaxCount:       1,
			TerminalStatus: "none",
		},
		{
			Name:           "TRACE_TO_SCAN",
			From:           "TRACE",
			To:             "SCAN",
			Guard:          "success_criteria_met AND (complete OR inconclusive OR degraded)",
			Action:         "prepare_ioc_scan",
			MaxCount:       0, // unlimited
			TerminalStatus: "none",
		},
		{
			Name:           "TRACE_TO_LOCK",
			From:           "TRACE",
			To:             "LOCK",
			Guard:          "failed_status OR max_retries_exceeded OR timeout",
			Action:         "finalize_case_with_error",
			MaxCount:       1,
			TerminalStatus: "failed",
		},
		{
			Name:           "SCAN_TO_EXPOSE",
			From:           "SCAN",
			To:             "EXPOSE",
			Guard:          "(complete OR inconclusive OR degraded) AND findings_exist",
			Action:         "prepare_verification",
			MaxCount:       0, // unlimited
			TerminalStatus: "none",
		},
		{
			Name:           "SCAN_TO_TRACE",
			From:           "SCAN",
			To:             "TRACE",
			Guard:          "gemma_requests_deeper_timeline AND attempt_count < 2",
			Action:         "reset_timeline_with_memory_focus",
			MaxCount:       2,
			TerminalStatus: "degraded",
		},
		{
			Name:           "SCAN_TO_LOCK",
			From:           "SCAN",
			To:             "LOCK",
			Guard:          "failed_status OR max_tool_failures OR budget_exceeded",
			Action:         "finalize_case_with_partial_findings",
			MaxCount:       1,
			TerminalStatus: "failed",
		},
		{
			Name:           "EXPOSE_TO_LOCK",
			From:           "EXPOSE",
			To:             "LOCK",
			Guard:          "(complete OR inconclusive OR degraded OR failed) AND (all_findings_processed OR no_findings)",
			Action:         "generate_final_report",
			MaxCount:       1,
			TerminalStatus: "(mirrors EXPOSE tier: complete|inconclusive|degraded|failed)",
		},
		{
			Name:           "EXPOSE_TO_SCAN",
			From:           "EXPOSE",
			To:             "SCAN",
			Guard:          "gemma_requests_additional_scans AND attempt_count < 2",
			Action:         "extend_scan_scope",
			MaxCount:       2,
			TerminalStatus: "degraded",
		},
	}
}

// ValidateTransition checks if a transition is valid based on the current state.
func ValidateTransition(fromState string, toState string) (bool, string) {
	rules := GetTransitionRules()
	for _, rule := range rules {
		if rule.From == fromState && rule.To == toState {
			return true, rule.Guard
		}
	}
	return false, ""
}

// GetNextState determines the next state based on current state and reason.
// This is a simplified version; full implementation would evaluate guards.
func GetNextState(currentState string) (string, error) {
	switch currentState {
	case "INITIALIZE":
		return "TRACE", nil
	case "TRACE":
		return "SCAN", nil
	case "SCAN":
		return "EXPOSE", nil
	case "EXPOSE":
		return "LOCK", nil
	case "LOCK":
		return "", ErrTerminalState()
	default:
		return "", ErrUnknownState(currentState)
	}
}

// ErrTerminalState returns an error for LOCK state (terminal).
func ErrTerminalState() error {
	return &TransitionError{Message: "cannot transition from LOCK state (terminal)"}
}

// ErrUnknownState returns an error for unknown state.
func ErrUnknownState(state string) error {
	return &TransitionError{Message: "unknown state: " + state}
}

// TransitionError represents a transition-related error.
type TransitionError struct {
	Message string
}

func (e *TransitionError) Error() string {
	return e.Message
}
