package markdown

import (
	"strings"
	"testing"
)

// --- ExtractDiagrams ---

func TestExtractDiagrams_Backtick(t *testing.T) {
	md := "```mermaid\ngraph TD;\n  A-->B;\n```"
	blocks := ExtractDiagrams(md)
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}
	if !strings.Contains(blocks[0].Definition, "A-->B") {
		t.Errorf("expected definition to contain 'A-->B', got %q", blocks[0].Definition)
	}
	if blocks[0].Index != 1 {
		t.Errorf("expected Index 1, got %d", blocks[0].Index)
	}
}

func TestExtractDiagrams_Colon(t *testing.T) {
	md := ":::mermaid\ngraph TD;\n  A-->B;\n:::"
	blocks := ExtractDiagrams(md)
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}
	if !strings.Contains(blocks[0].Definition, "A-->B") {
		t.Errorf("expected definition to contain 'A-->B', got %q", blocks[0].Definition)
	}
}

func TestExtractDiagrams_Multiple(t *testing.T) {
	md := "```mermaid\ngraph TD;\n  A-->B;\n```\n\nSome text\n\n```mermaid\nsequenceDiagram\n  Alice->>Bob: Hi\n```"
	blocks := ExtractDiagrams(md)
	if len(blocks) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(blocks))
	}
	if blocks[0].Index != 1 {
		t.Errorf("expected first Index 1, got %d", blocks[0].Index)
	}
	if blocks[1].Index != 2 {
		t.Errorf("expected second Index 2, got %d", blocks[1].Index)
	}
	if !strings.Contains(blocks[1].Definition, "sequenceDiagram") {
		t.Errorf("expected second definition to contain 'sequenceDiagram', got %q", blocks[1].Definition)
	}
}

func TestExtractDiagrams_None(t *testing.T) {
	md := "# Just a heading\n\nSome regular markdown."
	blocks := ExtractDiagrams(md)
	if len(blocks) != 0 {
		t.Errorf("expected 0 blocks, got %d", len(blocks))
	}
}

func TestExtractDiagrams_MixedContent(t *testing.T) {
	md := `# Title

Some paragraph.

` + "```mermaid\ngraph LR;\n  X-->Y;\n```" + `

More text here.

` + "```go\nfunc main() {}\n```" + `

Final paragraph.
`
	blocks := ExtractDiagrams(md)
	if len(blocks) != 1 {
		t.Fatalf("expected 1 mermaid block among mixed content, got %d", len(blocks))
	}
	if !strings.Contains(blocks[0].Definition, "X-->Y") {
		t.Errorf("expected definition to contain 'X-->Y', got %q", blocks[0].Definition)
	}
}

// --- MarkdownImage ---

func TestMarkdownImage_Basic(t *testing.T) {
	img := MarkdownImage(ImageRef{URL: "diagram.png", Alt: "My Diagram"})
	want := "![My Diagram](diagram.png)"
	if img != want {
		t.Errorf("expected %q, got %q", want, img)
	}
}

func TestMarkdownImage_WithTitle(t *testing.T) {
	img := MarkdownImage(ImageRef{URL: "diagram.png", Alt: "My Diagram", Title: "A title"})
	want := `![My Diagram](diagram.png "A title")`
	if img != want {
		t.Errorf("expected %q, got %q", want, img)
	}
}

func TestMarkdownImage_EmptyAlt(t *testing.T) {
	img := MarkdownImage(ImageRef{URL: "diagram.png"})
	want := "![diagram](diagram.png)"
	if img != want {
		t.Errorf("expected %q, got %q", want, img)
	}
}

func TestMarkdownImage_SpecialChars(t *testing.T) {
	img := MarkdownImage(ImageRef{
		URL:   "diagram.png",
		Alt:   "diagram [1]",
		Title: `say "hello"`,
	})
	if !strings.Contains(img, `\[1\]`) {
		t.Errorf("expected brackets to be escaped in alt, got %q", img)
	}
	if !strings.Contains(img, `\"hello\"`) {
		t.Errorf("expected quotes to be escaped in title, got %q", img)
	}
}

// --- ReplaceDiagrams ---

func TestReplaceDiagrams(t *testing.T) {
	md := "Before\n\n```mermaid\ngraph TD;\n  A-->B;\n```\n\nAfter"
	images := []ImageRef{{URL: "out.png", Alt: "Diagram 1"}}
	result := ReplaceDiagrams(md, images)

	if strings.Contains(result, "```mermaid") {
		t.Error("expected mermaid block to be replaced")
	}
	if !strings.Contains(result, "![Diagram 1](out.png)") {
		t.Errorf("expected image reference in output, got %q", result)
	}
	if !strings.Contains(result, "Before") || !strings.Contains(result, "After") {
		t.Error("expected surrounding text to be preserved")
	}
}

func TestReplaceDiagrams_MoreImagesThanBlocks(t *testing.T) {
	md := "```mermaid\ngraph TD;\n  A-->B;\n```"
	images := []ImageRef{
		{URL: "first.png", Alt: "First"},
		{URL: "extra.png", Alt: "Extra"},
	}
	result := ReplaceDiagrams(md, images)

	if !strings.Contains(result, "![First](first.png)") {
		t.Errorf("expected first image, got %q", result)
	}
	// Extra image is simply ignored
	if strings.Contains(result, "extra.png") {
		t.Error("extra image should not appear in output")
	}
}

func TestReplaceDiagrams_FewerImagesThanBlocks(t *testing.T) {
	md := "```mermaid\ngraph TD;\n  A-->B;\n```\n\n```mermaid\nsequenceDiagram\n  Alice->>Bob: Hi\n```"
	images := []ImageRef{{URL: "first.png", Alt: "First"}}
	result := ReplaceDiagrams(md, images)

	if !strings.Contains(result, "![First](first.png)") {
		t.Errorf("expected first block replaced, got %q", result)
	}
	// Second block should remain as-is
	if !strings.Contains(result, "```mermaid") {
		t.Error("expected unmatched mermaid block to be left as-is")
	}
}
