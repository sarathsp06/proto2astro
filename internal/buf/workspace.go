// Package buf provides integration with Buf workspaces for discovering proto files.
package buf

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// bufWorkConfig represents a buf.work.yaml file.
type bufWorkConfig struct {
	Version     string   `yaml:"version"`
	Directories []string `yaml:"directories"`
}

// bufConfig represents a buf.yaml file (v1 or v2).
type bufConfig struct {
	Version string `yaml:"version"`
	Modules []struct {
		Path string `yaml:"path"`
	} `yaml:"modules"`
}

// DiscoverProtoFiles finds .proto files from a Buf workspace or module root.
// If modules is non-empty, only those modules are included.
func DiscoverProtoFiles(workspaceRoot string, modules []string) ([]string, error) {
	// Check for buf.work.yaml (workspace mode)
	workPath := filepath.Join(workspaceRoot, "buf.work.yaml")
	if _, err := os.Stat(workPath); err == nil {
		return discoverFromWorkspace(workspaceRoot, workPath, modules)
	}

	// Check for buf.yaml (single module mode, v2 with modules)
	bufYamlPath := filepath.Join(workspaceRoot, "buf.yaml")
	if _, err := os.Stat(bufYamlPath); err == nil {
		return discoverFromBufYaml(workspaceRoot, bufYamlPath, modules)
	}

	// Fallback: treat the directory as a proto root
	return discoverProtoDir(workspaceRoot)
}

// discoverFromWorkspace reads buf.work.yaml and collects proto files from listed directories.
func discoverFromWorkspace(root, workPath string, filterModules []string) ([]string, error) {
	data, err := os.ReadFile(workPath)
	if err != nil {
		return nil, fmt.Errorf("read buf.work.yaml: %w", err)
	}

	var cfg bufWorkConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse buf.work.yaml: %w", err)
	}

	dirs := cfg.Directories
	if len(filterModules) > 0 {
		dirs = filterDirs(dirs, filterModules)
	}

	var allFiles []string
	for _, dir := range dirs {
		full := filepath.Join(root, dir)
		files, err := discoverProtoDir(full)
		if err != nil {
			return nil, err
		}
		allFiles = append(allFiles, files...)
	}
	return allFiles, nil
}

// discoverFromBufYaml reads buf.yaml (v2 format) and collects proto files from modules.
func discoverFromBufYaml(root, bufYamlPath string, filterModules []string) ([]string, error) {
	data, err := os.ReadFile(bufYamlPath)
	if err != nil {
		return nil, fmt.Errorf("read buf.yaml: %w", err)
	}

	var cfg bufConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse buf.yaml: %w", err)
	}

	// v2 modules
	if len(cfg.Modules) > 0 {
		var allFiles []string
		for _, mod := range cfg.Modules {
			modPath := mod.Path
			if modPath == "" {
				modPath = "."
			}
			if len(filterModules) > 0 && !containsStr(filterModules, modPath) {
				continue
			}
			full := filepath.Join(root, modPath)
			files, err := discoverProtoDir(full)
			if err != nil {
				return nil, err
			}
			allFiles = append(allFiles, files...)
		}
		return allFiles, nil
	}

	// v1 or simple: just scan root
	return discoverProtoDir(root)
}

// discoverProtoDir recursively finds all .proto files in a directory.
func discoverProtoDir(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".proto") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// filterDirs filters a list of directories to only those matching the filter list.
func filterDirs(dirs, filter []string) []string {
	var result []string
	for _, d := range dirs {
		if containsStr(filter, d) {
			result = append(result, d)
		}
	}
	return result
}

// containsStr checks if a string slice contains a value.
func containsStr(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}
