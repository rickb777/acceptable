package acceptable_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/rickb777/acceptable"
	"github.com/rickb777/acceptable/header"
	. "github.com/rickb777/acceptable/headername"
	"github.com/rickb777/acceptable/offer"
	"github.com/rickb777/acceptable/templates"
)

func Test_should_use_default_processor_if_no_accept_header(t *testing.T) {
	g := NewWithT(t)

	// Given ...
	a := offer.Of(offer.TXTProcessor(), "text/test")
	b := offer.Of(offer.TXTProcessor(), "text/plain")

	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	// When ...
	err := acceptable.RenderBestMatch(w, req, 0, "", a, b)

	// Then ...
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(w.Code).To(Equal(204))
	g.Expect(w.Header()).To(HaveLen(1))
	g.Expect(w.Header().Get(ContentType)).To(Equal("text/test;charset=utf-8"))
	g.Expect(w.Body.String()).To(Equal(""))
}

func Test_should_use_catch_all_if_no_matching_accept_header(t *testing.T) {
	g := NewWithT(t)

	// Given ...
	a := offer.Of(offer.TXTProcessor(), "text/csv").With("foo", "*")
	b := offer.Of(offer.TXTProcessor(), "").With("bar", "*")

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add(Accept, "image/*, application/*")
	w := httptest.NewRecorder()

	// When ...
	err := acceptable.RenderBestMatch(w, req, 201, "", a, b)

	// Then ...
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(w.Code).To(Equal(201))
	g.Expect(w.Header()).To(HaveLen(2))
	g.Expect(w.Header().Get(ContentType)).To(Equal("application/octet-stream"))
	g.Expect(w.Header().Get(Vary)).To(Equal("Accept"))
	g.Expect(w.Body.String()).To(Equal("bar\n"))
}

func Test_should_return_406_if_no_matching_accept_header(t *testing.T) {
	g := NewWithT(t)

	cases := []string{"application/xml", "text/test"}

	for _, c := range cases {
		// Given ...
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Add(Accept, "image/png")
		w := httptest.NewRecorder()
		a := offer.Of(offer.JSONProcessor(), c)

		// When ...
		err := acceptable.RenderBestMatch(w, req, 0, "", a)

		// Then ...
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(w.Code).To(Equal(406))
	}
}

func Test_should_return_406_if_there_are_no_offers(t *testing.T) {
	g := NewWithT(t)

	// Given ...
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add(Accept, "image/png")
	w := httptest.NewRecorder()

	// When ...
	err := acceptable.RenderBestMatch(w, req, 0, "")

	// Then ...
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(w.Code).To(Equal(406))
}

func Test_should_give_JSON_response_for_ajax_requests(t *testing.T) {
	g := NewWithT(t)

	// Given ...
	a := offer.Of(offer.JSONProcessor(), "application/json")

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add(XRequestedWith, header.XMLHttpRequest)
	w := httptest.NewRecorder()

	// When ...
	err := acceptable.RenderBestMatch(w, req, 0, "", a)

	// Then ...
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(w.HeaderMap).To(HaveLen(1))
	g.Expect(w.Header().Get(ContentType)).To(Equal("application/json;charset=utf-8"))
}

func Test_should_give_406_for_unmatched_ajax_requests(t *testing.T) {
	g := NewWithT(t)

	// Given ...
	a := offer.Of(offer.JSONProcessor(), "text/plain")

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add(XRequestedWith, header.XMLHttpRequest)
	w := httptest.NewRecorder()

	// When ...
	err := acceptable.RenderBestMatch(w, req, 0, "", a)

	// Then ...
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(w.Code).To(Equal(406))
}

func Test_should_return_406_if_there_are_no_offers_for_ajax(t *testing.T) {
	g := NewWithT(t)

	// Given ...
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add(Accept, "image/png")
	req.Header.Add(XRequestedWith, header.XMLHttpRequest)
	w := httptest.NewRecorder()

	// When ...
	err := acceptable.RenderBestMatch(w, req, 0, "")

	// Then ...
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(w.Code).To(Equal(406))
}

func Test_should_return_406_with_fallback_offer(t *testing.T) {
	g := NewWithT(t)

	// Given ...
	a := offer.Of(offer.TXTProcessor(), "text/foo").With(nil, "en-GB")
	b := offer.Of(offer.TXTProcessor(), "text/bar").With(`bad stuff`, "en").CanHandle406As(400)

	req, _ := http.NewRequest("GET", "/", nil)
	// this header means "anything but text/test"
	req.Header.Add(Accept, "text/foo;q=0, text/bar;q=0, */*") // excluded
	req.Header.Add(AcceptLanguage, "en")                      // accepted
	w := httptest.NewRecorder()

	// When ...
	err := acceptable.RenderBestMatch(w, req, 0, "", a, b)

	// Then ...
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(w.Code).To(Equal(400))
	g.Expect(w.Header()).To(HaveLen(2))
	g.Expect(w.Header().Get(ContentType)).To(Equal("text/bar;charset=utf-8"))
	g.Expect(w.Header().Get(ContentLanguage)).To(Equal("en"))
	g.Expect(w.Body.String()).To(Equal("bad stuff\n"))
}

