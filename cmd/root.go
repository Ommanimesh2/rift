// Package cmd contains all cobra command definitions for imgdiff.
package cmd

import (
	"fmt"
	"os"

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
		image1 := args[0]
		image2 := args[1]

		platform := flags.platform
		if platform == "" {
			platform = "host"
		}

		fmt.Printf("Comparing %s vs %s (format: %s, platform: %s", image1, image2, flags.format, platform)
		if flags.securityOnly {
			fmt.Printf(", security-only: true")
		}
		if flags.quick {
			fmt.Printf(", quick: true")
		}
		fmt.Println(")")

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
