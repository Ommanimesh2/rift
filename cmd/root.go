// Package cmd contains all cobra command definitions for imgdiff.
package cmd

import (
	"fmt"
	"os"

	"github.com/ommmishra/imgdiff/internal/diff"
	"github.com/ommmishra/imgdiff/internal/exitcode"
	"github.com/ommmishra/imgdiff/internal/output"
	"github.com/ommmishra/imgdiff/internal/security"
	"github.com/ommmishra/imgdiff/internal/source"
	"github.com/ommmishra/imgdiff/internal/tree"
	"github.com/spf13/cobra"
)

// flags holds the values for all persistent flags defined on the root command.
var flags struct {
	format         string
	securityOnly   bool
	quick          bool
	platform       string
	username       string // explicit registry username; overrides DefaultKeychain if non-empty
	password       string // explicit registry password; paired with username
	exitOnChange   bool   // --exit-code: exit 2 if any file changes found
	exitOnSecurity bool   // --fail-on-security: exit 2 if security events detected
	sizeThreshold  string // --size-threshold: exit 2 if net size increase exceeds threshold
}

// rootCmd is the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "imgdiff <image1> <image2>",
	Short: "Compare two container images and show file-level differences",
	Long: `imgdiff is a file-level, security-aware container image diff tool.

Compare two container images and see exactly what changed — files added,
removed, or modified — with size impact, permission changes, and security-
relevant mutations highlighted. Output is color-coded in the terminal and
also available as JSON for CI/CD pipelines or Markdown for PR comments.

Image sources supported:
  - Remote registries (docker.io, ghcr.io, ECR, GCR, etc.)
  - Local Docker daemon (running image name or image ID)
  - OCI tarball archives (./image.tar)`,
	Example: `  imgdiff nginx:1.24 nginx:1.25
  imgdiff myapp:latest myapp:v2.0
  imgdiff --format json alpine:3.18 alpine:3.19
  imgdiff --security-only ubuntu:22.04 ubuntu:24.04
  imgdiff ./old-image.tar ./new-image.tar`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := source.Options{
			Platform: flags.platform,
			Username: flags.username,
			Password: flags.password,
		}

		// Open first image
		img1, err := source.Open(args[0], opts)
		if err != nil {
			return fmt.Errorf("failed to open image %q: %w", args[0], err)
		}

		// Open second image
		img2, err := source.Open(args[1], opts)
		if err != nil {
			return fmt.Errorf("failed to open image %q: %w", args[1], err)
		}

		// Quick mode: manifest-only comparison, no content download.
		if flags.quick {
			summary, err := output.CompareLayers(img1, img2)
			if err != nil {
				fmt.Fprintf(os.Stderr, "warning: layer comparison failed: %v\n", err)
			}
			fmt.Print(output.FormatQuick(summary, args[0], args[1], flags.format))
			return nil
		}

		// Compute identical leading layer count without downloading any content.
		skipCount, err := tree.IdenticalLeadingLayers(img1, img2)
		if err != nil {
			// Non-fatal: fall back to building full trees.
			skipCount = 0
		}

		// Build file tree for first image, skipping identical leading layers.
		tree1, err := tree.BuildFromImageSkipFirst(img1, skipCount)
		if err != nil {
			return fmt.Errorf("failed to build file tree for %q: %w", args[0], err)
		}

		// Build file tree for second image, skipping identical leading layers.
		tree2, err := tree.BuildFromImageSkipFirst(img2, skipCount)
		if err != nil {
			return fmt.Errorf("failed to build file tree for %q: %w", args[1], err)
		}

		result := diff.Diff(tree1, tree2)

		// Compute layer breakdown; skip gracefully on error.
		layerSummary, err := output.CompareLayers(img1, img2)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not compute layer breakdown: %v\n", err)
			layerSummary = nil
		}

		// Run security analysis (pure function, always succeeds).
		events := security.Analyze(result)

		// Handle --security-only flag.
		if flags.securityOnly {
			if len(events) == 0 {
				fmt.Println("No security findings.")
				return nil
			}
			// Build a set of paths that have security events.
			secPaths := make(map[string]struct{}, len(events))
			for _, ev := range events {
				secPaths[ev.Path] = struct{}{}
			}
			// Filter result.Entries to only those with security events.
			filtered := result.Entries[:0]
			for _, entry := range result.Entries {
				if _, ok := secPaths[entry.Path]; ok {
					filtered = append(filtered, entry)
				}
			}
			result.Entries = filtered
		}

		switch flags.format {
		case "terminal", "":
			rendered := output.RenderTerminalWithSecurity(result, args[0], args[1], layerSummary, events)
			fmt.Print(rendered)
		case "json":
			data, err := output.FormatJSON(result, args[0], args[1], events)
			if err != nil {
				return fmt.Errorf("json formatting failed: %w", err)
			}
			fmt.Printf("%s\n", data)
		case "markdown":
			fmt.Print(output.FormatMarkdown(result, args[0], args[1], events))
		default:
			return fmt.Errorf("unknown format %q: supported formats are terminal, json, markdown", flags.format)
		}

		// Parse --size-threshold (validate; error if invalid).
		threshold, thresholdErr := exitcode.ParseSizeThreshold(flags.sizeThreshold)
		if thresholdErr != nil {
			return fmt.Errorf("invalid --size-threshold %q: %w", flags.sizeThreshold, thresholdErr)
		}

		// Evaluate exit code conditions AFTER all output is written.
		// os.Exit(2) is used directly — not returned as error — to avoid cobra
		// printing "Error: ..." to stderr and exiting 1 instead of 2.
		ecOpts := exitcode.Options{
			ExitOnChange:   flags.exitOnChange,
			ExitOnSecurity: flags.exitOnSecurity,
			SizeThreshold:  threshold,
		}
		if code := exitcode.Evaluate(result, events, ecOpts); code != 0 {
			os.Exit(code)
		}

		return nil
	},
}

// Execute runs the root command and exits with code 1 on error.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flags.format, "format", "terminal", "output format: terminal, json, markdown")
	rootCmd.PersistentFlags().BoolVar(&flags.securityOnly, "security-only", false, "show only security-relevant changes")
	rootCmd.PersistentFlags().BoolVar(&flags.quick, "quick", false, "manifest-only comparison (no content download)")
	rootCmd.PersistentFlags().StringVar(&flags.platform, "platform", "", "target platform for multi-arch images (e.g., linux/amd64)")
	rootCmd.PersistentFlags().StringVar(&flags.username, "username", "", "registry username for explicit authentication")
	rootCmd.PersistentFlags().StringVar(&flags.password, "password", "", "registry password for explicit authentication")
	rootCmd.PersistentFlags().BoolVar(&flags.exitOnChange, "exit-code", false, "exit 2 if any file changes are found")
	rootCmd.PersistentFlags().BoolVar(&flags.exitOnSecurity, "fail-on-security", false, "exit 2 if security events are detected")
	rootCmd.PersistentFlags().StringVar(&flags.sizeThreshold, "size-threshold", "", "exit 2 if net size increase exceeds threshold (e.g., 10MB, 500KB)")
}
