package processor

import (
	"encoding"
	"fmt"
	"io"
	"net/http"

	"github.com/rickb777/acceptable"
)

const TextPlain = "text/plain"

// DefaultTXTOffer is an Offer for text/plain content using the TXT() processor.
var DefaultTXTOffer = acceptable.OfferOf(TextPlain).Using(TXT())

// TXT creates an output processor that serialises strings in a form suitable for text/plain responses.
// Model values should be one of the following:
//
// * string
//
// * fmt.Stringer
//
// * encoding.TextMarshaler
func TXT() acceptable.Processor {
	return func(rw http.ResponseWriter, match acceptable.Match, template string) (err error) {
		w := match.ApplyHeaders(rw)

		if fn, isFunc := match.Data.(acceptable.Supplier); isFunc {
			match.Data, err = fn()
			if err != nil {
				return err
			}
		}

		s, ok := match.Data.(string)
		if ok {
			return writeWithNewline(w, []byte(s))
		}

		st, ok := match.Data.(fmt.Stringer)
		if ok {
			return writeWithNewline(w, []byte(st.String()))
		}

		tm, ok := match.Data.(encoding.TextMarshaler)
		if ok {
			b, e2 := tm.MarshalText()
			if e2 != nil {
				return e2
			}
			return writeWithNewline(w, b)
		}

		_, err = fmt.Fprintf(w, "%v\n", match.Data)
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
