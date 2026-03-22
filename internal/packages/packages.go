// Package packages detects and compares OS packages between container images.
package packages

import (
	"bufio"
	"bytes"
	"sort"
	"strings"
)

// Package represents an installed OS package.
type Package struct {
	Name    string
	Version string
}

// PackageChange represents a change in package state between images.
type PackageChange struct {
	Name       string
	Type       string // "added", "removed", "upgraded", "downgraded"
	OldVersion string
	NewVersion string
}

// DiffPackages compares two package lists and returns changes.
func DiffPackages(old, new []Package) []PackageChange {
	oldMap := make(map[string]string, len(old))
	for _, p := range old {
		oldMap[p.Name] = p.Version
	}
	newMap := make(map[string]string, len(new))
	for _, p := range new {
		newMap[p.Name] = p.Version
	}

	var changes []PackageChange

	// Added and upgraded/downgraded
	for name, newVer := range newMap {
		oldVer, exists := oldMap[name]
		if !exists {
			changes = append(changes, PackageChange{
				Name:       name,
				Type:       "added",
				NewVersion: newVer,
			})
		} else if oldVer != newVer {
			changes = append(changes, PackageChange{
				Name:       name,
				Type:       "upgraded",
				OldVersion: oldVer,
				NewVersion: newVer,
			})
		}
	}

	// Removed
	for name, oldVer := range oldMap {
		if _, exists := newMap[name]; !exists {
			changes = append(changes, PackageChange{
				Name:       name,
				Type:       "removed",
				OldVersion: oldVer,
			})
		}
	}

	sort.Slice(changes, func(i, j int) bool {
		return changes[i].Name < changes[j].Name
	})

	return changes
}

// ParseAPK parses an Alpine APK installed database (/lib/apk/db/installed).
func ParseAPK(data []byte) []Package {
	var packages []Package
	var name, version string

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "P:") {
			name = strings.TrimPrefix(line, "P:")
		} else if strings.HasPrefix(line, "V:") {
			version = strings.TrimPrefix(line, "V:")
		} else if line == "" && name != "" {
			packages = append(packages, Package{Name: name, Version: version})
			name = ""
			version = ""
		}
	}
	// Handle last entry without trailing blank line
	if name != "" {
		packages = append(packages, Package{Name: name, Version: version})
	}

	return packages
}

// ParseDEB parses a Debian dpkg status file (/var/lib/dpkg/status).
func ParseDEB(data []byte) []Package {
	var packages []Package
	var name, version string

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Package: ") {
			name = strings.TrimPrefix(line, "Package: ")
		} else if strings.HasPrefix(line, "Version: ") {
			version = strings.TrimPrefix(line, "Version: ")
		} else if line == "" && name != "" {
			packages = append(packages, Package{Name: name, Version: version})
			name = ""
			version = ""
		}
	}
	if name != "" {
		packages = append(packages, Package{Name: name, Version: version})
	}

	return packages
}

// DetectFormat returns "apk", "deb", or "" based on which package database
// paths are present in the file list.
func DetectFormat(paths map[string]bool) string {
	if paths["lib/apk/db/installed"] {
		return "apk"
	}
	if paths["var/lib/dpkg/status"] {
		return "deb"
	}
	return ""
}

// PackageDBPath returns the filesystem path for a given package format.
func PackageDBPath(format string) string {
	switch format {
	case "apk":
		return "lib/apk/db/installed"
	case "deb":
		return "var/lib/dpkg/status"
	default:
		return ""
	}
}
