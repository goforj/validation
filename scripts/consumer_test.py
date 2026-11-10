"""Regression checks for public-consumer isolation and failure handling."""

import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import unittest


class ConsumerTests(unittest.TestCase):
    """Exercise the shell workflow without contacting external services."""

    def run_consumer(self, mode="success"):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / "scripts").mkdir()
            (root / "examples").mkdir()
            (root / "bin").mkdir()
            shutil.copyfile(Path(__file__).with_name("verify-consumer.sh"), root / "scripts/verify-consumer.sh")
            fake_go = root / "bin/go"
            fake_go.write_text("""#!/usr/bin/env python3
import json, os, sys
args = sys.argv[1:]
with open(os.environ['CONSUMER_COMMAND_LOG'], 'a') as log:
    log.write(json.dumps({'args': args, 'cache': os.environ.get('GOMODCACHE'), 'proxy': os.environ.get('GOPROXY'), 'sumdb': os.environ.get('GOSUMDB'), 'private': os.environ.get('GOPRIVATE'), 'work': os.environ.get('GOWORK')}) + '\\n')
mode = os.environ['CONSUMER_TEST_MODE']
if args[:3] == ['list', '-m', '-f']:
    if mode == 'lookup-failure': sys.exit(1)
    print('v0.0.0-20261110172507-727f1e547319')
elif args[:1] == ['get']:
    if mode == 'download-failure': sys.exit(1)
elif args == ['list', '-m', '-json', 'all']:
    module = {'Path': 'github.com/goforj/validation', 'Version': 'v0.0.0-20261110172507-727f1e547319'}
    if mode == 'replacement': module['Replace'] = {'Dir': '/local/checkout'}
    if mode == 'version-mismatch': module['Version'] = 'v0.1.0'
    if mode == 'missing-module': module['Path'] = 'example.com/other'
    print(json.dumps(module))
""")
            fake_go.chmod(0o755)
            log = root / "commands.jsonl"
            env = dict(os.environ, PATH=str(root / "bin") + os.pathsep + os.environ["PATH"],
                       CONSUMER_COMMAND_LOG=str(log), CONSUMER_TEST_MODE=mode,
                       GOPROXY="https://untrusted.invalid,direct", GOSUMDB="off",
                       GOPRIVATE="*", GOWORK="/unrelated/go.work")
            result = subprocess.run(["bash", str(root / "scripts/verify-consumer.sh"), "727f1e547319"],
                                    env=env, capture_output=True, text=True)
            calls = [json.loads(line) for line in log.read_text().splitlines()]
            return result, calls

    def test_public_download_uses_separate_cache_and_strict_proxy(self):
        result, calls = self.run_consumer()
        self.assertEqual(result.returncode, 0, result.stderr)
        lookup = next(call for call in calls if call["args"][:3] == ["list", "-m", "-f"])
        download = next(call for call in calls if call["args"][0] == "get")
        self.assertNotEqual(lookup["cache"], download["cache"])
        self.assertEqual(lookup["proxy"], "https://proxy.golang.org,direct")
        self.assertEqual(download["proxy"], "https://proxy.golang.org")
        self.assertEqual(download["args"][1], "github.com/goforj/validation@v0.0.0-20261110172507-727f1e547319")
        self.assertEqual(download["sumdb"], "sum.golang.org")
        self.assertEqual(download["private"], "")
        self.assertEqual(download["work"], "off")
        self.assertTrue(any(call["args"][:1] == ["test"] for call in calls))
        self.assertEqual({call["cache"] for call in calls if call["args"] == ["clean", "-modcache"]},
                         {lookup["cache"], download["cache"]})
        self.assertFalse(Path(download["cache"]).parent.exists())

    def test_lookup_failure_stops_before_download(self):
        result, calls = self.run_consumer("lookup-failure")
        self.assertNotEqual(result.returncode, 0)
        self.assertFalse(any(call["args"][0] in ("get", "test") for call in calls))

    def test_proxy_download_failure_has_no_direct_fallback(self):
        result, calls = self.run_consumer("download-failure")
        self.assertNotEqual(result.returncode, 0)
        downloads = [call for call in calls if call["args"][0] == "get"]
        self.assertEqual(len(downloads), 1)
        self.assertEqual(downloads[0]["proxy"], "https://proxy.golang.org")
        self.assertFalse(any(call["args"][0] == "test" for call in calls))

    def test_invalid_module_selection_stops_before_examples(self):
        for mode in ("replacement", "version-mismatch", "missing-module"):
            with self.subTest(mode=mode):
                result, calls = self.run_consumer(mode)
                self.assertNotEqual(result.returncode, 0)
                self.assertFalse(any(call["args"][0] == "test" for call in calls))
