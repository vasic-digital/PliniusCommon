# PliniusCommon

> **Status (round 213, 2026-05-19):** FUNCTIONAL — `go build ./...` exits 0;
> `go test -count=1 -race -p 1 ./pkg/...` all-green across 5 packages
> (config, errors, grpcclient, i18n, types); 36 i18n bundle keys seeded;
> CONST-035 anti-bluff Challenge present at
> `challenges/pliniuscommon_validation_challenge.sh` with paired mutation
> proof.

Foundational common library for the Plinius Go service family:
config validators, structured error types, gRPC client primitives,
i18n translator seam, and shared value types.

---

## Status banner

| Surface              | State           | Evidence                                                                          |
|----------------------|-----------------|-----------------------------------------------------------------------------------|
| Compile              | green           | `go build ./...` exits 0                                                          |
| Unit tests (race)    | green           | `GOMAXPROCS=2 nice -n 19 go test -count=1 -race -p 1 ./pkg/...` — 5 packages PASS |
| Bench coverage       | present         | `*_bench_test.go` in config / errors / grpcclient                                 |
| i18n bundle (CONST-046) | seeded — 36 keys | `pkg/i18n/bundles/active.en.yaml`                                              |
| Challenge (CONST-035) | present + mutation | `challenges/pliniuscommon_validation_challenge.sh`                            |
| CONST-033 host-power | source clean    | `challenges/scripts/no_suspend_calls_challenge.sh`                                |
| Governance cascade   | inherits root   | constitution submodule HEAD pinned by parent project                              |

---

## Purpose

PliniusCommon is the **foundational utility library** consumed by every
member of the Plinius Go service family — AutoTemp, HyperTune, I-LLM,
Veritas, LeakHub, Claritas, Ouroborous, GandalfSolutions — and
transitively by the HelixAgent / HelixCode platform binaries through
those modules. It is intentionally **project-not-aware** per CONST-051(B):
no consuming-project hostname, asset path, or runtime assumption may
appear in this tree. All locale resolution flows through the
injectable `i18n.Translator` seam (`pkg/i18n/translator.go`); the
default `NoopTranslator` returns keys verbatim so legacy programmatic
contracts (sentinel errors, ErrorCode constants, `fmt.Errorf` wrappers)
stay byte-for-byte stable.

---

## API surface

### `pkg/config` — functional-options configuration

```go
type Config struct {
    ServiceName, Address                                               string
    Timeout, ConnectionTimeout, RetryBackoff, MaxRetryBackoff          time.Duration
    MaxRetries, MaxRecvMsgSize, MaxSendMsgSize                         int
    EnableTLS, InsecureSkipVerify                                      bool
    TLSCertPath, TLSKeyPath, TLSCAPath, TLSServerName                  string
    KeepaliveTime, KeepaliveTimeout                                    time.Duration
    AuthToken, Compression                                             string
    Metadata                                                           map[string]string
}

func New(serviceName string, opts ...Option) *Config       // defaults baked in
func FromEnv(serviceName string) *Config                   // env-var overlay
func FromFile(path, serviceName string) (*Config, error)   // YAML overlay
func (c *Config) Validate() error                          // 7 invariants
```

**Validators (7) — all surface bundle-key-paired diagnostics:**

| Invariant                            | Bundle key                                                       |
|--------------------------------------|------------------------------------------------------------------|
| `service_name` non-empty             | `pliniuscommon_config_err_service_name_required`                 |
| `address` non-empty                  | `pliniuscommon_config_err_address_required`                      |
| `timeout` > 0                        | `pliniuscommon_config_err_timeout_must_be_positive`              |
| `connection_timeout` > 0             | `pliniuscommon_config_err_connection_timeout_must_be_positive`   |
| `max_retries` ≥ 0                    | `pliniuscommon_config_err_max_retries_cannot_be_negative`        |
| `retry_backoff` > 0                  | `pliniuscommon_config_err_retry_backoff_must_be_positive`        |
| TLS enabled ⇒ `tls_cert_path` set    | `pliniuscommon_config_err_tls_cert_path_required`                |

**File-load wrappers (5):**
`pliniuscommon_config_err_{read,parse,marshal_service_config,unmarshal_service_config,unmarshal_file}_failed`.

### `pkg/errors` — structured error types

```go
type ErrorCode string
type PliniusError struct {
    Code            ErrorCode
    Message, Service string
    Retryable        bool
    RetryAfterSeconds int
    Details          map[string]any
    cause            error
}

func New(code ErrorCode, service, message string) *PliniusError
func Wrap(code ErrorCode, service, message string, cause error) *PliniusError
func Is(err error, code ErrorCode) bool
func IsRetryableError(err error) bool
```

