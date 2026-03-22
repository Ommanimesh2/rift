# rift Product Upgrade Plan: v1.1 → v2.0

## Context

rift is a "git diff for container images" — a Go CLI tool that compares two container images at the file level, showing added/removed/modified files with security analysis and CI/CD integration. It fills the gap left by Google's archived `container-diff`.

**Current state (v1.1):** Feature-complete core with clean modular architecture (6333 LOC, 168 tests, ~75% coverage). Supports remote registries, Docker daemon, OCI tarballs, 4 output formats, 7 security event types, and CI/CD exit codes.

**Problem:** The tool works well but is invisible. No CI pipeline, no release automation, no package manager distribution, no GitHub Action. Users must `git clone && go build`. Core UX features like path filtering and progress indicators are missing.

**Goal:** Transform rift from "working CLI" to "product people discover, adopt, and rely on" through 4 milestones shipping incrementally.

**Principle:** Ship what lets people *find* you before shipping what *impresses* experts.

---

## Milestone 1: "Shippable Product" (v1.2)

Make it installable, testable, and releasable without manual intervention. Everything after this depends on having a proper release pipeline.

### Phase 12: Build Infrastructure
**Scope:** Medium | **Depends on:** Nothing

Build:
1. **Makefile** — targets: `build`, `test`, `lint`, `cover`, `install`, `clean`. Inject version/commit/date via `-ldflags` matching `cmd/version.go:11-15`
2. **`.golangci.yml`** — govet, errcheck, staticcheck, unused, gosimple, ineffassign
3. **`.github/workflows/ci.yml`** — on push/PR to main: `go test ./...`, `golangci-lint run`, coverage upload
4. **`.gitignore`** updates — `rift` binary, coverage profiles, `dist/`

Files to create: `Makefile`, `.golangci.yml`, `.github/workflows/ci.yml`
Files to modify: `.gitignore`

---

### Phase 13: Release Automation
**Scope:** Medium | **Depends on:** Phase 12

Build:
1. **`.goreleaser.yml`** — cross-compile linux/darwin (amd64+arm64), windows/amd64. Checksums. ldflags from git tag matching `cmd/version.go`
2. **`.github/workflows/release.yml`** — triggered on `v*` tag push, runs goreleaser, creates GitHub Release
3. **`Dockerfile`** — multi-stage (Go builder → distroless), for `docker run ghcr.io/Ommanimesh2/rift`
4. **Docker image push** in release workflow to ghcr.io

Files to create: `.goreleaser.yml`, `.github/workflows/release.yml`, `Dockerfile`

---

### Phase 14: Homebrew Distribution
**Scope:** Small | **Depends on:** Phase 13

Build:
1. **Homebrew tap** — separate repo `Ommanimesh2/homebrew-tap` with auto-generated formula
2. **goreleaser `brews` section** in `.goreleaser.yml`
3. **README update** — `brew install Ommanimesh2/tap/rift` as primary install method

Files to modify: `.goreleaser.yml`, `README.md`
External: create `Ommanimesh2/homebrew-tap` repo

---

### Phase 15: Shell Completions
**Scope:** Small | **Depends on:** Phase 12

Build:
1. **`cmd/completion.go`** — `rift completion bash|zsh|fish|powershell` using cobra built-ins
2. **README update** with completion install instructions

Files to create: `cmd/completion.go`
Files to modify: `README.md`

---

## Milestone 2: "CI/CD First-Class Citizen" (v1.3)

Make rift the default image comparison step in CI/CD pipelines. The GitHub Action is the single most impactful thing to build after release automation.

### Phase 16: GitHub Action
**Scope:** Large | **Depends on:** Phase 13

Build (separate repo `Ommanimesh2/rift-action`):
1. **`action.yml`** — inputs: `image1`, `image2`, `format`, `fail-on-security`, `size-threshold`, `platform`, `exit-code`, `rift-version`
2. **Outputs:** `diff-output`, `has-changes`, `has-security-events`, `exit-code`
3. **Step summary** — when `format=markdown`, pipe to `$GITHUB_STEP_SUMMARY`
4. **PR comment mode** — optionally post markdown diff as PR comment
5. **Example workflows** and marketplace listing

Files in main repo to modify: `README.md` (add GitHub Action usage section)

---

### Phase 17: Path Filtering (--include / --exclude)
**Scope:** Medium | **Depends on:** Nothing

