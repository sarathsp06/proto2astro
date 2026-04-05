package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/sarathsp06/proto2astro/internal/config"
	"github.com/sarathsp06/proto2astro/internal/parser"
)

// Regex patterns for annotation stripping in cleanDescription.
var (
	cleanDefaultRE = regexp.MustCompile(`Default:\s*\S+\.?`)
	cleanRangeRE   = regexp.MustCompile(`Range:\s*\S+\s*-\s*\S+`)
)

// generateDataFiles generates TypeScript data files (one per service + one per enum) in outDir/src/data/api/.
func generateDataFiles(result *parser.ParseResult, cfg *config.Config, outDir string) error {
	dataDir := filepath.Join(outDir, "src", "data", "api")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return fmt.Errorf("mkdir data dir: %w", err)
	}

	for _, pkg := range result.Packages {
		for _, svc := range pkg.Services {
			overlay := cfg.Services[svc.Name]
			ts := buildTSService(svc, pkg, cfg, overlay)
			content, err := marshalServiceTS(ts)
			if err != nil {
				return err
			}
			filename := toKebab(svc.Name) + ".ts"
			path := filepath.Join(dataDir, filename)
			if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
				return fmt.Errorf("write %s: %w", path, err)
			}
		}

		// Generate enum data files
		for _, enum := range pkg.Enums {
			ts := buildTSEnum(enum, pkg)
			content, err := marshalEnumTS(ts)
			if err != nil {
				return err
			}
			filename := "enum-" + toKebab(enum.Name) + ".ts"
			path := filepath.Join(dataDir, filename)
			if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
				return fmt.Errorf("write %s: %w", path, err)
			}
		}
	}
	return nil
}

// buildTSService converts parsed proto data + overlays into a TSService struct.
func buildTSService(svc *parser.ProtoService, pkg *parser.ProtoPackage, cfg *config.Config, overlay config.OverlayService) TSService {
	desc := svc.Description
	if overlay.Description != "" {
		desc = overlay.Description
	}

	ts := TSService{
		Service:     svc.Name,
		Package:     pkg.Name,
		Description: desc,
		Notes:       overlay.Notes,
		Footer:      overlay.Footer,
	}

	for _, rpc := range svc.RPCs {
		ts.RPCs = append(ts.RPCs, buildTSRPC(rpc, pkg, cfg, overlay))
	}

	return ts
}

// buildTSRPC converts a parsed RPC + overlays into a TSRPC struct.
func buildTSRPC(rpc parser.ProtoRPC, pkg *parser.ProtoPackage, cfg *config.Config, overlay config.OverlayService) TSRPC {
	rpcOverlay := overlay.RPCs[rpc.Name]

	rpcDesc := rpc.Description
	if rpcOverlay.Description != "" {
		rpcDesc = rpcOverlay.Description
	}

	ts := TSRPC{
		Name:        rpc.Name,
		Description: rpcDesc,
		Request:     []TSField{}, // ensure non-nil for JSON serialization
	}

	// Request fields
	if reqMsg := pkg.Messages[rpc.RequestType]; reqMsg != nil {
		reqFields := flattenFields(reqMsg, pkg, cfg.EntityTypes, "", rpcOverlay.Fields)
		for _, f := range reqFields {
			ts.Request = append(ts.Request, toTSField(f))
		}
	}

	// Response fields
	if respMsg := pkg.Messages[rpc.ResponseType]; respMsg != nil {
		respFields := flattenFields(respMsg, pkg, cfg.EntityTypes, "", rpcOverlay.Fields)
		for _, f := range respFields {
			ts.Response = append(ts.Response, toTSField(f))
		}
	}

	// Errors
	for _, e := range rpc.Errors {
		ts.Errors = append(ts.Errors, TSErrorCode{
			Code:        e.Code,
			Description: e.Description,
		})
	}

	return ts
}

// toTSField converts a flatField to a TSField.
func toTSField(f flatField) TSField {
	return TSField(f)
}

// buildTSEnum converts a parsed enum into a TSEnum struct.
func buildTSEnum(enum *parser.ProtoEnum, pkg *parser.ProtoPackage) TSEnum {
	ts := TSEnum{
		Name:        enum.Name,
		Package:     pkg.Name,
		Description: enum.Description,
	}

	for _, v := range enum.Values {
		ts.Values = append(ts.Values, TSEnumValue{
			Name:        v.Name,
			Number:      v.Number,
			Description: v.Description,
		})
	}

	ts.UsedBy = findEnumUsage(enum.Name, pkg)
	return ts
}

// findEnumUsage finds all services that reference a given enum type.
func findEnumUsage(enumName string, pkg *parser.ProtoPackage) []string {
	var result []string
	seen := make(map[string]bool)

	for _, svc := range pkg.Services {
		for _, rpc := range svc.RPCs {
			// Check request message
			if reqMsg, ok := pkg.Messages[rpc.RequestType]; ok {
				if messageUsesEnum(reqMsg, enumName, pkg, nil) && !seen[svc.Name] {
					result = append(result, svc.Name)
					seen[svc.Name] = true
				}
			}
			// Check response message
			if respMsg, ok := pkg.Messages[rpc.ResponseType]; ok {
				if messageUsesEnum(respMsg, enumName, pkg, nil) && !seen[svc.Name] {
					result = append(result, svc.Name)
					seen[svc.Name] = true
				}
			}
		}
	}
	sort.Strings(result)
	return result
}

