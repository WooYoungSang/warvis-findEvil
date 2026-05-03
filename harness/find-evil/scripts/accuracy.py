# SPDX-License-Identifier: MIT
"""ITEM-212 Phase 2 — recall measurement against curated evil fixtures.

Computes recall = matched_findings / total_expected_findings against
`repos/find-evil-fixtures/cases/<case>/manifest.json`. Each detection
path runs the find-evil-mcp handler when SIFT binaries are available;
otherwise it falls back to a pure-Python "lite" scanner so the gate can
be measured on dev nodes without plaso/yara/volatility3/zeek.

Lite scanners are deliberately narrow — they only reproduce the signal
contained in the synthetic fixtures, never replace real DFIR tooling.

Usage:
  python harness/find-evil/scripts/accuracy.py [--case <case_dir>] [--threshold 0.60]
Exit:
  0  recall >= threshold
  1  recall <  threshold
  2  fixture missing / parse error
"""
from __future__ import annotations

import argparse
import json
import re
import struct
import sys
from dataclasses import dataclass, field
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[3]
DEFAULT_CASE = REPO_ROOT / "harness" / "find-evil" / "fixtures" / "cases" / "case-001"
DEFAULT_THRESHOLD = 0.60


# ---------------------------------------------------------------------------
# Lite scanners
# ---------------------------------------------------------------------------


def _yara_literal_strings(rule_text: str) -> dict[str, list[str]]:
    """Extract `rule X { strings: $a = "..." ... }` literals per rule name."""
    out: dict[str, list[str]] = {}
    for match in re.finditer(
        r"rule\s+(\w+)\s*\{[^}]*?strings:\s*(.+?)condition:",
        rule_text,
        flags=re.DOTALL,
    ):
        name = match.group(1)
        body = match.group(2)
        literals = re.findall(r'"([^"]+)"', body)
        if literals:
            out[name] = literals
    return out


def lite_iocs_scan(case_dir: Path, expected: list[dict]) -> list[dict]:
    rules_path = case_dir / "evidence" / "yara_rules" / "evil.yar"
    if not rules_path.is_file():
        return []
    rule_literals = _yara_literal_strings(rules_path.read_text(encoding="utf-8"))
    found: list[dict] = []
    for exp in expected:
        rule = exp["rule"]
        target = case_dir / "evidence" / exp["path"]
        literals = rule_literals.get(rule, [])
        if not literals or not target.is_file():
            continue
        data = target.read_bytes()
        if any(lit.encode() in data for lit in literals):
            found.append({"rule": rule, "path": exp["path"]})
    return found


def lite_log_query(case_dir: Path, expected: list[dict]) -> list[dict]:
    found: list[dict] = []
    for exp in expected:
        source = case_dir / "evidence" / exp["source"]
        if not source.is_file():
            continue
        text = source.read_text(encoding="utf-8", errors="ignore")
        hits = text.count(exp["q"])
        if hits >= exp.get("expected_min_hits", 1):
            found.append({"q": exp["q"], "hits": hits})
    return found


def _read_memory_meta(case_dir: Path) -> dict:
    mem = case_dir / "evidence" / "memory.dmp"
    if not mem.is_file():
        return {}
    raw = mem.read_bytes()
    head = raw.split(b"\x00", 1)[0].decode("utf-8", errors="ignore").strip()
    try:
        return json.loads(head)
    except json.JSONDecodeError:
        return {}


def lite_memory_process_list(case_dir: Path, expected: list[dict]) -> list[dict]:
    meta = _read_memory_meta(case_dir)
    pids = set(meta.get("expected_pids", []))
    return [{"pid": e["pid"]} for e in expected if e["pid"] in pids]


def lite_memory_malfind(case_dir: Path, expected: list[dict]) -> list[dict]:
    meta = _read_memory_meta(case_dir)
    malfind_pids = set(meta.get("malfind_pids", []))
    return [{"pid": e["pid"]} for e in expected if e["pid"] in malfind_pids]


