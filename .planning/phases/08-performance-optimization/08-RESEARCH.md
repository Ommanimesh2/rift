# Phase 8: Performance Optimization - Research

**Researched:** 2026-03-12
**Domain:** go-containerregistry v0.21.2 — manifest API, layer digest comparison, streaming tar
**Confidence:** HIGH (verified directly against codebase)

<research_summary>
## Summary

Researched the go-containerregistry API and current codebase implementation to understand what's already optimized and what needs to be built. The key finding: **streaming is already implemented** correctly — `tar.NewReader` processes layers without buffering. Two of the three roadmap optimizations are straightforward wiring tasks.

The codebase has a `--quick` flag that is completely defined but never checked in `cmd/root.go`. Activating it requires a single early-return in RunE after opening images. Layer digest skipping requires comparing `DiffID()` values before calling `Uncompressed()`, with the skip logic added to `BuildTree()` in `internal/tree/tree.go`.

**Primary recommendation:** Three optimizations, two are simple wiring tasks, one requires a focused change to `BuildTree()`. No new dependencies needed. All optimizations use existing `DiffID()` API available from the manifest without network round-trips.
</research_summary>

<standard_stack>
## Standard Stack

### Core (already in go.mod)
| Library | Version | Purpose | Notes |
|---------|---------|---------|-------|
| google/go-containerregistry | v0.21.2 | Image access + manifest API | Already in use |
| archive/tar | stdlib | Streaming tar reader | Already streaming correctly |

### Key APIs (no new dependencies needed)
| API | Download required? | Returns |
|-----|-------------------|---------|
| `layer.DiffID()` | NO — from manifest JSON | `v1.Hash` (uncompressed digest) |
| `layer.Digest()` | NO — from manifest JSON | `v1.Hash` (compressed digest) |
| `layer.Size()` | NO — from manifest JSON | `int64` (compressed bytes) |
| `layer.Uncompressed()` | YES — triggers download | `io.ReadCloser` |
| `img.Manifest()` | NO — first network call | `*v1.Manifest` |
| `img.Layers()` | NO — returns lazy objects | `[]v1.Layer` |

### No alternatives needed
Phase 8 uses only what's already in the project. No new libraries.
</standard_stack>

<architecture_patterns>
## Architecture Patterns

### Current Layer Processing Flow
```
remote.Image(ref)              [1 HTTP GET for manifest JSON]
  └─ img.Layers()              [no network — returns lazy Layer objects]
      └─ layer.DiffID()        [no network — reads from manifest]
      └─ layer.Uncompressed()  [HTTP GET — downloads compressed layer]
          └─ tar.NewReader(rc) [streams — no buffering]
              └─ tr.Next()     [processes one entry at a time]
```

### Pattern 1: Layer Skip Optimization
**What:** Compare DiffIDs before calling `Uncompressed()`. Skip layers present in both images.
**When to use:** Always — called in `BuildTree()` (tree.go:266-272) which currently downloads all layers.
**Implementation target:** `internal/tree/tree.go` — modify `BuildFromImage()` to accept a skip-set, or add a `BuildFromImageSkipLayers()` variant.

```go
// Get DiffIDs from both images without downloading
func sharedLayerDiffIDs(img1, img2 v1.Image) (map[string]bool, error) {
    layers1, err := img1.Layers()
    if err != nil {
        return nil, err
    }
    layers2, err := img2.Layers()
    if err != nil {
        return nil, err
    }

    // Build set of img2 DiffIDs
    img2IDs := make(map[string]bool, len(layers2))
    for _, l := range layers2 {
        id, err := l.DiffID()
        if err != nil {
            continue // non-fatal: just don't skip
        }
        img2IDs[id.String()] = true
    }

    // Find DiffIDs in img1 that are also in img2
    shared := make(map[string]bool)
    for _, l := range layers1 {
        id, err := l.DiffID()
        if err != nil {
            continue
        }
        if img2IDs[id.String()] {
            shared[id.String()] = true
        }
    }
    return shared, nil
}
```

**In BuildTree():** Check DiffID before calling `layer.Uncompressed()`. If DiffID is in shared set, skip.

### Pattern 2: --quick Manifest-Only Mode
**What:** Early return in RunE using existing `CompareLayers()` which already does layer-level comparison.
**When to use:** User passes `--quick` flag.
**Implementation target:** `cmd/root.go` — add check for `flags.quick` after opening both images.

