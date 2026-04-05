package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/sarathsp06/proto2astro/internal/config"
	"github.com/sarathsp06/proto2astro/internal/parser"
)

// ServicePageData is the template context for generating MDX stubs.
type ServicePageData struct {
	ServiceName string
	Description string
	Slug        string // kebab-case filename without extension
	Package     string
}

// IndexPageData is the template context for generating the API index page.
type IndexPageData struct {
	Title       string
	Description string
	Services    []ServicePageData
	Enums       []EnumPageData
}

// EnumPageData is the template context for generating enum MDX stubs.
type EnumPageData struct {
	EnumName    string
	Description string
	Slug        string // e.g., "enum-webhook-status"
	Package     string
}

// yamlString escapes a string for safe use inside double-quoted YAML values.
// It escapes backslashes and double quotes.
func yamlString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}

// pageFuncMap provides template functions for page generation.
var pageFuncMap = template.FuncMap{
	"yaml": yamlString,
}

const serviceMDXTemplate = `---
title: "{{yaml .ServiceName}}"
description: "API reference for {{yaml .ServiceName}}"
---

import ApiServicePage from '../../../../components/ApiServicePage.astro';
import service from '../../../../data/api/{{.Slug}}';

<ApiServicePage {...service} />
`

const enumMDXTemplate = `---
title: "{{yaml .EnumName}} (Enum)"
description: "Enum reference for {{yaml .EnumName}}"
---

import EnumPage from '../../../../components/EnumPage.astro';
import enumData from '../../../../data/api/{{.Slug}}';

<EnumPage {...enumData} />
`

const indexMDTemplate = `---
title: "{{yaml .Title}}"
description: "{{yaml .Description}}"
---

Browse the full API reference below. Each service page includes request and
response schemas, field constraints, and auto-generated curl examples you can
copy and run directly. Enum types are documented separately with all values
and descriptions.

## Services

{{range .Services}}- [{{.ServiceName}}](./{{.Slug}}/) — {{.Description}}
{{end}}
{{- if .Enums}}
## Enums

{{range .Enums}}- [{{.EnumName}}](./{{.Slug}}/) — {{.Description}}
{{end}}
{{- end}}
`

const rootIndexMDXTemplate = `---
title: "{{yaml .Title}}"
description: "{{yaml .Description}}"
template: splash
hero:
  title: "{{yaml .Title}}"
  tagline: "{{yaml .Description}}"
  actions:
    - text: API Reference
      link: /reference/api/
      icon: right-arrow
    - text: Comment Guide
      link: /guides/comment-guide/
      variant: minimal
---

import { Card, CardGrid } from '@astrojs/starlight/components';

## Features

<CardGrid stagger>
  <Card title="Two-column layout" icon="document">
    Service and RPC documentation is presented in a clean two-column layout
    with request/response details alongside auto-generated curl examples.
  </Card>
  <Card title="Auto-generated curl" icon="rocket">
    Every RPC gets a ready-to-use curl command built from your proto field
    types and ` + "`@example`" + ` annotations — no manual work required.
  </Card>
  <Card title="Enum reference" icon="list-format">
    Enums get dedicated pages with every value documented, including
    descriptions and numeric values parsed from your proto comments.
  </Card>
  <Card title="Comment-driven" icon="pencil">
    Rich documentation is extracted directly from your proto comments using
    conventions like ` + "`Required`" + `, ` + "`@example`" + `, ` + "`Errors:`" + `, ` + "`Default:`" + `, and ` + "`Range:`" + `.
  </Card>
</CardGrid>

## Quick start

Generate your docs in three commands:

` + "```sh" + `
# Initialize the Astro Starlight site scaffold
proto2astro init

# Parse your protos and generate content
proto2astro generate

# Install dependencies and build
proto2astro install && proto2astro build
` + "```" + `

The output is a standard Astro Starlight site — customize components,
styles, and add new pages freely. See the
[Comment Guide](/guides/comment-guide/) for proto annotation conventions.
`

