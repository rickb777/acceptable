package acceptable

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/onsi/gomega"
)

func Test_should_use_default_processor_if_no_accept_header(t *testing.T) {
	g := gomega.NewWithT(t)

	a := OfferOf("text/test")
	b := OfferOf("text/plain")

	req, _ := http.NewRequest("GET", "/", nil)

	best := BestRequestMatch(req, a, b)

	g.Expect(best).To(gomega.Equal(&Match{
		Type:     "text",
		Subtype:  "test",
		Language: "*",
		Charset:  "utf-8",
	}))
}

func Test_should_give_JSON_response_for_ajax_requests(t *testing.T) {
	g := gomega.NewWithT(t)

	a := OfferOf("")

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add(XRequestedWith, XMLHttpRequest)

	best := BestRequestMatch(req, a)

	g.Expect(best).To(gomega.Equal(&Match{
		Type:     "*",
		Subtype:  "*",
		Language: "*",
		Charset:  "utf-8",
	}))
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

func Test_should_return_406_if_there_are_no_offers(t *testing.T) {
	g := gomega.NewWithT(t)

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("Accept", "image/png")

	best := BestRequestMatch(req)

	g.Expect(best).To(gomega.BeNil())
}

func Test_should_return_406_if_there_are_no_offers_for_ajax(t *testing.T) {
	g := gomega.NewWithT(t)

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("Accept", "image/png")
	req.Header.Add(XRequestedWith, XMLHttpRequest)

	best := BestRequestMatch(req)

	g.Expect(best).To(gomega.BeNil())
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
	req.Header.Add("Accept-Language", "en;q=0, *") // anything but "en"

	best := BestRequestMatch(req, a)

	g.Expect(best).To(gomega.Equal(&Match{
		Type:     "text",
		Subtype:  "test",
		Language: "en",
		Charset:  "utf-8",
	}))
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

	g.Expect(best).To(gomega.Equal(&Match{
		Type:     "text",
		Subtype:  "test",
		Language: "en",
		Charset:  "utf-8",
	}))
}

func Test_should_match_subtype_wildcard1(t *testing.T) {
	g := gomega.NewWithT(t)

	a := OfferOf("text/test")

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("Accept", "text/*") // <-- wildcard

	best := BestRequestMatch(req, a)

	g.Expect(best).To(gomega.Equal(&Match{
		Type:     "text",
		Subtype:  "test",
		Language: "*",
		Charset:  "utf-8",
	}))
}

func Test_should_match_subtype_wildcard2(t *testing.T) {
	g := gomega.NewWithT(t)

	a := OfferOf("text/*") // <-- wildcard

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("Accept", "text/test")

	best := BestRequestMatch(req, a)

	g.Expect(best).To(gomega.Equal(&Match{
		Type:     "text",
		Subtype:  "test",
		Language: "*",
		Charset:  "utf-8",
	}))
}

func Test_should_match_language_when_offer_language_is_not_specified(t *testing.T) {
	g := gomega.NewWithT(t)

	a := OfferOf("text/html")

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("Accept", "application/json, text/html")
	req.Header.Add("Accept-Language", "en, fr")
	req.Header.Add("Accept-Charset", "utf-8, iso-8859-1")

	best := BestRequestMatch(req, a)

	g.Expect(best).To(gomega.Equal(&Match{
		Type:     "text",
		Subtype:  "html",
		Language: "en",
		Charset:  "utf-8",
	}))
}

func Test_should_match_language_wildcard_and_return_selected_language(t *testing.T) {
	g := gomega.NewWithT(t)

	a := OfferOf("", "en")
	b := OfferOf("", "de")

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("Accept-Language", "*")

	best := BestRequestMatch(req, a, b)

	g.Expect(best).To(gomega.Equal(&Match{
		Type:     "*",
		Subtype:  "*",
		Language: "en",
		Charset:  "utf-8",
	}))
}

func Test_should_negotiate_a_default_processor(t *testing.T) {
	g := gomega.NewWithT(t)

	wildcard := OfferOf("text/test")
	a := OfferOf("text/test")

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("Accept", "*/*")

	best := BestRequestMatch(req, wildcard)

	g.Expect(best).To(gomega.Equal(&Match{
		Type:     "text",
		Subtype:  "test",
		Language: "*",
		Charset:  "utf-8",
	}))

	best = BestRequestMatch(req, a)

	g.Expect(best).To(gomega.Equal(&Match{
		Type:     "text",
		Subtype:  "test",
		Language: "*",
		Charset:  "utf-8",
	}))
}

func Test_should_negotiate_one_of_the_processors(t *testing.T) {
	g := gomega.NewWithT(t)

	a := OfferOf("text/a")
	b := OfferOf("text/b")

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("Accept", "text/a, text/b")

	best := BestRequestMatch(req, a)

	g.Expect(best).To(gomega.Equal(&Match{
		Type:     "text",
		Subtype:  "a",
		Language: "*",
		Charset:  "utf-8",
	}))

	best = BestRequestMatch(req, b)

	g.Expect(best).To(gomega.Equal(&Match{
		Type:     "text",
		Subtype:  "b",
		Language: "*",
		Charset:  "utf-8",
	}))
}

func TestMain(m *testing.M) {
	flag.Parse()
	if testing.Verbose() {
		Debug = func(m string, a ...interface{}) { fmt.Printf(m, a...) }
	}
	os.Exit(m.Run())
}
