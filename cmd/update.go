package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

const repoAPI = "https://api.github.com/repos/Ommanimesh2/rift/releases/latest"

type githubRelease struct {
	TagName string        `json:"tag_name"`
	Assets  []githubAsset `json:"assets"`
}

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update rift to the latest version",
	Long:  "Check for the latest release on GitHub and replace the current binary.",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("Current version: %s\n", version)
		fmt.Println("Checking for updates...")

		// Fetch latest release info.
		resp, err := http.Get(repoAPI)
		if err != nil {
			return fmt.Errorf("failed to check for updates: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return fmt.Errorf("GitHub API returned %d", resp.StatusCode)
		}

		var release githubRelease
		if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
			return fmt.Errorf("failed to parse release info: %w", err)
		}

		latest := release.TagName
		if latest == version {
			fmt.Println("Already up to date.")
			return nil
		}

		fmt.Printf("New version available: %s\n", latest)

		// Find the right asset for this OS/arch.
		assetName := fmt.Sprintf("rift_%s_%s_%s.tar.gz", strings.TrimPrefix(latest, "v"), runtime.GOOS, runtime.GOARCH)
		var downloadURL string
		for _, asset := range release.Assets {
			if asset.Name == assetName {
				downloadURL = asset.BrowserDownloadURL
				break
			}
		}

		if downloadURL == "" {
			return fmt.Errorf("no binary found for %s/%s in release %s", runtime.GOOS, runtime.GOARCH, latest)
		}

		fmt.Printf("Downloading %s...\n", assetName)

		// Download the tarball.
		dlResp, err := http.Get(downloadURL)
		if err != nil {
			return fmt.Errorf("download failed: %w", err)
		}
		defer dlResp.Body.Close()

		if dlResp.StatusCode != 200 {
			return fmt.Errorf("download returned %d", dlResp.StatusCode)
		}

		// Write to temp file.
		tmpFile, err := os.CreateTemp("", "rift-update-*.tar.gz")
		if err != nil {
			return fmt.Errorf("failed to create temp file: %w", err)
		}
		tmpPath := tmpFile.Name()
		defer os.Remove(tmpPath)

		if _, err := io.Copy(tmpFile, dlResp.Body); err != nil {
			tmpFile.Close()
			return fmt.Errorf("download failed: %w", err)
		}
		tmpFile.Close()

		// Find the current binary path.
		execPath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("cannot determine binary path: %w", err)
		}

		// Extract the rift binary from the tarball and replace current binary.
		fmt.Printf("Installing to %s...\n", execPath)

		extractCmd := fmt.Sprintf("tar xzf %s -C %s rift", tmpPath, os.TempDir())
		if err := runShell(extractCmd); err != nil {
			return fmt.Errorf("extraction failed: %w", err)
		}

		extractedPath := fmt.Sprintf("%s/rift", os.TempDir())
		defer os.Remove(extractedPath)

		// Replace the binary.
		if err := os.Rename(extractedPath, execPath); err != nil {
			// Rename fails across devices; fall back to copy.
			if err := copyFile(extractedPath, execPath); err != nil {
				return fmt.Errorf("failed to install: %w", err)
			}
		}

		if err := os.Chmod(execPath, 0o755); err != nil {
			return fmt.Errorf("failed to set permissions: %w", err)
		}

		fmt.Printf("Updated to %s\n", latest)
		return nil
	},
}

func runShell(command string) error {
	c := exec.Command("sh", "-c", command)
	c.Stderr = os.Stderr
	return c.Run()
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