// messageUsesEnum checks whether a message (or any of its nested message
// fields) references the given enum. The seen map prevents infinite recursion
// from self-referencing or mutually-recursive message types.
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

// flatField is an intermediate representation for a flattened field.
type flatField struct {
	Name        string
	Type        string
	Required    bool
	Description string
	Example     any
}

// flattenFields recursively flattens message fields using dot-notation.
// Entity types (from config) are NOT flattened — they appear as a single field.
func flattenFields(msg *parser.ProtoMessage, pkg *parser.ProtoPackage, entityTypes []string, prefix string, overlayFields map[string]config.OverlayField) []flatField {
	return flattenFieldsWithSeen(msg, pkg, entityTypes, prefix, overlayFields, nil)
}

// flattenFieldsWithSeen is the recursive implementation with cycle detection.
// The seen map tracks message type names currently being expanded to prevent
// infinite recursion from self-referencing or mutually-recursive types.
func flattenFieldsWithSeen(msg *parser.ProtoMessage, pkg *parser.ProtoPackage, entityTypes []string, prefix string, overlayFields map[string]config.OverlayField, seen map[string]bool) []flatField {
	if seen == nil {
		seen = make(map[string]bool)
	}

	var fields []flatField
	for _, f := range msg.Fields {
		if f.Deprecated {
			continue
		}

		fullName := f.Name
		if prefix != "" {
			fullName = prefix + "." + f.Name
		}

		// Apply overlay
		ov := overlayFields[fullName]

		// Should we flatten into this message?
		if f.IsMessage && !f.IsMap && shouldFlatten(f.RawType, entityTypes) {
			// Cycle detection: skip if we're already expanding this type
			if seen[f.RawType] {
				// Emit as a non-flattened reference to break the cycle
				fields = append(fields, buildFlatField(f, fullName, ov))
				continue
			}
			subMsg, ok := pkg.Messages[f.RawType]
			if ok {
				seen[f.RawType] = true
				subPrefix := fullName
				if f.IsRepeated {
					subPrefix = fullName + "[]"
				}
				subFields := flattenFieldsWithSeen(subMsg, pkg, entityTypes, subPrefix, overlayFields, seen)
				fields = append(fields, subFields...)
				delete(seen, f.RawType) // allow this type in other branches
				continue
			}
		}

		fields = append(fields, buildFlatField(f, fullName, ov))
	}
	return fields
}

// buildFlatField constructs a flatField from a parser field and its overlay.
func buildFlatField(f parser.ProtoField, fullName string, ov config.OverlayField) flatField {
	desc := f.Description
	if ov.Description != "" {
		desc = ov.Description
	}
	desc = cleanDescription(desc)

	required := f.Required
	if ov.Required != nil {
		required = *ov.Required
	}

	example := f.Example
	if ov.Example != nil {
		example = ov.Example
	}

	typeName := f.Type
	if f.IsEnum {
		typeName = f.RawType
	}

	return flatField{
		Name:        fullName,
		Type:        typeName,
		Required:    required,
		Description: desc,
		Example:     example,
	}
}

// shouldFlatten returns true if a type should be flattened (not in entity types list).
func shouldFlatten(typeName string, entityTypes []string) bool {
	for _, et := range entityTypes {
		if et == typeName {
			return false
		}
	}
	return true
}

// cleanDescription strips annotation patterns from the visible description.
// Annotations like @example, Default:, and Range: are extracted as structured
// data elsewhere; leaving them in the description creates redundant noise.
func cleanDescription(desc string) string {
	desc = strings.TrimSpace(desc)
	desc = strings.TrimPrefix(desc, "Required. ")
	desc = strings.TrimPrefix(desc, "Required ")
	// Strip @example and everything after it (usually last in comment)
	if idx := strings.Index(desc, "@example"); idx >= 0 {
		desc = desc[:idx]
	}
	// Strip Default: VALUE. and Range: MIN-MAX patterns
	desc = cleanDefaultRE.ReplaceAllString(desc, "")
	desc = cleanRangeRE.ReplaceAllString(desc, "")
	return strings.TrimSpace(desc)
}

// toKebab converts a PascalCase name to kebab-case.
func toKebab(name string) string {
	var result []rune
	for i, r := range name {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '-')
		}
		result = append(result, r)
	}
	return strings.ToLower(string(result))
}

// sortedServiceNames returns service names in the order specified by config,
// with remaining services sorted alphabetically.
func sortedServiceNames(pkg *parser.ProtoPackage, serviceOrder []string) []string {
	seen := make(map[string]bool)
	var result []string

	// Add ordered services first
	for _, name := range serviceOrder {
		if _, ok := pkg.Services[name]; ok {
			result = append(result, name)
			seen[name] = true
		}
	}

	// Add remaining services alphabetically
	var remaining []string
	for name := range pkg.Services {
		if !seen[name] {
			remaining = append(remaining, name)
		}
	}
	sort.Strings(remaining)
	result = append(result, remaining...)

	return result
}