```go
// In RunE, after both images opened:
if flags.quick {
    layerSummary, err := output.CompareLayers(img1, img2)
    if err != nil {
        fmt.Fprintf(os.Stderr, "warning: layer comparison: %v\n", err)
    }
    // Display layer-level summary only (no file tree, no download)
    // Need a dedicated quick-mode output path
    return nil
}
// ... existing full diff code continues
```

**Output for quick mode:** Show layer digest comparison table (already computed by `CompareLayers()`). Consider adding a `FormatQuick()` function or reusing the layer section from terminal output.

### Pattern 3: Streaming (Already Correct — No Change Needed)
**What:** `tar.NewReader` processes tar entries sequentially without buffering.
**Current state:** `ParseLayer()` in tree.go already does this correctly.
**Memory profile:** O(file count × metadata size) — NOT O(layer size). Even 10GB layers process in constant memory.
**Action:** No change needed. Document as already optimized.

### Anti-Patterns to Avoid
- **Collecting all layers into a slice before processing:** The current `BuildFromImage` loop is fine; don't change to a goroutine-based approach (synchronization complexity not worth it for this tool)
- **Downloading layers in parallel:** Risk of overwhelming registry rate limits; current sequential approach is correct
- **Buffering layer content:** Don't add `bytes.Buffer` wrapping around `layer.Uncompressed()` — defeats streaming
</architecture_patterns>

<dont_hand_roll>
## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Layer equality check | Content hashing both layers | `DiffID()` comparison | DiffID is the uncompressed content digest from the manifest — free, no download |
| Manifest fetching | Custom HTTP client | `img.Manifest()` | go-containerregistry handles auth, redirects, retries |
| Streaming tar | Manual io.Read loops | `tar.NewReader()` (already used) | stdlib handles tar format edge cases |
| Layer download | Direct registry HTTP | `layer.Uncompressed()` | go-containerregistry handles decompression, auth |

**Key insight:** `DiffID()` is the canonical "is this layer the same content?" check. It's the uncompressed SHA256 digest stored in the image config. Two layers with the same DiffID are guaranteed identical file content. Use it, don't invent alternatives.
</dont_hand_roll>

<common_pitfalls>
## Common Pitfalls

### Pitfall 1: Skipping Layers Changes Squashing Semantics
**What goes wrong:** If you skip a layer that has whiteouts, the deletion is lost. The squashed tree will incorrectly include files that should be deleted.
**Why it happens:** Whiteout layers must be processed even if the underlying content layer is shared.
**How to avoid:** A layer should only be skipped if it's shared AND neither image has any subsequent layers that differ. In practice: only skip bottom-most identical layers. OR: compare the full layer list pairwise from the bottom up and only skip leading identical layers.
**Warning signs:** Files appear in diff output that shouldn't exist.

**Safer approach:** Only skip layers when the full squashed result would be identical:
- If img1 has layers [A, B, C] and img2 has layers [A, B, D]: only A and B are shared. Skip A and B for both images, then diff B-squash vs D-squash. Actually, simpler: compare DiffID prefix — if layers[0..N] are all identical, skip them; any difference breaks the sequence.

### Pitfall 2: DiffID vs Digest Confusion
**What goes wrong:** Using `layer.Digest()` instead of `layer.DiffID()` for equality comparison.
**Why it happens:** Both return `v1.Hash`. `Digest()` is the compressed layer digest; same content compressed differently → different Digest but same DiffID.
**How to avoid:** Always use `DiffID()` for content equality. `Digest()` is for registry addressing.
**Warning signs:** Layers that should match don't, false "different" results.

### Pitfall 3: --quick Output Format Incomplete
**What goes wrong:** `--quick` mode prints nothing useful or panics because it skips tree building but output functions expect `*diff.Result`.
**Why it happens:** The entire output pipeline assumes a full `diff.Result` struct.
**How to avoid:** Create a dedicated quick-mode output path that only uses `output.LayerSummary`. Don't try to pass nil/empty `diff.Result` to existing formatters.
**Warning signs:** nil pointer dereference on `result.Added`, etc.

### Pitfall 4: Non-TDD for Layer Skip Logic
**What goes wrong:** Layer skip logic has subtle bugs (e.g., pitfall 1 above) that are hard to catch without tests.
**How to avoid:** TDD the `sharedLayerDiffIDs()` and modified `BuildFromImage()` functions with table-driven tests covering: all shared, none shared, partial prefix shared, partial non-prefix shared (should NOT skip).
</common_pitfalls>

<code_examples>
## Code Examples

