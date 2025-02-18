package header

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/rickb777/acceptable/internal"
)

const (
	// XMLHttpRequest is the value used always with XRequestedWith for XHR.
	XMLHttpRequest = "xmlhttprequest"

	// RFC1123 is similar to the textual time format required by HTTP.
	// Use DateTimeFormat instead. It is used for parsing because it allows any
	// three-letter timezone.
	RFC1123 = time.RFC1123

	// DateTimeFormat is the canonical textual time format required by HTTP (see
	// RFC-9110 5.6.7). The timezone is always GMT.
	DateTimeFormat = "Mon, 02 Jan 2006 15:04:05 GMT"
)

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

	parts := Split(acceptHeader, ",").TrimSpace()
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

//-------------------------------------------------------------------------------------------------

// ParseHTTPDateTime can be used for headers including Date, Expires, Last-Modified,
// If-Modified-Since, If-Unmodified-Since etc. Also, some headers such as If-Range and
// Retry-After may optionally contain a date.
//
// This first tries the preferred RFC-9110 format, although it allows UTC as well as
// GMT (i.e. as per RFC-1123), before also trying the two obsolete but still supported
// formats RFC-850 and ANSI-C.
//
// An error is returned if the input is blank or could not be parsed as an HTTP-Date
// (see RFC-9110 sec 5.6.7).
func ParseHTTPDateTime(dateString string) (time.Time, error) {
	if dateString == "" {
		return time.Time{}, fmt.Errorf(`cannot parse "" as an HTTP date`)
	}
	t, err := time.Parse(RFC1123, dateString)
	if err == nil {
		return t.UTC(), nil
	}
	t, err = time.Parse(time.RFC850, dateString)
	if err == nil {
		return t.UTC(), nil
	}
	t, err = time.Parse(time.ANSIC, dateString)
	if err == nil {
		return t.UTC(), nil
	}
	return time.Time{}, fmt.Errorf("cannot parse %q as an HTTP date", dateString)
}

// FormatHTTPDateTime formats the canonical representation of the date-time
// (see RFC-9110 section 5.6.7). The timezone GMT is always used.
func FormatHTTPDateTime(t time.Time) string {
	return t.Format(DateTimeFormat)
}
