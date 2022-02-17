package acceptable

import (
	"encoding/json"
	"io"
	"net/http"

	datapkg "github.com/rickb777/acceptable/data"
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

	return func(w io.Writer, _ *http.Request, data datapkg.Data, template, language string) (err error) {
		if data == nil {
			return nil
		}

		p := &internal.WriterProxy{W: w}

		enc := json.NewEncoder(p)

		d, more, err := data.Content(template, language)
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

			d, stillMore, err = data.Content(template, language)
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
