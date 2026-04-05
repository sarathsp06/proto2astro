// Package generator produces a complete Astro Starlight documentation site
// from parsed proto definitions and a proto2astro.yaml configuration.
package generator

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// TSField mirrors the TypeScript Field interface in types.ts.
type TSField struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required,omitempty"`
	Description string `json:"description"`
	Example     any    `json:"example,omitempty"`
}

// TSErrorCode mirrors the TypeScript ErrorCode interface.
type TSErrorCode struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

// TSRPC mirrors the TypeScript Rpc interface.
type TSRPC struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Request     []TSField     `json:"request"`
	Response    []TSField     `json:"response,omitempty"`
	Errors      []TSErrorCode `json:"errors,omitempty"`
}

// TSService mirrors the TypeScript ApiService interface.
type TSService struct {
	Service     string  `json:"service"`
	Package     string  `json:"package"`
	Description string  `json:"description"`
	Notes       string  `json:"notes,omitempty"`
	RPCs        []TSRPC `json:"rpcs"`
	Footer      string  `json:"footer,omitempty"`
}

// TSEnumValue mirrors the TypeScript EnumValue interface.
type TSEnumValue struct {
	Name        string `json:"name"`
	Number      int    `json:"number"`
	Description string `json:"description"`
}

// TSEnum mirrors the TypeScript ApiEnum interface.
type TSEnum struct {
	Name        string        `json:"name"`
	Package     string        `json:"package"`
	Description string        `json:"description"`
	Values      []TSEnumValue `json:"values"`
	UsedBy      []string      `json:"usedBy,omitempty"`
}

// marshalServiceTS marshals a TSService to a TypeScript source file.
func marshalServiceTS(svc TSService) (string, error) {
	data, err := marshalJSON(svc)
	if err != nil {
		return "", fmt.Errorf("marshal service %s: %w", svc.Service, err)
	}
	return fmt.Sprintf(
		"import type { ApiService } from \"./types\";\n\nconst service: ApiService = %s;\n\nexport default service;\n",
		string(data),
	), nil
}

// marshalEnumTS marshals a TSEnum to a TypeScript source file.
func marshalEnumTS(enum TSEnum) (string, error) {
	data, err := marshalJSON(enum)
	if err != nil {
		return "", fmt.Errorf("marshal enum %s: %w", enum.Name, err)
	}
	return fmt.Sprintf(
		"import type { ApiEnum } from \"./types\";\n\nconst enumData: ApiEnum = %s;\n\nexport default enumData;\n",
		string(data),
	), nil
}

// marshalJSON encodes a value as indented JSON without HTML escaping.
// Go's default json.Marshal escapes <, >, and & as \u003c, \u003e, \u0026
// which corrupts proto descriptions containing comparison operators or ampersands.
func marshalJSON(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	// Encode appends a trailing newline; trim it so callers control formatting.
	return bytes.TrimRight(buf.Bytes(), "\n"), nil
}
