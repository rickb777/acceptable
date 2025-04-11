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

	expect.Any(ct1).ToBe(t, ContentType{
		Type:    "*",
		Subtype: "*",
	})

	hdrs.Set(headername.ContentType, "text/plain")

	ct2 := ParseContentTypeFromHeaders(hdrs)

	expect.Any(ct2).ToBe(t, ContentType{
		Type:    "text",
		Subtype: "plain",
	})
}

func TestParseContentType(t *testing.T) {
	// blank value is treated as star-star
	expect.Any(ParseContentType("")).ToBe(t, ContentType{
		Type:    "*",
		Subtype: "*",
	})

	// illegal value is treated as star-star
	expect.Any(ParseContentType("/")).ToBe(t, ContentType{
		Type:    "*",
		Subtype: "*",
	})

	// illegal value is treated as star-star
	expect.Any(ParseContentType("/plain")).ToBe(t, ContentType{
		Type:    "*",
		Subtype: "*",
	})

	// error case handled silently
	expect.Any(ParseContentType("text/")).ToBe(t, ContentType{
		Type:    "text",
		Subtype: "*",
	})

	expect.Any(ParseContentType("text/plain")).ToBe(t, ContentType{
		Type:    "text",
		Subtype: "plain",
	})

	expect.Any(ParseContentType("text/html; charset=utf-8")).ToBe(t, ContentType{
		Type:    "text",
		Subtype: "html",
		Params: []KV{{
			Key:   "charset",
			Value: "utf-8",
		}},
	})
}

func TestContentType_IsTextual(t *testing.T) {
	cases := []ContentType{
		{Type: "text", Subtype: "plain"},
		{Type: "application", Subtype: "json"},
		{Type: "application", Subtype: "geo+json"},
		{Type: "application", Subtype: "xml"},
		{Type: "application", Subtype: "xv+xml"},
		{Type: "image", Subtype: "svg+xml"},
		{Type: "message", Subtype: "imdn+xml"},
		{Type: "model", Subtype: "x3d+xml"},
		{Type: "model", Subtype: "gltf+json"},
	}
	for _, c := range cases {
		expect.Bool(c.IsTextual()).I(c.String).ToBeTrue(t)
	}
	expect.Bool(ContentType{Type: "video", Subtype: "mp4"}.IsTextual()).ToBeFalse(t)
}

func TestContentType_String(t *testing.T) {
	ct := ContentType{
		Type:    "text",
		Subtype: "html",
		Params: []KV{
			{Key: "charset", Value: "utf-8"},
			{Key: "level", Value: "1"},
		},
	}

	expect.String(ct.String()).ToBe(t, "text/html;charset=utf-8;level=1")
}