const commentGuideMD = `---
title: Proto Comment Guide
description: How to write proto comments for rich documentation
---

# Proto Comment Guide

This page documents the comment conventions supported by proto2astro. 
When you annotate your ` + "`.proto`" + ` files using these patterns, the documentation 
generator will automatically extract structured information and display it in the API reference.

## Leading Comments

Place comments directly above the message, field, RPC, or enum you want to document:

` + "```" + `protobuf
// RegisterWebhook creates a new webhook subscription.
// The webhook will start receiving events immediately after registration.
rpc RegisterWebhook(RegisterWebhookRequest) returns (RegisterWebhookResponse);
` + "```" + `

## Required Keyword

Include the word ` + "`Required`" + ` (capital R) anywhere in a field comment to mark it as required:

` + "```" + `protobuf
// Required. The URL to deliver webhook events to.
string url = 1;
` + "```" + `

## Deprecated Detection

Fields are marked as deprecated (and excluded from docs) in two ways:

1. The proto ` + "`deprecated`" + ` option:
` + "```" + `protobuf
string old_field = 5 [deprecated = true];
` + "```" + `

2. A comment starting with ` + "`Deprecated:`" + `:
` + "```" + `protobuf
// Deprecated: Use new_field instead.
string old_field = 5;
` + "```" + `

## Default Values

Use the ` + "`Default: VALUE`" + ` pattern to document default values:

` + "```" + `protobuf
// Maximum number of retry attempts. Default: 5.
int32 max_retries = 3;
` + "```" + `

## Range Constraints

Use the ` + "`Range: MIN-MAX`" + ` pattern to document valid ranges:

` + "```" + `protobuf
// Number of items per page. Range: 1-100. Default: 50.
int32 page_size = 2;
` + "```" + `

## Error Codes

Document RPC error codes using the ` + "`Errors: CODE description`" + ` pattern in RPC comments:

` + "```" + `protobuf
// RegisterWebhook creates a new webhook subscription.
// Errors: ALREADY_EXISTS if a webhook with the same URL already exists.
// Errors: INVALID_ARGUMENT if the URL is malformed.
rpc RegisterWebhook(RegisterWebhookRequest) returns (RegisterWebhookResponse);
` + "```" + `

## @example Annotation

Provide example values for fields using ` + "`@example`" + `. These are used to auto-generate curl commands and response JSON:

` + "```" + `protobuf
// The webhook endpoint URL. Required. @example "https://example.com/webhook"
string url = 1;

// Maximum retry attempts. @example 5
int32 max_retries = 2;

// Whether the webhook is active. @example true
bool active = 3;
` + "```" + `

JSON values are parsed automatically. If parsing fails, the value is treated as a string.
`

// generatePages generates MDX stubs, index page, and comment guide.
func generatePages(result *parser.ParseResult, cfg *config.Config, outDir string) error {
	docsDir := filepath.Join(outDir, "src", "content", "docs", "reference", "api")
	if err := os.MkdirAll(docsDir, 0o755); err != nil {
		return fmt.Errorf("mkdir docs dir: %w", err)
	}

	// Collect all services across packages for the index
	var allServices []ServicePageData
	var allEnums []EnumPageData

	for _, pkg := range result.Packages {
		svcNames := sortedServiceNames(pkg, cfg.ServiceOrder)
		for _, name := range svcNames {
			svc := pkg.Services[name]
			overlay := cfg.Services[name]
			desc := svc.Description
			if overlay.Description != "" {
				desc = overlay.Description
			}
			// Truncate description for index listing
			shortDesc := desc
			if len(shortDesc) > 120 {
				shortDesc = shortDesc[:117] + "..."
			}

			data := ServicePageData{
				ServiceName: name,
				Description: shortDesc,
				Slug:        toKebab(name),
				Package:     pkg.Name,
			}
			allServices = append(allServices, data)

			// Generate MDX stub
			if err := generateServiceMDX(data, docsDir); err != nil {
				return err
			}
		}

		// Generate enum pages
		enumNames := sortedEnumNames(pkg)
		for _, name := range enumNames {
			enum := pkg.Enums[name]
			shortDesc := enum.Description
			if len(shortDesc) > 120 {
				shortDesc = shortDesc[:117] + "..."
			}

			data := EnumPageData{
				EnumName:    name,
				Description: shortDesc,
				Slug:        "enum-" + toKebab(name),
				Package:     pkg.Name,
			}
			allEnums = append(allEnums, data)

			if err := generateEnumMDX(data, docsDir); err != nil {
				return err
			}
		}
	}

	// Generate index page
	if err := generateIndexPage(allServices, allEnums, cfg, docsDir); err != nil {
		return err
	}

	// Generate comment guide (scaffold-only: skip if file already exists)
	guideDir := filepath.Join(outDir, "src", "content", "docs", "guides")
	if err := os.MkdirAll(guideDir, 0o755); err != nil {
		return fmt.Errorf("mkdir guides dir: %w", err)
	}
	guidePath := filepath.Join(guideDir, "comment-guide.md")
	if _, err := os.Stat(guidePath); os.IsNotExist(err) {
		if err := os.WriteFile(guidePath, []byte(commentGuideMD), 0o644); err != nil {
			return fmt.Errorf("write comment guide: %w", err)
		}
	}

	// Generate root index page (scaffold-only: skip if file already exists)
	rootDocsDir := filepath.Join(outDir, "src", "content", "docs")
	if err := generateRootIndex(cfg, rootDocsDir); err != nil {
		return err
	}

	return nil
}

