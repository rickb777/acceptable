package processor

import (
	"encoding/json"
	"net/http"

	"github.com/rickb777/acceptable"
)

const defaultJSONContentType = "application/json; charset=utf-8"

// DefaultJSONOffer is an Offer for application/json content using the JSON() processor without indentation.
var DefaultJSONOffer = acceptable.Offer{
	ContentType: acceptable.ContentType{
		Type:    "application",
		Subtype: "json",
	},
	Processor: JSON(),
}

// JSON creates a new processor for JSON with a specified indentation.
// It handles all requests except Ajax requests.
func JSON(indent ...string) acceptable.Processor {
	in := ""
	if len(indent) > 0 {
		in = indent[0]
	}

	return func(w http.ResponseWriter, match acceptable.Match, template string, dataModel interface{}) error {
		match.ApplyHeaders(w)

		p := &writerProxy{w: w}

		enc := json.NewEncoder(p)
		enc.SetIndent("", in)

		err := enc.Encode(dataModel)
		if err != nil {
			return err
		}

		return p.FinalNewline()
	}
}
