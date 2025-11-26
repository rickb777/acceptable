package acceptable

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/rickb777/acceptable/header"
	"github.com/rickb777/acceptable/headername"
	offerpkg "github.com/rickb777/acceptable/offer"
)

// IsAjax tests whether a request has the Ajax header sent by browsers for XHR requests.
func IsAjax(req *http.Request) bool {
	return strings.ToLower(req.Header.Get(headername.XRequestedWith)) == header.XMLHttpRequest
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
// the request headers, and this determines what defaults will be used when exact matching
// is not possible.
//
// If no available offers are provided, the response will normally be nil. Note too that
// Ajax requests will result in nil being returned if no offer is capable of handling
// JSON, even if other offers are provided.
//
// However, when no available offers are provided, a match will still be returned if any
// of the offers has its Handle406 set non-zero. This fallback match allows custom error
// messages to be returned according to the context. The
func BestRequestMatch(req *http.Request, available ...offerpkg.Offer) *offerpkg.Match {
	accept, accLang, vary := readHeaders(req)

	availables := offerpkg.Offers(available)

	mrs := header.ParseMediaRanges(accept).WithDefault()
	languages := header.ParsePrecedenceValues(accLang).WithDefault()

	if IsAjax(req) {
		availables = availables.Filter("application", "json")
	}

	c := context(fmt.Sprintf("%s %s", req.Method, req.URL))
	best := c.bestMatch(mrs, languages, availables, vary)

	if best != nil {
		charsets := header.ParsePrecedenceValues(req.Header.Get(headername.AcceptCharset))
		best.Charset = "utf-8"
		// If at all possible, stick with utf-8 because (a) it is recommended; (b) no transcoding is necessary.
		// If other charsets are listed, choose one only if utf-8 is not included.
		if len(charsets) > 0 && !(charsets.Contains("utf-8") || charsets.Contains("utf8")) {
			// something other than utf-8 is legacy and deprecated, but supported anyway
			best.Charset = charsets[0].Value
			best.Vary = append(best.Vary, headername.AcceptCharset)
		}
		return best
	}

	return c.searchForFallbackOffer(availables.CanHandle406(), mrs)
}

func (c context) searchForFallbackOffer(available offerpkg.Offers, mrs header.MediaRanges) *offerpkg.Match {
	availableFor406 := available.CanHandle406()
	if len(availableFor406) == 1 {
		return availableFor406[0].BuildFallbackMatch()
	} else if len(availableFor406) > 1 {
		// matching an excluded media range is the worst case so we try to avoid this
		remainingFor406 := c.removeExcludedOffers(mrs, availableFor406)
		if len(remainingFor406) > 0 {
			return remainingFor406[0].BuildFallbackMatch()
		}
		// nope, go ahead anyway
		return availableFor406[0].BuildFallbackMatch()
	}
	return nil
}

func readHeaders(req *http.Request) (accept, accLang string, vary []string) {
	accept = req.Header.Get(headername.Accept)
	accLang = req.Header.Get(headername.AcceptLanguage)
	if accept != "" {
		vary = []string{headername.Accept}
	}
	if accLang != "" {
		vary = append(vary, headername.AcceptLanguage)
	}
	return accept, accLang, vary
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
func (c context) bestMatch(mrs header.MediaRanges, languages header.PrecedenceValues, availables offerpkg.Offers, vary []string) (best *offerpkg.Match) {
	// first pass - remove offers that match exclusions
	// (this doesn't apply to language exclusions because we always allow at least one language match)
	remaining := c.removeExcludedOffers(mrs, availables)

	foundCtMatch := false

	for i := 1; i <= 2; i++ {
		// second pass - find the first exact-match media-range and language combination
		for _, offer := range remaining {
			best, foundCtMatch = c.findBestMatch(mrs, languages, offer, vary, exactMatch, equalOrPrefix, "exact")
			if best != nil {
				return best
			}
		}

		// third pass - find the first near-match media-range and language combination
		for _, offer := range remaining {
			best, foundCtMatch = c.findBestMatch(mrs, languages, offer, vary, nearMatch, equalOrWildcard, "near")
			if best != nil {
				return best
			}
		}

		if foundCtMatch {
			// RFC-7231 recommends that it is better to return the default language
			// than nothing at all in the case when there is no matched language.
			// So go round another loop trying to match just the content type.
			// Use a wildcard in place of the accepted language.
			languages = header.WildcardPrecedenceValue
		} else {
			break
		}
	}

	Debug("%s is not acceptable for %d offers (%d available)\n", c, len(remaining), len(availables))
	return nil // 406 - Not Acceptable
}

func (c context) removeExcludedOffers(mrs header.MediaRanges, available offerpkg.Offers) []offerpkg.Offer {
	excluded := make([]bool, len(available))
	for i, offer := range available {
		for _, accepted := range mrs {
			if accepted.Quality <= 0 &&
				accepted.MediaType == offer.MediaType {
				excluded[i] = true
			}
		}
	}

	n := 0
	for _, e := range excluded {
		if e {
			n++
		}
	}

	if n == 0 {
		return available
	}

	remaining := make(offerpkg.Offers, 0, len(available)-n)
	for i, offer := range available {
		if !excluded[i] {
			remaining = append(remaining, offer)
		} else {
			Debug("%s excluding offerpkg %s\n", c, offer)
		}
	}

	return remaining
}

func (c context) findBestMatch(mrs header.MediaRanges, languages header.PrecedenceValues, offer offerpkg.Offer, vary []string,
	contentTypeMatch func(header.MediaRange, offerpkg.Offer) bool,
	langMatch func(acceptedLang, offeredLang string) bool,
	kind string) (*offerpkg.Match, bool) {

	foundCtMatch := false

	for _, acceptedCT := range mrs {
		if contentTypeMatch(acceptedCT, offer) {
			foundCtMatch = true

			for _, prefLang := range languages {
				for _, offeredLang := range offer.Langs {
					if langMatch(prefLang.Value, offeredLang) {
						Debug("%s try matching %s, lang=%s to %s, lang=%s\n", c, acceptedCT, prefLang, offer.ContentType, offeredLang)

						if prefLang.Quality > 0 {
							Debug("%s successfully matched %s, lang=%s to %s\n", c, acceptedCT, prefLang, offer)
							if offeredLang == "*" && prefLang.Value != "*" {
								offeredLang = prefLang.Value
							}
							m := offer.BuildMatch(acceptedCT.ContentType, offeredLang, 0)
							m.Vary = vary
							return m, true
						}
					}
				}
			}
		}
	}

	Debug("%s no %s match for offerpkg %s\n", c, kind, offer)
	return nil, foundCtMatch
}

//-------------------------------------------------------------------------------------------------

func exactMatch(accepted header.MediaRange, offer offerpkg.Offer) bool {
	return accepted.MediaType == offer.MediaType
}

func nearMatch(accepted header.MediaRange, offer offerpkg.Offer) bool {
	return equalOrWildcard(accepted.Type(), offer.Type()) &&
		equalOrWildcard(accepted.Subtype(), offer.Subtype())
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

// Debug can be used for observing decisions made by the negotiation algorithm. By default it is no-op.
var Debug = func(string, ...interface{}) {}
