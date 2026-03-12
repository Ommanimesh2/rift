---
phase: 10-vhs-demo-gif
type: research
domain: terminal-recording
confidence: HIGH
---

# Phase 10 Research: charmbracelet/vhs

## Summary

vhs is a terminal recording tool from Charmbracelet (same ecosystem as lipgloss). It executes script-based `.tape` files deterministically — same output every time. Required for producing the imgdiff demo GIF.

---

## Standard Stack

### Installation (macOS)

```bash
brew install vhs ffmpeg ttyd
```

- **vhs**: The tape recorder itself
- **ffmpeg**: Required — converts frames to GIF/MP4/WebM
- **ttyd**: Required — virtual terminal vhs controls

Verify:
```bash
vhs --version
which ttyd ffmpeg
```

---

## Tape File Syntax

### File Structure

All `Set` and `Output` commands **must appear first**, before any typing/interaction. Commands after interaction are silently ignored.

```tape
# demo.tape
Output demo.gif

Set FontSize 16
Set Width 1200
Set Height 700
Set Theme "Catppuccin Mocha"
Set Padding 20
Set TypingSpeed 75ms
Set Shell "zsh"

# Now interaction begins
Type "imgdiff nginx:1.24 nginx:1.25"
Enter
Wait /Done/
Sleep 2
```

### Key Commands

| Command | Description | Example |
|---------|-------------|---------|
| `Output <file>` | Output path (.gif, .mp4, .webm) | `Output demo.gif` |
| `Set <Key> <Value>` | Configure terminal | `Set Width 1200` |
| `Type "<text>"` | Type text | `Type "imgdiff nginx:1.24 nginx:1.25"` |
| `Enter` | Press Enter | `Enter` |
| `Ctrl+C` | Send Ctrl+C | `Ctrl+C` |
| `Sleep <duration>` | Pause | `Sleep 2`, `Sleep 500ms` |
| `Wait /<regex>/` | Wait for pattern (last line) | `Wait /\$` |
| `Wait+Screen /<regex>/` | Wait for pattern (full screen) | `Wait+Screen /Done/` |
| `Hide` | Hide subsequent commands from output | — |
| `Show` | Resume visible output | — |
| `Screenshot <path>` | Capture PNG frame | `Screenshot frame.png` |

### Configuration Settings

```tape
Set FontSize 16          # Font size in pixels
Set Width 1200           # Terminal width in pixels
Set Height 700           # Terminal height in pixels
Set Padding 20           # Padding around terminal
Set Theme "Catppuccin Mocha"  # Color theme
Set TypingSpeed 75ms     # Per-character typing speed
Set Shell "zsh"          # Shell to use
Set WindowBar Colorful   # macOS-style window decorations
```

---

## GIF Size Optimization (target: <3MB)

**Biggest levers (in order of impact):**

1. **Resolution** — `Set Width 1000` instead of 1200 saves ~25% file size
2. **Typing speed** — Slower typing = fewer frames = smaller file
3. **Sleep duration** — Only sleep as long as needed for readability
4. **Post-process with gifsicle** (if needed):
   ```bash
   brew install gifsicle
   gifsicle -O3 --colors 128 demo.gif -o demo-optimized.gif
   ```

**Recommended starting point for imgdiff:**
```tape
Set Width 1000
Set Height 600
Set FontSize 14
Set TypingSpeed 75ms
```

A 30-60s GIF at these dimensions should stay well under 3MB. Adjust if oversized.

---

## Common Pitfalls

1. **Set commands AFTER interaction are ignored** — always put all `Set` at top
2. **Wait timeout defaults to 15s** — use `Wait@30s /pattern/` for slow commands
3. **LF line endings only** — Windows CRLF causes parse errors
4. **Quote all Type strings** — `Type "text"` not `Type text`
5. **Shell prompt pattern** — the default `Wait` pattern expects `>$`; for zsh with custom prompt, use `Wait /\$ /` or `Wait+Screen /<expected-output>/`
6. **Long-running commands** — imgdiff pulls images from registry; first run may take 30-60s. Use `Wait+Screen` with long timeout.
7. **Image pull caching** — pre-pull images before recording with `Hide`/`Show` to avoid network wait in GIF

---

## Theme Recommendation

Use **Catppuccin Mocha** — dark, high-contrast, from same Charmbracelet ecosystem. Consistent aesthetic with lipgloss rendering.

```tape
Set Theme "Catppuccin Mocha"
```

Other options if needed: `Dracula`, `Nord`, `Tokyo Night`

---

## Demo Script Strategy for imgdiff

**Pre-pull approach** (hide network latency from GIF):

```tape
Output demo.gif

Set Width 1000
Set Height 600
Set FontSize 14
Set Theme "Catppuccin Mocha"
Set Padding 20
Set TypingSpeed 75ms
Set Shell "zsh"

# Pre-pull images silently (hide from GIF)
Hide
Type "imgdiff nginx:1.24 nginx:1.25 --format json > /dev/null 2>&1"
Enter
Wait+Screen /\$ /
Type "imgdiff ubuntu:22.04 ubuntu:24.04 --format json > /dev/null 2>&1"
Enter
Wait+Screen /\$ /
Show

# Scene 1: File diff
Type "imgdiff nginx:1.24 nginx:1.25"
Enter
Wait+Screen /\$ /
Sleep 3

# Scene 2: Security
Type "imgdiff ubuntu:22.04 ubuntu:24.04"
Enter
Wait+Screen /\$ /
Sleep 3

# Scene 3: JSON
Type "imgdiff nginx:1.24 nginx:1.25 --format json | head -30"
Enter
Wait+Screen /\$ /
Sleep 2
```

---

## Don't Hand-Roll

- **GIF production**: Use vhs — don't use screen recording tools or manual ffmpeg
- **Color optimization**: Let vhs + ffmpeg handle palette; only use gifsicle if GIF is too large
- **Timing**: Use `Wait+Screen` not `Sleep` for commands with variable duration (network, parsing)

---

## Confidence

**HIGH** — Official docs + multiple verified sources. vhs is well-documented with stable API.

Install: `brew install vhs ffmpeg ttyd`
