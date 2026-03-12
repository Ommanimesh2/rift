---
phase: 08-performance-optimization
plan: 02
type: summary
status: complete
---

## Summary: Quick-mode output formatter (08-02)

### What was built

`FormatQuick` in `internal/output/quick.go` — the output function for `--quick`
mode that formats a `*LayerSummary` without any `*diff.DiffResult`, respecting
the same `--format` flag values as the full diff path.

### FormatQuick signature

```go
func FormatQuick(summary *LayerSummary, image1, image2, format string) string
```

### Struct names (exported for test unmarshalling)

```go
type QuickReport struct {
    Image1 string      `json:"image1"`
    Image2 string      `json:"image2"`
    Mode   string      `json:"mode"`
    Layers QuickLayers `json:"layers"`
}

type QuickLayers struct {
    TotalA       int   `json:"total_a"`
    TotalB       int   `json:"total_b"`
    Shared       int   `json:"shared"`
    SharedBytes  int64 `json:"shared_bytes"`
    OnlyInA      int   `json:"only_in_a"`
    OnlyInABytes int64 `json:"only_in_a_bytes"`
    OnlyInB      int   `json:"only_in_b"`
    OnlyInBBytes int64 `json:"only_in_b_bytes"`
}
```

### Format branches

| format | behaviour |
|--------|-----------|
| `"terminal"` or `""` | Plain-text: header + `FormatLayerSummary` + note. No lipgloss dependency. |
| `"json"` | `json.MarshalIndent` of `QuickReport`; trailing newline included. |
| `"markdown"` | `## Quick Image Comparison` header + blockquote + fenced `FormatLayerSummary`. |
| anything else | `"unsupported format: <format>"` |

### Test coverage

16 tests in `internal/output/quick_test.go`:

- `TestFormatQuick_Terminal_ContainsImageNames`
- `TestFormatQuick_Terminal_ContainsSharedCount`
- `TestFormatQuick_Terminal_ContainsNote`
- `TestFormatQuick_EmptyFormat_TreatedAsTerminal`
- `TestFormatQuick_JSON_IsValidJSON`
- `TestFormatQuick_JSON_Image1Field`
- `TestFormatQuick_JSON_ModeField`
- `TestFormatQuick_JSON_LayersSharedField`
- `TestFormatQuick_JSON_TrailingNewline`
- `TestFormatQuick_Markdown_ContainsHeader`
- `TestFormatQuick_Markdown_ContainsImageNames`
- `TestFormatQuick_Markdown_ContainsBlockquote`
- `TestFormatQuick_NilSummary_Terminal_DoesNotPanic`
- `TestFormatQuick_NilSummary_JSON_DoesNotPanicAndIsValidJSON`
- `TestFormatQuick_NilSummary_Markdown_DoesNotPanic`
- `TestFormatQuick_UnknownFormat_ReturnsErrorString`

### TDD commits

| hash | message |
|------|---------|
| `70499d1` | `test(08-02): add failing tests for FormatQuick` |
| `ba5df7f` | `feat(08-02): implement FormatQuick with terminal/json/markdown branches` |

Refactor commit was not needed — implementation was clean on first pass.

### Deviations from plan

None. Field names in `LayerSummary` (`SharedCount`, `OnlyInACount`, `OnlyInBCount`
etc.) matched the plan exactly. All success criteria met.

### Next step

08-03: Wire `FormatQuick` into `cmd/root.go` — hook up the `--quick` flag's
`RunE` handler to call `FormatQuick` and print the result.
