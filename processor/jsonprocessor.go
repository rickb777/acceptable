package processor

import (
	"encoding/json"
	"net/http"

	"github.com/rickb777/acceptable"
	"github.com/rickb777/acceptable/internal"
)

const JSONContentType = "application/json"

// DefaultJSONOffer is an Offer for application/json content using the JSON() processor without indentation.
var DefaultJSONOffer = acceptable.OfferOf(JSONContentType).Using(JSON())

// JSON creates a new processor for JSON with a specified indentation.
// If match.Data is an acceptable.Supplier function, it is used to provide the model data.
func JSON(indent ...string) acceptable.Processor {
	in := ""
	if len(indent) > 0 {
		in = indent[0]
	}

	return func(rw http.ResponseWriter, match acceptable.Match, template string) (err error) {
		w := match.ApplyHeaders(rw)

		p := &internal.WriterProxy{W: w}

		enc := json.NewEncoder(p)
		enc.SetIndent("", in)

		data, err := internal.CallDataSuppliers(match.Data, template, match.Language)
		if err != nil {
			return err
		}

		err = enc.Encode(data)
		if err != nil {
			return err
		}

		return p.FinalNewline()
	}
}
