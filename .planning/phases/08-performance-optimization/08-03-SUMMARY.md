---
phase: 08-performance-optimization
plan: 03
type: summary
status: complete
commits: [5b3d082, edc3a65]
---

# 08-03 Summary: CLI Wiring

## Objective

Wire the two completed optimizations (08-01 layer skip, 08-02 FormatQuick) into
`cmd/root.go` and add a streaming benchmark to `internal/tree/tree_test.go`.

## Changes to cmd/root.go

### Quick-mode early return (Task 1)

Inserted immediately after both images are opened, before any tree building:

```go
if flags.quick {
    summary, err := output.CompareLayers(img1, img2)
    if err != nil {
        fmt.Fprintf(os.Stderr, "warning: layer comparison failed: %v\n", err)
    }
    fmt.Print(output.FormatQuick(summary, args[0], args[1], flags.format))
    return nil
}
```

`--quick` now produces real output in terminal, JSON, or Markdown format
without downloading any layer content. `CompareLayers` uses only `DiffID()`
and `Size()` from the manifest.

### Layer skip optimization (Task 2)

Replaced the two `tree.BuildFromImage` calls with:

```go
skipCount, err := tree.IdenticalLeadingLayers(img1, img2)
if err != nil {
    skipCount = 0  // non-fatal fallback
}
tree1, err := tree.BuildFromImageSkipFirst(img1, skipCount)
tree2, err := tree.BuildFromImageSkipFirst(img2, skipCount)
```

`skipCount=0` on error produces identical behavior to the old code. No calls
to `tree.BuildFromImage` remain in `cmd/root.go`.

## Export Status of IdenticalLeadingLayers

Confirmed exported as `IdenticalLeadingLayers` in `internal/tree/tree.go`
(per 08-01-SUMMARY.md). No rename needed.

## Benchmark Added (Task 3)

`BenchmarkParseLayer_Streaming` in `internal/tree/tree_test.go`:

```
BenchmarkParseLayer_Streaming-8   231   5235270 ns/op   33735295 B/op   19144 allocs/op
```

- 1000 files × 1024 bytes each (1 MB synthetic layer)
- Comment documents the streaming guarantee and how to verify it
- To verify: increase `perFileBytes` to 10000 — alloc count stays ~19144
  while layer size grows 10x

## Verification

- `go build ./...` — clean
- `go vet ./...` — clean
- `go test ./...` — all 6 packages pass, no regressions
- `go test -bench=BenchmarkParseLayer -benchmem ./internal/tree/...` — benchmark runs

## Phase 8 Complete

All three plans complete:
- 08-01: `IdenticalLeadingLayers` + `BuildFromImageSkipFirst` (TDD)
- 08-02: `FormatQuick` with terminal/json/markdown branches (TDD)
- 08-03: CLI wiring + streaming benchmark (this plan)
