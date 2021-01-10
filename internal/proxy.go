package internal

import "io"

type WriterProxy struct {
	W          io.Writer
	hasNewline bool
}

func (d *WriterProxy) Write(p []byte) (n int, err error) {
	n, err = d.W.Write(p)
	d.hasNewline = len(p) > 0 && p[len(p)-1] == '\n'
	return n, err
}

func (d *WriterProxy) FinalNewline() error {
	if d.hasNewline {
		return nil
	}
	_, err := d.W.Write([]byte{'\n'})
	return err
}
