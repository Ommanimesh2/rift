package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

// dockerPluginMetadata is the metadata Docker CLI expects from plugins.
type dockerPluginMetadata struct {
	SchemaVersion    string `json:"SchemaVersion"`
	Vendor           string `json:"Vendor"`
	Version          string `json:"Version"`
	ShortDescription string `json:"ShortDescription"`
}

var dockerCLIPluginMetadataCmd = &cobra.Command{
	Use:    "docker-cli-plugin-metadata",
	Short:  "Docker CLI plugin metadata",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		meta := dockerPluginMetadata{
			SchemaVersion:    "0.1.0",
			Vendor:           "Ommanimesh2",
			Version:          version,
			ShortDescription: "File-level diff for container images",
		}
		data, _ := json.Marshal(meta)
		fmt.Println(string(data))
	},
}

func init() {
	rootCmd.AddCommand(dockerCLIPluginMetadataCmd)
}
