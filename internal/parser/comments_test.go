package parser

import (
	"fmt"
	"testing"
)

func TestJoinCommentLines(t *testing.T) {
	tests := []struct {
		name  string
		lines []string
		want  string
	}{
		{
			name:  "single line",
			lines: []string{"Hello world"},
			want:  "Hello world",
		},
		{
			name:  "multiple lines same paragraph",
			lines: []string{"Line one", "line two", "line three"},
			want:  "Line one line two line three",
		},
		{
			name:  "two paragraphs",
			lines: []string{"Para one", "", "Para two"},
			want:  "Para one Para two",
		},
		{
			name:  "empty input",
			lines: nil,
			want:  "",
		},
		{
			name:  "only blanks",
			lines: []string{"", "", ""},
			want:  "",
		},
		{
			name:  "trailing blank",
			lines: []string{"Hello", ""},
			want:  "Hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := joinCommentLines(tt.lines)
			if got != tt.want {
				t.Errorf("joinCommentLines() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExtractRequired(t *testing.T) {
	tests := []struct {
		name    string
		comment string
		want    bool
	}{
		{"@required annotation", "@required The webhook endpoint URL.", true},
		{"@required in middle", "The URL. @required field.", true},
		{"@requiredfoo no match", "This has @requiredfoo which is not valid.", false},
		{"no @required", "The webhook endpoint URL.", false},
		{"empty", "", false},
		{"bare Required word", "Required. The URL.", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractRequired(tt.comment)
			if got != tt.want {
				t.Errorf("extractRequired(%q) = %v, want %v", tt.comment, got, tt.want)
			}
		})
	}
}

func TestExtractDeprecated(t *testing.T) {
	tests := []struct {
		name    string
		comment string
		want    bool
	}{
		{"@deprecated annotation", "@deprecated Use new_field instead.", true},
		{"@deprecated case insensitive", "@Deprecated Use old.", true},
		{"@deprecated not at start", "This field is @deprecated.", false},
		{"@deprecated with whitespace", "  @deprecated old field", true},
		{"no deprecated", "The webhook URL.", false},
		{"empty", "", false},
		{"legacy Deprecated: not matched", "Deprecated: Use new_field instead.", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractDeprecated(tt.comment)
			if got != tt.want {
				t.Errorf("extractDeprecated(%q) = %v, want %v", tt.comment, got, tt.want)
			}
		})
	}
}

func TestExtractDefault(t *testing.T) {
	tests := []struct {
		name    string
		comment string
		want    string
	}{
		{"@default numeric", "Max retries. @default 5", "5"},
		{"@default string", "Name. @default Untitled", "Untitled"},
		{"@default with other annotations", "Count. @default 100 @example 42", "100"},
		{"@default quoted", `URL. @default "https://localhost"`, `"https://localhost"`},
		{"no default", "The webhook URL.", ""},
		{"empty", "", ""},
		{"legacy Default: not matched", "Default: 50", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractDefault(tt.comment)
			if got != tt.want {
				t.Errorf("extractDefault(%q) = %q, want %q", tt.comment, got, tt.want)
			}
		})
	}
}

func TestExtractRange(t *testing.T) {
	tests := []struct {
		name    string
		comment string
		wantMin string
		wantMax string
	}{
		{"@range basic", "Count. @range 1-200", "1", "200"},
		{"@range with spaces", "@range 0 - 100", "0", "100"},
		{"@range with trailing period", "Size. @range 1-50.", "1", "50."},
		{"no range", "The webhook URL.", "", ""},
		{"empty", "", "", ""},
		{"legacy Range: not matched", "Range: 1-100.", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMin, gotMax := extractRange(tt.comment)
			if gotMin != tt.wantMin || gotMax != tt.wantMax {
				t.Errorf("extractRange(%q) = (%q, %q), want (%q, %q)",
					tt.comment, gotMin, gotMax, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestExtractExample(t *testing.T) {
	tests := []struct {
		name    string
		comment string
		want    any
	}{
		{"string example", `The URL. @example "https://example.com"`, "https://example.com"},
		{"numeric example", "Max retries. @example 5", float64(5)},
		{"bool example", "Active flag. @example true", true},
		{"json array", `Tags. @example ["a", "b"]`, []any{"a", "b"}},
		{"no example", "The webhook URL.", nil},
		{"empty", "", nil},
		{"bare string", "The ID. @example item-123", "item-123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractExample(tt.comment)
			// Compare using fmt.Sprintf for deep equality since any types
			if fmt.Sprintf("%v", got) != fmt.Sprintf("%v", tt.want) {
				t.Errorf("extractExample(%q) = %v (%T), want %v (%T)",
					tt.comment, got, got, tt.want, tt.want)
			}
		})
	}
}

func TestSplitRPCComment(t *testing.T) {
	tests := []struct {
		name       string
		comment    string
		wantDesc   string
		wantErrors []ProtoError
	}{
		{
			name:       "no errors",
			comment:    "Create a new webhook subscription.",
			wantDesc:   "Create a new webhook subscription.",
			wantErrors: nil,
		},
		{
			name:       "empty",
			comment:    "",
			wantDesc:   "",
			wantErrors: nil,
		},
		{
			name:     "@error annotations",
			comment:  "Update an item. @error NOT_FOUND if the item does not exist. @error INVALID_ARGUMENT if the name is empty.",
			wantDesc: "Update an item.",
			wantErrors: []ProtoError{
				{Code: "NOT_FOUND", Description: "The item does not exist."},
				{Code: "INVALID_ARGUMENT", Description: "The name is empty."},
			},
		},
		{
			name:     "single @error",
			comment:  "Create an item. @error ALREADY_EXISTS if the name is taken.",
			wantDesc: "Create an item.",
			wantErrors: []ProtoError{
				{Code: "ALREADY_EXISTS", Description: "The name is taken."},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDesc, gotErrors := splitRPCComment(tt.comment)
			if gotDesc != tt.wantDesc {
				t.Errorf("splitRPCComment(%q) desc = %q, want %q", tt.comment, gotDesc, tt.wantDesc)
			}
			if len(gotErrors) != len(tt.wantErrors) {
				t.Errorf("splitRPCComment(%q) errors count = %d, want %d",
					tt.comment, len(gotErrors), len(tt.wantErrors))
				return
			}
			for i, e := range gotErrors {
				if e.Code != tt.wantErrors[i].Code || e.Description != tt.wantErrors[i].Description {
					t.Errorf("splitRPCComment(%q) error[%d] = {%q, %q}, want {%q, %q}",
						tt.comment, i, e.Code, e.Description,
						tt.wantErrors[i].Code, tt.wantErrors[i].Description)
				}
			}
		})
	}
}

func TestCleanErrorDescription(t *testing.T) {
	tests := []struct {
		name string
		desc string
		want string
	}{
		{"normal", "a webhook with the same URL exists", "A webhook with the same URL exists."},
		{"trailing period", "the URL is malformed.", "The URL is malformed."},
		{"already capitalized", "The URL is bad", "The URL is bad."},
		{"empty", "", "."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cleanErrorDescription(tt.desc)
			if got != tt.want {
				t.Errorf("cleanErrorDescription(%q) = %q, want %q", tt.desc, got, tt.want)
			}
		})
	}
}

func TestHasDeprecatedOption(t *testing.T) {
	// This function requires proto.Option which needs the proto library.
	// We test it indirectly through integration tests.
	// Here we just verify it returns false for nil input.
	got := hasDeprecatedOption(nil)
	if got != false {
		t.Errorf("hasDeprecatedOption(nil) = %v, want false", got)
	}
}

func TestProcessExampleBlocks(t *testing.T) {
	tests := []struct {
		name  string
		lines []string
		want  []string
	}{
		{
			name:  "no fenced block",
			lines: []string{"A description.", `@example "hello"`},
			want:  []string{"A description.", `@example "hello"`},
		},
		{
			name:  "fenced JSON block",
			lines: []string{"Schema definition.", "@example ```", `{`, `  "type": "object"`, `}`, "```"},
			want:  []string{"Schema definition.", `@example { "type": "object" }`},
		},
		{
			name:  "fenced block with whitespace in opening",
			lines: []string{"Desc.", "@example  ```", "value", "```"},
			want:  []string{"Desc.", "@example value"},
		},
		{
			name:  "empty fenced block",
			lines: []string{"Desc.", "@example ```", "```"},
			want:  []string{"Desc."},
		},
		{
			name:  "fenced block at start",
			lines: []string{"@example ```", `{"id": "123"}`, "```"},
			want:  []string{`@example {"id": "123"}`},
		},
		{
			name:  "no fenced block at all",
			lines: []string{"A description.", "More text."},
			want:  []string{"A description.", "More text."},
		},
		{
			name:  "unclosed fenced block collects to end",
			lines: []string{"Desc.", "@example ```", "line1", "line2"},
			want:  []string{"Desc.", "@example line1 line2"},
		},
		{
			name:  "multi-line JSON object",
			lines: []string{"Config.", "@example ```", `{`, `  "name": "test",`, `  "count": 5`, `}`, "```"},
			want:  []string{"Config.", `@example { "name": "test", "count": 5 }`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := processExampleBlocks(tt.lines)
			if len(got) != len(tt.want) {
				t.Errorf("processExampleBlocks() = %v (len %d), want %v (len %d)",
					got, len(got), tt.want, len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("processExampleBlocks()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestExtractExample_FencedBlock(t *testing.T) {
	// Test that a fenced @example block, after processing by
	// processExampleBlocks + joinCommentLines, produces a valid example.
	lines := []string{"JSON schema.", "@example ```", `{`, `  "type": "object"`, `}`, "```"}
	processed := processExampleBlocks(lines)
	joined := joinCommentLines(processed)
	got := extractExample(joined)

	// The fenced content should be parsed as JSON
	if got == nil {
		t.Fatal("extractExample returned nil for fenced block")
	}
	m, ok := got.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any, got %T: %v", got, got)
	}
	if m["type"] != "object" {
		t.Errorf("expected type=object, got %v", m["type"])
	}
}
