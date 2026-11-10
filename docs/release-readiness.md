# Release review

The release target is the standalone root module, github.com/goforj/validation. The docs module is repository tooling. The November 10 review covers the standalone API documented in [validation contracts](contracts.md). Existing public calls retain their defaults, module layout, and minimum Go language version of 1.24.

## Review passes

| Pass | Scope and outcome |
| --- | --- |
| Consumer API and compatibility | Reviewed typed bounds, defined types, custom validator types and functions, rule composition, optional pointers, JSON presence, dates, collection rules, policy scoping, concurrent reuse, and owned configuration. Additive builders and policies preserve existing calls and first-violation-per-field defaults. |
| Correctness and malformed input | Tested string and numeric boundaries, exact full-width divisibility, map key ordering and escaping, incomparable uniqueness keys, date parsing, missing/null/zero states, composition short circuits, operational error precedence, cancellation, and explicit message overrides. Original URL fixes reject unbracketed IPv6 authorities and invalid query escapes. |
| Security and resources | Standard and URL-safe Base64 use fixed scratch storage with strict canonical decoding. Successful checks have zero decoded-payload allocations, including a prepared 4 MiB input. Uniqueness checks honor cancellation for empty input and check cancellation before allocating their key map. Reviewed error redaction, borrowed values, lazy callbacks, immutable policies, formatter parameter ownership, and dependency ownership. Vulnerability scanning covers both modules on a patched toolchain. |
| Documentation and operations | Fourteen canonical programs run under the output-checking runner. README snippets and API navigation derive from those sources and GoDoc. Checked generated parity, exact coverage accounting, module inventory, tidy parity, public-consumer setup, CI, and Codecov upload configuration. |

The original URL correction rejects malformed values that earlier revisions accepted under the same documented syntax contract. The breadth expansion adds APIs and opt-in policies. Existing source/API calls, configuration, default collection behavior, and the minimum Go version remain supported. This utility has no persisted data or operational migration.

## Verification

The [quality report](quality.json) records the tested source fingerprint, platform, toolchain, exact statement counts, and executed Go tests and seeds. Unit coverage is enforced at 100% for the runtime, examples, and documentation tooling. Python tests separately verify the coverage-reporting gates and public-consumer cache and proxy isolation. Concurrent reuse is exercised under the race detector, including the new builders, composed rules, and branched validator policies.

Both modules were tested and vetted on Go 1.24.0, the first version allowed by the module directive. Race tests, vet, module inventory, generated parity, and vulnerability checks also ran on Go 1.27.3. Public-consumer verification resolves the pushed revision through the public Go proxy and checksum database in a fresh module cache, rejects local replacements, and executes all canonical examples outside the checkout. CI repeats that check on the minimum Go line and stable Go. Inspect the chosen revision's [CI](https://github.com/goforj/validation/actions/workflows/ci.yml), [Security](https://github.com/goforj/validation/actions/workflows/security.yml), and [Codecov](https://app.codecov.io/gh/goforj/validation) results before tagging.

The original 15-second fuzz campaigns completed without failure: 2,442,505 standard Base64 inputs compared with the strict standard decoder and 192,036 inputs across the original formats. See the [Base64 log](review/2026-11-10/fuzz-base64.txt) and [original format log](review/2026-11-10/fuzz-formats.txt). Additional 15-second campaigns covered 2,322,427 URL-safe Base64 inputs against the strict standard decoder and 736,117 inputs across every expanded format. See the [URL-safe Base64 log](review/2026-11-10/fuzz-base64-url.txt) and [expanded format log](review/2026-11-10/fuzz-expanded-formats.txt). These runs supplement direct boundary and configuration tests.

The [allocation samples before](review/2026-11-10/allocations-before.txt) and [after](review/2026-11-10/allocations-after.txt) use the same committed benchmark harness on Go 1.27.3, darwin/arm64. The baseline implementations are from a26aaf5; the original fixes and harness are in 57f3702. Each sample uses a 100 ms interval. Full validation of a prepared 4 MiB standard Base64 value changed from approximately 3 MiB of allocation per call to 120 bytes, matching the small-input case. The current direct zero-allocation tests cover both Base64 alphabets. These establish allocation behavior, not a general throughput ranking.

## v0.1.0 candidate review

A fresh review examined the current public API and all runtime implementations through four passes: consumer behavior and compatibility, correctness and cancellation, security and resources, and documentation and release operations. The runtime fixes have direct regression tests:

- Direct `Input.UnmarshalJSON` calls accepted non-JSON whitespace around null and replaced the previous value. Null recognition now permits only JSON whitespace; malformed input preserves the prior state.
- `Time.Equal` retained the expected timestamp in issue parameters. It now keeps equality targets private, matching the string and numeric equality contract. Time bounds still expose their configured limits.
- The decoder documentation now states that enclosing `DisallowUnknownFields` and `UseNumber` options do not propagate through `Input`'s custom unmarshaler. Tests cover typed integer precision and the documented decoder boundary.

These are runtime correctness and error-redaction corrections. Exported signatures, configuration, module layout, and minimum Go version are unchanged. No persisted data or operational migration is involved.

The final local gate passed all 1,035 instrumented statements, 157 Go tests and fuzz seeds, ten Python tests, and fourteen runnable programs. Both modules passed tests and vet on Go 1.24.0. Go 1.27.3 passed race tests, vet, tidy parity, module inventory, generated parity, and vulnerability scans. A 20-second [JSON input fuzz run](review/2026-11-10/fuzz-input-json.txt) exercised 2,324,210 inputs against the standard JSON scanner without a failure.

Fresh [allocation samples](review/2026-11-10/allocations-v0.1.0.txt) confirm that a prepared 4 MiB Base64 value still uses 120 bytes per full validation, matching the 256-byte case. The direct standard and URL-safe Base64 checks retain their zero-allocation regression tests. These measurements establish resource behavior and do not rank this library against others.

Codecov is active for this repository. CI uploads the unit report through OIDC, treats upload failures as errors, and configures project and patch targets at 100% without flag carryforward. The coverage badge links to the hosted report. Require a processed 100% report for the exact final pushed candidate alongside successful CI, Security, and public-consumer checks before publishing v0.1.0.

## Scope and release procedure

The standalone core covers typed values, common scalar and format constraints, collections, conditional and presence-aware inputs, dates, composition, structured errors, localization hooks, and pluggable application rules. File inspection, database adapters, request binding, coercion, schema export, maintained country/currency registries, and translation bundles remain application or adapter concerns.

No unresolved source blocker remains from these review passes. Require the chosen pushed candidate's CI, Security, Codecov upload, and public-consumer checks to pass. Publish only the root module when a version is authorized; the docs module does not need a tag. Verify the resulting version independently through the public Go proxy before advertising it in installation instructions. v0.1.0 is the first root-module release.
