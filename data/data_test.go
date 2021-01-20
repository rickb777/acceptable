package data

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/rickb777/acceptable/header"
)

var t1 = time.Date(2020, 1, 1, 1, 1, 1, 0, time.UTC)
var t2 = time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)

func TestLazyValue_should_pass_template_and_language(t *testing.T) {
	g := NewGomegaWithT(t)

	for i := 1; i <= 2; i++ {
		// Given ...
		count := 0
		expectedTemplate := fmt.Sprintf("p%d.html", i)
		expectedLanguage := fmt.Sprintf("en-x%d", i)
		d := Lazy(func(template, language string, dr bool) (interface{}, *Metadata, error) {
			// Then ...
			count++
			g.Expect(template).To(Equal(expectedTemplate))
			g.Expect(language).To(Equal(expectedLanguage))
			return conditional(dr, "foo"), &Metadata{Hash: "abcdef", LastModified: t1}, nil
		})

		req, _ := http.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		// When ...
		c, err := GetContentAndApplyExtraHeaders(w, req, d, expectedTemplate, expectedLanguage)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(c).To(Equal("foo"))
		g.Expect(count).To(Equal(2))
	}
}

func TestLazyValue_returning_metadata(t *testing.T) {
	g := NewGomegaWithT(t)

	for i := 1; i <= 2; i++ {
		// Given ...
		count := 0
		d := Lazy(func(template, language string, dr bool) (interface{}, *Metadata, error) {
			count++
			if i == 1 {
				return "foo", &Metadata{Hash: "abcdef", LastModified: t1}, nil
			}
			return conditional(dr, "foo"), &Metadata{Hash: "abcdef", LastModified: t1}, nil
		})

		req, _ := http.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		// When ...
		c, err := GetContentAndApplyExtraHeaders(w, req, d, "home.html", "en")

		// Then ...
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(c).To(Equal("foo"))
		g.Expect(w.Header()).To(HaveLen(2))
		g.Expect(w.Header().Get("Last-Modified")).To(Equal("Wed, 01 Jan 2020 01:01:01 UTC"))
		g.Expect(w.Header().Get("ETag")).To(Equal(`"abcdef"`))
		g.Expect(count).To(Equal(i))
	}
}

func TestLazyValue_attaching_metadata(t *testing.T) {
	g := NewGomegaWithT(t)

	for i := 1; i <= 2; i++ {
		// Given ...
		count := 0
		d := Lazy(func(template, language string, dr bool) (interface{}, *Metadata, error) {
			count++
			if i == 1 {
				return "foo", nil, nil
			}
			return conditional(dr, "foo"), nil, nil
		}).
			ETag("abcdef").
			LastModified(t1)

		req, _ := http.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		// When ...
		c, err := GetContentAndApplyExtraHeaders(w, req, d, "home.html", "en")

		// Then ...
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(c).To(Equal("foo"))
		g.Expect(w.Header()).To(HaveLen(2))
		g.Expect(w.Header().Get("Last-Modified")).To(Equal("Wed, 01 Jan 2020 01:01:01 UTC"))
		g.Expect(w.Header().Get("ETag")).To(Equal(`"abcdef"`))
		g.Expect(count).To(Equal(i))
	}
}

func TestLazyValue_merging_metadata(t *testing.T) {
	g := NewGomegaWithT(t)

	for i := 1; i <= 2; i++ {
		// Given ...
		count := 0
		d := Lazy(func(template, language string, dr bool) (interface{}, *Metadata, error) {
			count++
			if i == 1 {
				return "foo", &Metadata{}, nil
			}
			return conditional(dr, "foo"), &Metadata{}, nil
		}).
			ETag("abcdef").
			LastModified(t1)

		req, _ := http.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		// When ...
		c, err := GetContentAndApplyExtraHeaders(w, req, d, "home.html", "en")

		// Then ...
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(c).To(Equal("foo"))
		g.Expect(w.Header()).To(HaveLen(2))
		g.Expect(w.Header().Get("Last-Modified")).To(Equal("Wed, 01 Jan 2020 01:01:01 UTC"))
		g.Expect(w.Header().Get("ETag")).To(Equal(`"abcdef"`))
		g.Expect(count).To(Equal(i))
	}
}

func TestLazyValue_returning_error(t *testing.T) {
	g := NewGomegaWithT(t)

	for i := 1; i <= 2; i++ {
		// Given ...
		count := 0
		d := Lazy(func(template, language string, dr bool) (interface{}, *Metadata, error) {
			count++
			e := errors.New("expected error")
			if i == 2 && !dr {
				return nil, nil, nil
			}
			return nil, nil, e
		})

		req := &http.Request{}
		w := httptest.NewRecorder()

		// When ...
		_, err := GetContentAndApplyExtraHeaders(w, req, d, "home.html", "en")

		// Then ...
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("expected error"))
		g.Expect(count).To(Equal(i))
	}
}

//-------------------------------------------------------------------------------------------------

func TestValue_future_expiry(t *testing.T) {
	g := NewGomegaWithT(t)

	// Given ...
	d := Of("foo").
		ETag("abcdef").
		Expires(t2).
		LastModified(t1).
		MaxAge(10 * time.Second)

	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	// When ...
	c, err := GetContentAndApplyExtraHeaders(w, req, d, "home.html", "en")

	// Then ...
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(c).To(Equal("foo"))
	g.Expect(w.Header()).To(HaveLen(4))
	g.Expect(w.Header().Get("Cache-Control")).To(Equal("max-age=10"))
	g.Expect(w.Header().Get("Expires")).To(Equal("Thu, 02 Jan 2020 03:04:05 UTC"))
	g.Expect(w.Header().Get("Last-Modified")).To(Equal("Wed, 01 Jan 2020 01:01:01 UTC"))
	g.Expect(w.Header().Get("ETag")).To(Equal(`"abcdef"`))
}

