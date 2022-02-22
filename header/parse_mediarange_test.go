package header_test

import (
	"fmt"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/rickb777/acceptable/header"
)

func TestParseAcceptHeader_parses_single(t *testing.T) {
	g := NewGomegaWithT(t)
	mr := header.ParseMediaRanges("application/json")

	g.Expect(len(mr)).To(Equal(1))
	g.Expect(mr[0].Type).To(Equal("application"))
	g.Expect(mr[0].Subtype).To(Equal("json"))
	g.Expect(mr[0].Quality).To(Equal(header.DefaultQuality))
}

func TestParseAcceptHeader_converts_mediaRange_to_lowercase(t *testing.T) {
	g := NewGomegaWithT(t)
	mr := header.ParseMediaRanges("Application/CEA")

	g.Expect(len(mr)).To(Equal(1))
	g.Expect(mr[0].Type).To(Equal("application"))
	g.Expect(mr[0].Subtype).To(Equal("cea"))
}

func TestParseAcceptHeader_defaults_quality_if_not_explicit(t *testing.T) {
	g := NewGomegaWithT(t)
	mr := header.ParseMediaRanges("text/plain")

	g.Expect(len(mr)).To(Equal(1))
	g.Expect(mr[0].Quality).To(Equal(header.DefaultQuality))
}

func TestParseAcceptHeader_should_parse_quality(t *testing.T) {
	g := NewGomegaWithT(t)
	mr := header.ParseMediaRanges("application/json; q=0.9")

	g.Expect(len(mr)).To(Equal(1))
	g.Expect(mr[0].Type).To(Equal("application"))
	g.Expect(mr[0].Subtype).To(Equal("json"))
	g.Expect(mr[0].Quality).To(BeNumerically("~", 0.9, 1e-4))
}

func TestParseAcceptHeader_extension_can_omit_value(t *testing.T) {
	g := NewGomegaWithT(t)
	mr := header.ParseMediaRanges("application/json; q=0.9; label")

	g.Expect(len(mr)).To(Equal(1))
	g.Expect(mr[0].Type).To(Equal("application"))
	g.Expect(mr[0].Subtype).To(Equal("json"))
	g.Expect(mr[0].Params).To(ConsistOf(header.KV{Key: "label"}))
}

func TestParseAcceptHeader_sorts_by_decending_quality(t *testing.T) {
	g := NewGomegaWithT(t)
	mr := header.ParseMediaRanges("application/json;q=0.8, application/xml, application/*;q=0.1")

	g.Expect(len(mr)).To(Equal(3))

	g.Expect(mr[0].Type).To(Equal("application"))
	g.Expect(mr[0].Subtype).To(Equal("xml"))
	g.Expect(mr[0].Quality).To(Equal(header.DefaultQuality))

	g.Expect(mr[1].Type).To(Equal("application"))
	g.Expect(mr[1].Subtype).To(Equal("json"))
	g.Expect(mr[1].Quality).To(BeNumerically("~", 0.8, 1e-4))

	g.Expect(mr[2].Type).To(Equal("application"))
	g.Expect(mr[2].Subtype).To(Equal("*"))
	g.Expect(mr[2].Quality).To(BeNumerically("~", 0.1, 1e-4))
}

func TestMediaRanges_should_ignore_invalid_quality(t *testing.T) {
	g := NewGomegaWithT(t)
	mr := header.ParseMediaRanges("text/html;q=blah")

	g.Expect(len(mr)).To(Equal(1))
	g.Expect(mr[0].Type).To(Equal("text"))
	g.Expect(mr[0].Subtype).To(Equal("html"))
	g.Expect(mr[0].Quality).To(Equal(header.DefaultQuality))
	g.Expect(mr[0].Params).To(HaveLen(0))
}

// If more than one media range applies to a
// given type, the most specific reference has precedence
func TestMediaRanges_should_handle_precedence(t *testing.T) {
	g := NewGomegaWithT(t)
	// from https://tools.ietf.org/html/rfc7231#section-5.3.2
	cases := []string{
		"text/*, text/plain, text/plain;format=flowed, */*",
		"*/*, text/*, text/plain, text/plain;format=flowed",
		"text/plain;format=flowed, */*, text/*, text/plain",
		"text/plain, text/plain;format=flowed, */*, text/*",
	}
	for _, c := range cases {
		mr := header.ParseMediaRanges(c)

		g.Expect(len(mr)).To(Equal(4))
		g.Expect(mr[0]).To(Equal(header.MediaRange{
			ContentType: header.ParseContentType("text/plain; format=flowed"),
			Quality:     header.DefaultQuality,
		}), c)
		g.Expect(mr[1]).To(Equal(header.MediaRange{
			ContentType: header.ParseContentType("text/plain"),
			Quality:     header.DefaultQuality,
		}), c)
		g.Expect(mr[2]).To(Equal(header.MediaRange{
			ContentType: header.ParseContentType("text/*"),
			Quality:     header.DefaultQuality,
		}), c)
		g.Expect(mr[3]).To(Equal(header.MediaRange{
			ContentType: header.ParseContentType("*/*"),
			Quality:     header.DefaultQuality,
		}), c)
	}
}

