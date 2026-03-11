---
phase: 05-terminal-output
plan: 03
status: complete
completed: 2026-03-11
---

# 05-03 Summary: Layer Comparison and Breakdown

## What Was Built

Created `internal/output/layers.go` with layer comparison logic and formatter, plus wired it into the CLI output between the header and file listing.

## Files Created / Modified

| File | Action | Description |
|------|--------|-------------|
| `internal/output/layers.go` | Created | `LayerSummary`, `CompareLayers`, `FormatLayerSummary` |
| `internal/output/terminal.go` | Modified | Added `RenderLayerSection`, `RenderTerminalWithLayers` |
| `cmd/root.go` | Modified | Calls `CompareLayers`, passes result to `RenderTerminalWithLayers` |

## Types and Functions

### `LayerSummary`
Struct with SharedCount/SharedBytes, OnlyInACount/OnlyInABytes, OnlyInBCount/OnlyInBBytes, TotalA, TotalB.

### `CompareLayers(imgA, imgB v1.Image) (*LayerSummary, error)`
- Gets layers from both images via `img.Layers()`
- Keyed by `DiffID` (digest of uncompressed content) for identity comparison
- Size collected via `layer.Size()` (compressed size — registry download cost)
- Algorithm: build maps for A and B, iterate A to classify shared vs unique, iterate B to find unique-to-B

### `FormatLayerSummary(summary *LayerSummary) string`
Multi-line output with conditional lines: OnlyInA/OnlyInB lines omitted when count is 0.

## TDD Commits

1. **RED**: `test(05-03): add failing tests for layer comparison` — 9 tests
2. **GREEN**: `feat(05-03): implement layer comparison and formatter` — all 9 pass
3. **WIRING**: `feat(05-03): wire layer breakdown into CLI output`

No REFACTOR commit needed.

## CLI Integration

Layer section appears between header and file listing:
```
Comparing image1 → image2

Layers: 3 → 4            (dim styled)
  Shared:    2 layers (...)
  Only in B: 2 layers (...)

+ added/file   (+X KB)
...
Summary: ...
```

If `CompareLayers` returns an error (e.g. image format doesn't support layer inspection), a warning is logged to stderr and the section is skipped — the diff output is never blocked.

## Test Coverage

9 tests in `output_test` external package:
- `CompareLayers`: identical images, completely different, partially shared, empty image A, single-layer
- `FormatLayerSummary`: all shared (no unique lines), both unique, OnlyInA omitted, OnlyInB omitted

Total output package: 27 tests (18 format + 9 layers).

## Verification

- `go test ./internal/output/... -v` — 27/27 pass
- `go build ./...` — clean
- `go vet ./...` — no issues
- `go test ./...` — all packages pass
