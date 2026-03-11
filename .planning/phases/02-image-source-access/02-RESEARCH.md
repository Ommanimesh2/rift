# Phase 2: Image Source Access - Research

**Researched:** 2026-03-11
**Domain:** go-containerregistry for container image access
**Confidence:** HIGH

<research_summary>
## Summary

Researched google/go-containerregistry (v0.21.2, March 2026) for accessing container images from three sources: remote registries, local Docker daemon, and OCI tarball archives. The library provides a unified `v1.Image` interface backed by different source implementations.

The standard approach is straightforward: use `name.ParseReference()` for parsing image references, then delegate to source-specific functions (`remote.Image()`, `daemon.Image()`, `tarball.ImageFromPath()`) that all return the same `v1.Image` interface. Authentication uses `authn.DefaultKeychain` which reads Docker config automatically.

**Primary recommendation:** Create a thin abstraction that auto-detects source type from the reference string and delegates to the appropriate go-containerregistry function.
</research_summary>

<standard_stack>
## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| github.com/google/go-containerregistry | v0.21.2 | Image access (remote, daemon, tarball) | The standard Go library for container images — used by crane, ko, kaniko, tekton |

### Key Packages
| Package | Purpose |
|---------|---------|
| `pkg/name` | Parse image references (tag, digest, registry) |
| `pkg/v1` | Core interfaces: `Image`, `Layer`, `Platform` |
| `pkg/v1/remote` | Pull images from registries |
| `pkg/v1/daemon` | Read images from Docker daemon |
| `pkg/v1/tarball` | Read images from .tar files |
| `pkg/authn` | Authentication (DefaultKeychain reads Docker config) |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| go-containerregistry | containers/image | containers/image is heavier, more C dependencies |
| authn.DefaultKeychain | Manual auth config | DefaultKeychain handles Docker config, credential helpers automatically |
</standard_stack>

<architecture_patterns>
## Architecture Patterns

### Recommended Package Structure
```
internal/
└── source/
    ├── source.go       # Types, detection, Open() resolver
    ├── remote.go       # openRemote() — registry access
    ├── daemon.go       # openDaemon() — Docker daemon
    ├── tarball.go      # openTarball() — .tar files
    └── source_test.go  # Table-driven tests for detection
```

### Pattern 1: Unified v1.Image Interface
**What:** All source types return `v1.Image` — downstream code never knows the source
**When to use:** Always — this is the core abstraction

### Pattern 2: Reference Auto-Detection
**What:** Classify references by inspecting the string (file path vs registry ref vs daemon prefix)
**When to use:** CLI tools that accept flexible input
**Detection logic:**
1. `daemon://` prefix → daemon source (strip prefix)
2. File exists on disk OR has .tar/.tar.gz/.tgz extension → tarball
3. Otherwise → remote registry (default)

### Anti-Patterns to Avoid
- **Wrapping v1.Image:** Don't create a custom image interface — use v1.Image directly
- **Eager layer download:** go-containerregistry is lazy by default — layers download on access, not on Image() call
- **Ignoring platform:** Multi-arch images need `remote.WithPlatform()` or you get the wrong architecture
</architecture_patterns>

<dont_hand_roll>
## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Image reference parsing | Regex for registry/tag/digest | `name.ParseReference()` | Handles all edge cases (default registry, default tag, digests, ports) |
| Registry auth | Manual credential reading | `authn.DefaultKeychain` | Reads Docker config.json, credential helpers, platform keychains |
| Tar manifest parsing | Manual JSON parsing from tarball | `tarball.ImageFromPath()` | Handles both OCI and Docker save formats |
| Platform resolution | Manual manifest index walking | `remote.WithPlatform()` | Resolves multi-arch indexes to correct platform |

**Key insight:** go-containerregistry handles all the OCI/Docker format complexity. The v1.Image interface is lazy — layers aren't downloaded until accessed.
</dont_hand_roll>

<common_pitfalls>
## Common Pitfalls

