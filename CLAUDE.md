# CLAUDE.md -- digital.vasic.pliniuscommon

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
