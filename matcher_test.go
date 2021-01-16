package acceptable_test

import (
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/onsi/gomega"
	"github.com/rickb777/acceptable"
	"github.com/rickb777/acceptable/data"
	"github.com/rickb777/acceptable/header"
	"github.com/rickb777/acceptable/processor"
	"github.com/rickb777/acceptable/templates"
)

func Test_should_use_default_processor_if_no_accept_header(t *testing.T) {
	g := gomega.NewWithT(t)

	// Given ...
	a := acceptable.OfferOf("text/test").Using(processor.TXT())
	b := acceptable.OfferOf("text/plain").Using(processor.TXT())

	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	// When ...
	err := acceptable.RenderBestMatch(w, req, "", a, b)

	// Then ...
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(w.HeaderMap).To(gomega.HaveLen(1))
	g.Expect(w.Header().Get("Content-Type")).To(gomega.Equal("text/test;charset=utf-8"))
}

func Test_should_give_JSON_response_for_ajax_requests(t *testing.T) {
	g := gomega.NewWithT(t)

	// Given ...
	a := acceptable.OfferOf("application/json").Using(processor.JSON())

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add(header.XRequestedWith, header.XMLHttpRequest)
	w := httptest.NewRecorder()

	// When ...
	best := acceptable.BestRequestMatch(req, a)
	err := best.Render(w, req, *best, "")

	// Then ...
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(w.HeaderMap).To(gomega.HaveLen(1))
	g.Expect(w.Header().Get("Content-Type")).To(gomega.Equal("application/json;charset=utf-8"))
}

func Test_should_give_406_for_unmatched_ajax_requests(t *testing.T) {
	g := gomega.NewWithT(t)

	// Given ...
	a := acceptable.OfferOf("text/plain").Using(processor.JSON())

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add(header.XRequestedWith, header.XMLHttpRequest)
	w := httptest.NewRecorder()

	// When ...
	err := acceptable.RenderBestMatch(w, req, "", a)

	// Then ...
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(w.Code).To(gomega.Equal(406))
}

func Test_should_return_406_if_no_matching_accept_header(t *testing.T) {
	g := gomega.NewWithT(t)

	cases := []string{"application/xml", "text/test"}

	for _, c := range cases {
		// Given ...
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Add("Accept", "image/png")
		w := httptest.NewRecorder()
		a := acceptable.OfferOf(c).Using(processor.JSON())

		// When ...
		err := acceptable.RenderBestMatch(w, req, "", a)

		// Then ...
		g.Expect(err).NotTo(gomega.HaveOccurred())
		g.Expect(w.Code).To(gomega.Equal(406))
	}
}

func Test_should_return_406_if_there_are_no_offers(t *testing.T) {
	g := gomega.NewWithT(t)

	// Given ...
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("Accept", "image/png")
	w := httptest.NewRecorder()

	// When ...
	err := acceptable.RenderBestMatch(w, req, "")

	// Then ...
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(w.Code).To(gomega.Equal(406))
}

func Test_should_return_406_if_there_are_no_offers_for_ajax(t *testing.T) {
	g := gomega.NewWithT(t)

	// Given ...
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("Accept", "image/png")
	req.Header.Add(header.XRequestedWith, header.XMLHttpRequest)
	w := httptest.NewRecorder()

	// When ...
	err := acceptable.RenderBestMatch(w, req, "")

	// Then ...
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(w.Code).To(gomega.Equal(406))
}

// RFC7231 suggests that 406 is sent when no media range matches are possible.
func Test_should_return_406_when_media_range_is_explicitly_excluded(t *testing.T) {
	g := gomega.NewWithT(t)

	// Given ...
	a := acceptable.OfferOf("text/test", "en").Using(processor.TXT())

	req, _ := http.NewRequest("GET", "/", nil)
	// this header means "anything but text/test"
	req.Header.Add("Accept", "text/test;q=0, */*") // excluded
	req.Header.Add("Accept-Language", "en")        // accepted
	w := httptest.NewRecorder()

	// When ...
	err := acceptable.RenderBestMatch(w, req, "", a)

	// Then ...
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(w.Code).To(gomega.Equal(406))
}

