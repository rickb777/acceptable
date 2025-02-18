package header_test

import (
	"net/http"
	"testing"

	"github.com/rickb777/acceptable/headername"

	"github.com/onsi/gomega"
	. "github.com/rickb777/acceptable/header"
)

func TestParseContentTypeFromHeaders(t *testing.T) {
	g := gomega.NewWithT(t)

	hdrs := make(http.Header)

	ct1 := ParseContentTypeFromHeaders(hdrs)

	g.Expect(ct1).To(gomega.Equal(ContentType{
		Type:    "*",
		Subtype: "*",
	}))

	hdrs.Set(headername.ContentType, "text/plain")

	ct2 := ParseContentTypeFromHeaders(hdrs)

	g.Expect(ct2).To(gomega.Equal(ContentType{
		Type:    "text",
		Subtype: "plain",
	}))
}

func TestParseContentType(t *testing.T) {
	g := gomega.NewWithT(t)

	// blank value is treated as star-star
	g.Expect(ParseContentType("")).To(gomega.Equal(ContentType{
		Type:    "*",
		Subtype: "*",
	}))

	// illegal value is treated as star-star
	g.Expect(ParseContentType("/")).To(gomega.Equal(ContentType{
		Type:    "*",
		Subtype: "*",
	}))

	// illegal value is treated as star-star
	g.Expect(ParseContentType("/plain")).To(gomega.Equal(ContentType{
		Type:    "*",
		Subtype: "*",
	}))

	// error case handled silently
	g.Expect(ParseContentType("text/")).To(gomega.Equal(ContentType{
		Type:    "text",
		Subtype: "*",
	}))

	g.Expect(ParseContentType("text/plain")).To(gomega.Equal(ContentType{
		Type:    "text",
		Subtype: "plain",
	}))

	g.Expect(ParseContentType("text/html; charset=utf-8")).To(gomega.Equal(ContentType{
		Type:    "text",
		Subtype: "html",
		Params: []KV{{
			Key:   "charset",
			Value: "utf-8",
		}},
	}))
}

func TestContentType_IsTextual(t *testing.T) {
	g := gomega.NewWithT(t)

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
		g.Expect(c.IsTextual()).To(gomega.BeTrue(), c.String())
	}
	g.Expect(ContentType{Type: "video", Subtype: "mp4"}.IsTextual()).To(gomega.BeFalse())
}

func TestContentType_String(t *testing.T) {
	g := gomega.NewWithT(t)

	ct := ContentType{
		Type:    "text",
		Subtype: "html",
		Params: []KV{
			{Key: "charset", Value: "utf-8"},
			{Key: "level", Value: "1"},
		},
	}

	g.Expect(ct.String()).To(gomega.Equal("text/html;charset=utf-8;level=1"))
}
