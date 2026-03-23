// Package cmd contains all cobra command definitions for rift.
package cmd

import (
	"fmt"
	"os"

	"github.com/Ommanimesh2/rift/internal/config"
	"github.com/Ommanimesh2/rift/internal/content"
	"github.com/Ommanimesh2/rift/internal/diff"
	"github.com/Ommanimesh2/rift/internal/exitcode"
	riftlog "github.com/Ommanimesh2/rift/internal/log"
	"github.com/Ommanimesh2/rift/internal/output"
	"github.com/Ommanimesh2/rift/internal/packages"
	"github.com/Ommanimesh2/rift/internal/policy"
	"github.com/Ommanimesh2/rift/internal/secrets"
	"github.com/Ommanimesh2/rift/internal/security"
	"github.com/Ommanimesh2/rift/internal/source"
	"github.com/Ommanimesh2/rift/internal/tree"
	"github.com/spf13/cobra"
)

// flags holds the values for all persistent flags defined on the root command.
var flags struct {
	format         string
	securityOnly   bool
	quick          bool
	platform       string
	username       string   // explicit registry username
	password       string   // explicit registry password
	exitOnChange   bool     // --exit-code
	exitOnSecurity bool     // --fail-on-security
	sizeThreshold  string   // --size-threshold
	include        []string // --include
	exclude        []string // --exclude
	verbose        bool     // --verbose
	contentDiff    bool     // --content-diff
	showPackages   bool     // --packages
	summary        bool     // --summary
	policyCheck    bool     // --policy
	scanSecrets    bool     // --secrets
	showLayers     bool     // --layers
	dockerfile     string   // --dockerfile
}

