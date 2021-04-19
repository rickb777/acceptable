package acceptable

import (
	"encoding"
	"fmt"
	"io"
	"net/http"

	"github.com/rickb777/acceptable/data"
	"github.com/rickb777/acceptable/offer"
)

// TXT creates an output processor that serialises strings in a form suitable for text/plain responses.
// Model values should be one of the following:
//
// * string
//
// * fmt.Stringer
//
// * encoding.TextMarshaler
func TXT() offer.Processor {
	return func(rw http.ResponseWriter, req *http.Request, match offer.Match, template string) (err error) {
		w := match.ApplyHeaders(rw)

		sendContent, err := data.ConditionalRequest(rw, req, match.Data, template, match.Language)
		if !sendContent || err != nil {
			return err
		}

		more := true
		for more {
			var d interface{}
			d, more, err = match.Data.Content(template, match.Language)
			if err != nil {
				return err
			}

			switch s := d.(type) {
			case string:
				err = writeWithNewline(w, []byte(s))

			case fmt.Stringer:
				err = writeWithNewline(w, []byte(s.String()))

			case encoding.TextMarshaler:
				b, e2 := s.MarshalText()
				if e2 != nil {
					return e2
				}
				err = writeWithNewline(w, b)
			}

			if err != nil {
				return err
			}
		}
		return err
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
