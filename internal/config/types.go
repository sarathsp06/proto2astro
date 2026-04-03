// Package config provides configuration loading for proto2docs.
package config

// Config is the top-level proto2docs.yaml configuration.
type Config struct {
	// Project metadata
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Site        string `yaml:"site"` // e.g. "https://example.github.io"
	Base        string `yaml:"base"` // e.g. "/my-api"
	Logo        string `yaml:"logo"` // optional logo path
	Social      []Link `yaml:"social"`
	EditLink    string `yaml:"edit_link"` // base URL for edit links

	// Proto input
	Proto ProtoInput `yaml:"proto"`

	// Output
	OutDir string `yaml:"out_dir"` // default: "./docs"

	// Overlay
	ServiceOrder []string                  `yaml:"service_order"`
	EntityTypes  []string                  `yaml:"entity_types"`
	Services     map[string]OverlayService `yaml:"services"`
	CustomPages  []CustomPage              `yaml:"custom_pages"`
}

// Link represents a social/external link.
type Link struct {
	Icon  string `yaml:"icon"`
	Label string `yaml:"label"`
	Href  string `yaml:"href"`
}

// ProtoInput defines where to find proto files.
type ProtoInput struct {
	// Direct file or directory paths
	Paths []string `yaml:"paths"`
	// Buf workspace root (alternative to paths)
	BufWorkspace string `yaml:"buf_workspace"`
	// Buf modules to include (if using buf workspace)
	BufModules []string `yaml:"buf_modules"`
}

// OverlayService provides per-service overrides.
type OverlayService struct {
	Description string                `yaml:"description"`
	Notes       string                `yaml:"notes"`
	Footer      string                `yaml:"footer"`
	RPCs        map[string]OverlayRPC `yaml:"rpcs"`
}

// OverlayRPC provides per-RPC overrides.
type OverlayRPC struct {
	Description string                  `yaml:"description"`
	Fields      map[string]OverlayField `yaml:"fields"`
}

// OverlayField provides per-field overrides.
type OverlayField struct {
	Example     any    `yaml:"example"`
	Description string `yaml:"description"`
	Required    *bool  `yaml:"required"`
}

// CustomPage defines an additional page to add to the sidebar.
type CustomPage struct {
	Title   string `yaml:"title"`
	Slug    string `yaml:"slug"`
	Content string `yaml:"content"` // markdown content
}
