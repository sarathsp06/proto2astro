package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sarathsp06/proto2astro/internal/config"
	"github.com/sarathsp06/proto2astro/internal/parser"
)

// TestGenerateIntegration parses testdata/basic.proto and generates a full site
// in a temp directory, then verifies the expected files exist and contain
// expected content.
func TestGenerateIntegration(t *testing.T) {
	protoFile := filepath.Join(testdataDir(t), "basic.proto")
	if _, err := os.Stat(protoFile); err != nil {
		t.Fatalf("fixture not found: %v", err)
	}

	// Parse the fixture
	result, err := parser.ParseFiles([]string{protoFile})
	if err != nil {
		t.Fatalf("ParseFiles() error = %v", err)
	}

	// Verify parse result has expected structure
	if len(result.Packages) != 1 {
		t.Fatalf("expected 1 package, got %d", len(result.Packages))
	}
	pkg, ok := result.Packages["testpkg.v1"]
	if !ok {
		t.Fatal("expected package testpkg.v1")
	}
	if len(pkg.Services) != 1 {
		t.Errorf("expected 1 service, got %d", len(pkg.Services))
	}
	if len(pkg.Enums) != 2 {
		t.Errorf("expected 2 enums, got %d", len(pkg.Enums))
	}

	// Create temp output directory
	outDir := t.TempDir()

	// Create minimal config
	cfg := &config.Config{
		Title:       "Test API",
		Description: "Test API docs",
		OutDir:      outDir,
		Proto: config.ProtoInput{
			Paths: []string{protoFile},
		},
		EntityTypes: []string{"ItemDetail"},
		EntityExamples: map[string]any{
			"ItemDetail": map[string]any{
				"id":     "item-fallback",
				"status": "active",
			},
		},
	}
	cfg.ApplyDefaults()

	// Scaffold the site first
	if err := Init(outDir, true); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// Generate data files
	if err := generateDataFiles(result, cfg, outDir); err != nil {
		t.Fatalf("generateDataFiles() error = %v", err)
	}

	// Generate pages
	if err := generatePages(result, cfg, outDir); err != nil {
		t.Fatalf("generatePages() error = %v", err)
	}

	// Generate config JSON
	if err := generateProto2AstroConfig(result, cfg, outDir); err != nil {
		t.Fatalf("generateProto2AstroConfig() error = %v", err)
	}

	// Verify expected files exist
	expectedFiles := []string{
		"src/data/api/item-service.ts",
		"src/data/api/enum-item-status.ts",
		"src/data/api/enum-priority.ts",
		"src/content/docs/reference/api/index.md",
		"src/content/docs/reference/api/item-service.mdx",
		"src/content/docs/reference/api/enum-item-status.mdx",
		"src/content/docs/reference/api/enum-priority.mdx",
		"src/content/docs/index.mdx",
		"src/content/docs/guides/comment-guide.md",
		"src/data/proto2astro-config.json",
	}

	for _, rel := range expectedFiles {
		path := filepath.Join(outDir, rel)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected file missing: %s", rel)
		}
	}

	// Verify service data file contains expected content
	svcData, err := os.ReadFile(filepath.Join(outDir, "src/data/api/item-service.ts"))
	if err != nil {
		t.Fatalf("read service data: %v", err)
	}
	svcContent := string(svcData)
	if !strings.Contains(svcContent, `"service": "ItemService"`) {
		t.Error("service data should contain service name")
	}
	if !strings.Contains(svcContent, `"CreateItem"`) {
		t.Error("service data should contain RPC name CreateItem")
	}
	if !strings.Contains(svcContent, `"UpdateItem"`) {
		t.Error("service data should contain RPC name UpdateItem")
	}

	// Verify bug #1 fix: HTML chars are NOT escaped
	if strings.Contains(svcContent, `\u003c`) {
		t.Error("service data should not contain \\u003c (HTML escaping should be disabled)")
	}
	if strings.Contains(svcContent, `\u0026`) {
		t.Error("service data should not contain \\u0026 (HTML escaping should be disabled)")
	}

	// Verify annotations are stripped from descriptions
	if strings.Contains(svcContent, `@example`) {
		t.Error("service data descriptions should not contain @example (should be stripped)")
	}
	if strings.Contains(svcContent, `@required`) {
		t.Error("service data descriptions should not contain @required (should be stripped)")
	}
	if strings.Contains(svcContent, `@default`) {
		t.Error("service data descriptions should not contain @default (should be stripped)")
	}
	if strings.Contains(svcContent, `@range`) {
		t.Error("service data descriptions should not contain @range (should be stripped)")
	}

	// Verify @error annotations produce error entries for UpdateItem
	if !strings.Contains(svcContent, `"NOT_FOUND"`) {
		t.Error("service data should contain NOT_FOUND error code from @error annotation")
	}
	if !strings.Contains(svcContent, `"UpdateItem"`) {
		t.Error("service data should contain UpdateItem RPC")
	}

	// Verify @required fields are marked required
	// The UpdateItem.id field uses @required — verify it shows required: true
	if !strings.Contains(svcContent, `"required": true`) {
		t.Error("service data should have required: true for @required fields")
	}

	// Verify enum data file contains expected content
	enumData, err := os.ReadFile(filepath.Join(outDir, "src/data/api/enum-item-status.ts"))
	if err != nil {
		t.Fatalf("read enum data: %v", err)
	}
	enumContent := string(enumData)
	if !strings.Contains(enumContent, `"name": "ItemStatus"`) {
		t.Error("enum data should contain enum name")
	}
	if !strings.Contains(enumContent, `"ITEM_STATUS_ACTIVE"`) {
		t.Error("enum data should contain enum value")
	}

	// Verify config JSON contains sidebar
	configData, err := os.ReadFile(filepath.Join(outDir, "src/data/proto2astro-config.json"))
	if err != nil {
		t.Fatalf("read config json: %v", err)
	}
	configContent := string(configData)
	if !strings.Contains(configContent, `"sidebar"`) {
		t.Error("config json should contain sidebar")
	}
	if !strings.Contains(configContent, "item-service") {
		t.Error("config json should reference item-service in sidebar")
	}

	// Verify entity type (ItemDetail) is NOT flattened — appears as a single field
	if !strings.Contains(svcContent, `"ItemDetail"`) {
		t.Error("service data should contain ItemDetail as entity type (not flattened)")
	}

	// Phase 5: Verify multi-line @example (fenced block) produces valid example data
	// The metadata field in UpdateItemRequest uses a fenced @example block
	if !strings.Contains(svcContent, `"metadata"`) {
		t.Error("service data should contain metadata field from UpdateItemRequest")
	}
	// The fenced example should produce a parsed JSON object with "key" and "count"
	if !strings.Contains(svcContent, `"key"`) || !strings.Contains(svcContent, `"value"`) {
		t.Error("service data should contain parsed fenced @example JSON for metadata field")
	}

	// Phase 5: Verify entity_examples config provides fallback examples
	// ItemDetail is an entity type with entity_examples configured
	if !strings.Contains(svcContent, `"item-fallback"`) {
		t.Error("service data should contain entity example fallback value for ItemDetail")
	}
}

