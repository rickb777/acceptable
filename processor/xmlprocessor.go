package processor

import (
	"encoding/xml"
	"io"
	"net/http"

	"github.com/rickb777/acceptable"
)

// XML creates a new processor for XML with optional indentation.
func XML(indent ...string) acceptable.Processor {
	in := ""
	if len(indent) > 0 {
		in = indent[0]
	}

	return func(w http.ResponseWriter, match acceptable.Match, template string, dataModel interface{}) error {
		match.ApplyHeaders(w)

		p := &writerProxy{w: w}

		enc := xml.NewEncoder(p)
		enc.Indent("", in)
		err := enc.Encode(dataModel)
		if err != nil {
			return err
		}

		return p.FinalNewline()
	}
}

//func (*xmlProcessor) CanProcess(mediaRange string, lang string) bool {
//	// see https://tools.ietf.org/html/rfc7303 XML Media Types
//	return mediaRange == "application/xml" || mediaRange == "text/xml" ||
//		strings.HasSuffix(mediaRange, "+xml") ||
//		strings.HasPrefix(mediaRange, "application/xml-") ||
//		strings.HasPrefix(mediaRange, "text/xml-")
//}

//-------------------------------------------------------------------------------------------------

type writerProxy struct {
	w          io.Writer
	hasNewline bool
}

func (d *writerProxy) Write(p []byte) (n int, err error) {
	n, err = d.w.Write(p)
	d.hasNewline = len(p) > 0 && p[len(p)-1] == '\n'
	return n, err
}

func (d *writerProxy) FinalNewline() error {
	if d.hasNewline {
		return nil
	}
	_, err := d.w.Write([]byte{'\n'})
	return err
}
