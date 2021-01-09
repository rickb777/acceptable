package acceptable_test

import (
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/rickb777/acceptable/processor"

	"github.com/onsi/gomega"
	"github.com/rickb777/acceptable"
)

func Test_should_use_default_processor_if_no_accept_header(t *testing.T) {
	g := gomega.NewWithT(t)

	a := acceptable.OfferOf("text/test")
	b := acceptable.OfferOf("text/plain")

	req, _ := http.NewRequest("GET", "/", nil)

	best := acceptable.BestRequestMatch(req, a, b)

	g.Expect(best).To(gomega.Equal(&acceptable.Match{
		Type:     "text",
		Subtype:  "test",
		Language: "*",
		Charset:  "utf-8",
	}))
}

func Test_should_give_JSON_response_for_ajax_requests(t *testing.T) {
	g := gomega.NewWithT(t)

	a := acceptable.OfferOf("")

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add(acceptable.XRequestedWith, acceptable.XMLHttpRequest)

	best := acceptable.BestRequestMatch(req, a)

	g.Expect(best).To(gomega.Equal(&acceptable.Match{
		Type:     "*",
		Subtype:  "*",
		Language: "*",
		Charset:  "utf-8",
	}))
}

func Test_should_give_406_for_unmatched_ajax_requests(t *testing.T) {
	g := gomega.NewWithT(t)

	a := acceptable.OfferOf("text/plain")

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add(acceptable.XRequestedWith, acceptable.XMLHttpRequest)

	best := acceptable.BestRequestMatch(req, a)

	g.Expect(best).To(gomega.BeNil())
}

func Test_should_return_406_if_no_matching_accept_header(t *testing.T) {
	g := gomega.NewWithT(t)

	cases := []string{"application/xml", "text/test"}

	for _, c := range cases {
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Add("Accept", "image/png")

		best := acceptable.BestRequestMatch(req, acceptable.OfferOf(c))

		g.Expect(best).To(gomega.BeNil())
	}
}

func Test_should_return_406_if_there_are_no_offers(t *testing.T) {
	g := gomega.NewWithT(t)

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("Accept", "image/png")

	best := acceptable.BestRequestMatch(req)

	g.Expect(best).To(gomega.BeNil())
}

func Test_should_return_406_if_there_are_no_offers_for_ajax(t *testing.T) {
	g := gomega.NewWithT(t)

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("Accept", "image/png")
	req.Header.Add(acceptable.XRequestedWith, acceptable.XMLHttpRequest)

	best := acceptable.BestRequestMatch(req)

	g.Expect(best).To(gomega.BeNil())
}

// RFC7231 suggests that 406 is sent when no media range matches are possible.
func Test_should_return_406_when_media_range_is_explicitly_excluded(t *testing.T) {
	g := gomega.NewWithT(t)

	a := acceptable.OfferOf("text/test", "en")

	req, _ := http.NewRequest("GET", "/", nil)
	// this header means "anything but text/test"
	req.Header.Add("Accept", "text/test;q=0, */*") // excluded
	req.Header.Add("Accept-Language", "en")        // accepted

	best := acceptable.BestRequestMatch(req, a)

	g.Expect(best).To(gomega.BeNil())
}

// RFC7231 recommends that, when no language matches are possible, a response should be sent anyway.
func Test_should_return_200_even_when_language_is_explicitly_excluded(t *testing.T) {
	g := gomega.NewWithT(t)

	a := acceptable.OfferOf("text/test", "en")

	req, _ := http.NewRequest("GET", "/", nil)
	// this header means "anything but text/test"
	req.Header.Add("Accept", "text/test, */*")
	req.Header.Add("Accept-Language", "en;q=0, *") // anything but "en"

	best := acceptable.BestRequestMatch(req, a)

	g.Expect(best).To(gomega.Equal(&acceptable.Match{
		Type:     "text",
		Subtype:  "test",
		Language: "en",
		Charset:  "utf-8",
	}))
}

func Test_should_negotiate_using_media_and_language(t *testing.T) {
	g := gomega.NewWithT(t)

	// should be skipped because of media mismatch
	a := acceptable.OfferOf("text/html", "en")
	// should be skipped because of language mismatch
	b := acceptable.OfferOf("text/test", "de")
	// should match
	c := acceptable.OfferOf("text/test", "en")

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("Accept", "text/test, text/*")
	req.Header.Add("Accept-Language", "en-GB, fr-FR")

	best := acceptable.BestRequestMatch(req, a, b, c)

	g.Expect(best).To(gomega.Equal(&acceptable.Match{
		Type:     "text",
		Subtype:  "test",
		Language: "en",
		Charset:  "utf-8",
	}))
}

func Test_should_match_subtype_wildcard1(t *testing.T) {
	g := gomega.NewWithT(t)

	a := acceptable.OfferOf("text/test")

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("Accept", "text/*") // <-- wildcard

	best := acceptable.BestRequestMatch(req, a)

	g.Expect(best).To(gomega.Equal(&acceptable.Match{
		Type:     "text",
		Subtype:  "test",
		Language: "*",
		Charset:  "utf-8",
	}))
}

