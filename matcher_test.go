package acceptable_test

import (
	"flag"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/benmoss/matchers"
	. "github.com/onsi/gomega"
	"github.com/rickb777/acceptable"
	"github.com/rickb777/acceptable/data"
	"github.com/rickb777/acceptable/header"
	. "github.com/rickb777/acceptable/header/headername"
	"github.com/rickb777/acceptable/offer"
	"github.com/rickb777/acceptable/templates"
)

func Test_should_use_default_processor_if_no_accept_header(t *testing.T) {
	g := NewWithT(t)

	// Given ...
	a := offer.Of(acceptable.TXT(), "text/test")
	b := offer.Of(acceptable.TXT(), "text/plain")

	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	// When ...
	err := acceptable.RenderBestMatch(w, req, "", a, b)

	// Then ...
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(w.Header()).To(HaveLen(1))
	g.Expect(w.Header().Get(ContentType)).To(Equal("text/test;charset=utf-8"))
}

func Test_should_give_JSON_response_for_ajax_requests(t *testing.T) {
	g := NewWithT(t)

	// Given ...
	a := offer.Of(acceptable.JSON(), "application/json")

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add(XRequestedWith, header.XMLHttpRequest)
	w := httptest.NewRecorder()

	// When ...
	err := acceptable.RenderBestMatch(w, req, "", a)

	// Then ...
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(w.HeaderMap).To(HaveLen(1))
	g.Expect(w.Header().Get(ContentType)).To(Equal("application/json;charset=utf-8"))
}

func Test_should_give_406_for_unmatched_ajax_requests(t *testing.T) {
	g := NewWithT(t)

	// Given ...
	a := offer.Of(acceptable.JSON(), "text/plain")

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add(XRequestedWith, header.XMLHttpRequest)
	w := httptest.NewRecorder()

	// When ...
	err := acceptable.RenderBestMatch(w, req, "", a)

	// Then ...
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(w.Code).To(Equal(406))
}

