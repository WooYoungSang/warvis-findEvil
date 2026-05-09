"""Test suite for real-hunt-trace UoW (B3 of P0 split).

Verifies that the committed real-hunt-trace artifacts contain a valid
end-to-end hunt audit trail (case_opened → state_transition → ≥1 gemma
response with non-empty reason → ≥1 tool_called → ≥1 tool_result) and
that the hash chain is intact.
"""

import json
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parent.parent
TRACE_DIR = REPO_ROOT / "repos/find-evil-fixtures/cases/sans-starter/real-hunt-trace"
AUDIT = TRACE_DIR / "audit.jsonl"
STATE = TRACE_DIR / "state.json"
TRACE_README = TRACE_DIR / "README.md"
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
