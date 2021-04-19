package acceptable

import (
	"html/template"

	"github.com/rickb777/acceptable/contenttype"
	"github.com/rickb777/acceptable/offer"
	"github.com/rickb777/acceptable/templates"
)

var (
	// DefaultImageOffer is an Offer for image/* content using the Binary() processor.
	DefaultImageOffer = offer.Of(Binary(), contenttype.ImageAny)

	// DefaultCSVOffer is an Offer for text/plain content using the CSV() processor.
	DefaultCSVOffer = offer.Of(CSV(), contenttype.TextCSV)

	// DefaultJSONOffer is an Offer for application/json content using the JSON() processor without indentation.
	DefaultJSONOffer = offer.Of(JSON(), contenttype.ApplicationJSON)

	// DefaultTXTOffer is an Offer for text/plain content using the TXT() processor.
	DefaultTXTOffer = offer.Of(TXT(), contenttype.TextPlain)

	// DefaultXMLOffer is an Offer for application/xml content using the XML("") processor without indentation.
	DefaultXMLOffer = offer.Of(XML("xml"), contenttype.ApplicationXML)
)

// TextHtmlOffer is an Offer for text/html content using the Template() processor.
func TextHtmlOffer(dir, suffix string, funcMap template.FuncMap) offer.Offer {
	return offer.Of(templates.Templates(dir, suffix, funcMap), contenttype.TextHTML)
}

// ApplicationXhtmlOffer is an Offer for application/xhtml+xml content using the Template() processor.
func ApplicationXhtmlOffer(dir, suffix string, funcMap template.FuncMap) offer.Offer {
	return offer.Of(templates.Templates(dir, suffix, funcMap), contenttype.ApplicationXHTML)
}
