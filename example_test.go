package acceptable_test

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"sort"

	"github.com/rickb777/acceptable"
	"github.com/rickb777/acceptable/data"
	"github.com/rickb777/acceptable/processor"
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
	es := data.Lazy(func(template string, language string, cr bool) (interface{}, *data.Metadata, error) {
		return "Hola!", nil, nil // get Spanish content - eg from database
	})

	// We're implementing an HTTP handler, so we are given a request and a response.

	req1, _ := http.NewRequest("GET", "/request1", nil) // some incoming request
	req1.Header.Set("Accept", "text/plain, text/html")
	req1.Header.Set("Accept-Language", "es, fr;q=0.8, en;q=0.6")

	req2, _ := http.NewRequest("GET", "/request2", nil) // some incoming request
	req2.Header.Set("Accept", "application/json")
	req2.Header.Set("Accept-Language", "fr")

	req3, _ := http.NewRequest("GET", "/request3", nil) // some incoming request
	req3.Header.Set("Accept", "text/html")
	req3.Header.Set("Accept-Language", "fr")
	req3.Header.Set("If-None-Match", `"hash1"`)

	requests := []*http.Request{req1, req2, req3}

	for _, req := range requests {
		res := httptest.NewRecorder() // replace with the server's http.ResponseWriter

		// Now do the content negotiation. This example has six supported content types, all of them
		// able to serve any of the three example languages.
		//
		// The first offer is for JSON - this is often the most widely used because it also supports
		// Ajax requests.

		err := acceptable.RenderBestMatch(res, req, "home.html",
			acceptable.OfferOf(processor.JSON("  "), "application/json").
				With(en, "en").With(fr, "fr").With(es, "es"),

			acceptable.OfferOf(processor.XML("  "), "application/xml").
				With(en, "en").With(fr, "fr").With(es, "es"),

			acceptable.OfferOf(processor.CSV(), "text/csv").
				With(en, "en").With(fr, "fr").With(es, "es"),

			acceptable.OfferOf(processor.TXT(), "text/plain").
				With(en, "en").With(fr, "fr").With(es, "es"),

			templates.TextHtmlOffer("example/templates/en", ".html", nil).
				With(en, "en").With(fr, "fr").With(es, "es"),

			templates.ApplicationXhtmlOffer("example/templates/en", ".html", nil).
				With(en, "en").With(fr, "fr").With(es, "es"),
		)

		if err != nil {
			log.Fatal(err) // replace with suitable error handling
		}

		fmt.Printf("%s %s %d\n", req.Method, req.URL, res.Code)
		fmt.Printf("%d headers\n", len(res.Header()))
		var hdrs []string
		for h, _ := range res.Header() {
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
	// 3 headers
	// Content-Language: es
	// Content-Type: text/plain;charset=utf-8
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
