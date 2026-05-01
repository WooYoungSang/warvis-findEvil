# ITEM-212 FIND EVIL — Secrets Policy

## Rules
1. **env-only** — all secrets via `harness/find-evil/.env` (gitignored). Never hardcoded.
2. **No print** — server logs MUST redact env values. Use `***REDACTED***` placeholder.
3. **No commit** — pre-commit hook (`scripts/secrets-scan.sh`) blocks staged files containing high-entropy strings or `.env`.
4. **No exfiltration** — MCP server tools must NOT read env vars on behalf of the LLM. Env stays in process memory only.
5. **Webhook URLs** (Discord, etc.) are mission-control owned. This node leaves them empty.

## Required gitignore entries
```
harness/find-evil/.env
harness/find-evil/.env.*
!harness/find-evil/env.sample
harness/find-evil/logs/*.jsonl
docs/find-evil/demo.mp4
repos/find-evil-fixtures/**/raw/**
```

## Pre-commit hook (sketch)
```bash
#!/usr/bin/env bash
# harness/find-evil/scripts/secrets-scan.sh
set -e
git diff --cached --name-only | while read f; do
  [ -z "$f" ] && continue
  if [[ "$f" == *".env"* && "$f" != *"env.sample"* ]]; then
    echo "REJECT: .env file staged: $f"; exit 1
  fi
  # Crude high-entropy check
  if git show ":$f" 2>/dev/null | grep -E '(AKIA|sk-[A-Za-z0-9]{32}|ghp_[A-Za-z0-9]{36}|xoxb-)' >/dev/null; then
    echo "REJECT: high-entropy secret in: $f"; exit 1
  fi
done
echo "secrets-scan: PASS"
```

## Operator checklist
- [ ] `.env` exists locally, gitignored, mode 0600.
- [ ] CI never reads `.env`; CI uses repo secrets / env injection.
- [ ] LLM tool surface does NOT include `os.environ` or `dotenv` reads.
- [ ] Demo video confirms no terminal shows real keys.
