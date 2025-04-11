package acceptable_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rickb777/acceptable"
	"github.com/rickb777/acceptable/header"
	. "github.com/rickb777/acceptable/headername"
	"github.com/rickb777/acceptable/offer"
	"github.com/rickb777/acceptable/templates"
	"github.com/rickb777/expect"
)

func Test_should_return_no_content_if_no_offers_have_data(t *testing.T) {
	// Given ...
	a := offer.Of(offer.TXTProcessor(), "text/test")
	b := offer.Of(offer.TXTProcessor(), "text/plain")

	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	// When ...
	err := acceptable.RenderBestMatch(w, req, 0, "", a, b)

	// Then ...
	expect.Error(err).Not().ToHaveOccurred(t)
	expect.Number(w.Code).ToBe(t, 204)
	expect.Map(w.Header()).ToHaveLength(t, 0)
	expect.String(w.Body.String()).ToBe(t, "")
}

func Test_should_use_default_processor_if_no_accept_header(t *testing.T) {
	// Given ...
	a := offer.Of(offer.TXTProcessor(), "text/test").With(nil, "en")
	b := offer.Of(offer.TXTProcessor(), "text/plain").With("foo", "en")

	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	// When ...
	err := acceptable.RenderBestMatch(w, req, 0, "", a, b)

	// Then ...
	expect.Error(err).Not().ToHaveOccurred(t)
	expect.Number(w.Code).ToBe(t, 204)
	expect.Map(w.Header()).ToHaveLength(t, 2)
	expect.String(w.Header().Get(ContentType)).ToBe(t, "text/test;charset=utf-8")
	expect.String(w.Header().Get(ContentLanguage)).ToBe(t, "en")
	expect.String(w.Body.String()).ToBe(t, "")
}

func Test_should_use_catch_all_if_no_matching_accept_header(t *testing.T) {
	// Given ...
	a := offer.Of(offer.TXTProcessor(), "text/csv").With("foo", "*")
	b := offer.Of(offer.TXTProcessor(), "").With("bar", "*")

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add(Accept, "image/*, application/*")
	w := httptest.NewRecorder()

	// When ...
	err := acceptable.RenderBestMatch(w, req, 201, "", a, b)

	// Then ...
	expect.Error(err).Not().ToHaveOccurred(t)
	expect.Number(w.Code).ToBe(t, 201)
	expect.Map(w.Header()).ToHaveLength(t, 2)
	expect.String(w.Header().Get(ContentType)).ToBe(t, "application/octet-stream")
	expect.String(w.Header().Get(Vary)).ToBe(t, "Accept")
	expect.String(w.Body.String()).ToBe(t, "bar\n")
}

func Test_should_return_406_if_no_matching_accept_header(t *testing.T) {
	cases := []string{"application/xml", "text/test"}

	for _, c := range cases {
		// Given ...
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Add(Accept, "image/png")
		w := httptest.NewRecorder()
		a := offer.Of(offer.JSONProcessor(), c).With("foo", "en")

		// When ...
		err := acceptable.RenderBestMatch(w, req, 0, "", a)

		// Then ...
		expect.Error(err).Not().ToHaveOccurred(t)
		expect.Number(w.Code).ToBe(t, 406)
	}
}

func Test_should_return_204_if_there_are_no_offers(t *testing.T) {
	// Given ...
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add(Accept, "image/png")
	w := httptest.NewRecorder()

	// When ...
	err := acceptable.RenderBestMatch(w, req, 0, "")

	// Then ...
	expect.Error(err).Not().ToHaveOccurred(t)
	expect.Number(w.Code).ToBe(t, 204)
}

func Test_should_give_JSON_response_for_ajax_requests(t *testing.T) {
	// Given ...
	a := offer.Of(offer.JSONProcessor(), "application/json").With(`"foo"`, "en")

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add(XRequestedWith, header.XMLHttpRequest)
	w := httptest.NewRecorder()

	// When ...
	err := acceptable.RenderBestMatch(w, req, 0, "", a)

	// Then ...
	expect.Error(err).Not().ToHaveOccurred(t)
	expect.Map(w.Header()).ToHaveLength(t, 2)
	expect.String(w.Header().Get(ContentType)).ToBe(t, "application/json;charset=utf-8")
	expect.String(w.Header().Get(ContentLanguage)).ToBe(t, "en")
}

