package svgmeta

import (
	"fmt"
	"regexp"
	"strings"
)

const namespace = "https://mermaid.js.org/source"

// EmbedSource inserts a <metadata> element containing the mermaid source
// as the first child of the <svg> root element.
func EmbedSource(svgData []byte, source string) ([]byte, error) {
	svg := string(svgData)

	// Find the end of the opening <svg ...> tag
	idx := findSVGOpenTagEnd(svg)
	if idx < 0 {
		return nil, fmt.Errorf("no <svg> element found")
	}

	metadata := buildMetadata(source)

	// Insert metadata as first child of <svg>
	result := svg[:idx] + "\n" + metadata + svg[idx:]
	return []byte(result), nil
}

// ExtractSource finds and returns the mermaid source from SVG metadata.
func ExtractSource(svgData []byte) (string, error) {
	svg := string(svgData)

	// Look for our metadata element with CDATA
	re := regexp.MustCompile(`(?s)<mmd:source\s+xmlns:mmd="` + regexp.QuoteMeta(namespace) + `">\s*<!\[CDATA\[(.*?)\]\]>\s*</mmd:source>`)
	matches := re.FindStringSubmatch(svg)
	if matches == nil {
		return "", fmt.Errorf("no mermaid source metadata found in SVG")
	}

	return matches[1], nil
}

// findSVGOpenTagEnd returns the position right after the closing > of the <svg ...> tag.
func findSVGOpenTagEnd(svg string) int {
	// Match <svg with possible namespace prefix
	re := regexp.MustCompile(`<svg\b[^>]*>`)
	loc := re.FindStringIndex(svg)
	if loc == nil {
		return -1
	}
	return loc[1]
}

// buildMetadata creates the metadata XML element.
func buildMetadata(source string) string {
	var sb strings.Builder
	sb.WriteString("<metadata>\n")
	sb.WriteString(`  <mmd:source xmlns:mmd="`)
	sb.WriteString(namespace)
	sb.WriteString("\">\n")
	sb.WriteString("    <![CDATA[")
	sb.WriteString(source)
	sb.WriteString("]]>\n")
	sb.WriteString("  </mmd:source>\n")
	sb.WriteString("</metadata>")
	return sb.String()
}
