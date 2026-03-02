package svgmeta

import (
	"testing"
)

const minimalSVG = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><rect width="100" height="100"/></svg>`

func TestRoundtrip(t *testing.T) {
	source := "graph TD; A-->B"

	embedded, err := EmbedSource([]byte(minimalSVG), source)
	if err != nil {
		t.Fatalf("EmbedSource failed: %v", err)
	}

	extracted, err := ExtractSource(embedded)
	if err != nil {
		t.Fatalf("ExtractSource failed: %v", err)
	}

	if extracted != source {
		t.Errorf("roundtrip mismatch: got %q, want %q", extracted, source)
	}
}

func TestRoundtripUTF8(t *testing.T) {
	source := "graph TD; 你好-->世界; A-->🎉; α→β"

	embedded, err := EmbedSource([]byte(minimalSVG), source)
	if err != nil {
		t.Fatalf("EmbedSource failed: %v", err)
	}

	extracted, err := ExtractSource(embedded)
	if err != nil {
		t.Fatalf("ExtractSource failed: %v", err)
	}

	if extracted != source {
		t.Errorf("roundtrip mismatch: got %q, want %q", extracted, source)
	}
}

func TestRoundtripWithSpecialChars(t *testing.T) {
	// Source containing XML-special characters that CDATA should handle
	source := `graph TD; A["<hello>"]-->B["&world"]; C-->D`

	embedded, err := EmbedSource([]byte(minimalSVG), source)
	if err != nil {
		t.Fatalf("EmbedSource failed: %v", err)
	}

	extracted, err := ExtractSource(embedded)
	if err != nil {
		t.Fatalf("ExtractSource failed: %v", err)
	}

	if extracted != source {
		t.Errorf("roundtrip mismatch: got %q, want %q", extracted, source)
	}
}

func TestNoSVGElement(t *testing.T) {
	_, err := EmbedSource([]byte("<html></html>"), "test")
	if err == nil {
		t.Error("expected error for non-SVG input")
	}
}

func TestExtractMissing(t *testing.T) {
	_, err := ExtractSource([]byte(minimalSVG))
	if err == nil {
		t.Error("expected error for SVG without metadata")
	}
}

func TestSVGWithExistingMetadata(t *testing.T) {
	svgWithMeta := `<svg xmlns="http://www.w3.org/2000/svg"><metadata><dc:title>Existing</dc:title></metadata><rect/></svg>`
	source := "graph TD; A-->B"

	embedded, err := EmbedSource([]byte(svgWithMeta), source)
	if err != nil {
		t.Fatalf("EmbedSource failed: %v", err)
	}

	extracted, err := ExtractSource(embedded)
	if err != nil {
		t.Fatalf("ExtractSource failed: %v", err)
	}

	if extracted != source {
		t.Errorf("roundtrip mismatch: got %q, want %q", extracted, source)
	}
}
