# rift

[![Go](https://img.shields.io/badge/go-%3E%3D1.21-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/Ommanimesh2/rift)](https://goreportcard.com/report/github.com/Ommanimesh2/rift)

**The `git diff` for container images.**

File-level diff for container images — files added, removed, modified — with size impact,
security analysis, secrets detection, layer attribution, package awareness, policy enforcement,
and CI/CD-ready output. Replaces
[Google's archived container-diff](https://github.com/GoogleContainerTools/container-diff).

---

## Install

**Go install (requires Go 1.21+):**

```sh
go install github.com/Ommanimesh2/rift@latest
```

**Download binary:**

Grab the latest release from [GitHub Releases](https://github.com/Ommanimesh2/rift/releases).

**Self-update:**

```sh
rift update
```

---

## Quick Start

```sh
# One-screen verdict — the fastest way to see what changed
rift --summary node:18-alpine node:20-alpine

# Full file-level diff
rift alpine:3.18 alpine:3.19

# Group changes by Dockerfile instruction
rift --layers node:18-alpine node:20-alpine

# Security audit with secrets scanning
rift --secrets --fail-on-security myapp:v1 myapp:v2

# Enforce policy rules from .rift.yml
rift --policy myapp:v1 myapp:v2

# JSON output for CI/CD pipelines
rift --format json alpine:3.18 alpine:3.19

# Interactive TUI browser
rift tui nginx:1.24 nginx:1.25
```

---

## Summary Mode

Get the full picture in 6 lines:

```sh
rift --summary node:18-alpine node:20-alpine
```

```
  Image:    node:18-alpine → node:20-alpine
  Size:     42.8 MB → 46.1 MB (+3.3 MB)
  Files:    48 added, 11 removed, 312 modified
  Packages: alpine-baselayout 3.6.8-r1→3.7.1-r8, alpine-keys 2.5-r0→2.6-r0, +13 more upgraded, +1 new
  Security: 1 perm escalation, 3 new executable, 2 world-writable, 1 SUID/SGID
  Verdict:  !! 7 security finding(s)
```

No scrolling, no noise. One command, full verdict.

---

## Layer Attribution

See which Dockerfile instruction caused each change:

```sh
rift --layers node:18-alpine node:20-alpine
```

```
Comparing node:18-alpine → node:20-alpine (by layer)

Layer 0 (ADD alpine-minirootfs-3.23.3-x86_64.tar.gz / # buildkit)
  ~ bin/busybox  (-4096 bytes)
  ~ sbin/apk  (+44.4 KB)
  + usr/lib/libapk.so.3.0.0  (+270.7 KB)
  ~ usr/lib/libcrypto.so.3  (+471.4 KB)
  Layer total: 601.3 KB

Layer 1 (CMD ["/bin/sh"])
  ~ usr/local/bin/node  (+7.5 MB)
  + usr/local/include/node/cppgc/  (+150 KB)
  ~ usr/local/lib/node_modules/corepack/dist/lib/corepack.cjs  (+4.0 KB)
  Layer total: 7.7 MB

Layer 2 (ENV NODE_VERSION=20.20.1)
  ~ lib/apk/db/installed  (+270 bytes)
  Layer total: -2382 bytes
```

Turns "312 file changes" into "7.7 MB came from the Node binary upgrade in Layer 1."

---

## Secrets Detection

Scan images for leaked credentials, keys, and tokens:

```sh
# Content-based scanning (private keys, AWS keys, API tokens)
rift --secrets myapp:v1 myapp:v2

# Fail CI if secrets are found
rift --secrets --fail-on-security myapp:v1 myapp:v2
```

**Path-based detection** runs automatically on every diff — flags files like `.env`, `id_rsa`, `credentials.json`, `*.pem`, `.aws/credentials`.

**Content-based detection** (`--secrets` flag) scans file content for:
- Private keys (`-----BEGIN RSA PRIVATE KEY-----`)
- AWS access keys (`AKIA...`)
- API tokens and secrets (`api_key=...`, `access_token=...`)

---

## Policy Enforcement

Define rules in `.rift.yml` and enforce them in CI:

```sh
rift init  # Creates .rift.yml template
```

Add a policy section:

```yaml
policy:
  max-size-growth: 50MB
  no-new-suid: true
  no-world-writable: true
  max-new-executables: 10
```

```sh
rift --policy myapp:v1 myapp:v2
```

```
Policy Evaluation
━━━━━━━━━━━━━━━━━━━━━━━━
  [PASS] max-size-growth: within limit
  [PASS] no-new-suid: no SUID/SGID changes
  [FAIL] no-world-writable: found world-writable file at lib/libz.so.1
  [PASS] max-new-executables: 1 new executables (limit: 10)
```

Exits with code 2 if any rule fails — drop it into CI and forget about it.

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

## Output Formats

### Terminal (default)

Color-coded diff with size deltas, layer breakdown, and security highlights:

```sh
rift alpine:3.18 alpine:3.19
```

```
Comparing alpine:3.18 → alpine:3.19

~ bin/busybox  [content]  (-8.0 KB)
- bin/ed  (0 bytes)
+ etc/udhcpc/udhcpc.conf  (+287 bytes)
~ lib/libcrypto.so.3  [content]  (+35.9 KB)

Security Findings (2)
━━━━━━━━━━━━━━━━━━━━━━━━
  [WORLD-WRITABLE] lib/libz.so.1  (0777 → 0777)
  [NEW EXEC] lib/libz.so.1.3.1

Summary: 3 added (+98.1 KB), 3 removed (-97.9 KB), 29 modified
```

### JSON

```sh
rift --format json alpine:3.18 alpine:3.19
```

```json
{
  "image1": "alpine:3.18",
  "image2": "alpine:3.19",
  "summary": {
    "added": 3, "removed": 3, "modified": 29,
    "added_bytes": 100495, "removed_bytes": 100280
  },
  "changes": [...],
  "security_events": [...]
}
```

### Markdown

```sh
rift --format markdown myapp:v1 myapp:v2
```

Ready to paste into PR descriptions or pipe to `$GITHUB_STEP_SUMMARY`.

### SARIF

Upload security findings directly to GitHub Code Scanning:

```sh
rift --format sarif myapp:v1 myapp:v2 > results.sarif
```

```yaml
- name: Run rift security scan
  run: rift --format sarif ${{ env.OLD_IMAGE }} ${{ env.NEW_IMAGE }} > results.sarif

- uses: github/codeql-action/upload-sarif@v3
  with:
    sarif_file: results.sarif
```

---

## Path Filtering

Focus on what matters:

```sh
# Only show changes in /etc
rift --include "etc/**" alpine:3.18 alpine:3.19

# Ignore cache and docs
rift --exclude "var/cache/**" --exclude "usr/share/doc/**" myapp:v1 myapp:v2
```

Glob patterns support `**` for recursive matching.

---

## Package-Level Changes

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

Generates unified diffs for text files like config files and scripts. Binary files and files over 1 MB are skipped.

---

## Interactive TUI

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

rift detects security-relevant changes automatically:

| Event | Trigger |
|-------|---------|
| `new_suid` | Added file with SUID bit set |
| `new_sgid` | Added file with SGID bit set |
| `suid_added` | Modified file gained SUID bit |
| `sgid_added` | Modified file gained SGID bit |
| `new_executable` | Added non-directory file with any execute bit |
| `world_writable` | Added or modified file is world-writable |
| `perm_escalation` | Modified file has strictly more permissive bits |
| `secret_private_key` | File contains a private key |
| `secret_aws_key` | File contains an AWS access key |
| `secret_api_token` | File contains an API key/token pattern |
| `secret_file_path` | File matches a known secret path (.env, id_rsa, etc.) |

```sh
rift --security-only ubuntu:22.04 ubuntu:24.04
rift --fail-on-security myapp:v1 myapp:v2
```

---

## CI/CD Integration

### Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Success — no conditions triggered |
| `1` | Tool error (bad arguments, unreachable image, etc.) |
| `2` | Condition triggered (changes, security, size, or policy failure) |

### GitHub Action

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

### CI pipeline example

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
rift --size-threshold 5MB myapp:v1 myapp:v2   # Fail if grew > 5 MB
rift --exit-code myapp:v1 myapp:v2            # Fail if anything changed
```

Threshold units: `B`, `KB`, `MB`, `GB` (case-insensitive, decimals supported).

### Private registries

```sh
rift --username $REGISTRY_USER --password $REGISTRY_PASS \
  ghcr.io/org/app:v1 ghcr.io/org/app:v2
```

### Multi-arch images

```sh
rift --platform linux/arm64 myapp:v1 myapp:v2
```

---

## Configuration

```sh
rift init  # Creates .rift.yml
```

```yaml
format: terminal
exclude:
  - "var/cache/**"
  - "**/*.pyc"
fail-on-security: true

policy:
  max-size-growth: 50MB
  no-new-suid: true
  no-world-writable: true
  max-new-executables: 10
```

CLI flags override config values. Config is loaded from `.rift.yml` in the current directory, then `~/.config/rift/config.yml`.

---

## All Flags

```
Usage:
  rift <image1> <image2> [flags]

Commands:
  tui            Interactive TUI for browsing image diffs
  init           Create a .rift.yml configuration file
  update         Update rift to the latest version
  completion     Generate shell completion scripts
  version        Print version information

Flags:
  --summary               One-screen verdict (file counts, packages, security, verdict)
  --layers                Group changes by Dockerfile layer
  --secrets               Scan file content for secrets (keys, tokens, credentials)
  --policy                Evaluate policy rules from .rift.yml
  --format string         Output format: terminal, json, markdown, sarif (default "terminal")
  --security-only         Show only security-relevant changes
  --packages              Show package-level changes (APK, DEB)
  --content-diff          Show unified diff for modified text files
  --include strings       Glob patterns to include (repeatable)
  --exclude strings       Glob patterns to exclude (repeatable)
  --quick                 Manifest-only comparison (no content download)
  --platform string       Target platform for multi-arch images
  --username string       Registry username
  --password string       Registry password
  --dockerfile string     Path to Dockerfile for layer-to-instruction mapping
  -v, --verbose           Enable verbose logging to stderr

CI/CD flags:
  --exit-code             Exit 2 if any file changes are found
  --fail-on-security      Exit 2 if security events are detected
  --size-threshold string Exit 2 if net size increase exceeds threshold (e.g., 10MB)
```

---

## Docker CLI Plugin

```sh
./scripts/install-docker-plugin.sh ./rift
docker rift nginx:1.24 nginx:1.25
```

---

## Shell Completions

```sh
source <(rift completion bash)                    # Bash
rift completion zsh > "${fpath[1]}/_rift"          # Zsh
rift completion fish | source                      # Fish
```

---

## Performance

| Feature | Description |
|---------|-------------|
| **Shared layer skip** | Layers with identical digests are not downloaded or parsed |
| **Streaming** | Tar entries are streamed in memory — no disk I/O |
| **`--quick` mode** | Manifest-only comparison, no content download |
| **`--summary` mode** | Computes everything but renders only the verdict |

---

## How It Works

1. **Source detection** — auto-detects registry, daemon, or tarball
2. **Layer skip** — compares layer digests; identical leading layers are skipped entirely
3. **Tree construction** — streams layers into in-memory file trees with layer attribution
4. **Diff** — two-pass comparison with per-file change flags and size deltas
5. **Security analysis** — flags SUID/SGID, world-writable, permission escalation, new executables
6. **Secrets detection** — path-based (always on) and content-based (`--secrets`) scanning
7. **Package detection** — parses APK/DEB databases for package-level changes
8. **Policy evaluation** — checks configurable rules from `.rift.yml`
9. **Output** — terminal, JSON, Markdown, SARIF, or one-screen summary

---

## License

[MIT](LICENSE)
