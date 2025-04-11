package header_test

import (
	"fmt"
	"testing"

	"github.com/rickb777/acceptable/header"
	"github.com/rickb777/expect"
)

func TestParseAcceptHeader_parses_single(t *testing.T) {
	mr := header.ParseMediaRanges("application/json")

	expect.Number(len(mr)).ToBe(t, 1)
	expect.String(mr[0].Type).ToBe(t, "application")
	expect.String(mr[0].Subtype).ToBe(t, "json")
	expect.Any(mr[0].Quality).ToBe(t, header.DefaultQuality)
}

func TestParseAcceptHeader_converts_mediaRange_to_lowercase(t *testing.T) {
	mr := header.ParseMediaRanges("Application/CEA")

	expect.Number(len(mr)).ToBe(t, 1)
	expect.String(mr[0].Type).ToBe(t, "application")
	expect.String(mr[0].Subtype).ToBe(t, "cea")
}

func TestParseAcceptHeader_defaults_quality_if_not_explicit(t *testing.T) {
	mr := header.ParseMediaRanges("text/plain")

	expect.Number(len(mr)).ToBe(t, 1)
	expect.Any(mr[0].Quality).ToBe(t, header.DefaultQuality)
}

func TestParseAcceptHeader_should_parse_quality(t *testing.T) {
	mr := header.ParseMediaRanges("application/json; q=0.9")

	expect.Number(len(mr)).ToBe(t, 1)
	expect.String(mr[0].Type).ToBe(t, "application")
	expect.String(mr[0].Subtype).ToBe(t, "json")
	expect.Any(mr[0].Quality).ToBe(t, 0.9) // 1e-4)
}

func TestParseAcceptHeader_extension_can_omit_value(t *testing.T) {
	mr := header.ParseMediaRanges("application/json; q=0.9; label")

	expect.Number(len(mr)).ToBe(t, 1)
	expect.String(mr[0].Type).ToBe(t, "application")
	expect.String(mr[0].Subtype).ToBe(t, "json")
	expect.Slice(mr[0].Params).ToBe(t, header.KV{Key: "label"})
}

func TestParseAcceptHeader_sorts_by_decending_quality(t *testing.T) {
	mr := header.ParseMediaRanges("application/json;q=0.8, application/xml, application/*;q=0.1")

	expect.Number(len(mr)).ToBe(t, 3)

	expect.String(mr[0].Type).ToBe(t, "application")
	expect.String(mr[0].Subtype).ToBe(t, "xml")
	expect.Any(mr[0].Quality).ToBe(t, header.DefaultQuality)

	expect.String(mr[1].Type).ToBe(t, "application")
	expect.String(mr[1].Subtype).ToBe(t, "json")
	expect.Any(mr[1].Quality).ToBe(t, 0.8) // 1e-4

	expect.String(mr[2].Type).ToBe(t, "application")
	expect.String(mr[2].Subtype).ToBe(t, "*")
	expect.Any(mr[2].Quality).ToBe(t, 0.1) // 1e-4
}

func TestMediaRanges_should_ignore_invalid_quality(t *testing.T) {
	mr := header.ParseMediaRanges("text/html;q=blah")

	expect.Number(len(mr)).ToBe(t, 1)
	expect.String(mr[0].Type).ToBe(t, "text")
	expect.String(mr[0].Subtype).ToBe(t, "html")
	expect.Any(mr[0].Quality).ToBe(t, header.DefaultQuality)
	expect.Slice(mr[0].Params).ToHaveLength(t, 0)
}

// If more than one media range applies to a
// given type, the most specific reference has precedence
func TestMediaRanges_should_handle_precedence(t *testing.T) {
	// from https://tools.ietf.org/html/rfc7231#section-5.3.2
	cases := []string{
		"text/*, text/plain, text/plain;format=flowed, */*",
		"*/*, text/*, text/plain, text/plain;format=flowed",
		"text/plain;format=flowed, */*, text/*, text/plain",
		"text/plain, text/plain;format=flowed, */*, text/*",
	}
	for _, c := range cases {
		mr := header.ParseMediaRanges(c)

		expect.Number(len(mr)).ToBe(t, 4)
		expect.Any(mr[0]).I(c).ToBe(t, header.MediaRange{
			ContentType: header.ParseContentType("text/plain; format=flowed"),
			Quality:     header.DefaultQuality,
		})
		expect.Any(mr[1]).I(c).ToBe(t, header.MediaRange{
			ContentType: header.ParseContentType("text/plain"),
			Quality:     header.DefaultQuality,
		})
		expect.Any(mr[2]).I(c).ToBe(t, header.MediaRange{
			ContentType: header.ParseContentType("text/*"),
			Quality:     header.DefaultQuality,
		})
		expect.Any(mr[3]).I(c).ToBe(t, header.MediaRange{
			ContentType: header.ParseContentType("*/*"),
			Quality:     header.DefaultQuality,
		})
	}
}

