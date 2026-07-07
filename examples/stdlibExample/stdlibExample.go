package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/rickb777/acceptable"
	"github.com/rickb777/acceptable/contenttype"
	"github.com/rickb777/acceptable/data"
	"github.com/rickb777/acceptable/examples"
	"github.com/rickb777/acceptable/offer"
	"github.com/rickb777/acceptable/templates"
)

// Some requests to try:
//
// curl -i -H 'Accept:' http://localhost:8080/
//     * gets the default, which is English as JSON
//
// curl -i -H 'Accept: application/json' -H 'Accept-Language: fr' http://localhost:8080/
//     * gets French as JSON
//
// curl -i -H 'Accept-Language: de' http://localhost:8080/
//     * gets English as JSON because there is no German and the first language offered is used instead
//
// curl -i -H 'Accept: text/html' -H 'Accept-Language: fr' http://localhost:8080/
//     * gets French as HTML using the page _index.html
//
// curl -i -H 'Accept: application/xhtml+xml' -H 'Accept-Language: ru' http://localhost:8080/home.html
//     * gets Russian as HTML using the page home.html

func main() {
	acceptable.Debug = func(msg string, args ...any) {
		fmt.Printf(msg, args...)
	}

	templates.ReloadOnTheFly = true // development mode

	mux := http.NewServeMux()

	svr := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// The main route
	mux.HandleFunc("GET /{path...}", hello)

	// a useful way to stop the demo automatically
	mux.HandleFunc("POST /stop",
		func(rw http.ResponseWriter, req *http.Request) {
			rw.WriteHeader(http.StatusNoContent)
			if f, ok := rw.(http.Flusher); ok {
				f.Flush()
			}
			time.Sleep(time.Millisecond)
			svr.Close()
		})

	err := svr.ListenAndServe()
	if err != nil {
		log.Println(err.Error())
	}
}

func hello(rw http.ResponseWriter, req *http.Request) {
	// example lazy data source (although this one just returns a fixed value)
	lazyEn := data.Lazy(func(chosen data.Chosen) (any, error) {
		return examples.EN, nil
	}).MaxAge(10 * time.Second).ETag("hash123") // replace "hash123" appropriately

	// a different example lazy data source
	lazyFr := func(chosen data.Chosen) (any, error) {
		return examples.FR, nil
	}

	template := req.URL.String()[1:]

	c := acceptable.RespondWith{Template: template}
	err := c.RenderBestMatch(rw, req,
		offer.JSON("  ").
			With(lazyEn, "en").WithFunc(lazyFr, "fr").With(examples.ES, "es").With(examples.RU, "ru"),

		offer.XML("xml", "  ").
			With(examples.EN, "en").With(examples.FR, "fr").With(examples.ES, "es").With(examples.RU, "ru"),

		offer.Of(offer.TXTProcessor(), contenttype.TextPlain).
			With(examples.EN, "en").With(examples.FR, "fr").With(examples.ES, "es").With(examples.RU, "ru"),

		templates.TextHtmlOffer("examples/templates/en", ".html", nil).
			With(examples.EN, "en").With(examples.FR, "fr").With(examples.ES, "es").With(examples.RU, "ru"),

		templates.ApplicationXhtmlOffer("examples/templates/en", ".html", nil).
			With(examples.EN, "en").With(examples.FR, "fr").With(examples.ES, "es").With(examples.RU, "ru"),
	)
	if err != nil {
		log.Println(err.Error())
	}
}
