Output demo.gif

Set FontSize 16
Set Width 1100
Set Height 680
Set Theme "Dracula"
Set Padding 30
Set BorderRadius 12
Set TypingSpeed 80ms
Set Shell "zsh"
Set WindowBar Colorful

# ── Scene 1: nginx file diff ──────────────────────────────────────────────────
Type "# What changed in nginx 1.24 -> 1.25?"
Enter
Sleep 1s
Type "./imgdiff daemon://nginx:1.24 daemon://nginx:1.25 | grep -E 'nginx|conf|\.so|Summary' | head -20"
Enter
Sleep 8s

# ── Scene 2: Security ─────────────────────────────────────────────────────────
Type "clear"
Enter
Sleep 300ms
Type "# Security changes: ubuntu 22.04 -> 24.04"
Enter
Sleep 1s
Type "./imgdiff daemon://ubuntu:22.04 daemon://ubuntu:24.04 --security-only"
Enter
Sleep 8s

# ── Scene 3: JSON for CI/CD ───────────────────────────────────────────────────
Type "clear"
Enter
Sleep 300ms
Type "# Pipeline-ready JSON output"
Enter
Sleep 1s
Type "./imgdiff daemon://nginx:1.24 daemon://nginx:1.25 --format json | head -30"
Enter
Sleep 10s
