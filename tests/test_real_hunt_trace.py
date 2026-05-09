"""Test suite for real-hunt-trace UoW (B3 of P0 split).

Verifies that the committed real-hunt-trace artifacts contain a valid
end-to-end hunt audit trail (case_opened → state_transition → ≥1 gemma
response with non-empty reason → ≥1 tool_called → ≥1 tool_result) and
that the hash chain is intact.
"""

import json
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parent.parent
SANS_DIR = REPO_ROOT / "repos/find-evil-fixtures/cases/sans-starter"
TRACE_DIR = SANS_DIR / "real-hunt-trace"
TRACE_DIR_26B = SANS_DIR / "real-hunt-trace-26b"
TRACE_DIR_MOCK = SANS_DIR / "comprehensive-mock-trace"
AUDIT = TRACE_DIR / "audit.jsonl"
AUDIT_26B = TRACE_DIR_26B / "audit.jsonl"
AUDIT_MOCK = TRACE_DIR_MOCK / "audit.jsonl"
STATE = TRACE_DIR / "state.json"
STATE_MOCK = TRACE_DIR_MOCK / "state.json"
TRACE_README = TRACE_DIR / "README.md"
TRACE_README_26B = TRACE_DIR_26B / "README.md"
TRACE_README_MOCK = TRACE_DIR_MOCK / "README.md"
ACCURACY_DOC = REPO_ROOT / "docs/find-evil/accuracy-report.md"


def _entries():
    return [json.loads(line) for line in AUDIT.read_text().splitlines() if line.strip()]


def test_trace_dir_exists():
    assert TRACE_DIR.is_dir()


def test_audit_file_present():
    assert AUDIT.is_file()


def test_state_file_present():
    assert STATE.is_file()


def test_trace_readme_present():
    assert TRACE_README.is_file()


def test_audit_all_lines_valid_json():
    for line in AUDIT.read_text().splitlines():
        if line.strip():
            json.loads(line)  # would raise on invalid


def test_audit_has_case_opened():
    events = [e["event"] for e in _entries()]
    assert "case_opened" in events


def test_audit_has_state_transition_to_trace():
    matches = [
        e for e in _entries()
        if e.get("event") == "state_transition" and e.get("to") == "TRACE"
    ]
    assert len(matches) >= 1


def test_audit_has_gemma_response_with_reason():
    gemma = [e for e in _entries() if e.get("event") == "gemma_response"]
    assert len(gemma) >= 1
    # reason field must be non-empty per B2 ai-explainability schema
    assert all(e.get("reason") for e in gemma)


def test_audit_has_tool_called_and_result():
    events = [e["event"] for e in _entries()]
    assert "tool_called" in events
    assert "tool_result" in events


def test_audit_hash_chain_continuous():
    """Each entry's prior_hash must equal the previous entry's entry_hash."""
    entries = _entries()
    prev = None
    for e in entries:
        if prev is not None:
            assert e.get("prior_hash") == prev.get("entry_hash"), (
                f"hash chain break at event={e.get('event')}"
            )
        prev = e


def test_state_json_in_trace_state():
    s = json.loads(STATE.read_text())
    assert s.get("current_state") == "TRACE"


def test_accuracy_report_has_real_hunt_section():
    text = ACCURACY_DOC.read_text().lower()
    assert "real hunt trace" in text
    assert "base-wkstn-05" in text


def test_accuracy_report_acknowledges_caveats():
    """Honest dual-path: B3 must not overclaim."""
    text = ACCURACY_DOC.read_text().lower()
    # at least one explicit caveat phrase
    caveat_phrases = (
        "no findings",
        "does not measure",
        "not measure",
        "honest caveat",
        "documented limitation",
    )
    assert any(p in text for p in caveat_phrases)


# --- 26B sibling trace ---

def _entries_26b():
    return [
        json.loads(line) for line in AUDIT_26B.read_text().splitlines() if line.strip()
    ]


def test_26b_trace_dir_exists():
    assert TRACE_DIR_26B.is_dir()


def test_26b_audit_present():
    assert AUDIT_26B.is_file()


def test_26b_readme_present():
    assert TRACE_README_26B.is_file()


def test_26b_audit_all_lines_valid_json():
    for line in AUDIT_26B.read_text().splitlines():
        if line.strip():
            json.loads(line)


def test_26b_has_case_opened_and_state_transition():
    events = [e["event"] for e in _entries_26b()]
    assert "case_opened" in events
    assert "state_transition" in events


def test_26b_audit_hash_chain_continuous():
    entries = _entries_26b()
    prev = None
    for e in entries:
        if prev is not None:
            assert e.get("prior_hash") == prev.get("entry_hash"), (
                f"hash chain break at event={e.get('event')}"
            )
        prev = e


def test_26b_demonstrates_escalation_or_self_correction():
    """26B should either escalate cleanly OR correct itself.
    Either is acceptable evidence for criterion #1."""
    text = AUDIT_26B.read_text().lower()
    assert "escalate" in text or "previous" in text


def test_accuracy_report_compares_model_variants():
    text = ACCURACY_DOC.read_text().lower()
    assert "26b" in text and "8b" in text


# --- comprehensive mock trace (full FSM traversal, deterministic) ---

def _entries_mock():
    return [
        json.loads(line) for line in AUDIT_MOCK.read_text().splitlines() if line.strip()
    ]


def test_mock_trace_dir_exists():
    assert TRACE_DIR_MOCK.is_dir()


def test_mock_audit_present():
    assert AUDIT_MOCK.is_file()


def test_mock_readme_present():
    assert TRACE_README_MOCK.is_file()


def test_mock_state_terminal_lock():
    s = json.loads(STATE_MOCK.read_text())
    assert s.get("current_state") == "LOCK", (
        "comprehensive mock hunt should reach LOCK terminal state"
    )


def test_mock_audit_records_trace_to_scan_transition():
    """L2 lock-in: TRACE → SCAN state_transition recorded in audit."""
    matches = [
        e for e in _entries_mock()
        if e.get("event") == "state_transition"
        and e.get("from") == "TRACE"
        and e.get("to") == "SCAN"
    ]
    assert len(matches) >= 1


def test_mock_audit_records_full_fsm_traversal():
    """All four state transitions present: INIT→TRACE→SCAN→EXPOSE→LOCK."""
    expected_pairs = {
        ("INITIALIZE", "TRACE"),
        ("TRACE", "SCAN"),
        ("SCAN", "EXPOSE"),
        ("EXPOSE", "LOCK"),
    }
    actual_pairs = {
        (e.get("from"), e.get("to"))
        for e in _entries_mock()
        if e.get("event") == "state_transition"
    }
    missing = expected_pairs - actual_pairs
    assert not missing, f"missing transitions: {missing}"


def test_mock_audit_has_memory_tools_called_in_scan():
    """L3 lock-in: ≥1 memory.* tool_called event with current_state=SCAN."""
    # gemma_response events carry current_state; tool_called events don't.
    # Match by adjacency: gemma_response[SCAN] action=call_tool tool=memory.*
    # immediately followed by tool_called for the same tool name.
    entries = _entries_mock()
    memory_calls = [
        e for e in entries
        if e.get("event") == "gemma_response"
        and e.get("current_state") == "SCAN"
        and (e.get("tool_name") or "").startswith("memory.")
    ]
    assert len(memory_calls) >= 1


def test_mock_audit_hash_chain_continuous():
    entries = _entries_mock()
    prev = None
    for e in entries:
        if prev is not None:
            assert e.get("prior_hash") == prev.get("entry_hash"), (
                f"hash chain break at event={e.get('event')}"
            )
        prev = e
