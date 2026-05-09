"""Tests for the paste-ready Devpost form draft.

The draft should project the long-form narrative into Devpost field shape so
the final submission can be copied without re-deriving required fields.
"""

from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parent.parent
FORM_DRAFT = REPO_ROOT / "docs/find-evil/devpost-form-draft.md"

REQUIRED_HEADINGS = (
    "## Project name",
    "## Tagline",
    "## Short description",
    "## Long description",
    "## How it's made",
    "## Built with",
    "## Try it out link",
    "## Video URL",
    "## Image gallery thumbnail spec",
    "## GitHub repo",
    "## Team members + W-8BEN reminder",
)


def test_devpost_form_draft_exists():
    assert FORM_DRAFT.exists(), "docs/find-evil/devpost-form-draft.md must exist"


def test_devpost_form_draft_has_required_field_headings():
    content = FORM_DRAFT.read_text()

    for heading in REQUIRED_HEADINGS:
        assert heading in content, f"missing Devpost field heading: {heading}"


def test_devpost_form_draft_preserves_required_limits_and_links():
    content = FORM_DRAFT.read_text()

    assert "W.A.R.V.I.S. — Find Evil" in content
    assert "≤ 80 chars" in content
    assert "≤ 200 chars" in content
    assert "1280×640" in content
    assert "https://github.com/WooYoungSang/warvis-findEvil" in content