// RFC7231 recommends that, when no language matches are possible, a response should be sent anyway.
func Test_should_return_200_even_when_language_is_explicitly_excluded(t *testing.T) {
	g := gomega.NewWithT(t)

	// Given ...
	a := acceptable.OfferOf("text/test").With("en", nil).Using(processor.TXT())

	req, _ := http.NewRequest("GET", "/", nil)
	// this header means "anything but text/test"
	req.Header.Add("Accept", "text/test, */*")
	req.Header.Add("Accept-Language", "en;q=0, *") // anything but "en"
	w := httptest.NewRecorder()

	// When ...
	err := acceptable.RenderBestMatch(w, req, "", a)

	// Then ...
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(w.HeaderMap).To(gomega.HaveLen(3))
	g.Expect(w.Header().Get("Content-Type")).To(gomega.Equal("text/test;charset=utf-8"))
	g.Expect(w.Header().Get("Content-Language")).To(gomega.Equal("en"))
	g.Expect(w.Header().Get("Vary")).To(gomega.Equal("Accept, Accept-Language"))
}

func Test_should_negotiate_using_media_and_language(t *testing.T) {
	g := gomega.NewWithT(t)

	// Given ...
	// should be skipped because of media mismatch
	a := acceptable.OfferOf("text/html", "en").Using(processor.TXT())
	// should be skipped because of language mismatch
	b := acceptable.OfferOf("text/test", "de").Using(processor.TXT())
	// should match
	c := acceptable.OfferOf("text/test", "en").Using(processor.TXT())

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("Accept", "text/test, text/*")
	req.Header.Add("Accept-Language", "en-GB, fr-FR")
	w := httptest.NewRecorder()

	// When ...
	err := acceptable.RenderBestMatch(w, req, "", a, b, c)

	// Then ...
	g.Expect(err).NotTo(gomega.HaveOccurred())

	g.Expect(w.HeaderMap).To(gomega.HaveLen(3))
	g.Expect(w.Header().Get("Content-Type")).To(gomega.Equal("text/test;charset=utf-8"))
	g.Expect(w.Header().Get("Content-Language")).To(gomega.Equal("en"))
	g.Expect(w.Header().Get("Vary")).To(gomega.Equal("Accept, Accept-Language"))
}

func Test_should_match_subtype_wildcard1(t *testing.T) {
	g := gomega.NewWithT(t)

	// Given ...
	a := acceptable.OfferOf("text/test")

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("Accept", "text/*") // <-- wildcard

	// When ...
	best := acceptable.BestRequestMatch(req, a)

	// Then ...
	g.Expect(best).To(gomega.Equal(&acceptable.Match{
		Type:     "text",
		Subtype:  "test",
		Language: "*",
		Charset:  "utf-8",
		Vary:     []string{"Accept"},
	}))
}

func Test_should_match_subtype_wildcard2(t *testing.T) {
	g := gomega.NewWithT(t)

	// Given ...
	a := acceptable.OfferOf("text/*") // <-- wildcard

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("Accept", "text/test")

	// When ...
	best := acceptable.BestRequestMatch(req, a)

	// Then ...
	g.Expect(best).To(gomega.Equal(&acceptable.Match{
		Type:     "text",
		Subtype:  "test",
		Language: "*",
		Charset:  "utf-8",
		Vary:     []string{"Accept"},
	}))
}

