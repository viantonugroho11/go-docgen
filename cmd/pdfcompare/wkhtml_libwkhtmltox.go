//go:build libwkhtmltox

package main

import (
	"bytes"
	"sync"

	wkpdf "github.com/adrg/go-wkhtmltopdf"
)

var (
	wkInitMu   sync.Mutex
	wkInited   bool
	wkInitErr  error
	wkTornDown bool
)

func wkHTMLTeardown() {
	wkInitMu.Lock()
	defer wkInitMu.Unlock()
	if wkInited && !wkTornDown {
		wkpdf.Destroy()
		wkTornDown = true
		wkInited = false
	}
}

func ensureWkHTMLInit() error {
	wkInitMu.Lock()
	defer wkInitMu.Unlock()
	if wkInited {
		return wkInitErr
	}
	wkInitErr = wkpdf.Init()
	if wkInitErr != nil {
		return wkInitErr
	}
	wkInited = true
	wkTornDown = false
	return nil
}

func runWkHTMLViaLib(htmlPath string) error {
	if err := ensureWkHTMLInit(); err != nil {
		return err
	}

	obj, err := wkpdf.NewObject(htmlPath)
	if err != nil {
		return err
	}
	conv, err := wkpdf.NewConverter()
	if err != nil {
		obj.Destroy()
		return err
	}
	conv.Add(obj)
	defer conv.Destroy()

	var buf bytes.Buffer
	return conv.Run(&buf)
}
