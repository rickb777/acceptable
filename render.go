package acceptable

import (
	"fmt"
	"net/http"

	"github.com/rickb777/acceptable/contenttype"
	datapkg "github.com/rickb777/acceptable/data"
	"github.com/rickb777/acceptable/headername"
	offerpkg "github.com/rickb777/acceptable/offer"
)

// NoMatchAccepted is a function used by RenderBestMatch when no match has been found.
// Replace this as needed. Note that offer.Offer can also handle 406-Not-Accepted cases,
// allowing customised error responses.
var NoMatchAccepted = func(rw http.ResponseWriter, _ *http.Request) {
	rw.Header().Set(headername.ContentType, contenttype.TextPlain+";"+contenttype.CharsetUTF8)
	rw.WriteHeader(http.StatusNotAcceptable)
	defaultNotAcceptableMessage := http.StatusText(http.StatusNotAcceptable) + "\n"
	rw.Write([]byte(defaultNotAcceptableMessage))
}

// RenderBestMatch uses BestRequestMatch to find the best matching offer for the request,
// and then renders the response. The returned error, if any, will have arisen from either
// the content provider (see data.Content) or the response processor (see offer.Processor).
func RenderBestMatch(rw http.ResponseWriter, req *http.Request, template string, available ...offerpkg.Offer) error {
	best := BestRequestMatch(req, available...)

	if best == nil {
		NoMatchAccepted(rw, req)
		return nil
	}

	if best.Render == nil {
		panic(fmt.Sprintf("misconfigured offers for %s/%s;charset=%s;lang=%s", best.Type, best.Subtype, best.Charset, best.Language))
	}

	if best.StatusCodeOverride != 0 {
		rw.WriteHeader(best.StatusCodeOverride)
	}

	w := best.ApplyHeaders(rw)

	sendContent, err := datapkg.ConditionalRequest(rw, req, best.Data, template, best.Language)
	if !sendContent || err != nil {
		return err
	}

	return best.Render(w, req, best.Data, template, best.Language)
}