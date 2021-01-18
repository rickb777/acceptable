package header

import (
	"sort"
	"strconv"
	"strings"

	"github.com/rickb777/acceptable/internal"
)

const (
	Accept         = "Accept"
	AcceptLanguage = "Accept-Language"
	AcceptCharset  = "Accept-Charset"
	// AcceptEncoding is handled effectively by net/http and can be disregarded here

	IfNoneMatch     = "If-None-Match"
	IfModifiedSince = "If-Modified-Since"

	// XRequestedWith defines the header strings used for XHR.
	XRequestedWith = "X-Requested-With"
	XMLHttpRequest = "xmlhttprequest"
)

//-------------------------------------------------------------------------------------------------

// ParseQuotedList extracts the comma-separated component parts from quoted headers such as If-None-Match.
// Surrounding spaces and quotes are removed.
func ParseQuotedList(listHeader string) internal.Strings {
	return internal.Split(strings.ToLower(listHeader), ",").TrimSpace().RemoveQuotes()
}

//-------------------------------------------------------------------------------------------------

// ParsePrecedenceValues splits a prioritised "Accept-Language", "Accept-Encoding" or "Accept-Charset"
// header value and sorts the parts. These are returned in order with the most
// preferred first.
func ParsePrecedenceValues(acceptXyzHeader string) PrecedenceValues {
	wvs := splitHeaderParts(strings.ToLower(acceptXyzHeader))
	sort.Stable(wvByPrecedence(wvs))
	return wvs
}

func splitHeaderParts(acceptHeader string) PrecedenceValues {
	if acceptHeader == "" {
		return nil
	}

	parts := internal.Split(acceptHeader, ",").TrimSpace()
	wvs := make(PrecedenceValues, 0, len(parts))

	for _, part := range parts {
		valueAndParams := strings.Split(part, ";")
		if len(valueAndParams) == 1 {
			wvs = append(wvs, PrecedenceValue{Value: strings.TrimSpace(valueAndParams[0]), Quality: DefaultQuality})
		} else {
			wvs = append(wvs, handlePartWithParams(valueAndParams[0], valueAndParams[1:]))
		}
	}

	return wvs
}

func handlePartWithParams(value string, acceptParams []string) PrecedenceValue {
	wv := new(PrecedenceValue)
	wv.Value = strings.TrimSpace(value)
	wv.Quality = DefaultQuality

	for _, ap := range acceptParams {
		ap = strings.TrimSpace(ap)
		k, v := internal.Split1(ap, '=')
		if strings.TrimSpace(k) == qualityParam {
			wv.Quality = parseQuality(v)
		}
	}
	return *wv
}

func parseQuality(qstring string) float64 {
	q64, err := strconv.ParseFloat(qstring, 64)
	if err != nil {
		q64 = 1.0
	}
	if q64 > DefaultQuality {
		q64 = DefaultQuality
	}
	if q64 < 0 {
		q64 = 0
	}
	return q64
}
