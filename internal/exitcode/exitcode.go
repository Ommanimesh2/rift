// Package exitcode provides exit code evaluation for CI/CD pipeline integration.
// It determines whether imgdiff should exit with a non-zero status based on
// configurable conditions: file changes, security events, and size growth thresholds.
package exitcode

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ommmishra/imgdiff/internal/diff"
	"github.com/ommmishra/imgdiff/internal/security"
)

// Options configures which conditions trigger a non-zero exit code.
type Options struct {
	// ExitOnChange triggers exit code 2 if any file was added, removed, or modified.
	ExitOnChange bool

	// ExitOnSecurity triggers exit code 2 if any security events are detected.
	ExitOnSecurity bool

	// SizeThreshold triggers exit code 2 if the net size delta (AddedBytes - RemovedBytes)
	// exceeds this value in bytes. Zero means disabled.
	SizeThreshold int64
}

// ParseSizeThreshold parses a human-readable size string to bytes.
// Supported suffixes (case-insensitive): B, KB/K, MB/M, GB/G.
// A bare number (no suffix) is treated as bytes. Empty string or "0" returns 0 (disabled).
// Decimal values are supported (e.g., "1.5MB" → 1572864).
// Returns an error for negative values, unknown suffixes, or non-numeric input.
func ParseSizeThreshold(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, nil
	}

	// Separate numeric prefix from suffix.
	// Find the first non-digit, non-decimal-point, non-sign character.
	i := 0
	for i < len(s) && (s[i] == '-' || s[i] == '+' || s[i] == '.' || (s[i] >= '0' && s[i] <= '9')) {
		i++
	}

	numStr := s[:i]
	suffix := strings.ToUpper(s[i:])

	if numStr == "" {
		return 0, fmt.Errorf("invalid size threshold %q: no numeric value", s)
	}

	value, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size threshold %q: %w", s, err)
	}

	if value < 0 {
		return 0, fmt.Errorf("invalid size threshold %q: negative values are not allowed", s)
	}

	var multiplier float64
	switch suffix {
	case "", "B":
		multiplier = 1
	case "K", "KB":
		multiplier = 1024
	case "M", "MB":
		multiplier = 1024 * 1024
	case "G", "GB":
		multiplier = 1024 * 1024 * 1024
	default:
		return 0, fmt.Errorf("invalid size threshold %q: unknown suffix %q (supported: B, KB, MB, GB)", s, suffix)
	}

	return int64(value * multiplier), nil
}

// Evaluate checks whether any configured condition is triggered by the given diff result
// and security events. Returns 0 if no conditions are triggered, or 2 if any condition
// is triggered. Exit code 2 follows the `diff` command convention for "differences found".
//
// Trigger rules (any one is sufficient to return 2):
//   - ExitOnChange && (result.Added + result.Removed + result.Modified > 0) → 2
//   - ExitOnSecurity && len(events) > 0 → 2
//   - SizeThreshold > 0 && (result.AddedBytes - result.RemovedBytes) > SizeThreshold → 2
func Evaluate(result *diff.DiffResult, events []security.SecurityEvent, opts Options) int {
	if opts.ExitOnChange && (result.Added+result.Removed+result.Modified > 0) {
		return 2
	}

	if opts.ExitOnSecurity && len(events) > 0 {
		return 2
	}

	if opts.SizeThreshold > 0 {
		netDelta := result.AddedBytes - result.RemovedBytes
		if netDelta > opts.SizeThreshold {
			return 2
		}
	}

	return 0
}
