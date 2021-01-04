package acceptable

import (
	"net/http"
	"strings"
)

const (
	Accept         = "Accept"
	AcceptLanguage = "Accept-Language"
	//AcceptCharset  = "Accept-Charset"
	// AcceptEncoding is handled effectively by net/http and can be disregarded here

	XRequestedWith = "X-Requested-With"
	XMLHttpRequest = "XMLHttpRequest"
)

//-------------------------------------------------------------------------------------------------

// IsAjax tests whether a request has the Ajax header sent by browsers for XHR requests.
func IsAjax(req *http.Request) bool {
	return req.Header.Get(XRequestedWith) == XMLHttpRequest
}

// BestRequestMatch finds the content type and language that best matches the accepted media
// ranges and languages contained in request headers.
// The result contains the best match, based on the rules of RFC-7231.
// Whenever the result is nil, the response should be 406-Not Acceptable.
func BestRequestMatch(req *http.Request, available ...Offer) *Offer {
	mrs := ParseMediaRanges(req.Header.Get(Accept)).WithDefault()
	languages := Parse(req.Header.Get(AcceptLanguage)).WithDefault()

	if IsAjax(req) {
		available = Offers(available).Filter("application", "json")
	}

	return BestMatch(mrs, languages, available...)
}

// BestMatch finds the content type and language that best matches the accepted media
// ranges and languages.
// The result contains the best match, based on the rules of RFC-7231.
// Whenever the result is nil, the response should be 406-Not Acceptable.
func BestMatch(mrs MediaRanges, languages PrecedenceValues, available ...Offer) *Offer {
	// first pass - remove offers that match exclusions
	// (this doesn't apply to language exclusions because we always allow at least one language match)
	remaining := removeExcludedOffers(mrs, available)

	// second pass - find the first exact-match media-range and language combination
	for _, offer := range remaining {
		best := findBestMatch(mrs, languages, offer, exactMatch)
		if best != nil {
			return best
		}
	}

	// third pass - find the first near-match media-range and language combination
	for _, offer := range remaining {
		best := findBestMatch(mrs, languages, offer, nearMatch)
		if best != nil {
			return best
		}
	}

	return nil // 406 - Not Acceptable
}

func removeExcludedOffers(mrs MediaRanges, available []Offer) []Offer {
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
		}
	}

	return remaining
}

func findBestMatch(mrs MediaRanges, languages PrecedenceValues, offer Offer, match func(MediaRange, PrecedenceValue, Offer) bool) *Offer {
	for _, accepted := range mrs {
		for _, lang := range languages {
			//info("compared", accepted.Value(), lang.Value, offer)

			if match(accepted, lang, offer) {
				if lang.Quality > 0 {
					return &offer
				}
			}
		}
	}

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
