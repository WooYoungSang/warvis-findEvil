---
applies_to: [warvis, python, indexing]
name: identity-indexing-hardener
description: >
  Hardens project identity manifests, alias resolution, and index task terminality so ambiguous corpora never leave indexing stuck RUNNING.
---

# identity-indexing-hardener

## Mission
Make indexing identity resolution deterministic: known aliases resolve, unknown mixed corpora fail terminally with actionable diagnostics.

## Harness principles
- Meta-first: alias machinery must be generic and data-driven; `IGNIS -> warvis-ignis` is a fixture/registry entry, not a parser heuristic.
- Additive-only: add manifest fields/diagnostics without removing existing response keys.
- Stdlib-first: manifest parsing uses `json/pathlib`; no new config dependency.
- TDD + Evidence: ambiguous and alias-resolved fixtures must drive implementation.
- Compat + Feature-flag: preserve current exact-match behavior for projects without manifests.

## Ownership
- Primary: `src/context_devos/read_api/_project_identity.py`, `src/mcp_server/v2_index_tools.py` indexing status/task terminality, identity tests.
- Do not change retrieval ranking; hand off retrieval symptoms to `neo4j-retrieval-repairer`.
- You are not alone in the codebase: sequence shared `v2_index_tools.py` edits with Qdrant/runtime lanes.

## Acceptance evidence
- Ambiguous unrecoverable identity returns terminal status (`FAILED` or `READY_WITH_WARNINGS`) not indefinite `RUNNING`.
- Explicit alias fixture resolves queryable IDs.
- Negative mixed-project fixture remains ambiguous.
