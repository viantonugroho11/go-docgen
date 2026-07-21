package pdf

import (
	"context"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// RenderChromeDP renders HTML to PDF using a fresh headless Chromium instance.
// Unlike Engine.Render, this does not fall back to the lightweight PDF path
// and does not reuse a persistent browser process.
func RenderChromeDP(ctx context.Context, html string) ([]byte, error) {
	// ctx carries the caller's deadline; allocCtx is a child of ctx so the
	// entire stack (allocator → browser → tab) is cancelled when ctx expires.
	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx, chromedp.DefaultExecAllocatorOptions[:]...)
	defer allocCancel()
	browserCtx, browserCancel := chromedp.NewContext(allocCtx)
	defer browserCancel()
	tabCtx, tabCancel := chromedp.NewContext(browserCtx)
	defer tabCancel()
	return renderInTab(tabCtx, html)
}

// renderInTab runs the PDF render pipeline using tabCtx (a chromedp tab context
// that is already bound to a browser). The caller is responsible for creating and
// cancelling tabCtx.
func renderInTab(tabCtx context.Context, html string) ([]byte, error) {
	var buf []byte
	err := chromedp.Run(tabCtx,
		// New tabs start at about:blank — Navigate is redundant and costs a round-trip.
		chromedp.ActionFunc(func(ctx context.Context) error {
			frameTree, err := page.GetFrameTree().Do(ctx)
			if err != nil {
				return err
			}
			return page.SetDocumentContent(frameTree.Frame.ID, html).Do(ctx)
		}),
		chromedp.WaitReady("html", chromedp.ByQuery),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			buf, _, err = page.PrintToPDF().WithPrintBackground(true).Do(ctx)
			return err
		}),
	)
	return buf, err
}
