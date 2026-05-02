package hunt

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRegistrySaveAndLoad(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "hunt-registry-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	registry := NewRegistry(tmpDir)
	caseID := "test-case-001"

	// Create initial state
	state := &InitializeState{
		CaseID:       caseID,
		SandboxRoot:  "/cases/test-case-001/",
		EvidenceKind: "disk_image",
		EvidencePath: "/evidence/test.img",
	}

	budgets := BudgetStatus{
		MaxLLMTurns:             50,
		CurrentLLMTurns:         5,
		MaxInvalidJSONAttempts:  10,
		CurrentInvalidJSONAttempts: 2,
	}

	// Save state
	if err := registry.Save(caseID, state, budgets); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	// Verify file was created
	stateFile := filepath.Join(tmpDir, caseID, "state.json")
	if _, err := os.Stat(stateFile); err != nil {
		t.Fatalf("state.json not created: %v", err)
	}

	// Load state
	record, err := registry.Load(caseID)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if record == nil {
		t.Fatal("loaded record is nil")
	}

	if record.CaseID != caseID {
		t.Errorf("expected CaseID %s, got %s", caseID, record.CaseID)
	}

	if record.CurrentState != "INITIALIZE" {
		t.Errorf("expected INITIALIZE state, got %s", record.CurrentState)
	}

	if record.Budgets.CurrentLLMTurns != 5 {
		t.Errorf("expected CurrentLLMTurns=5, got %d", record.Budgets.CurrentLLMTurns)
	}
}

func TestRegistryBudgetsPersist(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hunt-registry-budgets")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	registry := NewRegistry(tmpDir)
	caseID := "budget-test-case"

	// Save with some budget consumed
	state := &TraceState{CaseID: caseID}
	budgets := BudgetStatus{
		MaxLLMTurns:             50,
		CurrentLLMTurns:         30,
		MaxInvalidJSONAttempts:  10,
		CurrentInvalidJSONAttempts: 7,
	}

	if err := registry.Save(caseID, state, budgets); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	// Load and verify budgets are NOT reset
	record, err := registry.Load(caseID)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if record.Budgets.CurrentLLMTurns != 30 {
		t.Errorf("CurrentLLMTurns should persist: expected 30, got %d", record.Budgets.CurrentLLMTurns)
	}

	if record.Budgets.CurrentInvalidJSONAttempts != 7 {
		t.Errorf("CurrentInvalidJSONAttempts should persist: expected 7, got %d", record.Budgets.CurrentInvalidJSONAttempts)
	}
}

func TestRegistryLoadNonexistent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hunt-registry-nonexist")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	registry := NewRegistry(tmpDir)
	record, err := registry.Load("nonexistent-case")

	if err != nil {
		t.Fatalf("load should not error for nonexistent case: %v", err)
	}

	if record != nil {
		t.Fatal("load should return nil for nonexistent case")
	}
}

func TestRegistryLoadState(t *testing.T) {
	registry := NewRegistry("/tmp")

	// Test loading each state type
	tests := []struct {
		name          string
		record        *CaseRecord
		expectedState string
	}{
		{
			name: "INITIALIZE",
			record: &CaseRecord{
				CurrentState:    "INITIALIZE",
				InitializeState: &InitializeState{CaseID: "test"},
			},
			expectedState: "INITIALIZE",
		},
		{
			name: "TRACE",
			record: &CaseRecord{
				CurrentState: "TRACE",
				TraceState:   &TraceState{CaseID: "test"},
			},
			expectedState: "TRACE",
		},
		{
			name: "SCAN",
			record: &CaseRecord{
				CurrentState: "SCAN",
				ScanState:    &ScanState{CaseID: "test"},
			},
			expectedState: "SCAN",
		},
		{
			name: "EXPOSE",
			record: &CaseRecord{
				CurrentState: "EXPOSE",
				ExposeState:  &ExposeState{CaseID: "test"},
			},
			expectedState: "EXPOSE",
		},
		{
			name: "LOCK",
			record: &CaseRecord{
				CurrentState: "LOCK",
				LockState:    &LockState{CaseID: "test"},
			},
			expectedState: "LOCK",
		},
	}

	for _, test := range tests {
		state := registry.LoadState(test.record)
		if state == nil {
			t.Errorf("%s: LoadState returned nil", test.name)
			continue
		}
		if state.Name() != test.expectedState {
			t.Errorf("%s: expected %s, got %s", test.name, test.expectedState, state.Name())
		}
	}
}

func TestRegistryAuditLogPath(t *testing.T) {
	registry := NewRegistry("/cases")
	caseID := "test-case-123"

	auditPath := registry.GetAuditLogPath(caseID)
	expected := filepath.Join("/cases", caseID, "audit.jsonl")

	if auditPath != expected {
		t.Errorf("expected %s, got %s", expected, auditPath)
	}
}

func TestRegistryStateJSONStructure(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hunt-registry-json")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	registry := NewRegistry(tmpDir)
	caseID := "json-test-case"

	state := &ScanState{
		CaseID:      caseID,
		Matches:     []Finding{},
		Processes:   []Process{},
		MalfindsVAD: []VADRegion{},
		NetFlows:    []NetworkFlow{},
		ToolsRun:    []string{"iocs.scan"},
	}

	budgets := BudgetStatus{
		MaxLLMTurns: 50,
	}

	if err := registry.Save(caseID, state, budgets); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	// Read raw JSON and verify structure
	stateFile := filepath.Join(tmpDir, caseID, "state.json")
	data, err := os.ReadFile(stateFile)
	if err != nil {
		t.Fatalf("failed to read state.json: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	// Verify required fields
	requiredFields := []string{"case_id", "current_state", "created_at", "updated_at", "budgets"}
	for _, field := range requiredFields {
		if _, ok := raw[field]; !ok {
			t.Errorf("missing required field: %s", field)
		}
	}

	// Verify timestamps are RFC3339 format
	if createdAt, ok := raw["created_at"].(string); ok {
		if _, err := time.Parse(time.RFC3339, createdAt); err != nil {
			t.Errorf("created_at is not RFC3339 format: %s", createdAt)
		}
	}
}
