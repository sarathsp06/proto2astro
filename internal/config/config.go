package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// DefaultConfigFile is the default filename for proto2astro configuration.
const DefaultConfigFile = "proto2astro.yaml"

// Load reads and parses a proto2astro.yaml configuration file.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	cfg.ApplyDefaults()
	return &cfg, nil
}

// ApplyDefaults fills in zero-value fields with sensible defaults.
func (c *Config) ApplyDefaults() {
	if c.Title == "" {
		c.Title = "API Documentation"
	}
	if c.Description == "" {
		c.Description = "API reference documentation"
	}
	if c.OutDir == "" {
		c.OutDir = "./docs"
	}
}

// Validate checks the configuration for common issues and returns warnings.
// It returns a list of non-fatal warnings and a fatal error if the config
// is unusable.
func (c *Config) Validate() (warnings []string, err error) {
	// Fatal: must have proto input
	if len(c.Proto.Paths) == 0 && c.Proto.BufWorkspace == "" {
		err = fmt.Errorf("no proto input configured: set proto.paths or proto.buf_workspace in %s", DefaultConfigFile)
		return
	}

	// Validate proto paths exist
	for _, p := range c.Proto.Paths {
		if _, statErr := os.Stat(p); statErr != nil {
			warnings = append(warnings, fmt.Sprintf("proto path does not exist: %s", p))
		}
	}

	// Validate buf workspace exists
	if c.Proto.BufWorkspace != "" {
		if _, statErr := os.Stat(c.Proto.BufWorkspace); statErr != nil {
			warnings = append(warnings, fmt.Sprintf("buf workspace does not exist: %s", c.Proto.BufWorkspace))
		}
	}

	// Warn about empty title
	if c.Title == "API Documentation" {
		warnings = append(warnings, "using default title \"API Documentation\" — consider setting title in config")
	}

	// Validate service_order references
	if len(c.ServiceOrder) > 0 && len(c.Services) > 0 {
		for _, name := range c.ServiceOrder {
			if _, ok := c.Services[name]; !ok {
				warnings = append(warnings, fmt.Sprintf("service_order references unknown service: %s", name))
			}
		}
	}

	// Validate custom pages
	for i, cp := range c.CustomPages {
		if cp.Slug == "" {
			warnings = append(warnings, fmt.Sprintf("custom_pages[%d] is missing a slug", i))
		}
		if cp.Title == "" {
			warnings = append(warnings, fmt.Sprintf("custom_pages[%d] is missing a title", i))
		}
	}

	// Validate site URL format
	if c.Site != "" && !strings.HasPrefix(c.Site, "http://") && !strings.HasPrefix(c.Site, "https://") {
		warnings = append(warnings, fmt.Sprintf("site URL should start with http:// or https://: %s", c.Site))
	}

	return
}
