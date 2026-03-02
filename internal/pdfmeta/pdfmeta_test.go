package pdfmeta

import (
	"testing"
)

// minimalPDF is a minimal valid PDF for testing purposes.
const minimalPDF = "%PDF-1.4\n1 0 obj<</Type/Catalog>>endobj\n%%EOF\n"

func TestRoundtrip(t *testing.T) {
	source := "graph TD; A-->B"

	embedded, err := EmbedSource([]byte(minimalPDF), source)
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

	embedded, err := EmbedSource([]byte(minimalPDF), source)
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

func TestRoundtripLargeSource(t *testing.T) {
	// Source that will span multiple base64 lines
	source := "graph TD;\n"
	for i := 0; i < 100; i++ {
		source += "  A-->B;\n"
	}

	embedded, err := EmbedSource([]byte(minimalPDF), source)
	if err != nil {
		t.Fatalf("EmbedSource failed: %v", err)
	}

	extracted, err := ExtractSource(embedded)
	if err != nil {
		t.Fatalf("ExtractSource failed: %v", err)
	}

	if extracted != source {
		t.Errorf("roundtrip mismatch for large source")
	}
}

func TestInvalidPDF(t *testing.T) {
	_, err := EmbedSource([]byte("not a pdf"), "test")
	if err == nil {
		t.Error("expected error for invalid PDF")
	}

	_, err = ExtractSource([]byte("not a pdf"))
	if err == nil {
		t.Error("expected error for invalid PDF")
	}
}

func TestMissingMetadata(t *testing.T) {
	_, err := ExtractSource([]byte(minimalPDF))
	if err == nil {
		t.Error("expected error for PDF without metadata")
	}
}
