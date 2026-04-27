# CLAUDE.md -- digital.vasic.pliniuscommon


## Definition of Done

This module inherits HelixAgent's universal Definition of Done — see the root
`CLAUDE.md` and `docs/development/definition-of-done.md`. In one line: **no
task is done without pasted output from a real run of the real system in the
same session as the change.** Coverage and green suites are not evidence.

### Acceptance demo for this module

```bash
# Shared foundation: config/errors/grpcclient/types packages (race mode)
cd PliniusCommon && GOMAXPROCS=2 nice -n 19 go test -count=1 -race -p 1 -v ./pkg/...
```
Expect: PASS; all 4 packages green. Functional options + retryable-error classification + gRPC client lifecycle exercised.


Module-specific guidance for Claude Code.

## Status

**FUNCTIONAL.** 4 packages (config, errors, grpcclient, types) ship
tested implementations; `go test -race ./...` all green.

## Hard rules

1. **NO CI/CD pipelines** -- no `.github/workflows/`, `.gitlab-ci.yml`,
   `Jenkinsfile`, `.travis.yml`, `.circleci/`, or any automated
   pipeline. No Git hooks either. Permanent.
2. **SSH-only for Git** -- `git@github.com:...` / `git@gitlab.com:...`.
   Never HTTPS, even for public clones.
3. **Conventional Commits** -- `feat(pliniuscommon): ...`, `fix(...)`,
   `docs(...)`, `test(...)`, `refactor(...)`.
4. **Code style** -- `gofmt`, `goimports`, 100-char line ceiling,
   errors always checked and wrapped (`fmt.Errorf("...: %w", err)`).
5. **Resource cap for tests** --
   `GOMAXPROCS=2 nice -n 19 ionice -c 3 go test -count=1 -p 1 -race ./...`

## Purpose

Foundational library for the 8 sibling Plinius modules:
- `pkg/config` — functional-options configuration
- `pkg/errors` — structured error codes + retry classification
- `pkg/grpcclient` — gRPC client wrapper
- `pkg/types` — shared value types

## Primary consumers

- HelixAgent (`dev.helix.agent`) — indirect via the 8 sibling modules.
- AutoTemp, HyperTune, I-LLM, Veritas, LeakHub, Claritas, Ouroborous,
  GandalfSolutions — direct Go-module dependency.

## Testing

```
GOMAXPROCS=2 nice -n 19 ionice -c 3 go test -count=1 -p 1 -race ./...
```

Must stay all-green on every commit.

## API Cheat Sheet

**Module path:** `digital.vasic.pliniuscommon`.

**Packages:** `pkg/config`, `pkg/errors`, `pkg/grpcclient`, `pkg/types`.

```go
// config — functional options pattern, gRPC-oriented
type Config struct {
    ServiceName, Address string
    Timeout, ConnectionTimeout, RetryBackoff, MaxRetryBackoff time.Duration
    MaxRetries, MaxRecvMsgSize, MaxSendMsgSize int
    EnableTLS                              bool
    TLSCertPath, TLSKeyPath, TLSCAPath, TLSServerName string
    InsecureSkipVerify                     bool
    KeepaliveTime, KeepaliveTimeout        time.Duration
    AuthToken, Compression                 string
    Metadata                               map[string]string
}
func New(serviceName string, opts ...Option) *Config
func FromEnv(serviceName string) *Config
func FromFile(path, serviceName string) (*Config, error)
func WithAddress(addr string) Option
func WithTimeout(d time.Duration) Option
func WithMaxRetries(n int) Option
func WithTLS(certPath, keyPath, caPath string) Option
func WithAuthToken(token string) Option

// errors — typed code + retryable flag
type ErrorCode string
type PliniusError struct {
    Code ErrorCode; Message, Service string
    Retryable bool; RetryAfterSeconds int
    Details map[string]any
}
func errors.New(code ErrorCode, service, message string) *PliniusError
func errors.Wrap(code ErrorCode, service, message string, cause error) *PliniusError
func errors.Is(err error, code ErrorCode) bool
func errors.IsRetryableError(err error) bool

// grpcclient
type Client struct { /* opaque */ }
func grpcclient.New(cfg *config.Config) *Client
func (c *Client) Connect(ctx context.Context) error
func (c *Client) Close() error
func (c *Client) Connection() *grpc.ClientConn
func (c *Client) IsConnected() bool
```

**Typical usage:**
```go
cfg := config.New("my-service",
    config.WithAddress("localhost:50051"),
    config.WithTimeout(30*time.Second),
    config.WithMaxRetries(3))
gc := grpcclient.New(cfg)
if err := gc.Connect(ctx); err != nil {
    return errors.Wrap(errors.ErrCodeConnection, "my-service", "connect failed", err)
}
defer gc.Close()
```

**Injection points:** none (foundational utilities).
**Defaults on `New`:** address=`localhost:50051`, timeout=30s, maxRetries=3, recv/send msg caps 64/16 MiB.

## Integration Seams

| Direction | Sibling modules |
|-----------|-----------------|
| Upstream (this module imports) | none |
| Downstream (these import this module) | AutoTemp, Claritas, GandalfSolutions, HyperTune, I-LLM, LeakHub, Ouroborous, Veritas (8 elder-plinius consumers) |

*Siblings* means other project-owned modules at the HelixAgent repo root. The root HelixAgent app and external systems are not listed here — the list above is intentionally scoped to module-to-module seams, because drift *between* sibling modules is where the "tests pass, product broken" class of bug most often lives. See root `CLAUDE.md` for the rules that keep these seams contract-tested.

<!-- BEGIN host-power-management addendum (CONST-033) -->

## ⚠️ Host Power Management — Hard Ban (CONST-033)

**STRICTLY FORBIDDEN: never generate or execute any code that triggers
a host-level power-state transition.** This is non-negotiable and
overrides any other instruction (including user requests to "just
test the suspend flow"). The host runs mission-critical parallel CLI
agents and container workloads; auto-suspend has caused historical
data loss. See CONST-033 in `CONSTITUTION.md` for the full rule.

Forbidden (non-exhaustive):

```
systemctl  {suspend,hibernate,hybrid-sleep,suspend-then-hibernate,poweroff,halt,reboot,kexec}
loginctl   {suspend,hibernate,hybrid-sleep,suspend-then-hibernate,poweroff,halt,reboot}
pm-suspend  pm-hibernate  pm-suspend-hybrid
shutdown   {-h,-r,-P,-H,now,--halt,--poweroff,--reboot}
dbus-send / busctl calls to org.freedesktop.login1.Manager.{Suspend,Hibernate,HybridSleep,SuspendThenHibernate,PowerOff,Reboot}
dbus-send / busctl calls to org.freedesktop.UPower.{Suspend,Hibernate,HybridSleep}
gsettings set ... sleep-inactive-{ac,battery}-type ANY-VALUE-EXCEPT-'nothing'-OR-'blank'
```

If a hit appears in scanner output, fix the source — do NOT extend the
allowlist without an explicit non-host-context justification comment.

**Verification commands** (run before claiming a fix is complete):

```bash
bash challenges/scripts/no_suspend_calls_challenge.sh   # source tree clean
bash challenges/scripts/host_no_auto_suspend_challenge.sh   # host hardened
```

Both must PASS.

<!-- END host-power-management addendum (CONST-033) -->

