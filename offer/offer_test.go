package offer

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
		"1.Accept: */*": {o: Of(nil, ""), n: 0},

		"2.Accept: text/*": {o: Of(nil, "text/*"), n: 0},

		"3.Accept: a/b": {o: Of(nil, "a/b").With(nil, "*"), n: 0},

		"4.Accept: a/b. Accept-Language: *": {o: Of(nil, "a/b").With("foo", "*"), n: 1},

		"5.Accept: a/b. Accept-Language: en": {o: Of(nil, "a/b").With("foo", "en"), n: 1},

		"6.Accept: a/b. Accept-Language: en": {o: Of(nil, "a/b").With(nil, "en"), n: 1},

		"7.Accept: a/b. Accept-Language: en,fr": {o: Of(nil, "a/b").With("foo", "en").With("bar", "fr"), n: 2},

		"8.Accept: a/b. Accept-Language: en,fr": {o: Of(nil, "a/b").With("foo", "en", "fr"), n: 2},

		"9.Accept: a/b. Accept-Language: en,fr": {o: Of(nil, "a/b").With(data.Of("foo"), "en", "fr"), n: 2},
	}

	for s, c := range cases {
		g.Expect(c.o.String()).To(gomega.Equal(s[2:]), s)
		g.Expect(len(c.o.data)).To(gomega.Equal(c.n), s)

		for l, d := range c.o.data {
			g.Expect(fmt.Sprintf("%T", d)).To(
				gomega.Or(
					gomega.Equal("*data.Value"),
					gomega.Equal("offer.empty"),
				), s+"|"+l)
		}
	}
}
