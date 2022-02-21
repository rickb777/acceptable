package offer_test

import (
	"fmt"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/rickb777/acceptable/data"
	"github.com/rickb777/acceptable/header"
	. "github.com/rickb777/acceptable/headername"
	"github.com/rickb777/acceptable/offer"
)

func TestApplyHeaders(t *testing.T) {
	g := NewWithT(t)

	// Given ...
	cases := []struct {
		str  string
		m    offer.Match
		hdrs map[string]string
		utf8 bool
	}{
		{
			str: "text/test; charset=windows-1252; lang=en vary=[Accept Accept-Language]",
			m: offer.Match{
				ContentType: header.ContentType{Type: "text", Subtype: "test"},
				Language:    "en",
				Charset:     "windows-1252",
				Data:        data.Of("data"),
				Vary:        []string{Accept, AcceptLanguage},
				Render:      offer.TXTProcessor(),
			},
			hdrs: map[string]string{
				ContentType:     "text/test;charset=windows-1252",
				ContentLanguage: "en",
				Vary:            "Accept, Accept-Language",
			},
			utf8: false,
		},
		{
			str: "application/xhtml+xml; charset=utf-8; lang=fr vary=[Accept Accept-Language]; no data; no renderer; not accepted",
			m: offer.Match{
				ContentType:        header.ContentType{Type: "application", Subtype: "xhtml+xml"},
				Language:           "fr",
				Charset:            "utf-8",
				Vary:               []string{Accept, AcceptLanguage},
				StatusCodeOverride: 400,
			},
			hdrs: map[string]string{
				ContentType:     "application/xhtml+xml;charset=utf-8",
				ContentLanguage: "fr",
				Vary:            "Accept, Accept-Language",
			},
			utf8: true,
		},
		{
			str: "image/png; charset=utf-8; lang=fr vary=[Accept]; no data; no renderer; not accepted",
			m: offer.Match{
				ContentType:        header.ContentType{Type: "image", Subtype: "png"},
				Language:           "fr",
				Charset:            "utf-8",
				Vary:               []string{Accept},
				StatusCodeOverride: 400,
			},
			hdrs: map[string]string{
				ContentType: "image/png",
				Vary:        Accept,
			},
			utf8: true,
		},
		{
			str: "text/plain; charset=utf-8; lang=fr vary=[]; no data; no renderer; not accepted",
			m: offer.Match{
				ContentType:        header.ContentType{Type: "text", Subtype: "plain"},
				Language:           "fr",
				Charset:            "utf-8",
				Vary:               nil,
				StatusCodeOverride: 400,
			},
			hdrs: map[string]string{
				ContentType:     "text/plain;charset=utf-8",
				ContentLanguage: "fr",
			},
			utf8: true,
		},
		{
			str: "application/octet-stream; charset=utf-8; lang=fr vary=[]; no data; no renderer",
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

	for i, c := range cases {
		rec := httptest.NewRecorder()

		// When ...
		w := c.m.ApplyHeaders(rec)

		// Then ...
		info := fmt.Sprintf("%d:%s", i, c.m)
		g.Expect(c.m.String()).To(Equal(c.str), info)
		if c.utf8 {
			g.Expect(w).To(BeIdenticalTo(rec), info)
		}
		g.Expect(rec.HeaderMap).To(HaveLen(len(c.hdrs)), info)
		for h, v := range c.hdrs {
			g.Expect(rec.Header().Get(h)).To(Equal(v), info)
		}
	}
}
