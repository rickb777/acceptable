package data

import (
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	. "github.com/onsi/gomega"
)

func TestValue_future_expiry(t *testing.T) {
	g := NewGomegaWithT(t)

	// Given ...
	d := Lazy(func(template, language string) (interface{}, string, error) {
		g.Expect(template).To(Equal("home.html"))
		g.Expect(language).To(Equal("en"))
		return "foo", "abcdef", nil
	}).
		Expires(time.Date(2020, 2, 3, 1, 1, 1, 0, time.UTC)).
		LastModified(time.Date(2020, 1, 1, 1, 1, 1, 0, time.UTC)).
		MaxAge(10 * time.Second)

	w := httptest.NewRecorder()

	// When ...
	c, err := GetContentAndApplyExtraHeaders(w, d, "home.html", "en")

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

	w := httptest.NewRecorder()

	// When ...
	c, err := GetContentAndApplyExtraHeaders(w, d, "home.html", "en")

	// Then ...
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(c).To(Equal("foo"))
	g.Expect(w.HeaderMap).To(HaveLen(4))
	g.Expect(w.Header().Get("Cache-Control")).To(Equal("no-cache, must-revalidate"))
	g.Expect(w.Header().Get("Pragma")).To(Equal("no-cache"))
	g.Expect(w.Header().Get("Abc")).To(Equal("1"))
	g.Expect(w.Header().Get("Def")).To(Equal("true"))
}

func TestValue_error(t *testing.T) {
	g := NewGomegaWithT(t)

	// Given ...
	d := Lazy(func(template, language string) (interface{}, string, error) {
		g.Expect(template).To(Equal("home.html"))
		g.Expect(language).To(Equal("en"))
		return nil, "", errors.New("expected error")
	})

	w := httptest.NewRecorder()

	// When ...
	_, err := GetContentAndApplyExtraHeaders(w, d, "home.html", "en")

	// Then ...
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(Equal("expected error"))
}
