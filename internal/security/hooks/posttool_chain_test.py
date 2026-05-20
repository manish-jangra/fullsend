#!/usr/bin/env python3
"""Integration tests for post-tool hook chain ordering (unicode before secret redact)."""

from __future__ import annotations

import json
import subprocess
import sys
import unittest
from pathlib import Path

HOOKS_DIR = Path(__file__).parent
UNICODE_HOOK = str(HOOKS_DIR / "unicode_posttool.py")
SECRET_HOOK = str(HOOKS_DIR / "secret_redact_posttool.py")

PLAIN_PAT = "ghp_FAKEtesttoken000000000000000000000000"


def obfuscate_with_char(text: str, char: str) -> str:
    """Insert invisible character between each codepoint."""
    return char.join(text)


def run_hook(script: str, tool_result: str) -> tuple[int, str, str]:
    proc = subprocess.run(
        [sys.executable, script],
        input=json.dumps({"tool_name": "Read", "tool_result": tool_result}),
        capture_output=True,
        text=True,
        timeout=10,
    )
    return proc.returncode, proc.stdout, proc.stderr


def run_chain(tool_result: str) -> str:
    """Run unicode_posttool then secret_redact_posttool (correct sandbox order)."""
    rc, stdout, stderr = run_hook(UNICODE_HOOK, tool_result)
    if rc != 0:
        raise RuntimeError(f"unicode hook failed: rc={rc}, stderr={stderr}")
    if stdout.strip():
        out = json.loads(stdout)
        tool_result = out["tool_result"]

    rc, stdout, stderr = run_hook(SECRET_HOOK, tool_result)
    if rc != 0:
        raise RuntimeError(f"secret_redact hook failed: rc={rc}, stderr={stderr}")
    if stdout.strip():
        out = json.loads(stdout)
        return out["tool_result"]
    return tool_result


class TestPostToolChain(unittest.TestCase):
    def test_plain_pat_redacted_by_chain(self):
        result = run_chain(PLAIN_PAT)
        self.assertNotIn("ghp_FAKEtest", result)
        self.assertIn("...", result)

    def test_zero_width_obfuscated_pat_redacted_by_chain(self):
        obfuscated = obfuscate_with_char(PLAIN_PAT, "\u200c")
        result = run_chain(obfuscated)
        self.assertNotIn("ghp_FAKEtest", result)
        self.assertIn("...", result)

    def test_ltr_mark_obfuscated_pat_redacted_by_chain(self):
        obfuscated = obfuscate_with_char(PLAIN_PAT, "\u200e")
        result = run_chain(obfuscated)
        self.assertNotIn("ghp_FAKEtest", result)
        self.assertIn("...", result)

    def test_redact_alone_misses_zero_width_obfuscated_pat(self):
        obfuscated = obfuscate_with_char(PLAIN_PAT, "\u200c")
        rc, stdout, _ = run_hook(SECRET_HOOK, obfuscated)
        self.assertEqual(rc, 0)
        # secret_redact alone does not modify output when regex cannot match
        self.assertEqual(stdout.strip(), "")
        # Obfuscated token still present in source (would leak after unicode strips ZWNJ)
        self.assertIn("\u200c", obfuscated)


if __name__ == "__main__":
    unittest.main()