// generateServiceMDX writes a single service MDX stub.
func generateServiceMDX(data ServicePageData, docsDir string) error {
	tmpl, err := template.New("service").Funcs(pageFuncMap).Parse(serviceMDXTemplate)
	if err != nil {
		return fmt.Errorf("parse service template: %w", err)
	}

	path := filepath.Join(docsDir, data.Slug+".mdx")
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	return tmpl.Execute(f, data)
}

// generateEnumMDX writes a single enum MDX stub.
func generateEnumMDX(data EnumPageData, docsDir string) error {
	tmpl, err := template.New("enum").Funcs(pageFuncMap).Parse(enumMDXTemplate)
	if err != nil {
		return fmt.Errorf("parse enum template: %w", err)
	}

	path := filepath.Join(docsDir, data.Slug+".mdx")
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	return tmpl.Execute(f, data)
}

// generateIndexPage writes the API reference index page.
func generateIndexPage(services []ServicePageData, enums []EnumPageData, cfg *config.Config, docsDir string) error {
	tmpl, err := template.New("index").Funcs(pageFuncMap).Parse(indexMDTemplate)
	if err != nil {
		return fmt.Errorf("parse index template: %w", err)
	}

	data := IndexPageData{
		Title:       cfg.Title,
		Description: cfg.Description,
		Services:    services,
		Enums:       enums,
	}

	path := filepath.Join(docsDir, "index.md")
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create index: %w", err)
	}
	defer func() { _ = f.Close() }()

	return tmpl.Execute(f, data)
}

// generateRootIndex writes the site root landing page at src/content/docs/index.mdx.
// Scaffold-only: skips if the file already exists so user customizations are preserved.
func generateRootIndex(cfg *config.Config, rootDocsDir string) error {
	path := filepath.Join(rootDocsDir, "index.mdx")

	// Skip if already exists (scaffold-only behavior)
	if _, err := os.Stat(path); err == nil {
		return nil
	}

	tmpl, err := template.New("root-index").Funcs(pageFuncMap).Parse(rootIndexMDXTemplate)
	if err != nil {
		return fmt.Errorf("parse root index template: %w", err)
	}

	data := IndexPageData{
		Title:       cfg.Title,
		Description: cfg.Description,
	}

	// Remove old .md if it exists (from previous versions)
	oldPath := filepath.Join(rootDocsDir, "index.md")
	_ = os.Remove(oldPath)

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create root index: %w", err)
	}
	defer func() { _ = f.Close() }()

	return tmpl.Execute(f, data)
}

// sortedEnumNames returns enum names sorted alphabetically.
func sortedEnumNames(pkg *parser.ProtoPackage) []string {
	var names []string
	for name := range pkg.Enums {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
