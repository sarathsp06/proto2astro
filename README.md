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
  <a href="LICENSE"><img src="https://img.shields.io/github/license/sarathsp06/proto2astro" alt="License"></a>
</p>

---

proto2astro generates a complete [Astro Starlight](https://starlight.astro.build/) documentation site from your [Protocol Buffer](https://protobuf.dev/) definitions. Two-column layout, auto-generated curl examples, organized sidebar — built for [ConnectRPC](https://connectrpc.com/) services. No `protoc` needed.

```
  .proto files  ──▶  proto2astro  ──▶  Static docs site (Astro Starlight)
```

**See it in action:** [Sparrow API Reference](https://sarathsp06.github.io/sparrow/reference/api/)

## Highlights

- **Zero `protoc`** — parses `.proto` source files directly, no compilation step
- **Two-column layout** — fields on the left, curl examples on the right
- **ConnectRPC-native** — curl examples use HTTP POST + JSON with correct service paths
- **Buf support** — discovers protos from `buf.yaml` workspaces automatically
- **Rich proto comments** — extracts `@required`, `@deprecated`, `@default`, `@range`, `@error`, `@example`

## Install

**Go:**

```sh
go install github.com/sarathsp06/proto2astro/cmd/proto2astro@latest
```

**Pre-built binaries** for macOS, Linux, and Windows (arm64/amd64) are on the [releases page](https://github.com/sarathsp06/proto2astro/releases).

## Quick Start

**1. Create a config file**

```yaml
# proto2astro.yaml
title: "My API"
proto:
  paths:
    - ./proto
```

**2. Generate the docs site**

```sh
proto2astro generate
```

Parses your `.proto` files and scaffolds an Astro Starlight project in `./docs` with MDX pages, TypeScript data files, and a configured sidebar.

**3. Install dependencies and build**

```sh
proto2astro install       # npm install
proto2astro build         # static HTML → docs/dist/
```

**4. Preview locally**

```sh
proto2astro preview       # http://localhost:4321
```

Your docs are ready. Upload `docs/dist/` to any static host.

## Commands

proto2astro manages the full lifecycle of your documentation site. All commands accept `-o` / `--out` to set the site directory (default: `./docs`).

### `proto2astro init`

Scaffold the Astro Starlight project structure — components, styles, types, `package.json` — without generating any API content. Useful when you want to customize the site before your first generate.

```sh
proto2astro init              # scaffold into ./docs
proto2astro init ./my-docs    # scaffold into ./my-docs
proto2astro init --force      # overwrite existing scaffold files
```

### `proto2astro generate`

The main command. Parses all `.proto` files from your config, resolves cross-package types, and generates TypeScript data files + MDX pages + `astro.config.mjs`. If the site doesn't exist yet, it scaffolds it first.

```sh
proto2astro generate                           # uses proto2astro.yaml
proto2astro generate -c custom-config.yaml     # custom config file
proto2astro generate -p ./proto -o ./docs      # override paths via flags
proto2astro generate --buf-workspace ./         # discover from buf workspace
```

### `proto2astro install`

Runs `npm install` in the site directory to install Astro, Starlight, and other dependencies.

```sh
proto2astro install
```

### `proto2astro dev`

Starts a local dev server with hot-reload. Use this when customizing components or styles.

```sh
proto2astro dev               # http://localhost:4321
```

### `proto2astro build`

Builds the site into static HTML at `<site>/dist/`.

```sh
proto2astro build
```

### `proto2astro preview`

Serves the built static site locally for a final check before deploying.

```sh
proto2astro preview           # http://localhost:4321
```

### Typical workflows

```sh
# First time setup
proto2astro generate && proto2astro install && proto2astro build

# After changing .proto files
proto2astro generate && proto2astro build

# Customizing the look and feel
proto2astro dev
```

## Configuration

All config lives in `proto2astro.yaml`.

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
<summary><strong>Full example</strong></summary>

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

# ── Sidebar ──────────────────────────────────────
# Sections rendered before/after the auto-generated API Reference group.
sidebar:
  before:
    - label: "Getting Started"
      items:
        - slug: getting-started/installation
        - slug: getting-started/quickstart
    - label: "Guides"
      items:
        - label: "Comment Guide"
          slug: guides/comment-guide
  after:
    - label: "Resources"
      items:
        - slug: resources/changelog

# ── Sidebar ordering (within API Reference) ─────
service_order:
  - PaymentService
  - RefundService

# ── Starlight component overrides ────────────────
components:
  Footer: "./src/components/Footer.astro"

# ── Additional CSS ───────────────────────────────
custom_css:
  - "./src/styles/brand.css"

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
    slug: webhooks                          # → guides/webhooks.md
    content: |
      # Webhook Events
      Payment events are sent to your configured endpoint...

  - title: "Kubernetes Deployment"
    path: deployment/kubernetes              # → deployment/kubernetes.md
    content: |
      # Deploying to Kubernetes
      Use the Helm chart to deploy...
```

</details>

<details>
<summary><strong>All config fields</strong></summary>

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
| `sidebar.before` | list | | Sidebar sections rendered before API Reference |
| `sidebar.after` | list | | Sidebar sections rendered after API Reference |
| `components` | map | | Starlight component overrides (e.g., `Footer: ./src/components/Footer.astro`) |
| `custom_css` | list | | Additional CSS files beyond the default `custom.css` |
| `service_order` | list | | Explicit sidebar ordering for services |
| `entity_types` | list | | Message types that should not be flattened |
| `services` | map | | Per-service overrides (descriptions, examples, fields) |
| `custom_pages` | list | | Additional pages (`slug` for `guides/`, `path` for arbitrary location) |

</details>

## Proto Comment Conventions

Write comments in your `.proto` files and proto2astro picks them up automatically. All annotations use the `@` prefix for consistency.

```protobuf
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
```

| Annotation | What it does |
|---|---|
| `@required` | Marks the field as required |
| `@deprecated` | Shows deprecation notice; field excluded from docs |
| `@default VALUE` | Displays the default value |
| `@range MIN-MAX` | Shows allowed range constraint |
| `@error CODE desc` | Lists error codes on the RPC |
| `@example VALUE` | Uses the value in generated curl examples |

> **Legacy syntax** — `Required.`, `Deprecated:`, `Default:`, `Range:`, and `Errors:` patterns are still supported for backward compatibility.

## Customizing the Site

The generated site is a standard Astro Starlight project. You can edit components, styles, and pages directly.

**Safe to edit** (scaffold-only, never overwritten by `generate`):
`astro.config.mjs`, `src/components/*.astro`, `src/styles/custom.css`, `src/content.config.ts`, `src/data/api/types.ts`, `package.json`, `tsconfig.json`, `src/content/docs/index.mdx` (landing page), `src/content/docs/guides/comment-guide.md`

**Regenerated** on every `generate` run:
`src/data/proto2astro-config.json` (sidebar + site settings), `src/data/api/*.ts` (service/enum data), `src/content/docs/reference/api/*.mdx` (service stubs), `src/content/docs/reference/api/index.md`

### How it works

`astro.config.mjs` is written once and imports `src/data/proto2astro-config.json` for all proto-derived and YAML-derived settings (title, sidebar, social links, etc.). When you re-run `generate`, only the JSON file is updated — your customizations to `astro.config.mjs` (custom integrations, Vite config, i18n, etc.) are preserved.

Add sidebar sections via `proto2astro.yaml`:

```yaml
sidebar:
  before:
    - label: "Getting Started"
      items:
        - slug: getting-started/installation
  after:
    - label: "Resources"
      items:
        - slug: resources/changelog
```

Or edit `astro.config.mjs` directly — it won't be overwritten.

Run `proto2astro init --force` to reset scaffold files to defaults after upgrading.

## Deployment

The built site is static HTML — deploy it anywhere.

```sh
proto2astro generate && proto2astro install && proto2astro build
# Upload docs/dist/ to GitHub Pages, Netlify, Vercel, or any static host
```

## License

[MIT](LICENSE)
