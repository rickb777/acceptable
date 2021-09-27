package acceptable

import (
	"encoding/json"
	"net/http"

	"github.com/rickb777/acceptable/data"
	"github.com/rickb777/acceptable/internal"
	"github.com/rickb777/acceptable/offer"
)

// JSON creates a new processor for JSON with a specified indentation. This converts
// a data item (or a sequence of data items) into JSON using the standard Go encoder.
//
// When writing a sequence of items, the overall result is a JSON array starting with "["
// and ending with "]", including commas where necessary.
//
// The optional indent argument is a string usually of zero or more space characters.
func JSON(indent ...string) offer.Processor {
	in := ""
	if len(indent) > 0 {
		in = indent[0]
	}

	return func(rw http.ResponseWriter, req *http.Request, match offer.Match, template string) (err error) {
		w := match.ApplyHeaders(rw)

		sendContent, err := data.ConditionalRequest(rw, req, match.Data, template, match.Language)
		if !sendContent || err != nil {
			return err
		}

		p := &internal.WriterProxy{W: w}

		enc := json.NewEncoder(p)

		d, more, err := match.Data.Content(template, match.Language)
		if err != nil {
			return err
		}

		var newline, comma []byte
		if len(in) > 0 {
			newline = []byte{'\n'}
		}

		prefix := ""
		if more {
			prefix = in
			comma = []byte{','}
			p.Write([]byte{'['})
			p.Write(newline)
		}

		enc.SetIndent(prefix, in)

		err = enc.Encode(d)
		if err != nil {
			return err
		}

		stillMore := more
		for stillMore {
			p.Write(comma)
			p.Write(newline)

			d, stillMore, err = match.Data.Content(template, match.Language)
			if err != nil {
				return err
			}

			err = enc.Encode(d)
			if err != nil {
				return err
			}
		}

		if more {
			p.Write(newline)
			p.Write([]byte{']'})
		}

		return p.FinalNewline()
	}
}
