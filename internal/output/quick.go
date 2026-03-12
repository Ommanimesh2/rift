package output

import (
	"encoding/json"
	"fmt"
	"strings"
)

// QuickReport is the top-level JSON structure produced by FormatQuick when
// format == "json". It is exported so tests can unmarshal it directly.
type QuickReport struct {
	Image1 string      `json:"image1"`
	Image2 string      `json:"image2"`
	Mode   string      `json:"mode"`
	Layers QuickLayers `json:"layers"`
}

// QuickLayers holds the layer-count and byte statistics for a quick-mode
// comparison. All fields are zero when the source LayerSummary is nil.
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

// FormatQuick formats a LayerSummary for --quick mode output, respecting the
// same --format values (terminal, json, markdown) as the full diff path.
//
// A nil summary is handled safely in all branches — no panic will occur and
// layer statistics will be zero in the output.
//
// format values:
//   - "terminal" or "": plain-text with header, layer summary, and note
//   - "json": indented JSON (2-space) ending with a trailing newline
//   - "markdown": GitHub-Flavored Markdown with fenced layer summary block
//   - anything else: "unsupported format: <format>"
func FormatQuick(summary *LayerSummary, image1, image2, format string) string {
	switch format {
	case "", "terminal":
		return formatQuickTerminal(summary, image1, image2)
	case "json":
		return formatQuickJSON(summary, image1, image2)
	case "markdown":
		return formatQuickMarkdown(summary, image1, image2)
	default:
		return fmt.Sprintf("unsupported format: %s", format)
	}
}

// formatQuickTerminal produces a plain-text quick-mode output.
// No lipgloss dependency — this file intentionally uses only fmt/strings/encoding/json.
func formatQuickTerminal(summary *LayerSummary, image1, image2 string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Quick comparison: %s → %s\n", image1, image2))
	sb.WriteByte('\n')

	if summary != nil {
		sb.WriteString(FormatLayerSummary(summary))
		sb.WriteByte('\n')
	}

	sb.WriteByte('\n')
	sb.WriteString("(No file content downloaded — run without --quick for file-level diff)\n")

	return sb.String()
}

// formatQuickJSON produces an indented JSON representation of the quick report.
// The output always ends with a trailing newline.
func formatQuickJSON(summary *LayerSummary, image1, image2 string) string {
	var layers QuickLayers
	if summary != nil {
		layers = QuickLayers{
			TotalA:       summary.TotalA,
			TotalB:       summary.TotalB,
			Shared:       summary.SharedCount,
			SharedBytes:  summary.SharedBytes,
			OnlyInA:      summary.OnlyInACount,
			OnlyInABytes: summary.OnlyInABytes,
			OnlyInB:      summary.OnlyInBCount,
			OnlyInBBytes: summary.OnlyInBBytes,
		}
	}

	report := QuickReport{
		Image1: image1,
		Image2: image2,
		Mode:   "quick",
		Layers: layers,
	}

	b, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		// json.MarshalIndent should never fail for this struct, but handle gracefully.
		return fmt.Sprintf(`{"error": %q}`, err.Error()) + "\n"
	}
	return string(b) + "\n"
}

// formatQuickMarkdown produces a GitHub-Flavored Markdown quick-mode report.
func formatQuickMarkdown(summary *LayerSummary, image1, image2 string) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "## Quick Image Comparison: `%s` → `%s`\n", image1, image2)
	sb.WriteByte('\n')
	sb.WriteString("> Manifest-only comparison — no file content downloaded.\n")
	sb.WriteByte('\n')

	if summary != nil {
		sb.WriteString("```\n")
		sb.WriteString(FormatLayerSummary(summary))
		sb.WriteByte('\n')
		sb.WriteString("```\n")
	}

	return sb.String()
}
