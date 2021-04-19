package acceptable

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/rickb777/acceptable/data"
	"github.com/rickb777/acceptable/offer"
)

// Binary creates an output processor that outputs binary data in a form suitable for image/* and similar responses.
// Model values should be one of the following:
//
// * []byte
// * io.Reader
// * io.WriterTo
// * nil
func Binary() offer.Processor {
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

			switch v := d.(type) {
			case io.WriterTo:
				_, err = v.WriteTo(w)
			case io.Reader:
				_, err = io.Copy(w, v)
			case []byte:
				//rw.Header().Set("Content-Length", strconv.Itoa(len(v)))
				_, err = io.Copy(w, bytes.NewBuffer(v))
			case nil:
				// no-op
			default:
				info := fmt.Sprintf("%T: unsupported binary data", d)
				panic(info)
			}
		}

		return err
	}
}
