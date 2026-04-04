package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/emicklei/proto"
)

// ParseFiles parses one or more proto files and returns a combined ParseResult.
func ParseFiles(paths []string) (*ParseResult, error) {
	result := &ParseResult{
		Packages: make(map[string]*ProtoPackage),
	}

	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			return nil, fmt.Errorf("stat %s: %w", p, err)
		}
		if info.IsDir() {
			if err := parseDir(p, result); err != nil {
				return nil, err
			}
		} else {
			if err := parseFile(p, result); err != nil {
				return nil, err
			}
		}
	}

	// Post-process: mark message/enum references on fields (including cross-package)
	resolveAllFieldTypes(result)

	return result, nil
}

// parseDir recursively finds .proto files in a directory.
func parseDir(dir string, result *ParseResult) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".proto") {
			return parseFile(path, result)
		}
		return nil
	})
}

// parseFile parses a single .proto file and merges into the result.
func parseFile(path string, result *ParseResult) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open proto %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	parser := proto.NewParser(f)
	definition, err := parser.Parse()
	if err != nil {
		return fmt.Errorf("parse proto %s: %w", path, err)
	}

	pf := &ProtoFile{
		Services: make(map[string]*ProtoService),
		Messages: make(map[string]*ProtoMessage),
		Enums:    make(map[string]*ProtoEnum),
	}

	// Extract package name
	proto.Walk(definition,
		proto.WithPackage(func(p *proto.Package) {
			pf.Package = p.Name
		}),
	)

	// First pass: collect messages (including nested) and enums
	proto.Walk(definition,
		proto.WithMessage(func(m *proto.Message) {
			msg := parseMessage(m)
			pf.Messages[m.Name] = msg
			// Collect nested messages and enums
			collectNested(m, pf)
		}),
		proto.WithEnum(func(e *proto.Enum) {
			en := parseEnum(e)
			pf.Enums[e.Name] = en
		}),
	)

	// Second pass: collect services and RPCs
	proto.Walk(definition,
		proto.WithService(func(s *proto.Service) {
			svc := &ProtoService{
				Name:        s.Name,
				Description: extractComment(s.Comment),
			}
			for _, el := range s.Elements {
				if rpc, ok := el.(*proto.RPC); ok {
					svc.RPCs = append(svc.RPCs, parseRPC(rpc))
				}
			}
			pf.Services[s.Name] = svc
		}),
	)

	// Merge into package
	pkgName := pf.Package
	if pkgName == "" {
		pkgName = "_default"
	}

	pkg, ok := result.Packages[pkgName]
	if !ok {
		pkg = &ProtoPackage{
			Name:     pkgName,
			Services: make(map[string]*ProtoService),
			Messages: make(map[string]*ProtoMessage),
			Enums:    make(map[string]*ProtoEnum),
		}
		result.Packages[pkgName] = pkg
	}

	pkg.Files = append(pkg.Files, pf)
	for k, v := range pf.Services {
		pkg.Services[k] = v
	}
	for k, v := range pf.Messages {
		pkg.Messages[k] = v
	}
	for k, v := range pf.Enums {
		pkg.Enums[k] = v
	}

	return nil
}

// parseMessage extracts a ProtoMessage from a proto.Message.
func parseMessage(m *proto.Message) *ProtoMessage {
	msg := &ProtoMessage{
		Name:        m.Name,
		Description: extractComment(m.Comment),
	}
	for _, e := range m.Elements {
		switch el := e.(type) {
		case *proto.NormalField:
			msg.Fields = append(msg.Fields, parseNormalField(el))
		case *proto.MapField:
			msg.Fields = append(msg.Fields, parseMapField(el))
		case *proto.Oneof:
			for _, oe := range el.Elements {
				if of, ok := oe.(*proto.OneOfField); ok {
					f := parseOneofField(of, el.Name)
					msg.Fields = append(msg.Fields, f)
				}
			}
		}
	}
	return msg
}

// parseEnum extracts a ProtoEnum from a proto.Enum.
func parseEnum(e *proto.Enum) *ProtoEnum {
	en := &ProtoEnum{
		Name:        e.Name,
		Description: extractComment(e.Comment),
	}
	for _, el := range e.Elements {
		if ev, ok := el.(*proto.EnumField); ok {
			desc := extractComment(ev.Comment)
			if desc == "" {
				desc = extractInlineComment(ev.InlineComment)
			}
			en.Values = append(en.Values, ProtoEnumValue{
				Name:        ev.Name,
				Number:      ev.Integer,
				Description: desc,
			})
		}
	}
	return en
}

// parseNormalField extracts a ProtoField from a normal proto field.
func parseNormalField(f *proto.NormalField) ProtoField {
	field := ProtoField{
		Name:       f.Name,
		RawType:    f.Type,
		IsRepeated: f.Repeated,
		IsOptional: f.Optional,
	}
	field.Description = extractComment(f.Comment)
	if field.Description == "" {
		field.Description = extractInlineComment(f.InlineComment)
	}
	field.Required = extractRequired(field.Description)
	field.Deprecated = hasDeprecatedOption(f.Options) || extractDeprecated(field.Description)
	field.DefaultValue = extractDefault(field.Description)
	field.RangeMin, field.RangeMax = extractRange(field.Description)
	field.Example = extractExample(field.Description)
	field.Type = mapProtoType(f.Type, f.Repeated, false, "", "")
	return field
}

