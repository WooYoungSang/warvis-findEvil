# Plan: real-tool-installation — volatility3 install + smoke-test (B2)

**UoW ID**: `real-tool-installation` (B2 of 3-step split for P0 real-sample-integration)
**dev_session_id**: `1dd7b55d-6138-496c-85a8-3391070c55d7`
**Project**: warvis-findEval

## Context

B1 acquired `base-wkstn-05-memory.img` (3.0 GiB raw Windows memory dump). B2 installs the analyzer binary that the Phase 1+2-locked `src/find_evil_mcp/sift_runner.py` BINARIES whitelist already names — and verifies end-to-end parsing.

## Scope

- Install **volatility3** via `pip install --user volatility3` (no sudo). Provides `~/.local/bin/vol`, matching `sift_runner.BINARIES["volatility3"] = "vol"` exactly.
- Smoke test: `vol -f base-wkstn-05-memory.img windows.info` → capture to `repos/find-evil-fixtures/cases/sans-starter/smoke-vol-windows-info.txt`.
- Document in `docs/find-evil/dataset.md` §Installed Tools.

## Out of Scope (deferred or N/A)

- **yara CLI** — `yara-python` is a Python module, not the CLI; sudo `apt install yara` not requested for B2. Documented as deferred.
- **plaso / zeek / suricata** — N/A; no disk image / PCAP acquired in B1.
- **sigma** — rule format, no binary.
- Modifying `src/find_evil_mcp/sift_runner.py` — Phase 1+2 locked.

## Milestones

- M1 RED: `tests/test_real_tool_installation.py` (7 assertions) confirms `vol` absent → fails.
- M2 GREEN: `pip install --user volatility3` → 2.28.0 installed.
- M3 SMOKE: `vol -f .img windows.info` → exit 0 + Windows 7 NT 6.1 metadata extracted; SystemTime equals dc3dd capture log to the second.
- M4 DOCS: dataset.md §Installed Tools written.
- M5 REGRESSION: pytest 62/62 PASS, ruff clean.

## Verification Strategy

- pytest tests/test_real_tool_installation.py PASS (7 tests)
- pytest tests/ regression — no break
- ruff clean
- Smoke evidence file contains Windows kernel markers (NtMajorVersion / Is64Bit / KdVersionBlock / Kernel Base)
- SystemTime equality assertion: parser-extracted == capture-log == 2018-09-06T19:51:09Z

## Risk

**LOW** — pip --user only, no sudo, no Phase 1+2 / Phase 3 code changes, regression gate held.

---

**Status**: shipped (2026-05-09, 62/62 tests PASS, ruff clean, vol parsed real SANS image, lesson_id=74766fec-1421-4350-be01-d93299ff4213)
**Risk Level**: LOW
