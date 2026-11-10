"""Regression tests for exact coverage and execution evidence."""

import json
import tempfile
import unittest
import xml.etree.ElementTree as ET
from pathlib import Path

from report import badge, coverage, fingerprint, passing_tests


class ReportTests(unittest.TestCase):
    """Check evidence parsing independently from the report produced by the real suite."""

    def test_overlapping_profiles(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            one, two = root / "one", root / "two"
            one.write_text("mode: atomic\np/file.go:1.1,2.1 2 0\np/file.go:3.1,4.1 1 1\n")
            two.write_text("mode: atomic\np/file.go:1.1,2.1 2 4\n")
            blocks, total = coverage([one, two])
            self.assertEqual(total, 3)
            self.assertEqual(len(blocks), 2)

    def test_incomplete_or_invalid_profiles(self):
        with tempfile.TemporaryDirectory() as directory:
            profile = Path(directory) / "profile"
            for content in ("", "mode: set\n", "mode: atomic\n", "mode: atomic\np/file.go:1.1,2.1 1000 999\np/file.go:3.1,4.1 1 0\n"):
                profile.write_text(content)
                with self.assertRaises(ValueError):
                    coverage([profile])

    def test_execution_counts(self):
        with tempfile.TemporaryDirectory() as directory:
            events = Path(directory) / "events"
            records = [{"Action": "pass", "Test": "TestParent"}, {"Action": "pass", "Test": "TestParent/Child"}, {"Action": "pass"}, {"Action": "skip", "Test": "TestSkipped"}]
            events.write_text("\n".join(json.dumps(event) for event in records))
            self.assertEqual(passing_tests([events]), 2)
            events.write_text('{"Action":"fail"}\n')
            with self.assertRaises(ValueError):
                passing_tests([events])

            events.write_text('{"Action":"pass"}\n')
            with self.assertRaises(ValueError):
                passing_tests([events])

    def test_each_profile_must_be_complete_and_consistent(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            one, two = root / "one", root / "two"
            one.write_text("mode: atomic\np/file.go:1.1,2.1 2 1\n")
            for content in (
                "mode: atomic\n",
                "mode: atomic\np/file.go:1.1,2.1 3 1\n",
                "mode: atomic\np/other.go:1.1,2.1 -1 1\n",
                "mode: atomic\np/other.go:1.1,2.1 1 -1\n",
            ):
                two.write_text(content)
                with self.assertRaises(ValueError):
                    coverage([one, two])

    def test_fingerprint_ignores_projections(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            source = root / "source.go"
            source.write_text("package source\n")
            before = fingerprint(root)
            (root / "README.md").write_text("generated")
            (root / ".artifacts").mkdir()
            (root / ".artifacts/generated.go").write_text("ignored")
            self.assertEqual(before, fingerprint(root))
            source.write_text("package changed\n")
            self.assertNotEqual(before, fingerprint(root))
            before = fingerprint(root)
            (root / "go.sum").write_text("dependency checksum\n")
            self.assertNotEqual(before, fingerprint(root))

    def test_badges_use_evidence(self):
        self.assertIn("100%", badge("unit coverage", "100%"))
        self.assertIn("82", badge("tests", "82"))
        tests = ET.fromstring(badge("tests", "111"))
        coverage = ET.fromstring(badge("unit coverage", "100%"))
        self.assertLess(int(tests.attrib["width"]), 100)
        self.assertLess(int(tests.attrib["width"]), int(coverage.attrib["width"]))
        self.assertEqual(tests.attrib["height"], "20")


if __name__ == "__main__":
    unittest.main()
