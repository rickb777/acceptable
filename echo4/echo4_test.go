package echo4_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	. "github.com/rickb777/acceptable/contenttype"
	"github.com/rickb777/acceptable/echo4"
	"github.com/rickb777/acceptable/header"
	. "github.com/rickb777/acceptable/headername"
	"github.com/rickb777/acceptable/offer"
	"github.com/rickb777/expect"
)

func TestBestRequestMatch_should_match_best_offer(t *testing.T) {
	// Given ...
	oa := offer.Of(offer.TXTProcessor(), TextPlain).With("foo", "en")
	ob := offer.Of(offer.CSVProcessor(), TextCSV).With("bar", "en")
	oc := offer.Of(offer.JSONProcessor(), ApplicationJSON).With("hello", "en")
	od := offer.Of(offer.XMLProcessor("x"), ApplicationXML).With("zzz", "en")
	oe := offer.Of(offer.BinaryProcessor(), ApplicationBinary).With("10101", "en")

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(Accept, "application/json, application/xml;q=0")
	w := httptest.NewRecorder()
	ec := e.NewContext(req, w)

	// When ...
	match := echo4.BestRequestMatch(ec, oa, ob, oc, od, oe)

	// Then ...
	expect.Any(match.Data.Content("", "en")).ToBe(t, "hello")
	expect.Map(w.Header()).ToHaveLength(t, 0)
}

func TestRenderBestMatch_should_use_default_processor_if_no_accept_header(t *testing.T) {
	// Given ...
	oa := offer.Of(offer.TXTProcessor(), "text/test").With("hello world", "*")
	ob := offer.Of(offer.TXTProcessor(), TextPlain).With("hello world", "*")
	oc := offer.Of(offer.CSVProcessor(), TextCSV).With("hello,world", "*")
	od := offer.Of(offer.XMLProcessor("x"), ApplicationXML)
	oe := offer.Of(offer.BinaryProcessor(), ApplicationBinary)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	ec := e.NewContext(req, w)
	//c.SetPath("/users/:email")
	//c.SetParamNames("email")
	//c.SetParamValues("jon@labstack.com")

	// When ...
	err := echo4.RenderBestMatch(ec, 200, "", oa, ob, oc, od, oe)

	// Then ...
	expect.Error(err).Not().ToHaveOccurred(t)
	expect.Number(w.Code).ToBe(t, 200)
	expect.Map(w.Header()).ToHaveLength(t, 1)
	expect.String(w.Header().Get(ContentType)).ToBe(t, "text/test;charset=utf-8")
}

func TestRenderBestMatch_should_give_JSON_response_for_ajax_requests(t *testing.T) {
	// Given ...
	oa := offer.JSON().With("foo", "en")

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Add(XRequestedWith, header.XMLHttpRequest)
	w := httptest.NewRecorder()
	ec := e.NewContext(req, w)

	// When ...
	err := echo4.RenderBestMatch(ec, 201, "", oa)

	// Then ...
	expect.Error(err).Not().ToHaveOccurred(t)
	expect.Number(w.Code).ToBe(t, 201)
	expect.Map(w.HeaderMap).ToHaveLength(t, 2)
	expect.String(w.Header().Get(ContentType)).ToBe(t, "application/json;charset=utf-8")
	expect.String(w.Header().Get(ContentLanguage)).ToBe(t, "en")
}
