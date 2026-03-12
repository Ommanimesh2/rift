# imgdiff

**File-level, security-aware container image diff tool.**

Compare two container images and see exactly what changed — files added, removed,
or modified — with size impact, permission changes, and security-relevant mutations
highlighted. Output is color-coded in the terminal and also available as JSON or
Markdown for CI/CD pipelines and PR comments.

Fills the gap left by [Google's archived container-diff](https://github.com/GoogleContainerTools/container-diff).

---

## Install

**Build from source (requires Go 1.21+):**

```sh
go install github.com/ommmishra/imgdiff@latest
```

**Clone and build:**

```sh
git clone https://github.com/ommmishra/imgdiff
cd imgdiff
go build -o imgdiff .
```

---

## Quick Start

```sh
# Compare two registry images
imgdiff nginx:1.24 nginx:1.25

# Compare local Docker daemon images
imgdiff myapp:latest myapp:v2.0

# Compare OCI tarball archives
imgdiff ./old.tar ./new.tar

# Output JSON for pipelines
imgdiff --format json alpine:3.18 alpine:3.19

# Show only security-relevant changes
imgdiff --security-only ubuntu:22.04 ubuntu:24.04

# Fast manifest-only check (no content download)
imgdiff --quick nginx:1.24 nginx:1.25
```

---

## Image Sources

imgdiff auto-detects the source type from the reference string:

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
  imgdiff <image1> <image2> [flags]

Flags:
  --format string         Output format: terminal, json, markdown (default "terminal")
  --security-only         Show only security-relevant changes
  --quick                 Manifest-only comparison (no content download)
  --platform string       Target platform for multi-arch images (e.g., linux/amd64)
  --username string       Registry username for explicit authentication
  --password string       Registry password for explicit authentication

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
imgdiff --format json alpine:3.18 alpine:3.19
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
imgdiff --format markdown myapp:v1 myapp:v2
```

Output is a formatted table ready to paste into a PR description or GitHub Actions step summary.

---

## Security Analysis

imgdiff automatically detects security-relevant changes and highlights them in the output.

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
imgdiff --security-only ubuntu:22.04 ubuntu:24.04

# Fail CI on any security finding
imgdiff --fail-on-security ubuntu:22.04 ubuntu:24.04
```

---

## CI/CD Integration

### Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Success — no conditions triggered |
| `1` | Tool error (bad arguments, unreachable image, etc.) |
| `2` | Condition triggered (changes, security events, or size threshold exceeded) |

### GitHub Actions

```yaml
- name: Image diff check
  run: |
    imgdiff \
      --fail-on-security \
      --size-threshold 10MB \
      --format markdown \
      myapp:${{ github.base_ref }} \
      myapp:${{ github.sha }} | tee $GITHUB_STEP_SUMMARY
```

### Size threshold gate

```sh
# Fail if image grew by more than 5 MB
imgdiff --size-threshold 5MB myapp:v1 myapp:v2

# Fail if image grew at all
imgdiff --exit-code myapp:v1 myapp:v2
```

Threshold units: `B`, `KB`, `MB`, `GB` (case-insensitive). Decimals supported: `1.5MB`.

### Private registries

imgdiff uses Docker credential helpers by default — if `docker pull` works, imgdiff works.

For explicit credentials (useful in CI without Docker config):

```sh
imgdiff --username $REGISTRY_USER --password $REGISTRY_PASS \
  ghcr.io/org/app:v1 ghcr.io/org/app:v2
```

### Multi-arch images

```sh
# Default: use host platform
imgdiff myapp:v1 myapp:v2

# Select specific platform
imgdiff --platform linux/arm64 myapp:v1 myapp:v2
imgdiff --platform linux/amd64 myapp:v1 myapp:v2
```

---

## Performance

imgdiff is fast by default and has explicit speed modes:

| Feature | Description |
|---------|-------------|
| **Shared layer skip** | Layers with identical digests are not downloaded or parsed |
| **Streaming** | Tar entries are streamed in memory — no full layer download to disk |
| **`--quick` mode** | Manifest-only comparison using layer digests, no content download at all |

```sh
# Instant comparison via manifest only (registry images)
imgdiff --quick nginx:1.24 nginx:1.25
```

`--quick` shows layer-level changes (which layers were added, removed, or replaced) without
downloading any content. Useful for a fast "did anything change?" check in CI.

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
