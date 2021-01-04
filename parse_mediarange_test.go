package acceptable

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestParseAcceptHeader_parses_single(t *testing.T) {
	g := NewGomegaWithT(t)
	mr := ParseMediaRanges("application/json")

	g.Expect(len(mr)).To(Equal(1))
	g.Expect(mr[0].Type).To(Equal("application"))
	g.Expect(mr[0].Subtype).To(Equal("json"))
	g.Expect(mr[0].Quality).To(Equal(DefaultQuality))
}

func TestParseAcceptHeader_converts_mediaRange_to_lowercase(t *testing.T) {
	g := NewGomegaWithT(t)
	mr := ParseMediaRanges("Application/CEA")

	g.Expect(len(mr)).To(Equal(1))
	g.Expect(mr[0].Type).To(Equal("application"))
	g.Expect(mr[0].Subtype).To(Equal("cea"))
}

func TestParseAcceptHeader_defaults_quality_if_not_explicit(t *testing.T) {
	g := NewGomegaWithT(t)
	mr := ParseMediaRanges("text/plain")

	g.Expect(len(mr)).To(Equal(1))
	g.Expect(mr[0].Quality).To(Equal(DefaultQuality))
}

func TestParseAcceptHeader_should_parse_quality(t *testing.T) {
	g := NewGomegaWithT(t)
	mr := ParseMediaRanges("application/json; q=0.9")

	g.Expect(len(mr)).To(Equal(1))
	g.Expect(mr[0].Type).To(Equal("application"))
	g.Expect(mr[0].Subtype).To(Equal("json"))
	g.Expect(mr[0].Quality).To(BeNumerically("~", 0.9, 1e-4))
}

func TestParseAcceptHeader_sorts_by_decending_quality(t *testing.T) {
	g := NewGomegaWithT(t)
	mr := ParseMediaRanges("application/json;q=0.8, application/xml, application/*;q=0.1")

	g.Expect(len(mr)).To(Equal(3))

	g.Expect(mr[0].Type).To(Equal("application"))
	g.Expect(mr[0].Subtype).To(Equal("xml"))
	g.Expect(mr[0].Quality).To(Equal(DefaultQuality))

	g.Expect(mr[1].Type).To(Equal("application"))
	g.Expect(mr[1].Subtype).To(Equal("json"))
	g.Expect(mr[1].Quality).To(BeNumerically("~", 0.8, 1e-4))

	g.Expect(mr[2].Type).To(Equal("application"))
	g.Expect(mr[2].Subtype).To(Equal("*"))
	g.Expect(mr[2].Quality).To(BeNumerically("~", 0.1, 1e-4))
}

func TestMediaRanges_should_ignore_invalid_quality(t *testing.T) {
	g := NewGomegaWithT(t)
	mr := ParseMediaRanges("text/html;q=blah")

	g.Expect(len(mr)).To(Equal(1))
	g.Expect(mr[0].Type).To(Equal("text"))
	g.Expect(mr[0].Subtype).To(Equal("html"))
	g.Expect(mr[0].Quality).To(Equal(DefaultQuality))
	g.Expect(mr[0].Params).To(HaveLen(0))
}

// If more than one media range applies to a
// given type, the most specific reference has precedence
func TestMediaRanges_should_handle_precedence(t *testing.T) {
	g := NewGomegaWithT(t)
	// from https://tools.ietf.org/html/rfc7231#section-5.3.2
	c := "text/*, text/plain, text/plain;format=flowed, */*"
	mr := ParseMediaRanges(c)

	g.Expect(len(mr)).To(Equal(4))
	g.Expect(mr[0]).To(Equal(MediaRange{
		ContentType: ContentTypeOf("text", "plain", "format=flowed"),
		Quality:     DefaultQuality,
	}), c)
	g.Expect(mr[1]).To(Equal(MediaRange{
		ContentType: ContentTypeOf("text", "plain"),
		Quality:     DefaultQuality,
	}), c)
	g.Expect(mr[2]).To(Equal(MediaRange{
		ContentType: ContentTypeOf("text", "*"),
		Quality:     DefaultQuality,
	}), c)
	g.Expect(mr[3]).To(Equal(MediaRange{
		ContentType: ContentTypeOf("*", "*"),
		Quality:     DefaultQuality,
	}), c)
}

func TestMediaRanges_should_not_remove_accept_extension(t *testing.T) {
	g := NewGomegaWithT(t)
	mr := ParseMediaRanges("text/html; q=0.5; a=1;b=2")

	g.Expect(len(mr)).To(Equal(1))
	g.Expect(mr[0].Type).To(Equal("text"))
	g.Expect(mr[0].Subtype).To(Equal("html"))
	g.Expect(mr[0].Quality).To(Equal(0.5))
	g.Expect(mr[0].Params).To(BeEmpty())
	g.Expect(mr[0].Extensions).To(ConsistOf(KV{"a", "1"}, KV{"b", "2"}))
}

func TestMediaRanges_string(t *testing.T) {
	g := NewGomegaWithT(t)
	header := "text/html;level=1;q=0.9;a=1;b=2, text/html;q=0.5, text/*;q=0.3"
	mr := ParseMediaRanges(header)
	g.Expect(mr[0].Value()).To(Equal("text/html;level=1"))
	g.Expect(mr.String()).To(Equal(header))
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
		mr := ParseMediaRanges(c)
		g.Expect(5, len(mr))

		g.Expect(mr[0]).To(Equal(MediaRange{
			ContentType: ContentTypeOf("text", "html", "level=1"),
			Quality:     DefaultQuality,
		}), c)

		g.Expect(mr[1]).To(Equal(MediaRange{
			ContentType: ContentTypeOf("text", "html"),
			Quality:     0.7,
		}), c)

		g.Expect(mr[2]).To(Equal(MediaRange{
			ContentType: ContentTypeOf("*", "*"),
			Quality:     0.5,
		}), c)

		g.Expect(mr[3]).To(Equal(MediaRange{
			ContentType: ContentTypeOf("text", "html", "level=2"),
			Quality:     0.4,
		}), c)

		g.Expect(mr[4]).To(Equal(MediaRange{
			ContentType: ContentTypeOf("text", "*"),
			Quality:     0.3,
		}), c)
	}
}

func TestMediaRanges_should_ignore_case_of_quality_and_whitespace(t *testing.T) {
	g := NewGomegaWithT(t)
	mr := ParseMediaRanges("text/* ; q=0.3, text/html ; Q=0.7, text/html;level=2; q=0.4, */*; q=0.5")

	g.Expect(4).To(Equal(len(mr)))

	g.Expect("text/html").To(Equal(mr[0].Value()))
	g.Expect(0.7).To(Equal(mr[0].Quality))

	g.Expect("*/*").To(Equal(mr[1].Value()))
	g.Expect(0.5).To(Equal(mr[1].Quality))

	g.Expect("text/html;level=2").To(Equal(mr[2].Value()))
	g.Expect(0.4).To(Equal(mr[2].Quality))

	g.Expect("text/*").To(Equal(mr[3].Value()))
	g.Expect(0.3).To(Equal(mr[3].Quality))
}
