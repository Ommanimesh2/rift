---
phase: 08-performance-optimization
plan: 01
status: complete
commit: 3ecac3b
---

# 08-01 Summary: Layer Skip Functions

## Objective

Implement two exported functions in `internal/tree/tree.go` to enable skipping
identical leading layers without downloading them, using only `DiffID()` which
reads from the manifest JSON without any network round-trip.

## Functions Added

```go
// IdenticalLeadingLayers returns the count of contiguous identical leading
// layers between two images, compared by DiffID (no network download).
func IdenticalLeadingLayers(img1, img2 v1.Image) (int, error)

// BuildFromImageSkipFirst builds a squashed FileTree skipping the first
// skipFirst layers. Returns an empty FileTree when skipFirst >= len(layers).
func BuildFromImageSkipFirst(img v1.Image, skipFirst int) (*FileTree, error)
```

## Key Design Decision: Contiguous Prefix Only

Only contiguous leading layers are counted as identical and eligible to skip.
Non-prefix identical layers (e.g., layer at index 2 matches when index 1 does
not) are never skipped. This is required for whiteout safety:

A whiteout entry `.wh.foo` in a skipped layer at index 1 would normally delete
`foo` from the merged tree. If we skip layer 1 but process layer 2 (which
contains `foo`), the file would incorrectly appear in the diff result. Only
contiguous prefix layers are guaranteed to contribute no diff when skipped.

## Test Cases Covered

### TestIdenticalLeadingLayers (table-driven, 10 sub-tests)
| Case | img1 | img2 | Expected |
|------|------|------|----------|
| Both empty | [] | [] | 0 |
| All identical (1) | [A] | [A] | 1 |
| All identical (3) | [A,B,C] | [A,B,C] | 3 |
| No identical | [A,B] | [X,Y] | 0 |
| Partial prefix (2/4) | [A,B,C,D] | [A,B,X,Y] | 2 |
| Prefix only | [A,B,X] | [A,B,Y] | 2 |
| Non-contiguous shared | [A,X,B] | [A,Y,B] | 1 |
| Different length, img1 shorter | [A,B] | [A,B,C] | 2 |
| Different length, img2 shorter | [A,B,C] | [A,B] | 2 |
| Single mismatch | [A] | [B] | 0 |

### TestIdenticalLeadingLayers_DiffIDError
- DiffID error on layer i returns i, nil (conservative, non-fatal)

### TestBuildFromImageSkipFirst (3 tests)
- skipFirst=0 produces identical result to BuildFromImage
- skipFirst=1 on a 3-layer image: layer 1 files absent, layers 2+3 present
- skipFirst>=len(layers): returns empty FileTree, no panic

## Test Infrastructure Added

- `fakeDiffIDLayer`: implements `v1.Layer` with configurable DiffID and optional
  error injection; also has optional tarData for Uncompressed() support
- `fakeImageWithLayers`: implements `v1.Image` with only `Layers()` functional;
  all other interface methods return errors (not needed for these tests)
- `hashFor(hex string) v1.Hash`: helper to construct deterministic DiffID values
- `makeImageWithDiffIDs(hexes []string)`: shorthand factory for test cases

## Test Count

- New tests: 14 (11 for IdenticalLeadingLayers, 3 for BuildFromImageSkipFirst)
- Pre-existing tests: 19
- Total: 33 tests, all passing

## Files Modified

- `internal/tree/tree.go`: added `IdenticalLeadingLayers` and `BuildFromImageSkipFirst`
- `internal/tree/tree_test.go`: added tests and helper types

## What's Next

08-03 will wire these functions into `cmd/root.go` to apply the skip optimization
in the main diff command path.
