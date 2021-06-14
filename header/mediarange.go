package header

import (
	"fmt"
	"strings"

	"github.com/rickb777/acceptable/internal"
)

const qualityParam = "q"

const (
	// DefaultQuality is the default quality of a media range without explicit "q"
	// https://tools.ietf.org/html/rfc7231#section-5.3.1
	DefaultQuality float64 = 1.0 //e.g text/html;q=1

	// NotAcceptable is the value indicating that its item is not acceptable
	// https://tools.ietf.org/html/rfc7231#section-5.3.1
	NotAcceptable float64 = 0.0 //e.g text/foo;q=0
)

// ContentType is a media type as defined in RFC-2045, RFC-2046, RFC-2231
// (https://tools.ietf.org/html/rfc2045, https://tools.ietf.org/html/rfc2046,
// https://tools.ietf.org/html/rfc2231)
// There may also be parameters (e.g. "charset=utf-8") and extension values.
type ContentType struct {
	// Type and Subtype carry the media type, e.g. "text" and "html"
	Type, Subtype string
	// Params and Extensions hold optional parameter information
	Params     []KV
	Extensions []KV
}

// AsMediaRange converts this ContentType to a MediaRange.
// The default quality should be 1.
func (ct ContentType) AsMediaRange(quality float64) MediaRange {
	return MediaRange{
		ContentType: ct,
		Quality:     quality,
	}
}

func (ct ContentType) String() string {
	buf := &strings.Builder{}
	fmt.Fprintf(buf, "%s/%s", ct.Type, ct.Subtype)
	for _, p := range ct.Params {
		fmt.Fprintf(buf, ";%s=%s", p.Key, p.Value)
	}
	for _, p := range ct.Extensions {
		fmt.Fprintf(buf, ";%s=%s", p.Key, p.Value)
	}
	return buf.String()
}

// ContentTypeOf builds a content type value with optional parameters.
// The parameters are passed in as literal strings, e.g. "charset=utf-8".
func ContentTypeOf(typ, subtype string, paramKV ...string) ContentType {
	if typ == "" {
		typ = "*"
	}

	if subtype == "" {
		subtype = "*"
	}

	var params []KV
	if len(paramKV) > 0 {
		params = make([]KV, 0, len(paramKV))
		for _, p := range paramKV {
			k, v := internal.Split1(p, '=')
			params = append(params, KV{Key: k, Value: v})
		}
	}

	return ContentType{
		Type:    typ,
		Subtype: subtype,
		Params:  params,
	}
}

//-------------------------------------------------------------------------------------------------

// MediaRange is a content type and associated quality between 0.0 and 1.0.
type MediaRange struct {
	ContentType
	Quality float64
}

// MediaRanges holds a slice of media ranges.
type MediaRanges []MediaRange

// mrByPrecedence implements sort.Interface for []MediaRange based
// on the precedence rules. The data will be returned sorted decending
type mrByPrecedence []MediaRange

func (a mrByPrecedence) Len() int      { return len(a) }
func (a mrByPrecedence) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

func (a mrByPrecedence) Less(i, j int) bool {
	return a[i].StrongerThan(a[j])
}

// StrongerThan compares a media range with another value using the precedence rules.
func (mr MediaRange) StrongerThan(other MediaRange) bool {
	// qualities are floats so we don't use == directly
	if mr.Quality > other.Quality {
		return true
	} else if mr.Quality < other.Quality {
		return false
	}

	if mr.Type != "*" {
		if other.Type == "*" {
			return true
		}
		if mr.Subtype != "*" && other.Subtype == "*" {
			return true
		}
	}

	if mr.Type == other.Type {
		if mr.Subtype == other.Subtype {
			return len(mr.Params) > len(other.Params)
		}
	}
	return false
}

// Value gets the conjoined type and subtype string, plus any parameters.
// It does not include the quality value nor any of the extensions.
func (mr MediaRange) Value() string {
	buf := &strings.Builder{}
	fmt.Fprintf(buf, "%s/%s", mr.Type, mr.Subtype)
	for _, p := range mr.Params {
		fmt.Fprintf(buf, ";%s=%s", p.Key, p.Value)
	}
	return buf.String()
}

func (mr MediaRange) String() string {
	buf := &strings.Builder{}
	fmt.Fprintf(buf, "%s/%s", mr.Type, mr.Subtype)
	for _, p := range mr.Params {
		fmt.Fprintf(buf, ";%s=%s", p.Key, p.Value)
	}
	if mr.Quality < DefaultQuality {
		fmt.Fprintf(buf, ";q=%g", mr.Quality)
	}
	for _, p := range mr.Extensions {
		fmt.Fprintf(buf, ";%s=%s", p.Key, p.Value)
	}
	return buf.String()
}

//-------------------------------------------------------------------------------------------------

// WithDefault returns a list of media ranges that is always non-empty. If the input
// list is empty, the result holds a wildcard entry ("*/*").
func (mrs MediaRanges) WithDefault() MediaRanges {
	if len(mrs) == 0 {
		return []MediaRange{{ContentType: ContentType{Type: "*", Subtype: "*"}, Quality: DefaultQuality}}
	}
	return mrs
}

func (mrs MediaRanges) String() string {
	buf := &strings.Builder{}
	comma := ""
	for _, mr := range mrs {
		buf.WriteString(comma)
		buf.WriteString(mr.String())
		comma = ", "
	}
	return buf.String()
}
