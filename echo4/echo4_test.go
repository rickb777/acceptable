package echo4_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/rickb777/acceptable/contenttype"
	. "github.com/rickb777/acceptable/header/headername"

	"github.com/labstack/echo/v4"
	. "github.com/onsi/gomega"
	"github.com/rickb777/acceptable/echo4"
	"github.com/rickb777/acceptable/header"
	"github.com/rickb777/acceptable/offer"
)

func TestBestRequestMatch_should_match_best_offer(t *testing.T) {
	g := NewWithT(t)

	// Given ...
	oa := offer.Of(echo4.TXT(), TextPlain).With("foo", "en")
	ob := offer.Of(echo4.CSV(), TextCSV).With("bar", "en")
	oc := offer.Of(echo4.JSON(), ApplicationJSON).With("hello", "en")
	od := offer.Of(echo4.XML("x"), ApplicationXML).With("zzz", "en")
	oe := offer.Of(echo4.Binary(), ApplicationBinary).With("10101", "en")

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(Accept, "application/json, application/xml;q=0")
	w := httptest.NewRecorder()
	ec := e.NewContext(req, w)

	// When ...
	match := echo4.BestRequestMatch(ec, oa, ob, oc, od, oe)

	// Then ...
	g.Expect(match.Data.Content("", "en")).To(Equal("hello"))
	g.Expect(w.Header()).To(HaveLen(0))
}

func TestRenderBestMatch_should_use_default_processor_if_no_accept_header(t *testing.T) {
	g := NewWithT(t)

	// Given ...
	oa := offer.Of(echo4.TXT(), "text/test")
	ob := offer.Of(echo4.TXT(), TextPlain)
	oc := offer.Of(echo4.CSV(), TextCSV)
	od := offer.Of(echo4.XML("x"), ApplicationXML)
	oe := offer.Of(echo4.Binary(), ApplicationBinary)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	ec := e.NewContext(req, w)
	//c.SetPath("/users/:email")
	//c.SetParamNames("email")
	//c.SetParamValues("jon@labstack.com")

	// When ...
	err := echo4.RenderBestMatch(ec, "", oa, ob, oc, od, oe)

	// Then ...
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(w.Header()).To(HaveLen(1))
	g.Expect(w.Header().Get(ContentType)).To(Equal("text/test;charset=utf-8"))
}

func TestRenderBestMatch_should_give_JSON_response_for_ajax_requests(t *testing.T) {
	g := NewWithT(t)

	// Given ...
	oa := offer.Of(echo4.JSON(), "application/json").With("foo", "en")

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Add(XRequestedWith, header.XMLHttpRequest)
	w := httptest.NewRecorder()
	ec := e.NewContext(req, w)

	// When ...
	err := echo4.RenderBestMatch(ec, "", oa)

	// Then ...
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(w.HeaderMap).To(HaveLen(2))
	g.Expect(w.Header().Get(ContentType)).To(Equal("application/json;charset=utf-8"))
	g.Expect(w.Header().Get(ContentLanguage)).To(Equal("en"))
}
