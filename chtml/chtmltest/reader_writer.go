package chtmltest

import (
	"testing"

	"github.com/gocopper/copper/chtml"
	"github.com/gocopper/copper/clogger"
)

// NewReaderWriter creates a *chtml.ReaderWriter suitable for use in tests
func NewReaderWriter(t *testing.T) *chtml.ReaderWriter {
	t.Helper()

	r := chtml.NewRenderer(chtml.NewRendererParams{
		HTMLDir:   HTMLDir,
		StaticDir: nil,
		Config:    chtml.Config{},
	})

	return chtml.NewReaderWriter(r, chtml.Config{}, clogger.NewNoop())
}