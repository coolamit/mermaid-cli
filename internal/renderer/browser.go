package renderer

import (
	"context"
	"sync"

	"github.com/coolamit/mermaid-cli/internal/config"
	"github.com/chromedp/chromedp"
)

// Browser manages a lazy-started headless Chrome instance that is reused across renders.
type Browser struct {
	mu         sync.Mutex
	allocCtx   context.Context
	allocCancel context.CancelFunc
	browserCtx context.Context
	browserCancel context.CancelFunc
	started    bool
	cfg        *config.BrowserConfig
}

// NewBrowser creates a new Browser manager with the given config.
func NewBrowser(cfg *config.BrowserConfig) *Browser {
	if cfg == nil {
		cfg = &config.BrowserConfig{}
	}
	return &Browser{cfg: cfg}
}

// Context returns a chromedp context, lazily starting the browser if needed.
func (b *Browser) Context(ctx context.Context) (context.Context, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.started {
		return b.browserCtx, nil
	}

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-setuid-sandbox", true),
	)

	if b.cfg.ExecutablePath != "" {
		opts = append(opts, chromedp.ExecPath(b.cfg.ExecutablePath))
	}

	for _, arg := range b.cfg.Args {
		opts = append(opts, chromedp.Flag(arg, true))
	}

	b.allocCtx, b.allocCancel = chromedp.NewExecAllocator(ctx, opts...)
	b.browserCtx, b.browserCancel = chromedp.NewContext(b.allocCtx)

	// Run a no-op to force the browser to start
	if err := chromedp.Run(b.browserCtx); err != nil {
		b.allocCancel()
		return nil, err
	}

	b.started = true
	return b.browserCtx, nil
}

// Close shuts down the browser.
func (b *Browser) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.started {
		return
	}

	if b.browserCancel != nil {
		b.browserCancel()
	}
	if b.allocCancel != nil {
		b.allocCancel()
	}
	b.started = false
}
