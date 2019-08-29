package syntax

import (
	"bytes"
	"io"
)

type printer struct {
	buf *bytes.Buffer
}

func newPrinter() *printer {
	return &printer{
		buf: &bytes.Buffer{},
	}
}

func (p *printer) printf(w io.Writer, n interface{}) error {
	return nil
}
