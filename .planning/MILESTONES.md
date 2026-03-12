# Project Milestones: imgdiff

---

## v1.0 MVP (Shipped: 2026-03-12)

**Delivered:** Full-featured, single-binary container image diff tool — file-level diff with security analysis, three output formats, performance optimizations, and CI/CD exit code integration.

**Phases completed:** 1–9 (21 plans total)

**Key accomplishments:**

- Complete CLI with cobra, multi-arch `--platform` flag, and three image sources (remote registry, Docker daemon, OCI tarball)
- Streaming in-memory file tree construction from OCI layers with full whiteout support
- File-level diff engine with content hash, permission, uid/gid, symlink, and type change tracking
- Color-coded terminal output via lipgloss with size impact summary and per-layer breakdown
- Security analysis: SUID/SGID detection, permission escalation, world-writable files, new executables
- JSON and Markdown formatters for CI/CD pipelines and PR comments
- Performance: shared layer skip (identical leading layer detection), streaming tar, `--quick` manifest-only mode
- CI/CD exit codes: `--exit-code`, `--fail-on-security`, `--size-threshold` (B/KB/MB/GB)
- Explicit registry auth via `--username`/`--password` (fallback: Docker DefaultKeychain)

**Stats:**

- 6,333 lines of Go
- 9 phases, 21 plans, ~50 tasks
- All packages tested with table-driven tests; core logic TDD (tree, diff, security, output formats, exitcode)
- 2 days from first commit to v1.0 (2026-03-11 → 2026-03-12)

**Git range:** `feat(01-01)` → `feat(09-02)`

**What's next:** v1.1 — goreleaser distribution (Homebrew, GitHub Releases, Docker), GitHub Action, interactive TUI mode exploration

---