func Test_should_render_iso8859_html_using_templates(t *testing.T) {
	g := gomega.NewWithT(t)

	// Given ...
	a := acceptable.OfferOf("text/html", "en").
		Using(templates.Templates("example/templates/en", ".html", nil))

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("Accept", "text/html")
	req.Header.Add("Accept-Language", "en, fr")
	req.Header.Add("Accept-Charset", "iso-8859-1")
	w := httptest.NewRecorder()

	// When ...
	err := acceptable.RenderBestMatch(w, req, "home.html", a)

	// Then ...
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(w.HeaderMap).To(gomega.HaveLen(3))
	g.Expect(w.Header().Get("Content-Type")).To(gomega.Equal("text/html;charset=windows-1252"))
	g.Expect(w.Header().Get("Content-Language")).To(gomega.Equal("en"))
	g.Expect(w.Header().Get("Vary")).To(gomega.Equal("Accept, Accept-Language, Accept-Charset"))
}

func Test_should_match_language_when_offer_language_is_not_specified(t *testing.T) {
	g := gomega.NewWithT(t)

	// Given ...
	a := acceptable.OfferOf("text/html")

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("Accept", "application/json, text/html")
	req.Header.Add("Accept-Language", "en, fr")
	req.Header.Add("Accept-Charset", "utf-8, iso-8859-1")

	// When ...
	best := acceptable.BestRequestMatch(req, a)

	// Then ...
	g.Expect(best).To(gomega.Equal(&acceptable.Match{
		Type:     "text",
		Subtype:  "html",
		Language: "en",
		Charset:  "utf-8",
		Vary:     []string{"Accept", "Accept-Language"},
	}))
}

func Test_should_match_language_wildcard_and_return_selected_language(t *testing.T) {
	g := gomega.NewWithT(t)

	// Given ...
	a := acceptable.OfferOf("", "en")
	b := acceptable.OfferOf("", "de")

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("Accept-Language", "*")

	// When ...
	best := acceptable.BestRequestMatch(req, a, b)

	// Then ...
	g.Expect(best).To(gomega.Equal(&acceptable.Match{
		Type:     "application",
		Subtype:  "octet-stream",
		Language: "en",
		Charset:  "utf-8",
		Vary:     []string{"Accept-Language"},
	}))
}

func Test_should_select_language_of_first_matched_offer_when_no_language_matches(t *testing.T) {
	g := gomega.NewWithT(t)

	// Given ...
	a := acceptable.OfferOf("text/csv", "es")
	b := acceptable.OfferOf("text/html", "en")
	c := acceptable.OfferOf("text/html", "de")

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("Accept", "text/html")
	req.Header.Add("Accept-Language", "fr")

	// When ...
	best := acceptable.BestRequestMatch(req, a, b, c)

	// Then ...
	g.Expect(best).To(gomega.Equal(&acceptable.Match{
		Type:     "text",
		Subtype:  "html",
		Language: "en",
		Charset:  "utf-8",
		Vary:     []string{"Accept", "Accept-Language"},
	}))
}

func Test_should_match_utf8_charset_when_acceptable(t *testing.T) {
	g := gomega.NewWithT(t)

	// Given ...
	a := acceptable.OfferOf("text/html").Using(processor.TXT())

	cases := []string{
		"utf-8, iso-8859-1",
		"utf8, iso-8859-1",
		"iso-8859-1, utf-8",
		"iso-8859-1, utf8",
	}

	for _, cs := range cases {
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Add("Accept", "text/html")
		req.Header.Add("Accept-Language", "en")
		req.Header.Add("Accept-Charset", cs)
		w := httptest.NewRecorder()

		// When ...
		err := acceptable.RenderBestMatch(w, req, "", a)

		// Then ...
		g.Expect(err).NotTo(gomega.HaveOccurred())
		g.Expect(w.HeaderMap).To(gomega.HaveLen(3))
		g.Expect(w.Header().Get("Content-Type")).To(gomega.Equal("text/html;charset=utf-8"))
		g.Expect(w.Header().Get("Content-Language")).To(gomega.Equal("en"))
		g.Expect(w.Header().Get("Vary")).To(gomega.Equal("Accept, Accept-Language"))
	}
}

