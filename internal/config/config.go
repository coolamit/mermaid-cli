package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// MermaidConfig holds mermaid.js configuration options.
type MermaidConfig map[string]interface{}

// BrowserConfig holds browser launch configuration.
type BrowserConfig struct {
	ExecutablePath string   `json:"executablePath,omitempty"`
	Args           []string `json:"args,omitempty"`
	Timeout        int      `json:"timeout,omitempty"`
	Headless       string   `json:"headless,omitempty"`
}

// LoadMermaidConfig reads a mermaid config JSON file and merges it with defaults.
func LoadMermaidConfig(configFile string, theme string) (MermaidConfig, error) {
	cfg := MermaidConfig{"theme": theme}

	if configFile == "" {
		return cfg, nil
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("configuration file %q doesn't exist", configFile)
	}

	var fileCfg MermaidConfig
	if err := json.Unmarshal(data, &fileCfg); err != nil {
		return nil, fmt.Errorf("invalid JSON in config file %q: %w", configFile, err)
	}

	// Merge file config over defaults (file takes precedence)
	for k, v := range fileCfg {
		cfg[k] = v
	}

	return cfg, nil
}

// LoadBrowserConfig reads a browser config JSON file.
func LoadBrowserConfig(configFile string) (*BrowserConfig, error) {
	cfg := &BrowserConfig{}

	if configFile == "" {
		return cfg, nil
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("configuration file %q doesn't exist", configFile)
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("invalid JSON in browser config file %q: %w", configFile, err)
	}

	return cfg, nil
}

// LoadCSSFile reads a CSS file and returns its contents.
func LoadCSSFile(cssFile string) (string, error) {
	if cssFile == "" {
		return "", nil
	}

	data, err := os.ReadFile(cssFile)
	if err != nil {
		return "", fmt.Errorf("CSS file %q doesn't exist", cssFile)
	}

	return string(data), nil
}

// ToJSON serializes a MermaidConfig to JSON string.
func (c MermaidConfig) ToJSON() (string, error) {
	data, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
