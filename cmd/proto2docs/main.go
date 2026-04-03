// proto2docs — Generate Astro Starlight API documentation from proto files.
package main

import (
	"fmt"
	"os"

	"github.com/sarathsp06/proto2docs/internal/buf"
	"github.com/sarathsp06/proto2docs/internal/config"
	"github.com/sarathsp06/proto2docs/internal/generator"
	"github.com/sarathsp06/proto2docs/internal/npm"
	"github.com/spf13/cobra"
)

var version = "dev"

func main() {
	root := &cobra.Command{
		Use:     "proto2docs",
		Short:   "Generate Astro Starlight API docs from proto files",
		Version: version,
	}

	// ── init ──────────────────────────────────────────────
	initCmd := &cobra.Command{
		Use:   "init [output-dir]",
		Short: "Scaffold a new Astro Starlight documentation site",
		Long:  "Creates the base Astro project structure (package.json, components, styles, types) without generating any API content.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			outDir := "./docs"
			if len(args) > 0 {
				outDir = args[0]
			}
			force, _ := cmd.Flags().GetBool("force")
			return generator.Init(outDir, force)
		},
	}
	initCmd.Flags().Bool("force", false, "Overwrite existing scaffold files")
	root.AddCommand(initCmd)

	// ── generate ─────────────────────────────────────────
	generateCmd := &cobra.Command{
		Use:   "generate",
		Short: "Parse proto files and generate documentation site",
		Long:  "Parses proto files, generates TypeScript data files, MDX pages, and astro.config.mjs. Scaffolds the site if it doesn't exist.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgPath, _ := cmd.Flags().GetString("config")
			cfg, err := loadConfig(cfgPath, cmd)
			if err != nil {
				return err
			}
			return generator.Generate(cfg)
		},
	}
	generateCmd.Flags().StringP("config", "c", "proto2docs.yaml", "Path to proto2docs.yaml config file")
	generateCmd.Flags().StringP("proto", "p", "", "Proto file or directory (overrides config)")
	generateCmd.Flags().StringP("out", "o", "", "Output directory (overrides config)")
	generateCmd.Flags().String("buf-workspace", "", "Buf workspace root (overrides config)")
	root.AddCommand(generateCmd)

	// ── install ──────────────────────────────────────────
	installCmd := &cobra.Command{
		Use:   "install",
		Short: "Run npm install in the generated site",
		Long:  "Installs Node.js dependencies (Astro, Starlight, etc.) declared in the site's package.json.",
		RunE: func(cmd *cobra.Command, args []string) error {
			outDir, _ := cmd.Flags().GetString("out")
			fmt.Printf("Installing dependencies in %s ...\n", outDir)
			return npm.Install(npm.RunOptions{Dir: outDir})
		},
	}
	installCmd.Flags().StringP("out", "o", "./docs", "Site directory")
	root.AddCommand(installCmd)

	// ── build ────────────────────────────────────────────
	buildCmd := &cobra.Command{
		Use:   "build",
		Short: "Build the generated site into static HTML",
		Long:  "Runs the Astro static-site-generation build. Output goes to <site-dir>/dist/.",
		RunE: func(cmd *cobra.Command, args []string) error {
			outDir, _ := cmd.Flags().GetString("out")
			fmt.Printf("Building site in %s ...\n", outDir)
			if err := npm.Build(npm.RunOptions{Dir: outDir}); err != nil {
				return err
			}
			fmt.Println("Build complete! Static site is in " + outDir + "/dist/")
			return nil
		},
	}
	buildCmd.Flags().StringP("out", "o", "./docs", "Site directory")
	root.AddCommand(buildCmd)

	// ── dev ──────────────────────────────────────────────
	devCmd := &cobra.Command{
		Use:   "dev",
		Short: "Start the Astro dev server for local preview",
		Long:  "Launches a local development server with hot-reload at http://localhost:4321.",
		RunE: func(cmd *cobra.Command, args []string) error {
			outDir, _ := cmd.Flags().GetString("out")
			fmt.Printf("Starting dev server in %s ...\n", outDir)
			return npm.Dev(npm.RunOptions{Dir: outDir})
		},
	}
	devCmd.Flags().StringP("out", "o", "./docs", "Site directory")
	root.AddCommand(devCmd)

	// ── preview ──────────────────────────────────────────
	previewCmd := &cobra.Command{
		Use:   "preview",
		Short: "Preview the built site locally",
		Long:  "Serves the static HTML from dist/ for final review before deployment.",
		RunE: func(cmd *cobra.Command, args []string) error {
			outDir, _ := cmd.Flags().GetString("out")
			fmt.Printf("Previewing site in %s ...\n", outDir)
			return npm.Preview(npm.RunOptions{Dir: outDir})
		},
	}
	previewCmd.Flags().StringP("out", "o", "./docs", "Site directory")
	root.AddCommand(previewCmd)

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

// loadConfig loads configuration from YAML file and applies CLI flag overrides.
func loadConfig(cfgPath string, cmd *cobra.Command) (*config.Config, error) {
	var cfg *config.Config

	// Try loading config file
	if _, err := os.Stat(cfgPath); err == nil {
		cfg, err = config.Load(cfgPath)
		if err != nil {
			return nil, fmt.Errorf("load config: %w", err)
		}
	} else {
		// No config file — create a default
		cfg = &config.Config{}
		cfg.ApplyDefaults()
	}

	// CLI flag overrides
	if protoPath, _ := cmd.Flags().GetString("proto"); protoPath != "" {
		cfg.Proto.Paths = []string{protoPath}
	}
	if outDir, _ := cmd.Flags().GetString("out"); outDir != "" {
		cfg.OutDir = outDir
	}

	// Buf workspace override or discovery
	if bufWS, _ := cmd.Flags().GetString("buf-workspace"); bufWS != "" {
		files, err := buf.DiscoverProtoFiles(bufWS, cfg.Proto.BufModules)
		if err != nil {
			return nil, fmt.Errorf("discover buf protos: %w", err)
		}
		cfg.Proto.Paths = files
	} else if cfg.Proto.BufWorkspace != "" {
		files, err := buf.DiscoverProtoFiles(cfg.Proto.BufWorkspace, cfg.Proto.BufModules)
		if err != nil {
			return nil, fmt.Errorf("discover buf protos: %w", err)
		}
		cfg.Proto.Paths = files
	}

	return cfg, nil
}