// TestGenerateIntegration_ScaffoldOnly verifies scaffold-only behavior:
// index.mdx and comment-guide.md should not be overwritten on second run.
func TestGenerateIntegration_ScaffoldOnly(t *testing.T) {
	protoFile := filepath.Join(testdataDir(t), "basic.proto")
	result, err := parser.ParseFiles([]string{protoFile})
	if err != nil {
		t.Fatalf("ParseFiles() error = %v", err)
	}

	outDir := t.TempDir()
	cfg := &config.Config{
		Title:       "Test API",
		Description: "Test API docs",
		OutDir:      outDir,
		Proto:       config.ProtoInput{Paths: []string{protoFile}},
	}
	cfg.ApplyDefaults()

	// First generation
	if err := Init(outDir, true); err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	if err := generatePages(result, cfg, outDir); err != nil {
		t.Fatalf("generatePages() first run error = %v", err)
	}

	// Modify scaffold-only files
	indexPath := filepath.Join(outDir, "src/content/docs/index.mdx")
	customContent := []byte("---\ntitle: Custom\n---\nMy custom landing page")
	if err := os.WriteFile(indexPath, customContent, 0o644); err != nil {
		t.Fatalf("write custom index: %v", err)
	}

	guidePath := filepath.Join(outDir, "src/content/docs/guides/comment-guide.md")
	customGuide := []byte("---\ntitle: Custom Guide\n---\nMy custom guide")
	if err := os.WriteFile(guidePath, customGuide, 0o644); err != nil {
		t.Fatalf("write custom guide: %v", err)
	}

	// Second generation — scaffold-only files should be preserved
	if err := generatePages(result, cfg, outDir); err != nil {
		t.Fatalf("generatePages() second run error = %v", err)
	}

	// Verify custom content is preserved
	gotIndex, _ := os.ReadFile(indexPath)
	if string(gotIndex) != string(customContent) {
		t.Error("index.mdx should not be overwritten (scaffold-only)")
	}

	gotGuide, _ := os.ReadFile(guidePath)
	if string(gotGuide) != string(customGuide) {
		t.Error("comment-guide.md should not be overwritten (scaffold-only)")
	}

	// But regenerated files SHOULD be updated
	apiIndex := filepath.Join(outDir, "src/content/docs/reference/api/index.md")
	gotAPIIndex, _ := os.ReadFile(apiIndex)
	if !strings.Contains(string(gotAPIIndex), "ItemService") {
		t.Error("API index should be regenerated with service listing")
	}
}

