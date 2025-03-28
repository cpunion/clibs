package build

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// libInfo represents the JSON output from go list -m -json
type libInfo struct {
	Path    string
	Version string
	Dir     string
	Sum     string
}

// pkgInfo represents the JSON output from go list -json
type pkgInfo struct {
	ImportPath string
	Dir        string
	Module     struct {
		Path    string
		Version string
		Dir     string
		Sum     string
	}
}

// ListLibs gets all C libraries from the current project dependencies
func ListLibs(patterns ...string) ([]Lib, error) {
	// Get module paths
	mods, err := listMods(patterns)
	if err != nil {
		return nil, err
	}

	// Process modules to find lib.yaml files
	return findLibs(mods)
}

// listMods gets modules from specified package patterns
func listMods(patterns []string) ([]string, error) {
	// Use go list -json -deps to get package info and all dependencies
	fmt.Printf("Executing: go list -json -deps %s\n", strings.Join(patterns, " "))
	args := append([]string{"list", "-json", "-deps"}, patterns...)
	cmd := exec.Command("go", args...)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list specified packages: %v", err)
	}

	return parseJSON(out)
}

// parseJSON parses module paths from JSON output
func parseJSON(data []byte) ([]string, error) {
	var mods []string

	// Parse JSON output
	// go list -json outputs a series of JSON objects separated by newlines
	decoder := json.NewDecoder(strings.NewReader(string(data)))

	// Track processed modules to avoid duplicates
	seen := make(map[string]bool)

	for decoder.More() {
		var pkg pkgInfo
		if err := decoder.Decode(&pkg); err != nil {
			fmt.Printf("Error parsing package info: %v\n", err)
			continue
		}

		// If package has an associated module and we haven't processed it yet
		if pkg.Module.Path != "" && !seen[pkg.Module.Path] {
			mods = append(mods, pkg.Module.Path)
			seen[pkg.Module.Path] = true
		}
	}

	return mods, nil
}

// findLibs processes modules to find lib.yaml files
func findLibs(mods []string) ([]Lib, error) {
	var libs []Lib

	for _, mod := range mods {
		lib, found, err := processLib(mod)
		if err != nil {
			fmt.Printf("Error processing module %s: %v\n", mod, err)
			continue
		}

		if found {
			libs = append(libs, lib)
		}
	}

	return libs, nil
}

// processLib processes a single module to find lib.yaml
func processLib(mod string) (Lib, bool, error) {
	// Get detailed module info including Sum field
	cmd := exec.Command("go", "list", "-m", "-json", mod)
	out, err := cmd.Output()
	if err != nil {
		return Lib{}, false, fmt.Errorf("error finding module info: %v", err)
	}

	// Parse module info
	var info libInfo
	if err := json.Unmarshal(out, &info); err != nil {
		return Lib{}, false, fmt.Errorf("error parsing module info: %v", err)
	}

	dir := info.Dir
	if dir == "" {
		return Lib{}, false, fmt.Errorf("no local path found")
	}

	// Check if lib.yaml exists
	yamlPath := filepath.Join(dir, "lib.yaml")
	fmt.Printf("  Checking for lib.yaml: %s\n", yamlPath)
	if _, err := os.Stat(yamlPath); err != nil {
		// lib.yaml doesn't exist
		return Lib{}, false, nil
	}

	// Create lib object
	lib := Lib{
		ModName: mod,
		Path:    dir,
		Sum:     info.Sum,
	}

	// Read config file
	data, err := os.ReadFile(yamlPath)
	if err != nil {
		return Lib{}, false, fmt.Errorf("error reading config: %v", err)
	}

	// Parse YAML
	var config LibSpec
	if err := yaml.Unmarshal(data, &config); err != nil {
		return Lib{}, false, fmt.Errorf("error parsing YAML: %v", err)
	}

	fmt.Printf("  Found lib.yaml: %s at %s\n", mod, yamlPath)
	fmt.Printf("  Config: %v\n", config)
	lib.Config = config

	return lib, true, nil
}
