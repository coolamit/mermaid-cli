package pdfmeta

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strings"
)

var (
	beginMarker = []byte("% mermaid-source-begin")
	endMarker   = []byte("% mermaid-source-end")
)

// EmbedSource appends the mermaid source as base64-encoded PDF comments at the end of the PDF.
func EmbedSource(pdfData []byte, source string) ([]byte, error) {
	if len(pdfData) < 5 || string(pdfData[:5]) != "%PDF-" {
		return nil, fmt.Errorf("not a valid PDF file")
	}

	encoded := base64.StdEncoding.EncodeToString([]byte(source))

	var buf bytes.Buffer
	buf.Write(pdfData)
	buf.WriteByte('\n')
	buf.Write(beginMarker)
	buf.WriteByte('\n')

	// Write base64 in 76-char lines as PDF comments
	for i := 0; i < len(encoded); i += 76 {
		end := i + 76
		if end > len(encoded) {
			end = len(encoded)
		}
		buf.WriteString("% ")
		buf.WriteString(encoded[i:end])
		buf.WriteByte('\n')
	}

	buf.Write(endMarker)
	buf.WriteByte('\n')

	return buf.Bytes(), nil
}

// ExtractSource finds and returns the mermaid source from PDF comment metadata.
func ExtractSource(pdfData []byte) (string, error) {
	if len(pdfData) < 5 || string(pdfData[:5]) != "%PDF-" {
		return "", fmt.Errorf("not a valid PDF file")
	}

	beginIdx := bytes.Index(pdfData, beginMarker)
	if beginIdx < 0 {
		return "", fmt.Errorf("no mermaid source metadata found in PDF")
	}

	endIdx := bytes.Index(pdfData[beginIdx:], endMarker)
	if endIdx < 0 {
		return "", fmt.Errorf("malformed mermaid source metadata in PDF")
	}

	// Extract the lines between markers
	section := pdfData[beginIdx+len(beginMarker) : beginIdx+endIdx]
	lines := strings.Split(string(section), "\n")

	var encoded strings.Builder
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Strip the "% " prefix
		if strings.HasPrefix(line, "% ") {
			encoded.WriteString(line[2:])
		}
	}

	decoded, err := base64.StdEncoding.DecodeString(encoded.String())
	if err != nil {
		return "", fmt.Errorf("failed to decode mermaid source: %w", err)
	}

	return string(decoded), nil
}