func TestValue_no_cache_and_additional_headers(t *testing.T) {
	g := NewGomegaWithT(t)

	// Given ...
	d := Of("foo").NoCache().With("Abc", "1", "Def", "true")

	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	// When ...
	c, err := GetContentAndApplyExtraHeaders(w, req, d, "home.html", "en")

	// Then ...
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(c).To(Equal("foo"))
	g.Expect(w.Header()).To(HaveLen(4))
	g.Expect(w.Header().Get("Cache-Control")).To(Equal("no-cache, must-revalidate"))
	g.Expect(w.Header().Get("Pragma")).To(Equal("no-cache"))
	g.Expect(w.Header().Get("Abc")).To(Equal("1"))
	g.Expect(w.Header().Get("Def")).To(Equal("true"))
}

func TestValue_if_none_match_not_modified_get_request(t *testing.T) {
	g := NewGomegaWithT(t)

	// Given ...
	d := Of("foo").With("Abc", "1", "Def", "true").MaxAge(10 * time.Second).
		ETag("hash123").
		LastModified(time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC))

	for _, method := range []string{"GET", "HEAD"} {
		req, _ := http.NewRequest(method, "/", nil)
		req.Header.Set(header.IfNoneMatch, `"foo", "hash123", "bar"`)
		w := httptest.NewRecorder()

		// When ...
		c, err := GetContentAndApplyExtraHeaders(w, req, d, "home.html", "en")

		// Then ...
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(c).To(BeNil())
		g.Expect(w.Code).To(Equal(304))
		g.Expect(w.Header()).To(HaveLen(5))
		g.Expect(w.Header().Get("ETag")).To(Equal(`"hash123"`))
		g.Expect(w.Header().Get("Last-Modified")).To(Equal(`Thu, 02 Jan 2020 03:04:05 UTC`))
		g.Expect(w.Header().Get("Cache-Control")).To(Equal("max-age=10"))
		g.Expect(w.Header().Get("Abc")).To(Equal("1"))
		g.Expect(w.Header().Get("Def")).To(Equal("true"))
	}
}

func TestValue_if_modified_since_not_modified_get_request(t *testing.T) {
	g := NewGomegaWithT(t)

	// Given ...
	d := Of("foo").ETag("hash123").With("Abc", "1", "Def", "true").MaxAge(10 * time.Second).
		LastModified(t2)

	for _, method := range []string{"GET", "HEAD"} {
		req, _ := http.NewRequest(method, "/", nil)
		req.Header.Set(header.IfModifiedSince, `Wed, 01 Jan 2020 00:00:00 UTC`)
		w := httptest.NewRecorder()

		// When ...
		c, err := GetContentAndApplyExtraHeaders(w, req, d, "home.html", "en")

		// Then ...
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(c).To(BeNil())
		g.Expect(w.Code).To(Equal(304))
		g.Expect(w.Header()).To(HaveLen(5))
		g.Expect(w.Header().Get("ETag")).To(Equal(`"hash123"`))
		g.Expect(w.Header().Get("Last-Modified")).To(Equal(`Thu, 02 Jan 2020 03:04:05 UTC`))
		g.Expect(w.Header().Get("Cache-Control")).To(Equal("max-age=10"))
		g.Expect(w.Header().Get("Abc")).To(Equal("1"))
		g.Expect(w.Header().Get("Def")).To(Equal("true"))
	}
}

func TestValue_not_modified_put_request(t *testing.T) {
	g := NewGomegaWithT(t)

	// Given ...
	d := Of("foo").ETag("hash123").NoCache().With("Abc", "1", "Def", "true")

	for _, method := range []string{"PUT", "POST", "DELETE"} {
		req, _ := http.NewRequest(method, "/", nil)
		req.Header.Set(header.IfNoneMatch, `"foo", "hash123", "bar"`)
		w := httptest.NewRecorder()

		// When ...
		c, err := GetContentAndApplyExtraHeaders(w, req, d, "home.html", "en")

		// Then ...
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(c).NotTo(BeNil())
		g.Expect(w.Code).To(Equal(200))
		g.Expect(w.Header()).To(HaveLen(4))
		g.Expect(w.Header().Get("Cache-Control")).To(Equal("no-cache, must-revalidate"))
		g.Expect(w.Header().Get("Pragma")).To(Equal("no-cache"))
		g.Expect(w.Header().Get("Abc")).To(Equal("1"))
		g.Expect(w.Header().Get("Def")).To(Equal("true"))
	}
}

func TestGetContentAndApplyExtraHeaders_nil_data(t *testing.T) {
	g := NewGomegaWithT(t)

	// Given ...
	req := &http.Request{}
	w := httptest.NewRecorder()

	// When ...
	d, err := GetContentAndApplyExtraHeaders(w, req, nil, "home.html", "en")

	// Then ...
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(d).To(BeNil())
}

func conditional(predicate bool, value interface{}) interface{} {
	if predicate {
		return value
	}
	return nil
}
