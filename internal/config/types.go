// Package config provides configuration loading for proto2astro.
package config

// Config is the top-level proto2astro.yaml configuration.
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

	// Sidebar customization
	Sidebar SidebarConfig `yaml:"sidebar"`

	// Starlight component overrides (e.g. Footer: "./src/components/Footer.astro")
	Components map[string]string `yaml:"components"`

	// Additional CSS files beyond the default custom.css
	CustomCSS []string `yaml:"custom_css"`

	// Scaffold options
	Scaffold ScaffoldConfig `yaml:"scaffold"`

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

// CustomPage defines an additional page to add to the docs.
type CustomPage struct {
	Title   string `yaml:"title"`
	Slug    string `yaml:"slug"`
	Path    string `yaml:"path"`    // full content path (e.g. "deployment/kubernetes"); overrides slug
	Content string `yaml:"content"` // markdown content
}

// SidebarConfig defines custom sidebar sections rendered around the auto-generated API Reference.
type SidebarConfig struct {
	// Sections rendered before the auto-generated API Reference group
	Before []SidebarSection `yaml:"before"`
	// Sections rendered after the auto-generated API Reference group
	After []SidebarSection `yaml:"after"`
}

// SidebarSection is a named sidebar group with ordered items.
type SidebarSection struct {
	Label string        `yaml:"label"`
	Items []SidebarItem `yaml:"items"`
}

// SidebarItem is a single page entry in a sidebar section.
type SidebarItem struct {
	Label string `yaml:"label"` // optional display label
	Slug  string `yaml:"slug"`  // content slug (e.g. "getting-started/installation")
}

// ScaffoldConfig controls which scaffold-only files are generated.
type ScaffoldConfig struct {
	// LandingPage controls the root index.mdx landing page.
	// Default: true (generated on first run, never overwritten).
	LandingPage *bool `yaml:"landing_page"`
	// CommentGuide controls the guides/comment-guide.md page.
	// Default: true (generated on first run, never overwritten).
	CommentGuide *bool `yaml:"comment_guide"`
}

// LandingPageEnabled returns whether the landing page scaffold should be generated.
func (s ScaffoldConfig) LandingPageEnabled() bool {
	return s.LandingPage == nil || *s.LandingPage
}

// CommentGuideEnabled returns whether the comment guide scaffold should be generated.
func (s ScaffoldConfig) CommentGuideEnabled() bool {
	return s.CommentGuide == nil || *s.CommentGuide
}
