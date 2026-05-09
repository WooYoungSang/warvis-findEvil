---
id: sift-vm-cold-start
project: warvis-findeval
status: blocked
priority: P1
created: 2026-05-09
owner: Codex
---

# STOP — SIFT VM cold-start validation blocked

## Target
Validate mandatory artifact #7: a fresh SIFT Workstation VM can clone this repo and run the documented `warvis hunt` quickstart to completion, producing `repos/find-evil-fixtures/cases/sans-starter/sift-vm-cold-start.log`.

## Blocker evidence (2026-05-09)

Local inspection found no usable SIFT VM or local virtualization runner:

- No SIFT OVA/OVF found under `/home/jang` or `/mnt` (only unrelated `node_modules/sift` and project `sift_runner.py`).
- Missing virtualization commands: `VBoxManage`, `qemu-system-x86_64`, `virt-install`, `vmrun`, `multipass`.
- Disk is probably sufficient (`/mnt/disk1` has ~304 GiB free), but execution cannot proceed without a SIFT VM runtime.

## Decision
Per handoff constraints, do **not** substitute Ubuntu/Debian/macOS/host validation for the SIFT Workstation requirement. The next required action is for a human/operator to provision a running SIFT Workstation VM or provide the OVA plus an approved virtualization path.

## Resume steps once provisioned

1. Boot a fresh SIFT Workstation VM.
2. Clone `https://github.com/WooYoungSang/warvis-findEvil.git`.
3. Follow only `README.md` quickstart steps; patch README for every missing step discovered.
4. Capture the full terminal session at `repos/find-evil-fixtures/cases/sans-starter/sift-vm-cold-start.log`.
5. Add `tests/test_sift_cold_start.py` asserting the log exists and includes `git clone`, `warvis hunt`, and a non-zero `audit.jsonl` line count.
6. Run the four standard gates before commit.

## Stop condition
Blocked until SIFT Workstation VM access exists. No code or README changes were made for this task in this STOP note.
