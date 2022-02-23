package offer

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/rickb777/acceptable/contenttype"
	datapkg "github.com/rickb777/acceptable/data"
	"github.com/rickb777/acceptable/internal"
)

// JSON constructs a JSON Offer easily.
func JSON(indent ...string) Offer {
	return Of(JSONProcessor(indent...), contenttype.ApplicationJSON)
}

// JSONProcessor creates a new processor for JSON with a specified indentation. This converts
// a data item (or a sequence of data items) into JSON using the standard Go encoder.
//
// When writing a sequence of items, the overall result is a JSON array starting with "["
// and ending with "]", including commas where necessary.
//
// The optional indent argument is a string usually of zero or more space characters.
func JSONProcessor(indent ...string) Processor {
	in := ""
	if len(indent) > 0 {
		in = indent[0]
	}

	return func(w io.Writer, _ *http.Request, data datapkg.Data, template, language string) (err error) {
		p := internal.EnsureNewline(w)

		enc := NewJSONEncoder(p)

		item, more, err := data.Content(template, language)
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

		err = enc.Encode(item)
		if err != nil {
			return err
		}

		stillMore := more
		for stillMore {
			p.Write(comma)
			p.Write(newline)

			item, stillMore, err = data.Content(template, language)
			if err != nil {
				return err
			}

			err = enc.Encode(item)
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

// JSONEncoder summarises the key methods of the standard JSON encoder.
type JSONEncoder interface {
	SetIndent(string, string)
	Encode(interface{}) error
}

// NewJSONEncoder is a pluggable JSON encoder, initialised with the standard library implementation.
var NewJSONEncoder = func(w io.Writer) JSONEncoder { return json.NewEncoder(w) }
