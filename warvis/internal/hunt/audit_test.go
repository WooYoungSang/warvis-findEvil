package hunt

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestAuditLogAppendOnly(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hunt-audit-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	auditPath := filepath.Join(tmpDir, "audit.jsonl")
	auditLog, err := NewAuditLog(auditPath)
	if err != nil {
		t.Fatalf("failed to create audit log: %v", err)
	}
	defer auditLog.Close()

	// Append first entry
	if err := auditLog.Append(map[string]interface{}{
		"event": "case_opened",
		"case_id": "test-case",
	}); err != nil {
		t.Fatalf("first append failed: %v", err)
	}

	// Append second entry
	if err := auditLog.Append(map[string]interface{}{
		"event": "state_transition",
		"from": "INITIALIZE",
		"to": "TRACE",
	}); err != nil {
		t.Fatalf("second append failed: %v", err)
	}

	// Verify file exists and has 2 lines
	file, err := os.Open(auditPath)
	if err != nil {
		t.Fatalf("failed to open audit log: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineCount := 0
	for scanner.Scan() {
		lineCount++
	}

	if lineCount != 2 {
		t.Errorf("expected 2 lines, got %d", lineCount)
	}
}

func TestAuditLogJSONLines(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hunt-audit-jsonl")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	auditPath := filepath.Join(tmpDir, "audit.jsonl")
	auditLog, err := NewAuditLog(auditPath)
	if err != nil {
		t.Fatalf("failed to create audit log: %v", err)
	}
	defer auditLog.Close()

	// Append multiple entries
	entries := []map[string]interface{}{
		{"event": "case_opened", "case_id": "test-case"},
		{"event": "state_transition", "from": "INITIALIZE", "to": "TRACE"},
		{"event": "tool_called", "tool": "timeline.build"},
	}

	for _, entry := range entries {
		if err := auditLog.Append(entry); err != nil {
			t.Fatalf("append failed: %v", err)
		}
	}

	// Verify each line is valid JSON
	file, err := os.Open(auditPath)
	if err != nil {
		t.Fatalf("failed to open audit log: %v", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	lineCount := 0
	for {
		var entry map[string]interface{}
		err := decoder.Decode(&entry)
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			t.Fatalf("failed to parse JSON line: %v", err)
		}
		lineCount++

		// Verify entry has event field
		if _, ok := entry["event"]; !ok {
			t.Error("entry missing 'event' field")
		}
	}

	if lineCount != len(entries) {
		t.Errorf("expected %d entries, got %d", len(entries), lineCount)
	}
}

func TestAuditLogHashChain(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hunt-audit-hash")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	auditPath := filepath.Join(tmpDir, "audit.jsonl")
	auditLog, err := NewAuditLog(auditPath)
	if err != nil {
		t.Fatalf("failed to create audit log: %v", err)
	}
	defer auditLog.Close()

	// Append entries
	if err := auditLog.Append(map[string]interface{}{"event": "test1"}); err != nil {
		t.Fatalf("first append failed: %v", err)
	}

	if err := auditLog.Append(map[string]interface{}{"event": "test2"}); err != nil {
		t.Fatalf("second append failed: %v", err)
	}

	// Read and verify hash chain
	file, err := os.Open(auditPath)
	if err != nil {
		t.Fatalf("failed to open audit log: %v", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	var priorHash string
	lineCount := 0

	for {
		var entry map[string]interface{}
		err := decoder.Decode(&entry)
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			t.Fatalf("failed to parse JSON: %v", err)
		}

		lineCount++

		// First entry should not have prior_hash
		if lineCount == 1 {
			if _, ok := entry["prior_hash"]; ok {
				t.Error("first entry should not have prior_hash")
			}
		} else {
			// Subsequent entries should have prior_hash
			if priorHashVal, ok := entry["prior_hash"]; !ok {
				t.Error("entry missing prior_hash")
			} else {
				if priorHashVal != priorHash {
					t.Errorf("prior_hash mismatch: expected %s, got %v", priorHash, priorHashVal)
				}
			}
		}

		// Store hash for next iteration
		if entryHash, ok := entry["entry_hash"]; ok {
			priorHash = entryHash.(string)
		}
	}

	if lineCount != 2 {
		t.Errorf("expected 2 lines, got %d", lineCount)
	}
}

func TestAuditLogCreateAndAppend(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hunt-audit-create")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create two separate audit logs to same file
	auditPath := filepath.Join(tmpDir, "audit.jsonl")

	// First session
	auditLog1, err := NewAuditLog(auditPath)
	if err != nil {
		t.Fatalf("failed to create audit log 1: %v", err)
	}

	if err := auditLog1.Append(map[string]interface{}{"event": "session1_event"}); err != nil {
		t.Fatalf("append 1 failed: %v", err)
	}
	auditLog1.Close()

	// Second session (append mode)
	auditLog2, err := NewAuditLog(auditPath)
	if err != nil {
		t.Fatalf("failed to create audit log 2: %v", err)
	}

	if err := auditLog2.Append(map[string]interface{}{"event": "session2_event"}); err != nil {
		t.Fatalf("append 2 failed: %v", err)
	}
	auditLog2.Close()

	// Verify both entries exist
	file, err := os.Open(auditPath)
	if err != nil {
		t.Fatalf("failed to open audit log: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineCount := 0
	for scanner.Scan() {
		lineCount++
	}

	if lineCount != 2 {
		t.Errorf("expected 2 lines total, got %d", lineCount)
	}
}

func TestAuditLogVerify(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hunt-audit-verify")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	auditPath := filepath.Join(tmpDir, "audit.jsonl")
	auditLog, err := NewAuditLog(auditPath)
	if err != nil {
		t.Fatalf("failed to create audit log: %v", err)
	}

	// Append entries
	if err := auditLog.Append(map[string]interface{}{"event": "event1"}); err != nil {
		t.Fatalf("append failed: %v", err)
	}
	if err := auditLog.Append(map[string]interface{}{"event": "event2"}); err != nil {
		t.Fatalf("append failed: %v", err)
	}
	auditLog.Close()

	// Verify the chain
	if err := auditLog.Verify(auditPath); err != nil {
		t.Fatalf("verify failed: %v", err)
	}
}

func TestAuditLogEmptyFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hunt-audit-empty")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	auditPath := filepath.Join(tmpDir, "audit.jsonl")
	auditLog, err := NewAuditLog(auditPath)
	if err != nil {
		t.Fatalf("failed to create audit log: %v", err)
	}
	auditLog.Close()

	// Verify empty file passes verification
	if err := auditLog.Verify(auditPath); err != nil {
		t.Fatalf("verify on empty file failed: %v", err)
	}
}
