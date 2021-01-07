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
		acceptable.OfferOf("application/json", "en").Using(processor.JSON("  ")).With(en),
		acceptable.OfferOf("application/xml").Using(processor.XML("  ")).With(en),
		acceptable.OfferOf("text/plain").Using(processor.TXT()).With(en),
		acceptable.OfferOf("application/json", "en").Using(processor.JSON("  ")).With(fr),
		acceptable.OfferOf("application/xml").Using(processor.XML("  ")).With(fr),
		acceptable.OfferOf("text/plain").Using(processor.TXT()).With(fr))

	if best == nil {
		return c.String(http.StatusNotAcceptable, http.StatusText(http.StatusNotAcceptable))
	}

	data := en
	switch best.Language {
	case "fr":
		data = fr
	}

	return best.Render(c.Response(), best, "", data)
}
