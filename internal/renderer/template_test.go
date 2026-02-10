package renderer

import (
	"strings"
	"testing"

	"github.com/coolamit/mermaid-cli/internal/config"
	"github.com/coolamit/mermaid-cli/internal/icons"
)

func defaultOpts() RenderOpts {
	return RenderOpts{
		MermaidConfig:   config.MermaidConfig{"theme": "default"},
		BackgroundColor: "white",
	}
}

func TestBuildPageHTML_Basic(t *testing.T) {
	html, err := BuildPageHTML("graph TD;\n  A-->B;", defaultOpts())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	checks := []struct {
		label    string
		contains string
	}{
		{"DOCTYPE", "<!DOCTYPE html>"},
		{"container div", `id="container"`},
		// json.Marshal escapes > to \u003e, so A-->B becomes A--\u003eB
		{"diagram definition", `A--\u003eB`},
		{"mermaid config", `"theme":"default"`},
	}
	for _, c := range checks {
		if !strings.Contains(html, c.contains) {
			t.Errorf("expected HTML to contain %s (%q)", c.label, c.contains)
		}
	}
}

func TestBuildPageHTML_WithCSS(t *testing.T) {
	opts := defaultOpts()
	opts.CSS = "svg { border: 1px solid red; }"

	html, err := BuildPageHTML("graph TD; A-->B;", opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, "border: 1px solid red") {
		t.Error("expected custom CSS in output")
	}
}

func TestBuildPageHTML_WithIconPacks(t *testing.T) {
	opts := defaultOpts()
	opts.IconPacks = []icons.IconPack{
		{Name: "logos", URL: "https://example.com/logos.json"},
	}

	html, err := BuildPageHTML("graph TD; A-->B;", opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, "mermaid.registerIconPacks") {
		t.Error("expected registerIconPacks in output")
	}
	if !strings.Contains(html, "logos") {
		t.Error("expected icon pack name in output")
	}
}

func TestBuildPageHTML_WithSVGId(t *testing.T) {
	opts := defaultOpts()
	opts.SVGId = "custom-svg-id"

	html, err := BuildPageHTML("graph TD; A-->B;", opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, "custom-svg-id") {
		t.Error("expected custom SVG ID in output")
	}
}

func TestBuildPageHTML_SpecialChars(t *testing.T) {
	// Definition contains literal quotes and a backslash
	definition := "graph TD; A[\"Node with quotes and \\\\backslash\"]-->B;"

	html, err := BuildPageHTML(definition, defaultOpts())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// json.Marshal escapes " to \" and \ to \\
	// In the HTML: \" before "Node" (raw string `\"` matches backslash + quote)
	if !strings.Contains(html, `\"Node with quotes`) {
		t.Errorf("expected JSON-escaped quotes in output")
	}
	// Two input backslashes become four in JSON (each \ → \\)
	if !strings.Contains(html, `\\\\backslash`) {
		t.Errorf("expected JSON-escaped backslash in output")
	}
}