func TestMediaRanges_should_not_remove_accept_extension(t *testing.T) {
	mr := header.ParseMediaRanges("text/html; q=0.5; a=1;b=2")

	expect.Number(len(mr)).ToBe(t, 1)
	expect.String(mr[0].Type).ToBe(t, "text")
	expect.String(mr[0].Subtype).ToBe(t, "html")
	expect.Any(mr[0].Quality).ToBe(t, 0.5)
	expect.Slice(mr[0].Params).ToBe(t, header.KV{"a", "1"}, header.KV{"b", "2"})
}

func TestMediaRanges_string(t *testing.T) {
	h := "text/html;level=1;a=1;b=2;q=0.9, text/html;q=0.5, text/*;q=0.3"
	mr := header.ParseMediaRanges(h)
	expect.String(mr[0].Value()).ToBe(t, "text/html;level=1;a=1;b=2")
	expect.String(mr.String()).ToBe(t, h)
}

func TestMediaRanges_should_handle_quality_precedence(t *testing.T) {
	cases := []string{
		// each example has a distinct quality for each part
		"text/*;q=0.3, text/html;q=0.7, text/html;level=1, text/html;level=2;q=0.4, */*;q=0.5",
		"text/html;q=0.7, text/html;level=1, text/html;level=2;q=0.4, */*;q=0.5, text/*;q=0.3",
		"text/html;level=1, text/html;level=2;q=0.4, */*;q=0.5, text/*;q=0.3, text/html;q=0.7",
		"text/html;level=2;q=0.4, */*;q=0.5, text/*;q=0.3, text/html;q=0.7, text/html;level=1",
	}
	for _, c := range cases {
		mr := header.ParseMediaRanges(c)
		expect.Number(len(mr)).ToBe(t, 5)

		expect.Any(mr[0]).I(c).ToBe(t, header.MediaRange{
			ContentType: header.ParseContentType("text/html; level=1"),
			Quality:     header.DefaultQuality,
		})

		expect.Any(mr[1]).I(c).ToBe(t, header.MediaRange{
			ContentType: header.ParseContentType("text/html"),
			Quality:     0.7,
		})

		expect.Any(mr[2]).I(c).ToBe(t, header.MediaRange{
			ContentType: header.ParseContentType("*/*"),
			Quality:     0.5,
		})

		expect.Any(mr[3]).I(c).ToBe(t, header.MediaRange{
			ContentType: header.ParseContentType("text/html; level=2"),
			Quality:     0.4,
		})

		expect.Any(mr[4]).I(c).ToBe(t, header.MediaRange{
			ContentType: header.ParseContentType("text/*"),
			Quality:     0.3,
		})
	}
}

func TestMediaRanges_should_ignore_case_of_quality_and_whitespace(t *testing.T) {
	mr := header.ParseMediaRanges("text/* ; q=0.3, TEXT/html ; Q=0.7, text/html;level=2; q=0.4, */*; q=0.5")

	expect.Number(len(mr)).ToBe(t, 4)

	expect.String(mr[0].Value()).ToBe(t, "text/html")
	expect.Any(mr[0].Quality).ToBe(t, 0.7)

	expect.String(mr[1].Value()).ToBe(t, "*/*")
	expect.Any(mr[1].Quality).ToBe(t, 0.5)

	expect.String(mr[2].Value()).ToBe(t, "text/html;level=2")
	expect.Any(mr[2].Quality).ToBe(t, 0.4)

	expect.String(mr[3].Value()).ToBe(t, "text/*")
	expect.Any(mr[3].Quality).ToBe(t, 0.3)
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
