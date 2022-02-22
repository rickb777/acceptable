package echo4

import (
	"github.com/labstack/echo/v4"
	"github.com/rickb777/acceptable"
	"github.com/rickb777/acceptable/offer"
)

// RenderBestMatch uses BestRequestMatch to find the best matching offer for the request,
// and then renders the response.
// If statusCode is 0, the default (200-status OK) will be used.
func RenderBestMatch(c echo.Context, statusCode int, template string, available ...offer.Offer) error {
	return acceptable.RenderBestMatch(c.Response(), c.Request(), statusCode, template, available...)
}

// BestRequestMatch finds the content type and language that best matches the accepted media
// ranges and languages contained in request headers.
// The result contains the best match, based on the rules of RFC-7231.
// On exit, the result will contain the preferred language and charset, if these are known.
//
// Whenever the result is nil, the response should be 406-Not Acceptable.
//
// For all Ajax requests, the available offers are filtered so that only those capable
// of providing an Ajax response are considered by the content negotiation algorithm.
// The other offers are discarded.
//
// The order of offers is important. It determines the order they are compared against
// the request headers, and it determines what defaults will be used when exact matching
// is not possible.
//
// If no available offers are provided, the response will always be nil. Note too that
// Ajax requests will result in nil being returned if no offer is capable of handling
// them, even if other offers are provided.
func BestRequestMatch(c echo.Context, available ...offer.Offer) *offer.Match {
	return acceptable.BestRequestMatch(c.Request(), available...)
}
