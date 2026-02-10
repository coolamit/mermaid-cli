package renderer

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// RenderResult contains the output of rendering a mermaid diagram.
type RenderResult struct {
	Data  []byte
	Title string
	Desc  string
}

// Renderer handles mermaid diagram rendering via chromedp.
type Renderer struct {
	browser *Browser
}

// NewRenderer creates a new Renderer with the given browser.
func NewRenderer(browser *Browser) *Renderer {
	return &Renderer{browser: browser}
}

// Render renders a mermaid diagram to the specified output format.
func (r *Renderer) Render(ctx context.Context, definition string, outputFormat string, opts RenderOpts) (*RenderResult, error) {
	browserCtx, err := r.browser.Context(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start browser: %w", err)
	}

	// Create a new tab
	tabCtx, tabCancel := chromedp.NewContext(browserCtx)
	defer tabCancel()

	// Set timeout
	tabCtx, timeoutCancel := context.WithTimeout(tabCtx, 60*time.Second)
	defer timeoutCancel()

	// Build the HTML page
	pageHTML, err := BuildPageHTML(definition, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to build page HTML: %w", err)
	}

	// Set viewport
	if err := chromedp.Run(tabCtx,
		emulation.SetDeviceMetricsOverride(int64(opts.Width), int64(opts.Height), float64(opts.Scale), false),
	); err != nil {
		return nil, fmt.Errorf("failed to set viewport: %w", err)
	}

	// Navigate to about:blank, then set the HTML content via CDP
	var frameTree *page.FrameTree
	if err := chromedp.Run(tabCtx,
		chromedp.Navigate("about:blank"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			frameTree, err = page.GetFrameTree().Do(ctx)
			return err
		}),
	); err != nil {
		return nil, fmt.Errorf("failed to navigate: %w", err)
	}

	if err := chromedp.Run(tabCtx, chromedp.ActionFunc(func(ctx context.Context) error {
		return page.SetDocumentContent(frameTree.Frame.ID, pageHTML).Do(ctx)
	})); err != nil {
		return nil, fmt.Errorf("failed to set page content: %w", err)
	}

	// Wait for rendering to complete
	if err := chromedp.Run(tabCtx,
		chromedp.WaitReady("#container svg", chromedp.ByQuery),
	); err != nil {
		// Check if there was a render error
		var resultJSON string
		_ = chromedp.Run(tabCtx,
			chromedp.Evaluate(`JSON.stringify(window.__mmd_result || {})`, &resultJSON),
		)
		return nil, fmt.Errorf("mermaid rendering failed (waited for SVG): %w\nrender result: %s", err, resultJSON)
	}

	// Check for errors in the render result
	var resultJSON string
	if err := chromedp.Run(tabCtx,
		chromedp.Evaluate(`JSON.stringify(window.__mmd_result || {})`, &resultJSON),
	); err != nil {
		return nil, fmt.Errorf("failed to get render result: %w", err)
	}

	var renderResult struct {
		Title   *string `json:"title"`
		Desc    *string `json:"desc"`
		Success bool    `json:"success"`
		Error   string  `json:"error"`
	}
	if err := json.Unmarshal([]byte(resultJSON), &renderResult); err != nil {
		return nil, fmt.Errorf("failed to parse render result: %w", err)
	}
	if !renderResult.Success {
		return nil, fmt.Errorf("mermaid rendering error: %s", renderResult.Error)
	}

	result := &RenderResult{}
	if renderResult.Title != nil {
		result.Title = *renderResult.Title
	}
	if renderResult.Desc != nil {
		result.Desc = *renderResult.Desc
	}

	switch outputFormat {
	case "svg":
		var data []byte
		var err error
		if opts.SvgFit {
			data, err = extractSVGFit(tabCtx)
		} else {
			data, err = extractSVG(tabCtx)
		}
		if err != nil {
			return nil, err
		}
		result.Data = data

	case "png":
		data, err := capturePNG(tabCtx, opts)
		if err != nil {
			return nil, err
		}
		result.Data = data

	case "pdf":
		data, err := capturePDF(tabCtx, opts)
		if err != nil {
			return nil, err
		}
		result.Data = data

	default:
		return nil, fmt.Errorf("unsupported output format: %s", outputFormat)
	}

	return result, nil
}

// Close closes the browser.
func (r *Renderer) Close() {
	r.browser.Close()
}

// extractSVG extracts the SVG XML from the page using XMLSerializer.
func extractSVG(ctx context.Context) ([]byte, error) {
	var svgXML string
	err := chromedp.Run(ctx,
		chromedp.Evaluate(`(() => {
			const svg = document.querySelector('#container svg');
			if (!svg) return '';
			const serializer = new XMLSerializer();
			return serializer.serializeToString(svg);
		})()`, &svgXML),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to extract SVG: %w", err)
	}
	if svgXML == "" {
		return nil, fmt.Errorf("no SVG element found in rendered output")
	}
	return []byte(svgXML), nil
}

