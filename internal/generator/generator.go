package generator

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sarathsp06/proto2astro/internal/config"
	"github.com/sarathsp06/proto2astro/internal/parser"
)

// Init scaffolds a new Astro Starlight site in outDir.
// This creates the base project structure with package.json, components,
// styles, and types. It does NOT generate any API content.
func Init(outDir string, force bool) error {
	fmt.Printf("Scaffolding site in %s ...\n", outDir)
	if err := scaffoldSite(outDir, force); err != nil {
		return fmt.Errorf("scaffold: %w", err)
	}
	fmt.Println("Site scaffold created.")
	fmt.Println("Next steps:")
	fmt.Println("  1. proto2astro generate   (parse protos and generate docs)")
	fmt.Println("  2. proto2astro install     (install npm dependencies)")
	fmt.Println("  3. proto2astro build       (build the static site)")
	return nil
}

// Generate parses proto files and generates a complete documentation site.
// It scaffolds the site (if needed), generates TS data files, MDX pages,
// and the astro.config.mjs with dynamic sidebar.
func Generate(cfg *config.Config) error {
	outDir := cfg.OutDir

	// Validate config
	warnings, err := cfg.Validate()
	for _, w := range warnings {
		fmt.Printf("  warning: %s\n", w)
	}
	if err != nil {
		return err
	}

	// 1. Scaffold site (won't overwrite existing files unless forced)
	fmt.Println("Ensuring site scaffold exists...")
	if err := scaffoldSite(outDir, false); err != nil {
		return fmt.Errorf("scaffold: %w", err)
	}

	// 2. Parse proto files
	paths := cfg.Proto.Paths
	if len(paths) == 0 {
		return fmt.Errorf("no proto paths configured (set proto.paths in proto2astro.yaml)")
	}

	fmt.Printf("Parsing proto files from %v ...\n", paths)
	result, err := parser.ParseFiles(paths)
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}

	// Summary
	totalSvcs := 0
	totalRPCs := 0
	totalMsgs := 0
	totalEnums := 0
	streamingRPCs := 0
	for _, pkg := range result.Packages {
		totalSvcs += len(pkg.Services)
		totalMsgs += len(pkg.Messages)
		totalEnums += len(pkg.Enums)
		for _, svc := range pkg.Services {
			totalRPCs += len(svc.RPCs)
			for _, rpc := range svc.RPCs {
				if rpc.StreamsRequest || rpc.StreamsResponse {
					streamingRPCs++
				}
			}
		}
	}
	fmt.Printf("Found %d package(s), %d service(s), %d RPC(s), %d message(s), %d enum(s)\n",
		len(result.Packages), totalSvcs, totalRPCs, totalMsgs, totalEnums)
	if streamingRPCs > 0 {
		fmt.Printf("  note: %d streaming RPC(s) detected (documented as unary)\n", streamingRPCs)
	}

	// 3. Generate TS data files
	fmt.Println("Generating TypeScript data files...")
	if err := generateDataFiles(result, cfg, outDir); err != nil {
		return fmt.Errorf("generate data: %w", err)
	}

	// 4. Generate pages (MDX stubs, index, comment guide)
	fmt.Println("Generating documentation pages...")
	if err := generatePages(result, cfg, outDir); err != nil {
		return fmt.Errorf("generate pages: %w", err)
	}

	// 5. Generate astro.config.mjs
	fmt.Println("Generating astro.config.mjs...")
	if err := generateAstroConfig(result, cfg, outDir); err != nil {
		return fmt.Errorf("generate config: %w", err)
	}

	// 6. Generate custom pages from config
	if err := generateCustomPages(cfg, outDir); err != nil {
		return fmt.Errorf("generate custom pages: %w", err)
	}

	// Final summary
	pagesGenerated := totalSvcs + totalEnums + 1 /* index */ + 1 /* comment guide */ + 1 /* landing */
	pagesGenerated += len(cfg.CustomPages)
	fmt.Printf("\nGeneration complete: %d pages, %d data files\n", pagesGenerated, totalSvcs+totalEnums)
	fmt.Printf("Output directory: %s\n", outDir)
	fmt.Println("Next steps:")
	fmt.Println("  1. proto2astro install     (install npm dependencies)")
	fmt.Println("  2. proto2astro build       (build the static site)")
	fmt.Println("  3. proto2astro dev         (start dev server for preview)")
	return nil
}

// generateCustomPages writes custom pages defined in proto2astro.yaml.
func generateCustomPages(cfg *config.Config, outDir string) error {
	for _, cp := range cfg.CustomPages {
		if cp.Slug == "" || cp.Content == "" {
			continue
		}
		dir := outDir + "/src/content/docs/guides"
		path := dir + "/" + cp.Slug + ".md"
		content := fmt.Sprintf("---\ntitle: %s\n---\n\n%s\n", cp.Title, cp.Content)
		if err := writeFileEnsureDir(path, []byte(content)); err != nil {
			return fmt.Errorf("write custom page %s: %w", cp.Slug, err)
		}
	}
	return nil
}

// writeFileEnsureDir creates parent directories and writes a file.
func writeFileEnsureDir(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
