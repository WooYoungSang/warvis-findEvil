#!/usr/bin/env bash
# ITEM-212 FIND EVIL — pre-commit secret scan
set -e
git diff --cached --name-only | while read f; do
  [ -z "$f" ] && continue
  if [[ "$f" == *".env"* && "$f" != *"env.sample"* ]]; then
    echo "REJECT: .env file staged: $f"; exit 1
  fi
  if git show ":$f" 2>/dev/null | grep -E '(AKIA|sk-[A-Za-z0-9]{32}|ghp_[A-Za-z0-9]{36}|xoxb-)' >/dev/null; then
    echo "REJECT: high-entropy secret in: $f"; exit 1
  fi
done
echo "secrets-scan: PASS"