// RFC7231 suggests that 406 is sent when no media range matches are possible.
func Test_should_return_406_when_media_range_is_explicitly_excluded(t *testing.T) {
	g := NewWithT(t)

	// Given ...
	a := offer.Of(offer.TXTProcessor(), "text/foo").With(nil, "en")
	b := offer.Of(offer.TXTProcessor(), "text/bar").With(nil, "en")

	req, _ := http.NewRequest("GET", "/", nil)
	// this header means "anything but text/test"
	req.Header.Add(Accept, "text/foo;q=0, text/bar;q=0, */*") // excluded
	req.Header.Add(AcceptLanguage, "en")                      // accepted
	w := httptest.NewRecorder()

	// When ...
	err := acceptable.RenderBestMatch(w, req, 0, "", a, b)

	// Then ...
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(w.Code).To(Equal(406))
	g.Expect(w.Header()).To(HaveLen(1))
	g.Expect(w.Header().Get(ContentType)).To(Equal("text/plain;charset=utf-8"))
	g.Expect(w.Body.String()).To(Equal("Not Acceptable\n")) // from http.StatusText()
}

// RFC7231 recommends that, when no language matches are possible, a response should be sent anyway.
func Test_should_return_200_even_when_language_is_explicitly_excluded(t *testing.T) {
	g := NewWithT(t)

	// Given ...
	a := offer.Of(offer.TXTProcessor(), "text/test").With(nil, "en")

	req, _ := http.NewRequest("GET", "/", nil)
	// this header means "anything but text/test"
	req.Header.Add(Accept, "text/test, */*")
	req.Header.Add(AcceptLanguage, "en;q=0, *") // anything but "en"
	w := httptest.NewRecorder()

	// When ...
	err := acceptable.RenderBestMatch(w, req, 0, "", a)

	// Then ...
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(w.HeaderMap).To(HaveLen(3))
	g.Expect(w.Header().Get(ContentType)).To(Equal("text/test;charset=utf-8"))
	g.Expect(w.Header().Get(ContentLanguage)).To(Equal("en"))
	g.Expect(w.Header().Get(Vary)).To(Equal("Accept, Accept-Language"))
}

func Test_should_negotiate_using_media_and_language(t *testing.T) {
	g := NewWithT(t)

	// Given ...
	// should be skipped because of media mismatch
	a := offer.Of(offer.TXTProcessor(), "text/html").With(nil, "en")
	// should be skipped because of language mismatch
	b := offer.Of(offer.TXTProcessor(), "text/test").With(nil, "de")
	// should match
	c := offer.Of(offer.TXTProcessor(), "text/test").With(nil, "en")

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add(Accept, "text/test, text/*")
	req.Header.Add(AcceptLanguage, "en-GB, fr-FR")
	w := httptest.NewRecorder()

	// When ...
	err := acceptable.RenderBestMatch(w, req, 0, "", a, b, c)

	// Then ...
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(w.HeaderMap).To(HaveLen(3))
	g.Expect(w.Header().Get(ContentType)).To(Equal("text/test;charset=utf-8"))
	g.Expect(w.Header().Get(ContentLanguage)).To(Equal("en"))
	g.Expect(w.Header().Get(Vary)).To(Equal("Accept, Accept-Language"))
}

func Test_should_render_iso8859_html_using_templates(t *testing.T) {
	g := NewWithT(t)

	// Given ...
	p := templates.Templates("example/templates/en", ".html", nil)
	a := offer.Of(p, "text/html").With(nil, "en")

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add(Accept, "text/html")
	req.Header.Add(AcceptLanguage, "en, fr")
	req.Header.Add(AcceptCharset, "iso-8859-1")
	w := httptest.NewRecorder()

	// When ...
	err := acceptable.RenderBestMatch(w, req, 0, "home.html", a)

	// Then ...
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(w.HeaderMap).To(HaveLen(3))
	g.Expect(w.Header().Get(ContentType)).To(Equal("text/html;charset=windows-1252"))
	g.Expect(w.Header().Get(ContentLanguage)).To(Equal("en"))
	g.Expect(w.Header().Get(Vary)).To(Equal("Accept, Accept-Language, Accept-Charset"))
}

func Test_should_match_utf8_charset_when_acceptable(t *testing.T) {
	g := NewWithT(t)

	// Given ...
	a := offer.Of(offer.TXTProcessor(), "text/html")

	// all these cases contain utf-8 and another
	cases := []string{
		"utf-8, iso-8859-1",
		"utf8, iso-8859-1",
		"iso-8859-1, utf-8",
		"iso-8859-1, utf8",
	}

	for _, cs := range cases {
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Add(Accept, "text/html")
		req.Header.Add(AcceptLanguage, "en")
		req.Header.Add(AcceptCharset, cs)
		w := httptest.NewRecorder()

		// When ...
		err := acceptable.RenderBestMatch(w, req, 0, "", a)

		// Then ...
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(w.HeaderMap).To(HaveLen(3))
		g.Expect(w.Header().Get(ContentType)).To(Equal("text/html;charset=utf-8"))
		g.Expect(w.Header().Get(ContentLanguage)).To(Equal("en"))
		g.Expect(w.Header().Get(Vary)).To(Equal("Accept, Accept-Language"))
	}
}
