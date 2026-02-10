package web

import _ "embed"

//go:embed template.html
var TemplateHTML string

//go:embed mermaid.min.js
var MermaidJS []byte

//go:embed mermaid-zenuml.js
var MermaidZenUMLJS []byte
