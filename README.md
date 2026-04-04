<p align="center">
  <strong>proto2astro</strong>
</p>

<p align="center">
  <code>.proto</code> files in, beautiful API docs out.
</p>

<p align="center">
  <a href="https://github.com/sarathsp06/proto2astro/actions/workflows/ci.yml"><img src="https://github.com/sarathsp06/proto2astro/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://github.com/sarathsp06/proto2astro/releases"><img src="https://img.shields.io/github/v/release/sarathsp06/proto2astro" alt="Release"></a>
  <a href="https://pkg.go.dev/github.com/sarathsp06/proto2astro"><img src="https://pkg.go.dev/badge/github.com/sarathsp06/proto2astro.svg" alt="Go Reference"></a>
  <a href="https://goreportcard.com/report/github.com/sarathsp06/proto2astro"><img src="https://goreportcard.com/badge/github.com/sarathsp06/proto2astro" alt="Go Report Card"></a>
  <a href="LICENSE"><img src="https://img.shields.io/github/license/sarathsp06/proto2astro" alt="License"></a>
</p>

<p align="center">
  <a href="#quick-start">Quick Start</a> &middot;
  <a href="https://sarathsp06.github.io/sparrow/reference/api/">Live Demo</a> &middot;
  <a href="#configuration">Configuration</a> &middot;
  <a href="#deployment">Deployment</a>
</p>

---

