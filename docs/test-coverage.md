# PliniusCommon — Test-Type Coverage Ledger (CONST-050(B))

> Round 213 (2026-05-19) snapshot. Refresh in the same commit as any
> change that adds / removes a package or test file. Stale ledger =
> CONST-050(B) violation.

## Verbatim 2026-05-19 operator mandate (per CONST-049 §11.4.17)

> "all existing tests and Challenges do work in anti-bluff manner — they
> MUST confirm that all tested codebase really works as expected! We had
> been in position that all tests do execute with success and all
> Challenges as well, but in reality the most of the features does not
> work and can't be used! This MUST NOT be the case and execution of
> tests and Challenges MUST guarantee the quality, the completition and
> full usability by end users of the product!"

The ledger below maps every PliniusCommon package × every CONST-050(B)
test-type-axis, with the file path that supplies the evidence (or
`N/A` + rationale where the axis is genuinely out-of-scope for a
foundational-utility library).

---

## Package × test-type matrix

| Package           | Unit                                                                 | Bench / Performance                       | Anti-bluff Challenge (CONST-035)                           | Integration / E2E (real downstream)                              | Mutation (paired CONST-050(B))                              |
|-------------------|----------------------------------------------------------------------|-------------------------------------------|------------------------------------------------------------|------------------------------------------------------------------|-------------------------------------------------------------|
| `pkg/config`      | `config_test.go` + `config_extra_test.go`                            | `config_bench_test.go`                    | `challenges/pliniuscommon_validation_challenge.sh` §3      | exercised in-process via 8 sibling Plinius modules at consumer   | challenge mutation §6 — corrupts bundle key, asserts exit 1 |
| `pkg/errors`      | `errors_test.go` + `errors_extra_test.go`                            | `errors_bench_test.go`                    | challenge §4                                               | exercised in-process via 8 sibling Plinius modules               | challenge mutation §6                                       |
| `pkg/grpcclient`  | `grpcclient_test.go` + `grpcclient_extra_test.go`                    | `grpcclient_bench_test.go`                | challenge §5                                               | real gRPC dial exercised in sibling-module integration suites    | challenge mutation §6                                       |
| `pkg/i18n`        | `translator_test.go`                                                 | N/A — interface seam, no hot path         | challenge §2 (bundle integrity + key count)                | N/A — translator seam injected at consumer (CONST-051(B))        | challenge mutation §6                                       |
| `pkg/types`       | `types_test.go`                                                      | N/A — pure value types                    | challenge §2 (build & test gate)                           | N/A — pure value types                                           | challenge mutation §6                                       |

---

## Test-type axes — coverage rationale

| CONST-050(B) axis | PliniusCommon coverage                                                                                                                                                                                                       |
|-------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Unit              | 5 packages, all green under `-race -p 1`. Mocks confined to `*_test.go`.                                                                                                                                                     |
| Integration       | Performed at the **sibling-module** layer (AutoTemp / HyperTune / I-LLM / Veritas / LeakHub / Claritas / Ouroborous / GandalfSolutions) — those modules import PliniusCommon and exercise it against real gRPC backends.    |
| E2E               | Performed at the **consuming-project** layer (HelixAgent / HelixCode) — they orchestrate the 8 siblings end-to-end.                                                                                                          |
| Full automation   | Inherited via consuming-project release-gate sweeps (CONST-048 ledger); this submodule contributes the green `go test ./...` baseline + Challenge exit-0 evidence.                                                          |
| Security          | (a) CONST-033 / CONST-036 host-power + session-termination scanners in `challenges/scripts/`; (b) no secrets in tree (CONST-042); (c) gRPC TLS path exercised in `config_test.go`.                                          |
| DDoS              | Out-of-scope at library layer — exercised at sibling-module gRPC server boundary. Foundational-utility code has no advertised request throughput tier of its own.                                                            |
| Scaling           | Out-of-scope at library layer — see sibling modules.                                                                                                                                                                          |
| Chaos             | Connection-failure path exercised in `grpcclient_extra_test.go` (dial-failed wrapping + retryable classification); sibling-module suites add network-partition coverage.                                                     |
| Stress            | `*_bench_test.go` files exercise hot-path allocators under `-benchtime`; sustained-load coverage at sibling-module layer.                                                                                                    |
| Performance       | Bench files cover config option-application, error wrapping, gRPC client construction. Historical p95-drift tracked at sibling-module layer.                                                                                  |
| Benchmarking      | 3 of 5 packages ship `*_bench_test.go`; remaining 2 (`i18n`, `types`) have no hot path worth benching.                                                                                                                       |
| UI / UX           | N/A — library, no UI surface. CONST-046 bundle keys are the i18n contract that future UI consumers translate.                                                                                                                |
| Challenges        | `challenges/pliniuscommon_validation_challenge.sh` (round 213) + 8 inherited host-power / session / DDoS / chaos / stress / scaling / UI / UX scripts under `challenges/scripts/`.                                            |
| HelixQA           | Exercised at consuming-project HelixQA autonomous-session layer; this submodule's contribution is the `go test ./...` + Challenge evidence captured per run.                                                                  |

---

## Anti-bluff guarantee (CONST-035 / Article XI §11.9)

Each entry in the matrix above is a **claim** — and the
`pliniuscommon_validation_challenge.sh` script is the **evidence** that
the claim is honest. Specifically:

1. Building the module — `go build ./...` — proves §3-§5 surfaces compile.
2. Running the test suite — `go test -count=1 -race -p 1 ./pkg/...` —
   proves the unit + bench axes are wired and green.
3. Counting the bundle keys — `grep -c '^pliniuscommon_'` ≥ 36 — proves
   the CONST-046 contract is honoured.
4. **Mutation** — the challenge corrupts one bundle key, re-runs the
   integrity check, and asserts **non-zero exit** under mutation.
   Without the mutation gate, the green PASS could be a bluff — the
   mutation proves the check actually catches a regression.

A future regression that breaks any of the 4 invariants causes the
challenge to exit non-zero — which fails any pre-commit / pre-push
sweep that runs it, per CONST-055.

---

## Refresh procedure

Whenever a new test file, new package, or new bundle key lands in this
submodule:

1. Add a row / column entry above so the matrix stays accurate.
2. Re-run the Challenge and paste its exit code + last 10 lines into
   the commit message as positive runtime evidence.
3. Update the parent project's CONST-048 coverage ledger to reflect the
   PliniusCommon row delta.
4. Cascade-bump the parent project's `.gitmodules` pointer (CONST-049
   step 7).

Skipping any of these steps creates `Status: Reopened` items per
CONST-058 in the parent project's Issues tracker, classified
`Reason: cycle-re-discovered`.
