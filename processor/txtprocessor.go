package processor

import (
	"encoding"
	"fmt"
	"io"
	"net/http"

	"github.com/rickb777/acceptable"
	"github.com/rickb777/acceptable/data"
)

const TextPlain = "text/plain"

// DefaultTXTOffer is an Offer for text/plain content using the TXT() processor.
var DefaultTXTOffer = acceptable.OfferOf(TXT(), TextPlain)

// TXT creates an output processor that serialises strings in a form suitable for text/plain responses.
// Model values should be one of the following:
//
// * string
//
// * fmt.Stringer
//
// * encoding.TextMarshaler
func TXT() acceptable.Processor {
	return func(rw http.ResponseWriter, req *http.Request, match acceptable.Match, template string) (err error) {
		w := match.ApplyHeaders(rw)

		d, err := data.GetContentAndApplyExtraHeaders(rw, req, match.Data, template, match.Language)
		if err != nil || d == nil {
			return err
		}

		s, ok := d.(string)
		if ok {
			return writeWithNewline(w, []byte(s))
		}

		st, ok := d.(fmt.Stringer)
		if ok {
			return writeWithNewline(w, []byte(st.String()))
		}

		tm, ok := d.(encoding.TextMarshaler)
		if ok {
			b, e2 := tm.MarshalText()
			if e2 != nil {
				return e2
			}
			return writeWithNewline(w, b)
		}

		_, err = fmt.Fprintf(w, "%v\n", d)
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