proto2astro generates a complete [Astro Starlight](https://starlight.astro.build/) documentation site from your [Protocol Buffer](https://protobuf.dev/) definitions. Two-column layout, auto-generated curl examples, organized sidebar — ready for [ConnectRPC](https://connectrpc.com/) services. No `protoc` needed.

```
  ┌─────────────┐         ┌──────────────┐         ┌─────────────────────┐
  │             │         │              │         │                     │
  │  .proto     │────────▶│  proto2astro  │────────▶│  Static docs site   │
  │  files      │         │              │         │  (Astro Starlight)  │
  │             │         │              │         │                     │
  └─────────────┘         └──────────────┘         └─────────────────────┘
                                                     ▲
                                                     │
                                                   Deploy to
                                               GitHub Pages,
                                              Netlify, Vercel,
                                              or any static host
```

**See it in action:** The [Sparrow API Reference](https://sarathsp06.github.io/sparrow/reference/api/) is generated entirely by proto2astro.

## Highlights

- **Zero `protoc`** — parses `.proto` source files directly, no compilation step
- **Scalar-inspired two-column layout** — fields on the left, curl examples on the right
- **ConnectRPC-native** — curl examples use HTTP POST + JSON with correct service paths
- **Just works with Buf** — discovers protos from `buf.yaml` workspaces automatically
- **Rich proto comments** — extracts `Required`, `Deprecated:`, `Default:`, `Range:`, `Errors:`, `@example`
- **Full site lifecycle** — `init` / `generate` / `install` / `build` / `dev` / `preview`
- **Your site, your rules** — components, styles, and pages are fully customizable

## Quick Start

Install and generate a documentation site in under a minute:

```sh
# Install
go install github.com/sarathsp06/proto2astro/cmd/proto2astro@latest

# Create a config file
cat > proto2astro.yaml <<EOF
title: "My API"
proto:
  paths:
    - ./proto
EOF

# Generate, install deps, build, and preview
proto2astro generate
proto2astro install
proto2astro build
proto2astro preview     # → http://localhost:4321
```

Your API docs are now at `docs/dist/`. Deploy them anywhere.

> Pre-built binaries for macOS and Linux (arm64/amd64) are available on the [releases page](https://github.com/sarathsp06/proto2astro/releases).

---

## How It Works

```
proto2astro generate
```

That single command:

1. **Parses** all `.proto` files from the configured paths (source-level, using [`emicklei/proto`](https://github.com/emicklei/proto))
2. **Resolves** cross-package types, nested messages, oneofs, and self-referencing structures
3. **Generates** TypeScript data files + MDX pages + `astro.config.mjs` with a dynamic sidebar
4. **Scaffolds** the Astro Starlight project if it doesn't exist yet

The result is a standard Astro project. Run `proto2astro build` and you get static HTML.

### What proto2astro understands

| Proto feature | How it's documented |
|---|---|
| Services & RPCs | Dedicated page per service, with all RPCs |
| Request / Response messages | Field tables with types, descriptions, constraints |
| Enums | Dedicated pages with all values |
| Nested messages | Qualified names (`Outer.Inner`) resolved and linked |
| Oneof groups | Documented with group metadata |
| Streaming RPCs | Marked as client/server/bidi stream |
| Cross-package references | Resolved and linked across proto packages |
| Recursive types | Handled without infinite loops |

---

## CLI

All commands accept `-o` / `--out` to set the site directory (default: `./docs`).

```sh
proto2astro init [dir]       # Scaffold the Astro project (no API content)
proto2astro generate         # Parse protos → generate docs
proto2astro install          # npm install
proto2astro build            # Build static HTML → <site>/dist/
proto2astro dev              # Dev server with hot-reload (localhost:4321)
proto2astro preview          # Serve the built site locally
```

### Generate options

```sh
proto2astro generate                              # uses proto2astro.yaml
proto2astro generate -c custom-config.yaml        # custom config
proto2astro generate -p ./proto -o ./docs         # override paths
proto2astro generate --buf-workspace ./            # discover from buf workspace
```

### Day-to-day workflow

```sh
# First time
proto2astro generate && proto2astro install && proto2astro build

# After proto changes
proto2astro generate && proto2astro build

# Tweaking styles/components
proto2astro dev
```

---

## Configuration

All config lives in a single `proto2astro.yaml`.

<details>
<summary><strong>Minimal config</strong></summary>

```yaml
title: "My API"
proto:
  paths:
    - ./proto
```

</details>

<details open>
<summary><strong>Full config</strong></summary>

```yaml
# ── Project ──────────────────────────────────────
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
  # Option A: file/directory paths
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
entity_types:                            # don't flatten these into parent fields
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

</details>

### All fields

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

Write comments in your `.proto` files and proto2astro picks them up automatically. All conventions are optional — plain `//` comments work as descriptions.

```protobuf
// Create a new user account.
// The email must be unique across the system.
// Errors: ALREADY_EXISTS if the email is taken.
rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);

message CreateUserRequest {
  // Required. The user's email address.
  // @example "alice@example.com"
  string email = 1;

  // Display name. Default: "Anonymous".
  string display_name = 2;

  // Number of invites to pre-allocate. Range: 0-100.
  int32 invite_count = 3;

  // Deprecated: Use display_name instead.
  string name = 4;
}
```

| Annotation | What it does |
|---|---|
| `Required.` | Marks the field as required |
| `Deprecated:` | Shows deprecation notice with reason |
| `Default:` | Displays the default value |
| `Range:` | Shows allowed range constraint |
| `Errors:` | Lists error codes on the RPC (can appear multiple times) |
| `@example` | Uses the value in generated curl examples (JSON-parsed when possible) |

---

## Customizing the Site

The generated site is a standard Astro Starlight project. proto2astro respects your changes.

### Safe to edit (written once, never overwritten)

```
src/components/*.astro     ← layout, endpoint rendering, field tables
src/styles/custom.css      ← all site styles
src/content.config.ts      ← Astro content config
src/data/api/types.ts      ← TypeScript type definitions
package.json               ← add npm dependencies here
tsconfig.json
```

### Regenerated on every `generate` run

```
astro.config.mjs           ← sidebar and site config
src/data/api/*.ts          ← data files from protos
src/content/docs/reference/api/*.mdx  ← service and enum pages
src/content/docs/index.md  ← landing page
```

> Use `proto2astro init --force` to reset scaffold files to defaults after upgrading.

### Components you can customize

| Component | Controls |
|---|---|
| `ApiServicePage.astro` | Full service page (two-column layout) |
| `ApiEndpoint.astro` | Individual RPC endpoint |
| `ApiRequest.astro` | Request fields table |
| `ApiResponse.astro` | Response fields table |
| `ErrorCodes.astro` | Error code display |
| `EnumPage.astro` | Enum values page |

---

## Building a Full Documentation Site

proto2astro generates the **API Reference** section. For a complete docs site like the [Sparrow documentation](https://sarathsp06.github.io/sparrow/), add your own sections around it.

### Recommended structure

```
docs/src/content/docs/
├── getting-started/
│   ├── installation.md          ✍ you write these
│   ├── quickstart.md
│   ├── how-it-works.md
│   └── configuration.md
├── reference/
│   ├── api/                     ← generated by proto2astro
│   │   ├── index.md
│   │   ├── payment-service.mdx
│   │   └── ...
│   ├── client-libraries.md      ✍ you write these
│   ├── error-handling.md
│   └── architecture.md
└── deployment/
    ├── docker-compose.md        ✍ you write these
    └── kubernetes.md
```

### Sidebar patching

Since `astro.config.mjs` is regenerated, use a wrapper script to add your custom sidebar sections:

```sh
#!/bin/sh
# generate-docs.sh — generate + patch + build
proto2astro generate

node -e "
const fs = require('fs');
let c = fs.readFileSync('docs/astro.config.mjs', 'utf8');
c = c.replace('sidebar: [', \`sidebar: [
    {
      label: 'Getting Started',
      items: [
        { label: 'Installation', slug: 'getting-started/installation' },
        { label: 'Quickstart', slug: 'getting-started/quickstart' },
        { label: 'How It Works', slug: 'getting-started/how-it-works' },
        { label: 'Configuration', slug: 'getting-started/configuration' },
      ],
    },\`);
fs.writeFileSync('docs/astro.config.mjs', c);
"

proto2astro install && proto2astro build
```

### Writing tips

| Page | Approach |
|---|---|
| **Installation** | Lead with the easiest method (Docker/binary). Include prerequisites. |
| **Quickstart** | Step-by-step curl walkthrough. End with "What just happened?" and "Next steps." |
| **How It Works** | High-level architecture. Diagrams help. |
| **Configuration** | Tables grouped by category. Cover every env var and config option. |
| **Client Libraries** | Code examples in Go, JavaScript, Python. Use MDX tabs for multi-language. |
| **Error Handling** | Document all error codes, response format, common errors. |
| **Deployment** | Copy-paste-ready `docker-compose.yml` and Kubernetes manifests. |

### MDX tabs for multi-language examples

```mdx
import { Tabs, TabItem } from '@astrojs/starlight/components';

<Tabs>
  <TabItem label="curl">
    ```sh
    curl -X POST http://localhost:8080/my.Service/MyRpc \
      -H "Content-Type: application/json" \
      -d '{"name": "alice"}'
    ```
  </TabItem>
  <TabItem label="Go">
    ```go
    res, err := client.MyRpc(ctx, connect.NewRequest(&pb.MyRequest{Name: "alice"}))
    ```
  </TabItem>
  <TabItem label="TypeScript">
    ```ts
    const res = await client.myRpc({ name: "alice" });
    ```
  </TabItem>
</Tabs>
```

---

## Deployment

The built site is static HTML. Deploy it anywhere.

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
# Upload docs/dist/
```

---

## Development

### Requirements

- **Go 1.21+**
- **Node.js 18+**
- **macOS or Linux**

### Build from source

```sh
git clone https://github.com/sarathsp06/proto2astro.git
cd proto2astro
make build       # → bin/proto2astro
make install     # → $GOPATH/bin/proto2astro
```

### Commands

```sh
make build       # Build the binary
make test        # Run tests with -race
make vet         # go vet
make lint        # golangci-lint
make fmt         # gofmt + goimports
make clean       # Remove bin/ and dist/
```

### Releasing

Tag and push — CI handles the rest via [GoReleaser](https://goreleaser.com/):

```sh
git tag v0.1.0
git push origin v0.1.0
```

Produces binaries for macOS (arm64, amd64) and Linux (arm64, amd64).

---

## License

[MIT](LICENSE)
