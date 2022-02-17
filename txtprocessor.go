package acceptable

import (
	"encoding"
	"fmt"
	"io"
	"net/http"

	datapkg "github.com/rickb777/acceptable/data"
	"github.com/rickb777/acceptable/internal"
	"github.com/rickb777/acceptable/offer"
)

// TXT creates an output processor that serialises strings in a form suitable for text/* responses (especially
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
// * nil
func TXT() offer.Processor {
	return func(w io.Writer, _ *http.Request, data datapkg.Data, template, language string) (err error) {
		p := &internal.WriterProxy{W: w}

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
