# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Cobra CLI with `init`, `generate`, `install`, `build`, `dev`, and `preview` commands
- Proto parser using `github.com/emicklei/proto` (source-level, no protoc required)
- Two-pass parsing: messages/enums first, then services/RPCs with type resolution
- Comment extraction supporting leading `//` comments with pattern detection:
  `Required`, `Deprecated:`, `Default: VALUE`, `Range: MIN-MAX`, `Errors: CODE`, `@example`
- YAML config overlay (`proto2astro.yaml`) for service order, entity types, per-field examples, and description overrides
- Full Astro Starlight site scaffold generation (Astro 6 + Starlight 0.38.2)
- TypeScript data file generation using Go structs + `json.MarshalIndent` with thin TS wrapper
- MDX stub pages for services and enums with two-column layout components
- Auto-generated curl examples for every RPC (ConnectRPC HTTP POST + JSON)
- Enum reference pages with all values, descriptions, and numeric values
- Sidebar auto-generated from proto packages via `text/template`
- Splash landing page with Starlight CardGrid feature highlights
- API reference index page listing all services and enums
- Proto comment guide page bundled in the generated site
- Buf workspace integration for proto file discovery
- npm orchestration in `internal/npm/` (install, build, dev, preview)
- Scaffold overwrite semantics: scaffold files are write-once, generated content always regenerated
- `--force` flag on `init` for upgrading scaffold files
- Comprehensive README with installation, usage, config reference, and customization guide
