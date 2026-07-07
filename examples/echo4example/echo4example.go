package main

import (
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rickb777/acceptable"
	"github.com/rickb777/acceptable/contenttype"
	"github.com/rickb777/acceptable/data"
	"github.com/rickb777/acceptable/echo4"
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

	// Echo instance
	e := echo.New()

	// Middleware
	//e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/*", hello)

	// Start server
	e.Logger.Fatal(e.Start(":8080"))
}

// Handler
func hello(c echo.Context) error {
	// example lazy data source (although this one just returns a fixed value)
	lazyEn := data.Lazy(func(params data.Chosen) (any, error) {
		return examples.EN, nil
	}).MaxAge(10 * time.Second).ETag("hash123") // replace "hash123" appropriately

	template := c.Request().URL.String()[1:]

	return echo4.RenderBestMatch(c, 200, template,
		offer.JSON("  ").
			With(lazyEn, "EN").With(examples.FR, "FR").With(examples.ES, "ES").With(examples.RU, "RU"),

		offer.XML("xml", "  ").
			With(examples.EN, "EN").With(examples.FR, "FR").With(examples.ES, "ES").With(examples.RU, "RU"),

		offer.Of(offer.TXTProcessor(), contenttype.TextPlain).
			With(examples.EN, "EN").With(examples.FR, "FR").With(examples.ES, "ES").With(examples.RU, "RU"),

		templates.TextHtmlOffer("examples/templates/EN", ".html", nil).
			With(examples.EN, "EN").With(examples.FR, "FR").With(examples.ES, "ES").With(examples.RU, "RU"),

		templates.ApplicationXhtmlOffer("examples/templates/EN", ".html", nil).
			With(examples.EN, "EN").With(examples.FR, "FR").With(examples.ES, "ES").With(examples.RU, "RU"),
	)
}
