package acceptable

import (
	"encoding/xml"
	"net/http"

	"github.com/rickb777/acceptable/data"
	"github.com/rickb777/acceptable/internal"
	"github.com/rickb777/acceptable/offer"
)

// XML creates a new processor for XML with optional indentation.
func XML(indent ...string) offer.Processor {
	in := ""
	if len(indent) > 0 {
		in = indent[0]
	}

	return func(rw http.ResponseWriter, req *http.Request, match offer.Match, template string) (err error) {
		w := match.ApplyHeaders(rw)

		d, err := data.GetContentAndApplyExtraHeaders(rw, req, match.Data, template, match.Language)
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