func Test_should_return_406_if_no_matching_accept_header(t *testing.T) {
	g := NewWithT(t)

	cases := []string{"application/xml", "text/test"}

	for _, c := range cases {
		// Given ...
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Add(Accept, "image/png")
		w := httptest.NewRecorder()
		a := offer.Of(acceptable.JSON(), c)

		// When ...
		err := acceptable.RenderBestMatch(w, req, "", a)

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
	err := acceptable.RenderBestMatch(w, req, "")

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
	err := acceptable.RenderBestMatch(w, req, "")

	// Then ...
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(w.Code).To(Equal(406))
}

// RFC7231 suggests that 406 is sent when no media range matches are possible.
func Test_should_return_406_when_media_range_is_explicitly_excluded(t *testing.T) {
	g := NewWithT(t)

	// Given ...
	a := offer.Of(acceptable.TXT(), "text/test").With(nil, "en")

	req, _ := http.NewRequest("GET", "/", nil)
	// this header means "anything but text/test"
	req.Header.Add(Accept, "text/test;q=0, */*") // excluded
	req.Header.Add(AcceptLanguage, "en")         // accepted
	w := httptest.NewRecorder()

	// When ...
	err := acceptable.RenderBestMatch(w, req, "", a)

	// Then ...
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(w.Code).To(Equal(406))
}

// RFC7231 recommends that, when no language matches are possible, a response should be sent anyway.
func Test_should_return_200_even_when_language_is_explicitly_excluded(t *testing.T) {
	g := NewWithT(t)

	// Given ...
	a := offer.Of(acceptable.TXT(), "text/test").With(nil, "en")

	req, _ := http.NewRequest("GET", "/", nil)
	// this header means "anything but text/test"
	req.Header.Add(Accept, "text/test, */*")
	req.Header.Add(AcceptLanguage, "en;q=0, *") // anything but "en"
	w := httptest.NewRecorder()

	// When ...
	err := acceptable.RenderBestMatch(w, req, "", a)

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
	a := offer.Of(acceptable.TXT(), "text/html").With(nil, "en")
	// should be skipped because of language mismatch
	b := offer.Of(acceptable.TXT(), "text/test").With(nil, "de")
	// should match
	c := offer.Of(acceptable.TXT(), "text/test").With(nil, "en")

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add(Accept, "text/test, text/*")
	req.Header.Add(AcceptLanguage, "en-GB, fr-FR")
	w := httptest.NewRecorder()

	// When ...
	err := acceptable.RenderBestMatch(w, req, "", a, b, c)

	// Then ...
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(w.HeaderMap).To(HaveLen(3))
	g.Expect(w.Header().Get(ContentType)).To(Equal("text/test;charset=utf-8"))
	g.Expect(w.Header().Get(ContentLanguage)).To(Equal("en"))
	g.Expect(w.Header().Get(Vary)).To(Equal("Accept, Accept-Language"))
}

func Test_should_return_wildcard_data_for_any_language(t *testing.T) {
	g := NewWithT(t)

	// Given ...
	a := offer.Of(acceptable.TXT(), "text/test").With(someSliceData, "*")

	for _, lang := range []string{"en", "de"} {
		req, _ := http.NewRequest("GET", "/", nil)
		// this header means "anything but text/test"
		req.Header.Add(Accept, "text/test, */*")
		req.Header.Add(AcceptLanguage, lang)

		// When ...
		best := acceptable.BestRequestMatch(req, a)

		// Then ...
		g.Expect(best.Render).NotTo(BeNil(), lang)
		best.Render = nil

		g.Expect(best).To(matchers.DeepEqual(&offer.Match{
			ContentType: header.ContentType{Type: "text", Subtype: "test"},
			Language:    lang,
			Charset:     "utf-8",
			Vary:        []string{Accept, AcceptLanguage},
			Data:        data.Of(someSliceData),
			Render:      nil,
		}), lang)
	}
}

func Test_should_match_subtype_wildcard1(t *testing.T) {
	g := NewWithT(t)

	// Given ...
	a := offer.Of(nil, "text/test")

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add(Accept, "text/*") // <-- wildcard

	// When ...
	best := acceptable.BestRequestMatch(req, a)

	// Then ...
	g.Expect(best).To(matchers.DeepEqual(&offer.Match{
		ContentType: header.ContentType{Type: "text", Subtype: "test"},
		Language:    "*",
		Charset:     "utf-8",
		Vary:        []string{Accept},
	}))
}

func Test_should_match_subtype_wildcard2(t *testing.T) {
	g := NewWithT(t)

	// Given ...
	a := offer.Of(nil, "text/*") // <-- wildcard

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add(Accept, "text/test")

	// When ...
	best := acceptable.BestRequestMatch(req, a)

	// Then ...
	g.Expect(best).To(matchers.DeepEqual(&offer.Match{
		ContentType: header.ContentType{Type: "text", Subtype: "test"},
		Language:    "*",
		Charset:     "utf-8",
		Vary:        []string{Accept},
	}))
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
	err := acceptable.RenderBestMatch(w, req, "home.html", a)

	// Then ...
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(w.HeaderMap).To(HaveLen(3))
	g.Expect(w.Header().Get(ContentType)).To(Equal("text/html;charset=windows-1252"))
	g.Expect(w.Header().Get(ContentLanguage)).To(Equal("en"))
	g.Expect(w.Header().Get(Vary)).To(Equal("Accept, Accept-Language, Accept-Charset"))
}

func Test_should_match_language_when_offer_language_is_not_specified(t *testing.T) {
	g := NewWithT(t)

	// Given ...
	a := offer.Of(nil, "text/html")

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add(Accept, "application/json, text/html")
	req.Header.Add(AcceptLanguage, "en, fr")
	req.Header.Add(AcceptCharset, "utf-8, iso-8859-1")

	// When ...
	best := acceptable.BestRequestMatch(req, a)

	// Then ...
	g.Expect(best).To(matchers.DeepEqual(&offer.Match{
		ContentType: header.ContentType{Type: "text", Subtype: "html"},
		Language:    "en",
		Charset:     "utf-8",
		Vary:        []string{Accept, AcceptLanguage},
	}))
}

func Test_should_match_language_wildcard_and_return_selected_language(t *testing.T) {
	g := NewWithT(t)

	// Given ...
	a := offer.Of(nil, "").With(nil, "en")
	b := offer.Of(nil, "").With(nil, "de")

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add(AcceptLanguage, "*")

	// When ...
	best := acceptable.BestRequestMatch(req, a, b)

	// Then ...
	g.Expect(best).To(matchers.DeepEqual(&offer.Match{
		ContentType: header.ContentType{Type: "application", Subtype: "octet-stream"},
		Language:    "en",
		Charset:     "utf-8",
		Vary:        []string{AcceptLanguage},
	}))
}

func Test_should_select_language_of_first_matched_offer_when_no_language_matches(t *testing.T) {
	g := NewWithT(t)

	// Given ...
	a := offer.Of(nil, "text/csv").With(someSliceData, "es")
	b := offer.Of(nil, "text/html").With(someMapData, "en")
	c := offer.Of(nil, "text/html").With(someMapData, "de")

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add(Accept, "text/html")
	req.Header.Add(AcceptLanguage, "fr")

	// When ...
	best := acceptable.BestRequestMatch(req, a, b, c)

	// Then ...
	g.Expect(best).To(matchers.DeepEqual(&offer.Match{
		ContentType: header.ContentType{Type: "text", Subtype: "html"},
		Language:    "en",
		Charset:     "utf-8",
		Vary:        []string{Accept, AcceptLanguage},
		Data:        data.Of(someMapData),
		Render:      nil,
	}))
}

func Test_should_match_utf8_charset_when_acceptable(t *testing.T) {
	g := NewWithT(t)

	// Given ...
	a := offer.Of(acceptable.TXT(), "text/html")

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
		err := acceptable.RenderBestMatch(w, req, "", a)

		// Then ...
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(w.HeaderMap).To(HaveLen(3))
		g.Expect(w.Header().Get(ContentType)).To(Equal("text/html;charset=utf-8"))
		g.Expect(w.Header().Get(ContentLanguage)).To(Equal("en"))
		g.Expect(w.Header().Get(Vary)).To(Equal("Accept, Accept-Language"))
	}
}

