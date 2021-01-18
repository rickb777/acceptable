package echo4_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/onsi/gomega"
	"github.com/rickb777/acceptable/echo4"
	"github.com/rickb777/acceptable/header"
	"github.com/rickb777/acceptable/offer"
)

func Test_should_use_default_processor_if_no_accept_header(t *testing.T) {
	g := gomega.NewWithT(t)

	// Given ...
	a := offer.Of(echo4.TXT(), "text/test")
	b := offer.Of(echo4.TXT(), "text/plain")

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	c := e.NewContext(req, w)
	//c.SetPath("/users/:email")
	//c.SetParamNames("email")
	//c.SetParamValues("jon@labstack.com")

	// When ...
	err := echo4.RenderBestMatch(c, "", a, b)

	// Then ...
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(w.Header()).To(gomega.HaveLen(1))
	g.Expect(w.Header().Get("Content-Type")).To(gomega.Equal("text/test;charset=utf-8"))
}

func Test_should_give_JSON_response_for_ajax_requests(t *testing.T) {
	g := gomega.NewWithT(t)

	// Given ...
	a := offer.Of(echo4.JSON(), "application/json")

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Add(header.XRequestedWith, header.XMLHttpRequest)
	w := httptest.NewRecorder()
	c := e.NewContext(req, w)

	// When ...
	best := echo4.BestRequestMatch(c, a)
	err := best.Render(w, req, *best, "")

	// Then ...
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(w.HeaderMap).To(gomega.HaveLen(1))
	g.Expect(w.Header().Get("Content-Type")).To(gomega.Equal("application/json;charset=utf-8"))
}
