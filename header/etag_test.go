package header

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestETagsOf(t *testing.T) {
	g := NewGomegaWithT(t)
	cases := []struct {
		input    string
		expected ETags
	}{
		{
			input:    "",
			expected: nil,
		},
		{
			input:    `*`,
			expected: ETags{ETag{Hash: "*"}},
		},
		{
			input:    `"xyzzy", "r2d2xxxx", "c3piozzzz"`,
			expected: ETags{ETag{Hash: "xyzzy"}, ETag{Hash: "r2d2xxxx"}, ETag{Hash: "c3piozzzz"}},
		},
		{
			input:    `W/"xyzzy", W/"r2d2xxxx", W/"c3piozzzz"`,
			expected: ETags{ETag{Hash: "xyzzy", Weak: true}, ETag{Hash: "r2d2xxxx", Weak: true}, ETag{Hash: "c3piozzzz", Weak: true}},
		},
	}
	for _, c := range cases {
		actual := ETagsOf(c.input)
		g.Expect(actual).To(ConsistOf(c.expected))
		g.Expect(actual.String()).To(Equal(c.input))
	}
}

func TestWeaklyMatches(t *testing.T) {
	g := NewGomegaWithT(t)

	etags1 := ETagsOf(`"xyzzy", "r2d2xxxx", "c3piozzzz"`)
	g.Expect(etags1.WeaklyMatches("c3piozzzz")).To(BeTrue())
	g.Expect(etags1.WeaklyMatches("zzzz")).To(BeFalse())

	etags2 := ETagsOf(`W/"xyzzy", W/"r2d2xxxx", W/"c3piozzzz"`)
	g.Expect(etags2.WeaklyMatches("c3piozzzz")).To(BeTrue())
	g.Expect(etags2.WeaklyMatches("zzzz")).To(BeFalse())
}

func TestStronglyMatches(t *testing.T) {
	g := NewGomegaWithT(t)

	etags1 := ETagsOf(`"xyzzy", "r2d2xxxx", "c3piozzzz"`)
	g.Expect(etags1.StronglyMatches("c3piozzzz")).To(BeTrue())
	g.Expect(etags1.StronglyMatches("zzzz")).To(BeFalse())

	etags2 := ETagsOf(`W/"xyzzy", W/"r2d2xxxx", W/"c3piozzzz"`)
	g.Expect(etags2.StronglyMatches("c3piozzzz")).To(BeFalse())
	g.Expect(etags2.StronglyMatches("zzzz")).To(BeFalse())
}
