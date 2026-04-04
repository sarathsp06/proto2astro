package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/sarathsp06/proto2astro/internal/config"
	"github.com/sarathsp06/proto2astro/internal/parser"
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
	return TSField{
		Name:        f.Name,
		Type:        f.Type,
		Required:    f.Required,
		Description: f.Description,
		Example:     f.Example,
	}
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
				if messageUsesEnum(reqMsg, enumName) && !seen[svc.Name] {
					result = append(result, svc.Name)
					seen[svc.Name] = true
				}
			}
			// Check response message
			if respMsg, ok := pkg.Messages[rpc.ResponseType]; ok {
				if messageUsesEnum(respMsg, enumName) && !seen[svc.Name] {
					result = append(result, svc.Name)
					seen[svc.Name] = true
				}
			}
		}
	}
	sort.Strings(result)
	return result
}

// messageUsesEnum checks whether a message has any field referencing the given enum.
func messageUsesEnum(msg *parser.ProtoMessage, enumName string) bool {
	for _, f := range msg.Fields {
		if f.IsEnum && f.RawType == enumName {
			return true
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
			subMsg, ok := pkg.Messages[f.RawType]
			if ok {
				subPrefix := fullName
				if f.IsRepeated {
					subPrefix = fullName + "[]"
				}
				subFields := flattenFields(subMsg, pkg, entityTypes, subPrefix, overlayFields)
				fields = append(fields, subFields...)
				continue
			}
		}

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

		fields = append(fields, flatField{
			Name:        fullName,
			Type:        typeName,
			Required:    required,
			Description: desc,
			Example:     example,
		})
	}
	return fields
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
func cleanDescription(desc string) string {
	desc = strings.TrimSpace(desc)
	desc = strings.TrimPrefix(desc, "Required. ")
	desc = strings.TrimPrefix(desc, "Required ")
	return desc
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
