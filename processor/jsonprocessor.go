package processor

import (
	"encoding/json"
	"net/http"

	"github.com/rickb777/acceptable"
	"github.com/rickb777/acceptable/data"
	"github.com/rickb777/acceptable/internal"
)

const JSONContentType = "application/json"

// DefaultJSONOffer is an Offer for application/json content using the JSON() processor without indentation.
var DefaultJSONOffer = acceptable.OfferOf(JSONContentType).Using(JSON())

// JSON creates a new processor for JSON with a specified indentation.
func JSON(indent ...string) acceptable.Processor {
	in := ""
	if len(indent) > 0 {
		in = indent[0]
	}

	return func(rw http.ResponseWriter, req *http.Request, match acceptable.Match, template string) (err error) {
		w := match.ApplyHeaders(rw)

		d, err := data.GetContentAndApplyExtraHeaders(rw, req, match.Data, template, match.Language)
		if err != nil || d == nil {
			return err
		}

		p := &internal.WriterProxy{W: w}

		enc := json.NewEncoder(p)
		enc.SetIndent("", in)

		err = enc.Encode(d)
		if err != nil {
			return err
		}

		return p.FinalNewline()
	}
}
