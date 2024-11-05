package header

import (
	"io"
	"net/http"
	"strings"

	"github.com/rickb777/acceptable/headername"

	"github.com/rickb777/acceptable/internal"
)

// ContentType is a media type as defined in RFC-2045, RFC-2046, RFC-2231
// (https://tools.ietf.org/html/rfc2045, https://tools.ietf.org/html/rfc2046,
// https://tools.ietf.org/html/rfc2231)
// There may also be parameters (e.g. "charset=utf-8") and extension values.
type ContentType struct {
	// Type and Subtype carry the media type, e.g. "text" and "html"
	Type, Subtype string
	// Params and Extensions hold optional parameter information
	Params []KV
}

// AsMediaRange converts this ContentType to a MediaRange.
// The default quality should be 1.
func (ct ContentType) AsMediaRange(quality float64) MediaRange {
	return MediaRange{
		ContentType: ct,
		Quality:     quality,
	}
}

// IsTextual returns true if the content represents a textual entity; false otherwise.
func (ct ContentType) IsTextual() bool {
	if ct.Type == "text" {
		return true
	}

	if ct.Type == "application" {
		return ct.Subtype == "json" ||
			ct.Subtype == "xml" ||
			strings.HasSuffix(ct.Subtype, "+xml") ||
			strings.HasSuffix(ct.Subtype, "+json")
	}

	if ct.Type == "model" {
		return strings.HasSuffix(ct.Subtype, "+xml") ||
			strings.HasSuffix(ct.Subtype, "+json")
	}

	if ct.Type == "image" || ct.Type == "message" {
		return strings.HasSuffix(ct.Subtype, "+xml")
	}

	return false
}

func (ct ContentType) writeTo(w io.StringWriter) {
	w.WriteString(ct.Type)
	w.WriteString("/")
	w.WriteString(ct.Subtype)
	for _, p := range ct.Params {
		w.WriteString(";")
		w.WriteString(p.Key)
		w.WriteString("=")
		w.WriteString(p.Value)
	}
}

func (ct ContentType) String() string {
	buf := &strings.Builder{}
	ct.writeTo(buf)
	return buf.String()
}

var starStar = ContentType{Type: "*", Subtype: "*"}

func ParseContentTypeFromHeaders(hdrs http.Header) ContentType {
	cts := hdrs[headername.ContentType]
	if len(cts) == 0 {
		return starStar
	}
	return ParseContentType(cts[0])
}

func ParseContentType(ct string) ContentType {
	if ct == "" {
		return starStar
	}

	valueAndParams := Split(ct, ";").TrimSpace()
	t, s := internal.Split1(valueAndParams[0], '/')
	return contentTypeOf(t, s, valueAndParams[1:])
}

// contentTypeOf builds a content type value with optional parameters.
// The parameters are passed in as literal strings, e.g. "charset=utf-8".
func contentTypeOf(typ, subtype string, paramKV []string) ContentType {
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
