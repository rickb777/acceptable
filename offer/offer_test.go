package offer

import (
	"fmt"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/rickb777/acceptable/data"
	"github.com/rickb777/acceptable/header"
)

func Test_offer_construction(t *testing.T) {
	g := NewWithT(t)

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
		g.Expect(base.String()).To(Equal("Accept: */*"), s)
		g.Expect(base.Langs).To(ConsistOf("*"), s)
		g.Expect(len(base.data)).To(Equal(0), s)

		g.Expect(c.o.String()).To(Equal(s[3:]), s)
		g.Expect(c.o.Langs).To(ConsistOf(strings.Split(c.el, ",")), s)
		g.Expect(len(c.o.data)).To(Equal(c.nd), s)

		for l, d := range c.o.data {
			g.Expect(fmt.Sprintf("%T", d)).To(
				Or(
					Equal("*data.Value"),
					Equal("offer.empty"),
				), s+"|"+l)
		}
	}
}

func Test_offer_with(t *testing.T) {
	g := NewWithT(t)

	o1 := Of(nil, "text/plain")
	o2 := o1.With("foo", "en")
	o3 := o2.With("bar", "fr")
	o4 := o3.With("baz", "pt")

	g.Expect(o1.Langs).To(HaveLen(1))
	g.Expect(o1.Langs).To(ConsistOf("*"))
	g.Expect(o2.Langs).To(HaveLen(1))
	g.Expect(o2.Langs).To(ConsistOf("en"))
	g.Expect(o3.Langs).To(HaveLen(2))
	g.Expect(o3.Langs).To(ConsistOf("en", "fr"))
	g.Expect(o4.Langs).To(HaveLen(3))
	g.Expect(o4.Langs).To(ConsistOf("en", "fr", "pt"))

	g.Expect(o1.data).To(HaveLen(0))
	g.Expect(o2.data).To(HaveLen(1))
	g.Expect(o3.data).To(HaveLen(2))
	g.Expect(o4.data).To(HaveLen(3))
}

func TestOffersAllEmpty(t *testing.T) {
	g := NewWithT(t)

	o1 := Of(nil, "text/plain")
	o2 := Of(nil, "image/png")

	e := Offers{o1, o2}.AllEmpty()
	g.Expect(e).To(BeTrue())

	o3 := Of(nil, "text/plain").With("foo", "*")

	e = Offers{o2, o3}.AllEmpty()
	g.Expect(e).To(BeFalse())
}

func TestBuildMatch(t *testing.T) {
	g := NewWithT(t)

	txt := TXTProcessor()
	cases := []struct {
		o        Offer
		accepted header.ContentType
		m        Match
	}{
		{
			o:        Of(txt, "text/*"),
			accepted: header.ContentType{Type: "text", Subtype: "plain"},
			m: Match{
				ContentType:        header.ContentType{Type: "text", Subtype: "plain"},
				Language:           "en",
				Data:               nil,
				StatusCodeOverride: 0,
			},
		},
		{
			o:        Of(txt, "*/*"),
			accepted: header.ContentType{Type: "text", Subtype: "plain"},
			m: Match{
				ContentType:        header.ContentType{Type: "text", Subtype: "plain"},
				Language:           "en",
				Data:               nil,
				StatusCodeOverride: 0,
			},
		},
		{
			o:        Of(txt, "text/*"),
			accepted: header.ContentType{Type: "text", Subtype: "*"},
			m: Match{
				ContentType:        header.ContentType{Type: "text", Subtype: "plain"},
				Language:           "en",
				Data:               nil,
				StatusCodeOverride: 0,
			},
		},
		{
			o:        Of(txt, "*/*"),
			accepted: header.ContentType{Type: "*", Subtype: "*"},
			m: Match{
				ContentType:        header.ContentType{Type: "application", Subtype: "octet-stream"},
				Language:           "en",
				Data:               nil,
				StatusCodeOverride: 0,
			},
		},
		{
			o:        Of(txt, "text/plain").With("foo", "fr").With("bar", "en"),
			accepted: header.ContentType{Type: "text", Subtype: "*"},
			m: Match{
				ContentType:        header.ContentType{Type: "text", Subtype: "plain"},
				Language:           "en",
				Data:               data.Of("bar"),
				StatusCodeOverride: 0,
			},
		},
	}

	for _, c := range cases {
		m := c.o.BuildMatch(c.accepted, "en", 0)
		m.Render = nil // comparing functions would always fail
		g.Expect(*m).To(Equal(c.m), c.o.String())
	}
}