func Test_should_match_subtype_wildcard2(t *testing.T) {
	g := gomega.NewWithT(t)

	a := acceptable.OfferOf("text/*") // <-- wildcard

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("Accept", "text/test")

	best := acceptable.BestRequestMatch(req, a)

	g.Expect(best).To(gomega.Equal(&acceptable.Match{
		Type:     "text",
		Subtype:  "test",
		Language: "*",
		Charset:  "utf-8",
	}))
}

func Test_should_match_language_when_offer_language_is_not_specified(t *testing.T) {
	g := gomega.NewWithT(t)

	a := acceptable.OfferOf("text/html")

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("Accept", "application/json, text/html")
	req.Header.Add("Accept-Language", "en, fr")
	req.Header.Add("Accept-Charset", "utf-8, iso-8859-1")

	best := acceptable.BestRequestMatch(req, a)

	g.Expect(best).To(gomega.Equal(&acceptable.Match{
		Type:     "text",
		Subtype:  "html",
		Language: "en",
		Charset:  "utf-8",
	}))
}

func Test_should_match_language_wildcard_and_return_selected_language(t *testing.T) {
	g := gomega.NewWithT(t)

	a := acceptable.OfferOf("", "en")
	b := acceptable.OfferOf("", "de")

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("Accept-Language", "*")

	best := acceptable.BestRequestMatch(req, a, b)

	g.Expect(best).To(gomega.Equal(&acceptable.Match{
		Type:     "*",
		Subtype:  "*",
		Language: "en",
		Charset:  "utf-8",
	}))
}

func Test_should_negotiate_a_default_processor(t *testing.T) {
	g := gomega.NewWithT(t)

	wildcard := acceptable.OfferOf("text/test")
	a := acceptable.OfferOf("text/test")

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("Accept", "*/*")

	best := acceptable.BestRequestMatch(req, wildcard)

	g.Expect(best).To(gomega.Equal(&acceptable.Match{
		Type:     "text",
		Subtype:  "test",
		Language: "*",
		Charset:  "utf-8",
	}))

	best = acceptable.BestRequestMatch(req, a)

	g.Expect(best).To(gomega.Equal(&acceptable.Match{
		Type:     "text",
		Subtype:  "test",
		Language: "*",
		Charset:  "utf-8",
	}))
}

func Test_should_negotiate_one_of_the_processors(t *testing.T) {
	g := gomega.NewWithT(t)

	a := acceptable.OfferOf("text/a")
	b := acceptable.OfferOf("text/b")

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("Accept", "text/a, text/b")

	best := acceptable.BestRequestMatch(req, a)

	g.Expect(best).To(gomega.Equal(&acceptable.Match{
		Type:     "text",
		Subtype:  "a",
		Language: "*",
		Charset:  "utf-8",
	}))

	best = acceptable.BestRequestMatch(req, b)

	g.Expect(best).To(gomega.Equal(&acceptable.Match{
		Type:     "text",
		Subtype:  "b",
		Language: "*",
		Charset:  "utf-8",
	}))
}

func TestMain(m *testing.M) {
	flag.Parse()
	if testing.Verbose() {
		acceptable.Debug = func(m string, a ...interface{}) { fmt.Printf(m, a...) }
	}
	os.Exit(m.Run())
}

func ExampleBestRequestMatch() {
	// In this example, the same content is available in three languages. We're using functions
	// for the sake of illustration, but simple values (often structs) will also work and will
	// sometimes be more appropriate.

	en := func() (interface{}, error) {
		return "Hello!", nil // get English content - eg from database
	}

	fr := func() (interface{}, error) {
		return "Bonjour!", nil // get French content - eg from database
	}

	es := func() (interface{}, error) {
		return "Hola!", nil // get Spanish content - eg from database
	}

	// We're implementing an HTTP handler, so we are given a request and a response.

	req := &http.Request{}                               // some incoming request
	var res http.ResponseWriter = httptest.NewRecorder() // replace with the server's response writer

	// Now do the content negotiation. This example has four supported content types, all of them
	// able to serve any of the three example languages.
	//
	// The first offer is for JSON - this is often the most widely used because it also supports
	// Ajax requests.

	best := acceptable.BestRequestMatch(req,
		acceptable.OfferOf("application/json", "en").Using(processor.JSON("  ")).
			With("en", en).With("fr", fr).With("es", es),

		acceptable.OfferOf("application/xml").Using(processor.XML("  ")).
			With("en", en).With("fr", fr).With("es", es),

		acceptable.OfferOf("text/csv").Using(processor.CSV()).
			With("en", en).With("fr", fr).With("es", es),

		acceptable.OfferOf("text/plain").Using(processor.TXT()).
			With("en", en).With("fr", fr).With("es", es))

	if best == nil {
		// The user agent asked for some content type that isn't available.

		res.WriteHeader(http.StatusNotAcceptable)
		res.Write([]byte(http.StatusText(http.StatusNotAcceptable)))
		return
	}

	// Happy case - we found a good match for the requested content so we serve it as the response.

	best.Render(res, *best, "")
}
