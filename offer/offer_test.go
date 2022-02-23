package offer

import (
	"fmt"
	"strings"
	"testing"

	"github.com/onsi/gomega"
	"github.com/rickb777/acceptable/data"
)

func Test_offer_construction(t *testing.T) {
	g := gomega.NewWithT(t)

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
		g.Expect(base.String()).To(gomega.Equal("Accept: */*"), s)
		g.Expect(base.Langs).To(gomega.ConsistOf("*"), s)
		g.Expect(len(base.data)).To(gomega.Equal(0), s)

		g.Expect(c.o.String()).To(gomega.Equal(s[3:]), s)
		g.Expect(c.o.Langs).To(gomega.ConsistOf(strings.Split(c.el, ",")), s)
		g.Expect(len(c.o.data)).To(gomega.Equal(c.nd), s)

		for l, d := range c.o.data {
			g.Expect(fmt.Sprintf("%T", d)).To(
				gomega.Or(
					gomega.Equal("*data.Value"),
					gomega.Equal("offer.empty"),
				), s+"|"+l)
		}
	}
}

func Test_offer_with(t *testing.T) {
	g := gomega.NewWithT(t)

	o1 := Of(nil, "text/plain")
	o2 := o1.With("foo", "en")
	o3 := o2.With("bar", "fr")
	o4 := o3.With("baz", "pt")

	g.Expect(o1.Langs).To(gomega.HaveLen(1))
	g.Expect(o1.Langs).To(gomega.ConsistOf("*"))
	g.Expect(o2.Langs).To(gomega.HaveLen(1))
	g.Expect(o2.Langs).To(gomega.ConsistOf("en"))
	g.Expect(o3.Langs).To(gomega.HaveLen(2))
	g.Expect(o3.Langs).To(gomega.ConsistOf("en", "fr"))
	g.Expect(o4.Langs).To(gomega.HaveLen(3))
	g.Expect(o4.Langs).To(gomega.ConsistOf("en", "fr", "pt"))

	g.Expect(o1.data).To(gomega.HaveLen(0))
	g.Expect(o2.data).To(gomega.HaveLen(1))
	g.Expect(o3.data).To(gomega.HaveLen(2))
	g.Expect(o4.data).To(gomega.HaveLen(3))
}

func TestBuildMatch(t *testing.T) {
	// TODO Test BuildMatch
}
