package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/Ommanimesh2/rift/internal/diff"
	"github.com/Ommanimesh2/rift/internal/security"
	"github.com/Ommanimesh2/rift/internal/source"
	"github.com/Ommanimesh2/rift/internal/tree"
	"github.com/Ommanimesh2/rift/internal/tui"
)

var tuiCmd = &cobra.Command{
	Use:   "tui <image1> <image2>",
	Short: "Interactive TUI for browsing image diffs",
	Long: `Launch an interactive terminal UI to browse file-level differences
between two container images. Navigate with arrow keys or j/k, search
with /, switch panels with tab, and quit with q.`,
	Example: `  rift tui nginx:1.24 nginx:1.25
  rift tui --platform linux/amd64 myapp:v1 myapp:v2`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := source.Options{
			Platform: flags.platform,
			Username: flags.username,
			Password: flags.password,
		}

		img1, err := source.Open(args[0], opts)
		if err != nil {
			return fmt.Errorf("failed to open image %q: %w", args[0], err)
		}

		img2, err := source.Open(args[1], opts)
		if err != nil {
			return fmt.Errorf("failed to open image %q: %w", args[1], err)
		}

		skipCount, err := tree.IdenticalLeadingLayers(img1, img2)
		if err != nil {
			skipCount = 0
		}

		tree1, err := tree.BuildFromImageSkipFirst(img1, skipCount)
		if err != nil {
			return fmt.Errorf("failed to build file tree for %q: %w", args[0], err)
		}

		tree2, err := tree.BuildFromImageSkipFirst(img2, skipCount)
		if err != nil {
			return fmt.Errorf("failed to build file tree for %q: %w", args[1], err)
		}

		result := diff.Diff(tree1, tree2)

		// Apply path filters if specified.
		if len(flags.include) > 0 || len(flags.exclude) > 0 {
			result = diff.FilterEntries(result, flags.include, flags.exclude)
		}

		events := security.Analyze(result)

		model := tui.New(result, events, args[0], args[1])
		p := tea.NewProgram(model, tea.WithAltScreen())

		if _, err := p.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
			os.Exit(1)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}
