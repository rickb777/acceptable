package header

import (
	"testing"

	"github.com/rickb777/expect"
)

func TestETagsOf(t *testing.T) {
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
	for i, c := range cases {
		actual := ETagsOf(c.input)
		expect.Slice(actual).I(i).ToBe(t, c.expected...)
		expect.String(actual.String()).I(i).ToBe(t, c.input)
	}
}

func TestWeaklyMatches(t *testing.T) {
	etags1 := ETagsOf(`"xyzzy", "r2d2xxxx", "c3piozzzz"`)
	expect.Bool(etags1.WeaklyMatches("c3piozzzz")).ToBeTrue(t)
	expect.Bool(etags1.WeaklyMatches("zzzz")).ToBeFalse(t)

	etags2 := ETagsOf(`W/"xyzzy", W/"r2d2xxxx", W/"c3piozzzz"`)
	expect.Bool(etags2.WeaklyMatches("c3piozzzz")).ToBeTrue(t)
	expect.Bool(etags2.WeaklyMatches("zzzz")).ToBeFalse(t)
}

func TestStronglyMatches(t *testing.T) {
	etags1 := ETagsOf(`"xyzzy", "r2d2xxxx", "c3piozzzz"`)
	expect.Bool(etags1.StronglyMatches("c3piozzzz")).ToBeTrue(t)
	expect.Bool(etags1.StronglyMatches("zzzz")).ToBeFalse(t)

	etags2 := ETagsOf(`W/"xyzzy", W/"r2d2xxxx", W/"c3piozzzz"`)
	expect.Bool(etags2.StronglyMatches("c3piozzzz")).ToBeFalse(t)
	expect.Bool(etags2.StronglyMatches("zzzz")).ToBeFalse(t)
}
