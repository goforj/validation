"""Regression tests for deterministic, offline currency selector generation."""

from pathlib import Path
import subprocess
import tempfile
import unittest
from unittest.mock import patch

from currencies import generate, read_codes, render


class CurrencyGeneratorTests(unittest.TestCase):
    """Check the source contract and exported constants alongside selector methods."""

    def test_source_is_validated(self):
        self.assertEqual(read_codes("# source\nUSD\n\n EUR \n"), ["EUR", "USD"])
        for invalid in ("", "# only comments", "usd", "US", "USDX", "ＵＳＤ", "USD\nUSD"):
            with self.subTest(source=invalid), self.assertRaises(ValueError):
                read_codes(invalid)

    def test_constants_and_selectors_are_projected_together(self):
        source = render(["EUR", "USD"])
        self.assertIn('EUR CurrencyCode = "EUR"', source)
        self.assertIn('USD CurrencyCode = "USD"', source)
        self.assertIn('func (b CurrencyField[T]) EUR() CurrencyField[T] { return b.allow(EUR) }', source)
        self.assertIn('func (b CurrencyField[T]) USD() CurrencyField[T] { return b.allow(USD) }', source)

    def test_generation_is_deterministic_and_preserves_output_on_failure(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / 'testdata').mkdir()
            data = root / 'testdata/currency-codes.txt'
            data.write_text('USD\nEUR\n')
            generate(root)
            first = (root / 'currency_codes.go').read_text()
            generate(root)
            self.assertEqual(first, (root / 'currency_codes.go').read_text())
            data.write_text('invalid')
            with self.assertRaises(ValueError):
                generate(root)
            self.assertEqual(first, (root / 'currency_codes.go').read_text())
            data.write_text('USD\n')
            with patch('currencies.subprocess.check_output', side_effect=subprocess.CalledProcessError(1, 'gofmt')):
                with self.assertRaises(subprocess.CalledProcessError):
                    generate(root)
            self.assertEqual(first, (root / 'currency_codes.go').read_text())
