package hunt

import (
	"testing"
)

func TestGetTransitionRules(t *testing.T) {
	rules := GetTransitionRules()

	if len(rules) != 8 {
		t.Errorf("expected 8 transition rules, got %d", len(rules))
	}

	// Verify all required rules exist
	ruleMap := make(map[string]bool)
	for _, rule := range rules {
		ruleMap[rule.Name] = true
	}

	expectedRules := []string{
		"INITIALIZE_TO_TRACE",
		"TRACE_TO_SCAN",
		"TRACE_TO_LOCK",
		"SCAN_TO_EXPOSE",
		"SCAN_TO_TRACE",
		"SCAN_TO_LOCK",
		"EXPOSE_TO_LOCK",
		"EXPOSE_TO_SCAN",
	}

	for _, expected := range expectedRules {
		if !ruleMap[expected] {
			t.Errorf("missing rule: %s", expected)
		}
	}
}

func TestValidateTransition(t *testing.T) {
	tests := []struct {
		from    string
		to      string
		valid   bool
		name    string
	}{
		{"INITIALIZE", "TRACE", true, "INITIALIZE_TO_TRACE"},
		{"TRACE", "SCAN", true, "TRACE_TO_SCAN"},
		{"TRACE", "LOCK", true, "TRACE_TO_LOCK"},
		{"SCAN", "EXPOSE", true, "SCAN_TO_EXPOSE"},
		{"SCAN", "TRACE", true, "SCAN_TO_TRACE"},
		{"SCAN", "LOCK", true, "SCAN_TO_LOCK"},
		{"EXPOSE", "LOCK", true, "EXPOSE_TO_LOCK"},
		{"EXPOSE", "SCAN", true, "EXPOSE_TO_SCAN"},
		{"INITIALIZE", "SCAN", false, "invalid transition"},
		{"LOCK", "EXPOSE", false, "cannot transition from LOCK"},
	}

	for _, test := range tests {
		valid, guard := ValidateTransition(test.from, test.to)
		if valid != test.valid {
			t.Errorf("%s (%s→%s): expected %v, got %v (guard: %s)", test.name, test.from, test.to, test.valid, valid, guard)
		}
		if test.valid && guard == "" {
			t.Errorf("%s: expected guard condition, got empty", test.name)
		}
	}
}

func TestTransitionRuleProperties(t *testing.T) {
	rules := GetTransitionRules()

	for _, rule := range rules {
		// Each rule must have required fields
		if rule.Name == "" {
			t.Error("rule missing Name")
		}
		if rule.From == "" {
			t.Error("rule missing From")
		}
		if rule.To == "" {
			t.Error("rule missing To")
		}
		if rule.Guard == "" {
			t.Errorf("rule %s missing Guard", rule.Name)
		}
		if rule.Action == "" {
			t.Errorf("rule %s missing Action", rule.Name)
		}

		// MaxCount should be 0 (unlimited) or >= 1
		if rule.MaxCount < 0 {
			t.Errorf("rule %s has invalid MaxCount: %d", rule.Name, rule.MaxCount)
		}

		// Terminal rules should have TerminalStatus set
		if rule.To == "LOCK" && rule.TerminalStatus == "" {
			t.Errorf("rule %s (→LOCK) should have TerminalStatus", rule.Name)
		}
	}
}

func TestGetNextState(t *testing.T) {
	tests := []struct {
		current   string
		expected  string
		expectErr bool
	}{
		{"INITIALIZE", "TRACE", false},
		{"TRACE", "SCAN", false},
		{"SCAN", "EXPOSE", false},
		{"EXPOSE", "LOCK", false},
		{"LOCK", "", true},
		{"UNKNOWN", "", true},
	}

	for _, test := range tests {
		next, err := GetNextState(test.current)

		if test.expectErr && err == nil {
			t.Errorf("GetNextState(%s): expected error, got nil", test.current)
		}

		if !test.expectErr && err != nil {
			t.Errorf("GetNextState(%s): unexpected error: %v", test.current, err)
		}

		if next != test.expected {
			t.Errorf("GetNextState(%s): expected %s, got %s", test.current, test.expected, next)
		}
	}
}

func TestTransitionRuleMaxCounts(t *testing.T) {
	rules := GetTransitionRules()
	ruleMap := make(map[string]int)

	for _, rule := range rules {
		ruleMap[rule.Name] = rule.MaxCount
	}

	// Terminal transitions should have MaxCount=1
	terminalRules := []string{
		"INITIALIZE_TO_TRACE",
		"TRACE_TO_LOCK",
		"SCAN_TO_LOCK",
		"EXPOSE_TO_LOCK",
	}

	for _, ruleName := range terminalRules {
		if maxCount, ok := ruleMap[ruleName]; ok && maxCount != 1 {
			t.Errorf("rule %s should have MaxCount=1, got %d", ruleName, maxCount)
		}
	}

	// Recursive transitions (SCAN↔TRACE, EXPOSE↔SCAN) should have MaxCount=2
	recursiveRules := []string{
		"SCAN_TO_TRACE",
		"EXPOSE_TO_SCAN",
	}

	for _, ruleName := range recursiveRules {
		if maxCount, ok := ruleMap[ruleName]; ok && maxCount != 2 {
			t.Errorf("rule %s should have MaxCount=2, got %d", ruleName, maxCount)
		}
	}

	// Non-terminal, non-recursive should have MaxCount=0 (unlimited)
	unlimitedRules := []string{
		"TRACE_TO_SCAN",
		"SCAN_TO_EXPOSE",
	}

	for _, ruleName := range unlimitedRules {
		if maxCount, ok := ruleMap[ruleName]; ok && maxCount != 0 {
			t.Errorf("rule %s should have MaxCount=0 (unlimited), got %d", ruleName, maxCount)
		}
	}
}
