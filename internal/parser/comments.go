package parser

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/emicklei/proto"
)

var (
	// Legacy (colon-suffix) patterns.
	errorLineRE  = regexp.MustCompile(`^Errors?:\s+([A-Z_]+)\s+(?:if\s+)?(.+?)\.?$`)
	requiredRE   = regexp.MustCompile(`\bRequired\b`)
	deprecatedRE = regexp.MustCompile(`(?i)^Deprecated:`)
	defaultRE    = regexp.MustCompile(`Default:\s*(.+?)(?:\.|$)`)
	rangeRE      = regexp.MustCompile(`Range:\s*(\S+)\s*-\s*(\S+)`)
	exampleRE    = regexp.MustCompile(`@example\s+(.+)`)

	// New @-prefix patterns (consistent annotation style).
	atRequiredRE   = regexp.MustCompile(`@required\b`)
	atDeprecatedRE = regexp.MustCompile(`(?i)^@deprecated\b`)
	atDefaultRE    = regexp.MustCompile(`@default\s+(\S+)`)
	atRangeRE      = regexp.MustCompile(`@range\s+(\S+)\s*-\s*(\S+)`)
	atErrorNormRE  = regexp.MustCompile(`@error\s+`)
)

// extractComment joins leading comment lines into a single string.
// Fenced @example blocks (delimited by ```) are collapsed into a single
// @example annotation before joining so that multi-line JSON examples
// are preserved as a single value.
func extractComment(c *proto.Comment) string {
	if c == nil {
		return ""
	}
	var lines []string
	for _, line := range c.Lines {
		line = strings.TrimSpace(line)
		line = strings.TrimPrefix(line, " ")
		lines = append(lines, line)
	}
	lines = processExampleBlocks(lines)
	return joinCommentLines(lines)
}

// extractInlineComment joins inline comment lines.
func extractInlineComment(c *proto.Comment) string {
	if c == nil {
		return ""
	}
	var lines []string
	for _, line := range c.Lines {
		lines = append(lines, strings.TrimSpace(line))
	}
	return strings.Join(lines, " ")
}

// joinCommentLines joins consecutive non-empty lines into paragraphs.
func joinCommentLines(lines []string) string {
	var result []string
	var current []string
	for _, line := range lines {
		if line == "" {
			if len(current) > 0 {
				result = append(result, strings.Join(current, " "))
				current = nil
			}
		} else {
			current = append(current, line)
		}
	}
	if len(current) > 0 {
		result = append(result, strings.Join(current, " "))
	}
	return strings.Join(result, " ")
}

// processExampleBlocks detects fenced @example blocks in comment lines
// and collapses them into a single @example annotation line.
//
// A fenced block starts with a line matching "@example ```" and ends with
// a line that is exactly "```". The lines between the fences are joined
// with spaces to produce a single-line value that extractExample can parse.
//
// Example input:
//
//	["Schema definition.", "@example ```", "{", `  "type": "object"`, "}", "```"]
//
// Output:
//
//	["Schema definition.", `@example { "type": "object" }`]
func processExampleBlocks(lines []string) []string {
	var result []string
	i := 0
	for i < len(lines) {
		line := lines[i]
		// Check for @example ``` (with optional trailing whitespace)
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "@example") {
			rest := strings.TrimSpace(strings.TrimPrefix(trimmed, "@example"))
			if rest == "```" {
				// Fenced block — collect lines until closing ```
				var blockLines []string
				i++
				for i < len(lines) {
					bl := strings.TrimSpace(lines[i])
					if bl == "```" {
						i++
						break
					}
					blockLines = append(blockLines, bl)
					i++
				}
				if len(blockLines) > 0 {
					content := strings.Join(blockLines, " ")
					result = append(result, "@example "+content)
				}
				continue
			}
		}
		result = append(result, line)
		i++
	}
	return result
}

// extractRequired checks if a comment contains the Required keyword or @required annotation.
func extractRequired(comment string) bool {
	return requiredRE.MatchString(comment) || atRequiredRE.MatchString(comment)
}

// extractDeprecated checks if a comment starts with Deprecated: or @deprecated.
func extractDeprecated(comment string) bool {
	trimmed := strings.TrimSpace(comment)
	return deprecatedRE.MatchString(trimmed) || atDeprecatedRE.MatchString(trimmed)
}

// extractDefault extracts a "Default: VALUE" or "@default VALUE" pattern from a comment.
func extractDefault(comment string) string {
	m := defaultRE.FindStringSubmatch(comment)
	if m != nil {
		return strings.TrimSpace(m[1])
	}
	m = atDefaultRE.FindStringSubmatch(comment)
	if m != nil {
		return strings.TrimSpace(m[1])
	}
	return ""
}

// extractRange extracts a "Range: MIN-MAX" or "@range MIN-MAX" pattern from a comment.
func extractRange(comment string) (min, max string) {
	m := rangeRE.FindStringSubmatch(comment)
	if m != nil {
		return strings.TrimSpace(m[1]), strings.TrimSpace(m[2])
	}
	m = atRangeRE.FindStringSubmatch(comment)
	if m != nil {
		return strings.TrimSpace(m[1]), strings.TrimSpace(m[2])
	}
	return "", ""
}

// extractExample extracts an "@example VALUE" annotation from a comment.
// It attempts to JSON-parse the value; if that fails, returns the raw string.
func extractExample(comment string) any {
	m := exampleRE.FindStringSubmatch(comment)
	if m == nil {
		return nil
	}
	raw := strings.TrimSpace(m[1])
	if raw == "" {
		return nil
	}

	// Try JSON parse
	var parsed any
	if err := json.Unmarshal([]byte(raw), &parsed); err == nil {
		return parsed
	}
	return raw
}

// splitRPCComment separates the description from error code patterns.
// Supports both "Errors: CODE desc" and "@error CODE desc" syntax.
func splitRPCComment(comment string) (string, []ProtoError) {
	// Normalize @error to Errors: for unified processing.
	comment = atErrorNormRE.ReplaceAllString(comment, "Errors: ")

	var errors []ProtoError
	var descParts []string

	sentences := splitOnErrors(comment)
	for _, s := range sentences {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		m := errorLineRE.FindStringSubmatch(s)
		if m != nil {
			errors = append(errors, ProtoError{
				Code:        m[1],
				Description: cleanErrorDescription(m[2]),
			})
		} else {
			descParts = append(descParts, s)
		}
	}

	return strings.Join(descParts, " "), errors
}

// splitOnErrors splits text on "Errors:" boundaries.
func splitOnErrors(text string) []string {
	parts := strings.Split(text, "Errors:")
	var result []string
	if len(parts) > 0 {
		desc := strings.TrimSpace(parts[0])
		if desc != "" {
			result = append(result, desc)
		}
		for _, p := range parts[1:] {
			p = strings.TrimSpace(p)
			if p != "" {
				result = append(result, "Errors: "+p)
			}
		}
	}
	return result
}

// cleanErrorDescription normalizes an error description.
func cleanErrorDescription(desc string) string {
	desc = strings.TrimSpace(desc)
	desc = strings.TrimSuffix(desc, ".")
	if len(desc) > 0 {
		desc = strings.ToUpper(desc[:1]) + desc[1:]
	}
	return desc + "."
}

// hasDeprecatedOption checks proto field options for [deprecated = true].
func hasDeprecatedOption(opts []*proto.Option) bool {
	for _, opt := range opts {
		if opt.Name == "deprecated" && opt.Constant.Source == "true" {
			return true
		}
	}
	return false
}
