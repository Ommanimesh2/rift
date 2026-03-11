package output

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/ommmishra/imgdiff/internal/diff"
)

// terminal styles — defined at package level so they are initialized once.
var (
	addedStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))   // green
	removedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))   // red
	modifiedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))   // yellow
	pathStyle     = lipgloss.NewStyle().Bold(true)
	headerStyle   = lipgloss.NewStyle().Bold(true).Underline(true)
	dimStyle      = lipgloss.NewStyle().Faint(true)
)

// RenderLayerSection renders a LayerSummary as a dim-styled block for display
// between the header and file listing. Returns an empty string if summary is nil.
func RenderLayerSection(summary *LayerSummary) string {
	if summary == nil {
		return ""
	}
	return dimStyle.Render(FormatLayerSummary(summary)) + "\n"
}

// RenderTerminal returns a fully styled terminal diff output string for the
// given DiffResult. image1Name and image2Name are used in the header line.
// layerSummary may be nil, in which case the layer section is omitted.
//
// Output structure:
//  1. Header: "Comparing {image1} → {image2}"
//  2. Blank line
//  3. Layer breakdown (if layerSummary is non-nil)
//  4. Blank line
//  5. One styled entry line per DiffEntry (added=green, removed=red, modified=yellow)
//  6. Blank line
//  7. Summary line with per-category colors
func RenderTerminal(result *diff.DiffResult, image1Name, image2Name string) string {
	return RenderTerminalWithLayers(result, image1Name, image2Name, nil)
}

// RenderTerminalWithLayers is like RenderTerminal but includes an optional
// layer breakdown section when layerSummary is non-nil.
func RenderTerminalWithLayers(result *diff.DiffResult, image1Name, image2Name string, layerSummary *LayerSummary) string {
	var sb strings.Builder

	// Header
	header := fmt.Sprintf("Comparing %s → %s", image1Name, image2Name)
	sb.WriteString(headerStyle.Render(header))
	sb.WriteString("\n\n")

	// Layer breakdown (optional)
	if layerSummary != nil {
		sb.WriteString(RenderLayerSection(layerSummary))
		sb.WriteByte('\n')
	}

	// Per-entry lines
	if len(result.Entries) == 0 {
		sb.WriteString(dimStyle.Render("No differences found."))
		sb.WriteString("\n\n")
	} else {
		for _, entry := range result.Entries {
			sb.WriteString(renderEntry(entry))
			sb.WriteByte('\n')
		}
		sb.WriteByte('\n')
	}

	// Summary
	sb.WriteString(renderSummary(result))
	sb.WriteByte('\n')

	return sb.String()
}

// renderEntry produces a single styled line for a DiffEntry.
func renderEntry(entry *diff.DiffEntry) string {
	delta := dimStyle.Render("(" + FormatSizeDelta(entry.SizeDelta) + ")")

	switch entry.Type {
	case diff.Added:
		prefix := addedStyle.Render("+")
		path := addedStyle.Inherit(pathStyle).Render(entry.Path)
		return fmt.Sprintf("%s %s  %s", prefix, path, delta)

	case diff.Removed:
		prefix := removedStyle.Render("-")
		path := removedStyle.Inherit(pathStyle).Render(entry.Path)
		return fmt.Sprintf("%s %s  %s", prefix, path, delta)

	case diff.Modified:
		prefix := modifiedStyle.Render("~")
		path := modifiedStyle.Inherit(pathStyle).Render(entry.Path)
		flags := dimStyle.Render("[" + collectFlags(entry) + "]")
		return fmt.Sprintf("%s %s  %s  %s", prefix, path, flags, delta)

	default:
		return fmt.Sprintf("? %s", entry.Path)
	}
}

// renderSummary produces a styled summary line with per-category colors.
func renderSummary(result *diff.DiffResult) string {
	var addedStr, removedStr string

	if result.AddedBytes > 0 {
		addedStr = fmt.Sprintf(" (+%s)", FormatBytes(result.AddedBytes))
	}
	if result.RemovedBytes > 0 {
		removedStr = fmt.Sprintf(" (-%s)", FormatBytes(result.RemovedBytes))
	}

	added := addedStyle.Render(fmt.Sprintf("%d added%s", result.Added, addedStr))
	removed := removedStyle.Render(fmt.Sprintf("%d removed%s", result.Removed, removedStr))
	modified := modifiedStyle.Render(fmt.Sprintf("%d modified", result.Modified))

	return fmt.Sprintf("Summary: %s, %s, %s", added, removed, modified)
}
