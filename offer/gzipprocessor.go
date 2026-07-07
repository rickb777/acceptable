package offer

import (
	gzippkg "compress/gzip"
	"io"
	"net/http"

	dpkg "github.com/rickb777/acceptable/data"
	"github.com/rickb777/acceptable/header"
	"github.com/rickb777/acceptable/headername"
)

const (
	NoCompression  = 0
	MidCompression = 5
)

// GZIPLevel sets the compression strength when gzip is applied to a response entity.
// This is in the range 1 to 9 inclusive (see gzip.NewWriterLevel). High values should
// be avoided because the cpu cost is high but the benefit may not be sufficient.
var GZIPLevel = MidCompression

func GZIPProcessor(level int, mainProc Processor) Processor {
	return func(w io.Writer, req *http.Request, data dpkg.Data, chosen dpkg.Chosen) (err error) {
		if level == gzippkg.NoCompression {
			return mainProc(w, req, data, chosen)
		}

		acceptEncoding := header.ParsePrecedenceValues(req.Header.Get(headername.AcceptEncoding))
		if !acceptEncoding.Contains(gzip) {
			return mainProc(w, req, data, chosen)
		}

		rw := w.(http.ResponseWriter)
		rw.Header().Add(headername.ContentEncoding, gzip)
		vary := rw.Header().Get(headername.Vary)
		rw.Header().Set(headername.Vary, joinWithComma(vary, headername.AcceptEncoding))

		gw, err := gzippkg.NewWriterLevel(w, level)
		if err != nil {
			panic(err.Error() + " (see offer.GZIPLevel)") // misconfiguration
		}
		defer gw.Close()
		return mainProc(gw, req, data, chosen)
	}
}

func joinWithComma(a, b string) string {
	if a == "" {
		return b
	}
	return a + ", " + b
}

const gzip = "gzip"
