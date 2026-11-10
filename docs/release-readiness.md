# Initial release review

The release target is the standalone root module, github.com/goforj/validation. The docs module is repository tooling. The review preserves the fluent API, public error contracts, module layout, and minimum Go language version of 1.24.

## Review passes

| Pass | Scope and outcome |
| --- | --- |
| Consumer API and compatibility | Reviewed typed bounds, defined types, custom rule structs and functions, PATCH presence, conditional requirements, nested errors, zero values, ordering, ownership, cancellation, and concurrent reuse. Added regression tests for nested concurrent execution and secret redaction in serialized errors. No public API change was needed. |
| Correctness and malformed input | Reproduced acceptance of unbracketed IPv6 authorities and invalid percent escapes in URL queries. Corrected both under the existing URL syntax contract, including consistent IPv4 and IPv6 authority spelling across Go versions. Added direct regressions, Base64 block-boundary tests, and format fuzzing. |
| Security and resources | Removed decoded-payload allocations from Base64 validation using fixed scratch storage and strict standard-library decoding. Removed the unnecessary element traversal for slices without Each. Reviewed error classification, default message redaction, borrowed inputs, immutable builder branches, and dependency ownership. Vulnerability scanning covers both Go modules on a patched toolchain. |
| Documentation and release operations | Checked canonical programs, generated README navigation and parity, module inventory, tidy checks, exact coverage accounting, public installation, CI, and Codecov. Strengthened coverage parsing and checksum-file parity checks. Added a reproducible public-consumer check that executes all nine canonical programs against the pushed module with no local replacements. |

The runtime behavior correction rejects malformed URLs that older revisions accepted. Valid URLs following the documented format remain supported. Source/API, configuration, persisted data, and the minimum Go version are unchanged. This library has no persisted data or operational migration.

## Verification

The [quality report](quality.json) records the tested source fingerprint, platform, toolchain, exact statement counts, and executed Go tests and seeds. Unit coverage is enforced at 100% for the runtime, examples, and documentation tooling. Python tests separately verify the coverage-reporting gates.

The review also executed both modules on Go 1.24.0, the first version allowed by the module directive. Public-consumer checks were executed on Go 1.24.13 and Go 1.27.3. Race tests, vet, tidy parity, module inventory, generated parity, and vulnerability checks cover the relevant modules. CI repeats the checks and public installation on the pushed revision; inspect [CI](https://github.com/goforj/validation/actions/workflows/ci.yml), [Security](https://github.com/goforj/validation/actions/workflows/security.yml), and [Codecov](https://app.codecov.io/gh/goforj/validation) for that revision before tagging.

Two 15-second fuzz campaigns completed without a failure: 2,442,505 Base64 inputs compared with the strict standard decoder, and 192,036 arbitrary inputs exercised across every format. The [Base64 log](review/2026-11-10/fuzz-base64.txt) and [format log](review/2026-11-10/fuzz-formats.txt) retain the executed evidence. These runs supplement the direct contract tests.

The [allocation samples before](review/2026-11-10/allocations-before.txt) and [after](review/2026-11-10/allocations-after.txt) use the same committed benchmark harness on Go 1.27.3, darwin/arm64. The baseline Base64 and slice implementations are from a26aaf5; the fixes and harness are in 57f3702. Each sample uses a 100 ms benchmark interval. Full validation of a prepared 4 MiB Base64 value changed from approximately 3 MiB of allocation per call to 120 bytes, matching the small-input case. The internal successful Base64 check also has a zero-allocation regression test. These short samples establish the allocation change rather than a general throughput claim.

## Release procedure

No unresolved source release blocker remains from these review passes. Require the chosen pushed candidate's CI, Security, Codecov upload, and public-consumer checks to pass. Publish only the root module when a version is authorized; the docs module does not need a tag. Verify the resulting version independently through the public Go proxy before updating installation instructions to advertise it. Tagging and publication remain separate from this readiness review.
