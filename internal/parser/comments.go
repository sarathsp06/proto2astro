package parser

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/emicklei/proto"
)

var (
	requiredRE   = regexp.MustCompile(`@required\b`)
	deprecatedRE = regexp.MustCompile(`(?i)^@deprecated\b`)
	defaultRE    = regexp.MustCompile(`@default\s+(\S+)`)
	rangeRE      = regexp.MustCompile(`@range\s+(\S+)\s*-\s*(\S+)`)
	exampleRE    = regexp.MustCompile(`@example\s+(.+)`)
	errorRE      = regexp.MustCompile(`@error\s+([A-Z_]+)\s+(?:if\s+)?(.+?)\.?$`)
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

// extractRequired checks if a comment contains the @required annotation.
func extractRequired(comment string) bool {
	return requiredRE.MatchString(comment)
}

// extractDeprecated checks if a comment starts with @deprecated.
func extractDeprecated(comment string) bool {
	trimmed := strings.TrimSpace(comment)
	return deprecatedRE.MatchString(trimmed)
}

// extractDefault extracts an "@default VALUE" pattern from a comment.
func extractDefault(comment string) string {
	m := defaultRE.FindStringSubmatch(comment)
	if m != nil {
		return strings.TrimSpace(m[1])
	}
	return ""
}

// extractRange extracts an "@range MIN-MAX" pattern from a comment.
func extractRange(comment string) (min, max string) {
	m := rangeRE.FindStringSubmatch(comment)
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

// splitRPCComment separates the description from @error annotations.
func splitRPCComment(comment string) (string, []ProtoError) {
	var errors []ProtoError
	var descParts []string

	// Split on @error boundaries.
	parts := splitOnAtError(comment)
	for _, s := range parts {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		m := errorRE.FindStringSubmatch(s)
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

// splitOnAtError splits text on "@error" boundaries.
func splitOnAtError(text string) []string {
	var result []string
	remaining := text
	for {
		idx := strings.Index(remaining, "@error")
		if idx < 0 {
			s := strings.TrimSpace(remaining)
			if s != "" {
				result = append(result, s)
			}
			break
		}
		before := strings.TrimSpace(remaining[:idx])
		if before != "" {
			result = append(result, before)
		}
		remaining = remaining[idx:]
		// Find next @error after current one
		nextIdx := strings.Index(remaining[1:], "@error")
		if nextIdx < 0 {
			result = append(result, strings.TrimSpace(remaining))
			break
		}
		result = append(result, strings.TrimSpace(remaining[:nextIdx+1]))
		remaining = remaining[nextIdx+1:]
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
