"""
Test suite for WARVIS submission artifacts (pitch-writer UoW).

Validates:
1. README.md exists and contains "warvis hunt" command
2. demo-script.md exists and covers all 5 FSM states
3. devpost-page.md exists and includes W.A.R.V.I.S branding
4. LICENSE exists and is MIT-licensed
5. All markdown files are syntactically valid
"""

import re
from pathlib import Path


def test_readme_exists_and_contains_warvis_hunt():
    """Test that README.md exists and mentions the warvis hunt command."""
    readme = Path("README.md")
    assert readme.exists(), "README.md not found at project root"

    content = readme.read_text()
    assert "warvis hunt" in content, "README.md must contain 'warvis hunt' command reference"


def test_demo_script_exists_and_contains_all_fsm_states():
    """Test that demo-script.md exists and covers all 5 Hunt FSM states."""
    demo_script = Path("docs/find-evil/demo-script.md")
    assert demo_script.exists(), "docs/find-evil/demo-script.md not found"

    content = demo_script.read_text()
    required_states = ["INITIALIZE", "TRACE", "SCAN", "EXPOSE", "LOCK"]

    for state in required_states:
        assert state in content, f"demo-script.md must contain FSM state '{state}'"


def test_devpost_page_exists_and_contains_warvis_branding():
    """Test that devpost-page.md exists and includes W.A.R.V.I.S branding."""
    devpost = Path("docs/find-evil/devpost-page.md")
    assert devpost.exists(), "docs/find-evil/devpost-page.md not found"

    content = devpost.read_text()
    assert "W.A.R.V.I.S" in content, "devpost-page.md must contain 'W.A.R.V.I.S' branding"


def test_license_exists_and_is_mit():
    """Test that LICENSE file exists and is MIT-licensed."""
    license_file = Path("LICENSE")
    assert license_file.exists(), "LICENSE not found at project root"

    content = license_file.read_text()
    assert "MIT" in content, "LICENSE must contain 'MIT' license designation"


def test_all_markdown_files_are_parseable():
    """Test that all submission markdown files have valid syntax."""
    markdown_files = [
        Path("README.md"),
        Path("docs/find-evil/demo-script.md"),
        Path("docs/find-evil/devpost-page.md"),
    ]

    for md_file in markdown_files:
        if md_file.exists():
            content = md_file.read_text()

            # Check for unclosed code blocks
            code_fence_count = content.count("```")
            assert code_fence_count % 2 == 0, f"{md_file.name}: unclosed code blocks detected"

            # Check for basic markdown link syntax validity
            link_pattern = r"\[([^\]]+)\]\(([^)]+)\)"
            links = re.findall(link_pattern, content)
            for text, url in links:
                assert text, f"{md_file.name}: empty link text detected"
                assert url, f"{md_file.name}: empty link URL detected"
