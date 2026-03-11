// Package cmd contains all cobra command definitions for imgdiff.
package cmd

import (
	"fmt"
	"os"

	"github.com/ommmishra/imgdiff/internal/diff"
	"github.com/ommmishra/imgdiff/internal/output"
	"github.com/ommmishra/imgdiff/internal/source"
	"github.com/ommmishra/imgdiff/internal/tree"
	"github.com/spf13/cobra"
)

// flags holds the values for all persistent flags defined on the root command.
var flags struct {
	format       string
	securityOnly bool
	quick        bool
	platform     string
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

		// Print confirmation with source type
		fmt.Printf("Opened %s (%s)\n", args[0], source.DetectSourceType(args[0]))

		// Build file tree for first image.
		tree1, err := tree.BuildFromImage(img1)
		if err != nil {
			return fmt.Errorf("failed to build file tree for %q: %w", args[0], err)
		}
		fmt.Printf("  Tree: %s\n", tree1)

		fmt.Printf("Opened %s (%s)\n", args[1], source.DetectSourceType(args[1]))

		// Build file tree for second image.
		tree2, err := tree.BuildFromImage(img2)
		if err != nil {
			return fmt.Errorf("failed to build file tree for %q: %w", args[1], err)
		}
		fmt.Printf("  Tree: %s\n", tree2)

		result := diff.Diff(tree1, tree2)

		fmt.Println()
		if flags.format == "terminal" {
			rendered := output.RenderTerminal(result, args[0], args[1])
			fmt.Print(rendered)
		} else {
			fmt.Fprintf(os.Stderr, "format %q not yet supported, falling back to terminal\n", flags.format)
			rendered := output.RenderTerminal(result, args[0], args[1])
			fmt.Print(rendered)
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
}