func TestMediaRanges_should_not_remove_accept_extension(t *testing.T) {
	g := NewGomegaWithT(t)
	mr := header.ParseMediaRanges("text/html; q=0.5; a=1;b=2")

	g.Expect(len(mr)).To(Equal(1))
	g.Expect(mr[0].Type).To(Equal("text"))
	g.Expect(mr[0].Subtype).To(Equal("html"))
	g.Expect(mr[0].Quality).To(Equal(0.5))
	g.Expect(mr[0].Params).To(ConsistOf(header.KV{"a", "1"}, header.KV{"b", "2"}))
}

func TestMediaRanges_string(t *testing.T) {
	g := NewGomegaWithT(t)
	h := "text/html;level=1;a=1;b=2;q=0.9, text/html;q=0.5, text/*;q=0.3"
	mr := header.ParseMediaRanges(h)
	g.Expect(mr[0].Value()).To(Equal("text/html;level=1;a=1;b=2"))
	g.Expect(mr.String()).To(Equal(h))
}

func TestMediaRanges_should_handle_quality_precedence(t *testing.T) {
	g := NewGomegaWithT(t)
	cases := []string{
		// each example has a distinct quality for each part
		"text/*;q=0.3, text/html;q=0.7, text/html;level=1, text/html;level=2;q=0.4, */*;q=0.5",
		"text/html;q=0.7, text/html;level=1, text/html;level=2;q=0.4, */*;q=0.5, text/*;q=0.3",
		"text/html;level=1, text/html;level=2;q=0.4, */*;q=0.5, text/*;q=0.3, text/html;q=0.7",
		"text/html;level=2;q=0.4, */*;q=0.5, text/*;q=0.3, text/html;q=0.7, text/html;level=1",
	}
	for _, c := range cases {
		mr := header.ParseMediaRanges(c)
		g.Expect(5, len(mr))

		g.Expect(mr[0]).To(Equal(header.MediaRange{
			ContentType: header.ParseContentType("text/html; level=1"),
			Quality:     header.DefaultQuality,
		}), c)

		g.Expect(mr[1]).To(Equal(header.MediaRange{
			ContentType: header.ParseContentType("text/html"),
			Quality:     0.7,
		}), c)

		g.Expect(mr[2]).To(Equal(header.MediaRange{
			ContentType: header.ParseContentType("*/*"),
			Quality:     0.5,
		}), c)

		g.Expect(mr[3]).To(Equal(header.MediaRange{
			ContentType: header.ParseContentType("text/html; level=2"),
			Quality:     0.4,
		}), c)

		g.Expect(mr[4]).To(Equal(header.MediaRange{
			ContentType: header.ParseContentType("text/*"),
			Quality:     0.3,
		}), c)
	}
}

func TestMediaRanges_should_ignore_case_of_quality_and_whitespace(t *testing.T) {
	g := NewGomegaWithT(t)
	mr := header.ParseMediaRanges("text/* ; q=0.3, TEXT/html ; Q=0.7, text/html;level=2; q=0.4, */*; q=0.5")

	g.Expect(len(mr)).To(Equal(4))

	g.Expect(mr[0].Value()).To(Equal("text/html"))
	g.Expect(mr[0].Quality).To(Equal(0.7))

	g.Expect(mr[1].Value()).To(Equal("*/*"))
	g.Expect(mr[1].Quality).To(Equal(0.5))

	g.Expect(mr[2].Value()).To(Equal("text/html;level=2"))
	g.Expect(mr[2].Quality).To(Equal(0.4))

	g.Expect(mr[3].Value()).To(Equal("text/*"))
	g.Expect(mr[3].Quality).To(Equal(0.3))
}

func ExampleParseMediaRanges() {
	mrs := header.ParseMediaRanges("text/* ; q=0.3, TEXT/html ; Q=0.7, text/html;level=2; q=0.4, */*; q=0.5")

	for i, mr := range mrs {
		fmt.Printf("mr%d = %s\n", i, mr)
	}
	// Output:
	// mr0 = text/html;q=0.7
	// mr1 = */*;q=0.5
	// mr2 = text/html;level=2;q=0.4
	// mr3 = text/*;q=0.3
}
