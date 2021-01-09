package main

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rickb777/acceptable"
	"github.com/rickb777/acceptable/processor"
)

func main() {
	acceptable.Debug = func(msg string, args ...interface{}) {
		fmt.Printf(msg, args...)
	}

	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/", hello)

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
			With("en", en).With("fr", fr).With("es", es).With("ru", ru))

	if best == nil {
		return c.String(http.StatusNotAcceptable, http.StatusText(http.StatusNotAcceptable))
	}

	return best.Render(c.Response(), *best, "")
}
