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
//
// If there are no available offers, the response is simply 204-No Content and the matching
// algorithm is skipped.
//
// A matching offer is then sought.
//
// If no match is found, a fallback match is sought. If a fallback offer is matched, its
// Handle406As status code will be used, and its data is rendered using its processor; no
// further processing follows. Otherwise NoMatchAccepted is called and processing ends.
//
// If a match is found, the following happens.
//
// If the matched offer has empty data, the response will be 204-No Content; no further
// processing occurs.
//
// A check is then made for a conditional request (ETag, If-Modified-Since etc). If this is
// successful, the response is 304-Not Modified and no response rendering occurs.
//
// Finally, if statusCode is non-zero it is applied to the response (200-OK otherwise).
// Then the matched offer's data is rendered using the offer's processor.
func RenderBestMatch(rw http.ResponseWriter, req *http.Request, statusCode int, template string, available ...offerpkg.Offer) error {
	if offerpkg.Offers(available).AllEmpty() {
		rw.WriteHeader(http.StatusNoContent)
		return nil
	}

	best := BestRequestMatch(req, available...)

	if best == nil {
		NoMatchAccepted(rw, req)
		return nil
	}

	if best.Render == nil {
		panic(fmt.Sprintf("misconfigured offers for %s;charset=%s;lang=%s", best.MediaType, best.Charset, best.Language))
	}

	w := best.ApplyHeaders(rw)

	// StatusCodeOverride is a mechanism for offers to behave as error handlers.
	// Conditional request handling is disabled in this case.
	if best.StatusCodeOverride != 0 {
		rw.WriteHeader(best.StatusCodeOverride)
		return best.Render(w, req, best.Data, template, best.Language)
	}

	if best.Data == nil {
		rw.WriteHeader(http.StatusNoContent)
		return nil
	}

	sendContent, err := datapkg.ConditionalRequest(rw, req, best.Data, template, best.Language)
	if err != nil {
		return err
	}

	if !sendContent {
		return nil // status will be 304
	}

	if statusCode > 0 {
		rw.WriteHeader(statusCode)
	}

	return best.Render(w, req, best.Data, template, best.Language)
}
