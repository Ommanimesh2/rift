---
phase: 10-vhs-demo-gif
plan: 02
subsystem: demo
tags: [vhs, gif, demo, terminal-recording]
requires: [10-01]
provides: [demo.gif, demo.tape, make-demo.sh]
affects: [11-readme-polish]
tech-stack:
  added: [vhs, ffmpeg, ttyd]
  patterns: [daemon:// source prefix for local Docker reads]
key-files:
  created: [demo.gif, demo.tape, make-demo.sh]
  modified: []
key-decisions:
  - "Use daemon:// prefix instead of remote refs to avoid network latency in GIF recording"
  - "Grep filter (nginx|conf|\\.so|Summary) for Scene 1 to surface interesting nginx changes"
  - "Remove Hide/Show — vhs 0.11.0 does not suppress Hide commands from GIF output"
duration: ~15 min
completed: 2026-03-12
---

# Phase 10 Plan 02: GIF Production Summary

vhs render + human approval + commit of demo GIF assets.

## Accomplishments

- Rendered `demo.gif` (571KB, 1000×620→1100×680, Dracula theme)
- Fixed Hide/Show issue: removed pre-pull block, switched to `daemon://` source prefix
- Three scenes recorded cleanly with daemon reads (~6s nginx, ~3s ubuntu)
- Human-approved after visual iteration (2 re-renders)
- Committed: `demo.tape`, `demo.gif`, `make-demo.sh`

## Files Created/Modified

- `demo.gif` — 571KB terminal recording, committed to git
- `demo.tape` — final vhs tape script (updated from Plan 10-01 draft)
- `make-demo.sh` — helper script: docker pull + vhs render

## Decisions Made

- **daemon:// prefix over remote refs**: Remote pulls take 30-120s per image, making GIF recording impractical. `daemon://` reads from local Docker cache in ~3-6s.
- **Removed Hide/Show**: vhs 0.11.0 does not suppress hidden commands from GIF output. Pre-pull moved to `make-demo.sh` wrapper script.
- **Grep filter for Scene 1**: `head -30` on nginx diff shows generic `/bin/bash`, `/bin/cat` removals — not compelling. Grep filter surfaces nginx binary, config, shared lib changes.

## Deviations from Plan

- [Rule 1 - Bug] vhs `Hide`/`Show` does not work as expected in 0.11.0 — hidden commands appear in GIF output. Fixed by removing the block and moving pre-pull to `make-demo.sh`.
- [Rule 1 - Bug] `Wait N` (bare number) syntax rejected by vhs 0.11.0 — requires `Sleep Ns`. Fixed in Plan 10-01.
- Scene 1 command changed from `| head -50` → `| grep -E 'nginx|conf|\.so|Summary' | head -20` for more compelling output.

## Issues Encountered

None — plan executed with deviations handled inline.

## Next Phase Readiness

Phase 11 (README Polish) is ready:
- `demo.gif` is committed and available at repo root
- Embed path: `![imgdiff demo](demo.gif)`
- File size is 571KB — well under GitHub's rendering threshold

## Performance

- Duration: ~15 min (including 2 re-renders for visual iteration)
- GIF size: 571KB (target: <3MB ✓)
- Re-renders: 2 (initial Hide/Show fix + visual polish iteration)
