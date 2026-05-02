package hunt

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// AuditLog provides append-only audit trail functionality.
// All entries are JSON-Lines format (one JSON object per line).
// File is opened in O_APPEND mode to enforce append-only semantics.
type AuditLog struct {
	file      *os.File
	mu        sync.Mutex
	priorHash string // hash of previous entry for chain-of-custody
}

// NewAuditLog creates or opens an audit log file at the given path.
// File is opened in append mode only (O_APPEND).
func NewAuditLog(filepath string) (*AuditLog, error) {
	// Open file in O_APPEND mode (append-only semantics)
	file, err := os.OpenFile(
		filepath,
		os.O_CREATE|os.O_APPEND|os.O_WRONLY,
		0644,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to open audit log: %w", err)
	}

	al := &AuditLog{
		file:      file,
		priorHash: "",
	}

	return al, nil
}

// Append adds a new entry to the audit log.
// Entry can be any JSON-serializable object.
// Each entry is augmented with timestamp and prior hash for chain-of-custody.
func (al *AuditLog) Append(entry interface{}) error {
	al.mu.Lock()
	defer al.mu.Unlock()

	// Convert entry to map if it's not already
	var entryMap map[string]interface{}
	switch v := entry.(type) {
	case map[string]interface{}:
		entryMap = v
	default:
		// Try to marshal/unmarshal to get map form
		data, _ := json.Marshal(v)
		_ = json.Unmarshal(data, &entryMap)
		if entryMap == nil {
			entryMap = make(map[string]interface{})
		}
	}

	// Add prior hash to chain-of-custody
	if al.priorHash != "" {
		entryMap["prior_hash"] = al.priorHash
	}

	// Compute entry hash before serialization
	entryJSON, err := json.Marshal(entryMap)
	if err != nil {
		return fmt.Errorf("failed to marshal entry: %w", err)
	}

	hash := sha256.Sum256(entryJSON)
	al.priorHash = "sha256:" + hex.EncodeToString(hash[:])
	entryMap["entry_hash"] = al.priorHash

	// Re-serialize with hash included
	finalJSON, err := json.Marshal(entryMap)
	if err != nil {
		return fmt.Errorf("failed to marshal entry with hash: %w", err)
	}

	// Write to file with newline
	_, err = al.file.Write(append(finalJSON, '\n'))
	if err != nil {
		return fmt.Errorf("failed to write audit log entry: %w", err)
	}

	return nil
}

// Close closes the audit log file.
func (al *AuditLog) Close() error {
	al.mu.Lock()
	defer al.mu.Unlock()

	if al.file != nil {
		return al.file.Close()
	}
	return nil
}

// Verify reads and validates the audit log chain.
// Returns an error if any entry fails hash chain validation.
func (al *AuditLog) Verify(filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("failed to open audit log for verification: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	priorHash := ""

	for {
		var entry map[string]interface{}
		err := decoder.Decode(&entry)
		if err != nil {
			// EOF is expected
			if err.Error() == "EOF" {
				break
			}
			return fmt.Errorf("failed to parse audit log entry: %w", err)
		}

		// Verify prior hash matches
		if priorHash != "" {
			if entryPriorHash, ok := entry["prior_hash"]; ok {
				if entryPriorHash != priorHash {
					return fmt.Errorf("audit log chain broken: expected %s, got %v", priorHash, entryPriorHash)
				}
			}
		}

		// Update prior hash for next iteration
		if entryHash, ok := entry["entry_hash"]; ok {
			priorHash = fmt.Sprintf("%v", entryHash)
		}
	}

	return nil
}
