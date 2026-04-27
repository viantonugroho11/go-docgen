//go:build !libwkhtmltox

package main

import "fmt"

func wkHTMLTeardown() {}

func runWkHTMLViaLib(htmlPath string) error {
	_ = htmlPath
	return fmt.Errorf("in-process libwkhtmltox: CGO_ENABLED=1 go run -tags libwkhtmltox . (from cmd/pdfcompare); libwkhtmltox must match GOARCH (see README)")
}
