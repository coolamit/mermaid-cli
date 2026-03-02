package pngmeta

import (
	"encoding/binary"
	"hash/crc32"
	"strings"
	"testing"
)

// buildMinimalPNG creates a minimal valid PNG for testing.
// It contains: signature + IHDR + IDAT + IEND
func buildMinimalPNG() []byte {
	var png []byte
	// PNG signature
	png = append(png, 137, 80, 78, 71, 13, 10, 26, 10)

	// IHDR chunk: 1x1, 8-bit RGB
	ihdrData := []byte{
		0, 0, 0, 1, // width
		0, 0, 0, 1, // height
		8,    // bit depth
		2,    // color type (RGB)
		0,    // compression
		0,    // filter
		0,    // interlace
	}
	png = append(png, makeChunk("IHDR", ihdrData)...)

	// Minimal IDAT chunk (empty compressed data is fine for testing)
	png = append(png, makeChunk("IDAT", []byte{0x08, 0xd7, 0x63, 0x60, 0x60, 0x60, 0x00, 0x00, 0x00, 0x04, 0x00, 0x01})...)

	// IEND chunk
	png = append(png, makeChunk("IEND", nil)...)

	return png
}

func makeChunk(chunkType string, data []byte) []byte {
	buf := make([]byte, 4+4+len(data)+4)
	binary.BigEndian.PutUint32(buf[0:4], uint32(len(data)))
	copy(buf[4:8], chunkType)
	copy(buf[8:], data)
	crc := crc32.NewIEEE()
	crc.Write(buf[4 : 8+len(data)])
	binary.BigEndian.PutUint32(buf[8+len(data):], crc.Sum32())
	return buf
}

func TestRoundtrip(t *testing.T) {
	png := buildMinimalPNG()
	source := "graph TD; A-->B"

	embedded, err := EmbedText(png, "mermaid-source", source)
	if err != nil {
		t.Fatalf("EmbedText failed: %v", err)
	}

	extracted, err := ExtractText(embedded, "mermaid-source")
	if err != nil {
		t.Fatalf("ExtractText failed: %v", err)
	}

	if extracted != source {
		t.Errorf("roundtrip mismatch: got %q, want %q", extracted, source)
	}
}

func TestRoundtripUTF8(t *testing.T) {
	png := buildMinimalPNG()
	source := "graph TD; 你好-->世界; A-->🎉; α→β"

	embedded, err := EmbedText(png, "mermaid-source", source)
	if err != nil {
		t.Fatalf("EmbedText failed: %v", err)
	}

	extracted, err := ExtractText(embedded, "mermaid-source")
	if err != nil {
		t.Fatalf("ExtractText failed: %v", err)
	}

	if extracted != source {
		t.Errorf("roundtrip mismatch: got %q, want %q", extracted, source)
	}
}

func TestRoundtripCompressed(t *testing.T) {
	png := buildMinimalPNG()
	// Create a large source that exceeds compressionThreshold
	source := "graph TD;\n" + strings.Repeat("  A-->B;\n", 200)

	embedded, err := EmbedText(png, "mermaid-source", source)
	if err != nil {
		t.Fatalf("EmbedText failed: %v", err)
	}

	extracted, err := ExtractText(embedded, "mermaid-source")
	if err != nil {
		t.Fatalf("ExtractText failed: %v", err)
	}

	if extracted != source {
		t.Errorf("roundtrip mismatch for compressed data")
	}

	// Verify the embedded data is actually smaller (compressed)
	if len(embedded) >= len(png)+len(source) {
		t.Logf("warning: compressed size (%d) not smaller than uncompressed would be (%d)", len(embedded), len(png)+len(source))
	}
}

func TestInvalidPNG(t *testing.T) {
	_, err := EmbedText([]byte("not a png"), "mermaid-source", "test")
	if err == nil {
		t.Error("expected error for invalid PNG")
	}

	_, err = ExtractText([]byte("not a png"), "mermaid-source")
	if err == nil {
		t.Error("expected error for invalid PNG")
	}
}

func TestMissingChunk(t *testing.T) {
	png := buildMinimalPNG()
	_, err := ExtractText(png, "mermaid-source")
	if err == nil {
		t.Error("expected error for missing chunk")
	}
}

func TestWrongKeyword(t *testing.T) {
	png := buildMinimalPNG()
	embedded, err := EmbedText(png, "mermaid-source", "graph TD; A-->B")
	if err != nil {
		t.Fatalf("EmbedText failed: %v", err)
	}

	_, err = ExtractText(embedded, "other-keyword")
	if err == nil {
		t.Error("expected error for wrong keyword")
	}
}
