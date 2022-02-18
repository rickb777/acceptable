package offer_test

import (
	"net/http/httptest"
	"testing"

	"github.com/onsi/gomega"
	"github.com/rickb777/acceptable/header"
	. "github.com/rickb777/acceptable/headername"
	"github.com/rickb777/acceptable/offer"
)

func TestApplyHeaders(t *testing.T) {
	g := gomega.NewWithT(t)

	// Given ...
	cases := []struct {
		m    offer.Match
		hdrs map[string]string
		utf8 bool
	}{
		{
			m: offer.Match{
				ContentType: header.ContentType{Type: "text", Subtype: "test"},
				Language:    "en",
				Charset:     "windows-1252",
				Vary:        []string{Accept, AcceptLanguage},
			},
			hdrs: map[string]string{
				ContentType:     "text/test;charset=windows-1252",
				ContentLanguage: "en",
				Vary:            "Accept, Accept-Language",
			},
			utf8: false,
		},
		{
			m: offer.Match{
				ContentType: header.ContentType{Type: "application", Subtype: "xhtml+xml"},
				Language:    "fr",
				Charset:     "utf-8",
				Vary:        []string{Accept, AcceptLanguage},
			},
			hdrs: map[string]string{
				ContentType:     "application/xhtml+xml;charset=utf-8",
				ContentLanguage: "fr",
				Vary:            "Accept, Accept-Language",
			},
			utf8: true,
		},
		{
			m: offer.Match{
				ContentType: header.ContentType{Type: "image", Subtype: "png"},
				Language:    "fr",
				Charset:     "utf-8",
				Vary:        []string{Accept},
			},
			hdrs: map[string]string{
				ContentType: "image/png",
				Vary:        Accept,
			},
			utf8: true,
		},
		{
			m: offer.Match{
				ContentType: header.ContentType{Type: "text", Subtype: "plain"},
				Language:    "fr",
				Charset:     "utf-8",
				Vary:        nil,
			},
			hdrs: map[string]string{
				ContentType:     "text/plain;charset=utf-8",
				ContentLanguage: "fr",
			},
			utf8: true,
		},
		{
			m: offer.Match{
				ContentType: header.ContentType{Type: "application", Subtype: "octet-stream"},
				Language:    "fr",
				Charset:     "utf-8",
				Vary:        nil,
			},
			hdrs: map[string]string{
				ContentType: "application/octet-stream",
			},
			utf8: true,
		},
	}

	for _, c := range cases {
		rec := httptest.NewRecorder()

		// When ...
		w := c.m.ApplyHeaders(rec)

		// Then ...
		info := c.m.String()
		if c.utf8 {
			g.Expect(w).To(gomega.BeIdenticalTo(rec), info)
		}
		g.Expect(rec.HeaderMap).To(gomega.HaveLen(len(c.hdrs)), info)
		for h, v := range c.hdrs {
			g.Expect(rec.Header().Get(h)).To(gomega.Equal(v), info)
		}
	}
}
