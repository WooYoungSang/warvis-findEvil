"""Acceptance checks for the SIFT Workstation cold-start evidence.

The real SIFT VM run is an external, mandatory artifact. This scaffold keeps the
expected evidence shape executable in CI without pretending that the VM evidence
has already been captured.
"""

from pathlib import Path

import pytest


REPO_ROOT = Path(__file__).resolve().parent.parent
COLD_START_LOG = (
    REPO_ROOT
    / "repos/find-evil-fixtures/cases/sans-starter/sift-vm-cold-start.log"
)
STOP_NOTE = REPO_ROOT / ".omc/plans/sift-vm-cold-start.md"


@pytest.fixture
def cold_start_log_text():
    if not COLD_START_LOG.exists():
        if STOP_NOTE.exists():
            pytest.skip("SIFT VM cold-start blocked; STOP note exists")
        pytest.skip("SIFT VM cold-start evidence has not been captured yet")
    return COLD_START_LOG.read_text(errors="replace")


def test_sift_cold_start_log_path_is_trackable():
    """The required log path must not be hidden by blanket fixture ignores."""
    assert COLD_START_LOG.parent.exists(), "sans-starter fixture directory missing"


def test_sift_cold_start_log_contains_git_clone(cold_start_log_text):
    assert "git clone" in cold_start_log_text


def test_sift_cold_start_log_contains_warvis_hunt(cold_start_log_text):
    assert "warvis hunt" in cold_start_log_text


def test_sift_cold_start_log_records_nonzero_audit_lines(cold_start_log_text):
    markers = (
        "audit.jsonl line count:",
        "audit.jsonl lines:",
        "non-zero audit.jsonl",
    )
    assert any(marker in cold_start_log_text for marker in markers)
    assert "audit.jsonl line count: 0" not in cold_start_log_text
    assert "audit.jsonl lines: 0" not in cold_start_log_text
