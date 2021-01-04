package acceptable

import (
	"fmt"
	"net/http"
	"strings"
)

const (
	Accept         = "Accept"
	AcceptLanguage = "Accept-Language"
	AcceptCharset  = "Accept-Charset"
	// AcceptEncoding is handled effectively by net/http and can be disregarded here

	// The header strings used for XHR.
	XRequestedWith = "X-Requested-With"
	XMLHttpRequest = "xmlhttprequest"
)

//-------------------------------------------------------------------------------------------------

// IsAjax tests whether a request has the Ajax header sent by browsers for XHR requests.
func IsAjax(req *http.Request) bool {
	return strings.ToLower(req.Header.Get(XRequestedWith)) == XMLHttpRequest
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
// If no available offers are provided, the response will always be nil. Note too that
// Ajax requests will result in nil being returned if no offer is capable of handling
// them, even if other offers are provided.
func BestRequestMatch(req *http.Request, available ...Offer) *Offer {
	mrs := ParseMediaRanges(req.Header.Get(Accept)).WithDefault()
	languages := Parse(req.Header.Get(AcceptLanguage)).WithDefault()

	if IsAjax(req) {
		available = Offers(available).Filter("application", "json")
	}

	c := context(fmt.Sprintf("%s %s", req.Method, req.URL))
	best := c.bestMatch(mrs, languages, available...)

	if best != nil {
		charsets := Parse(req.Header.Get(AcceptCharset))
		if len(charsets) > 0 {
			best.Charset = charsets[0].Value
		}
	}

	return best
}

// used for diagnostics
type context string

// bestMatch finds the content type and language that best matches the accepted media
// ranges and languages.
// The result contains the best match, based on the rules of RFC-7231.
//
// On exit, the result will contain the preferred language, if this is known.
//
// Whenever the result is nil, the response should be 406-Not Acceptable.
// If no available offers are provided, the response will always be nil.
func (c context) bestMatch(mrs MediaRanges, languages PrecedenceValues, available ...Offer) *Offer {
	// first pass - remove offers that match exclusions
	// (this doesn't apply to language exclusions because we always allow at least one language match)
	remaining := c.removeExcludedOffers(mrs, available)

	// second pass - find the first exact-match media-range and language combination
	for _, offer := range remaining {
		best := c.findBestMatch(mrs, languages, offer, exactMatch, "exact")
		if best != nil {
			return best
		}
	}

	// third pass - find the first near-match media-range and language combination
	for _, offer := range remaining {
		best := c.findBestMatch(mrs, languages, offer, nearMatch, "near")
		if best != nil {
			return best
		}
	}

	Debug("%s is not acceptable for %d offers (%d available)\n", c, len(remaining), len(available))
	return nil // 406 - Not Acceptable
}

func (c context) removeExcludedOffers(mrs MediaRanges, available []Offer) []Offer {
	excluded := make([]bool, len(available))
	for i, offer := range available {
		for _, accepted := range mrs {
			if accepted.Quality <= 0 &&
				accepted.Type == offer.Type &&
				accepted.Subtype == offer.Subtype {

				excluded[i] = true
			}
		}
	}

	remaining := make([]Offer, 0, len(available))
	for i, offer := range available {
		if !excluded[i] {
			remaining = append(remaining, offer)
		} else {
			Debug("%s excluding offer %s, lang=%s\n", c, offer.ContentType, offer.Language)
		}
	}

	return remaining
}

func (c context) findBestMatch(mrs MediaRanges, languages PrecedenceValues, offer Offer, match func(MediaRange, PrecedenceValue, Offer) bool, kind string) *Offer {
	for _, accepted := range mrs {
		for _, lang := range languages {
			Debug("%s try matching %s, lang=%s to %s, lang=%s\n", c, accepted, lang, offer.ContentType, offer.Language)

			if match(accepted, lang, offer) {
				if lang.Quality > 0 {
					Debug("%s successfully matched %s, lang=%s to %s, lang=%s\n", c, accepted, lang, offer.ContentType, offer.Language)

					if offer.Language == "*" && lang.Value != "*" {
						offer.Language = lang.Value
					}
					return &offer
				}
			}
		}
	}

	Debug("%s no %s match for offer %s, lang=%s\n", c, kind, offer.ContentType, offer.Language)
	return nil
}

//-------------------------------------------------------------------------------------------------

func exactMatch(accepted MediaRange, lang PrecedenceValue, offer Offer) bool {
	return accepted.Type == offer.Type &&
		accepted.Subtype == offer.Subtype &&
		equalOrPrefix(lang.Value, offer.Language)
}

func nearMatch(accepted MediaRange, lang PrecedenceValue, offer Offer) bool {
	return equalOrWildcard(accepted.Type, offer.Type) &&
		equalOrWildcard(accepted.Subtype, offer.Subtype) &&
		equalOrPrefix(lang.Value, offer.Language)
}

func equalOrPrefix(acceptedLang, offeredLang string) bool {
	return acceptedLang == "*" ||
		offeredLang == "*" ||
		acceptedLang == offeredLang ||
		strings.HasPrefix(acceptedLang, offeredLang+"-")
}

func equalOrWildcard(accepted, offered string) bool {
	return offered == "*" ||
		accepted == "*" ||
		accepted == offered
}

//-------------------------------------------------------------------------------------------------

// Debug can be used for observing decisions made by the negotiator. By default it is no-op.
var Debug = func(string, ...interface{}) {}
