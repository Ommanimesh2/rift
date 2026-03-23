package output

import (
	"fmt"
	"strings"

	"github.com/Ommanimesh2/rift/internal/diff"
)

// LayerGroup holds diff entries attributed to a single image layer.
type LayerGroup struct {
	Index      int
	Command    string
	Entries    []*diff.DiffEntry
	TotalDelta int64
}

// GroupByLayer groups diff entries by their source layer index.
// Uses After.LayerIndex for Added/Modified, Before.LayerIndex for Removed.
func GroupByLayer(entries []*diff.DiffEntry) []LayerGroup {
	groups := make(map[int]*LayerGroup)

	for _, entry := range entries {
		idx := -1
		switch entry.Type {
		case diff.Added:
			if entry.After != nil {
				idx = entry.After.LayerIndex
			}
		case diff.Removed:
			if entry.Before != nil {
				idx = entry.Before.LayerIndex
			}
		case diff.Modified:
			if entry.After != nil {
				idx = entry.After.LayerIndex
			}
		}

		g, ok := groups[idx]
		if !ok {
			g = &LayerGroup{Index: idx}
			groups[idx] = g
		}
		g.Entries = append(g.Entries, entry)
		g.TotalDelta += entry.SizeDelta
	}

	// Sort by layer index
	maxIdx := -1
	for idx := range groups {
		if idx > maxIdx {
			maxIdx = idx
		}
	}

	result := make([]LayerGroup, 0, len(groups))
	for i := -1; i <= maxIdx; i++ {
		if g, ok := groups[i]; ok {
			result = append(result, *g)
		}
	}

	return result
}

// FormatLayerAttribution renders diff entries grouped by layer with their commands.
func FormatLayerAttribution(groups []LayerGroup, image1, image2 string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Comparing %s → %s (by layer)\n\n", image1, image2))

	for _, g := range groups {
		cmd := g.Command
		if cmd == "" {
			cmd = "(unknown)"
		}
		// Truncate long commands
		if len(cmd) > 80 {
			cmd = cmd[:77] + "..."
		}

		header := fmt.Sprintf("Layer %d (%s)", g.Index, cmd)
		sb.WriteString(header)
		sb.WriteString("\n")

		for _, entry := range g.Entries {
			var symbol string
			switch entry.Type {
			case diff.Added:
				symbol = "+"
			case diff.Removed:
				symbol = "-"
			case diff.Modified:
				symbol = "~"
			}

			delta := ""
			if entry.SizeDelta != 0 {
				if entry.SizeDelta > 0 {
					delta = fmt.Sprintf("  (+%s)", FormatBytes(entry.SizeDelta))
				} else {
					delta = fmt.Sprintf("  (%s)", FormatBytes(entry.SizeDelta))
				}
			}

			sb.WriteString(fmt.Sprintf("  %s %s%s\n", symbol, entry.Path, delta))
		}

		sb.WriteString(fmt.Sprintf("  Layer total: %s\n\n", FormatBytes(g.TotalDelta)))
	}

	return sb.String()
}
