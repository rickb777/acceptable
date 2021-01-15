package acceptable_test

import (
	"net/http/httptest"
	"testing"

	"github.com/onsi/gomega"
	"github.com/rickb777/acceptable"
)

func TestApplyHeaders(t *testing.T) {
	g := gomega.NewWithT(t)

	// Given ...
	cases := []struct {
		m    acceptable.Match
		hdrs map[string]string
		utf8 bool
	}{
		{
			m: acceptable.Match{
				Type:     "text",
				Subtype:  "test",
				Language: "en",
				Charset:  "windows-1252",
				Vary:     []string{"Accept", "Accept-Language"},
			},
			hdrs: map[string]string{
				"Content-Type":     "text/test;charset=windows-1252",
				"Content-Language": "en",
				"Vary":             "Accept, Accept-Language",
			},
			utf8: false,
		},
		{
			m: acceptable.Match{
				Type:     "application",
				Subtype:  "xhtml+xml",
				Language: "fr",
				Charset:  "utf-8",
				Vary:     []string{"Accept", "Accept-Language"},
			},
			hdrs: map[string]string{
				"Content-Type":     "application/xhtml+xml;charset=utf-8",
				"Content-Language": "fr",
				"Vary":             "Accept, Accept-Language",
			},
			utf8: true,
		},
		{
			m: acceptable.Match{
				Type:     "image",
				Subtype:  "png",
				Language: "fr",
				Charset:  "utf-8",
				Vary:     []string{"Accept"},
			},
			hdrs: map[string]string{
				"Content-Type": "image/png",
				"Vary":         "Accept",
			},
			utf8: true,
		},
		{
			m: acceptable.Match{
				Type:     "text",
				Subtype:  "plain",
				Language: "fr",
				Charset:  "utf-8",
				Vary:     nil,
			},
			hdrs: map[string]string{
				"Content-Type":     "text/plain;charset=utf-8",
				"Content-Language": "fr",
			},
			utf8: true,
		},
		{
			m: acceptable.Match{
				Type:     "application",
				Subtype:  "octet-stream",
				Language: "fr",
				Charset:  "utf-8",
				Vary:     nil,
			},
			hdrs: map[string]string{
				"Content-Type": "application/octet-stream",
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
