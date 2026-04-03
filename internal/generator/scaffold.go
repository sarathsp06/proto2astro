package generator

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed all:templates/site
var siteFS embed.FS

// scaffoldSite copies the embedded Starlight scaffold to the output directory.
// This creates the base Astro project structure with package.json, components,
// styles, types, etc. It will NOT overwrite existing files if force is false.
func scaffoldSite(outDir string, force bool) error {
	return fs.WalkDir(siteFS, "templates/site", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Compute destination path: strip "templates/site/" prefix
		rel, err := filepath.Rel("templates/site", path)
		if err != nil {
			return err
		}
		dest := filepath.Join(outDir, rel)

		if d.IsDir() {
			return os.MkdirAll(dest, 0o755)
		}

		// Check if file exists and skip if not forcing
		if !force {
			if _, err := os.Stat(dest); err == nil {
				return nil // already exists, skip
			}
		}

		data, err := siteFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read embedded %s: %w", path, err)
		}

		// Ensure parent dir
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}

		return os.WriteFile(dest, data, 0o644)
	})
}
