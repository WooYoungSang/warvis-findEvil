package agent

import (
	"strings"
	"testing"

	"github.com/woopsfactory/warvis/internal/hunt"
)

func TestBuildSystemPromptContainsCaseID(t *testing.T) {
	state := &hunt.TraceState{CaseID: "test-case-uuid-abc"}
	tools := []ToolInfo{{Name: "timeline.build", Description: "x"}}
	prompt := BuildSystemPrompt(state, tools, "test-case-uuid-abc")
	if !strings.Contains(prompt, "test-case-uuid-abc") {
		t.Error("prompt must include case_id verbatim so Gemma can reuse it in tool calls")
	}
	if !strings.Contains(strings.ToLower(prompt), "case_id") {
		t.Error("prompt must mention case_id parameter explicitly")
	}
}

func TestBuildSystemPromptContainsState(t *testing.T) {
	state := &hunt.TraceState{}
	tools := []ToolInfo{
		{
			Name:        "timeline.build",
			Description: "Build forensic timeline",
		},
	}

	prompt := BuildSystemPrompt(state, tools, "")

	if !strings.Contains(prompt, "TRACE") {
		t.Error("prompt should contain state name")
	}
	if !strings.Contains(prompt, "timeline.build") {
		t.Error("prompt should contain tool name")
	}
}

func TestBuildSystemPromptContainsFormatSpec(t *testing.T) {
	state := &hunt.TraceState{}
	tools := []ToolInfo{}

	prompt := BuildSystemPrompt(state, tools, "")

	if !strings.Contains(prompt, `"action"`) {
		t.Error("prompt should contain action field spec")
	}
	if !strings.Contains(prompt, `"tool_name"`) {
		t.Error("prompt should contain tool_name field spec")
	}
	if !strings.Contains(prompt, `"arguments"`) {
		t.Error("prompt should contain arguments field spec")
	}
}

func TestBuildSystemPromptMultipleTools(t *testing.T) {
	state := &hunt.ScanState{}
	tools := []ToolInfo{
		{
			Name:        "iocs.scan",
			Description: "Scan for indicators of compromise",
			Parameters: map[string]interface{}{
				"pattern": "string",
			},
		},
		{
			Name:        "memory.process_list",
			Description: "Extract process list",
			Parameters: map[string]interface{}{
				"pid": "integer",
			},
		},
	}

	prompt := BuildSystemPrompt(state, tools, "")

	if !strings.Contains(prompt, "iocs.scan") {
		t.Error("prompt should contain iocs.scan")
	}
	if !strings.Contains(prompt, "memory.process_list") {
		t.Error("prompt should contain memory.process_list")
	}
	if !strings.Contains(prompt, "Scan for indicators of compromise") {
		t.Error("prompt should contain tool description")
	}
}

func TestBuildSystemPromptStateDescriptions(t *testing.T) {
	states := []struct {
		state hunt.State
		key   string
	}{
		{&hunt.InitializeState{}, "case.open"},
		{&hunt.TraceState{}, "timeline.build"},
		{&hunt.ScanState{}, "iocs.scan"},
		{&hunt.ExposeState{}, "verify.cross_check"},
		{&hunt.LockState{}, "terminal state"},
	}

	for _, tt := range states {
		prompt := BuildSystemPrompt(tt.state, []ToolInfo{}, "")
		if !strings.Contains(prompt, tt.key) {
			t.Errorf("state %s prompt should mention %s", tt.state.Name(), tt.key)
		}
	}
}

func TestBuildSystemPromptActionSemantics(t *testing.T) {
	state := &hunt.TraceState{}
	tools := []ToolInfo{}

	prompt := BuildSystemPrompt(state, tools, "")

	if !strings.Contains(prompt, "call_tool") {
		t.Error("prompt should explain call_tool action")
	}
	if !strings.Contains(prompt, "state_complete") {
		t.Error("prompt should explain state_complete action")
	}
	if !strings.Contains(prompt, "escalate") {
		t.Error("prompt should explain escalate action")
	}
}

func TestBuildSystemPromptLength(t *testing.T) {
	state := &hunt.TraceState{}
	tools := []ToolInfo{
		{Name: "timeline.build", Description: "Build timeline"},
	}

	prompt := BuildSystemPrompt(state, tools, "")

	// Prompt should be substantial (at least 500 chars)
	if len(prompt) < 500 {
		t.Errorf("prompt too short: %d chars", len(prompt))
	}
}
