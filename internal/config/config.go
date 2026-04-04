package config

import (
	"fmt"
	"os"

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