// parseMapField extracts a ProtoField from a map proto field.
func parseMapField(f *proto.MapField) ProtoField {
	field := ProtoField{
		Name:     f.Name,
		RawType:  f.Type,
		IsMap:    true,
		MapKey:   f.KeyType,
		MapValue: f.Type,
	}
	field.Description = extractComment(f.Comment)
	if field.Description == "" {
		field.Description = extractInlineComment(f.InlineComment)
	}
	field.Required = extractRequired(field.Description)
	field.Deprecated = hasDeprecatedOption(f.Options) || extractDeprecated(field.Description)
	field.DefaultValue = extractDefault(field.Description)
	field.RangeMin, field.RangeMax = extractRange(field.Description)
	field.Example = extractExample(field.Description)
	field.Type = mapProtoType(f.Type, false, true, f.KeyType, f.Type)
	return field
}

// parseRPC extracts a ProtoRPC from a proto.RPC.
func parseRPC(rpc *proto.RPC) ProtoRPC {
	fullComment := extractComment(rpc.Comment)
	desc, errors := splitRPCComment(fullComment)
	return ProtoRPC{
		Name:            rpc.Name,
		Description:     desc,
		RequestType:     rpc.RequestType,
		ResponseType:    rpc.ReturnsType,
		StreamsRequest:  rpc.StreamsRequest,
		StreamsResponse: rpc.StreamsReturns,
		Errors:          errors,
	}
}

// parseOneofField extracts a ProtoField from a oneof field.
func parseOneofField(f *proto.OneOfField, groupName string) ProtoField {
	field := ProtoField{
		Name:       f.Name,
		RawType:    f.Type,
		IsOneof:    true,
		OneofGroup: groupName,
		IsOptional: true, // oneof fields are inherently optional
	}
	field.Description = extractComment(f.Comment)
	if field.Description == "" {
		field.Description = extractInlineComment(f.InlineComment)
	}
	field.Required = extractRequired(field.Description)
	field.Deprecated = extractDeprecated(field.Description)
	field.DefaultValue = extractDefault(field.Description)
	field.RangeMin, field.RangeMax = extractRange(field.Description)
	field.Example = extractExample(field.Description)
	field.Type = mapProtoType(f.Type, false, false, "", "")
	return field
}

// collectNested recursively collects nested message and enum definitions
// from within a message. Nested types are stored with their parent-qualified
// name (e.g. "Outer.Inner") to match how proto field types reference them.
func collectNested(m *proto.Message, pf *ProtoFile) {
	for _, el := range m.Elements {
		switch e := el.(type) {
		case *proto.Message:
			qualifiedName := m.Name + "." + e.Name
			msg := parseMessage(e)
			msg.Name = qualifiedName
			pf.Messages[qualifiedName] = msg
			// Also store with short name for unqualified references
			if _, exists := pf.Messages[e.Name]; !exists {
				pf.Messages[e.Name] = msg
			}
			// Recurse into deeper nesting
			collectNestedQualified(e, qualifiedName, pf)
		case *proto.Enum:
			qualifiedName := m.Name + "." + e.Name
			en := parseEnum(e)
			en.Name = qualifiedName
			pf.Enums[qualifiedName] = en
			if _, exists := pf.Enums[e.Name]; !exists {
				pf.Enums[e.Name] = en
			}
		}
	}
}

// collectNestedQualified is the recursive helper for deeply nested types.
func collectNestedQualified(m *proto.Message, parentQualified string, pf *ProtoFile) {
	for _, el := range m.Elements {
		switch e := el.(type) {
		case *proto.Message:
			qualifiedName := parentQualified + "." + e.Name
			msg := parseMessage(e)
			msg.Name = qualifiedName
			pf.Messages[qualifiedName] = msg
			if _, exists := pf.Messages[e.Name]; !exists {
				pf.Messages[e.Name] = msg
			}
			collectNestedQualified(e, qualifiedName, pf)
		case *proto.Enum:
			qualifiedName := parentQualified + "." + e.Name
			en := parseEnum(e)
			en.Name = qualifiedName
			pf.Enums[qualifiedName] = en
			if _, exists := pf.Enums[e.Name]; !exists {
				pf.Enums[e.Name] = en
			}
		}
	}
}

// mapProtoType maps proto types to human-readable names.
func mapProtoType(typeName string, repeated, isMap bool, mapKey, mapValue string) string {
	if isMap {
		k := mapScalarType(mapKey)
		v := mapScalarType(mapValue)
		return fmt.Sprintf("map<%s, %s>", k, v)
	}
	base := mapScalarType(typeName)
	if repeated {
		return base + "[]"
	}
	return base
}

// mapScalarType maps scalar proto types to human-readable names.
func mapScalarType(t string) string {
	switch t {
	case "string":
		return "string"
	case "int32", "sint32", "sfixed32":
		return "int32"
	case "int64", "sint64", "sfixed64":
		return "int64"
	case "uint32", "fixed32":
		return "uint32"
	case "uint64", "fixed64":
		return "uint64"
	case "bool":
		return "bool"
	case "double", "float":
		return "double"
	case "bytes":
		return "bytes"
	case "google.protobuf.Timestamp":
		return "Timestamp"
	case "google.protobuf.Duration":
		return "Duration"
	case "google.protobuf.Struct":
		return "Struct (JSON)"
	case "google.protobuf.Value":
		return "Value (JSON)"
	case "google.protobuf.Any":
		return "Any"
	case "google.protobuf.Empty":
		return "Empty"
	case "google.protobuf.StringValue":
		return "string?"
	case "google.protobuf.BoolValue":
		return "bool?"
	case "google.protobuf.Int32Value":
		return "int32?"
	case "google.protobuf.Int64Value":
		return "int64?"
	case "google.protobuf.DoubleValue":
		return "double?"
	default:
		return t
	}
}