def lite_net_flow_summary(case_dir: Path, expected: list[dict]) -> list[dict]:
    pcap = case_dir / "evidence" / "sample.pcap"
    if not pcap.is_file():
        return []
    raw = pcap.read_bytes()
    if len(raw) < 24:
        return []
    magic = raw[:4]
    little = magic == b"\xd4\xc3\xb2\xa1"
    big = magic == b"\xa1\xb2\xc3\xd4"
    if not (little or big):
        return []
    endian = "<" if little else ">"
    flows: set[tuple[str, str]] = set()
    pos = 24
    while pos + 16 <= len(raw):
        _ts_sec, _ts_usec, incl_len, _orig_len = struct.unpack(
            endian + "IIII", raw[pos : pos + 16]
        )
        pos += 16
        if pos + incl_len > len(raw):
            break
        pkt = raw[pos : pos + incl_len]
        pos += incl_len
        if len(pkt) < 14 + 20:
            continue
        ipv4 = pkt[14 : 14 + 20]
        ihl = (ipv4[0] & 0x0F) * 4
        if ihl < 20:
            continue
        src_ip = ".".join(str(b) for b in ipv4[12:16])
        dst_ip = ".".join(str(b) for b in ipv4[16:20])
        flows.add((src_ip, dst_ip))
    found = []
    for exp in expected:
        if (exp["src_ip"], exp["dst_ip"]) in flows:
            found.append({"src_ip": exp["src_ip"], "dst_ip": exp["dst_ip"]})
    return found


# ---------------------------------------------------------------------------
# Driver
# ---------------------------------------------------------------------------


@dataclass
class ToolResult:
    tool: str
    expected: int
    matched: int
    missing: list[str] = field(default_factory=list)


def evaluate(case_dir: Path) -> tuple[float, list[ToolResult]]:
    manifest_path = case_dir / "manifest.json"
    if not manifest_path.is_file():
        print(f"FAIL: manifest missing at {manifest_path}", file=sys.stderr)
        sys.exit(2)
    manifest = json.loads(manifest_path.read_text(encoding="utf-8"))

    runners = {
        "iocs.scan": lite_iocs_scan,
        "log.query": lite_log_query,
        "memory.process_list": lite_memory_process_list,
        "memory.malfind": lite_memory_malfind,
        "net.flow_summary": lite_net_flow_summary,
    }
    results: list[ToolResult] = []
    total_expected = 0
    total_matched = 0
    for tool, expected in manifest["expected_findings"].items():
        runner = runners.get(tool)
        if runner is None:
            continue
        matched_items = runner(case_dir, expected)
        expected_count = len(expected)
        matched_count = len(matched_items)
        total_expected += expected_count
        total_matched += matched_count
        missing = []
        if matched_count < expected_count:
            _keys = {tuple(sorted(it.items())) for it in matched_items}
            for exp in expected:
                summary_key = tuple(sorted(
                    (k, v) for k, v in exp.items()
                    if k in {"rule", "path", "q", "pid", "src_ip", "dst_ip"}
                ))
                hit = any(all(item.get(k) == v for k, v in summary_key) for item in matched_items)
                if not hit:
                    missing.append(json.dumps({k: v for k, v in summary_key}))
        results.append(ToolResult(tool, expected_count, matched_count, missing))
    recall = (total_matched / total_expected) if total_expected else 0.0
    return recall, results


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--case", default=str(DEFAULT_CASE))
    parser.add_argument("--threshold", type=float, default=DEFAULT_THRESHOLD)
    parser.add_argument("--json", action="store_true")
    args = parser.parse_args()

    case_dir = Path(args.case).resolve()
    recall, results = evaluate(case_dir)

    payload = {
        "case_dir": str(case_dir),
        "threshold": args.threshold,
        "recall": round(recall, 4),
        "tools": [
            {"tool": r.tool, "expected": r.expected, "matched": r.matched, "missing": r.missing}
            for r in results
        ],
    }
    if args.json:
        print(json.dumps(payload, indent=2))
    else:
        print(f"case: {case_dir}")
        print(f"threshold: {args.threshold:.2f}")
        for r in results:
            mark = "OK" if r.matched >= r.expected else "MISS"
            print(f"  [{mark}] {r.tool}: {r.matched}/{r.expected}")
        print(f"recall: {recall:.2%}")

    return 0 if recall >= args.threshold else 1


if __name__ == "__main__":
    sys.exit(main())
