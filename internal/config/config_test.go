package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- LoadMermaidConfig ---

func TestLoadMermaidConfig_EmptyFile(t *testing.T) {
	cfg, err := LoadMermaidConfig("", "default")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg["theme"] != "default" {
		t.Errorf("expected theme %q, got %q", "default", cfg["theme"])
	}
}

func TestLoadMermaidConfig_WithFile(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "config.json")
	os.WriteFile(p, []byte(`{"theme":"dark","logLevel":"error"}`), 0644)

	cfg, err := LoadMermaidConfig(p, "default")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// File key overrides the default theme
	if cfg["theme"] != "dark" {
		t.Errorf("expected theme %q, got %q", "dark", cfg["theme"])
	}
	if cfg["logLevel"] != "error" {
		t.Errorf("expected logLevel %q, got %q", "error", cfg["logLevel"])
	}
}

func TestLoadMermaidConfig_MissingFile(t *testing.T) {
	_, err := LoadMermaidConfig("/nonexistent/config.json", "default")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoadMermaidConfig_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "bad.json")
	os.WriteFile(p, []byte(`{not json}`), 0644)

	_, err := LoadMermaidConfig(p, "default")
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
	if !strings.Contains(err.Error(), "invalid JSON") {
		t.Errorf("expected 'invalid JSON' in error, got: %v", err)
	}
}

// --- LoadBrowserConfig ---

func TestLoadBrowserConfig_EmptyFile(t *testing.T) {
	cfg, err := LoadBrowserConfig("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ExecutablePath != "" {
		t.Errorf("expected empty ExecutablePath, got %q", cfg.ExecutablePath)
	}
	if len(cfg.Args) != 0 {
		t.Errorf("expected empty Args, got %v", cfg.Args)
	}
}

func TestLoadBrowserConfig_WithFile(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "browser.json")
	os.WriteFile(p, []byte(`{"executablePath":"/usr/bin/chromium","args":["--no-sandbox"],"timeout":30000,"headless":"new"}`), 0644)

	cfg, err := LoadBrowserConfig(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ExecutablePath != "/usr/bin/chromium" {
		t.Errorf("expected executablePath %q, got %q", "/usr/bin/chromium", cfg.ExecutablePath)
	}
	if len(cfg.Args) != 1 || cfg.Args[0] != "--no-sandbox" {
		t.Errorf("expected args [--no-sandbox], got %v", cfg.Args)
	}
	if cfg.Timeout != 30000 {
		t.Errorf("expected timeout 30000, got %d", cfg.Timeout)
	}
	if cfg.Headless != "new" {
		t.Errorf("expected headless %q, got %q", "new", cfg.Headless)
	}
}

func TestLoadBrowserConfig_MissingFile(t *testing.T) {
	_, err := LoadBrowserConfig("/nonexistent/browser.json")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoadBrowserConfig_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "bad.json")
	os.WriteFile(p, []byte(`{nope}`), 0644)

	_, err := LoadBrowserConfig(p)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
	if !strings.Contains(err.Error(), "invalid JSON") {
		t.Errorf("expected 'invalid JSON' in error, got: %v", err)
	}
}

// --- LoadCSSFile ---

func TestLoadCSSFile_Empty(t *testing.T) {
	css, err := LoadCSSFile("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if css != "" {
		t.Errorf("expected empty string, got %q", css)
	}
}

func TestLoadCSSFile_WithFile(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "style.css")
	content := "body { background: red; }\n"
	os.WriteFile(p, []byte(content), 0644)

	css, err := LoadCSSFile(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if css != content {
		t.Errorf("expected %q, got %q", content, css)
	}
}

func TestLoadCSSFile_MissingFile(t *testing.T) {
	_, err := LoadCSSFile("/nonexistent/style.css")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

// --- ToJSON ---

func TestToJSON(t *testing.T) {
	cfg := MermaidConfig{"theme": "forest"}
	j, err := cfg.ToJSON()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(j, `"theme":"forest"`) {
		t.Errorf("expected JSON to contain theme, got %q", j)
	}
}
