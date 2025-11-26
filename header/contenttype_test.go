package header_test

import (
	"net/http"
	"testing"

	. "github.com/rickb777/acceptable/header"
	"github.com/rickb777/acceptable/headername"
	"github.com/rickb777/expect"
)

func TestParseContentTypeFromHeaders(t *testing.T) {
	hdrs := make(http.Header)

	ct1 := ParseContentTypeFromHeaders(hdrs)

	expect.Any(ct1).ToBe(t, ContentType{MediaType: "*/*"})

	hdrs.Set(headername.ContentType, "text/plain")

	ct2 := ParseContentTypeFromHeaders(hdrs)

	expect.Any(ct2).ToBe(t, ContentType{MediaType: "text/plain"})
}

func TestParseContentType(t *testing.T) {
	// blank value is treated as star-star
	expect.Any(ParseContentType("")).ToBe(t, ContentType{MediaType: "*/*"})

	// illegal value is treated as star-star
	expect.Any(ParseContentType("/")).ToBe(t, ContentType{MediaType: "*/*"})

	// illegal value is treated as star-star
	expect.Any(ParseContentType("/plain")).ToBe(t, ContentType{MediaType: "*/*"})

	// error case handled silently
	expect.Any(ParseContentType("text/")).ToBe(t, ContentType{MediaType: "*/*"})

	expect.Any(ParseContentType("text/plain")).ToBe(t, ContentType{MediaType: "text/plain"})

	expect.Any(ParseContentType("text/html; charset=utf-8")).ToBe(t, ContentType{
		MediaType: "text/html",
		Params: []KV{{
			Key:   "charset",
			Value: "utf-8",
		}},
	})
}

func TestContentType_IsTextual(t *testing.T) {
	cases := []ContentType{
		{MediaType: "text/plain"},
		{MediaType: "application/json"},
		{MediaType: "application/geo+json"},
		{MediaType: "application/xml"},
		{MediaType: "application/xv+xml"},
		{MediaType: "image/svg+xml"},
		{MediaType: "message/imdn+xml"},
		{MediaType: "model/x3d+xml"},
		{MediaType: "model/gltf+json"},
	}
	for _, c := range cases {
		expect.Bool(c.IsTextual()).I(c.String).ToBeTrue(t)
	}
	expect.Bool(ContentType{MediaType: "video/mp4"}.IsTextual()).ToBeFalse(t)
}

func TestContentType_String(t *testing.T) {
	ct := ContentType{
		MediaType: "text/html",
		Params: []KV{
			{Key: "charset", Value: "utf-8"},
			{Key: "level", Value: "1"},
		},
	}

	expect.String(ct.String()).ToBe(t, "text/html;charset=utf-8;level=1")
}
