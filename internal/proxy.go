package internal

import "io"

type WriterProxy struct {
	w          io.Writer
	hasNewline bool
}

func EnsureNewline(w io.Writer) *WriterProxy {
	return &WriterProxy{w: w, hasNewline: true}
}

func (wp *WriterProxy) Write(p []byte) (n int, err error) {
	if len(p) > 0 {
		n, err = wp.w.Write(p)
		wp.hasNewline = p[len(p)-1] == '\n'
	}
	return n, err
}

func (wp *WriterProxy) FinalNewline() error {
	if wp.hasNewline {
		return nil
	}
	_, err := wp.w.Write([]byte{'\n'})
	return err
}
