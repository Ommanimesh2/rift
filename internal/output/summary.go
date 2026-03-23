package output

import (
	"fmt"
	"strings"

	"github.com/Ommanimesh2/rift/internal/diff"
	"github.com/Ommanimesh2/rift/internal/packages"
	"github.com/Ommanimesh2/rift/internal/security"
)

// FormatSummaryReport produces a concise one-screen verdict of the diff.
func FormatSummaryReport(result *diff.DiffResult, image1, image2 string,
	layerSummary *LayerSummary, events []security.SecurityEvent,
	pkgChanges []packages.PackageChange) string {

	var sb strings.Builder

	// Header
	sb.WriteString(fmt.Sprintf("  Image:    %s → %s\n", image1, image2))

	// Size
	if layerSummary != nil {
		sizeA := layerSummary.OnlyInABytes + layerSummary.SharedBytes
		sizeB := layerSummary.OnlyInBBytes + layerSummary.SharedBytes
		delta := sizeB - sizeA
		sign := "+"
		if delta < 0 {
			sign = ""
		}
		sb.WriteString(fmt.Sprintf("  Size:     %s → %s (%s%s)\n",
			FormatBytes(sizeA), FormatBytes(sizeB), sign, FormatBytes(delta)))
	}

	// Files
	sb.WriteString(fmt.Sprintf("  Files:    %d added, %d removed, %d modified\n",
		result.Added, result.Removed, result.Modified))

	// Packages
	if len(pkgChanges) > 0 {
		sb.WriteString(fmt.Sprintf("  Packages: %s\n", formatPackageSummary(pkgChanges)))
	}

	// Security
	sb.WriteString(fmt.Sprintf("  Security: %s\n", formatSecuritySummary(events)))

	// Verdict
	sb.WriteString(fmt.Sprintf("  Verdict:  %s\n", formatVerdict(result, events)))

	return sb.String()
}

func formatPackageSummary(changes []packages.PackageChange) string {
	var upgraded, added, removed int
	var samples []string

	for _, pc := range changes {
		switch pc.Type {
		case "upgraded":
			upgraded++
			if len(samples) < 3 {
				samples = append(samples, fmt.Sprintf("%s %s→%s", pc.Name, pc.OldVersion, pc.NewVersion))
			}
		case "added":
			added++
		case "removed":
			removed++
		}
	}

	var parts []string
	if len(samples) > 0 {
		parts = append(parts, strings.Join(samples, ", "))
		if upgraded > len(samples) {
			parts = append(parts, fmt.Sprintf("+%d more upgraded", upgraded-len(samples)))
		}
	}
	if added > 0 {
		parts = append(parts, fmt.Sprintf("+%d new", added))
	}
	if removed > 0 {
		parts = append(parts, fmt.Sprintf("-%d removed", removed))
	}

	if len(parts) == 0 {
		return "no changes"
	}
	return strings.Join(parts, ", ")
}

func formatSecuritySummary(events []security.SecurityEvent) string {
	if len(events) == 0 {
		return "no findings"
	}

	counts := make(map[string]int)
	for _, ev := range events {
		switch ev.Kind {
		case security.KindNewSUID, security.KindSUIDAdded, security.KindNewSGID, security.KindSGIDAdded:
			counts["SUID/SGID"]++
		case security.KindNewExecutable:
			counts["new executable"]++
		case security.KindWorldWritable:
			counts["world-writable"]++
		case security.KindPermEscalation:
			counts["perm escalation"]++
		default:
			// Covers secret detection kinds
			counts["secret"]++
		}
	}

	var parts []string
	for kind, count := range counts {
		parts = append(parts, fmt.Sprintf("%d %s", count, kind))
	}
	return strings.Join(parts, ", ")
}

func formatVerdict(result *diff.DiffResult, events []security.SecurityEvent) string {
	if len(events) == 0 && result.Added == 0 && result.Removed == 0 && result.Modified == 0 {
		return "no changes"
	}
	if len(events) > 0 {
		return fmt.Sprintf("!! %d security finding(s)", len(events))
	}
	total := result.Added + result.Removed + result.Modified
	return fmt.Sprintf("%d file change(s), no security findings", total)
}
