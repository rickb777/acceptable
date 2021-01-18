package processor

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/rickb777/acceptable"
	"github.com/rickb777/acceptable/data"
)

// DefaultImageOffer is an Offer for image/* content using the Binary() processor.
var DefaultImageOffer = acceptable.OfferOf(Binary(), "image/*")

// Binary creates an output processor that outputs binary data in a form suitable for image/* and similar responses.
// Model values should be one of the following:
//
// * []byte
// * io.Reader
// * nil
func Binary() acceptable.Processor {
	return func(rw http.ResponseWriter, req *http.Request, match acceptable.Match, template string) (err error) {
		w := match.ApplyHeaders(rw)

		d, err := data.GetContentAndApplyExtraHeaders(rw, req, match.Data, template, match.Language)
		if err != nil || d == nil {
			return err
		}

		switch v := d.(type) {
		case io.Reader:
			_, err = io.Copy(w, v)
		case []byte:
			rw.Header().Set("Content-Length", strconv.Itoa(len(v)))
			_, err = io.Copy(w, bytes.NewBuffer(v))
		case nil:
			// no-op
		default:
			info := fmt.Sprintf("%T: unsupported binary data", d)
			panic(info)
		}

		return err
	}
}
