// Package parser provides proto file parsing and type resolution.
package parser

// ProtoField represents a single field in a proto message.
type ProtoField struct {
	Name        string
	Type        string // mapped type: "string", "int32", "Timestamp", etc.
	RawType     string // original proto type (for message resolution)
	Description string
	Required    bool
	Deprecated  bool
	IsRepeated  bool
	IsMap       bool
	MapKey      string
	MapValue    string
	IsOptional  bool
	IsOneof     bool   // whether this field is part of a oneof group
	OneofGroup  string // name of the containing oneof group
	IsMessage   bool   // whether the type references a message
	IsEnum      bool   // whether the type references an enum
	// Extended comment annotations
	DefaultValue string // from "@default VALUE" annotation
	RangeMin     string // from "@range MIN-MAX" annotation
	RangeMax     string
	Example      any // from "@example VALUE" annotation
}

// ProtoError represents an error code extracted from RPC comments.
type ProtoError struct {
	Code        string
	Description string
}

// ProtoRPC represents an RPC method in a service.
type ProtoRPC struct {
	Name            string
	Description     string
	RequestType     string
	ResponseType    string
	StreamsRequest  bool // client-streaming or bidi
	StreamsResponse bool // server-streaming or bidi
	Errors          []ProtoError
}

// ProtoService represents a gRPC/ConnectRPC service.
type ProtoService struct {
	Name        string
	Description string
	RPCs        []ProtoRPC
}

// ProtoMessage represents a proto message definition.
type ProtoMessage struct {
	Name        string
	Description string
	Fields      []ProtoField
}

// ProtoEnumValue represents a single value in an enum.
type ProtoEnumValue struct {
	Name        string
	Number      int
	Description string
}

// ProtoEnum represents a proto enum definition.
type ProtoEnum struct {
	Name        string
	Description string
	Values      []ProtoEnumValue
}

// ProtoFile represents the parsed contents of one or more .proto files.
type ProtoFile struct {
	Package  string
	Services map[string]*ProtoService
	Messages map[string]*ProtoMessage
	Enums    map[string]*ProtoEnum
}

// ProtoPackage groups parsed proto files by package name.
type ProtoPackage struct {
	Name     string
	Files    []*ProtoFile
	Services map[string]*ProtoService
	Messages map[string]*ProtoMessage
	Enums    map[string]*ProtoEnum
}

// ParseResult is the top-level result from parsing one or more proto files.
type ParseResult struct {
	Packages map[string]*ProtoPackage
}
