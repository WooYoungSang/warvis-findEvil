package hunt

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// CaseRecord represents the persisted state of a case in state.json.
type CaseRecord struct {
	CaseID            string        `json:"case_id"`
	CurrentState      string        `json:"current_state"`
	CreatedAt         string        `json:"created_at"`
	UpdatedAt         string        `json:"updated_at"`
	Budgets           BudgetStatus  `json:"budgets"`
	InitializeState   *InitializeState `json:"initialize_state,omitempty"`
	TraceState        *TraceState      `json:"trace_state,omitempty"`
	ScanState         *ScanState       `json:"scan_state,omitempty"`
	ExposeState       *ExposeState     `json:"expose_state,omitempty"`
	LockState         *LockState       `json:"lock_state,omitempty"`
}

// Registry provides persistence for Hunt FSM state.
type Registry struct {
	casesRootDir string
}

// NewRegistry creates a new Registry with the given cases root directory.
func NewRegistry(casesRootDir string) *Registry {
	return &Registry{
		casesRootDir: casesRootDir,
	}
}

// Save persists the FSM state to /cases/<case_id>/state.json.
func (r *Registry) Save(caseID string, state State, budgets BudgetStatus) error {
	caseDir := filepath.Join(r.casesRootDir, caseID)

	// Create directory if it doesn't exist
	if err := os.MkdirAll(caseDir, 0755); err != nil {
		return fmt.Errorf("failed to create case directory: %w", err)
	}

	// Build CaseRecord
	record := CaseRecord{
		CaseID:       caseID,
		CurrentState: state.Name(),
		UpdatedAt:    time.Now().UTC().Format(time.RFC3339),
		Budgets:      budgets,
	}

	// If CreatedAt not set, set it now
	existingState, _ := r.Load(caseID)
	if existingState != nil && existingState.CreatedAt != "" {
		record.CreatedAt = existingState.CreatedAt
	} else {
		record.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}

	// Store state-specific data
	switch s := state.(type) {
	case *InitializeState:
		record.InitializeState = s
	case *TraceState:
		record.TraceState = s
	case *ScanState:
		record.ScanState = s
	case *ExposeState:
		record.ExposeState = s
	case *LockState:
		record.LockState = s
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// Write to file
	stateFilePath := filepath.Join(caseDir, "state.json")
	if err := os.WriteFile(stateFilePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write state.json: %w", err)
	}

	return nil
}

// Load restores FSM state from /cases/<case_id>/state.json.
// Returns nil, nil if file doesn't exist (new case).
// Budgets are NOT reset on load (critical for resume behavior).
func (r *Registry) Load(caseID string) (*CaseRecord, error) {
	stateFilePath := filepath.Join(r.casesRootDir, caseID, "state.json")

	data, err := os.ReadFile(stateFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // New case
		}
		return nil, fmt.Errorf("failed to read state.json: %w", err)
	}

	var record CaseRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return nil, fmt.Errorf("failed to parse state.json: %w", err)
	}

	return &record, nil
}

// LoadState restores the State interface from CaseRecord.
func (r *Registry) LoadState(record *CaseRecord) State {
	if record == nil {
		return nil
	}

	switch record.CurrentState {
	case "INITIALIZE":
		if record.InitializeState != nil {
			return record.InitializeState
		}
	case "TRACE":
		if record.TraceState != nil {
			return record.TraceState
		}
	case "SCAN":
		if record.ScanState != nil {
			return record.ScanState
		}
	case "EXPOSE":
		if record.ExposeState != nil {
			return record.ExposeState
		}
	case "LOCK":
		if record.LockState != nil {
			return record.LockState
		}
	}

	return nil
}

// GetAuditLogPath returns the path to the audit log for a case.
func (r *Registry) GetAuditLogPath(caseID string) string {
	return filepath.Join(r.casesRootDir, caseID, "audit.jsonl")
}
