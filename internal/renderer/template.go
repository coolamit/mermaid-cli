package renderer

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/coolamit/mermaid-cli/internal/config"
	"github.com/coolamit/mermaid-cli/internal/icons"
	"github.com/coolamit/mermaid-cli/web"
)

// RenderOpts contains all options needed to render a mermaid diagram.
type RenderOpts struct {
	MermaidConfig   config.MermaidConfig
	BackgroundColor string
	CSS             string
	SVGId           string
	Width           int
	Height          int
	Scale           int
	PdfFit          bool
	SvgFit          bool
	IconPacks       []icons.IconPack
}

// BuildPageHTML constructs the full HTML page with embedded mermaid.js, config, and diagram.
func BuildPageHTML(definition string, opts RenderOpts) (string, error) {
	mermaidConfigJSON, err := opts.MermaidConfig.ToJSON()
	if err != nil {
		return "", fmt.Errorf("failed to serialize mermaid config: %w", err)
	}

	definitionJSON, err := json.Marshal(definition)
	if err != nil {
		return "", fmt.Errorf("failed to serialize diagram definition: %w", err)
	}

	svgIdJSON, err := json.Marshal(opts.SVGId)
	if err != nil {
		return "", fmt.Errorf("failed to serialize svgId: %w", err)
	}

	bgColorJSON, err := json.Marshal(opts.BackgroundColor)
	if err != nil {
		return "", fmt.Errorf("failed to serialize backgroundColor: %w", err)
	}

	cssJSON, err := json.Marshal(opts.CSS)
	if err != nil {
		return "", fmt.Errorf("failed to serialize CSS: %w", err)
	}

	iconPackJS := icons.GenerateIconPackJS(opts.IconPacks)

	// Build the full HTML page
	var sb strings.Builder
	sb.WriteString(`<!DOCTYPE html>
<html>
<head>
  <style>
    body { margin: 0; padding: 0; font-family: sans-serif; }
  </style>
</head>
<body>
  <div id="container"></div>
  <script>`)
	// Embed mermaid.js inline
	sb.Write(web.MermaidJS)
	sb.WriteString(`</script>
  <script>`)
	// Embed mermaid-zenuml.js inline
	sb.Write(web.MermaidZenUMLJS)
	sb.WriteString(`</script>
  <script>
    async function renderDiagram() {
      try {
        const zenuml = globalThis['mermaid-zenuml'];
        if (zenuml && zenuml.default) {
          await mermaid.registerExternalDiagrams([zenuml.default]);
        } else if (zenuml) {
          await mermaid.registerExternalDiagrams([zenuml]);
        }
`)
	sb.WriteString(iconPackJS)
	sb.WriteString(fmt.Sprintf(`
        mermaid.initialize({ startOnLoad: false, ...%s });

        const definition = %s;
        const svgId = %s || 'my-svg';
        const backgroundColor = %s;
        const myCSS = %s;

        const container = document.getElementById('container');
        const { svg: svgText } = await mermaid.render(svgId, definition, container);
        container.innerHTML = svgText;

        const svg = container.getElementsByTagName('svg')[0];
        if (svg && svg.style) {
          svg.style.backgroundColor = backgroundColor;
        }

        if (myCSS) {
          const style = document.createElementNS('http://www.w3.org/2000/svg', 'style');
          style.appendChild(document.createTextNode(myCSS));
          svg.appendChild(style);
        }

        // Extract metadata
        let title = null;
        let desc = null;
        if (svg.firstChild && svg.firstChild.nodeName === 'title') {
          title = svg.firstChild.textContent;
        }
        for (const node of svg.children) {
          if (node.nodeName === 'desc') {
            desc = node.textContent;
            break;
          }
        }

        window.__mmd_result = { title, desc, success: true };
      } catch (e) {
        window.__mmd_result = { error: e.message || String(e), success: false };
      }
    }
    renderDiagram();
  </script>
</body>
</html>`, mermaidConfigJSON, string(definitionJSON), string(svgIdJSON), string(bgColorJSON), string(cssJSON)))

	return sb.String(), nil
}

