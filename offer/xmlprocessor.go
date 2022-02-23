package offer

import (
	"encoding/xml"
	"io"
	"net/http"
	"strings"

	"github.com/rickb777/acceptable/contenttype"
	datapkg "github.com/rickb777/acceptable/data"
	"github.com/rickb777/acceptable/internal"
)

// XML constructs an XML Offer easily.
func XML(root string, indent ...string) Offer {
	return Of(XMLProcessor(root, indent...), contenttype.ApplicationXML)
}

// XMLProcessor creates a new processor for XML with root element and optional indentation. This
// converts a data item (or a sequence of data items) into XML using the standard Go encoder.
//
// The root element is used only when processing content that is a sequence of data items. It
// can be a name such as "root" or an XML element such as "<html lang='en'>".
//
// The optional indent argument is a string usually of zero or more space characters.
func XMLProcessor(root string, indent ...string) Processor {
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

	return func(w io.Writer, _ *http.Request, data datapkg.Data, template, language string) (err error) {
		p := internal.EnsureNewline(w)

		enc := xml.NewEncoder(p)

		d, more, err := data.Content(template, language)
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
