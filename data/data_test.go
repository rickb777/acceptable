package data

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rickb777/acceptable/header"

	. "github.com/onsi/gomega"
)

func TestValue_future_expiry(t *testing.T) {
	g := NewGomegaWithT(t)

	// Given ...
	d := Lazy(func(template, language string, cr bool) (interface{}, string, error) {
		g.Expect(template).To(Equal("home.html"))
		g.Expect(language).To(Equal("en"))
		return "foo", "abcdef", nil
	}).
		Expires(time.Date(2020, 2, 3, 1, 1, 1, 0, time.UTC)).
		LastModified(time.Date(2020, 1, 1, 1, 1, 1, 0, time.UTC)).
		MaxAge(10 * time.Second)

	req := &http.Request{}
	w := httptest.NewRecorder()

	// When ...
	c, err := GetContentAndApplyExtraHeaders(w, req, d, "home.html", "en")

	// Then ...
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(c).To(Equal("foo"))
	g.Expect(w.HeaderMap).To(HaveLen(4))
	g.Expect(w.Header().Get("Cache-Control")).To(Equal("max-age=10"))
	g.Expect(w.Header().Get("Expires")).To(Equal("Mon, 03 Feb 2020 01:01:01 UTC"))
	g.Expect(w.Header().Get("Last-Modified")).To(Equal("Wed, 01 Jan 2020 01:01:01 UTC"))
	g.Expect(w.Header().Get("ETag")).To(Equal(`"abcdef"`))
}

func TestValue_no_cache(t *testing.T) {
	g := NewGomegaWithT(t)

	// Given ...
	d := Of("foo").NoCache().With("Abc", "1", "Def", "true")

	req := &http.Request{}
	w := httptest.NewRecorder()

	// When ...
	c, err := GetContentAndApplyExtraHeaders(w, req, d, "home.html", "en")

	// Then ...
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(c).To(Equal("foo"))
	g.Expect(w.HeaderMap).To(HaveLen(4))
	g.Expect(w.Header().Get("Cache-Control")).To(Equal("no-cache, must-revalidate"))
	g.Expect(w.Header().Get("Pragma")).To(Equal("no-cache"))
	g.Expect(w.Header().Get("Abc")).To(Equal("1"))
	g.Expect(w.Header().Get("Def")).To(Equal("true"))
}

func TestValue_not_modified(t *testing.T) {
	g := NewGomegaWithT(t)

	// Given ...
	d := Of("foo", "hash123").NoCache().With("Abc", "1", "Def", "true")

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set(header.IfNoneMatch, `"foo", "hash123", "bar"`)
	w := httptest.NewRecorder()

	// When ...
	c, err := GetContentAndApplyExtraHeaders(w, req, d, "home.html", "en")

	// Then ...
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(c).To(BeNil())
	g.Expect(w.Code).To(Equal(304))
	g.Expect(w.HeaderMap).To(HaveLen(5))
	g.Expect(w.Header().Get("ETag")).To(Equal(`"hash123"`))
	g.Expect(w.Header().Get("Cache-Control")).To(Equal("no-cache, must-revalidate"))
	g.Expect(w.Header().Get("Pragma")).To(Equal("no-cache"))
	g.Expect(w.Header().Get("Abc")).To(Equal("1"))
	g.Expect(w.Header().Get("Def")).To(Equal("true"))
}

func TestValue_error(t *testing.T) {
	g := NewGomegaWithT(t)

	// Given ...
	d := Lazy(func(template, language string, cr bool) (interface{}, string, error) {
		g.Expect(template).To(Equal("home.html"))
		g.Expect(language).To(Equal("en"))
		return nil, "", errors.New("expected error")
	})

	req := &http.Request{}
	w := httptest.NewRecorder()

	// When ...
	_, err := GetContentAndApplyExtraHeaders(w, req, d, "home.html", "en")

	// Then ...
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(Equal("expected error"))
}
