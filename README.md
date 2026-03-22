# rift

![rift demo](demo.gif)

[![Go](https://img.shields.io/badge/go-%3E%3D1.21-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/Ommanimesh2/rift)](https://goreportcard.com/report/github.com/Ommanimesh2/rift)

**The `git diff` for container images.**

File-level diff for container images — files added, removed, modified — with size impact,
security highlights, and CI/CD-ready JSON output. Replaces
[Google's archived container-diff](https://github.com/GoogleContainerTools/container-diff).

---

## Install

**Homebrew (macOS/Linux):**

```sh
brew install Ommanimesh2/tap/rift
```

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

Flags:
  --format string         Output format: terminal, json, markdown, sarif (default "terminal")
  --security-only         Show only security-relevant changes
  --quick                 Manifest-only comparison (no content download)
  --platform string       Target platform for multi-arch images (e.g., linux/amd64)
  --username string       Registry username for explicit authentication
  --password string       Registry password for explicit authentication
  --include strings       Glob patterns to include (repeatable, e.g., --include "etc/**")
  --exclude strings       Glob patterns to exclude (repeatable, e.g., --exclude "var/cache/**")
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
nginx:1.24 → nginx:1.25

  + usr/share/nginx/html/50x.html          +1.2 KB
  ~ etc/nginx/nginx.conf                   +340 B
  - etc/nginx/conf.d/default.conf.bak      -2.1 KB

Summary: 1 added, 1 removed, 1 modified (+1.5 KB / -2.1 KB)

Layers:
  Layer 1  sha256:a3…  200 MB  shared
  Layer 2  sha256:b7…   12 MB  → sha256:c9…  14 MB  +2 MB

Security: no findings
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
    "removed": 1,
    "modified": 12,
    "added_bytes": 1245184,
    "removed_bytes": 4096
  },
  "changes": [
    {
      "path": "etc/alpine-release",
      "type": "modified",
      "size_delta": 2,
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
6. **Output** — renders to terminal (lipgloss), JSON, or Markdown

---

## License

MIT
