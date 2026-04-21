# PliniusCommon

Foundational common library for the Plinius Go service family:
config, error types, gRPC client primitives, and shared value types.

## Status

- ✅ Compiles: `go build ./...` exits 0.
- ✅ Tests pass under `-race`: 4 packages, all green.
- ✅ Integration-ready: serves as a library dependency for the 8
  sibling Plinius modules (AutoTemp, HyperTune, I-LLM, Veritas,
  LeakHub, Claritas, Ouroborous, GandalfSolutions).

## Purpose

- `pkg/config` — functional-options configuration with env + file
  loading (`FromEnv`, `FromFile`, `Validate`).
- `pkg/errors` — structured error types with error codes
  (`ErrCodeUnavailable`, `ErrCodeInvalidArgument`, `ErrCodeNotFound`, …),
  retry-hint classification, and causal chain via `errors.Wrap` /
  `errors.Is`.
- `pkg/grpcclient` — gRPC client wrapper with connection timeout,
  retry backoff, TLS, auth-token metadata, and keepalive.
- `pkg/types` — shared value types used across the service family.

## Usage

```go
import (
    "context"

    "digital.vasic.pliniuscommon/pkg/config"
    "digital.vasic.pliniuscommon/pkg/errors"
    "digital.vasic.pliniuscommon/pkg/grpcclient"
)

cfg := config.New("autotemp",
    config.WithAddress("autotemp:50051"),
    config.WithTimeout(45 * time.Second),
    config.WithAuthToken(os.Getenv("PLINIUS_AUTH_TOKEN")),
)

client := grpcclient.New(cfg)
ctx := client.ContextWithMetadata(context.Background())
// … issue gRPC calls with `ctx` …

if err := someCall(); err != nil {
    if errors.IsRetryable(err) {
        // back off and retry
    }
}
```

## Module path

```go
import "digital.vasic.pliniuscommon"
```

## Lineage

Originated as a research scaffold imported into HelixAgent on
2026-04-21; graduated to functional status on the same day after
the sibling `digital.vasic.redteam` and `digital.vasic.normalize`
extractions proved the submodule flow. Historical research corpus
(unused) remains at
`docs/research/go-elder-plinius-v3/go-elder-plinius/go-plinius-common/`
inside the HelixAgent repository.

## License

Apache-2.0
