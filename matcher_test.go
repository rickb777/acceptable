package acceptable_test

import (
	"flag"
	"net/http"
	"os"
	"testing"

	"github.com/rickb777/acceptable"
	"github.com/rickb777/acceptable/data"
	"github.com/rickb777/acceptable/header"
	. "github.com/rickb777/acceptable/headername"
	"github.com/rickb777/acceptable/offer"
	"github.com/rickb777/expect"
)

func Test_should_return_wildcard_data_for_any_language(t *testing.T) {
	// Given ...
	a := offer.Of(offer.TXTProcessor(), "text/test").With(someSliceData, "*")

	for _, lang := range []string{"en", "de"} {
		req, _ := http.NewRequest("GET", "/", nil)
		// this header means "anything but text/test"
		req.Header.Add(Accept, "text/test, */*")
		req.Header.Add(AcceptLanguage, lang)

		// When ...
		best := acceptable.BestRequestMatch(req, a)

		// Then ...
		expect.Any(best.Render).I(lang).Not().ToBeNil(t)
		best.Render = nil

		expect.Any(best).I(lang).ToBe(t, &offer.Match{
			ContentType: header.ContentType{Type: "text", Subtype: "test"},
			Language:    lang,
			Charset:     "utf-8",
			Vary:        []string{Accept, AcceptLanguage},
			Data:        data.Of(someSliceData),
		})
	}
}

func Test_should_match_subtype_wildcard1(t *testing.T) {
	// Given ...
	a := offer.Of(nil, "text/test")

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add(Accept, "text/*") // <-- wildcard

	// When ...
	best := acceptable.BestRequestMatch(req, a)

	// Then ...
	expect.Any(best).ToBe(t, &offer.Match{
		ContentType: header.ContentType{Type: "text", Subtype: "test"},
		Language:    "*",
		Charset:     "utf-8",
		Vary:        []string{Accept},
	})
}

func Test_should_match_subtype_wildcard2(t *testing.T) {
	// Given ...
	a := offer.Of(nil, "text/*") // <-- wildcard

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add(Accept, "text/test")

	// When ...
	best := acceptable.BestRequestMatch(req, a)

	// Then ...
	expect.Any(best).ToBe(t, &offer.Match{
		ContentType: header.ContentType{Type: "text", Subtype: "test"},
		Language:    "*",
		Charset:     "utf-8",
		Vary:        []string{Accept},
	})
}

func Test_should_match_language_when_offer_language_is_not_specified(t *testing.T) {
	// Given ...
	a := offer.Of(nil, "text/html")

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add(Accept, "application/json, text/html")
	req.Header.Add(AcceptLanguage, "en, fr")
	req.Header.Add(AcceptCharset, "utf-8, iso-8859-1")

	// When ...
	best := acceptable.BestRequestMatch(req, a)

	// Then ...
	expect.Any(best).ToBe(t, &offer.Match{
		ContentType: header.ContentType{Type: "text", Subtype: "html"},
		Language:    "en",
		Charset:     "utf-8",
		Vary:        []string{Accept, AcceptLanguage},
	})
}

func Test_should_match_language_wildcard_and_return_selected_language(t *testing.T) {
	// Given ...
	a := offer.Of(nil, "").With(nil, "en")
	b := offer.Of(nil, "").With(nil, "de")

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add(AcceptLanguage, "*")

	// When ...
	best := acceptable.BestRequestMatch(req, a, b)

	// Then ...
	expect.Any(best).ToBe(t, &offer.Match{
		ContentType: header.ContentType{Type: "application", Subtype: "octet-stream"},
		Language:    "en",
		Charset:     "utf-8",
		Vary:        []string{AcceptLanguage},
	})
}

func Test_should_select_language_of_first_matched_offer_when_no_language_matches(t *testing.T) {
	// Given ...
	a := offer.Of(nil, "text/csv").With(someSliceData, "es")
	b := offer.Of(nil, "text/html").With(someMapData, "en").With(someMapData, "pt")
	c := offer.Of(nil, "text/html").With(someMapData, "de")

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add(Accept, "text/html")
	req.Header.Add(AcceptLanguage, "fr")

	// When ...
	best := acceptable.BestRequestMatch(req, a, b, c)

	// Then ...
	expect.Any(best).ToBe(t, &offer.Match{
		ContentType: header.ContentType{Type: "text", Subtype: "html"},
		Language:    "en",
		Charset:     "utf-8",
		Vary:        []string{Accept, AcceptLanguage},
		Data:        data.Of(someMapData),
	})
}

func Test_should_negotiate_a_default_processor(t *testing.T) {
	// Given ...
	wildcard := offer.Of(nil, "text/*")
	a := offer.Of(nil, "text/test")

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add(Accept, "*/*")

	// When ...
	best := acceptable.BestRequestMatch(req, wildcard)

	// Then ...
	expect.Any(best).ToBe(t, &offer.Match{
		ContentType: header.ContentType{Type: "text", Subtype: "plain"},
		Language:    "*",
		Charset:     "utf-8",
		Vary:        []string{Accept},
	})

	// And when ...
	best = acceptable.BestRequestMatch(req, a)

	// Then ...
	expect.Any(best).ToBe(t, &offer.Match{
		ContentType: header.ContentType{Type: "text", Subtype: "test"},
		Language:    "*",
		Charset:     "utf-8",
		Vary:        []string{Accept},
	})
}

func Test_should_negotiate_one_of_the_processors(t *testing.T) {
	// Given ...
	a := offer.Of(nil, "text/a")
	b := offer.Of(nil, "text/b")

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add(Accept, "text/a, text/b")

	// When ...
	best := acceptable.BestRequestMatch(req, a)

	// Then ...
	expect.Any(best).ToBe(t, &offer.Match{
		ContentType: header.ContentType{Type: "text", Subtype: "a"},
		Language:    "*",
		Charset:     "utf-8",
		Vary:        []string{Accept},
	})

	// And when ...
	best = acceptable.BestRequestMatch(req, b)

	// Then ...
	expect.Any(best).ToBe(t, &offer.Match{
		ContentType: header.ContentType{Type: "text", Subtype: "b"},
		Language:    "*",
		Charset:     "utf-8",
		Vary:        []string{Accept},
	})
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
