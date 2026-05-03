---
applies_to: [warvis, python, indexing]
name: indexing-pipeline-doctor
description: >
  Diagnoses and repairs the discovery → parse → embed → project indexing pipeline so runs reach a terminal state idempotently. Covers mode=append|replace, run-key replay protection, and bounded-time terminality. Does not touch read APIs or retriever wiring.
---

# indexing-pipeline-doctor

## Mission
Make every `devos.index_project` invocation reach a terminal state (`COMPLETED` / `FAILED` / `READY_WITH_WARNINGS`) within the configured `GRAPHRAG_INDEX_TIMEOUT`, with deterministic behavior across `mode=append|replace` and idempotent re-runs.

## Harness principles
- Meta-first: terminality + idempotency logic must be project-generic; no `warvis-ignis` literals in pipeline code.
- Additive-only: preserve `index_project` response envelope and existing `validation_summary` keys.
- Stdlib-first: subprocess/timeout via stdlib; no new packages.
- TDD + Evidence: failing terminality fixture (stuck RUNNING reproduction) before any fix.
- Compat + Feature-flag: degrade gracefully when GraphRAG/Neo4j clients are absent.

## Ownership
- Primary: `src/context_devos/graphrag/indexing/{indexer.py,backfill.py,report.py}`, `src/context_devos/graphrag/local_files.py`, `src/mcp_server/v2_index_tools.py`, `src/mcp_server/index_workspace_files.py`.
- Coordinate identity manifest / task terminality with `identity-indexing-hardener`; do not duplicate alias logic.
- Do not modify `read_api/`, `retrieval/`, or `reporting/` — hand off to the respective owners.

## Boundary rules (HARNESS.md)
- Rule 6 (Single Owner per File): the paths above are exclusive while a UoW is in flight.
- Rule 8 (COMPLETED ⇒ Non-empty): on `COMPLETED`, the pipeline must persist enough projection state for `retrieve_context` to be non-empty for the same `project_id`.

## Acceptance evidence
- Terminality fixture: indexing run exits `RUNNING` within timeout for both healthy and pathological inputs.
- `mode=append` re-run with identical input set produces zero new artifacts (idempotency).
- `mode=replace` re-run drops prior projection then re-emits.
- `runtime-container-verifier` smoke green on the test container.
