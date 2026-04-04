# proto2astro

Generate a complete [Astro Starlight](https://starlight.astro.build/) API documentation site from `.proto` files — designed for [ConnectRPC](https://connectrpc.com/) services.

The generated site features a Scalar-inspired two-column layout with auto-generated `curl` examples, organized by proto package. Run `proto2astro generate` and you get a fully buildable static site, ready for deployment.

## Features

- **Source-level parsing** — uses `emicklei/proto` to parse `.proto` files directly (no `protoc` required)
- **Two-column layout** — request/response fields on the left, `curl` examples on the right
- **ConnectRPC HTTP POST + JSON** — curl examples use `Content-Type: application/json` with correct service paths
- **Organized by package** — multi-package protos get automatic sidebar grouping
- **Enum documentation** — dedicated pages for enums with all values documented
- **Rich comment extraction** — `Required`, `Deprecated:`, `Default:`, `Range:`, `Errors:`, `@example` annotations
- **YAML overlay** — customize descriptions, examples, ordering, and add custom pages without touching protos
- **Buf workspace support** — discover proto files from `buf.yaml` workspaces
- **Customizable** — edit components, styles, and add your own pages directly in the generated site
- **Full site lifecycle** — `init`, `generate`, `install`, `build`, `dev`, `preview` commands

## Requirements

- **Go 1.21+** (to build proto2astro)
- **Node.js 18+** (to build the generated Astro site)
- **macOS or Linux**

## Installation

```sh
go install github.com/sarathsp06/proto2astro/cmd/proto2astro@latest
```

Or build from source:

```sh
git clone https://github.com/sarathsp06/proto2astro.git
cd proto2astro
go build -o proto2astro ./cmd/proto2astro
```

## Quick Start

```sh
# 1. Create a config file
cat > proto2astro.yaml <<EOF
title: "My API"
description: "API reference for My Service"

proto:
  paths:
    - ./proto

out_dir: ./docs
EOF

# 2. Generate the documentation site
proto2astro generate

# 3. Install dependencies and build
proto2astro install
proto2astro build

# 4. Preview the result
proto2astro preview
```

The static site is output to `./docs/dist/` — ready to deploy to any static host.

---

## CLI Commands

Every command that operates on the generated site accepts `-o` / `--out` to specify the site directory. The default is `./docs`.

### `proto2astro init [output-dir]`

Scaffold a new Astro Starlight project without generating any API content. This creates the base project: `package.json`, Astro components, styles, TypeScript types.

```sh
proto2astro init                     # scaffold into ./docs
proto2astro init ./my-docs           # scaffold into ./my-docs
proto2astro init --force ./docs      # overwrite existing scaffold files
```

Scaffold files (components, styles, `package.json`, etc.) are only written if they don't already exist. Use `--force` to overwrite them — useful when upgrading proto2astro to pick up new component changes.

**You don't need to run `init` separately.** The `generate` command runs the scaffold automatically if the site directory doesn't exist yet.

### `proto2astro generate`

The main command. Parses `.proto` files and generates the complete documentation site. This does the following in order:

1. Scaffolds the site if it doesn't exist (same as `init`, without `--force`)
2. Parses all proto files from the configured paths
3. Generates TypeScript data files (`src/data/api/*.ts`)
4. Generates MDX page stubs (`src/content/docs/reference/api/*.mdx`)
5. Generates the landing page, API index, and comment guide
6. Generates `astro.config.mjs` with the sidebar
7. Writes any custom pages from the config

```sh
proto2astro generate                              # uses proto2astro.yaml
proto2astro generate -c custom-config.yaml        # custom config path
proto2astro generate -p ./proto -o ./docs         # override paths via flags
proto2astro generate --buf-workspace ./            # discover protos from buf workspace
```

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--config` | `-c` | `proto2astro.yaml` | Config file path |
| `--proto` | `-p` | *(from config)* | Proto file or directory (overrides config) |
| `--out` | `-o` | *(from config)* | Output directory (overrides config) |
| `--buf-workspace` | | *(from config)* | Buf workspace root (overrides config) |

### `proto2astro install`

Run `npm install` in the generated site directory. Required once before the first build, or after updating `package.json`.

```sh
proto2astro install                  # default: ./docs
proto2astro install -o ./my-docs
```

### `proto2astro build`

Build the site into static HTML. Output goes to `<site-dir>/dist/`.

```sh
proto2astro build                    # default: ./docs
proto2astro build -o ./my-docs
```

### `proto2astro dev`

Start the Astro dev server with hot-reload at `http://localhost:4321`. Useful while iterating on custom components or styles.

```sh
proto2astro dev                      # default: ./docs
proto2astro dev -o ./my-docs
```

### `proto2astro preview`

Serve the already-built static site locally for final review before deployment. You must run `build` first.

```sh
proto2astro preview                  # default: ./docs
proto2astro preview -o ./my-docs
```

### Typical Workflow

```sh
# First time
proto2astro generate
proto2astro install
proto2astro build

# After changing protos
proto2astro generate
proto2astro build

# Iterate on styling/components
proto2astro dev                      # hot-reload while editing files in docs/src/
```

---

## Configuration

All configuration lives in `proto2astro.yaml`. The generated `astro.config.mjs` is **fully regenerated on each `generate` run** — customize the site through this YAML file and through direct edits to components/styles (see [Customizing the Generated Site](#customizing-the-generated-site)).

### Minimal Example

```yaml
title: "My API"
proto:
  paths:
    - ./proto
```

### Full Example

```yaml
# ── Project metadata ─────────────────────────────────
title: "Payment API"
description: "API reference for the Payment service"
site: "https://example.github.io"       # base URL for deployed site (enables sitemap)
base: "/payment-api"                     # URL base path
logo: "./src/assets/logo.svg"            # logo shown in the header
edit_link: "https://github.com/example/payment/edit/main/proto"

social:
  - icon: github
    label: GitHub
    href: https://github.com/example/payment

# ── Proto input ──────────────────────────────────────
proto:
  # Option A: direct file/directory paths
  paths:
    - ./proto/payment
    - ./proto/common/types.proto

  # Option B: Buf workspace (alternative to paths)
  # buf_workspace: ./
  # buf_modules:
  #   - proto/payment
  #   - proto/common

# ── Output ───────────────────────────────────────────
out_dir: ./docs                          # default

# ── Sidebar ordering ─────────────────────────────────
service_order:                           # services appear in this order
  - PaymentService
  - RefundService
  - WebhookService

# ── Type handling ────────────────────────────────────
entity_types:                            # nested messages that should NOT be flattened
  - PaymentIntent
  - Customer

# ── Per-service overrides ────────────────────────────
services:
  PaymentService:
    description: "Handles payment processing and lifecycle management."
    notes: |
      All monetary amounts are in the smallest currency unit (e.g., cents for USD).
    footer: |
      See the [webhook guide](/guides/webhooks) for payment event notifications.
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

# ── Custom pages ─────────────────────────────────────
custom_pages:
  - title: "Webhooks"
    slug: webhooks
    content: |
      # Webhook Events
      Payment events are sent to your configured endpoint...
```

### Configuration Reference

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `title` | string | `"API Documentation"` | Site title (shown in header and landing page) |
| `description` | string | `"API reference documentation"` | Site description |
| `site` | string | | Base URL for the deployed site (enables sitemap generation) |
| `base` | string | | URL base path (e.g., `/my-api`) |
| `logo` | string | | Path to a logo image (relative to the site root) |
| `social` | list | | Social/external links in the header (`icon`, `label`, `href`) |
| `edit_link` | string | | Base URL for "edit this page" links |
| `proto.paths` | list | | Proto files or directories to parse |
| `proto.buf_workspace` | string | | Buf workspace root directory |
| `proto.buf_modules` | list | | Specific buf modules to include |
| `out_dir` | string | `"./docs"` | Output directory for the generated site |
| `service_order` | list | | Explicit sidebar ordering for services |
| `entity_types` | list | | Message types that should not be flattened into parent fields |
| `services` | map | | Per-service overrides (see overlay format above) |
| `custom_pages` | list | | Additional markdown pages (added under `/guides/`) |

---

## Customizing the Generated Site

The generated site is a standard Astro Starlight project. You can customize it directly — proto2astro is designed to coexist with your manual changes.

### What Gets Overwritten on Each `generate` Run

| Path | Behavior |
|------|----------|
| `astro.config.mjs` | **Always overwritten.** Regenerated from `proto2astro.yaml`. |
| `src/data/api/*.ts` | **Always overwritten.** Generated from proto files. |
| `src/content/docs/reference/api/*.mdx` | **Always overwritten.** MDX stubs for each service/enum. |
| `src/content/docs/reference/api/index.md` | **Always overwritten.** API index page. |
| `src/content/docs/index.md` | **Always overwritten.** Root landing page. |
| `src/content/docs/guides/comment-guide.md` | **Always overwritten.** Bundled comment guide. |
| Custom pages from `custom_pages` config | **Always overwritten.** Written to `src/content/docs/guides/`. |

### What Is Safe to Edit

| Path | Behavior |
|------|----------|
| `src/components/*.astro` | **Safe.** Scaffold files are only written once (skipped if they already exist). Edit freely. |
| `src/styles/custom.css` | **Safe.** Only written on first scaffold. |
| `src/content.config.ts` | **Safe.** Only written on first scaffold. |
| `src/data/api/types.ts` | **Safe.** Only written on first scaffold. |
| `package.json` | **Safe.** Only written on first scaffold. |
| `tsconfig.json` | **Safe.** Only written on first scaffold. |
| Any new files you add | **Safe.** proto2astro never deletes files. |

In short: **scaffold files** (components, styles, package.json, tsconfig) are write-once. **Generated content** (data, pages, astro config) is overwritten every run.

If you need to reset a scaffold file to the proto2astro default (e.g., after upgrading proto2astro), use `proto2astro init --force`.

### Adding Custom Pages

**Via config (recommended for simple pages):**

```yaml
custom_pages:
  - title: "Authentication"
    slug: authentication
    content: |
      # Authentication
      All API requests require a Bearer token...
```

These appear under `/guides/` and are added to the sidebar automatically.

**By adding files directly:**

You can add any `.md` or `.mdx` file anywhere under `src/content/docs/`. These pages will be included in the Astro build. However, they won't appear in the sidebar unless you manually add them — because `astro.config.mjs` is regenerated each run.

To add a page that persists in the sidebar, use the `custom_pages` config. For pages that don't need sidebar entries (or that you link to from other pages), just drop the file in and link to it manually.

### Customizing Components

The Astro components in `src/components/` control how API pages are rendered:

| Component | Purpose |
|-----------|---------|
| `ApiServicePage.astro` | Full service page layout (two-column, endpoint list) |
| `ApiEndpoint.astro` | Single RPC endpoint section |
| `ApiRequest.astro` | Request fields table |
| `ApiResponse.astro` | Response fields table |
| `ErrorCodes.astro` | Error codes display |
| `EnumPage.astro` | Enum values documentation |

These are standard Astro components. Edit them to change layout, add features, or adjust styling. They won't be overwritten on subsequent `generate` runs.

### Customizing Styles

Edit `src/styles/custom.css` to change colors, spacing, or layout. This file is imported via `astro.config.mjs` and applies globally. It is not overwritten after the initial scaffold.

### Customizing the Landing Page

The root landing page (`src/content/docs/index.md`) uses Starlight's `splash` template with a hero section. It is regenerated on each `generate` run. To customize the landing page, you have two options:

1. **Set `title` and `description` in `proto2astro.yaml`** — these populate the hero section.
2. **Replace `index.md` with your own `index.mdx`** — note that `generate` will recreate `index.md` each run. If you want full control, rename your file to `index.mdx` (MDX takes precedence over MD in Astro) and proto2astro won't interfere.

### Adding npm Dependencies

Since `package.json` is write-once, you can add npm packages normally:

```sh
cd docs
npm install some-astro-plugin
```

Then configure the plugin in your code. Just remember that `astro.config.mjs` is regenerated — if the plugin requires config there, you'll need a different integration approach (e.g., importing it in a component).

---

## Proto Comment Conventions

proto2astro extracts structured information from your proto comments. All conventions are optional — plain `//` comments work as descriptions.

### Leading Comments

```protobuf
// Create a new user account.
// The email must be unique across the system.
rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);
```

### Required Fields

```protobuf
// Required. The user's email address.
string email = 1;
```

### Deprecated

```protobuf
// Deprecated: Use display_name instead.
string name = 1;
```

### Default Values

```protobuf
// The page size for results. Default: 20.
int32 page_size = 1;
```

### Range Constraints

```protobuf
// Number of items to return. Range: 1-100.
int32 limit = 1;
```

### Error Codes

```protobuf
// Create a new payment.
// Errors: ALREADY_EXISTS if a payment with this idempotency key exists.
// Errors: INVALID_ARGUMENT if the amount is negative.
rpc CreatePayment(CreatePaymentRequest) returns (CreatePaymentResponse);
```

### Example Values

```protobuf
// The user's email. @example "alice@example.com"
string email = 1;

// Tags for the resource. @example ["production", "us-east"]
repeated string tags = 2;
```

The `@example` value is JSON-parsed when possible (objects, arrays, numbers, booleans), otherwise treated as a string.

---

## Generated Site Structure

After running `proto2astro generate`, the output directory contains:

```
docs/
├── package.json              # Astro + Starlight deps (scaffold, write-once)
├── astro.config.mjs          # Sidebar + site config (regenerated each run)
├── tsconfig.json              # TypeScript config (scaffold, write-once)
├── src/
│   ├── content.config.ts      # Astro content config (scaffold, write-once)
│   ├── content/docs/
│   │   ├── index.md           # Landing page (regenerated)
│   │   ├── guides/
│   │   │   ├── comment-guide.md     # Proto comment conventions (regenerated)
│   │   │   └── <custom-page>.md     # Custom pages from config (regenerated)
│   │   └── reference/api/
│   │       ├── index.md             # API index listing (regenerated)
│   │       ├── <service>.mdx        # Service page stubs (regenerated)
│   │       └── enum-<name>.mdx      # Enum page stubs (regenerated)
│   ├── data/api/
│   │   ├── types.ts           # TS type definitions (scaffold, write-once)
│   │   ├── <service>.ts       # Service data files (regenerated)
│   │   └── enum-<name>.ts     # Enum data files (regenerated)
│   ├── components/            # Astro components (scaffold, write-once)
│   │   ├── ApiServicePage.astro
│   │   ├── ApiEndpoint.astro
│   │   ├── ApiRequest.astro
│   │   ├── ApiResponse.astro
│   │   ├── ErrorCodes.astro
│   │   └── EnumPage.astro
│   └── styles/
│       └── custom.css         # Site styles (scaffold, write-once)
├── node_modules/              # After proto2astro install
└── dist/                      # Built static HTML (after proto2astro build)
```

---

## How It Works

1. **Parse** — Reads `.proto` files using `emicklei/proto` (source-level, no protoc needed). Two-pass approach: first collects all messages/enums, then processes services/RPCs with full type resolution.

2. **Resolve** — Marks fields whose types are known messages or enums (`IsMessage`, `IsEnum`) so the frontend can link between pages.

3. **Generate data** — Emits TypeScript data files (one per service, one per enum) using Go structs serialized via `json.MarshalIndent` wrapped in typed TS exports.

4. **Generate pages** — Creates MDX stubs that import the TS data and render it through Astro components. Also generates a landing page, API index, and comment guide.

5. **Generate config** — Writes `astro.config.mjs` with a dynamic sidebar reflecting the proto package structure, social links, logo, and custom pages.

6. **Build** — The standard Astro SSG build produces static HTML in `dist/`.

---

## Deployment

The built site in `<out_dir>/dist/` is plain static HTML. Deploy it anywhere:

**GitHub Pages:**

```yaml
# .github/workflows/docs.yml
name: Deploy API Docs
on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
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
      - uses: actions/deploy-pages@v4
```

**Netlify / Vercel / Any static host:**

```sh
proto2astro generate && proto2astro install && proto2astro build
# Upload docs/dist/
```

## License

MIT
