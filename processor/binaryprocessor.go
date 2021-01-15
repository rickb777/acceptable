package processor

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/rickb777/acceptable"
	"github.com/rickb777/acceptable/internal"
)

// DefaultImageOffer is an Offer for image/* content using the Binary() processor.
var DefaultImageOffer = acceptable.OfferOf("image/*").Using(Binary())

// Binary creates an output processor that outputs binary data in a form suitable for image/* and similar responses.
// Model values should be one of the following:
//
// * []byte
// * io.Reader
// * acceptable.Supplier function returning one of the above
// * nil
func Binary() acceptable.Processor {
	return func(rw http.ResponseWriter, match acceptable.Match, template string) (err error) {
		w := match.ApplyHeaders(rw)

		data, err := internal.CallDataSuppliers(match.Data, template, match.Language)
		if err != nil {
			return err
		}

		switch v := data.(type) {
		case io.Reader:
			_, err = io.Copy(w, v)
		case []byte:
			rw.Header().Set("Content-Length", strconv.Itoa(len(v)))
			_, err = io.Copy(w, bytes.NewBuffer(v))
		case nil:
			// no-op
		default:
			info := fmt.Sprintf("%T: unsupported binary data", match.Data)
			panic(info)
		}

		return err
	}
}
