package processor

import (
	"encoding"
	"fmt"
	"io"
	"net/http"

	"github.com/rickb777/acceptable"
)

const defaultTxtContentType = "text/plain; charset=utf-8"

// TXT creates an output processor that serialises strings in text/plain form.
// Model values should be one of the following:
//
// * string
//
// * fmt.Stringer
//
// * encoding.TextMarshaler
func TXT() acceptable.Processor {
	return func(w http.ResponseWriter, match acceptable.Match, template string, dataModel interface{}) error {
		match.ApplyHeaders(w)

		s, ok := dataModel.(string)
		if ok {
			return writeWithNewline(w, []byte(s))
		}

		st, ok := dataModel.(fmt.Stringer)
		if ok {
			return writeWithNewline(w, []byte(st.String()))
		}

		tm, ok := dataModel.(encoding.TextMarshaler)
		if ok {
			b, err := tm.MarshalText()
			if err != nil {
				return err
			}
			return writeWithNewline(w, b)
		}

		return fmt.Errorf("Unsupported type for TXT: %T", dataModel)
	}
}

// writeWithNewline is a helper function that writes some bytes to a Writer. If the
// byte slice is empty or if the last byte is *not* newline, an extra newline is
// also written, as required for HTTP responses.
func writeWithNewline(w io.Writer, x []byte) error {
	_, err := w.Write(x)
	if err != nil {
		return err
	}

	if len(x) == 0 || x[len(x)-1] != '\n' {
		_, err = w.Write([]byte{'\n'})
	}
	return err
}
