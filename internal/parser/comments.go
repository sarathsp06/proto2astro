package parser

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/emicklei/proto"
)

var (
	errorLineRE  = regexp.MustCompile(`^Errors?:\s+([A-Z_]+)\s+(?:if\s+)?(.+?)\.?$`)
	requiredRE   = regexp.MustCompile(`\bRequired\b`)
	deprecatedRE = regexp.MustCompile(`(?i)^Deprecated:`)
	defaultRE    = regexp.MustCompile(`Default:\s*(.+?)(?:\.|$)`)
	rangeRE      = regexp.MustCompile(`Range:\s*(\S+)\s*-\s*(\S+)`)
	exampleRE    = regexp.MustCompile(`@example\s+(.+)`)
)

// extractComment joins leading comment lines into a single string.
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

// extractRequired checks if a comment contains the Required keyword.
func extractRequired(comment string) bool {
	return requiredRE.MatchString(comment)
}

// extractDeprecated checks if a comment starts with Deprecated:.
func extractDeprecated(comment string) bool {
	return deprecatedRE.MatchString(strings.TrimSpace(comment))
}

// extractDefault extracts a "Default: VALUE" pattern from a comment.
func extractDefault(comment string) string {
	m := defaultRE.FindStringSubmatch(comment)
	if m != nil {
		return strings.TrimSpace(m[1])
	}
	return ""
}

// extractRange extracts a "Range: MIN-MAX" pattern from a comment.
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

// splitRPCComment separates the description from error code patterns.
func splitRPCComment(comment string) (string, []ProtoError) {
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

// cleanFieldDescription strips "Required. " prefix and annotation patterns.
func cleanFieldDescription(desc string) string {
	desc = strings.TrimSpace(desc)
	desc = strings.TrimPrefix(desc, "Required. ")
	desc = strings.TrimPrefix(desc, "Required ")

	// Remove @example annotations from the visible description
	desc = exampleRE.ReplaceAllString(desc, "")

	// Remove Default: and Range: patterns from visible description
	// (they're extracted separately)
	desc = defaultRE.ReplaceAllString(desc, "")
	desc = rangeRE.ReplaceAllString(desc, "")

	return strings.TrimSpace(desc)
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
