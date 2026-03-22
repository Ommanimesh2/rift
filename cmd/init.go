package cmd

import (
	"fmt"
	"os"

	"github.com/Ommanimesh2/rift/internal/config"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a .rift.yml configuration file",
	Long:  "Generate a commented .rift.yml template in the current directory.",
	RunE: func(cmd *cobra.Command, args []string) error {
		path := ".rift.yml"
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("%s already exists", path)
		}

		if err := os.WriteFile(path, []byte(config.DefaultTemplate()), 0o644); err != nil {
			return fmt.Errorf("failed to write %s: %w", path, err)
		}

		fmt.Printf("Created %s\n", path)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
