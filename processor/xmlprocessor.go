package processor

import (
	"encoding/xml"
	"io"
	"net/http"

	"github.com/rickb777/acceptable"
)

// XML creates a new processor for XML with optional indentation.
func XML(indent ...string) acceptable.Processor {
	if len(indent) == 0 || len(indent[0]) == 0 {
		return func(w http.ResponseWriter, match acceptable.Match, template string, dataModel interface{}) error {
			match.ApplyHeaders(w)

			return xml.NewEncoder(w).Encode(dataModel)
		}
	}

	return func(w http.ResponseWriter, match acceptable.Match, template string, dataModel interface{}) error {
		match.ApplyHeaders(w)

		x, err := xml.MarshalIndent(dataModel, "", indent[0])
		if err != nil {
			return err
		}

		return WriteWithNewline(w, x)
	}
}

//func (*xmlProcessor) CanProcess(mediaRange string, lang string) bool {
//	// see https://tools.ietf.org/html/rfc7303 XML Media Types
//	return mediaRange == "application/xml" || mediaRange == "text/xml" ||
//		strings.HasSuffix(mediaRange, "+xml") ||
//		strings.HasPrefix(mediaRange, "application/xml-") ||
//		strings.HasPrefix(mediaRange, "text/xml-")
//}

// WriteWithNewline is a helper function that writes some bytes to a Writer. If the
// byte slice is empty or if the last byte is *not* newline, an extra newline is
// also written, as required for HTTP responses.
func WriteWithNewline(w io.Writer, x []byte) error {
	_, err := w.Write(x)
	if err != nil {
		return err
	}

	if len(x) == 0 || x[len(x)-1] != '\n' {
		_, err = w.Write([]byte{'\n'})
	}
	return err
}