func Test_should_negotiate_a_default_processor(t *testing.T) {
	g := NewWithT(t)

	// Given ...
	wildcard := offer.Of(nil, "text/*")
	a := offer.Of(nil, "text/test")

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add(Accept, "*/*")

	// When ...
	best := acceptable.BestRequestMatch(req, wildcard)

	// Then ...
	g.Expect(best).To(matchers.DeepEqual(&offer.Match{
		ContentType: header.ContentType{Type: "text", Subtype: "plain"},
		Language:    "*",
		Charset:     "utf-8",
		Vary:        []string{Accept},
	}))

	// And when ...
	best = acceptable.BestRequestMatch(req, a)

	// Then ...
	g.Expect(best).To(matchers.DeepEqual(&offer.Match{
		ContentType: header.ContentType{Type: "text", Subtype: "test"},
		Language:    "*",
		Charset:     "utf-8",
		Vary:        []string{Accept},
	}))
}

func Test_should_negotiate_one_of_the_processors(t *testing.T) {
	g := NewWithT(t)

	// Given ...
	a := offer.Of(nil, "text/a")
	b := offer.Of(nil, "text/b")

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add(Accept, "text/a, text/b")

	// When ...
	best := acceptable.BestRequestMatch(req, a)

	// Then ...
	g.Expect(best).To(matchers.DeepEqual(&offer.Match{
		ContentType: header.ContentType{Type: "text", Subtype: "a"},
		Language:    "*",
		Charset:     "utf-8",
		Vary:        []string{Accept},
	}))

	// And when ...
	best = acceptable.BestRequestMatch(req, b)

	// Then ...
	g.Expect(best).To(matchers.DeepEqual(&offer.Match{
		ContentType: header.ContentType{Type: "text", Subtype: "b"},
		Language:    "*",
		Charset:     "utf-8",
		Vary:        []string{Accept},
	}))
}

func TestMain(m *testing.M) {
	flag.Parse()
	//if testing.Verbose() {
	//	acceptable.Debug = func(m string, a ...interface{}) { fmt.Printf(m, a...) }
	//}
	os.Exit(m.Run())
}

var someMapData = map[string]string{
	"hay":  "horses",
	"beef": "or mutton",
}

var someSliceData = []string{
	"hay is for horses",
	"beef or mutton",
}
