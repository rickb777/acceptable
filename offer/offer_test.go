package offer

import (
	"fmt"
	"strings"
	"testing"

	"github.com/rickb777/acceptable/data"
	"github.com/rickb777/acceptable/header"
	"github.com/rickb777/expect"
)

func Test_offer_construction(t *testing.T) {
	base := Of(nil, "")

	cases := map[string]struct {
		o  Offer
		el string
		nd int
	}{
		"01.Accept: */*": {o: base, el: "*", nd: 0},

		"02.Accept: text/*": {o: Of(nil, "text/*"), el: "*", nd: 0},

		"03.Accept: a/b": {o: Of(nil, "a/b").With(nil, "*"), el: "*", nd: 0},

		"04.Accept: */*. Accept-Language: *": {o: base.With("foo", "*"), el: "*", nd: 1},

		"05.Accept: a/b. Accept-Language: en": {o: Of(nil, "a/b").With("foo", "en"), el: "en", nd: 1},

		"06.Accept: a/b. Accept-Language: en": {o: Of(nil, "a/b").With(nil, "en"), el: "en", nd: 1},

		"07.Accept: */*. Accept-Language: en,fr,pt": {o: base.With("foo", "en").With("bar", "fr").With("baz", "pt"), el: "en,fr,pt", nd: 3},

		"08.Accept: a/b. Accept-Language: en,fr,pt": {o: Of(nil, "a/b").With("foo", "en").With("bar", "fr").With("baz", "pt"), el: "en,fr,pt", nd: 3},

		"09.Accept: a/b. Accept-Language: en,fr": {o: Of(nil, "a/b").With("foo", "en", "fr"), el: "en,fr", nd: 2},

		"10.Accept: a/b. Accept-Language: en,fr": {o: Of(nil, "a/b").With(data.Of("foo"), "en", "fr"), el: "en,fr", nd: 2},
	}

	for s, c := range cases {
		// invariants
		expect.String(base.String()).I(s).ToBe(t, "Accept: */*")
		expect.Slice(base.Langs).I(s).ToBe(t, "*")
		expect.Map(base.data).I(s).ToBeEmpty(t)

		expect.String(c.o.String()).I(s).ToBe(t, s[3:])
		expect.Slice(c.o.Langs).I(s).ToBe(t, strings.Split(c.el, ",")...)
		expect.Map(c.o.data).I(s).ToHaveLength(t, c.nd)

		for l, d := range c.o.data {
			expect.String(fmt.Sprintf("%T", d)).I("%v|%v", s, l).
				ToBe(nil, "*data.Value").Or().ToBe(t, "offer.empty")
		}
	}
}

func Test_offer_with(t *testing.T) {
	o1 := Of(nil, "text/plain")
	o2 := o1.With("foo", "en")
	o3 := o2.With("bar", "fr")
	o4 := o3.With("baz", "pt")

	expect.Slice(o1.Langs).ToHaveLength(t, 1)
	expect.Slice(o1.Langs).ToBe(t, "*")
	expect.Slice(o2.Langs).ToHaveLength(t, 1)
	expect.Slice(o2.Langs).ToBe(t, "en")
	expect.Slice(o3.Langs).ToHaveLength(t, 2)
	expect.Slice(o3.Langs).ToBe(t, "en", "fr")
	expect.Slice(o4.Langs).ToHaveLength(t, 3)
	expect.Slice(o4.Langs).ToBe(t, "en", "fr", "pt")

	expect.Map(o1.data).ToHaveLength(t, 0)
	expect.Map(o2.data).ToHaveLength(t, 1)
	expect.Map(o3.data).ToHaveLength(t, 2)
	expect.Map(o4.data).ToHaveLength(t, 3)
}

func TestOffersAllEmpty(t *testing.T) {
	o1 := Of(nil, "text/plain")
	o2 := Of(nil, "image/png")

	e := Offers{o1, o2}.AllEmpty()
	expect.Bool(e).ToBeTrue(t)

	o3 := Of(nil, "text/plain").With("foo", "*")

	e = Offers{o2, o3}.AllEmpty()
	expect.Bool(e).ToBeFalse(t)
}

func TestBuildMatch(t *testing.T) {
	txt := TXTProcessor()
	cases := []struct {
		o        Offer
		accepted header.ContentType
		m        Match
	}{
		{
			o:        Of(txt, "text/*"),
			accepted: header.ContentType{MediaType: "text/plain"},
			m: Match{
				ContentType:        header.ContentType{MediaType: "text/plain"},
				Language:           "en",
				Data:               nil,
				StatusCodeOverride: 0,
			},
		},
		{
			o:        Of(txt, "*/*"),
			accepted: header.ContentType{MediaType: "text/plain"},
			m: Match{
				ContentType:        header.ContentType{MediaType: "text/plain"},
				Language:           "en",
				Data:               nil,
				StatusCodeOverride: 0,
			},
		},
		{
			o:        Of(txt, "text/*"),
			accepted: header.ContentType{MediaType: "text/*"},
			m: Match{
				ContentType:        header.ContentType{MediaType: "text/plain"},
				Language:           "en",
				Data:               nil,
				StatusCodeOverride: 0,
			},
		},
		{
			o:        Of(txt, "*/*"),
			accepted: header.ContentType{MediaType: "*/*"},
			m: Match{
				ContentType:        header.ContentType{MediaType: "application/octet-stream"},
				Language:           "en",
				Data:               nil,
				StatusCodeOverride: 0,
			},
		},
		{
			o:        Of(txt, "text/plain").With("foo", "fr").With("bar", "en"),
			accepted: header.ContentType{MediaType: "text/*"},
			m: Match{
				ContentType:        header.ContentType{MediaType: "text/plain"},
				Language:           "en",
				Data:               data.Of("bar"),
				StatusCodeOverride: 0,
			},
		},
	}

	for _, c := range cases {
		m := c.o.BuildMatch(c.accepted, "en", 0)
		m.Render = nil // comparing functions would always fail
		expect.Any(*m).I(c.o).ToBe(t, c.m)
	}
}
