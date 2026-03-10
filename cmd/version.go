package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// These variables are set at build time via -ldflags.
// Defaults make the binary self-describing in development builds.
var (
	version   = "dev"
	commitHash = "none"
	buildDate = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  "Print the version, commit hash, and build date of this imgdiff binary.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("imgdiff version %s (commit: %s, built: %s)\n", version, commitHash, buildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
