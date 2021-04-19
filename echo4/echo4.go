package echo4

import (
	"github.com/labstack/echo/v4"
	"github.com/rickb777/acceptable"
	"github.com/rickb777/acceptable/offer"
)

// RenderBestMatch uses BestRequestMatch to find the best matching offer for the request,
// and then renders the response.
func RenderBestMatch(c echo.Context, template string, available ...offer.Offer) error {
	return acceptable.RenderBestMatch(c.Response(), c.Request(), template, available...)
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

//-------------------------------------------------------------------------------------------------

// Binary creates an output processor that outputs binary data in a form suitable for image/* and similar responses.
// Model values should be one of the following:
//
// * []byte
// * io.Reader
// * nil
func Binary() offer.Processor {
	return acceptable.Binary()
}

// CSV creates an output processor that serialises a dataModel in CSV form. With no arguments, the default
// format is comma-separated; you can supply any rune to be used as an alternative separator.
//
// Model values should be one of the following:
//
// * string or []string, or [][]string
//
// * fmt.Stringer or []fmt.Stringer, or [][]fmt.Stringer
//
// * []int or similar (bool, int8, int16, int32, int64, uint8, uint16, uint32, uint63, float32, float64, complex)
//
// * [][]int or similar (bool, int8, int16, int32, int64, uint8, uint16, uint32, uint63, float32, float64, complex)
//
// * struct for some struct in which all the fields are exported and of simple types (as above).
//
// * []struct for some struct in which all the fields are exported and of simple types (as above).
func CSV(comma ...rune) offer.Processor {
	return acceptable.CSV(comma...)
}

// JSON creates a new processor for JSON with a specified indentation.
func JSON(indent ...string) offer.Processor {
	return acceptable.JSON(indent...)
}

// TXT creates an output processor that serialises strings in a form suitable for text/plain responses.
// Model values should be one of the following:
//
// * string
//
// * fmt.Stringer
//
// * encoding.TextMarshaler
func TXT() offer.Processor {
	return acceptable.TXT()
}

// XML creates a new processor for XML with root element and optional indentation.
// The root element is used only when processing content that is a sequence of data items.
func XML(root string, indent ...string) offer.Processor {
	return acceptable.XML(root, indent...)
}
