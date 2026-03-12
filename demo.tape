Output demo.gif

Set FontSize 14
Set Width 1000
Set Height 620
Set Theme "Catppuccin Mocha"
Set Padding 20
Set TypingSpeed 80ms
Set Shell "zsh"
Set WindowBar Colorful

# ── Pre-pull images silently ──────────────────────────────────────────────────
Hide
Type "./imgdiff nginx:1.24 nginx:1.25 --format json > /dev/null 2>&1; echo DONE"
Enter
Wait+Screen /DONE/
Wait 1
Type "./imgdiff ubuntu:22.04 ubuntu:24.04 --format json > /dev/null 2>&1; echo DONE"
Enter
Wait+Screen /DONE/
Wait 1
Show

# ── Clear terminal for clean start ────────────────────────────────────────────
Type "clear"
Enter
Wait 0.5

# ── Scene 1: File diff — nginx upgrade ────────────────────────────────────────
Type "# Scene 1: What changed in nginx 1.24 → 1.25?"
Enter
Sleep 1
Type "./imgdiff nginx:1.24 nginx:1.25"
Enter
Wait+Screen /Summary:/
Sleep 4

# ── Scene 2: Security — ubuntu major upgrade ──────────────────────────────────
Type "clear"
Enter
Wait 0.3
Type "# Scene 2: Security changes in ubuntu 22.04 → 24.04"
Enter
Sleep 1
Type "./imgdiff ubuntu:22.04 ubuntu:24.04"
Enter
Wait+Screen /Summary:/
Sleep 4

# ── Scene 3: JSON for CI/CD ───────────────────────────────────────────────────
Type "clear"
Enter
Wait 0.3
Type "# Scene 3: JSON output for CI/CD pipelines"
Enter
Sleep 1
Type "./imgdiff nginx:1.24 nginx:1.25 --format json | head -30"
Enter
Wait+Screen /}/
Sleep 3