// rootCmd is the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "rift <image1> <image2>",
	Short: "Compare two container images and show file-level differences",
	Long: `rift is a file-level, security-aware container image diff tool.

Compare two container images and see exactly what changed — files added,
removed, or modified — with size impact, permission changes, and security-
relevant mutations highlighted. Output is color-coded in the terminal and
also available as JSON for CI/CD pipelines or Markdown for PR comments.

Image sources supported:
  - Remote registries (docker.io, ghcr.io, ECR, GCR, etc.)
  - Local Docker daemon (running image name or image ID)
  - OCI tarball archives (./image.tar)`,
	Example: `  rift nginx:1.24 nginx:1.25
  rift myapp:latest myapp:v2.0
  rift --summary alpine:3.18 alpine:3.19
  rift --format json alpine:3.18 alpine:3.19
  rift --security-only ubuntu:22.04 ubuntu:24.04
  rift --secrets --fail-on-security myapp:v1 myapp:v2
  rift --layers alpine:3.18 alpine:3.19
  rift --policy myapp:v1 myapp:v2
  rift --exclude "var/cache/**" alpine:3.18 alpine:3.19
  rift ./old-image.tar ./new-image.tar`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load config file and apply defaults for flags not set on CLI.
		applyConfig(cmd)

		// Summary mode auto-enables packages.
		if flags.summary {
			flags.showPackages = true
		}

		log := riftlog.New(flags.verbose)

		opts := source.Options{
			Platform: flags.platform,
			Username: flags.username,
			Password: flags.password,
		}

		// Open first image
		log.Stepf("Opening image %s", args[0])
		img1, err := source.Open(args[0], opts)
		if err != nil {
			return fmt.Errorf("failed to open image %q: %w", args[0], err)
		}

		// Open second image
		log.Stepf("Opening image %s", args[1])
		img2, err := source.Open(args[1], opts)
		if err != nil {
			return fmt.Errorf("failed to open image %q: %w", args[1], err)
		}

		// Quick mode: manifest-only comparison, no content download.
		if flags.quick {
			log.Step("Quick mode: comparing manifests only")
			summary, err := output.CompareLayers(img1, img2)
			if err != nil {
				fmt.Fprintf(os.Stderr, "warning: layer comparison failed: %v\n", err)
			}
			fmt.Print(output.FormatQuick(summary, args[0], args[1], flags.format))
			return nil
		}

		// Compute identical leading layer count without downloading any content.
		log.Step("Comparing layer digests")
		skipCount, err := tree.IdenticalLeadingLayers(img1, img2)
		if err != nil {
			skipCount = 0
		}
		if skipCount > 0 {
			log.Stepf("Skipping %d identical leading layers", skipCount)
		}

		// Build file trees, skipping identical leading layers.
		log.Stepf("Building file tree for %s", args[0])
		tree1, err := tree.BuildFromImageSkipFirst(img1, skipCount)
		if err != nil {
			return fmt.Errorf("failed to build file tree for %q: %w", args[0], err)
		}

		log.Stepf("Building file tree for %s", args[1])
		tree2, err := tree.BuildFromImageSkipFirst(img2, skipCount)
		if err != nil {
			return fmt.Errorf("failed to build file tree for %q: %w", args[1], err)
		}

		log.Step("Computing diff")
		result := diff.Diff(tree1, tree2)
		log.Stepf("Found %d added, %d removed, %d modified", result.Added, result.Removed, result.Modified)

		// Apply path filters if specified.
		if len(flags.include) > 0 || len(flags.exclude) > 0 {
			log.Stepf("Filtering paths (include=%v, exclude=%v)", flags.include, flags.exclude)
			result = diff.FilterEntries(result, flags.include, flags.exclude)
			log.Stepf("After filtering: %d added, %d removed, %d modified", result.Added, result.Removed, result.Modified)
		}

		// Compute layer breakdown; skip gracefully on error.
		layerSummary, err := output.CompareLayers(img1, img2)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not compute layer breakdown: %v\n", err)
			layerSummary = nil
		}

		// Run security analysis (pure function, always succeeds).
		log.Step("Running security analysis")
		events := security.Analyze(result)

		// Secrets detection: path-based is always on, content-based gated behind --secrets.
		log.Step("Scanning for secret file paths")
		pathFindings := secrets.AnalyzePaths(result)
		secretEvents := secrets.ToSecurityEvents(pathFindings)
		events = append(events, secretEvents...)

		if flags.scanSecrets {
			log.Step("Scanning file content for secrets")
			for _, entry := range result.Entries {
				if entry.Type == diff.Removed {
					continue
				}
				fileContent, err := content.ExtractFile(img2, entry.Path)
				if err != nil || fileContent == nil || !content.IsText(fileContent) {
					continue
				}
				contentFindings := secrets.AnalyzeContent(entry.Path, fileContent)
				events = append(events, secrets.ToSecurityEvents(contentFindings)...)
			}
		}

		if len(events) > 0 {
			log.Stepf("Found %d security events", len(events))
		}

		// Handle --security-only flag.
		if flags.securityOnly {
			if len(events) == 0 {
				fmt.Println("No security findings.")
				return nil
			}
			secPaths := make(map[string]struct{}, len(events))
			for _, ev := range events {
				secPaths[ev.Path] = struct{}{}
			}
			filtered := result.Entries[:0]
			for _, entry := range result.Entries {
				if _, ok := secPaths[entry.Path]; ok {
					filtered = append(filtered, entry)
				}
			}
			result.Entries = filtered
		}

		// Extract content diffs for modified text files if requested.
		var contentDiffs map[string]string
		if flags.contentDiff {
			log.Step("Extracting content diffs for modified text files")
			contentDiffs = make(map[string]string)
			for _, entry := range result.Entries {
				if entry.Type != diff.Modified || !entry.ContentChanged {
					continue
				}
				oldContent, err := content.ExtractFile(img1, entry.Path)
				if err != nil || oldContent == nil {
					continue
				}
				newContent, err := content.ExtractFile(img2, entry.Path)
				if err != nil || newContent == nil {
					continue
				}
				if !content.IsDiffable(oldContent) || !content.IsDiffable(newContent) {
					continue
				}
				d := content.UnifiedDiff(oldContent, newContent, "a/"+entry.Path, "b/"+entry.Path)
				if d != "" {
					contentDiffs[entry.Path] = d
				}
			}
			log.Stepf("Generated %d content diffs", len(contentDiffs))
		}

		// Package-level analysis.
		var pkgChanges []packages.PackageChange
		if flags.showPackages {
			log.Step("Detecting package manager format")
			allPaths := make(map[string]bool)
			for _, e := range result.Entries {
				allPaths[e.Path] = true
			}
			for _, p := range []string{"lib/apk/db/installed", "var/lib/dpkg/status"} {
				allPaths[p] = true
			}

			pkgFormat := packages.DetectFormat(allPaths)
			if pkgFormat != "" {
				dbPath := packages.PackageDBPath(pkgFormat)
				log.Stepf("Found %s package database at %s", pkgFormat, dbPath)

				oldDB, _ := content.ExtractFile(img1, dbPath)
				newDB, _ := content.ExtractFile(img2, dbPath)

				var oldPkgs, newPkgs []packages.Package
				switch pkgFormat {
				case "apk":
					oldPkgs = packages.ParseAPK(oldDB)
					newPkgs = packages.ParseAPK(newDB)
				case "deb":
					oldPkgs = packages.ParseDEB(oldDB)
					newPkgs = packages.ParseDEB(newDB)
				}
				pkgChanges = packages.DiffPackages(oldPkgs, newPkgs)
				log.Stepf("Found %d package changes", len(pkgChanges))
			}
		}

		// Policy evaluation.
		if flags.policyCheck {
			cfg, _ := config.Load()
			if cfg.Policy != nil {
				log.Step("Evaluating policy rules")
				policyResults := policy.Evaluate(*cfg.Policy, result, events)
				fmt.Println("\nPolicy Evaluation")
				fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━")
				for _, r := range policyResults {
					status := "PASS"
					if !r.Passed {
						status = "FAIL"
					}
					fmt.Printf("  [%s] %s: %s\n", status, r.Name, r.Message)
				}
				fmt.Println()
				if policy.HasFailures(policyResults) {
					// Print the normal diff output, then exit 2.
					defer os.Exit(2)
				}
			}
		}

		// Summary mode: one-screen verdict.
		if flags.summary {
			fmt.Print(output.FormatSummaryReport(result, args[0], args[1], layerSummary, events, pkgChanges))
			goto exitEval
		}

		// Layer attribution mode.
		if flags.showLayers {
			log.Step("Grouping changes by layer")
			groups := output.GroupByLayer(result.Entries)

			// Enrich groups with layer commands from image history.
			if cfg2, err := img2.ConfigFile(); err == nil && cfg2 != nil {
				for i := range groups {
					idx := groups[i].Index
					if idx >= 0 && idx < len(cfg2.History) {
						groups[i].Command = cfg2.History[idx].CreatedBy
					}
				}
			}

			fmt.Print(output.FormatLayerAttribution(groups, args[0], args[1]))
			goto exitEval
		}

		// Standard output.
		switch flags.format {
		case "terminal", "":
			rendered := output.RenderTerminalWithSecurity(result, args[0], args[1], layerSummary, events)
			fmt.Print(rendered)
			if len(contentDiffs) > 0 {
				fmt.Println("\n--- Content Diffs ---")
				for _, d := range contentDiffs {
					fmt.Println(d)
				}
			}
			if len(pkgChanges) > 0 {
				fmt.Println("\n--- Package Changes ---")
				for _, pc := range pkgChanges {
					switch pc.Type {
					case "added":
						fmt.Printf("  + %s %s\n", pc.Name, pc.NewVersion)
					case "removed":
						fmt.Printf("  - %s %s\n", pc.Name, pc.OldVersion)
					case "upgraded":
						fmt.Printf("  ~ %s %s → %s\n", pc.Name, pc.OldVersion, pc.NewVersion)
					}
				}
			}
		case "json":
			data, err := output.FormatJSON(result, args[0], args[1], events)
			if err != nil {
				return fmt.Errorf("json formatting failed: %w", err)
			}
			fmt.Printf("%s\n", data)
		case "markdown":
			fmt.Print(output.FormatMarkdown(result, args[0], args[1], events))
		case "sarif":
			data, err := output.FormatSARIF(events, args[0], args[1], version)
			if err != nil {
				return fmt.Errorf("sarif formatting failed: %w", err)
			}
			fmt.Printf("%s\n", data)
		default:
			return fmt.Errorf("unknown format %q: supported formats are terminal, json, markdown, sarif", flags.format)
		}

	exitEval:
		// Parse --size-threshold (validate; error if invalid).
		threshold, thresholdErr := exitcode.ParseSizeThreshold(flags.sizeThreshold)
		if thresholdErr != nil {
			return fmt.Errorf("invalid --size-threshold %q: %w", flags.sizeThreshold, thresholdErr)
		}

		// Evaluate exit code conditions AFTER all output is written.
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
	rootCmd.PersistentFlags().StringVar(&flags.format, "format", "terminal", "output format: terminal, json, markdown, sarif")
	rootCmd.PersistentFlags().BoolVar(&flags.securityOnly, "security-only", false, "show only security-relevant changes")
	rootCmd.PersistentFlags().BoolVar(&flags.quick, "quick", false, "manifest-only comparison (no content download)")
	rootCmd.PersistentFlags().StringVar(&flags.platform, "platform", "", "target platform for multi-arch images (e.g., linux/amd64)")
	rootCmd.PersistentFlags().StringVar(&flags.username, "username", "", "registry username for explicit authentication")
	rootCmd.PersistentFlags().StringVar(&flags.password, "password", "", "registry password for explicit authentication")
	rootCmd.PersistentFlags().BoolVar(&flags.exitOnChange, "exit-code", false, "exit 2 if any file changes are found")
	rootCmd.PersistentFlags().BoolVar(&flags.exitOnSecurity, "fail-on-security", false, "exit 2 if security events are detected")
	rootCmd.PersistentFlags().StringVar(&flags.sizeThreshold, "size-threshold", "", "exit 2 if net size increase exceeds threshold (e.g., 10MB, 500KB)")
	rootCmd.PersistentFlags().StringArrayVar(&flags.include, "include", nil, "glob patterns to include (repeatable)")
	rootCmd.PersistentFlags().StringArrayVar(&flags.exclude, "exclude", nil, "glob patterns to exclude (repeatable)")
	rootCmd.PersistentFlags().BoolVarP(&flags.verbose, "verbose", "v", false, "enable verbose logging to stderr")
	rootCmd.PersistentFlags().BoolVar(&flags.contentDiff, "content-diff", false, "show unified diff for modified text files")
	rootCmd.PersistentFlags().BoolVar(&flags.showPackages, "packages", false, "show package-level changes (APK, DEB)")
	rootCmd.PersistentFlags().BoolVar(&flags.summary, "summary", false, "show one-screen summary instead of per-file listing")
	rootCmd.PersistentFlags().BoolVar(&flags.policyCheck, "policy", false, "evaluate policy rules from .rift.yml")
	rootCmd.PersistentFlags().BoolVar(&flags.scanSecrets, "secrets", false, "scan file content for secrets (keys, tokens, credentials)")
	rootCmd.PersistentFlags().BoolVar(&flags.showLayers, "layers", false, "group changes by Dockerfile layer")
	rootCmd.PersistentFlags().StringVar(&flags.dockerfile, "dockerfile", "", "path to Dockerfile for layer-to-instruction mapping")
}