### Check if layers can be safely skipped (leading prefix only)
```go
// Source: derived from codebase analysis + go-containerregistry API
// Safe: only skip contiguous identical layers from the beginning
func identicalLeadingLayers(img1, img2 v1.Image) (int, error) {
    layers1, err := img1.Layers()
    if err != nil {
        return 0, err
    }
    layers2, err := img2.Layers()
    if err != nil {
        return 0, err
    }

    n := len(layers1)
    if len(layers2) < n {
        n = len(layers2)
    }

    for i := 0; i < n; i++ {
        id1, err := layers1[i].DiffID()
        if err != nil {
            return i, nil
        }
        id2, err := layers2[i].DiffID()
        if err != nil {
            return i, nil
        }
        if id1 != id2 {
            return i, nil
        }
    }
    return n, nil
}
```

### Modified BuildFromImage with skip count
```go
// Source: derived from tree.go BuildFromImage
func BuildFromImageSkipFirst(img v1.Image, skipFirst int) (*FileTree, error) {
    layers, err := img.Layers()
    if err != nil {
        return nil, fmt.Errorf("get layers: %w", err)
    }

    tree := NewFileTree()
    for i, layer := range layers {
        if i < skipFirst {
            continue // skip identical leading layers
        }
        if err := tree.BuildTree(layer); err != nil {
            return nil, fmt.Errorf("layer %d: %w", i, err)
        }
    }
    return tree, nil
}
```

### Early return for --quick mode in cmd/root.go
```go
// In RunE, after img1 and img2 are opened:
if flags.quick {
    summary, err := output.CompareLayers(img1, img2)
    if err != nil {
        // non-fatal
        fmt.Fprintf(os.Stderr, "warning: %v\n", err)
    }
    fmt.Print(output.FormatQuick(summary))
    return nil
}
// ... rest of full diff logic
```
</code_examples>

<sota_updates>
## State of the Art

| Old Approach | Current State | Impact |
|--------------|---------------|--------|
| Always download all layers | Skip identical leading layers | Massive speedup for images sharing a base |
| Full diff only | --quick manifest-only mode | Sub-second comparison for registry images |
| Custom buffering concerns | tar.NewReader already streams | No action needed — streaming is correct |

**What changed:** The codebase already implemented streaming correctly. The two missing optimizations (layer skip and --quick) were roadmapped from day one but not yet wired. This is a "complete what was planned" phase, not a "discover new approach" phase.
</sota_updates>

<open_questions>
## Open Questions

1. **How to handle --quick output format**
   - What we know: `CompareLayers()` returns `*output.LayerSummary`
   - What's unclear: Should --quick output be terminal/JSON/markdown aware? Should it respect `--format`?
   - Recommendation: Yes — check `flags.format` in quick mode too. Create `FormatQuick(summary, format)` or route through existing formatters with a "quick" diff result.

2. **Shared layer skip with different layer counts**
   - What we know: Safe to skip leading identical layers
   - What's unclear: Edge case where img1=[A,B,C] and img2=[A,B,C,D] — should skip A,B,C and only diff against D?
   - Recommendation: Yes — skip ALL leading identical layers regardless of total count difference. Only layers after the first difference need to be processed.
</open_questions>

<sources>
## Sources

### Primary (HIGH confidence)
- Direct codebase analysis: `internal/tree/tree.go`, `cmd/root.go`, `internal/output/layers.go`, `go.mod`
- go-containerregistry v0.21.2 API (verified against actual function calls in codebase)

### Secondary (MEDIUM confidence)
- go-containerregistry source: DiffID is uncompressed layer digest from image config JSON
- OCI spec: DiffID is canonical content identity for layers

### Tertiary (LOW confidence)
- None — all findings are from direct codebase inspection
</sources>

<metadata>
## Metadata

**Research scope:**
- Core technology: go-containerregistry v0.21.2 manifest/layer API
- Ecosystem: stdlib archive/tar (already in use)
- Patterns: Layer skip optimization, streaming (already done), quick mode wiring
- Pitfalls: Whiteout safety, DiffID vs Digest confusion, quick output format

**Confidence breakdown:**
- Standard stack: HIGH — verified against go.mod and existing code
- Architecture: HIGH — derived from actual codebase structure
- Pitfalls: HIGH — pitfall 1 (whiteout) is a logical consequence of squashing semantics
- Code examples: HIGH — derived from existing code patterns in tree.go

**Research date:** 2026-03-12
**Valid until:** 2026-06-12 (90 days — go-containerregistry API is stable)
</metadata>

---

*Phase: 08-performance-optimization*
*Research completed: 2026-03-12*
*Ready for planning: yes*
