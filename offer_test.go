package acceptable

import (
	"fmt"
	"testing"

	"github.com/onsi/gomega"
	"github.com/rickb777/acceptable/data"
)

func Test_offer_construction(t *testing.T) {
	g := gomega.NewWithT(t)

	cases := map[string]struct {
		o Offer
		n int
	}{
		"1.Accept: */*": {o: OfferOf(nil, ""), n: 0},

		"2.Accept: a/b": {o: OfferOf(nil, "a/b").With(nil, "*"), n: 0},

		"3.Accept: a/b. Accept-Language: *": {o: OfferOf(nil, "a/b").With("foo", "*"), n: 1},

		"4.Accept: a/b. Accept-Language: en": {o: OfferOf(nil, "a/b").With("foo", "en"), n: 1},

		"5.Accept: a/b. Accept-Language: en,fr": {o: OfferOf(nil, "a/b").With("foo", "en").With("bar", "fr"), n: 2},

		"6.Accept: a/b. Accept-Language: en,fr": {o: OfferOf(nil, "a/b").With("foo", "en", "fr"), n: 2},

		"7.Accept: a/b. Accept-Language: en,fr": {o: OfferOf(nil, "a/b").With(data.Of("foo"), "en", "fr"), n: 2},
	}

	for s, c := range cases {
		g.Expect(c.o.String()).To(gomega.Equal(s[2:]), s)
		g.Expect(len(c.o.data)).To(gomega.Equal(c.n), s)
		for l, d := range c.o.data {
			g.Expect(fmt.Sprintf("%T", d)).To(gomega.Equal("*data.Value"), s+l)
		}
	}
}
