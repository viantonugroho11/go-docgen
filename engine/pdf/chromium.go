package pdf

import (
	"context"
	"net/url"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// RenderChromeDP renders HTML to PDF using headless Chromium via chromedp.
// Unlike Engine.Render, this does not fall back to the lightweight PDF path.
func RenderChromeDP(ctx context.Context, html string) ([]byte, error) {
	return chromium(ctx, html)
}

func chromium(ctx context.Context, html string) ([]byte, error) {
	target := "data:text/html;charset=utf-8," + url.PathEscape(html)

	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	var buf []byte
	err := chromedp.Run(ctx,
		chromedp.Navigate(target),
		chromedp.WaitReady("html", chromedp.ByQuery),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			buf, _, err = page.PrintToPDF().WithPrintBackground(true).Do(ctx)
			return err
		}),
	)
	return buf, err
}
