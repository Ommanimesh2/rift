---
phase: 10-vhs-demo-gif
plan: 01
status: complete
date: 2026-03-12
---

## Summary

Installed vhs tooling, built the imgdiff binary, wrote demo.tape, and validated the
tape with a successful test render.

## Tool Versions Installed

| Tool   | Version        | Path                      |
|--------|----------------|---------------------------|
| vhs    | 0.11.0         | /opt/homebrew/bin/vhs     |
| ffmpeg | 8.0.1          | /opt/homebrew/bin/ffmpeg  |
| ttyd   | 1.7.7-unknown  | /opt/homebrew/bin/ttyd    |

All installed via `brew install vhs ffmpeg ttyd`.

## imgdiff Binary

Built fresh with `go build -o imgdiff .`. Binary verified working at
`/Users/ommmishra/opensource/container-image-diff/imgdiff`.

## demo.tape Structure

Written to project root at `demo.tape`.

Structure:
- `Output demo.gif` + 8 `Set` commands at the top (Width 1000, Height 620, FontSize 14,
  Theme "Catppuccin Mocha", Padding 20, TypingSpeed 80ms, Shell "zsh", WindowBar Colorful)
- **Hidden pre-pull block**: both nginx and ubuntu images pulled silently before
  the visible GIF starts, eliminating registry wait time from the demo
- **Scene 1** (nginx:1.24 → nginx:1.25): file-level diff with layer breakdown
- **Scene 2** (ubuntu:22.04 → ubuntu:24.04): security events and major upgrade changes
- **Scene 3** (nginx:1.24 → nginx:1.25 --format json | head -30): CI/CD JSON output

Total scenes: 3. Estimated runtime: ~30–45 seconds.

## Wait Pattern Adjustments

The plan template specified `Wait+Screen /Total files/` as the completion signal.
After inspecting `internal/output/terminal.go`, the actual summary line produced by
imgdiff is:

```
Summary: N added (+X), N removed (-X), N modified
```

`Wait+Screen` patterns were updated to `/Summary:/` for Scenes 1 and 2.
Scene 3 retains `Wait+Screen /}/` which matches the closing brace of the JSON output.

## Test Render

Minimal tape rendered successfully to `/tmp/test-vhs.gif` (10 KB).
- vhs rendered without errors when PATH included `/opt/homebrew/bin`
- Theme "Catppuccin Mocha" confirmed working
- Note: vhs Output path must be relative (not absolute) — demo.tape already uses
  `Output demo.gif` which is correct

## Files Modified

- `demo.tape` — created in project root
- `.planning/phases/10-vhs-demo-gif/10-01-SUMMARY.md` — this file