// extractSVGFit extracts the SVG XML with dimensions set to match the viewBox (for standalone viewing).
func extractSVGFit(ctx context.Context) ([]byte, error) {
	var svgXML string
	err := chromedp.Run(ctx,
		chromedp.Evaluate(`(() => {
			const svg = document.querySelector('#container svg');
			if (!svg) return '';
			const viewBox = svg.getAttribute('viewBox');
			if (viewBox) {
				const parts = viewBox.split(/\s+/);
				if (parts.length === 4) {
					svg.setAttribute('width', parts[2]);
					svg.setAttribute('height', parts[3]);
					svg.style.removeProperty('max-width');
				}
			}
			const serializer = new XMLSerializer();
			return serializer.serializeToString(svg);
		})()`, &svgXML),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to extract SVG: %w", err)
	}
	if svgXML == "" {
		return nil, fmt.Errorf("no SVG element found in rendered output")
	}
	return []byte(svgXML), nil
}

// clipRect represents a bounding rectangle.
type clipRect struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// getSVGBounds gets the bounding rect of the SVG element.
func getSVGBounds(ctx context.Context) (*clipRect, error) {
	var boundsJSON string
	err := chromedp.Run(ctx,
		chromedp.Evaluate(`(() => {
			const svg = document.querySelector('#container svg');
			if (!svg) return JSON.stringify({x:0, y:0, width:800, height:600});
			const rect = svg.getBoundingClientRect();
			return JSON.stringify({
				x: Math.floor(rect.left),
				y: Math.floor(rect.top),
				width: Math.ceil(rect.width),
				height: Math.ceil(rect.height)
			});
		})()`, &boundsJSON),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get SVG bounds: %w", err)
	}

	var bounds clipRect
	if err := json.Unmarshal([]byte(boundsJSON), &bounds); err != nil {
		return nil, fmt.Errorf("failed to parse SVG bounds: %w", err)
	}
	return &bounds, nil
}

// capturePNG captures a PNG screenshot clipped to the SVG bounds.
func capturePNG(ctx context.Context, opts RenderOpts) ([]byte, error) {
	bounds, err := getSVGBounds(ctx)
	if err != nil {
		return nil, err
	}

	// Resize viewport to fit the SVG
	newWidth := int64(bounds.X + bounds.Width)
	newHeight := int64(bounds.Y + bounds.Height)
	if err := chromedp.Run(ctx,
		emulation.SetDeviceMetricsOverride(newWidth, newHeight, float64(opts.Scale), false),
	); err != nil {
		return nil, fmt.Errorf("failed to resize viewport for PNG: %w", err)
	}

	// Small delay to let the resize settle
	time.Sleep(100 * time.Millisecond)

	clip := &page.Viewport{
		X:      bounds.X,
		Y:      bounds.Y,
		Width:  bounds.Width,
		Height: bounds.Height,
		Scale:  1,
	}

	// Set transparent background if requested
	if opts.BackgroundColor == "transparent" {
		if err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
			return emulation.SetDefaultBackgroundColorOverride().WithColor(&cdp.RGBA{R: 0, G: 0, B: 0, A: 0}).Do(ctx)
		})); err != nil {
			return nil, fmt.Errorf("failed to set transparent background: %w", err)
		}
	}

	var buf []byte
	captureParams := page.CaptureScreenshot().
		WithFormat(page.CaptureScreenshotFormatPng).
		WithClip(clip).
		WithCaptureBeyondViewport(true)

	if err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		var err error
		buf, err = captureParams.Do(ctx)
		return err
	})); err != nil {
		return nil, fmt.Errorf("failed to capture PNG: %w", err)
	}

	// Reset background color override
	if opts.BackgroundColor == "transparent" {
		_ = chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
			return emulation.SetDefaultBackgroundColorOverride().Do(ctx)
		}))
	}

	return buf, nil
}

// capturePDF captures a PDF of the page.
func capturePDF(ctx context.Context, opts RenderOpts) ([]byte, error) {
	// Set transparent background if requested
	if opts.BackgroundColor == "transparent" {
		if err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
			return emulation.SetDefaultBackgroundColorOverride().WithColor(&cdp.RGBA{R: 0, G: 0, B: 0, A: 0}).Do(ctx)
		})); err != nil {
			return nil, fmt.Errorf("failed to set transparent background: %w", err)
		}
	}

	printParams := page.PrintToPDF()

	if opts.PdfFit {
		bounds, err := getSVGBounds(ctx)
		if err != nil {
			return nil, err
		}

		// Convert px to inches (96 DPI)
		widthInches := (math.Ceil(bounds.Width) + bounds.X*2) / 96.0
		heightInches := (math.Ceil(bounds.Height) + bounds.Y*2) / 96.0

		printParams = printParams.
			WithPaperWidth(widthInches).
			WithPaperHeight(heightInches).
			WithMarginTop(0).
			WithMarginBottom(0).
			WithMarginLeft(0).
			WithMarginRight(0).
			WithPageRanges("1-1")
	}

	printParams = printParams.WithPrintBackground(true)

	var buf []byte
	if err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		var err error
		buf, _, err = printParams.Do(ctx)
		return err
	})); err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	// Reset background color override
	if opts.BackgroundColor == "transparent" {
		_ = chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
			return emulation.SetDefaultBackgroundColorOverride().Do(ctx)
		}))
	}

	return buf, nil
}
