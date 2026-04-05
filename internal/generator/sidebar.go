package generator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/sarathsp06/proto2astro/internal/config"
	"github.com/sarathsp06/proto2astro/internal/parser"
)

// proto2astroConfig is the JSON structure written to src/data/proto2astro-config.json.
// It is imported by the scaffold's astro.config.mjs so that all proto-derived
// and YAML-derived settings survive regeneration while astro.config.mjs itself
// remains user-editable.
type proto2astroConfig struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Site        string            `json:"site,omitempty"`
	Base        string            `json:"base,omitempty"`
	Logo        string            `json:"logo,omitempty"`
	Social      []socialLink      `json:"social,omitempty"`
	EditLink    string            `json:"editLink,omitempty"`
	Components  map[string]string `json:"components,omitempty"`
	CustomCSS   []string          `json:"customCss,omitempty"`
	Sidebar     []sidebarGroup    `json:"sidebar"`
}

// socialLink is the JSON representation of a social/external link.
type socialLink struct {
	Icon  string `json:"icon"`
	Label string `json:"label"`
	Href  string `json:"href"`
}

// sidebarGroup is a single sidebar section in the JSON output.
type sidebarGroup struct {
	Label string             `json:"label"`
	Items []sidebarGroupItem `json:"items"`
}

// sidebarGroupItem is a single entry in a sidebar group.
type sidebarGroupItem struct {
	Label string `json:"label,omitempty"`
	Slug  string `json:"slug"`
}

// generateProto2AstroConfig writes src/data/proto2astro-config.json.
// This file is regenerated on every `generate` run and contains all
// config-derived and proto-derived settings that astro.config.mjs reads.
func generateProto2AstroConfig(result *parser.ParseResult, cfg *config.Config, outDir string) error {
	data := buildConfigJSON(result, cfg)

	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal proto2astro config: %w", err)
	}

	path := filepath.Join(outDir, "src", "data", "proto2astro-config.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("mkdir for config json: %w", err)
	}
	if err := os.WriteFile(path, b, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

// buildConfigJSON constructs the full JSON config from parsed protos and YAML config.
func buildConfigJSON(result *parser.ParseResult, cfg *config.Config) proto2astroConfig {
	// Build sidebar
	var sidebar []sidebarGroup

	// Before sections
	for _, sec := range cfg.Sidebar.Before {
		sidebar = append(sidebar, configSectionToGroup(sec))
	}

	// API Reference section(s)
	sidebar = append(sidebar, buildAPIReferenceSidebar(result, cfg)...)

	// After sections
	for _, sec := range cfg.Sidebar.After {
		sidebar = append(sidebar, configSectionToGroup(sec))
	}

	// Social links
	var social []socialLink
	for _, l := range cfg.Social {
		social = append(social, socialLink{Icon: l.Icon, Label: l.Label, Href: l.Href})
	}

	return proto2astroConfig{
		Title:       cfg.Title,
		Description: cfg.Description,
		Site:        cfg.Site,
		Base:        cfg.Base,
		Logo:        cfg.Logo,
		Social:      social,
		EditLink:    cfg.EditLink,
		Components:  cfg.Components,
		CustomCSS:   cfg.CustomCSS,
		Sidebar:     sidebar,
	}
}

// configSectionToGroup converts a YAML sidebar section to a JSON sidebar group.
func configSectionToGroup(sec config.SidebarSection) sidebarGroup {
	var items []sidebarGroupItem
	for _, item := range sec.Items {
		items = append(items, sidebarGroupItem{Label: item.Label, Slug: item.Slug})
	}
	return sidebarGroup{Label: sec.Label, Items: items}
}

// buildAPIReferenceSidebar builds the API Reference sidebar groups from parsed protos.
func buildAPIReferenceSidebar(result *parser.ParseResult, cfg *config.Config) []sidebarGroup {
	var pkgNames []string
	for name := range result.Packages {
		pkgNames = append(pkgNames, name)
	}
	sort.Strings(pkgNames)

	multiPackage := len(pkgNames) > 1

	if multiPackage {
		// One sidebar group per package
		var groups []sidebarGroup
		for _, pkgName := range pkgNames {
			pkg := result.Packages[pkgName]
			var items []sidebarGroupItem
			items = append(items, sidebarGroupItem{Slug: "reference/api"})
			for _, name := range sortedServiceNames(pkg, cfg.ServiceOrder) {
				items = append(items, sidebarGroupItem{Slug: "reference/api/" + toKebab(name)})
			}
			for _, name := range sortedEnumNames(pkg) {
				items = append(items, sidebarGroupItem{Slug: "reference/api/enum-" + toKebab(name)})
			}
			groups = append(groups, sidebarGroup{
				Label: formatPackageLabel(pkgName),
				Items: items,
			})
		}
		return groups
	}

	// Single package: one "API Reference" group
	var items []sidebarGroupItem
	items = append(items, sidebarGroupItem{Slug: "reference/api"})
	for _, pkgName := range pkgNames {
		pkg := result.Packages[pkgName]
		for _, name := range sortedServiceNames(pkg, cfg.ServiceOrder) {
			items = append(items, sidebarGroupItem{Slug: "reference/api/" + toKebab(name)})
		}
		for _, name := range sortedEnumNames(pkg) {
			items = append(items, sidebarGroupItem{Slug: "reference/api/enum-" + toKebab(name)})
		}
	}
	return []sidebarGroup{{Label: "API Reference", Items: items}}
}

// formatPackageLabel converts a proto package name to a sidebar label.
// e.g. "webhook.v1" -> "Webhook V1", "mypackage" -> "Mypackage"
func formatPackageLabel(pkg string) string {
	if pkg == "_default" {
		return "API Reference"
	}
	parts := strings.Split(pkg, ".")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, " ")
}