func Test_should_negotiate_a_default_processor(t *testing.T) {
	g := gomega.NewWithT(t)

	// Given ...
	wildcard := acceptable.OfferOf("text/*")
	a := acceptable.OfferOf("text/test")

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("Accept", "*/*")

	// When ...
	best := acceptable.BestRequestMatch(req, wildcard)

	// Then ...
	g.Expect(best).To(gomega.Equal(&acceptable.Match{
		Type:     "text",
		Subtype:  "plain",
		Language: "*",
		Charset:  "utf-8",
		Vary:     []string{"Accept"},
	}))

	// And when ...
	best = acceptable.BestRequestMatch(req, a)

	// Then ...
	g.Expect(best).To(gomega.Equal(&acceptable.Match{
		Type:     "text",
		Subtype:  "test",
		Language: "*",
		Charset:  "utf-8",
		Vary:     []string{"Accept"},
	}))
}

func Test_should_negotiate_one_of_the_processors(t *testing.T) {
	g := gomega.NewWithT(t)

	// Given ...
	a := acceptable.OfferOf("text/a")
	b := acceptable.OfferOf("text/b")

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("Accept", "text/a, text/b")

	// When ...
	best := acceptable.BestRequestMatch(req, a)

	// Then ...
	g.Expect(best).To(gomega.Equal(&acceptable.Match{
		Type:     "text",
		Subtype:  "a",
		Language: "*",
		Charset:  "utf-8",
		Vary:     []string{"Accept"},
	}))

	// And when ...
	best = acceptable.BestRequestMatch(req, b)

	// Then ...
	g.Expect(best).To(gomega.Equal(&acceptable.Match{
		Type:     "text",
		Subtype:  "b",
		Language: "*",
		Charset:  "utf-8",
		Vary:     []string{"Accept"},
	}))
}

func TestMain(m *testing.M) {
	flag.Parse()
	if testing.Verbose() {
		acceptable.Debug = func(m string, a ...interface{}) { fmt.Printf(m, a...) }
	}
	os.Exit(m.Run())
}

func ExampleRenderBestMatch() {
	// In this example, the same content is available in three languages. Three different
	// approaches can be used.

	// 1. simple values can be used
	en := "Hello!" // get English content

	// 2. values can be wrapped in a data.Data
	fr := data.Of("Bonjour!") // get French content

	// 3. this uses a lazy evaluation function, wrapped in a data.Data
	es := data.Lazy(func(template string, language string, cr bool) (interface{}, *data.Metadata, error) {
		return "Hola!", nil, nil // get Spanish content - eg from database
	})

	// We're implementing an HTTP handler, so we are given a request and a response.

	req := &http.Request{}                               // some incoming request
	var res http.ResponseWriter = httptest.NewRecorder() // replace with the server's response writer

	// Now do the content negotiation. This example has six supported content types, all of them
	// able to serve any of the three example languages.
	//
	// The first offer is for JSON - this is often the most widely used because it also supports
	// Ajax requests.

	acceptable.RenderBestMatch(res, req, "home.html",
		acceptable.OfferOf("application/json", "en").Using(processor.JSON("  ")).
			With("en", en).With("fr", fr).With("es", es),

		acceptable.OfferOf("application/xml").Using(processor.XML("  ")).
			With("en", en).With("fr", fr).With("es", es),

		acceptable.OfferOf("text/csv").Using(processor.CSV()).
			With("en", en).With("fr", fr).With("es", es),

		acceptable.OfferOf("text/plain").Using(processor.TXT()).
			With("en", en).With("fr", fr).With("es", es),

		templates.TextHtmlOffer("en", "templates/en", ".html", nil).
			With("en", en).With("fr", fr).With("es", es),

		templates.ApplicationXhtmlOffer("en", "templates/en", ".html", nil).
			With("en", en).With("fr", fr).With("es", es),
	)

	// RenderBestMatch returns an error which should be checked
}
