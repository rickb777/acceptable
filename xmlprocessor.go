package acceptable

import (
	"encoding/xml"
	"net/http"
	"strings"

	"github.com/rickb777/acceptable/data"
	"github.com/rickb777/acceptable/internal"
	"github.com/rickb777/acceptable/offer"
)

// XML creates a new processor for XML with root element and optional indentation.
// The root element is used only when processing content that is a sequence of data items.
func XML(root string, indent ...string) offer.Processor {
	if root == "" {
		root = "<xml>"
	}
	if !strings.HasPrefix(root, "<") {
		root = "<" + root
	}
	if !strings.HasSuffix(root, ">") {
		root += ">"
	}

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

		enc := xml.NewEncoder(p)

		d, more, err := match.Data.Content(template, match.Language)
		if err != nil {
			return err
		}

		var newline []byte
		if len(in) > 0 {
			newline = []byte{'\n'}
		}

		prefix := ""
		if more {
			prefix = in
			p.Write([]byte(root))
			p.Write(newline)
		}

		enc.Indent(prefix, in)

		err = enc.Encode(d)
		if err != nil {
			return err
		}

		stillMore := more
		for stillMore {
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
			p.Write([]byte(closing(root)))
		}

		return p.FinalNewline()
	}
}

func closing(root string) string {
	parts := strings.SplitN(root, " ", 2)
	clse := "</" + parts[0][1:]
	if !strings.HasSuffix(clse, ">") {
		return clse + ">"
	}
	return clse
}
