package templates

import (
	"html/template"

	"github.com/rickb777/acceptable"
)

// TextHtmlOffer is an Offer for text/html content using the Template() processor.
func TextHtmlOffer(dir, suffix string, funcMap template.FuncMap) acceptable.Offer {
	return acceptable.OfferOf(Templates(dir, suffix, funcMap), TextHtml)
}

// ApplicationXhtmlOffer is an Offer for application/xhtml+xml content using the Template() processor.
func ApplicationXhtmlOffer(dir, suffix string, funcMap template.FuncMap) acceptable.Offer {
	return acceptable.OfferOf(Templates(dir, suffix, funcMap), ApplicationXhtml)
}

const (
	TextHtml         = "text/html"
	ApplicationXhtml = "application/xhtml+xml"
)
