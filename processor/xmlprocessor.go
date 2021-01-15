package processor

import (
	"encoding/xml"
	"net/http"

	"github.com/rickb777/acceptable"
	"github.com/rickb777/acceptable/data"
	"github.com/rickb777/acceptable/internal"
)

const XMLContentType = "application/xml"

// DefaultXMLOffer is an Offer for application/xml content using the XML() processor without indentation.
var DefaultXMLOffer = acceptable.OfferOf(XMLContentType).Using(XML())

// XML creates a new processor for XML with optional indentation.
func XML(indent ...string) acceptable.Processor {
	in := ""
	if len(indent) > 0 {
		in = indent[0]
	}

	return func(rw http.ResponseWriter, match acceptable.Match, template string) (err error) {
		w := match.ApplyHeaders(rw)

		d, err := data.GetContentAndApplyExtraHeaders(rw, match.Data, template, match.Language)
		if err != nil || d == nil {
			return err
		}

		p := &internal.WriterProxy{W: w}

		enc := xml.NewEncoder(p)
		enc.Indent("", in)

		err = enc.Encode(d)
		if err != nil {
			return err
		}

		return p.FinalNewline()
	}
}
