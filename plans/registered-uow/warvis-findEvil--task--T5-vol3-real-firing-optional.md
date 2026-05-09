---
note_type: task
artifact_type: task
id: T5-vol3-real-firing-optional
project_id: warvis-findEvil
title: Optional Volatility3 real firing evidence residual
status: open
phase: build
related_bet: BET-BIG-warvis-findEvil-warvis-findevil-bet-cc74
priority: P4
size: medium
updated_at: 2026-05-09T15:25:03Z
---
Attempt one trace where the agent loop invokes volatility3 pslist on the SANS memory image, or document honest miss if time-box expires. Only after T1-T4 complete and buffer remains before D-5 freeze.

STOP/TODO: Do not claim real Volatility3 agent-loop firing until a trace records the subprocess/tool_result evidence. Current acceptable state is the honest caveat in `docs/find-evil/accuracy-report.md`.