Build:
1. **`--include` and `--exclude` flags** — glob patterns, repeatable (e.g., `--exclude "var/cache/**" --exclude "*.pyc"`)
2. **`internal/diff/filter.go`** — `FilterEntries()` pure function on `DiffResult`, recomputes summary counts
3. **Use `doublestar` library** for `**` glob support (stdlib `path.Match` lacks it)

Design: filter post-diff (on `DiffResult`) not pre-diff (on `FileTree`). Preserves correct unfiltered stats.

Files to create: `internal/diff/filter.go`, `internal/diff/filter_test.go`
Files to modify: `cmd/root.go` (flags + wire filter between diff and output, around line 99-132), `go.mod`

---

### Phase 18: SARIF Output Format
**Scope:** Medium | **Depends on:** Nothing

Build:
1. **`internal/output/sarif.go`** — SARIF v2.1.0 schema. Map security events to SARIF Results with ruleId (`rift/new_suid` etc), level (error/warning), artifact location
2. **`--format sarif`** option in `cmd/root.go`
3. Enables upload to GitHub Code Scanning via `codeql-action/upload-sarif`

Files to create: `internal/output/sarif.go`, `internal/output/sarif_test.go`
Files to modify: `cmd/root.go` (add `sarif` case to format switch at line 132)

---

### Phase 19: Verbose Mode & Progress Indicators
**Scope:** Medium | **Depends on:** Nothing

Build:
1. **`--verbose` / `-v` flag** — structured log to stderr via `log/slog`: "Opening image1...", "Skipping 3 identical layers", "Building tree (5 layers)...", "Found 42 changes"
2. **Spinner/progress** — charmbracelet spinner on stderr during layer download (only when stderr is TTY)

Files to create: `internal/log/log.go`
Files to modify: `cmd/root.go` (add flag, wire logging around source.Open/tree.Build/diff.Diff)

---

## Milestone 3: "Power Features" (v1.4)

Features that make power users stay and recommend the tool.

### Phase 20: Config File Support
**Scope:** Medium | **Depends on:** Phase 17

Build:
1. **Config loading** from `.rift.yml` (cwd) then `$HOME/.config/rift/config.yml`
2. **Config structure** mirrors CLI flags: format, security-only, platform, exit-code, fail-on-security, size-threshold, include, exclude, verbose
3. **Precedence:** CLI flags > config file > defaults
4. **`rift init`** subcommand generating commented template
5. Use `gopkg.in/yaml.v3` (lightweight, no viper)

Files to create: `internal/config/config.go`, `internal/config/config_test.go`, `cmd/init.go`
Files to modify: `cmd/root.go`, `go.mod`

---

### Phase 21: Content Diff for Text Files
**Scope:** Large | **Depends on:** Nothing

Build:
1. **`--content-diff` flag** — for modified text files, show unified diff of actual content
2. **`internal/content/` package** — takes two `v1.Image` objects + list of paths, returns content pairs. Text detection via null-byte heuristic (first 512 bytes), size cap at 1MB
3. **Lazy extraction (Approach A)** — re-read layer tars only for content-changed files after diff computed. Memory-efficient
4. **Unified diff** via `github.com/sergi/go-diff/diffmatchpatch`
5. **Output integration** — inline diffs in terminal (red/green), `"diff"` field in JSON, fenced blocks in markdown

Files to create: `internal/content/content.go`, `internal/content/content_test.go`, `internal/content/udiff.go`, `internal/content/udiff_test.go`
Files to modify: `internal/tree/tree.go` (content re-read capability), `cmd/root.go`, `internal/output/terminal.go`, `json.go`, `markdown.go`

---

### Phase 22: Integration Tests
**Scope:** Medium | **Depends on:** Phase 12

Build:
1. **`test/integration_test.go`** with `//go:build integration` tag
2. **Pinned images:** `alpine:3.18@sha256:...` vs `alpine:3.19@sha256:...`
3. **Scenarios:** remote comparison, JSON validation, security events, quick mode, exit codes, platform selection
4. **CI:** separate job on push to main only (avoid registry rate limits on PRs)

Files to create: `test/integration_test.go`
Files to modify: `.github/workflows/ci.yml`

---

## Milestone 4: "Ecosystem Integration" (v2.0)

Position rift as part of the container DevOps ecosystem.

### Phase 23: Docker CLI Plugin
**Scope:** Medium | **Depends on:** Phase 13

