#!/usr/bin/env python3
"""Merge real unit execution profiles and enforce exact statement coverage."""

import hashlib
import json
import subprocess
from html import escape
from pathlib import Path


def coverage(paths):
    """Coalesce repeated blocks from overlapping Go coverage instrumentation."""
    blocks = {}
    statement_counts = {}
    for path in paths:
        lines = path.read_text().splitlines()
        if not lines or lines[0] != "mode: atomic":
            raise ValueError(f"expected atomic coverage profile: {path}")
        profile_statements = 0
        for line in lines[1:]:
            location, statements, hits = line.split()
            statements, hits = int(statements), int(hits)
            if statements < 0 or hits < 0:
                raise ValueError(f"negative coverage count: {path}")
            if location in statement_counts and statement_counts[location] != statements:
                raise ValueError(f"inconsistent statement count: {location}")
            statement_counts[location] = statements
            profile_statements += statements
            key = (location, statements)
            blocks[key] = max(blocks.get(key, 0), hits)
        if not profile_statements:
            raise ValueError(f"empty coverage profile: {path}")
    total = sum(statements for _, statements in blocks)
    covered = sum(statements for (_, statements), hits in blocks.items() if hits)
    if not total or total != covered:
        missing = [location for (location, statements), hits in blocks.items() if statements and not hits]
        raise ValueError(f"coverage requires 100%: {covered}/{total} statements; uncovered: {missing}")
    return blocks, total


def passing_tests(paths):
    """Count completed test and subtest events rather than source declarations."""
    count = 0
    for path in paths:
        for line in path.read_text().splitlines():
            event = json.loads(line)
            if event["Action"] == "fail":
                raise ValueError(f"failed test event: {path}")
            if event["Action"] == "pass" and event.get("Test"):
                count += 1
    if not count:
        raise ValueError("no executed tests")
    return count


def fingerprint(root):
    """Bind reports to relevant source while excluding generated projections."""
    digest = hashlib.sha256()
    paths = sorted(p for p in root.rglob("*") if p.is_file() and ".git" not in p.parts and ".artifacts" not in p.parts and (p.suffix in {".go", ".py", ".sh"} or p.name in {"go.mod", "go.sum", "module-manifest.txt"}))
    for path in paths:
        digest.update(path.relative_to(root).as_posix().encode() + b"\0" + path.read_bytes() + b"\0")
    return digest.hexdigest()


def badge(label, value):
    """Render a compact repository-owned status badge from measured evidence."""
    left, right = len(label) * 7 + 14, len(value) * 7 + 14
    width = left + right
    label, value = escape(label), escape(value)
    return f'<svg xmlns="http://www.w3.org/2000/svg" width="{width}" height="20" role="img" aria-label="{label}: {value}"><title>{label}: {value}</title><clipPath id="r"><rect width="{width}" height="20" rx="3"/></clipPath><g clip-path="url(#r)"><path fill="#555" d="M0 0h{left}v20H0z"/><path fill="#4c1" d="M{left} 0h{right}v20H{left}z"/></g><g fill="#fff" text-anchor="middle" font-family="Verdana,DejaVu Sans,sans-serif" font-size="11"><text x="{left / 2:g}" y="14">{label}</text><text x="{left + right / 2:g}" y="14">{value}</text></g></svg>\n'


def main():
    """Write evidence only after both modules and executable programs pass."""
    root = Path(__file__).resolve().parent.parent
    artifacts = root / ".artifacts"
    profiles = [artifacts / name for name in ("runtime.out", "examples.out", "docs.out")]
    blocks, total = coverage(profiles)
    count = passing_tests([artifacts / "runtime-tests.json", artifacts / "docs-tests.json"])
    merged = "mode: atomic\n" + "".join(f"{location} {statements} {hits}\n" for (location, statements), hits in sorted(blocks.items()))
    (artifacts / "unit-coverage.out").write_text(merged)
    version, goos, goarch = subprocess.check_output(["go", "env", "GOVERSION", "GOOS", "GOARCH"], text=True).splitlines()
    report = {"source_sha256": fingerprint(root), "go_version": version, "goos": goos, "goarch": goarch, "unit_covered_statements": total, "unit_total_statements": total, "executed_tests": count, "modules": [".", "docs"], "example_programs": len(list((root / "examples").glob("*/main.go")))}
    (root / "docs/quality.json").write_text(json.dumps(report, indent=2) + "\n")
    (root / "docs/assets/coverage.svg").write_text(badge("unit coverage", "100%"))
    (root / "docs/assets/tests.svg").write_text(badge("tests", str(count)))
    print(f"unit coverage: {total}/{total} statements; {count} executed tests; {report['example_programs']} runnable programs")


if __name__ == "__main__":
    main()
