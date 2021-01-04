package acceptable

import (
	sort "sort"
	"strings"
)

const (
	// DefaultQuality is the default quality of a media range without explicit "q"
	// https://tools.ietf.org/html/rfc7231#section-5.3.1
	DefaultQuality float64 = 1.0 //e.g text/html;q=1

	// NotAcceptable is the value indicating that its item is not acceptable
	// https://tools.ietf.org/html/rfc7231#section-5.3.1
	NotAcceptable float64 = 0.0 //e.g text/foo;q=0
)

// ParseMediaRanges splits a prioritised "Accept" header value and sorts the
// parts based on quality values and precedence rules.
// These are returned in order with the most preferred first.
//
// A request without any Accept header field implies that the user agent
// will accept any media type in response.  If the header field is
// present in a request and none of the available representations for
// the response have a media type that is listed as acceptable, the
// origin server can either honor the header field by sending a 406 (Not
// Acceptable) response or disregard the header field by treating the
// response as if it is not subject to content negotiation.
func ParseMediaRanges(acceptHeader string) MediaRanges {
	result := parseMediaRangeHeader(acceptHeader)
	sort.Stable(mrByPrecedence(result))
	return result
}

func parseMediaRangeHeader(acceptHeader string) MediaRanges {
	if acceptHeader == "" {
		return nil
	}

	parts := strings.Split(strings.ToLower(acceptHeader), ",")
	wvs := make(MediaRanges, 0, len(parts))

	for _, part := range parts {
		valueAndParams := strings.Split(part, ";")
		if len(valueAndParams) == 1 {
			t, s := split(strings.TrimSpace(valueAndParams[0]), '/')
			wvs = append(wvs, MediaRange{ContentType: ContentType{Type: t, Subtype: s}, Quality: DefaultQuality})
		} else {
			wvs = append(wvs, handleMediaRangeWithParams(valueAndParams[0], valueAndParams[1:]))
		}
	}

	return wvs
}

func handleMediaRangeWithParams(value string, acceptParams []string) MediaRange {
	wv := new(MediaRange)
	wv.Type, wv.Subtype = split(strings.TrimSpace(value), '/')
	wv.Quality = DefaultQuality

	hasQ := false
	for _, ap := range acceptParams {
		ap = strings.TrimSpace(ap)
		k, v := split(ap, '=')
		if strings.TrimSpace(k) == qualityParam {
			wv.Quality = parseQuality(v)
			hasQ = true
		} else if hasQ {
			wv.Extensions = append(wv.Extensions, KV{Key: k, Value: v})
		} else {
			wv.Params = append(wv.Params, KV{Key: k, Value: v})
		}
	}
	return *wv
}

func split(value string, b byte) (string, string) {
	i := strings.IndexByte(value, b)
	if i < 0 {
		return value, ""
	}
	return value[:i], value[i+1:]
}
