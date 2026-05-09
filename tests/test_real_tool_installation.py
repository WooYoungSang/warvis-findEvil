"""Test suite for real-tool-installation UoW (B2 of P0 split).

Verifies that volatility3 is installed and can parse the B1-acquired
SANS memory dump, and that dataset.md documents the install.
"""

import shutil
import subprocess
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parent.parent
SANS_DIR = REPO_ROOT / "repos/find-evil-fixtures/cases/sans-starter"
SANS_IMG = SANS_DIR / "evidence/base-wkstn-05-memory.img"
DATASET_DOC = REPO_ROOT / "docs/find-evil/dataset.md"
SMOKE_EVIDENCE = SANS_DIR / "smoke-vol-windows-info.txt"


def test_vol_binary_available():
    """vol must be on PATH (provided by `pip install --user volatility3`)."""
    assert shutil.which("vol") is not None, "vol binary not on PATH"


def test_vol_help_runs():
    """vol --help must exit 0."""
    result = subprocess.run(
        ["vol", "--help"], capture_output=True, text=True, timeout=30
    )
    assert result.returncode == 0, f"vol --help exit {result.returncode}: {result.stderr[:200]}"


def test_volatility3_python_module_importable():
    import volatility3  # noqa: F401


def test_smoke_evidence_file_exists():
    """The smoke-test output (vol windows.info) must be captured."""
    assert SMOKE_EVIDENCE.is_file(), f"missing: {SMOKE_EVIDENCE}"


def test_smoke_evidence_contains_windows_marker():
    """The smoke-test output must contain Windows OS metadata
    (e.g. 'NTBuildLab', 'NtMajorVersion', or 'Is64Bit')."""
    content = SMOKE_EVIDENCE.read_text()
    markers = ("NTBuildLab", "NtMajorVersion", "Is64Bit", "KdVersionBlock", "Kernel Base")
    assert any(m in content for m in markers), (
        f"smoke evidence has none of {markers}"
    )


def test_dataset_doc_has_installed_tools_section():
    content = DATASET_DOC.read_text().lower()
    assert "installed tools" in content or "installed-tools" in content


def test_dataset_doc_mentions_volatility3():
    content = DATASET_DOC.read_text().lower()
    assert "volatility3" in content
