"""Test suite for valhuntir-analysis UoW.

Verifies docs/find-evil/valhuntir-comparison.md structural completeness
and factual claim consistency.
"""

from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parent.parent
DOC = REPO_ROOT / "docs/find-evil/valhuntir-comparison.md"


def _read() -> str:
    return DOC.read_text()


class TestValhuntirComparisonFile:
    def test_file_exists(self):
        assert DOC.exists()

    def test_valhuntir_keyword_min_5(self):
        assert _read().count("Valhuntir") >= 5

    def test_criterion_keyword_min_6(self):
        assert _read().lower().count("criterion") >= 6

    def test_has_differentiation_or_advantage(self):
        content = _read().lower()
        assert "differentiation" in content or "advantage" in content

    def test_min_3_tables(self):
        # markdown table separator rows
        content = _read()
        sep_count = content.count("|---") + content.count("| ---")
        assert sep_count >= 3, f"expected >=3 table separators, got {sep_count}"

    def test_has_honesty_phrase(self):
        content = _read().lower()
        assert any(
            p in content
            for p in ("we don't claim", "we do not claim", "we focus", "honest")
        )

    def test_has_executive_summary(self):
        assert "executive summary" in _read().lower()

    def test_has_architecture_compare(self):
        content = _read().lower()
        assert "architecture compare" in content or "## 2." in content

    def test_has_six_criteria_section(self):
        content = _read().lower()
        assert "six-criteria" in content or "6-criteria" in content or "six criteria" in content

    def test_has_differentiation_section(self):
        content = _read().lower()
        assert "differentiat" in content

    def test_has_honest_positioning(self):
        content = _read().lower()
        assert "honest positioning" in content or "positioning statement" in content

    def test_has_action_items(self):
        content = _read().lower()
        assert "action item" in content

    def test_no_overclaims(self):
        # we should NOT claim production-grade accuracy on real samples
        content = _read().lower()
        assert "we claim 100%" not in content
        if "production-ready" in content:
            assert (
                "we don't claim" in content
                or "we do not claim" in content
            )

    def test_references_kill_switch_harness(self):
        assert "kill-switch" in _read().lower() or "kill switch" in _read().lower()

    def test_references_fsm(self):
        content = _read().lower()
        assert "fsm" in content or "finite state machine" in content or "state machine" in content

    def test_references_go_binary(self):
        content = _read().lower()
        assert "go binary" in content or "single go" in content
