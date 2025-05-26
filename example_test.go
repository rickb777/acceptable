package acceptable_test

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"sort"

	"github.com/rickb777/acceptable"
	"github.com/rickb777/acceptable/contenttype"
	"github.com/rickb777/acceptable/data"
	"github.com/rickb777/acceptable/headername"
	"github.com/rickb777/acceptable/offer"
	"github.com/rickb777/acceptable/templates"
)

func Example() {
	// In this example, the same content is available in three languages. Three different
	// approaches can be used.

	// 1. simple values can be used
	en := "Hello!" // get English content

	// 2. values can be wrapped in a data.Data
	fr := data.Of("Bonjour!").ETag("hash1") // get French content and some metadata

	// 3. this uses a lazy evaluation function, wrapped in a data.Data
	es := data.Lazy(func(template string, language string) (interface{}, error) {
		return "Hola!", nil // get Spanish content - eg from database
	}).ETagUsing(func(template, language string) (string, error) {
		// allows us to obtain the etag lazily, should we need to
		return "hash2", nil
	})

	// We're implementing an HTTP handler, so we are given a request and a response.

	req1, _ := http.NewRequest("GET", "/request1", nil) // some incoming request
	req1.Header.Set(headername.Accept, "text/plain, text/html")
	req1.Header.Set(headername.AcceptLanguage, "es, fr;q=0.8, en;q=0.6")

	req2, _ := http.NewRequest("GET", "/request2", nil) // some incoming request
	req2.Header.Set(headername.Accept, "application/json")
	req2.Header.Set(headername.AcceptLanguage, "fr")

	req3, _ := http.NewRequest("GET", "/request3", nil) // some incoming request
	req3.Header.Set(headername.Accept, "text/html")
	req3.Header.Set(headername.AcceptLanguage, "fr")
	req3.Header.Set(headername.IfNoneMatch, `"hash1"`)

	requests := []*http.Request{req1, req2, req3}

	for _, req := range requests {
		res := httptest.NewRecorder() // replace with the server's http.ResponseWriter

		// Now do the content negotiation. This example has six supported content types, all of them
		// able to serve any of the three example languages.
		//
		// The first offer is for JSON - this is often the most widely used because it also supports
		// Ajax requests.

		err := acceptable.RenderBestMatch(res, req, 200, "home.html", offer.JSON("  ").
			With(en, "en").With(fr, "fr").With(es, "es"), offer.XML("xml", "  ").
			With(en, "en").With(fr, "fr").With(es, "es"), offer.CSV().
			With(en, "en").With(fr, "fr").With(es, "es"), offer.Of(offer.TXTProcessor(), contenttype.TextPlain).
			With(en, "en").With(fr, "fr").With(es, "es"), templates.TextHtmlOffer("example/templates/en", ".html", nil).
			With(en, "en").With(fr, "fr").With(es, "es"), templates.ApplicationXhtmlOffer("example/templates/en", ".html", nil).
			With(en, "en").With(fr, "fr").With(es, "es"))

		if err != nil {
			log.Fatal(err) // replace with suitable error handling
		}

		// ----- ignore the following, which is needed only for the example test to run -----
		fmt.Printf("%s %s %d\n", req.Method, req.URL, res.Code)
		fmt.Printf("%d headers\n", len(res.Header()))
		var hdrs []string
		for h := range res.Header() {
			hdrs = append(hdrs, h)
		}
		sort.Strings(hdrs)
		for _, h := range hdrs {
			fmt.Printf("%s: %s\n", h, res.Header().Get(h))
		}
		fmt.Println()
		fmt.Println(res.Body.String())
	}

	// Output:
	// GET /request1 200
	// 4 headers
	// Content-Language: es
	// Content-Type: text/plain;charset=utf-8
	// Etag: "hash2"
	// Vary: Accept, Accept-Language
	//
	// Hola!
	//
	// GET /request2 200
	// 4 headers
	// Content-Language: fr
	// Content-Type: application/json;charset=utf-8
	// Etag: "hash1"
	// Vary: Accept, Accept-Language
	//
	// "Bonjour!"
	//
	// GET /request3 304
	// 4 headers
	// Content-Language: fr
	// Content-Type: text/html;charset=utf-8
	// Etag: "hash1"
	// Vary: Accept, Accept-Language
	//
}
