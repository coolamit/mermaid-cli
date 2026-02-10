package markdown

import (
	"fmt"
	"regexp"
	"strings"
)

// mermaidBlockRegex matches ```mermaid ... ``` and :::mermaid ... ::: code blocks.
// Mirrors the official CLI regex: /^[^\S\n]*[`:]{3}(?:mermaid)([^\S\n]*\r?\n([\s\S]*?))[`:]{3}[^\S\n]*$/gm
var mermaidBlockRegex = regexp.MustCompile(`(?m)^[^\S\n]*[\x60:]{3}(?:mermaid)([^\S\n]*\r?\n([\s\S]*?))[\x60:]{3}[^\S\n]*$`)

// DiagramBlock represents a mermaid diagram found in markdown.
type DiagramBlock struct {
	// FullMatch is the entire matched text including fences
	FullMatch string
	// Definition is the mermaid diagram definition (inner content)
	Definition string
	// Index is the 1-based index of this diagram in the markdown
	Index int
}

// ExtractDiagrams finds all mermaid code blocks in markdown content.
func ExtractDiagrams(content string) []DiagramBlock {
	matches := mermaidBlockRegex.FindAllStringSubmatch(content, -1)
	blocks := make([]DiagramBlock, 0, len(matches))

	for i, match := range matches {
		blocks = append(blocks, DiagramBlock{
			FullMatch:  match[0],
			Definition: strings.TrimSpace(match[2]),
			Index:      i + 1,
		})
	}

	return blocks
}

// ImageRef holds information about a rendered diagram image.
type ImageRef struct {
	URL   string
	Alt   string
	Title string
}

// MarkdownImage creates a markdown image reference: ![alt](url "title")
func MarkdownImage(ref ImageRef) string {
	alt := escapeMarkdownAlt(ref.Alt)
	if alt == "" {
		alt = "diagram"
	}

	if ref.Title != "" {
		title := escapeMarkdownTitle(ref.Title)
		return fmt.Sprintf("![%s](%s \"%s\")", alt, ref.URL, title)
	}
	return fmt.Sprintf("![%s](%s)", alt, ref.URL)
}

// ReplaceDiagrams replaces mermaid code blocks in markdown with image references.
func ReplaceDiagrams(content string, images []ImageRef) string {
	idx := 0
	return mermaidBlockRegex.ReplaceAllStringFunc(content, func(match string) string {
		if idx >= len(images) {
			return match
		}
		img := images[idx]
		idx++
		return MarkdownImage(img)
	})
}

func escapeMarkdownAlt(s string) string {
	replacer := strings.NewReplacer(
		"[", "\\[",
		"]", "\\]",
		"\\", "\\\\",
	)
	return replacer.Replace(s)
}

func escapeMarkdownTitle(s string) string {
	replacer := strings.NewReplacer(
		`"`, `\"`,
		`\`, `\\`,
	)
	return replacer.Replace(s)
}
