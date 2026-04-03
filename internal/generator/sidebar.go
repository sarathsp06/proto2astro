package generator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/sarathsp06/proto2docs/internal/config"
	"github.com/sarathsp06/proto2docs/internal/parser"
)

// sidebarData is the template context for generating astro.config.mjs.
type sidebarData struct {
	Site         string
	Base         string
	Title        string
	Description  string
	Logo         string
	Social       []config.Link
	EditLink     string
	CustomPages  []config.CustomPage
	MultiPackage bool
	Packages     []sidebarPackage
}

// sidebarPackage represents a proto package in the sidebar.
type sidebarPackage struct {
	Label    string
	Services []sidebarItem
	Enums    []sidebarItem
}

// sidebarItem is a single sidebar entry.
type sidebarItem struct {
	Slug string
}

// jsString is a template function that JSON-encodes a string for use in JS source.
func jsString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

const astroConfigTemplate = `import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

export default defineConfig({
{{- if .Site}}
  site: {{js .Site}},
{{- end}}
{{- if .Base}}
  base: {{js .Base}},
{{- end}}
  integrations: [
    starlight({
      title: {{js .Title}},
      description: {{js .Description}},
{{- if .Logo}}
      logo: { src: {{js .Logo}} },
{{- end}}
{{- if .Social}}
      social: [
{{- range .Social}}
        { icon: {{js .Icon}}, label: {{js .Label}}, href: {{js .Href}} },
{{- end}}
      ],
{{- end}}
{{- if .EditLink}}
      editLink: { baseUrl: {{js .EditLink}} },
{{- end}}
      sidebar: [
        {
          label: 'Guides',
          items: [
            { slug: 'guides/comment-guide' },
{{- range .CustomPages}}
            { slug: 'guides/{{.Slug}}' },
{{- end}}
          ],
        },
{{- if .MultiPackage}}
{{- range .Packages}}
        {
          label: {{js .Label}},
          items: [
            { slug: 'reference/api' },
{{- range .Services}}
            { slug: 'reference/api/{{.Slug}}' },
{{- end}}
{{- range .Enums}}
            { slug: 'reference/api/{{.Slug}}' },
{{- end}}
          ],
        },
{{- end}}
{{- else}}
        {
          label: 'API Reference',
          items: [
            { slug: 'reference/api' },
{{- range .Packages}}
{{- range .Services}}
            { slug: 'reference/api/{{.Slug}}' },
{{- end}}
{{- range .Enums}}
            { slug: 'reference/api/{{.Slug}}' },
{{- end}}
{{- end}}
          ],
        },
{{- end}}
      ],
      customCss: ['./src/styles/custom.css'],
    }),
  ],
});
`

// generateAstroConfig generates the astro.config.mjs with a dynamic sidebar
// based on parsed proto packages and services.
func generateAstroConfig(result *parser.ParseResult, cfg *config.Config, outDir string) error {
	tmpl, err := template.New("astro-config").Funcs(template.FuncMap{
		"js": jsString,
	}).Parse(astroConfigTemplate)
	if err != nil {
		return fmt.Errorf("parse astro config template: %w", err)
	}

	data := buildSidebarData(result, cfg)

	path := filepath.Join(outDir, "astro.config.mjs")
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	return tmpl.Execute(f, data)
}

// buildSidebarData constructs the template context for astro.config.mjs.
func buildSidebarData(result *parser.ParseResult, cfg *config.Config) sidebarData {
	var pkgNames []string
	for name := range result.Packages {
		pkgNames = append(pkgNames, name)
	}
	sort.Strings(pkgNames)

	var packages []sidebarPackage
	for _, pkgName := range pkgNames {
		pkg := result.Packages[pkgName]

		var services []sidebarItem
		for _, name := range sortedServiceNames(pkg, cfg.ServiceOrder) {
			services = append(services, sidebarItem{Slug: toKebab(name)})
		}

		var enums []sidebarItem
		for _, name := range sortedEnumNames(pkg) {
			enums = append(enums, sidebarItem{Slug: "enum-" + toKebab(name)})
		}

		packages = append(packages, sidebarPackage{
			Label:    formatPackageLabel(pkgName),
			Services: services,
			Enums:    enums,
		})
	}

	return sidebarData{
		Site:         cfg.Site,
		Base:         cfg.Base,
		Title:        cfg.Title,
		Description:  cfg.Description,
		Logo:         cfg.Logo,
		Social:       cfg.Social,
		EditLink:     cfg.EditLink,
		CustomPages:  cfg.CustomPages,
		MultiPackage: len(pkgNames) > 1,
		Packages:     packages,
	}
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
