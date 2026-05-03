package agent

import (
	"encoding/json"
	"testing"
)

func TestParseActionCallTool(t *testing.T) {
	response := `{"action": "call_tool", "tool_name": "timeline.build", "arguments": {"limit": 100}, "reason": "Testing"}`
	action := ParseAction(response)

	if action.Type != "call_tool" {
		t.Fatalf("expected call_tool, got %s", action.Type)
	}
	if action.ToolName != "timeline.build" {
		t.Fatalf("expected timeline.build, got %s", action.ToolName)
	}
	if action.Arguments == nil {
		t.Fatal("arguments should not be nil")
	}
}

func TestParseActionStateComplete(t *testing.T) {
	response := `{"action": "state_complete", "reason": "Found enough evidence"}`
	action := ParseAction(response)

	if action.Type != "state_complete" {
		t.Fatalf("expected state_complete, got %s", action.Type)
	}
}

func TestParseActionEscalate(t *testing.T) {
	response := `{"action": "escalate", "reason": "Need expert intervention"}`
	action := ParseAction(response)

	if action.Type != "escalate" {
		t.Fatalf("expected escalate, got %s", action.Type)
	}
}

func TestParseActionMarkdownWrapped(t *testing.T) {
	response := "Here's the action I decided on:\n\n```json\n" +
		`{"action": "call_tool", "tool_name": "iocs.scan", "arguments": {}}` +
		"\n```\n\nThis will scan for indicators."
	action := ParseAction(response)

	if action.Type != "call_tool" {
		t.Fatalf("expected call_tool, got %s", action.Type)
	}
	if action.ToolName != "iocs.scan" {
		t.Fatalf("expected iocs.scan, got %s", action.ToolName)
	}
}

func TestParseActionEmbeddedInText(t *testing.T) {
	response := `I think I should call this tool:

{"action": "call_tool", "tool_name": "memory.dump", "arguments": {"pid": 1234}}

This will help identify malware.`
	action := ParseAction(response)

	if action.Type != "call_tool" {
		t.Fatalf("expected call_tool, got %s", action.Type)
	}
	if action.ToolName != "memory.dump" {
		t.Fatalf("expected memory.dump, got %s", action.ToolName)
	}
}

func TestParseActionMalformedJSON(t *testing.T) {
	response := `{"action": "call_tool", "tool_name": "timeline.build"`
	action := ParseAction(response)

	// Should escalate after 3 retries
	if action.Type != "escalate" {
		t.Fatalf("expected escalate on malformed JSON, got %s", action.Type)
	}
}

func TestParseActionInvalidActionType(t *testing.T) {
	response := `{"action": "invalid_action", "reason": "Test"}`
	action := ParseAction(response)

	// Should escalate on invalid action type
	if action.Type != "escalate" {
		t.Fatalf("expected escalate on invalid type, got %s", action.Type)
	}
}

func TestParseActionMissingToolName(t *testing.T) {
	response := `{"action": "call_tool", "arguments": {}, "reason": "Test"}`
	action := ParseAction(response)

	// Should escalate because call_tool requires tool_name
	if action.Type != "escalate" {
		t.Fatalf("expected escalate on missing tool_name, got %s", action.Type)
	}
}

func TestParseActionWithNestedJSON(t *testing.T) {
	response := `{"action": "call_tool", "tool_name": "memory.dump", "arguments": {"nested": {"key": "value"}}, "reason": "Test"}`
	action := ParseAction(response)

	if action.Type != "call_tool" {
		t.Fatalf("expected call_tool, got %s", action.Type)
	}
	if action.ToolName != "memory.dump" {
		t.Fatalf("expected memory.dump, got %s", action.ToolName)
	}

	// Verify nested structure
	if nested, ok := action.Arguments["nested"].(map[string]interface{}); !ok || nested["key"] != "value" {
		t.Fatal("nested arguments not parsed correctly")
	}
}

func TestExtractJSONFromMarkdownCodeBlock(t *testing.T) {
	response := "```json\n" + `{"test": "value"}` + "\n```"
	extracted := extractJSONFromMarkdown(response)

	if extracted != `{"test": "value"}` {
		t.Fatalf("expected to extract JSON, got: %s", extracted)
	}
}

func TestExtractJSONFromMarkdownWithoutLanguage(t *testing.T) {
	response := "```\n" + `{"test": "value"}` + "\n```"
	extracted := extractJSONFromMarkdown(response)

	if extracted != `{"test": "value"}` {
		t.Fatalf("expected to extract JSON, got: %s", extracted)
	}
}

func TestExtractJSONFromText(t *testing.T) {
	response := `Some text before {"action": "call_tool"} and more text after`
	extracted := extractJSONFromText(response)

	if extracted != `{"action": "call_tool"}` {
		t.Fatalf("expected to extract JSON, got: %s", extracted)
	}
}

func TestExtractJSONFromTextNestedBraces(t *testing.T) {
	response := `Start {"outer": {"inner": "value"}} end`
	extracted := extractJSONFromText(response)

	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(extracted), &obj); err != nil {
		t.Fatalf("extracted JSON is invalid: %v", err)
	}

	if outer, ok := obj["outer"].(map[string]interface{}); !ok || outer["inner"] != "value" {
		t.Fatal("nested structure not extracted correctly")
	}
}

func TestValidateActionCallTool(t *testing.T) {
	action := &Action{
		Type:      "call_tool",
		ToolName:  "timeline.build",
		Arguments: map[string]interface{}{},
	}
	err := validateAction(action)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestValidateActionCallToolMissingToolName(t *testing.T) {
	action := &Action{
		Type:      "call_tool",
		Arguments: map[string]interface{}{},
	}
	err := validateAction(action)
	if err == nil {
		t.Fatal("expected error for missing tool_name")
	}
}

func TestValidateActionInvalidType(t *testing.T) {
	action := &Action{
		Type: "unknown",
	}
	err := validateAction(action)
	if err == nil {
		t.Fatal("expected error for invalid type")
	}
}

func TestSanitizeToolOutput(t *testing.T) {
	output := "This is a very long output that should be truncated"
	sanitized := SanitizeToolOutput(output, 20)

	// 20 chars + " [truncated]" = ~33 chars
	if len(sanitized) > 40 {
		t.Fatalf("output not properly truncated: %d chars", len(sanitized))
	}
	if !contains(sanitized, "[truncated]") {
		t.Fatal("expected [truncated] marker")
	}
}

func TestSafeEnvelopeToolOutput(t *testing.T) {
	output := "some forensic data"
	enveloped := SafeEnvelopeToolOutput("timeline.build", output)

	if !contains(enveloped, "untrusted source") {
		t.Fatal("expected safety prefix")
	}
	if !contains(enveloped, "timeline.build") {
		t.Fatal("expected tool name in envelope")
	}
	if !contains(enveloped, output) {
		t.Fatal("expected output in envelope")
	}
}

// Helper function
func contains(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