// TestGenerateIntegration_ScaffoldDisabled verifies that scaffold files are
// not created when disabled via config.
func TestGenerateIntegration_ScaffoldDisabled(t *testing.T) {
	protoFile := filepath.Join(testdataDir(t), "basic.proto")
	result, err := parser.ParseFiles([]string{protoFile})
	if err != nil {
		t.Fatalf("ParseFiles() error = %v", err)
	}

	outDir := t.TempDir()

	falseBool := false
	cfg := &config.Config{
		Title:       "Test API",
		Description: "Test API docs",
		OutDir:      outDir,
		Proto:       config.ProtoInput{Paths: []string{protoFile}},
		Scaffold: config.ScaffoldConfig{
			LandingPage:  &falseBool,
			CommentGuide: &falseBool,
		},
	}
	cfg.ApplyDefaults()

	// Scaffold the site first
	if err := Init(outDir, true); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// Generate pages
	if err := generatePages(result, cfg, outDir); err != nil {
		t.Fatalf("generatePages() error = %v", err)
	}

	// Verify landing page was NOT created
	indexPath := filepath.Join(outDir, "src/content/docs/index.mdx")
	if _, err := os.Stat(indexPath); err == nil {
		t.Error("index.mdx should not be created when scaffold.landing_page is false")
	}

	// Verify comment guide was NOT created
	guidePath := filepath.Join(outDir, "src/content/docs/guides/comment-guide.md")
	if _, err := os.Stat(guidePath); err == nil {
		t.Error("comment-guide.md should not be created when scaffold.comment_guide is false")
	}

	// Verify API reference pages ARE still generated (they are not scaffold-only)
	apiIndex := filepath.Join(outDir, "src/content/docs/reference/api/index.md")
	if _, err := os.Stat(apiIndex); err != nil {
		t.Error("API index should still be generated when scaffold is disabled")
	}
}

