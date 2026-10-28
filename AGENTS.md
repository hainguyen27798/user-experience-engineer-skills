# Repository Guidelines

## Project Structure

This repository publishes the dependency-free Go CLI `create-ux-engineer`. `cmd/create-ux-engineer/main.go` owns flags and prompts. Workspace orchestration, safety checks, downloads, and official skill installation belong in `internal/scaffold/`. Generated project files belong in `templates/` and are embedded in the binary. Do not author substitutes for third-party skills; install their unmodified directories from official repositories and record resolved commits in `skills-lock.json`.

## Commands

- `go run ./cmd/create-ux-engineer --help` — inspect CLI usage.
- `go test ./...` — run the offline test suite.
- `go test -race ./...` — run concurrency-safe verification.
- `go build ./cmd/create-ux-engineer` — build the CLI.
- `./scripts/build-release.sh` — build all GitHub Release assets locally.
- `./release.sh` — verify, bump the patch version, commit/tag, and atomically push a CLI release.

A real initialization writes the Nuxt scratch template and downloads current npm packages, framework references, and official skills. Use a disposable directory for full-stack initialization. UI-only initialization may also add a new `<project>-ui` service to an existing worktree and must preserve its existing files.

## Style

Use standard Go formatting and the standard library unless a dependency removes meaningful lifecycle complexity. Keep command execution behind `Runner`, network retrieval behind `Downloader`, and official skill installation behind `SkillInstaller` so routine tests remain offline. Markdown uses ATX headings and fenced code blocks with language identifiers.

## Testing

Tests live beside their Go package as `*_test.go`. Mock external commands and downloads. Cover observable behavior: empty-directory protection, command arguments, optional FastAPI setup, generated workspace files, root agent configuration, official skill lock data, current reference destinations, and the Taste temporary-artifact rule.

## Initializer Invariants

The CLI derives the project name from the target directory and always creates a new `<project>-ui`. Full-stack initialization and `--api` require an empty worktree; UI-only initialization may run in an existing worktree when `<project>-ui` does not exist. In that mode, preserve existing root and backend files, merge only missing ignore patterns into `.gitignore`, and do not generate root README, AGENTS, or release files that could misdescribe the existing project. Preserve an existing `.git` directory or gitfile and skip `git init`. A failure may leave a partial project for inspection; never silently delete user files. Nuxt dependencies use latest stable npm tags. Caveman, UI UX Pro Max, and Taste come directly from their official default-branch HEAD; preserve their contents and record the resolved commit and hash. Taste and other design tools must keep temporary artifacts outside generated repositories.

## CLI Release Invariants

Ordinary builds never mutate `VERSION`. Only the root `release.sh` increments it, locally cross-builds assets, and publishes them with `gh`; releases do not depend on GitHub Actions. It must preserve a failed push or asset upload as a retryable local release. A tag must equal `v$(cat VERSION)`. GitHub Release asset names stay unversioned so the documented `/releases/latest/download/install.sh` URL remains stable. The installer verifies SHA-256 before atomically replacing the binary.