// applyConfig loads the config file and applies defaults for flags not explicitly set on CLI.
func applyConfig(cmd *cobra.Command) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to load config: %v\n", err)
		return
	}

	if !cmd.Flags().Changed("format") && cfg.Format != "" {
		flags.format = cfg.Format
	}
	if !cmd.Flags().Changed("security-only") && cfg.SecurityOnly != nil {
		flags.securityOnly = *cfg.SecurityOnly
	}
	if !cmd.Flags().Changed("platform") && cfg.Platform != "" {
		flags.platform = cfg.Platform
	}
	if !cmd.Flags().Changed("exit-code") && cfg.ExitCode != nil {
		flags.exitOnChange = *cfg.ExitCode
	}
	if !cmd.Flags().Changed("fail-on-security") && cfg.FailSecurity != nil {
		flags.exitOnSecurity = *cfg.FailSecurity
	}
	if !cmd.Flags().Changed("size-threshold") && cfg.SizeThreshold != "" {
		flags.sizeThreshold = cfg.SizeThreshold
	}
	if !cmd.Flags().Changed("include") && len(cfg.Include) > 0 {
		flags.include = cfg.Include
	}
	if !cmd.Flags().Changed("exclude") && len(cfg.Exclude) > 0 {
		flags.exclude = cfg.Exclude
	}
	if !cmd.Flags().Changed("verbose") && cfg.Verbose != nil {
		flags.verbose = *cfg.Verbose
	}
	if !cmd.Flags().Changed("content-diff") && cfg.ContentDiff != nil {
		flags.contentDiff = *cfg.ContentDiff
	}
}
