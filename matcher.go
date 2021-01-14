package acceptable

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/rickb777/acceptable/header"
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
// The order of offers is important. It determines the order they are compared against
// the request headers, and it determines what defaults will be used when exact matching
// is not possible.
//
// If no available offers are provided, the response will always be nil. Note too that
// Ajax requests will result in nil being returned if no offer is capable of handling
// them, even if other offers are provided.
func BestRequestMatch(req *http.Request, available ...Offer) *Match {
	accept, accLang, vary := readHeaders(req)

	mrs := header.ParseMediaRanges(accept).WithDefault()
	languages := header.Parse(accLang).WithDefault()

	if IsAjax(req) {
		available = Offers(available).Filter("application", "json")
	}

	c := context(fmt.Sprintf("%s %s", req.Method, req.URL))
	best := c.bestMatch(mrs, languages, available, vary)

	if best != nil {
		charsets := header.Parse(req.Header.Get(AcceptCharset))
		best.Charset = "utf-8"
		// If at all possible, stick with utf-8 because (a) it is recommended; (b) no trancoding is necessary.
		// If other charsets are listed, choose one only if utf-8 is not included.
		if len(charsets) > 0 && !(charsets.Contains("utf-8") || charsets.Contains("utf8")) {
			// something other than utf-8 is legacy and deprecated, but supported anyway
			best.Charset = charsets[0].Value
			best.Vary = append(best.Vary, AcceptCharset)
		}
	}

	return best
}

func readHeaders(req *http.Request) (accept, accLang string, vary []string) {
	accept = req.Header.Get(Accept)
	accLang = req.Header.Get(AcceptLanguage)
	if accept != "" {
		vary = []string{Accept}
	}
	if accLang != "" {
		vary = append(vary, AcceptLanguage)
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
func (c context) bestMatch(mrs header.MediaRanges, languages header.PrecedenceValues, available Offers, vary []string) (best *Match) {
	// first pass - remove offers that match exclusions
	// (this doesn't apply to language exclusions because we always allow at least one language match)
	remaining := c.removeExcludedOffers(mrs, available)

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

	Debug("%s is not acceptable for %d offers (%d available)\n", c, len(remaining), len(available))
	return nil // 406 - Not Acceptable
}

func (c context) removeExcludedOffers(mrs header.MediaRanges, available []Offer) []Offer {
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
			Debug("%s excluding offer %s, langs=%v\n", c, offer.ContentType, offer.langs)
		}
	}

	return remaining
}

func (c context) findBestMatch(mrs header.MediaRanges, languages header.PrecedenceValues, offer Offer, vary []string,
	ctMatch func(header.MediaRange, Offer) bool,
	langMatch func(acceptedLang, offeredLang string) bool,
	kind string) (*Match, bool) {

	foundCtMatch := false

	for _, acceptedCT := range mrs {
		if ctMatch(acceptedCT, offer) {
			foundCtMatch = true

			for _, prefLang := range languages {
				for _, offeredLang := range offer.langs {
					if langMatch(prefLang.Value, offeredLang) {
						Debug("%s try matching %s, lang=%s to %s, lang=%s\n", c, acceptedCT, prefLang, offer.ContentType, offeredLang)

						if prefLang.Quality > 0 {
							Debug("%s successfully matched %s, lang=%s to %s, langs=%v\n", c, acceptedCT, prefLang, offer.ContentType, offer.langs)
							return buildMatch(offer, offeredLang, acceptedCT, prefLang, vary), true
						}
					}
				}
			}
		}
	}

	Debug("%s no %s match for offer %s, langs=%v\n", c, kind, offer.ContentType, offer.langs)
	return nil, foundCtMatch
}

func buildMatch(offer Offer, offeredLang string, acceptedCT header.MediaRange, prefLang header.PrecedenceValue, vary []string) *Match {
	m := &Match{
		Type:     offer.Type,
		Subtype:  offer.Subtype,
		Language: offeredLang,
		Vary:     vary,
		Data:     offer.data[offeredLang],
		Render:   offer.processor,
	}
	if offer.Subtype == "*" && acceptedCT.Subtype != "*" {
		m.Subtype = acceptedCT.Subtype
		if offer.Type == "*" && acceptedCT.Type != "*" {
			m.Type = acceptedCT.Type
		}
	}
	if offeredLang == "*" && prefLang.Value != "*" {
		m.Language = prefLang.Value
		m.Data = offer.data[prefLang.Value]
	}
	if m.Data == emptyValue {
		m.Data = nil
	}
	return m
}

//-------------------------------------------------------------------------------------------------

func exactMatch(accepted header.MediaRange, offer Offer) bool {
	return accepted.Type == offer.Type &&
		accepted.Subtype == offer.Subtype
}

func nearMatch(accepted header.MediaRange, offer Offer) bool {
	return equalOrWildcard(accepted.Type, offer.Type) &&
		equalOrWildcard(accepted.Subtype, offer.Subtype)
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
