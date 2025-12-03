package header

import (
	"io"
	"mime"
	"net/http"
	"strings"

	"github.com/rickb777/acceptable/headername"
)

// ContentType is a media type as defined in RFC-2045, RFC-2046, RFC-2231
// (https://tools.ietf.org/html/rfc2045, https://tools.ietf.org/html/rfc2046,
// https://tools.ietf.org/html/rfc2231)
// There may also be parameters (e.g. "charset=utf-8") and extension values.
type ContentType struct {
	// Type and Subtype carry the media type, e.g. "text" and "html"
	MediaType string
	// Params and Extensions hold optional parameter information
	Params []KV
}

// WithDefault returns "*/*" if ct has a blank media type.
func (ct ContentType) WithDefault() ContentType {
	if ct.MediaType == "" {
		ct.MediaType = "*/*"
	}
	return ct
}

func (ct ContentType) Split() (string, string) {
	t, s, _ := strings.Cut(ct.MediaType, "/")
	return t, s
}

func (ct ContentType) Type() string {
	t, _ := ct.Split()
	return t
}

func (ct ContentType) Subtype() string {
	_, s := ct.Split()
	return s
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
// Textual types are
//
//   - "text/*"
//   - "application/json"
//   - "application/xml"
//   - "application/*+json"
//   - "application/*+xml"
//   - "image/*+xml"
//   - "message/*+xml"
//   - "model/*+json"
//   - "model/*+xml"
//
// where "*" is a wildcard.
func (ct ContentType) IsTextual() bool {
	t, s := ct.Split()
	if t == "text" {
		return true
	}

	if t == "application" {
		return s == "json" ||
			s == "xml" ||
			strings.HasSuffix(s, "+xml") ||
			strings.HasSuffix(s, "+json")
	}

	if t == "model" {
		return strings.HasSuffix(s, "+xml") ||
			strings.HasSuffix(s, "+json")
	}

	if t == "image" || t == "message" {
		return strings.HasSuffix(s, "+xml")
	}

	return false
}

func (ct ContentType) writeTo(w io.StringWriter) {
	w.WriteString(ct.MediaType)
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

// ParseContentTypeFromHeaders gets the "Content-Type" header and returns
// its parsed value. An absent or malformed input yields a blank media type.
func ParseContentTypeFromHeaders(hdrs http.Header) ContentType {
	cts := hdrs[headername.ContentType]
	if len(cts) == 0 {
		return ContentType{}
	}
	return ParseContentType(cts[0])
}

// ParseContentType parses a content type value.
// An absent or malformed input yields a blank media type.
func ParseContentType(ct string) ContentType {
	if ct == "" {
		return ContentType{}
	}

	mt, params, err := mime.ParseMediaType(ct)
	if err != nil {
		return ContentType{}
	}

	var paramsKV []KV
	if len(params) > 0 {
		paramsKV = make([]KV, 0, len(params))
		for k, v := range params {
			paramsKV = append(paramsKV, KV{Key: k, Value: v})
		}
	}
	return ContentType{MediaType: mt, Params: paramsKV}
}
