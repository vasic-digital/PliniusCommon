# CLAUDE.md -- digital.vasic.pliniuscommon

Module-specific guidance for Claude Code.

## Status

**SCAFFOLD / WIP.** All exported method bodies return
`ErrCodeUnimplemented`. The module compiles but is not yet
functional. Phase-A implementation is a future milestone.

## Hard rules

1. **NO CI/CD pipelines** -- no `.github/workflows/`, `.gitlab-ci.yml`,
   `Jenkinsfile`, `.travis.yml`, `.circleci/`, or any automated
   pipeline. No Git hooks either. Permanent.
2. **SSH-only for Git** -- `git@github.com:...` / `git@gitlab.com:...`.
   Never HTTPS, even for public clones.
3. **Conventional Commits** -- `feat(pliniuscommon): ...`, `fix(...)`,
   `docs(...)`, `test(...)`, `refactor(...)`.
4. **Code style** -- `gofmt`, `goimports`, 100-char line ceiling,
   errors always checked and wrapped.
5. **Resource cap for tests** --
   `GOMAXPROCS=2 nice -n 19 ionice -c 3 go test -count=1 -p 1 -race ./...`

## Purpose (intended)

Common types/errors library (shared foundation for the 8 sibling modules).

## Primary consumer

HelixAgent (`dev.helix.agent`). See the consuming-side Phase-A spec
at `docs/superpowers/specs/2026-04-21-elder-plinius-phaseA-go-plinius-common.md` in the HelixAgent repository.
