package output

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Ommanimesh2/rift/internal/diff"
	"github.com/Ommanimesh2/rift/internal/packages"
	"github.com/Ommanimesh2/rift/internal/security"
	"github.com/Ommanimesh2/rift/internal/tree"
)

var _ = fmt.Sprintf // ensure fmt is used

func TestFormatSummaryReport_Empty(t *testing.T) {
	result := &diff.DiffResult{}
	out := FormatSummaryReport(result, "img1", "img2", nil, nil, nil)

	if !strings.Contains(out, "img1 → img2") {
		t.Error("expected image names")
	}
	if !strings.Contains(out, "0 added, 0 removed, 0 modified") {
		t.Error("expected zero counts")
	}
	if !strings.Contains(out, "no findings") {
		t.Error("expected no findings")
	}
	if !strings.Contains(out, "no changes") {
		t.Error("expected no changes verdict")
	}
}

func TestFormatSummaryReport_WithChanges(t *testing.T) {
	result := &diff.DiffResult{
		Added: 3, Removed: 1, Modified: 10,
		AddedBytes: 5000, RemovedBytes: 1000,
		Entries: []*diff.DiffEntry{
			{Path: "a", Type: diff.Added, After: &tree.FileNode{Size: 100}},
		},
	}
	events := []security.SecurityEvent{
		{Kind: security.KindNewExecutable, Path: "usr/bin/new"},
	}
	pkgChanges := []packages.PackageChange{
		{Name: "curl", Type: "upgraded", OldVersion: "7.88", NewVersion: "8.5"},
		{Name: "wget", Type: "added", NewVersion: "1.21"},
	}

	out := FormatSummaryReport(result, "alpine:3.18", "alpine:3.19", nil, events, pkgChanges)

	if !strings.Contains(out, "3 added, 1 removed, 10 modified") {
		t.Error("expected file counts")
	}
	if !strings.Contains(out, "curl 7.88→8.5") {
		t.Error("expected package upgrade")
	}
	if !strings.Contains(out, "+1 new") {
		t.Error("expected new package count")
	}
	if !strings.Contains(out, "1 new executable") {
		t.Error("expected security summary")
	}
	if !strings.Contains(out, "1 security finding") {
		t.Error("expected security verdict")
	}
}

func TestFormatSummaryReport_WithLayers(t *testing.T) {
	result := &diff.DiffResult{Added: 1}
	ls := &LayerSummary{
		SharedBytes: 100000, OnlyInABytes: 50000, OnlyInBBytes: 70000,
	}

	out := FormatSummaryReport(result, "a", "b", ls, nil, nil)
	if !strings.Contains(out, "Size:") {
		t.Error("expected size line when layer summary provided")
	}
}

func TestFormatSecuritySummary(t *testing.T) {
	events := []security.SecurityEvent{
		{Kind: security.KindNewSUID},
		{Kind: security.KindNewSGID},
		{Kind: security.KindNewExecutable},
		{Kind: security.KindWorldWritable},
	}
	out := formatSecuritySummary(events)
	if !strings.Contains(out, "SUID/SGID") {
		t.Error("expected SUID/SGID in summary")
	}
	if !strings.Contains(out, "new executable") {
		t.Error("expected new executable in summary")
	}
}

func TestFormatPackageSummary_NoChanges(t *testing.T) {
	out := formatPackageSummary(nil)
	if out != "no changes" {
		t.Errorf("expected 'no changes', got %q", out)
	}
}

func TestFormatPackageSummary_ManyUpgrades(t *testing.T) {
	changes := make([]packages.PackageChange, 10)
	for i := range changes {
		changes[i] = packages.PackageChange{
			Name: fmt.Sprintf("pkg%d", i), Type: "upgraded",
			OldVersion: "1.0", NewVersion: "2.0",
		}
	}
	out := formatPackageSummary(changes)
	if !strings.Contains(out, "+7 more upgraded") {
		t.Errorf("expected '+7 more upgraded', got %q", out)
	}
}
