package templates

import (
	"html/template"

	"github.com/rickb777/acceptable"
)

// TextHtmlOffer is an Offer for text/html content using the Template() processor.
func TextHtmlOffer(language, dir, suffix string, funcMap template.FuncMap) acceptable.Offer {
	return acceptable.OfferOf(TextHtml, language).Using(Templates(dir, suffix, funcMap, false))
}

func ApplicationXhtmlOffer(language, dir, suffix string, funcMap template.FuncMap) acceptable.Offer {
	return acceptable.OfferOf(ApplicationXhtml, language).Using(Templates(dir, suffix, funcMap, false))
}

const (
	TextHtml         = "text/html"
	ApplicationXhtml = "application/xhtml+xml"
)
