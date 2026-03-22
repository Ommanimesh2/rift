package diff

import (
	"github.com/bmatcuk/doublestar/v4"
)

// FilterEntries returns a new DiffResult containing only entries whose paths
// match at least one include pattern (if any) and no exclude patterns.
// Summary counts and byte totals are recomputed for the filtered set.
func FilterEntries(result *DiffResult, include, exclude []string) *DiffResult {
	if len(include) == 0 && len(exclude) == 0 {
		return result
	}

	filtered := &DiffResult{
		Entries: make([]*DiffEntry, 0, len(result.Entries)),
	}

	for _, entry := range result.Entries {
		if !matchesFilter(entry.Path, include, exclude) {
			continue
		}
		filtered.Entries = append(filtered.Entries, entry)

		switch entry.Type {
		case Added:
			filtered.Added++
			filtered.AddedBytes += entry.After.Size
		case Removed:
			filtered.Removed++
			filtered.RemovedBytes += entry.Before.Size
		case Modified:
			filtered.Modified++
		}
	}

	return filtered
}

// matchesFilter returns true if path passes the include/exclude filters.
// If include is non-empty, path must match at least one include pattern.
// If exclude is non-empty, path must not match any exclude pattern.
func matchesFilter(path string, include, exclude []string) bool {
	if len(include) > 0 {
		matched := false
		for _, pattern := range include {
			if ok, _ := doublestar.Match(pattern, path); ok {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	for _, pattern := range exclude {
		if ok, _ := doublestar.Match(pattern, path); ok {
			return false
		}
	}

	return true
}
