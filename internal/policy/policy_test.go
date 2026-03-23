package policy

import (
	"testing"

	"github.com/Ommanimesh2/rift/internal/config"
	"github.com/Ommanimesh2/rift/internal/diff"
	"github.com/Ommanimesh2/rift/internal/security"
	"github.com/Ommanimesh2/rift/internal/tree"
)

func boolPtr(v bool) *bool { return &v }
func intPtr(v int) *int    { return &v }

func TestEvaluate_EmptyConfig(t *testing.T) {
	cfg := config.PolicyConfig{}
	result := &diff.DiffResult{}
	results := Evaluate(cfg, result, nil)
	if len(results) != 0 {
		t.Errorf("expected 0 results for empty config, got %d", len(results))
	}
}

func TestEvaluate_MaxSizeGrowth_Pass(t *testing.T) {
	cfg := config.PolicyConfig{MaxSizeGrowth: "10MB"}
	result := &diff.DiffResult{AddedBytes: 1000, RemovedBytes: 500}
	results := Evaluate(cfg, result, nil)
	if len(results) != 1 || !results[0].Passed {
		t.Error("expected pass for small size growth")
	}
}

func TestEvaluate_MaxSizeGrowth_Fail(t *testing.T) {
	cfg := config.PolicyConfig{MaxSizeGrowth: "1KB"}
	result := &diff.DiffResult{AddedBytes: 5000, RemovedBytes: 0}
	results := Evaluate(cfg, result, nil)
	if len(results) != 1 || results[0].Passed {
		t.Error("expected fail for exceeding size growth")
	}
}

func TestEvaluate_MaxSizeGrowth_InvalidThreshold(t *testing.T) {
	cfg := config.PolicyConfig{MaxSizeGrowth: "invalid"}
	result := &diff.DiffResult{}
	results := Evaluate(cfg, result, nil)
	if len(results) != 1 || results[0].Passed {
		t.Error("expected fail for invalid threshold")
	}
}

func TestEvaluate_NoNewSUID_Pass(t *testing.T) {
	cfg := config.PolicyConfig{NoNewSUID: boolPtr(true)}
	events := []security.SecurityEvent{
		{Kind: security.KindNewExecutable, Path: "usr/bin/app"},
	}
	results := Evaluate(cfg, &diff.DiffResult{}, events)
	if len(results) != 1 || !results[0].Passed {
		t.Error("expected pass when no SUID events")
	}
}

func TestEvaluate_NoNewSUID_Fail(t *testing.T) {
	cfg := config.PolicyConfig{NoNewSUID: boolPtr(true)}
	events := []security.SecurityEvent{
		{Kind: security.KindNewSUID, Path: "usr/bin/sudo"},
	}
	results := Evaluate(cfg, &diff.DiffResult{}, events)
	if len(results) != 1 || results[0].Passed {
		t.Error("expected fail when SUID event found")
	}
}

func TestEvaluate_NoNewSUID_SGID(t *testing.T) {
	cfg := config.PolicyConfig{NoNewSUID: boolPtr(true)}
	events := []security.SecurityEvent{
		{Kind: security.KindSGIDAdded, Path: "usr/bin/wall"},
	}
	results := Evaluate(cfg, &diff.DiffResult{}, events)
	if len(results) != 1 || results[0].Passed {
		t.Error("expected fail when SGID event found")
	}
}

func TestEvaluate_NoWorldWritable_Pass(t *testing.T) {
	cfg := config.PolicyConfig{NoWorldWritable: boolPtr(true)}
	results := Evaluate(cfg, &diff.DiffResult{}, nil)
	if len(results) != 1 || !results[0].Passed {
		t.Error("expected pass when no events")
	}
}

func TestEvaluate_NoWorldWritable_Fail(t *testing.T) {
	cfg := config.PolicyConfig{NoWorldWritable: boolPtr(true)}
	events := []security.SecurityEvent{
		{Kind: security.KindWorldWritable, Path: "tmp/scratch"},
	}
	results := Evaluate(cfg, &diff.DiffResult{}, events)
	if len(results) != 1 || results[0].Passed {
		t.Error("expected fail for world-writable file")
	}
}

func TestEvaluate_MaxNewExecutables_Pass(t *testing.T) {
	cfg := config.PolicyConfig{MaxNewExecutables: intPtr(5)}
	events := []security.SecurityEvent{
		{Kind: security.KindNewExecutable, Path: "a"},
		{Kind: security.KindNewExecutable, Path: "b"},
	}
	results := Evaluate(cfg, &diff.DiffResult{}, events)
	if len(results) != 1 || !results[0].Passed {
		t.Error("expected pass for 2 executables under limit of 5")
	}
}

func TestEvaluate_MaxNewExecutables_Fail(t *testing.T) {
	cfg := config.PolicyConfig{MaxNewExecutables: intPtr(1)}
	events := []security.SecurityEvent{
		{Kind: security.KindNewExecutable, Path: "a"},
		{Kind: security.KindNewExecutable, Path: "b"},
		{Kind: security.KindNewExecutable, Path: "c"},
	}
	results := Evaluate(cfg, &diff.DiffResult{}, events)
	if len(results) != 1 || results[0].Passed {
		t.Error("expected fail for 3 executables over limit of 1")
	}
}

func TestEvaluate_AllRules(t *testing.T) {
	cfg := config.PolicyConfig{
		MaxSizeGrowth:     "1MB",
		NoNewSUID:         boolPtr(true),
		NoWorldWritable:   boolPtr(true),
		MaxNewExecutables: intPtr(10),
	}
	result := &diff.DiffResult{
		AddedBytes: 500, RemovedBytes: 100,
		Entries: []*diff.DiffEntry{
			{Path: "a", Type: diff.Added, After: &tree.FileNode{Size: 500}},
		},
	}
	results := Evaluate(cfg, result, nil)
	if len(results) != 4 {
		t.Fatalf("expected 4 results, got %d", len(results))
	}
	for _, r := range results {
		if !r.Passed {
			t.Errorf("expected all rules to pass, %s failed: %s", r.Name, r.Message)
		}
	}
}

func TestHasFailures(t *testing.T) {
	if HasFailures(nil) {
		t.Error("nil should have no failures")
	}
	if HasFailures([]RuleResult{{Passed: true}}) {
		t.Error("all passed should have no failures")
	}
	if !HasFailures([]RuleResult{{Passed: true}, {Passed: false}}) {
		t.Error("one failed should have failures")
	}
}
