# proto2astro Improvements

Bug fixes and feature requests discovered during real-world integration with the [Sparrow](https://github.com/sarathsp06/sparrow) webhook delivery platform (~1700-line proto file, 5 services, 2 enums, 60+ RPCs, entity type flattening, `@example` annotations on 70+ fields).

---

## Bugs

### 1. HTML entity encoding in generated JSON/TS files

**Severity**: Medium — corrupts rendered output for any field description or example containing `<`, `>`, or `&`.

**Location**: `internal/generator/emit.go:62`, `emit.go:74`, `sidebar.go:57`

**Problem**: `json.MarshalIndent()` uses Go's default `json.Marshal` behavior which escapes `<`, `>`, and `&` as `\u003c`, `\u003e`, `\u0026`. This means any proto comment containing HTML-like content (e.g., `"values < 100"`, `"A & B"`) or angle brackets in examples gets mangled in the TypeScript data files and `proto2astro-config.json`.

**Fix**:
```go
// Before (emit.go:62)
data, err := json.MarshalIndent(svc, "", "  ")

// After
var buf bytes.Buffer
enc := json.NewEncoder(&buf)
enc.SetEscapeHTML(false)
enc.SetIndent("", "  ")
if err := enc.Encode(svc); err != nil { ... }
data := bytes.TrimRight(buf.Bytes(), "\n")
```

Apply the same fix to `marshalEnumTS()` (emit.go:74) and `generateProto2AstroConfig()` (sidebar.go:57).

---

### 2. `@example` annotation text not stripped from rendered description

**Severity**: Medium — every field with `@example` shows the raw annotation text in the description column of the docs page.

**Location**: `internal/generator/data.go:285-291` (`cleanDescription()`)

**Problem**: `cleanDescription()` only strips the `Required.` prefix. When a proto comment contains `@example "some_value"`, the `@example` text is extracted for the example column but is also left verbatim in the description string. The rendered docs show something like:

> UUID of the webhook to query. @example "d290f1ee-6c54-4b01-90e6-d701748f0851"

**Fix**: Add `@example` stripping to `cleanDescription()`:
```go
func cleanDescription(desc string) string {
    desc = strings.TrimSpace(desc)
    desc = strings.TrimPrefix(desc, "Required. ")
    desc = strings.TrimPrefix(desc, "Required ")
    // Strip @example annotation and everything after it
    if idx := strings.Index(desc, "@example"); idx >= 0 {
        desc = strings.TrimSpace(desc[:idx])
    }
    return desc
}
```

**Note**: The same issue may apply to `Default:` and `Range:` annotations — they are extracted as structured data but also left in the description text. Consider stripping all recognized annotations from the description.

---

### 3. `usedBy` misses enum usage inside entity type messages

**Severity**: Low — `usedBy` list on enum pages is incomplete. Enums used inside entity types (nested messages) are not detected.

**Location**: `internal/generator/data.go:142-167` (`findEnumUsage()`, `messageUsesEnum()`)

**Problem**: `findEnumUsage()` only checks direct fields of RPC request/response messages. If an enum is used inside a nested message (especially an entity type that isn't flattened), it's missed.

**Example**: `WebhookDeliveryStatus` enum is used in `WebhookDelivery` message (an entity type). `WebhookDelivery` appears as a field in `GetDeliveryStatusResponse`, `ListDeliveriesResponse`, etc. But `findEnumUsage()` only checks direct fields of `GetDeliveryStatusResponse` — it finds `WebhookDelivery` (a message type), not `WebhookDeliveryStatus` (an enum inside it).

Meanwhile, `WebhookHealth` enum IS found because it appears directly in `GetWebhookHealthResponse.health`.

**Fix**: Make `messageUsesEnum()` recursive — walk into nested message fields:
```go
func messageUsesEnum(msg *parser.ProtoMessage, enumName string, pkg *parser.ProtoPackage, seen map[string]bool) bool {
    if seen == nil {
        seen = make(map[string]bool)
    }
    if seen[msg.Name] {
        return false // cycle guard
    }
    seen[msg.Name] = true

    for _, f := range msg.Fields {
        if f.IsEnum && f.RawType == enumName {
            return true
        }
        if f.IsMessage {
            if sub, ok := pkg.Messages[f.RawType]; ok {
                if messageUsesEnum(sub, enumName, pkg, seen) {
                    return true
                }
            }
        }
    }
    return false
}
```

---

## Feature Requests

### 4. Test suite

**Priority**: High

**Problem**: Zero test files in the entire codebase. No unit tests, no integration tests, no golden file tests. This makes it risky to fix any of the bugs above without introducing regressions.

**Suggested approach**:

- **Golden file tests for generated output**: Parse a small fixture `.proto` file, generate data files, and compare against checked-in golden files. This catches encoding issues (#1), description stripping issues (#2), and enum usage issues (#3) in one pass.
- **Unit tests for comment parsing**: `internal/parser/comments.go` has clear pure-function boundaries (`extractExample`, `extractRequired`, `extractDefault`, `extractRange`, `splitRPCComment`). Table-driven tests are straightforward.
- **Unit tests for `cleanDescription`**: Test that all annotation patterns are stripped.
- **Unit tests for `shouldFlatten`**: Entity type vs non-entity type behavior.
- **Integration test**: Parse `webhook.proto` (or a trimmed version) end-to-end and verify the generated TS files are valid JSON.

---

### 5. Scaffold conflict detection

**Priority**: Medium

**Problem**: `proto2astro generate` scaffolds `index.mdx` and `guides/comment-guide.md` into the content directory. If the user already has a custom landing page (e.g., `src/pages/index.astro` in Sparrow), the scaffolded `index.mdx` creates a route conflict. Starlight renders the scaffolded page, shadowing the user's custom page, or produces confusing warnings.

**Suggested approach**:

- Before writing scaffold-only files, check if a file already exists at the target path (any extension: `.md`, `.mdx`, `.astro`).
- Also check `src/pages/` for route collisions (e.g., `src/pages/index.astro` vs `src/content/docs/index.mdx`).
- Skip scaffold files that would conflict, or print a warning and skip.
- Consider a `--no-scaffold` flag or a config option `scaffold: false` to skip all scaffold files on `generate`.

---

### 6. Config option to exclude scaffold files

**Priority**: Low

**Problem**: The `comment-guide.md` scaffold is useful for proto2astro developers learning the annotation conventions, but it's noise in a production docs site. There's no way to prevent it from being created (short of deleting it after every `generate` run).

**Suggested approach**: Add a config option:
```yaml
scaffold:
  index: false          # skip scaffolded index.mdx
  comment_guide: false  # skip scaffolded comment-guide.md
```

Or a simpler blocklist:
```yaml
exclude_scaffold:
  - index.mdx
  - guides/comment-guide.md
```

---

### 7. Strip all recognized annotations from description text

**Priority**: Medium — related to bug #2 but broader.

**Problem**: `@example`, `Default:`, and `Range:` annotations are parsed into structured data but also left verbatim in the description string. This creates redundant/noisy descriptions.

**Suggested approach**: `cleanDescription()` should strip all recognized annotation patterns:
```go
func cleanDescription(desc string) string {
    desc = strings.TrimSpace(desc)
    desc = strings.TrimPrefix(desc, "Required. ")
    desc = strings.TrimPrefix(desc, "Required ")
    // Strip @example and everything after it (usually last in comment)
    if idx := strings.Index(desc, "@example"); idx >= 0 {
        desc = strings.TrimSpace(desc[:idx])
    }
    // Strip Default: VALUE. pattern
    desc = defaultRE.ReplaceAllString(desc, "")
    // Strip Range: MIN-MAX pattern
    desc = rangeRE.ReplaceAllString(desc, "")
    return strings.TrimSpace(desc)
}
```

**Consideration**: This changes behavior for existing users who may rely on seeing "Default: 50" in the description. Could be gated behind a config flag: `strip_annotations: true`.

---

### 8. Multi-line `@example` support

**Priority**: Low

**Problem**: The `@example` regex (`@example\s+(.+)`) captures to end of the joined comment string. Since proto comments are joined into a single line before parsing, multi-line JSON examples must be written as a single line in the proto comment. This is awkward for complex examples:

```protobuf
// JSON schema defining the event payload structure.
// @example {"type": "object", "properties": {"order_id": {"type": "string"}, "amount": {"type": "number"}, "currency": {"type": "string"}}}
string json_schema = 3;
```

**Suggested approach**: Support a block syntax using fenced delimiters:
```protobuf
// JSON schema defining the event payload structure.
// @example ```
// {
//   "type": "object",
//   "properties": {
//     "order_id": {"type": "string"},
//     "amount": {"type": "number"}
//   }
// }
// ```
string json_schema = 3;
```

This would require changes to `joinCommentLines()` in `comments.go` to preserve newlines within fenced blocks, and to `extractExample()` to detect the fenced pattern.

---

### 9. `@example` on entity type fields

**Priority**: Medium

**Problem**: Entity types (listed in `entity_types` config) are not flattened — they appear as a single field with their message type as the type name. But there's no way to provide an `@example` for the entity field itself from the proto comment, because the `@example` annotation is parsed from the field that *references* the entity type (which is in the request/response message), not from the entity type definition.

In practice, for a response field like:
```protobuf
// The delivery record.
WebhookDelivery delivery = 1;
```

You'd want to put `@example` on this field, but the example would need to be the entire JSON representation of a `WebhookDelivery`, which is unwieldy on one line.

**Suggested approach**: Allow entity types to have a default example defined once in the message itself (or in config), rather than requiring it on every field reference. A proto annotation like `@entity_example` on the message comment, or a config option:
```yaml
entity_types:
  - name: WebhookDelivery
    example_file: examples/webhook-delivery.json
```

---

### 10. Consistent `@` prefix for all annotations

**Priority**: High — this is a developer experience and API design consistency issue.

**Problem**: The annotation syntax is inconsistent. `@example` uses an `@` prefix, but all other annotations use different conventions:

| Annotation | Current Syntax | Style |
|---|---|---|
| Required | `Required.` or `Required ` | Bare word, period-terminated |
| Default | `Default: VALUE.` | Colon-suffix, period-terminated |
| Range | `Range: MIN-MAX` | Colon-suffix |
| Deprecated | `Deprecated: reason` | Colon-suffix |
| Errors | `Errors: CODE description` | Colon-suffix |
| Example | `@example VALUE` | `@` prefix |

This is confusing for users. The `@` prefix is the more natural convention for structured annotations in comments (it mirrors Javadoc `@param`, JSDoc `@returns`, Doxygen `@brief`, etc.). All annotations should use the `@` prefix for consistency.

**Proposed unified syntax**:
```protobuf
// @required The webhook endpoint URL.
// @example "https://example.com/webhook"
// @default "https://localhost"
string url = 1;

// Number of items per page.
// @range 1-100
// @default 50
// @example 25
int32 page_size = 2;

// @deprecated Use new_field instead.
string old_field = 5;
```

For RPCs:
```protobuf
// Create a new webhook subscription.
// @error ALREADY_EXISTS if a webhook with the same URL already exists.
// @error INVALID_ARGUMENT if the URL is malformed.
rpc RegisterWebhook(RegisterWebhookRequest) returns (RegisterWebhookResponse);
```

**Migration strategy**: Support both old and new syntax during a transition period. Parse `@required`, `@default VALUE`, `@range MIN-MAX`, `@deprecated`, `@error CODE desc` alongside the existing patterns. Deprecate the old syntax in docs and emit warnings when the old patterns are detected.

**Files to update**:
- `internal/parser/comments.go` — Add new `@`-prefixed regexes, update `extractRequired()`, `extractDefault()`, `extractRange()`, `extractDeprecated()`, `splitRPCComment()`
- `internal/generator/pages.go` — Update `commentGuideMD` constant (the embedded comment guide) with new syntax
- `internal/generator/data.go` — Update `cleanDescription()` to strip all `@`-prefixed annotations
- `README.md` — Update the "Proto Comment Conventions" section with new syntax and migration notes
- `CHANGELOG.md` — Document the syntax change

---

### 11. Deterministic output for diffable regeneration

**Priority**: Low

**Problem**: Go map iteration order is non-deterministic. While service ordering is handled via `sortedServiceNames()`, other map iterations (e.g., overlay fields, package iteration in multi-package mode) may produce non-deterministic output across runs. This makes `git diff` noisy after regeneration even when nothing changed.

**Suggested approach**: Audit all map iterations and ensure stable sorting. Key locations:
- `data.go:21` — `result.Packages` iteration
- `sidebar.go:122` — `result.Packages` iteration (already sorted)
- Field overlay application order

---

## Files to Update

When implementing these improvements, the following files will need changes. This list is a cross-reference to help plan work:

| File | Relevant Items |
|------|---------------|
| `internal/generator/emit.go` | #1 (HTML encoding) |
| `internal/generator/sidebar.go` | #1 (HTML encoding) |
| `internal/generator/data.go` | #2, #3, #7, #10 (cleanDescription, findEnumUsage, annotation stripping, consistent @ prefix) |
| `internal/parser/comments.go` | #2, #7, #8, #10 (annotation extraction, multi-line, consistent @ prefix) |
| `internal/generator/pages.go` | #5, #6, #10 (scaffold conflict, exclude config, update commentGuideMD) |
| `internal/generator/scaffold.go` | #5, #6 (scaffold conflict detection) |
| `internal/config/types.go` | #6, #9 (scaffold exclusion config, entity example config) |
| `internal/config/config.go` | #6, #9 (config validation) |
| `README.md` | #10 (update Proto Comment Conventions section with `@` prefix syntax) |
| `CHANGELOG.md` | All items (document changes) |
| `internal/generator/templates/` | #5 (scaffold awareness) |
| New: `internal/parser/comments_test.go` | #4 (unit tests) |
| New: `internal/generator/data_test.go` | #4 (unit tests) |
| New: `internal/generator/emit_test.go` | #4 (golden file tests) |
| New: `testdata/` | #4 (fixture proto files + golden output files) |

---

## Summary

| # | Type | Severity | Effort | Description |
|---|------|----------|--------|-------------|
| 1 | Bug | Medium | Small | HTML entity encoding in JSON output |
| 2 | Bug | Medium | Small | `@example` not stripped from description |
| 3 | Bug | Low | Small | `usedBy` misses enums in nested messages |
| 4 | Feature | High | Medium | Add test suite |
| 5 | Feature | Medium | Small | Scaffold conflict detection |
| 6 | Feature | Low | Small | Config to exclude scaffold files |
| 7 | Feature | Medium | Small | Strip all annotations from description |
| 8 | Feature | Low | Medium | Multi-line `@example` support |
| 9 | Feature | Medium | Medium | `@example` on entity type fields |
| 10 | Feature | High | Medium | Consistent `@` prefix for all annotations |
| 11 | Feature | Low | Small | Deterministic output ordering |
