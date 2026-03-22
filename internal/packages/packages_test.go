package packages

import (
	"testing"
)

func TestParseAPK(t *testing.T) {
	data := []byte(`P:musl
V:1.2.4-r2
A:x86_64

P:busybox
V:1.36.1-r5
A:x86_64

P:alpine-keys
V:2.4-r1
A:x86_64
`)
	pkgs := ParseAPK(data)
	if len(pkgs) != 3 {
		t.Fatalf("expected 3 packages, got %d", len(pkgs))
	}
	if pkgs[0].Name != "musl" || pkgs[0].Version != "1.2.4-r2" {
		t.Errorf("expected musl 1.2.4-r2, got %s %s", pkgs[0].Name, pkgs[0].Version)
	}
	if pkgs[1].Name != "busybox" {
		t.Errorf("expected busybox, got %s", pkgs[1].Name)
	}
}

func TestParseAPK_NoTrailingNewline(t *testing.T) {
	data := []byte("P:curl\nV:8.5.0-r0\nA:x86_64")
	pkgs := ParseAPK(data)
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}
	if pkgs[0].Name != "curl" {
		t.Errorf("expected curl, got %s", pkgs[0].Name)
	}
}

func TestParseDEB(t *testing.T) {
	data := []byte(`Package: base-files
Status: install ok installed
Version: 13.3
Description: Debian base system miscellaneous files

Package: bash
Status: install ok installed
Version: 5.2.21-2
Description: GNU Bourne Again SHell

`)
	pkgs := ParseDEB(data)
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}
	if pkgs[0].Name != "base-files" || pkgs[0].Version != "13.3" {
		t.Errorf("expected base-files 13.3, got %s %s", pkgs[0].Name, pkgs[0].Version)
	}
	if pkgs[1].Name != "bash" || pkgs[1].Version != "5.2.21-2" {
		t.Errorf("expected bash 5.2.21-2, got %s %s", pkgs[1].Name, pkgs[1].Version)
	}
}

func TestDiffPackages(t *testing.T) {
	old := []Package{
		{Name: "curl", Version: "7.88.1-10"},
		{Name: "bash", Version: "5.2.15-2"},
		{Name: "vim", Version: "9.0.1378-2"},
	}
	new := []Package{
		{Name: "curl", Version: "8.5.0-2"},
		{Name: "bash", Version: "5.2.15-2"},
		{Name: "wget", Version: "1.21.3-1"},
	}

	changes := DiffPackages(old, new)
	if len(changes) != 3 {
		t.Fatalf("expected 3 changes, got %d", len(changes))
	}

	// Sorted alphabetically
	if changes[0].Name != "curl" || changes[0].Type != "upgraded" {
		t.Errorf("expected curl upgraded, got %s %s", changes[0].Name, changes[0].Type)
	}
	if changes[0].OldVersion != "7.88.1-10" || changes[0].NewVersion != "8.5.0-2" {
		t.Errorf("unexpected curl versions: %s -> %s", changes[0].OldVersion, changes[0].NewVersion)
	}
	if changes[1].Name != "vim" || changes[1].Type != "removed" {
		t.Errorf("expected vim removed, got %s %s", changes[1].Name, changes[1].Type)
	}
	if changes[2].Name != "wget" || changes[2].Type != "added" {
		t.Errorf("expected wget added, got %s %s", changes[2].Name, changes[2].Type)
	}
}

func TestDiffPackages_NoChanges(t *testing.T) {
	pkgs := []Package{{Name: "a", Version: "1.0"}}
	changes := DiffPackages(pkgs, pkgs)
	if len(changes) != 0 {
		t.Errorf("expected 0 changes, got %d", len(changes))
	}
}

func TestDetectFormat(t *testing.T) {
	tests := []struct {
		name  string
		paths map[string]bool
		want  string
	}{
		{"alpine", map[string]bool{"lib/apk/db/installed": true}, "apk"},
		{"debian", map[string]bool{"var/lib/dpkg/status": true}, "deb"},
		{"unknown", map[string]bool{"usr/bin/ls": true}, ""},
		{"both prefers apk", map[string]bool{"lib/apk/db/installed": true, "var/lib/dpkg/status": true}, "apk"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectFormat(tt.paths)
			if got != tt.want {
				t.Errorf("DetectFormat() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPackageDBPath(t *testing.T) {
	if PackageDBPath("apk") != "lib/apk/db/installed" {
		t.Error("wrong apk path")
	}
	if PackageDBPath("deb") != "var/lib/dpkg/status" {
		t.Error("wrong deb path")
	}
	if PackageDBPath("unknown") != "" {
		t.Error("expected empty for unknown")
	}
}