**Sentinel errors (3)** — keys
`pliniuscommon_err_{connection,timeout,cancelled}`.

**ErrorCode constants (16)** — keys
`pliniuscommon_errcode_{unavailable,invalid_argument,not_found,already_exists,
permission_denied,unauthenticated,resource_exhausted,failed_precondition,
aborted,out_of_range,unimplemented,internal,unknown,timeout,cancelled,connection}`.

Codes `unavailable`, `resource_exhausted`, `aborted`, `timeout`,
`cancelled`, `connection` are flagged `Retryable=true` by default;
`IsRetryableError(err)` returns true for any wrapped chain ending in
one of those codes.

### `pkg/grpcclient` — gRPC client wrapper

```go
type Client struct { /* opaque */ }
func New(cfg *config.Config) *Client
func (c *Client) Connect(ctx context.Context) error
func (c *Client) Close() error
func (c *Client) Connection() *grpc.ClientConn
func (c *Client) IsConnected() bool
func (c *Client) ContextWithMetadata(ctx context.Context) context.Context
```

Lifecycle diagnostics — keys
`pliniuscommon_grpc_err_{already_connected,not_connected,dial_failed,
close_failed,invocation_failed}`. Dial wraps with `errors.Wrap(
ErrCodeConnection, ...)` so transient connect failures are
`IsRetryableError`-true at the consumer layer.

### `pkg/i18n` — translator seam (CONST-046)

```go
type Translator interface {
    T(key string, params map[string]any) string
}
type NoopTranslator struct{}                                  // returns key verbatim
```

Project-side wiring example (NOT in this submodule — CONST-051(B)):

```go
// In HelixCode-side adapter package
type yamlTranslator struct{ msgs map[string]string }
func (y *yamlTranslator) T(key string, _ map[string]any) string {
    if v, ok := y.msgs[key]; ok { return v }
    return key // fall back per Translator contract
}
```

### `pkg/types` — shared value types

Cross-cutting structs reused by the 8 sibling Plinius modules.

---

## Build & test

```bash
# Compile sanity
GOMAXPROCS=2 nice -n 19 go build ./...

# Full test run with race detector (resource-capped per CLAUDE.md §Hard rules)
GOMAXPROCS=2 nice -n 19 ionice -c 3 go test -count=1 -p 1 -race ./pkg/...

# Anti-bluff Challenge — round 213 §11.4 deliverable
bash challenges/pliniuscommon_validation_challenge.sh

# CONST-033 host-power source-tree scan
bash challenges/scripts/no_suspend_calls_challenge.sh
```

All four MUST exit 0 before a commit touching this submodule is allowed.

---

## Governance pointers

- **Constitution** — see `CONSTITUTION.md` in this submodule (cascaded
  anchors CONST-035 / CONST-046 / CONST-047 / CONST-048 / CONST-049 /
  CONST-050 / CONST-051 / CONST-052 / CONST-053 / CONST-054 / CONST-055
  / CONST-056 / CONST-057 / CONST-058 / CONST-059 / CONST-060 /
  CONST-061).
- **Agent manual** — `CLAUDE.md` (mirrored in `AGENTS.md`).
- **Test coverage ledger** — `docs/test-coverage.md` (CONST-050(B)
  matrix, refreshed round 213).
- **Host power hard ban** — `docs/HOST_POWER_MANAGEMENT.md`
  (CONST-033 reference).

The single source of truth for universal rules is the constitution
submodule at `constitution/{Constitution,CLAUDE,AGENTS}.md` of every
consuming project — never edit this submodule's governance files to
override them per CONST-059.

---

## Module path

```go
import "digital.vasic.pliniuscommon"
```

## Lineage

Originated as a research scaffold imported into HelixAgent on
2026-04-21; graduated to functional status the same day. Round 135
(2026-05-15) seeded the 36-key i18n bundle infrastructure to satisfy
CONST-046 without breaking programmatic contracts. Round 213
(2026-05-19) ships this expanded README, the `docs/test-coverage.md`
ledger, and the `pliniuscommon_validation_challenge.sh` anti-bluff
Challenge per the verbatim 2026-05-19 operator mandate:

> "all existing tests and Challenges do work in anti-bluff manner —
> they MUST confirm that all tested codebase really works as expected!"

## License

Apache-2.0
