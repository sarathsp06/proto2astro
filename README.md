# Protostar

Generate beautiful API documentation from `.proto` files. Built for [ConnectRPC](https://connectrpc.com/).

Protostar parses your Protocol Buffer definitions and generates a complete [Astro Starlight](https://starlight.astro.build/) documentation site with a two-column layout, auto-generated curl examples, and organized sidebar navigation. No `protoc` required.

**[Live example](https://sarathsp06.github.io/sparrow/reference/api/)** — generated from Sparrow's proto files.

---

## Features

| | |
|---|---|
| **Source-level parsing** | Parses `.proto` files directly using `emicklei/proto`. No `protoc` or code generation needed. |
| **Two-column layout** | Request/response fields on the left, curl examples on the right. Inspired by Scalar. |
| **ConnectRPC examples** | Curl examples use `POST` with `Content-Type: application/json` and correct service paths. |
| **Package organization** | Multi-package protos get automatic sidebar grouping. |
| **Rich comment extraction** | `Required`, `Deprecated:`, `Default:`, `Range:`, `Errors:`, `@example` annotations. |
| **Enum documentation** | Dedicated pages for enums with all values documented. |
| **Oneof / nested messages** | Oneof groups and qualified type names (`Outer.Inner`) are resolved and documented. |
| **Streaming RPCs** | Detected and noted in the generated output. |
| **Cross-package resolution** | Types referenced across proto packages are resolved correctly. |
| **Cycle-safe** | Recursive and self-referencing message types are handled without infinite loops. |
| **Buf workspace support** | Discover proto files from `buf.yaml` workspaces. |
| **YAML overlay** | Customize descriptions, examples, ordering, and add custom pages without touching protos. |
| **Fully customizable** | Edit Astro components, styles, and add your own pages directly in the generated site. |

---

## Quick Start

### 1. Install

```sh
go install github.com/sarathsp06/proto2astro/cmd/proto2astro@latest
```

Or download a pre-built binary from the [releases page](https://github.com/sarathsp06/proto2astro/releases) (macOS and Linux, arm64/amd64).

### 2. Configure

Create a `proto2astro.yaml` in your project root:

```yaml
title: "My API"
description: "API reference for My Service"

proto:
  paths:
    - ./proto

out_dir: ./docs
```

### 3. Generate and build

```sh
proto2astro generate    # parse protos, generate the site
proto2astro install     # npm install (first time only)
proto2astro build       # build static HTML to docs/dist/
```

### 4. Preview

```sh
proto2astro preview     # serve locally at http://localhost:4321
```

The static site is output to `docs/dist/` — ready to deploy anywhere.

---

## CLI Reference

All commands accept `-o` / `--out` to specify the site directory (default: `./docs`).

| Command | Description |
|---------|-------------|
| `proto2astro init [dir]` | Scaffold a new Astro Starlight project without generating API content. |
| `proto2astro generate` | Parse proto files and generate the complete documentation site. |
| `proto2astro install` | Run `npm install` in the generated site directory. |
| `proto2astro build` | Build the site into static HTML (`<site-dir>/dist/`). |
| `proto2astro dev` | Start the Astro dev server with hot-reload at `localhost:4321`. |
| `proto2astro preview` | Serve the built static site locally for review. |

### Generate flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--config` | `-c` | `proto2astro.yaml` | Config file path |
| `--proto` | `-p` | *(from config)* | Proto file or directory (overrides config) |
| `--out` | `-o` | *(from config)* | Output directory (overrides config) |
| `--buf-workspace` | | *(from config)* | Buf workspace root (overrides config) |

### Typical workflow

```sh
# First time
proto2astro generate && proto2astro install && proto2astro build

# After changing protos
proto2astro generate && proto2astro build

# Iterate on styling/components
proto2astro dev
```

---

## Configuration

All configuration lives in `proto2astro.yaml`.

### Full example

```yaml
# ── Project metadata ─────────────────────────────
title: "Payment API"
description: "API reference for the Payment service"
site: "https://example.github.io"       # base URL (enables sitemap)
base: "/payment-api"                     # URL base path
logo: "./src/assets/logo.svg"
edit_link: "https://github.com/example/payment/edit/main/proto"

social:
  - icon: github
    label: GitHub
    href: https://github.com/example/payment

# ── Proto input ──────────────────────────────────
proto:
  # Option A: direct paths
  paths:
    - ./proto/payment
    - ./proto/common/types.proto

  # Option B: Buf workspace
  # buf_workspace: ./
  # buf_modules:
  #   - proto/payment
  #   - proto/common

# ── Output ───────────────────────────────────────
out_dir: ./docs

# ── Sidebar ordering ─────────────────────────────
service_order:
  - PaymentService
  - RefundService

# ── Type handling ────────────────────────────────
entity_types:
  - PaymentIntent
  - Customer

# ── Per-service overrides ────────────────────────
services:
  PaymentService:
    description: "Handles payment processing."
    rpcs:
      CreatePayment:
        description: "Create a new payment intent."
        fields:
          amount:
            example: 2500
            description: "Amount in cents."
            required: true
          currency:
            example: "usd"

# ── Custom pages ─────────────────────────────────
custom_pages:
  - title: "Webhooks"
    slug: webhooks
    content: |
      # Webhook Events
      Payment events are sent to your configured endpoint...
```

### Field reference

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `title` | string | `"API Documentation"` | Site title |
| `description` | string | `"API reference documentation"` | Site description |
| `site` | string | | Base URL for deployed site (enables sitemap) |
| `base` | string | | URL base path (e.g., `/my-api`) |
| `logo` | string | | Logo image path (relative to site root) |
| `social` | list | | Header links (`icon`, `label`, `href`) |
| `edit_link` | string | | Base URL for "edit this page" links |
| `proto.paths` | list | | Proto files or directories to parse |
| `proto.buf_workspace` | string | | Buf workspace root directory |
| `proto.buf_modules` | list | | Specific Buf modules to include |
| `out_dir` | string | `"./docs"` | Output directory |
| `service_order` | list | | Explicit sidebar ordering for services |
| `entity_types` | list | | Message types that should not be flattened |
| `services` | map | | Per-service overrides (descriptions, examples, fields) |
| `custom_pages` | list | | Additional pages (added under `/guides/`) |

---

## Proto Comment Conventions

Protostar extracts structured information from your proto comments. All conventions are optional.

```protobuf
// Create a new user account.
// The email must be unique across the system.
rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);
```

### Supported annotations

| Pattern | Example | Effect |
|---------|---------|--------|
| `Required.` | `// Required. The user's email.` | Field marked as required |
| `Deprecated:` | `// Deprecated: Use display_name instead.` | Field shown as deprecated |
| `Default:` | `// Default: 20.` | Default value displayed |
| `Range:` | `// Range: 1-100.` | Constraint displayed |
| `Errors:` | `// Errors: NOT_FOUND if missing.` | Error codes listed on the RPC |
| `@example` | `// @example "alice@example.com"` | Value used in curl examples |

The `@example` value is JSON-parsed when possible (objects, arrays, numbers, booleans), otherwise treated as a string.

---

## Customizing the Generated Site

The generated site is a standard Astro Starlight project. Protostar is designed to coexist with your manual changes.

### What gets overwritten on each `generate` run

These files are regenerated every time. Do not edit them directly:

- `astro.config.mjs` — sidebar and site configuration
- `src/data/api/*.ts` — TypeScript data files from protos
- `src/content/docs/reference/api/*.mdx` — service and enum pages
- `src/content/docs/index.md` — landing page
- `src/content/docs/guides/comment-guide.md` — comment conventions guide
- Custom pages from `custom_pages` config

### What is safe to edit

These are written once during initial scaffold and never touched again:

- `src/components/*.astro` — Astro components (layout, endpoint rendering)
- `src/styles/custom.css` — site styles
- `src/content.config.ts` — Astro content config
- `src/data/api/types.ts` — TypeScript type definitions
- `package.json`, `tsconfig.json`
- Any new files you add

Use `proto2astro init --force` to reset scaffold files to defaults (e.g., after upgrading).

### Components

| Component | Purpose |
|-----------|---------|
| `ApiServicePage.astro` | Full service page layout (two-column, endpoint list) |
| `ApiEndpoint.astro` | Single RPC endpoint section |
| `ApiRequest.astro` | Request fields table |
| `ApiResponse.astro` | Response fields table |
| `ErrorCodes.astro` | Error codes display |
| `EnumPage.astro` | Enum values documentation |

---

## Building a Complete Documentation Site

Protostar generates the **API Reference** section out of the box. To build a documentation site on par with the [Sparrow docs](https://sarathsp06.github.io/sparrow/), you'll want to add additional sections manually. Here's how.

### Recommended site structure

```
Getting Started
  ├── Installation
  ├── Quickstart
  ├── How It Works
  └── Configuration

API Reference              ← generated by protostar
  ├── Overview
  ├── ServiceA
  ├── ServiceB
  └── ...

Reference
  ├── Client Libraries
  ├── Error Handling
  └── Architecture

Deployment
  ├── Docker Compose
  └── Kubernetes
```

### Adding sections

Protostar generates pages under `src/content/docs/reference/api/`. To add other sections:

**1. Create the markdown files:**

```
docs/src/content/docs/
├── getting-started/
│   ├── installation.md
│   ├── quickstart.md
│   ├── how-it-works.md
│   └── configuration.md
├── reference/
│   ├── api/                 ← generated by protostar
│   ├── client-libraries.md
│   ├── error-handling.md
│   └── architecture.md
└── deployment/
    ├── docker-compose.md
    └── kubernetes.md
```

**2. Update the sidebar in `proto2astro.yaml`:**

Since `astro.config.mjs` is regenerated each run, you cannot edit it directly. Instead, use `custom_pages` for simple pages, or — for full sidebar control — **override `astro.config.mjs` after generation**.

A practical approach: create a post-generation script that patches the sidebar:

```sh
#!/bin/sh
# generate-docs.sh
proto2astro generate

# Patch the sidebar to include custom sections
node -e "
const fs = require('fs');
let config = fs.readFileSync('docs/astro.config.mjs', 'utf8');

const customSidebar = \`
  sidebar: [
    {
      label: 'Getting Started',
      items: [
        { label: 'Installation', slug: 'getting-started/installation' },
        { label: 'Quickstart', slug: 'getting-started/quickstart' },
        { label: 'How It Works', slug: 'getting-started/how-it-works' },
        { label: 'Configuration', slug: 'getting-started/configuration' },
      ],
    },
\`;

// Insert custom sections before the generated API Reference section
config = config.replace('sidebar: [', customSidebar);
fs.writeFileSync('docs/astro.config.mjs', config);
"

proto2astro install
proto2astro build
```

**3. Write content following Starlight conventions:**

Each markdown file needs frontmatter:

```markdown
---
title: Installation
description: How to install and set up the service.
---

## Docker Compose (Recommended)

The simplest way to get started...

## Build from Source

### Prerequisites

- Go 1.21+
- Node.js 18+
- PostgreSQL 15+

### Build

\`\`\`sh
make build
\`\`\`
```

### Page writing tips

Drawing from the Sparrow docs as a reference:

- **Installation page** — lead with the easiest method (Docker, pre-built binary), then build-from-source. Include prerequisites.
- **Quickstart page** — a step-by-step walkthrough using curl. Show a complete flow: create something, use it, verify the result. End with "What just happened?" and "Next steps."
- **How It Works** — explain the system architecture. Diagrams help. Keep it high-level.
- **Configuration** — list all environment variables or config options in a table. Group by category.
- **Client Libraries** — show how to call the API from Go, JavaScript, Python, etc.
- **Error Handling** — document error codes, response format, and common errors.
- **Deployment pages** — provide copy-paste-ready configs (docker-compose.yml, Kubernetes manifests). Include an observability section if applicable.

### Using MDX for richer pages

Starlight supports MDX, which lets you use Astro components in your documentation. For example, to add tabbed content (e.g., showing both curl and SDK examples):

```mdx
---
title: Quickstart
---

import { Tabs, TabItem } from '@astrojs/starlight/components';

<Tabs>
  <TabItem label="curl">
    ```sh
    curl -X POST http://localhost:8080/my.Service/MyRpc ...
    ```
  </TabItem>
  <TabItem label="Go">
    ```go
    client.MyRpc(ctx, connect.NewRequest(&pb.MyRequest{}))
    ```
  </TabItem>
</Tabs>
```

---

## Deployment

The built site in `<out_dir>/dist/` is static HTML. Deploy it anywhere.

### GitHub Pages

```yaml
# .github/workflows/docs.yml
name: Deploy API Docs
on:
  push:
    branches: [main]

permissions:
  pages: write
  id-token: write

jobs:
  deploy:
    runs-on: ubuntu-latest
    environment:
      name: github-pages
      url: ${{ steps.deployment.outputs.page_url }}
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'
      - uses: actions/setup-node@v4
        with:
          node-version: '18'
      - run: go install github.com/sarathsp06/proto2astro/cmd/proto2astro@latest
      - run: proto2astro generate
      - run: proto2astro install
      - run: proto2astro build
      - uses: actions/upload-pages-artifact@v3
        with:
          path: docs/dist
      - id: deployment
        uses: actions/deploy-pages@v4
```

### Netlify / Vercel / Any static host

```sh
proto2astro generate && proto2astro install && proto2astro build
# Deploy the contents of docs/dist/
```

---

## Development

```sh
make build       # build binary to bin/proto2astro
make test        # run tests with -race
make vet         # go vet
make lint        # golangci-lint (install separately)
make fmt         # gofmt + goimports
make clean       # remove bin/ and dist/
```

### Requirements

- **Go 1.21+** (to build proto2astro)
- **Node.js 18+** (to build the generated Astro site)
- **macOS or Linux**

### Building from source

```sh
git clone https://github.com/sarathsp06/proto2astro.git
cd proto2astro
make build       # binary: bin/proto2astro
make install     # installs to $GOPATH/bin
```

### Releasing

Releases use [GoReleaser](https://goreleaser.com/). Tag and push:

```sh
git tag v0.1.0
git push origin v0.1.0
make release
```

Produces binaries for macOS (arm64, amd64) and Linux (arm64, amd64).

---

## License

MIT
