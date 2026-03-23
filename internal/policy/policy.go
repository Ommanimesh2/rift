// Package policy evaluates configurable rules against diff results.
package policy

import (
	"fmt"

	"github.com/Ommanimesh2/rift/internal/config"
	"github.com/Ommanimesh2/rift/internal/diff"
	"github.com/Ommanimesh2/rift/internal/exitcode"
	"github.com/Ommanimesh2/rift/internal/security"
)

// RuleResult holds the outcome of a single policy rule evaluation.
type RuleResult struct {
	Name    string
	Passed  bool
	Message string
}

// Evaluate runs all configured policy rules against the diff result and security events.
// Returns a result for each configured rule. Rules with nil/zero config values are skipped.
func Evaluate(cfg config.PolicyConfig, result *diff.DiffResult, events []security.SecurityEvent) []RuleResult {
	var results []RuleResult

	if cfg.MaxSizeGrowth != "" {
		results = append(results, evalMaxSizeGrowth(cfg.MaxSizeGrowth, result))
	}
	if cfg.NoNewSUID != nil && *cfg.NoNewSUID {
		results = append(results, evalNoNewSUID(events))
	}
	if cfg.NoWorldWritable != nil && *cfg.NoWorldWritable {
		results = append(results, evalNoWorldWritable(events))
	}
	if cfg.MaxNewExecutables != nil {
		results = append(results, evalMaxNewExecutables(*cfg.MaxNewExecutables, events))
	}

	return results
}

// HasFailures returns true if any rule failed.
func HasFailures(results []RuleResult) bool {
	for _, r := range results {
		if !r.Passed {
			return true
		}
	}
	return false
}

func evalMaxSizeGrowth(threshold string, result *diff.DiffResult) RuleResult {
	limit, err := exitcode.ParseSizeThreshold(threshold)
	if err != nil {
		return RuleResult{
			Name:    "max-size-growth",
			Passed:  false,
			Message: fmt.Sprintf("invalid threshold %q: %v", threshold, err),
		}
	}

	netGrowth := result.AddedBytes - result.RemovedBytes
	if netGrowth > limit {
		return RuleResult{
			Name:    "max-size-growth",
			Passed:  false,
			Message: fmt.Sprintf("net size growth %d bytes exceeds limit %d bytes", netGrowth, limit),
		}
	}

	return RuleResult{Name: "max-size-growth", Passed: true, Message: "within limit"}
}

func evalNoNewSUID(events []security.SecurityEvent) RuleResult {
	for _, ev := range events {
		switch ev.Kind {
		case security.KindNewSUID, security.KindSUIDAdded, security.KindNewSGID, security.KindSGIDAdded:
			return RuleResult{
				Name:    "no-new-suid",
				Passed:  false,
				Message: fmt.Sprintf("found %s at %s", ev.Kind, ev.Path),
			}
		}
	}
	return RuleResult{Name: "no-new-suid", Passed: true, Message: "no SUID/SGID changes"}
}

func evalNoWorldWritable(events []security.SecurityEvent) RuleResult {
	for _, ev := range events {
		if ev.Kind == security.KindWorldWritable {
			return RuleResult{
				Name:    "no-world-writable",
				Passed:  false,
				Message: fmt.Sprintf("found world-writable file at %s", ev.Path),
			}
		}
	}
	return RuleResult{Name: "no-world-writable", Passed: true, Message: "no world-writable files"}
}

func evalMaxNewExecutables(limit int, events []security.SecurityEvent) RuleResult {
	count := 0
	for _, ev := range events {
		if ev.Kind == security.KindNewExecutable {
			count++
		}
	}
	if count > limit {
		return RuleResult{
			Name:    "max-new-executables",
			Passed:  false,
			Message: fmt.Sprintf("found %d new executables, limit is %d", count, limit),
		}
	}
	return RuleResult{
		Name:    "max-new-executables",
		Passed:  true,
		Message: fmt.Sprintf("%d new executables (limit: %d)", count, limit),
	}
}