// TestGenerateIntegration_ScaffoldPartial verifies that individual scaffold
// files can be independently disabled.
func TestGenerateIntegration_ScaffoldPartial(t *testing.T) {
	protoFile := filepath.Join(testdataDir(t), "basic.proto")
	result, err := parser.ParseFiles([]string{protoFile})
	if err != nil {
		t.Fatalf("ParseFiles() error = %v", err)
	}

	outDir := t.TempDir()

	falseBool := false
	trueBool := true
	cfg := &config.Config{
		Title:       "Test API",
		Description: "Test API docs",
		OutDir:      outDir,
		Proto:       config.ProtoInput{Paths: []string{protoFile}},
		Scaffold: config.ScaffoldConfig{
			LandingPage:  &trueBool,
			CommentGuide: &falseBool,
		},
	}
	cfg.ApplyDefaults()

	if err := Init(outDir, true); err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	if err := generatePages(result, cfg, outDir); err != nil {
		t.Fatalf("generatePages() error = %v", err)
	}

	// Landing page should be created
	indexPath := filepath.Join(outDir, "src/content/docs/index.mdx")
	if _, err := os.Stat(indexPath); err != nil {
		t.Error("index.mdx should be created when scaffold.landing_page is true")
	}

	// Comment guide should NOT be created
	guidePath := filepath.Join(outDir, "src/content/docs/guides/comment-guide.md")
	if _, err := os.Stat(guidePath); err == nil {
		t.Error("comment-guide.md should not be created when scaffold.comment_guide is false")
	}
}

// TestGenerateIntegration_RouteCollision verifies that the landing page scaffold
// is skipped when src/pages/index.astro already exists.
func TestGenerateIntegration_RouteCollision(t *testing.T) {
	protoFile := filepath.Join(testdataDir(t), "basic.proto")
	result, err := parser.ParseFiles([]string{protoFile})
	if err != nil {
		t.Fatalf("ParseFiles() error = %v", err)
	}

	outDir := t.TempDir()
	cfg := &config.Config{
		Title:       "Test API",
		Description: "Test API docs",
		OutDir:      outDir,
		Proto:       config.ProtoInput{Paths: []string{protoFile}},
	}
	cfg.ApplyDefaults()

	if err := Init(outDir, true); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// Create a conflicting src/pages/index.astro file
	pagesDir := filepath.Join(outDir, "src", "pages")
	if err := os.MkdirAll(pagesDir, 0o755); err != nil {
		t.Fatalf("mkdir pages: %v", err)
	}
	customPage := filepath.Join(pagesDir, "index.astro")
	if err := os.WriteFile(customPage, []byte("<html>Custom</html>"), 0o644); err != nil {
		t.Fatalf("write custom page: %v", err)
	}

	// Generate pages — should skip index.mdx due to route collision
	if err := generatePages(result, cfg, outDir); err != nil {
		t.Fatalf("generatePages() error = %v", err)
	}

	// The scaffold index.mdx should NOT be created
	indexPath := filepath.Join(outDir, "src/content/docs/index.mdx")
	if _, err := os.Stat(indexPath); err == nil {
		t.Error("index.mdx should not be created when src/pages/index.astro exists (route collision)")
	}

	// But comment guide and API pages should still be created
	guidePath := filepath.Join(outDir, "src/content/docs/guides/comment-guide.md")
	if _, err := os.Stat(guidePath); err != nil {
		t.Error("comment-guide.md should still be created despite route collision for index")
	}
}

