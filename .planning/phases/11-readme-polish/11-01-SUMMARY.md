---
phase: 11-readme-polish
plan: 01
status: complete
completed: 2026-03-12
---

# Summary: 11-01 — README Polish

## Tasks Completed: 2/2

### Task 1: Add GIF embed and badges below title
- Embedded `demo.gif` immediately after `# imgdiff` title
- Added three badges on one line: Go version, MIT license, Go Report Card
- Updated tagline from "File-level, security-aware container image diff tool." to "The `git diff` for container images."
- Tightened opening paragraph — cut verbose preamble, folded attribution naturally into description

### Task 2: Commit polished README
- Committed README.md as `feat(11-01): polish README — embed demo.gif, add badges, sharpen copy`
- Commit hash: 21091c8

## Verification

- `head -5 README.md` shows title then GIF embed line ✓
- 3 badges present (Go version, MIT license, Go Report Card) ✓
- Tagline updated to "The `git diff` for container images." ✓
- `git show --stat HEAD` shows README.md changed (15 +++++++++------) ✓

## Outcome

Phase 11 complete. v1.1 Demo milestone shipped 2026-03-12.

README now leads with the demo GIF for maximum first-scroll impact, backed by three badges and a punchy tagline that anchors the mental model ("git diff for container images").
