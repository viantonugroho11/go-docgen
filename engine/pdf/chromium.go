package pdf

import (
	"context"
	"net/url"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

func chromium(ctx context.Context, html string) ([]byte, error) {
	target := "data:text/html;charset=utf-8," + url.PathEscape(html)

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
