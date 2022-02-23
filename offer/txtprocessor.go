package offer

import (
	"encoding"
	"fmt"
	"io"
	"net/http"

	"github.com/rickb777/acceptable/contenttype"
	datapkg "github.com/rickb777/acceptable/data"
	"github.com/rickb777/acceptable/internal"
)

// TextPlain returns an Offer for text/plain content using TXTProcessor.
func TextPlain() Offer { return textPlainOffer }

var textPlainOffer = Of(TXTProcessor(), contenttype.TextPlain)

// TXTProcessor creates an output processor that serialises strings in a form suitable for text/* responses (especially
// text/plain and text/html). It is also useful for JSON, XML etc that is already encoded.
//
// As required by IETF RFC, the response will always be sent with a trailing newline, even if the supplied
// content doesn't end with a newline.
//
// Model values should be one of the following:
//
// * string
// * []byte
// * fmt.Stringer
// * encoding.TextMarshaler
// * io.WriterTo
// * io.Reader
// * nil
//
// Because it handles io.Reader and io.WriterTo, TXTProcessor can be used to stream large responses (without any
// further encoding).
func TXTProcessor() Processor {
	return func(w io.Writer, _ *http.Request, data datapkg.Data, template, language string) (err error) {
		p := internal.EnsureNewline(w)

		more := data != nil

		for more {
			var d interface{}
			d, more, err = data.Content(template, language)
			if err != nil {
				return err
			}

			switch s := d.(type) {
			case []byte:
				_, err = p.Write(s)

			case string:
				_, err = p.Write([]byte(s))

			case fmt.Stringer:
				_, err = p.Write([]byte(s.String()))

			case encoding.TextMarshaler:
				b, e2 := s.MarshalText()
				if e2 != nil {
					return e2
				}
				_, err = p.Write(b)

			case io.WriterTo:
				_, err = s.WriteTo(w)

			case io.Reader:
				_, err = io.Copy(w, s)

			case nil:
				// no-op

			default:
				info := fmt.Sprintf("%T: unsupported text data", d)
				panic(info)
			}

			if err != nil {
				return err
			}
		}

		return p.FinalNewline()
	}
}
