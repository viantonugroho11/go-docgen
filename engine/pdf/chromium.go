package pdf

import (
	"context"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// RenderChromeDP renders HTML to PDF using headless Chromium via chromedp.
// Unlike Engine.Render, this does not fall back to the lightweight PDF path.
func RenderChromeDP(ctx context.Context, html string) ([]byte, error) {
	return chromium(ctx, html)
}

// chromium injects HTML via Page.setDocumentContent instead of a data: URL to avoid url.PathEscape
// duplicating the full markup on the Go heap (CDP still carries the payload to the browser).
func chromium(ctx context.Context, html string) ([]byte, error) {
	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	var buf []byte
	err := chromedp.Run(ctx,
		chromedp.Navigate("about:blank"),
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
