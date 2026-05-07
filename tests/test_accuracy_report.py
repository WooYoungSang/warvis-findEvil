"""ITEM-212 Phase 4 — accuracy-report.md validation tests.

Tests verify that docs/find-evil/accuracy-report.md exists and contains
all required sections and keywords for evaluator credibility.
"""
import pytest
from pathlib import Path


REPORT_PATH = Path(__file__).resolve().parents[1] / "docs" / "find-evil" / "accuracy-report.md"


class TestAccuracyReportExists:
    """Test 1: File exists"""

    def test_accuracy_report_exists(self):
        assert REPORT_PATH.exists(), f"accuracy-report.md missing at {REPORT_PATH}"


class TestAccuracyReportContent:
    """Tests 2-7: Required keywords and sections"""

    @pytest.fixture(autouse=True)
    def report_content(self):
        """Load report content once per test class."""
        if not REPORT_PATH.exists():
            pytest.skip("accuracy-report.md does not exist yet")
        self.content = REPORT_PATH.read_text(encoding="utf-8")

    def test_has_recall_keyword(self):
        """Test 2: Contains 'recall'"""
        assert "recall" in self.content.lower(), "report must contain keyword 'recall'"

    def test_has_synthetic_keyword(self):
        """Test 3: Contains 'synthetic'"""
        assert "synthetic" in self.content.lower(), "report must contain keyword 'synthetic'"

    def test_has_case_001_reference(self):
        """Test 4: Contains 'case-001'"""
        assert "case-001" in self.content, "report must reference case-001"

    def test_has_13_ground_truth(self):
        """Test 5: Contains '13 ground-truth' or equivalent"""
        assert "13" in self.content and "ground" in self.content.lower(), \
            "report must mention 13 ground-truth findings"

    def test_has_accuracy_py_command(self):
        """Test 6: Contains 'accuracy.py'"""
        assert "accuracy.py" in self.content, "report must reference accuracy.py script"

    def test_has_lite_vs_real_divergence(self):
        """Test 7: Contains both 'lite' and 'real' in divergence section"""
        content_lower = self.content.lower()
        has_lite = "lite" in content_lower
        has_real = "real" in content_lower
        assert has_lite and has_real, \
            "report must discuss both lite and real paths (divergence section)"

    def test_has_evaluator_methodology(self):
        """Test 8: Contains 'evaluator' or 'methodology'"""
        content_lower = self.content.lower()
        has_evaluator = "evaluator" in content_lower
        has_methodology = "methodology" in content_lower
        assert has_evaluator or has_methodology, \
            "report must include evaluator methodology or methodology section"
