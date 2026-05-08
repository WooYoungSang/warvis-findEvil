"""Test suite for narrative-pivot UoW.

Verifies the new tagline + Valhuntir honest positioning + 3 differentiator
keywords are propagated to README.md, devpost-page.md, and CLAUDE.md.
"""

from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parent.parent
TAGLINE = "smallest IR agent whose architecture"
HONESTY_PHRASES = ("we don't claim", "we do not claim", "we focus")


def _read(rel_path: str) -> str:
    return (REPO_ROOT / rel_path).read_text()


def _has_any(content: str, phrases: tuple) -> bool:
    lower = content.lower()
    return any(p in lower for p in phrases)


def test_tagline_in_readme():
    assert TAGLINE in _read("README.md")


def test_tagline_in_devpost():
    assert TAGLINE in _read("docs/find-evil/devpost-page.md")


def test_tagline_in_claude_md():
    assert TAGLINE in _read("CLAUDE.md")


def test_readme_has_single_go_binary():
    assert "single go binary" in _read("README.md").lower()


def test_readme_has_fsm():
    content = _read("README.md").lower()
    assert "fsm" in content or "finite state machine" in content


def test_readme_has_resume():
    assert "resume" in _read("README.md").lower()


def test_readme_mentions_valhuntir():
    assert "Valhuntir" in _read("README.md")


def test_devpost_mentions_valhuntir():
    assert "Valhuntir" in _read("docs/find-evil/devpost-page.md")


def test_readme_honesty_phrase():
    assert _has_any(_read("README.md"), HONESTY_PHRASES)


def test_devpost_honesty_phrase():
    assert _has_any(_read("docs/find-evil/devpost-page.md"), HONESTY_PHRASES)


def test_devpost_has_compare_section():
    content = _read("docs/find-evil/devpost-page.md").lower()
    has_heading = any(
        marker in content
        for marker in (
            "## how we compare to valhuntir",
            "## compared to valhuntir",
            "## vs valhuntir",
            "## valhuntir comparison",
        )
    )
    assert has_heading, "expected a 'compare to Valhuntir' heading in devpost-page.md"
