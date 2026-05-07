"""Test suite for architecture.md document validation.

Ensures Phase 3+4 expansion meets minimum line count, keyword coverage,
and diagram requirements.
"""

def load_architecture_doc():
    """Load the architecture.md file."""
    with open("docs/find-evil/architecture.md", "r") as f:
        return f.read()


def test_arch_doc_lines_min_600():
    """Test that architecture.md has at least 600 lines."""
    content = load_architecture_doc()
    lines = content.splitlines()
    assert len(lines) >= 600, f"Expected ≥600 lines, got {len(lines)}"


def test_has_all_fsm_states():
    """Test that all 5 FSM states are mentioned."""
    content = load_architecture_doc()
    states = ["INITIALIZE", "TRACE", "SCAN", "EXPOSE", "LOCK"]
    for state in states:
        assert state in content, f"Missing FSM state: {state}"


def test_has_jsonrpc_keyword():
    """Test that JSON-RPC is mentioned."""
    content = load_architecture_doc()
    assert "JSON-RPC" in content, "Missing 'JSON-RPC' keyword"


def test_has_gemma_keyword():
    """Test that Gemma is mentioned (case-sensitive)."""
    content = load_architecture_doc()
    assert "Gemma" in content, "Missing 'Gemma' keyword (case-sensitive)"


def test_has_audit_jsonl():
    """Test that audit.jsonl is mentioned."""
    content = load_architecture_doc()
    assert "audit.jsonl" in content, "Missing 'audit.jsonl' keyword"


def test_has_mcp_keyword():
    """Test that MCP is mentioned."""
    content = load_architecture_doc()
    assert "MCP" in content, "Missing 'MCP' keyword"


def test_has_two_diagrams():
    """Test that document has at least 2 diagrams (mermaid or ascii art)."""
    content = load_architecture_doc()
    # Count mermaid blocks
    mermaid_blocks = content.count("```mermaid")
    # Count ascii art indicators
    ascii_markers = (
        content.count("┌") + content.count("┐")
        + content.count("└") + content.count("┘")
    )

    total_diagrams = mermaid_blocks + (1 if ascii_markers >= 4 else 0)
    assert total_diagrams >= 2, (
        f"Expected ≥2 diagrams, got {total_diagrams} "
        f"(mermaid:{mermaid_blocks}, ascii:{ascii_markers})"
    )
