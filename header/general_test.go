package header_test

import (
	"fmt"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	. "github.com/rickb777/acceptable/header"
)

func TestParseDate(t *testing.T) {
	//est, _ := time.LoadLocation("America/New_York")
	g := NewGomegaWithT(t)
	cases := []struct {
		input     string
		canonical string
		expected  time.Time
	}{
		{
			input:     "Sun, 06 Nov 1994 08:49:37 GMT", // canonical
			canonical: "Sun, 06 Nov 1994 08:49:37 GMT",
			expected:  time.Date(1994, 11, 6, 8, 49, 37, 0, time.UTC),
		},
		{
			input:     "Wed, 01 Jan 2020 01:01:01 UTC", // non-standard UTC
			canonical: "Wed, 01 Jan 2020 01:01:01 GMT",
			expected:  time.Date(2020, 1, 1, 1, 1, 1, 0, time.UTC),
		},
		{
			input:     "Sunday, 06-Nov-94 08:49:37 GMT", // obsolete RFC-850
			canonical: "Sun, 06 Nov 1994 08:49:37 GMT",
			expected:  time.Date(1994, 11, 6, 8, 49, 37, 0, time.UTC),
		},
		{
			input:     "Sun Nov  6 08:49:37 1994", // obsolete ANSI-C
			canonical: "Sun, 06 Nov 1994 08:49:37 GMT",
			expected:  time.Date(1994, 11, 6, 8, 49, 37, 0, time.UTC),
		},
	}

	for i, c := range cases {
		actual, err := ParseHTTPDateTime(c.input)
		g.Expect(err).NotTo(HaveOccurred(), "%d", i)
		g.Expect(actual.Equal(c.expected)).To(BeTrue(), "%d %s != %s", i, actual, c.expected)
		g.Expect(FormatHTTPDateTime(actual)).To(Equal(c.canonical), "%d", i)
	}

	_, err := ParseHTTPDateTime("")
	g.Expect(err).To(HaveOccurred())

	_, err = ParseHTTPDateTime("not a date")
	g.Expect(err).To(HaveOccurred())
}

func TestFormatDate(t *testing.T) {
	est, _ := time.LoadLocation("America/New_York")
	g := NewGomegaWithT(t)
	cases := []struct {
		dateTime  time.Time
		canonical string
	}{
		{
			dateTime:  time.Date(2020, 1, 1, 1, 1, 1, 0, time.UTC),
			canonical: "Wed, 01 Jan 2020 01:01:01 GMT",
		},
		{
			dateTime:  time.Date(1994, 11, 6, 8, 49, 37, 0, time.UTC),
			canonical: "Sun, 06 Nov 1994 08:49:37 GMT",
		},
		{
			dateTime:  time.Date(1994, 11, 6, 8, 49, 37, 0, est),
			canonical: "Sun, 06 Nov 1994 08:49:37 GMT",
		},
	}

	for i, c := range cases {
		actual := FormatHTTPDateTime(c.dateTime)
		g.Expect(actual).To(Equal(c.canonical), "%d", i)
	}

	_, err := ParseHTTPDateTime("")
	g.Expect(err).To(HaveOccurred())
}

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
		actual := ParsePrecedenceValues(c.actual)
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
		actual := ParsePrecedenceValues(c.actual)
		g.Expect(actual).To(Equal(c.expected))
	}
}

func ExampleParsePrecedenceValues() {
	pvs := ParsePrecedenceValues("da, en-gb;q=0.8, en;q=0.7")

	for i, pv := range pvs {
		fmt.Printf("pv%d = %s\n", i, pv)
	}
	// Output:
	// pv0 = da
	// pv1 = en-gb;q=0.8
	// pv2 = en;q=0.7
}