### Pitfall 1: Missing Platform Specification for Multi-Arch Images
**What goes wrong:** `remote.Image()` without platform returns wrong architecture or error on index
**Why it happens:** Many images are multi-arch manifests, not single images
**How to avoid:** Always pass `remote.WithPlatform()` when platform flag is set; detect host platform as default
**Warning signs:** "unsupported media type" errors or wrong binary architecture

### Pitfall 2: Daemon Image Buffering
**What goes wrong:** Memory spikes when reading large images from daemon
**Why it happens:** Default daemon.Image() buffers entire image in memory
**How to avoid:** Consider `daemon.WithUnbufferedOpener()` or `daemon.WithFileBufferedOpener()` for large images
**Warning signs:** OOM on large images (>1GB)

### Pitfall 3: Tarball Tag Requirement
**What goes wrong:** tarball.ImageFromPath fails or returns wrong image
**Why it happens:** Tarballs can contain multiple images; tag selects which one
**How to avoid:** Pass `nil` tag to get the first (and usually only) image; handle multi-image tarballs gracefully
**Warning signs:** Wrong layers returned from tarball
</common_pitfalls>

<code_examples>
## Code Examples

### Remote Image with Auth and Platform
```go
// Source: pkg.go.dev/github.com/google/go-containerregistry/pkg/v1/remote
ref, err := name.ParseReference("nginx:1.25")
img, err := remote.Image(ref,
    remote.WithAuthFromKeychain(authn.DefaultKeychain),
    remote.WithPlatform(v1.Platform{
        Architecture: "amd64",
        OS:           "linux",
    }),
)
```

### Daemon Image
```go
// Source: pkg.go.dev/github.com/google/go-containerregistry/pkg/v1/daemon
ref, err := name.NewTag("nginx:latest")
img, err := daemon.Image(ref)
```

### Tarball Image
```go
// Source: pkg.go.dev/github.com/google/go-containerregistry/pkg/v1/tarball
img, err := tarball.ImageFromPath("./image.tar", nil) // nil = first image
```

### Platform Parsing
```go
// Parse "linux/amd64" into v1.Platform
func parsePlatform(s string) (v1.Platform, error) {
    parts := strings.Split(s, "/")
    if len(parts) != 2 {
        return v1.Platform{}, fmt.Errorf("invalid platform %q: expected os/arch", s)
    }
    return v1.Platform{OS: parts[0], Architecture: parts[1]}, nil
}
```
</code_examples>

<sources>
## Sources

### Primary (HIGH confidence)
- [pkg.go.dev/github.com/google/go-containerregistry/pkg/v1/remote](https://pkg.go.dev/github.com/google/go-containerregistry/pkg/v1/remote) — remote.Image(), options, auth
- [pkg.go.dev/github.com/google/go-containerregistry/pkg/v1/daemon](https://pkg.go.dev/github.com/google/go-containerregistry/pkg/v1/daemon) — daemon.Image(), buffering options
- [pkg.go.dev/github.com/google/go-containerregistry/pkg/v1/tarball](https://pkg.go.dev/github.com/google/go-containerregistry/pkg/v1/tarball) — tarball.ImageFromPath(), Opener

### Secondary (MEDIUM confidence)
- [github.com/google/go-containerregistry](https://github.com/google/go-containerregistry) — latest version v0.21.2 confirmed
</sources>

<metadata>
## Metadata

**Research scope:**
- Core technology: go-containerregistry v0.21.2
- Ecosystem: remote, daemon, tarball, authn packages
- Patterns: Reference detection, unified v1.Image, lazy loading
- Pitfalls: Multi-arch, daemon buffering, tarball tags

**Confidence breakdown:**
- Standard stack: HIGH — verified with official pkg.go.dev docs
- Architecture: HIGH — well-established patterns in Go ecosystem
- Pitfalls: HIGH — documented in official package docs
- Code examples: HIGH — from official documentation

**Research date:** 2026-03-11
**Valid until:** 2026-04-11 (30 days — go-containerregistry is stable)
</metadata>

---

*Phase: 02-image-source-access*
*Research completed: 2026-03-11*
*Ready for planning: yes*
