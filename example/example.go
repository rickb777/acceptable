package main

import (
	"fmt"
	"net/http"

	"github.com/rickb777/acceptable/templates"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rickb777/acceptable"
	"github.com/rickb777/acceptable/processor"
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
	acceptable.Debug = func(msg string, args ...interface{}) {
		fmt.Printf(msg, args...)
	}

	templates.ReloadOnTheFly = true // development mode

	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/*", hello)

	// Start server
	e.Logger.Fatal(e.Start(":8080"))
}

// Handler
func hello(c echo.Context) error {
	best := acceptable.BestRequestMatch(c.Request(),
		acceptable.OfferOf("application/json", "en").Using(processor.JSON("  ")).
			With("en", en).With("fr", fr).With("es", es).With("ru", ru),

		acceptable.OfferOf("application/xml").Using(processor.XML("  ")).
			With("en", en).With("fr", fr).With("es", es).With("ru", ru),

		acceptable.OfferOf("text/plain").Using(processor.TXT()).
			With("en", en).With("fr", fr).With("es", es).With("ru", ru),

		templates.TextHtmlOffer("en", "example/templates/en", ".html", nil).
			With("en", en).With("fr", fr).With("es", es).With("ru", ru),

		templates.ApplicationXhtmlOffer("en", "example/templates/en", ".html", nil).
			With("en", en).With("fr", fr).With("es", es).With("ru", ru),
	)

	if best == nil {
		return c.String(http.StatusNotAcceptable, http.StatusText(http.StatusNotAcceptable))
	}

	return best.Render(c.Response(), c.Request(), *best, c.Request().URL.String()[1:])
}
