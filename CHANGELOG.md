# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.4.0] - 2026-04-05

### Added

- Comprehensive test suite for parser and generator with golden file tests
- Consistent `@`-prefix annotation syntax (`@required`, `@deprecated`, `@default`, `@range`, `@error`, `@example`)
- Multi-line fenced `@example` blocks for complex JSON/object examples
- `entity_examples` config for type-level example values
- Scaffold config options (`landing_page`, `comment_guide`) to control generated pages
- Route collision detection to prevent conflicting page paths
- Deterministic output — sorted map iterations for reproducible generation
- MIT license

### Fixed

- HTML encoding in TypeScript data files (ampersands, angle brackets no longer escaped)
- Annotation markers leaking into rendered descriptions
- Recursive enum detection in nested message types

### Changed

- Sidebar generation rewritten as JSON config with scaffold-only behavior and custom page paths
- README restructured around command lifecycle

### Removed

- Legacy annotation syntax (bare `Required`, `Deprecated:`, `Default:`, `Range:`, `Errors:`) — only `@`-prefix annotations are now supported

## [0.3.0] - 2025-12-15

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
- Oneof field support — fields inside `oneof` blocks are parsed with `IsOneof`/`OneofGroup` metadata
- Nested message/enum support — recursively collects qualified names (e.g., `Outer.Inner`) for both messages and enums
- Streaming RPC detection — `StreamsRequest`/`StreamsResponse` fields populated and noted in generation output
- Cross-package type resolution — two-pass resolver with global type index and fully-qualified name support
- Recursive type cycle detection — `flattenFields` uses a `seen` set to break infinite loops on self-referencing messages
- MDX frontmatter YAML escaping — `yamlString` helper prevents broken frontmatter from special characters in titles
- Config validation — `Validate()` checks proto paths exist, validates buf workspace, validates service_order/custom_pages, and warns about defaults
- Enhanced generation summary — shows RPC counts, streaming RPC notes, and page/data file totals
- Build version injection via ldflags (`version`, `commit`, `date`)
- Makefile for dev tasks (`build`, `test`, `vet`, `lint`, `fmt`, `clean`, `snapshot`, `release`)
- goreleaser config for cross-platform releases (macOS + Linux, amd64 + arm64)
