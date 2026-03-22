# rift

[![Go](https://img.shields.io/badge/go-%3E%3D1.21-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/Ommanimesh2/rift)](https://goreportcard.com/report/github.com/Ommanimesh2/rift)

**The `git diff` for container images.**

File-level diff for container images — files added, removed, modified — with size impact,
security analysis, package-level awareness, and CI/CD-ready output. Replaces
[Google's archived container-diff](https://github.com/GoogleContainerTools/container-diff).

---

## Install

**Go install (requires Go 1.21+):**

```sh
go install github.com/Ommanimesh2/rift@latest
```

**Docker:**

```sh
docker run --rm ghcr.io/Ommanimesh2/rift nginx:1.24 nginx:1.25
```

**Download binary:**

Grab the latest release from [GitHub Releases](https://github.com/Ommanimesh2/rift/releases).

---

## Quick Start

```sh
# Compare two registry images
rift nginx:1.24 nginx:1.25

# Compare local Docker daemon images
rift myapp:latest myapp:v2.0

# Compare OCI tarball archives
rift ./old.tar ./new.tar

# Output JSON for pipelines
rift --format json alpine:3.18 alpine:3.19

# Show only security-relevant changes
rift --security-only ubuntu:22.04 ubuntu:24.04

# Show package-level changes (APK/DEB)
rift --packages alpine:3.18 alpine:3.19

# Interactive TUI browser
rift tui nginx:1.24 nginx:1.25

# Fast manifest-only check (no content download)
rift --quick nginx:1.24 nginx:1.25
```

---

## Image Sources

rift auto-detects the source type from the reference string:

| Reference | Source | Example |
|-----------|--------|---------|
| `name:tag` | Remote registry | `nginx:1.25`, `ghcr.io/org/app:sha` |
| `daemon://name` | Local Docker daemon | `daemon://myapp:latest` |
| `./file.tar` | OCI tarball | `./image.tar`, `./export.tar.gz` |

Registry images use Docker credential helpers automatically (`~/.docker/config.json`).

---

## Flags

```
Usage:
  rift <image1> <image2> [flags]

Commands:
  tui            Interactive TUI for browsing image diffs
  init           Create a .rift.yml configuration file
  completion     Generate shell completion scripts (bash, zsh, fish, powershell)
  version        Print version information

Flags:
  --format string         Output format: terminal, json, markdown, sarif (default "terminal")
  --security-only         Show only security-relevant changes
  --quick                 Manifest-only comparison (no content download)
  --platform string       Target platform for multi-arch images (e.g., linux/amd64)
  --username string       Registry username for explicit authentication
  --password string       Registry password for explicit authentication
  --include strings       Glob patterns to include (repeatable, e.g., --include "etc/**")
  --exclude strings       Glob patterns to exclude (repeatable, e.g., --exclude "var/cache/**")
  --content-diff          Show unified diff for modified text files
  --packages              Show package-level changes (APK, DEB)
  -v, --verbose           Enable verbose logging to stderr

CI/CD flags:
  --exit-code             Exit 2 if any file changes are found
  --fail-on-security      Exit 2 if security events are detected
  --size-threshold string Exit 2 if net size increase exceeds threshold (e.g., 10MB, 500KB)

  -h, --help              Show help
      --version           Show version
```

---

## Output Formats

### Terminal (default)

Color-coded diff with size impact summary, per-layer breakdown, and security highlights:

```
Comparing alpine:3.18 → alpine:3.19

~ bin/busybox  [content]  (-8.0 KB)
- bin/ed  (0 bytes)
~ etc/alpine-release  [content]  (-1 bytes)
+ etc/udhcpc/udhcpc.conf  (+287 bytes)
~ lib/ld-musl-x86_64.so.1  [content]  (+32.0 KB)

Security Findings (2)
━━━━━━━━━━━━━━━━━━━━━━━━
  [WORLD-WRITABLE] lib/libz.so.1  (0777 → 0777)
  [NEW EXEC] lib/libz.so.1.3.1

Summary: 3 added (+98.1 KB), 3 removed (-97.9 KB), 29 modified
```

### JSON

Machine-readable output for CI/CD pipelines:

```sh
rift --format json alpine:3.18 alpine:3.19
```

```json
{
  "image1": "alpine:3.18",
  "image2": "alpine:3.19",
  "summary": {
    "added": 3,
    "removed": 3,
    "modified": 29,
    "added_bytes": 100495,
    "removed_bytes": 100280
  },
  "changes": [
    {
      "path": "etc/alpine-release",
      "type": "modified",
      "size_delta": -1,
      "changes": ["content"]
    }
  ],
  "security_events": []
}
```

### Markdown

GitHub-Flavored Markdown for PR comments and documentation:

```sh
rift --format markdown myapp:v1 myapp:v2
```

Output is a formatted table ready to paste into a PR description or GitHub Actions step summary.

### SARIF

Upload security findings directly to GitHub Code Scanning:

```sh
rift --format sarif myapp:v1 myapp:v2 > results.sarif
```

Use with GitHub Actions:

```yaml
- name: Run rift security scan
  run: rift --format sarif ${{ env.OLD_IMAGE }} ${{ env.NEW_IMAGE }} > results.sarif

- uses: github/codeql-action/upload-sarif@v3
  with:
    sarif_file: results.sarif
```

---

## Path Filtering

Focus on what matters by including or excluding paths:

```sh
# Only show changes in /etc
rift --include "etc/**" alpine:3.18 alpine:3.19

# Ignore cache and docs
rift --exclude "var/cache/**" --exclude "usr/share/doc/**" myapp:v1 myapp:v2

# Combine include and exclude
rift --include "usr/**" --exclude "**/*.pyc" myapp:v1 myapp:v2
```

Glob patterns support `**` for recursive matching.

---

## Package-Level Changes

See what packages changed instead of raw file diffs:

```sh
rift --packages alpine:3.18 alpine:3.19
```

```
--- Package Changes ---
  ~ busybox 1.36.1-r7 → 1.36.1-r20
  ~ libcrypto3 3.1.8-r0 → 3.1.8-r1
  ~ musl 1.2.4-r3 → 1.2.4_git20230717-r5
  ~ zlib 1.2.13-r1 → 1.3.1-r0
```

Supports Alpine (APK) and Debian/Ubuntu (DEB) package databases.

---

## Content Diff

See exactly what changed inside modified text files:

```sh
rift --content-diff alpine:3.18 alpine:3.19
```

Generates unified diffs for text files like config files, scripts, and package databases. Binary files and files over 1 MB are skipped.

---

## Interactive TUI

Browse diffs interactively:

```sh
rift tui alpine:3.18 alpine:3.19
```

| Key | Action |
|-----|--------|
| `j` / `k` / arrows | Navigate |
| `/` | Search by path |
| `tab` | Switch between file list and detail panel |
| `esc` | Clear search |
| `q` | Quit |

---

## Security Analysis

rift automatically detects security-relevant changes and highlights them in the output.

| Event | Trigger |
|-------|---------|
| `new_suid` | Added file with SUID bit set |
| `new_sgid` | Added file with SGID bit set |
| `suid_added` | Modified file gained SUID bit |
| `sgid_added` | Modified file gained SGID bit |
| `new_executable` | Added non-directory file with any execute bit |
| `world_writable` | Added or modified file is world-writable |
| `perm_escalation` | Modified file has strictly more permissive bits |

```sh
# Show only security-relevant paths
rift --security-only ubuntu:22.04 ubuntu:24.04

# Fail CI on any security finding
rift --fail-on-security ubuntu:22.04 ubuntu:24.04
```

---

## CI/CD Integration

### Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Success — no conditions triggered |
| `1` | Tool error (bad arguments, unreachable image, etc.) |
| `2` | Condition triggered (changes, security events, or size threshold exceeded) |

### GitHub Action

Use rift directly in your workflows:

```yaml
- name: Image diff
  uses: Ommanimesh2/rift@v1
  with:
    image1: myapp:${{ github.event.pull_request.base.sha }}
    image2: myapp:${{ github.sha }}
    format: markdown
    fail-on-security: true
    size-threshold: 10MB
```

Or run manually:

```yaml
- name: Image diff check
  run: |
    rift \
      --fail-on-security \
      --size-threshold 10MB \
      --format markdown \
      myapp:${{ github.base_ref }} \
      myapp:${{ github.sha }} | tee $GITHUB_STEP_SUMMARY
```

### Size threshold gate

```sh
# Fail if image grew by more than 5 MB
rift --size-threshold 5MB myapp:v1 myapp:v2

# Fail if image grew at all
rift --exit-code myapp:v1 myapp:v2
```

Threshold units: `B`, `KB`, `MB`, `GB` (case-insensitive). Decimals supported: `1.5MB`.

### Private registries

rift uses Docker credential helpers by default — if `docker pull` works, rift works.

For explicit credentials (useful in CI without Docker config):

```sh
rift --username $REGISTRY_USER --password $REGISTRY_PASS \
  ghcr.io/org/app:v1 ghcr.io/org/app:v2
```

### Multi-arch images

```sh
# Default: use host platform
rift myapp:v1 myapp:v2

# Select specific platform
rift --platform linux/arm64 myapp:v1 myapp:v2
rift --platform linux/amd64 myapp:v1 myapp:v2
```

---

## Configuration

Create a `.rift.yml` in your project to set default flags:

```sh
rift init
```

Example `.rift.yml`:

```yaml
format: terminal
exclude:
  - "var/cache/**"
  - "**/*.pyc"
fail-on-security: true
verbose: false
```

CLI flags always override config file values. Config is loaded from `.rift.yml` in the current directory, then `~/.config/rift/config.yml`.

---

## Docker CLI Plugin

Use rift as a Docker subcommand:

```sh
# Install
./scripts/install-docker-plugin.sh ./rift

# Use
docker rift nginx:1.24 nginx:1.25
```

---

## Performance

rift is fast by default and has explicit speed modes:

| Feature | Description |
|---------|-------------|
| **Shared layer skip** | Layers with identical digests are not downloaded or parsed |
| **Streaming** | Tar entries are streamed in memory — no full layer download to disk |
| **`--quick` mode** | Manifest-only comparison using layer digests, no content download at all |

```sh
# Instant comparison via manifest only (registry images)
rift --quick nginx:1.24 nginx:1.25
```

`--quick` shows layer-level changes (which layers were added, removed, or replaced) without
downloading any content. Useful for a fast "did anything change?" check in CI.

---

## Shell Completions

```sh
# Bash
source <(rift completion bash)

# Zsh
rift completion zsh > "${fpath[1]}/_rift"

# Fish
rift completion fish | source
```

---

## How It Works

1. **Source detection** — auto-detects registry, daemon, or tarball from reference string
2. **Layer skip** — compares layer digests; identical leading layers are skipped entirely
3. **Tree construction** — streams remaining layers, building an in-memory file tree per image (handles OCI whiteout files for layer deletions)
4. **Diff** — compares the two squashed file trees, computing per-file changes with size deltas and attribute flags (content, mode, uid/gid, symlink target)
5. **Security analysis** — pure-function pass over the diff result, flagging security-relevant permission changes
6. **Package detection** — parses APK/DEB package databases from both images to show package-level changes
7. **Output** — renders to terminal (lipgloss), JSON, Markdown, or SARIF

---

## License

[MIT](LICENSE)
