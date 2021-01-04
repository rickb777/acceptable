package acceptable

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestParseAcceptXyzHeader_with_inverse_string(t *testing.T) {
	g := NewGomegaWithT(t)
	cases := []struct {
		actual   string
		expected PrecedenceValues
	}{
		// nil handling
		{actual: "", expected: nil},

		// single
		{actual: "utf-8", expected: []PrecedenceValue{{Value: "utf-8", Quality: DefaultQuality}}},
		{actual: "gzip", expected: []PrecedenceValue{{Value: "gzip", Quality: DefaultQuality}}},
		{actual: "en-gb", expected: []PrecedenceValue{{Value: "en-gb", Quality: DefaultQuality}}},

		// with quality - in order
		{
			actual:   "iso-8859-5, unicode-1-1;q=0.8",
			expected: []PrecedenceValue{{Value: "iso-8859-5", Quality: DefaultQuality}, {Value: "unicode-1-1", Quality: 0.8}},
		},
		{
			actual:   "gzip, identity;q=0.5",
			expected: []PrecedenceValue{{Value: "gzip", Quality: DefaultQuality}, {Value: "identity", Quality: 0.5}},
		},
		{
			actual:   "da, en-gb;q=0.8, en;q=0.7",
			expected: []PrecedenceValue{{Value: "da", Quality: DefaultQuality}, {Value: "en-gb", Quality: 0.8}, {Value: "en", Quality: 0.7}},
		},

		// with quality - sorted
		{
			actual:   "iso-8859-5, unicode-1-1;q=0.8",
			expected: []PrecedenceValue{{Value: "iso-8859-5", Quality: DefaultQuality}, {Value: "unicode-1-1", Quality: 0.8}},
		},
		{
			actual:   "gzip, identity;q=0.5",
			expected: []PrecedenceValue{{Value: "gzip", Quality: DefaultQuality}, {Value: "identity", Quality: 0.5}},
		},
		{
			actual:   "da, en-gb;q=0.8, en;q=0.7",
			expected: []PrecedenceValue{{Value: "da", Quality: DefaultQuality}, {Value: "en-gb", Quality: 0.8}, {Value: "en", Quality: 0.7}},
		},
		{
			actual:   "en-gb, en-us, en;q=0.7",
			expected: []PrecedenceValue{{Value: "en-gb", Quality: DefaultQuality}, {Value: "en-us", Quality: DefaultQuality}, {Value: "en", Quality: 0.7}},
		},
	}

	for _, c := range cases {
		actual := Parse(c.actual)
		g.Expect(actual).To(Equal(c.expected))
		g.Expect(actual.String()).To(Equal(c.actual))
	}
}

func TestParseAcceptXyzHeader_special_cases(t *testing.T) {
	g := NewGomegaWithT(t)
	cases := []struct {
		actual   string
		expected PrecedenceValues
	}{
		// ignore invalid quality
		{actual: "UTF-8;q=z", expected: []PrecedenceValue{{Value: "utf-8", Quality: DefaultQuality}}},
		{actual: "gzip;q=z", expected: []PrecedenceValue{{Value: "gzip", Quality: DefaultQuality}}},
		{actual: "en-gb;q=z", expected: []PrecedenceValue{{Value: "en-gb", Quality: DefaultQuality}}},

		// with quality - in order
		{
			actual:   "iso-8859-5, unicode-1-1; q=0.8\n",
			expected: []PrecedenceValue{{Value: "iso-8859-5", Quality: DefaultQuality}, {Value: "unicode-1-1", Quality: 0.8}},
		},
		{
			actual:   " gzip; q=1.0, identity; q=0.5",
			expected: []PrecedenceValue{{Value: "gzip", Quality: DefaultQuality}, {Value: "identity", Quality: 0.5}},
		},
		{
			actual:   " DA, en-gb;q=0.8, en; q=0.7",
			expected: []PrecedenceValue{{Value: "da", Quality: DefaultQuality}, {Value: "en-gb", Quality: 0.8}, {Value: "en", Quality: 0.7}},
		},

		// with quality - sorted
		{
			actual:   "unicode-1-1;q=0.8, ISO-8859-5\n",
			expected: []PrecedenceValue{{Value: "iso-8859-5", Quality: DefaultQuality}, {Value: "unicode-1-1", Quality: 0.8}},
		},
		{
			actual:   "identity; q=0.5, gzip; q=1.0",
			expected: []PrecedenceValue{{Value: "gzip", Quality: DefaultQuality}, {Value: "identity", Quality: 0.5}},
		},
		{
			actual:   "en;q=0.7, en-gb;q=0.8, da",
			expected: []PrecedenceValue{{Value: "da", Quality: DefaultQuality}, {Value: "en-gb", Quality: 0.8}, {Value: "en", Quality: 0.7}},
		},
		{
			actual:   "en-gb, en-us, en;q=0.7",
			expected: []PrecedenceValue{{Value: "en-gb", Quality: DefaultQuality}, {Value: "en-us", Quality: DefaultQuality}, {Value: "en", Quality: 0.7}},
		},
		{
			actual:   "en;q=-1",
			expected: []PrecedenceValue{{Value: "en", Quality: 0}},
		},
		{
			actual:   "en;q=13",
			expected: []PrecedenceValue{{Value: "en", Quality: 1}},
		},
	}

	for _, c := range cases {
		actual := Parse(c.actual)
		g.Expect(actual).To(Equal(c.expected))
	}
}