// TestDetectRouteCollision tests the route collision detection helper.
func TestDetectRouteCollision(t *testing.T) {
	dir := t.TempDir()

	// No collision when directory doesn't exist
	if got := detectRouteCollision(dir, "index"); got != "" {
		t.Errorf("expected no collision, got %q", got)
	}

	// Create src/pages directory
	pagesDir := filepath.Join(dir, "src", "pages")
	if err := os.MkdirAll(pagesDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// No collision when src/pages/ is empty
	if got := detectRouteCollision(dir, "index"); got != "" {
		t.Errorf("expected no collision, got %q", got)
	}

	// Collision with .astro file
	astroFile := filepath.Join(pagesDir, "index.astro")
	if err := os.WriteFile(astroFile, []byte("test"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := detectRouteCollision(dir, "index"); got == "" {
		t.Error("expected collision with index.astro")
	} else if !strings.Contains(got, "index.astro") {
		t.Errorf("collision path should mention index.astro, got %q", got)
	}
	_ = os.Remove(astroFile)

	// Collision with .tsx file
	tsxFile := filepath.Join(pagesDir, "index.tsx")
	if err := os.WriteFile(tsxFile, []byte("test"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := detectRouteCollision(dir, "index"); got == "" {
		t.Error("expected collision with index.tsx")
	}
	_ = os.Remove(tsxFile)

	// No collision for different name
	if got := detectRouteCollision(dir, "about"); got != "" {
		t.Errorf("expected no collision for 'about', got %q", got)
	}
}

// TestScaffoldConfigHelpers tests the ScaffoldConfig helper methods.
func TestScaffoldConfigHelpers(t *testing.T) {
	trueBool := true
	falseBool := false

	tests := []struct {
		name      string
		cfg       config.ScaffoldConfig
		wantPage  bool
		wantGuide bool
	}{
		{
			name:      "zero value (defaults to true)",
			cfg:       config.ScaffoldConfig{},
			wantPage:  true,
			wantGuide: true,
		},
		{
			name:      "explicitly true",
			cfg:       config.ScaffoldConfig{LandingPage: &trueBool, CommentGuide: &trueBool},
			wantPage:  true,
			wantGuide: true,
		},
		{
			name:      "explicitly false",
			cfg:       config.ScaffoldConfig{LandingPage: &falseBool, CommentGuide: &falseBool},
			wantPage:  false,
			wantGuide: false,
		},
		{
			name:      "mixed",
			cfg:       config.ScaffoldConfig{LandingPage: &trueBool, CommentGuide: &falseBool},
			wantPage:  true,
			wantGuide: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.LandingPageEnabled(); got != tt.wantPage {
				t.Errorf("LandingPageEnabled() = %v, want %v", got, tt.wantPage)
			}
			if got := tt.cfg.CommentGuideEnabled(); got != tt.wantGuide {
				t.Errorf("CommentGuideEnabled() = %v, want %v", got, tt.wantGuide)
			}
		})
	}
}

// TestGenerateIntegration_DeterministicOutput verifies that generating the same
// proto files twice produces identical output (Phase 6: deterministic output).
func TestGenerateIntegration_DeterministicOutput(t *testing.T) {
	protoFile := filepath.Join(testdataDir(t), "basic.proto")
	result, err := parser.ParseFiles([]string{protoFile})
	if err != nil {
		t.Fatalf("ParseFiles() error = %v", err)
	}

	generate := func() map[string]string {
		outDir := t.TempDir()
		cfg := &config.Config{
			Title:       "Test API",
			Description: "Test API docs",
			OutDir:      outDir,
			Proto:       config.ProtoInput{Paths: []string{protoFile}},
			EntityTypes: []string{"ItemDetail"},
		}
		cfg.ApplyDefaults()

		if err := Init(outDir, true); err != nil {
			t.Fatalf("Init() error = %v", err)
		}
		if err := generateDataFiles(result, cfg, outDir); err != nil {
			t.Fatalf("generateDataFiles() error = %v", err)
		}
		if err := generatePages(result, cfg, outDir); err != nil {
			t.Fatalf("generatePages() error = %v", err)
		}
		if err := generateProto2AstroConfig(result, cfg, outDir); err != nil {
			t.Fatalf("generateProto2AstroConfig() error = %v", err)
		}

		// Read regenerated files (not scaffold-only)
		files := map[string]string{}
		for _, rel := range []string{
			"src/data/api/item-service.ts",
			"src/data/api/enum-item-status.ts",
			"src/data/api/enum-priority.ts",
			"src/content/docs/reference/api/index.md",
			"src/data/proto2astro-config.json",
		} {
			data, err := os.ReadFile(filepath.Join(outDir, rel))
			if err != nil {
				t.Fatalf("read %s: %v", rel, err)
			}
			files[rel] = string(data)
		}
		return files
	}

	// Generate twice and compare
	run1 := generate()
	run2 := generate()

	for path, content1 := range run1 {
		content2, ok := run2[path]
		if !ok {
			t.Errorf("file %s missing in second run", path)
			continue
		}
		if content1 != content2 {
			t.Errorf("non-deterministic output for %s:\n--- run 1 ---\n%s\n--- run 2 ---\n%s",
				path, content1[:min(200, len(content1))], content2[:min(200, len(content2))])
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
