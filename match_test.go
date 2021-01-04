package acceptable

import (
	"net/http"
	"testing"

	"github.com/onsi/gomega"
)

func Test_should_use_default_processor_if_no_accept_header(t *testing.T) {
	g := gomega.NewWithT(t)

	a := OfferOf("text/test")
	b := OfferOf("text/plain")

	req, _ := http.NewRequest("GET", "/", nil)

	best := BestRequestMatch(req, a, b)

	g.Expect(best).To(gomega.Equal(&a))
}

func Test_should_give_JSON_response_for_ajax_requests(t *testing.T) {
	g := gomega.NewWithT(t)

	a := OfferOf("")

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add(XRequestedWith, XMLHttpRequest)

	best := BestRequestMatch(req, a)

	g.Expect(best).To(gomega.Equal(&a))
}

func Test_should_give_406_for_unmatched_ajax_requests(t *testing.T) {
	g := gomega.NewWithT(t)

	a := OfferOf("text/plain")

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add(XRequestedWith, XMLHttpRequest)

	best := BestRequestMatch(req, a)

	g.Expect(best).To(gomega.BeNil())
}

func Test_should_return_406_if_no_matching_accept_header(t *testing.T) {
	g := gomega.NewWithT(t)

	cases := []string{"application/xml", "text/test"}

	for _, c := range cases {
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Add("Accept", "image/png")

		best := BestRequestMatch(req, OfferOf(c))

		g.Expect(best).To(gomega.BeNil())
	}
}

// RFC7231 suggests that 406 is sent when no media range matches are possible.
func Test_should_return_406_when_media_range_is_explicitly_excluded(t *testing.T) {
	g := gomega.NewWithT(t)

	a := OfferOf("text/test", "en")

	req, _ := http.NewRequest("GET", "/", nil)
	// this header means "anything but text/test"
	req.Header.Add("Accept", "text/test;q=0, */*") // excluded
	req.Header.Add("Accept-Language", "en")        // accepted

	best := BestRequestMatch(req, a)

	g.Expect(best).To(gomega.BeNil())
}

// RFC7231 recommends that, when no language matches are possible, a response should be sent anyway.
func Test_should_return_200_even_when_language_is_explicitly_excluded(t *testing.T) {
	g := gomega.NewWithT(t)

	a := OfferOf("text/test", "en")

	req, _ := http.NewRequest("GET", "/", nil)
	// this header means "anything but text/test"
	req.Header.Add("Accept", "text/test, */*")
	req.Header.Add("Accept-Language", "en;q=0 *") // anything but "en"

	best := BestRequestMatch(req, a)

	g.Expect(best).To(gomega.Equal(&a))
}

func Test_should_negotiate_using_media_and_language(t *testing.T) {
	g := gomega.NewWithT(t)

	// should be skipped because of media mismatch
	a := OfferOf("text/html", "en")
	// should be skipped because of language mismatch
	b := OfferOf("text/test", "de")
	// should match
	c := OfferOf("text/test", "en")

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("Accept", "text/test, text/*")
	req.Header.Add("Accept-Language", "en-GB, fr-FR")

	best := BestRequestMatch(req, a, b, c)

	g.Expect(best).To(gomega.Equal(&c))
}

func Test_should_match_subtype_wildcard(t *testing.T) {
	g := gomega.NewWithT(t)

	a := OfferOf("text/test")

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("Accept", "text/*")

	best := BestRequestMatch(req, a)

	g.Expect(best).To(gomega.Equal(&a))
}

func Test_should_match_language_when_offer_language_is_not_specified(t *testing.T) {
	g := gomega.NewWithT(t)

	a := OfferOf("text/html")

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("Accept", "text/html")
	req.Header.Add("Accept-Language", "en, fr")

	best := BestRequestMatch(req, a)

	g.Expect(best).To(gomega.Equal(&a))
}

func Test_should_match_language_wildcard_and_send_content_language_header(t *testing.T) {
	g := gomega.NewWithT(t)

	var a = OfferOf("", "en")

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("Accept-Language", "*")

	best := BestRequestMatch(req, OfferOf("", "en"))

	g.Expect(best).To(gomega.Equal(&a))
}

func Test_should_negotiate_a_default_processor(t *testing.T) {
	g := gomega.NewWithT(t)

	wildcard := OfferOf("text/test")
	a := OfferOf("text/test")

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("Accept", "*/*")

	best := BestRequestMatch(req, wildcard)

	g.Expect(best).To(gomega.Equal(&wildcard))

	best = BestRequestMatch(req, a)

	g.Expect(best).To(gomega.Equal(&a))
}

func Test_should_negotiate_one_of_the_processors(t *testing.T) {
	g := gomega.NewWithT(t)

	a := OfferOf("text/a")
	b := OfferOf("text/b")

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("Accept", "text/a, text/b")

	best := BestRequestMatch(req, a)

	g.Expect(best).To(gomega.Equal(&a))

	best = BestRequestMatch(req, b)

	g.Expect(best).To(gomega.Equal(&b))
}
