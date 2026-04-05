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

	// Verify bug #1 fix: HTML chars are NOT escaped
	if strings.Contains(svcContent, `\u003c`) {
		t.Error("service data should not contain \\u003c (HTML escaping should be disabled)")
	}
	if strings.Contains(svcContent, `\u0026`) {
		t.Error("service data should not contain \\u0026 (HTML escaping should be disabled)")
	}

	// Verify bug #2+#7 fix: @example, Default:, Range: are stripped from descriptions
	if strings.Contains(svcContent, `@example`) {
		t.Error("service data descriptions should not contain @example (should be stripped)")
	}
	if strings.Contains(svcContent, `Default:`) {
		t.Error("service data descriptions should not contain Default: (should be stripped)")
	}
	if strings.Contains(svcContent, `Range:`) {
		t.Error("service data descriptions should not contain Range: (should be stripped)")
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
