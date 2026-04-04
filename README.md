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

proto2astro generates a complete [Astro Starlight](https://starlight.astro.build/) documentation site from your [Protocol Buffer](https://protobuf.dev/) definitions. Two-column layout, auto-generated curl examples, organized sidebar — ready for [ConnectRPC](https://connectrpc.com/) services. No `protoc` needed.

```
  ┌─────────────┐         ┌──────────────┐         ┌─────────────────────┐
  │  .proto      │────────▶│  proto2astro  │────────▶│  Static docs site   │
  │  files       │         │              │         │  (Astro Starlight)  │
  └─────────────┘         └──────────────┘         └─────────────────────┘
```

**See it in action:** [Sparrow API Reference](https://sarathsp06.github.io/sparrow/reference/api/)

## Highlights

- **Zero `protoc`** — parses `.proto` source files directly, no compilation step
- **Two-column layout** — fields on the left, curl examples on the right
- **ConnectRPC-native** — curl examples use HTTP POST + JSON with correct service paths
- **Buf support** — discovers protos from `buf.yaml` workspaces automatically
- **Rich proto comments** — extracts `Required`, `Deprecated:`, `Default:`, `Range:`, `Errors:`, `@example`
- **Full site lifecycle** — `init` / `generate` / `install` / `build` / `dev` / `preview`

## Quick Start

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

## CLI

All commands accept `-o` / `--out` to set the site directory (default: `./docs`).

| Command | Description |
|---------|-------------|
| `proto2astro init [dir]` | Scaffold the Astro project (no API content) |
| `proto2astro generate` | Parse protos and generate docs |
| `proto2astro install` | Run `npm install` in the site directory |
| `proto2astro build` | Build static HTML to `<site>/dist/` |
| `proto2astro dev` | Dev server with hot-reload (localhost:4321) |
| `proto2astro preview` | Serve the built site locally |

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

## Configuration

All config lives in a single `proto2astro.yaml`.

<details>
<summary><strong>Minimal</strong></summary>

```yaml
title: "My API"
proto:
  paths:
    - ./proto
```

</details>

<details>
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
| `Errors:` | Lists error codes on the RPC |
| `@example` | Uses the value in generated curl examples |

## Customizing the Site

The generated site is a standard Astro Starlight project. You can customize components, styles, and pages.

**Written once (safe to edit):**
`src/components/*.astro`, `src/styles/custom.css`, `src/content.config.ts`, `src/data/api/types.ts`, `package.json`, `tsconfig.json`

**Regenerated on every `generate` run:**
`astro.config.mjs`, `src/data/api/*.ts`, `src/content/docs/reference/api/*.mdx`, `src/content/docs/index.md`

### Customizable components

| Component | Controls |
|---|---|
| `ApiServicePage.astro` | Full service page (two-column layout) |
| `ApiEndpoint.astro` | Individual RPC endpoint |
| `ApiRequest.astro` | Request fields table |
| `ApiResponse.astro` | Response fields table |
| `ErrorCodes.astro` | Error code display |
| `EnumPage.astro` | Enum values page |

> Use `proto2astro init --force` to reset scaffold files to defaults after upgrading.

## Deployment

The built site is static HTML — deploy it anywhere.

```sh
proto2astro generate && proto2astro install && proto2astro build
# Upload docs/dist/ to GitHub Pages, Netlify, Vercel, or any static host
```

## Development

### Requirements

- Go 1.25+
- Node.js 18+
- macOS or Linux

### Build from source

```sh
git clone https://github.com/sarathsp06/proto2astro.git
cd proto2astro
make build       # → bin/proto2astro
make install     # → $GOPATH/bin/proto2astro
```

### Make targets

```sh
make build       # Build the binary
make test        # Run tests with -race
make vet         # go vet
make lint        # golangci-lint
make fmt         # gofmt + goimports
make clean       # Remove bin/ and dist/
```

## License

[MIT](LICENSE)
