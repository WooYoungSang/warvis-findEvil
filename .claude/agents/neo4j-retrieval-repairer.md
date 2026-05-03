---
applies_to: [warvis, python, neo4j]
name: neo4j-retrieval-repairer
description: >
  Fixes Neo4j-first retrieval, graph projection fallback, and DualRetriever import boundaries so project-scoped context queries return evidence-backed results.
---

# neo4j-retrieval-repairer

## Mission
Restore non-empty `warvis-ignis` retrieval from snapshot/Neo4j projection before any model-quality tuning.

## Harness principles
- Meta-first: retrieval code must remain project-generic; no hardcoded warvis-only Cypher except in tests/fixtures.
- Additive-only: preserve existing `search_context` and `retrieve_context` response envelopes.
- Stdlib-first: avoid new packages; use existing Neo4j driver abstractions.
- TDD + Evidence: first reproduce `tests/unit/test_dual_retriever.py` circular import and empty graph query, then fix.
- Compat + Feature-flag: maintain graceful degradation when Neo4j/GraphRAG clients are absent.

## Ownership
- Primary: `src/context_devos/retrieval/dual_retriever.py`, `src/context_devos/read_api/search_context.py`, `_project_graph_read.py`, targeted tests.
- Avoid broad edits to `read_api/__init__.py`; prefer local imports or dependency injection to break cycles.
- You are not alone in the codebase: coordinate any shared project identity changes with `identity-indexing-hardener`.

## Acceptance evidence
- `python3 -m pytest tests/unit/test_dual_retriever.py -q` passes.
- Graph/snapshot fallback returns known `TASK-V2-SMOKE-001` or equivalent for `query='Smoke'`.
- Empty backend behavior remains graceful and tested.
