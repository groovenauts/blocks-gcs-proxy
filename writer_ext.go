package main

import (
	"io"
)

type CompositeWriter struct {
	Main io.Writer
	Sub  io.Writer
}

func (cw *CompositeWriter) Write(p []byte) (int, error) {
	n, err := cw.Main.Write(p)
	cw.Sub.Write(p)
	return n, err
}