Build:
1. **`cmd/docker_plugin.go`** — hidden `docker-cli-plugin-metadata` cobra command returning JSON
2. **goreleaser** `docker-rift` binary target
3. **Install script** for copying to `~/.docker/cli-plugins/`
4. Enables `docker rift nginx:1.24 nginx:1.25`

Files to create: `cmd/docker_plugin.go`, `scripts/install-docker-plugin.sh`
Files to modify: `.goreleaser.yml`, `README.md`

---

### Phase 24: Interactive TUI
**Scope:** Large | **Depends on:** Phase 17

Build:
1. **`rift tui <image1> <image2>`** subcommand using bubbletea
2. **Tree view** — hierarchical, expandable/collapsible, colored by change type
3. **Detail panel** — metadata diff for selected file
4. **Security panel** — sidebar with security events
5. **Keyboard:** arrows, j/k, `/` search, `q` quit, tab to switch panels

Files to create: `internal/tui/model.go`, `tree.go`, `detail.go`, `security.go`, `keys.go`, `styles.go`, `cmd/tui.go`
Files to modify: `go.mod`, `README.md`

---

### Phase 25: Package-Level Awareness
**Scope:** Large | **Depends on:** Phase 21

Build:
1. **Package detection** — parse APK (`/lib/apk/db/installed`), DEB (`/var/lib/dpkg/status`), RPM, pip dist-info
2. **Package diff** — added/removed/upgraded packages between images
3. **`--packages` flag** — show package changes alongside file changes
4. Transform "47 file changes" into "curl upgraded 1.2→1.3"

Files to create: `internal/packages/detect.go`, `apk.go`, `deb.go`, `packages_test.go`
Files to modify: `cmd/root.go`, all output formatters

---

## Execution Order (Priority)

| # | Phase | Scope | Milestone | Rationale |
|---|-------|-------|-----------|-----------|
| 1 | 12: Build Infra | M | v1.2 | Everything depends on CI |
| 2 | 13: Release Automation | M | v1.2 | Users can't install without binaries |
| 3 | 14: Homebrew | S | v1.2 | Primary discovery path |
| 4 | 15: Shell Completions | S | v1.2 | Trivial, signals polish |
| 5 | 16: GitHub Action | L | v1.3 | #1 adoption driver |
| 6 | 17: Path Filtering | M | v1.3 | Most impactful UX feature |
| 7 | 19: Verbose/Progress | M | v1.3 | Needed for large images |
| 8 | 22: Integration Tests | M | v1.3 | Safety net before more features |
| 9 | 18: SARIF Output | M | v1.3 | Unique differentiator |
| 10 | 20: Config File | M | v1.4 | Teams need consistent settings |
| 11 | 23: Docker Plugin | M | v2.0 | Natural invocation |
| 12 | 21: Content Diff | L | v1.4 | High value, architecturally complex |
| 13 | 24: Interactive TUI | L | v2.0 | "Wow" factor, needs distribution first |
| 14 | 25: Package Awareness | L | v2.0 | Nice-to-have, not urgent |

## Explicitly Out of Scope

- No web UI (CLI only)
- No vulnerability scanning (use trivy/grype)
- No image optimization (use SlimToolkit)
- No single-image exploration (use dive)
- No Cosign/signature verification
- No SBOM generation (use syft)

## Architectural Principles

1. **New `internal/` packages** per feature — preserves modularity
2. **`cmd/root.go` is the orchestrator** — wiring only, no business logic
3. **Pure functions** — follow `security.Analyze()` pattern (no I/O, no state)
4. **Test-driven** — `_test.go` with table-driven tests before implementation
5. **Formatters are independent** — each output format in its own file

## Key Files

- `cmd/root.go:17-28` — flag definitions, all new flags go here
- `cmd/root.go:51-167` — RunE orchestration, new feature wiring goes here
- `cmd/version.go:11-15` — ldflags pattern goreleaser must match
- `internal/diff/diff.go` — DiffEntry/DiffResult structs that filters and content diff extend
- `internal/tree/tree.go` — FileNode/BuildTree that content diff needs to extend
- `internal/output/json.go` — DiffReport struct pattern for new formatters (SARIF)

## Verification

After each phase:
1. `make test` — all existing + new tests pass
2. `make lint` — no new lint issues
3. `make build` — binary builds successfully
4. Manual smoke test: `./rift alpine:3.18 alpine:3.19` produces correct output
5. For CI phases: verify GitHub Actions run green on a test PR
6. For distribution phases: verify `brew install` / `docker run` works end-to-end
