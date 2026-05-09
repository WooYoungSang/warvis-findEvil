"""Test suite for real-sample-acquisition UoW (B1 of P0 split).

Verifies that the SANS FIND EVIL starter case data has been acquired,
documented in MANIFEST.md, and referenced from dataset.md.
"""

import re
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parent.parent
SANS_DIR = REPO_ROOT / "repos/find-evil-fixtures/cases/sans-starter"
MANIFEST = SANS_DIR / "MANIFEST.md"
EVIDENCE_DIR = SANS_DIR / "evidence"
DATASET_DOC = REPO_ROOT / "docs/find-evil/dataset.md"

EGNYTE_URL_FRAGMENT = "sansorg.egnyte.com/fl/HhH7crTYT4JK"
SHA256_RE = re.compile(r"\b[0-9a-fA-F]{64}\b")


def test_sans_dir_exists():
    assert SANS_DIR.is_dir(), f"missing dir: {SANS_DIR}"


def test_evidence_dir_resolves():
    # evidence/ is a symlink to /mnt/disk1/...
    assert EVIDENCE_DIR.exists(), f"missing evidence dir: {EVIDENCE_DIR}"
    assert EVIDENCE_DIR.is_dir()


def test_evidence_has_at_least_one_file():
    files = [p for p in EVIDENCE_DIR.iterdir() if p.is_file()]
    assert len(files) >= 1, "evidence/ has no files yet"


def test_manifest_exists():
    assert MANIFEST.is_file(), f"missing: {MANIFEST}"


def test_manifest_has_sha256_hash():
    content = MANIFEST.read_text()
    assert SHA256_RE.search(content), (
        "MANIFEST.md must contain at least one 64-char SHA256 hash"
    )


def test_manifest_references_source_url():
    content = MANIFEST.read_text()
    assert EGNYTE_URL_FRAGMENT in content, (
        f"MANIFEST.md must reference source URL fragment '{EGNYTE_URL_FRAGMENT}'"
    )


def test_dataset_doc_has_sans_starter_section():
    content = DATASET_DOC.read_text().lower()
    assert "sans-starter" in content or "sans starter" in content


def test_dataset_doc_references_source_url():
    content = DATASET_DOC.read_text()
    assert EGNYTE_URL_FRAGMENT in content
