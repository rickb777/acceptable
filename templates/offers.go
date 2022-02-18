package templates

import (
	"html/template"

	"github.com/rickb777/acceptable/contenttype"
	"github.com/rickb777/acceptable/offer"
)

// TextHtmleOffer is an Offer for text/html content using the Template() processor.
func TextHtmlOffer(dir, suffix string, funcMap template.FuncMap) offer.Offer {
	return offer.Of(Templates(dir, suffix, funcMap), contenttype.TextHTML)
}

// ApplicationXhtmlOffer is an Offer for application/xhtml+xml content using the Template() processor.
func ApplicationXhtmlOffer(dir, suffix string, funcMap template.FuncMap) offer.Offer {
	return offer.Of(Templates(dir, suffix, funcMap), contenttype.ApplicationXHTML)
}