func Test_should_give_406_for_unmatched_ajax_requests(t *testing.T) {
	// Given ...
	a := offer.Of(offer.JSONProcessor(), "text/plain").With("foo", "en")

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add(XRequestedWith, header.XMLHttpRequest)
	w := httptest.NewRecorder()

	// When ...
	err := acceptable.RenderBestMatch(w, req, 0, "", a)

	// Then ...
	expect.Error(err).Not().ToHaveOccurred(t)
	expect.Number(w.Code).ToBe(t, 406)
}

func Test_should_return_204_if_there_are_no_offers_for_ajax(t *testing.T) {
	// Given ...
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add(Accept, "image/png")
	req.Header.Add(XRequestedWith, header.XMLHttpRequest)
	w := httptest.NewRecorder()

	// When ...
	err := acceptable.RenderBestMatch(w, req, 0, "")

	// Then ...
	expect.Error(err).Not().ToHaveOccurred(t)
	expect.Number(w.Code).ToBe(t, 204)
}

func Test_should_return_406_with_fallback_offer(t *testing.T) {
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
	expect.Error(err).Not().ToHaveOccurred(t)
	expect.Number(w.Code).ToBe(t, 400)
	expect.Map(w.Header()).ToHaveLength(t, 2)
	expect.String(w.Header().Get(ContentType)).ToBe(t, "text/bar;charset=utf-8")
	expect.String(w.Header().Get(ContentLanguage)).ToBe(t, "en")
	expect.String(w.Body.String()).ToBe(t, "bad stuff\n")
}

// RFC7231 suggests that 406 is sent when no media range matches are possible.
func Test_should_return_406_when_media_range_is_explicitly_excluded(t *testing.T) {
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
	expect.Error(err).Not().ToHaveOccurred(t)
	expect.Number(w.Code).ToBe(t, 406)
	expect.Map(w.Header()).ToHaveLength(t, 1)
	expect.String(w.Header().Get(ContentType)).ToBe(t, "text/plain;charset=utf-8")
	expect.String(w.Body.String()).ToBe(t, "Not Acceptable\n") // from http.StatusText()
}

// RFC7231 recommends that, when no language matches are possible, a response should be sent anyway.
func Test_should_return_200_even_when_language_is_explicitly_excluded(t *testing.T) {
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
	expect.Error(err).Not().ToHaveOccurred(t)
	expect.Map(w.Header()).ToHaveLength(t, 3)
	expect.String(w.Header().Get(ContentType)).ToBe(t, "text/test;charset=utf-8")
	expect.String(w.Header().Get(ContentLanguage)).ToBe(t, "en")
	expect.String(w.Header().Get(Vary)).ToBe(t, "Accept, Accept-Language")
}

func Test_should_negotiate_using_media_and_language(t *testing.T) {
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
	expect.Error(err).Not().ToHaveOccurred(t)

	expect.Map(w.Header()).ToHaveLength(t, 3)
	expect.String(w.Header().Get(ContentType)).ToBe(t, "text/test;charset=utf-8")
	expect.String(w.Header().Get(ContentLanguage)).ToBe(t, "en")
	expect.String(w.Header().Get(Vary)).ToBe(t, "Accept, Accept-Language")
}

func Test_should_render_iso8859_html_using_templates(t *testing.T) {
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
	expect.Error(err).Not().ToHaveOccurred(t)
	expect.Error(err).Not().ToHaveOccurred(t)
	expect.Map(w.Header()).ToHaveLength(t, 3)
	expect.String(w.Header().Get(ContentType)).ToBe(t, "text/html;charset=windows-1252")
	expect.String(w.Header().Get(ContentLanguage)).ToBe(t, "en")
	expect.String(w.Header().Get(Vary)).ToBe(t, "Accept, Accept-Language, Accept-Charset")
}

func Test_should_match_utf8_charset_when_acceptable(t *testing.T) {
	// Given ...
	a := offer.Of(offer.TXTProcessor(), "text/html").With("foo", "en")

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
		expect.Error(err).Not().ToHaveOccurred(t)
		expect.Map(w.Header()).ToHaveLength(t, 3)
		expect.String(w.Header().Get(ContentType)).ToBe(t, "text/html;charset=utf-8")
		expect.String(w.Header().Get(ContentLanguage)).ToBe(t, "en")
		expect.String(w.Header().Get(Vary)).ToBe(t, "Accept, Accept-Language")
	}
}
