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
    annotations like ` + "`@required`" + `, ` + "`@example`" + `, ` + "`@error`" + `, ` + "`@default`" + `, and ` + "`@range`" + `.
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

## Annotations

All annotations use the ` + "`@`" + ` prefix for consistency, following conventions from Javadoc, JSDoc, and similar systems.

### @required

Mark a field as required:

` + "```" + `protobuf
// @required The URL to deliver webhook events to.
// @example "https://example.com/webhook"
string url = 1;
` + "```" + `

### @deprecated

Mark a field or message as deprecated. Deprecated fields are excluded from generated documentation:

` + "```" + `protobuf
// @deprecated Use new_field instead.
string old_field = 5;
` + "```" + `

Fields with the proto ` + "`deprecated`" + ` option are also detected:
` + "```" + `protobuf
string old_field = 5 [deprecated = true];
` + "```" + `

### @default

Document a field's default value:

` + "```" + `protobuf
// Maximum number of retry attempts. @default 5
int32 max_retries = 3;
` + "```" + `

### @range

Document valid value ranges:

` + "```" + `protobuf
// Number of items per page. @range 1-100 @default 50
int32 page_size = 2;
` + "```" + `

### @error

Document RPC error codes in RPC comments:

` + "```" + `protobuf
// RegisterWebhook creates a new webhook subscription.
// @error ALREADY_EXISTS if a webhook with the same URL already exists.
// @error INVALID_ARGUMENT if the URL is malformed.
rpc RegisterWebhook(RegisterWebhookRequest) returns (RegisterWebhookResponse);
` + "```" + `

### @example

Provide example values for fields. These are used to auto-generate curl commands and response JSON:

` + "```" + `protobuf
// @required The webhook endpoint URL.
// @example "https://example.com/webhook"
string url = 1;

// Maximum retry attempts. @example 5
int32 max_retries = 2;

// Whether the webhook is active. @example true
bool active = 3;
` + "```" + `

JSON values are parsed automatically. If parsing fails, the value is treated as a string.

#### Multi-line examples

For complex JSON values, use a fenced block with triple backticks:

` + "```" + `protobuf
// JSON metadata for the item.
// @example ` + "```" + `
// {"key": "value", "count": 1}
// ` + "```" + `
string metadata = 4;
` + "```" + `

The lines between the fences are joined into a single value and parsed as JSON.

## Complete Example

` + "```" + `protobuf
// Create a new user account.
// @error ALREADY_EXISTS if the email is taken.
rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);

message CreateUserRequest {
  // @required The user's email address.
  // @example "alice@example.com"
  string email = 1;

  // Display name. @default "Anonymous"
  string display_name = 2;

  // Number of invites to pre-allocate. @range 0-100
  int32 invite_count = 3;

  // @deprecated Use display_name instead.
  string name = 4;
}
` + "```" + `

## Legacy Syntax

The following legacy patterns are still supported for backward compatibility:

| Legacy | Preferred |
|--------|-----------|
| ` + "`Required.`" + ` / ` + "`Required `" + ` | ` + "`@required`" + ` |
| ` + "`Deprecated: reason`" + ` | ` + "`@deprecated reason`" + ` |
| ` + "`Default: VALUE.`" + ` | ` + "`@default VALUE`" + ` |
| ` + "`Range: MIN-MAX`" + ` | ` + "`@range MIN-MAX`" + ` |
| ` + "`Errors: CODE desc`" + ` | ` + "`@error CODE desc`" + ` |
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

	// Generate comment guide (scaffold-only: skip if file already exists or disabled via config)
	if cfg.Scaffold.CommentGuideEnabled() {
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
	}

	// Generate root index page (scaffold-only: skip if file already exists or disabled via config)
	if cfg.Scaffold.LandingPageEnabled() {
		rootDocsDir := filepath.Join(outDir, "src", "content", "docs")
		// Route collision detection: warn if src/pages/index.* exists
		if collision := detectRouteCollision(outDir, "index"); collision != "" {
			fmt.Printf("  warning: scaffold index.mdx skipped — route collision with %s\n", collision)
		} else if err := generateRootIndex(cfg, rootDocsDir); err != nil {
			return err
		}
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

// detectRouteCollision checks if a src/pages/{name}.* file exists that would
// conflict with a scaffold-only content page. Returns the relative path of the
// conflicting file, or empty string if no collision.
func detectRouteCollision(outDir string, name string) string {
	pagesDir := filepath.Join(outDir, "src", "pages")
	for _, ext := range []string{".astro", ".tsx", ".jsx", ".md", ".mdx"} {
		candidate := filepath.Join(pagesDir, name+ext)
		if _, err := os.Stat(candidate); err == nil {
			// Return relative path for a readable warning
			rel, _ := filepath.Rel(outDir, candidate)
			if rel == "" {
				rel = candidate
			}
			return rel
		}
	}
	return ""
}
