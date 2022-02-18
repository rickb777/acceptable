package data

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	. "github.com/rickb777/acceptable/headername"
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

		d := Lazy(func(template, language string) (interface{}, error) {
			// Then ...
			count++
			g.Expect(template).To(Equal(expectedTemplate))
			g.Expect(language).To(Equal(expectedLanguage))
			return "foo", nil
		})

		req, _ := http.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		// When ...
		send, e1 := ConditionalRequest(w, req, d, expectedTemplate, expectedLanguage)
		c, more, e2 := d.Content(expectedTemplate, expectedLanguage)

		// Then ...
		g.Expect(e1).NotTo(HaveOccurred())
		g.Expect(send).To(BeTrue())

		g.Expect(e2).NotTo(HaveOccurred())
		g.Expect(more).To(BeFalse())
		g.Expect(c).To(Equal("foo"))
		g.Expect(count).To(Equal(1))
	}
}

func TestLazyValue_attaching_eager_metadata(t *testing.T) {
	g := NewGomegaWithT(t)

	// Given ...
	d := Lazy(func(template, language string) (interface{}, error) {
		return "foo", nil
	}).
		ETag("abcdef").
		LastModified(t1)

	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	// When ...
	send, e1 := ConditionalRequest(w, req, d, "home.html", "en")
	_, _, e2 := d.Content("home.html", "en")

	// Then ...
	g.Expect(e1).NotTo(HaveOccurred())
	g.Expect(send).To(BeTrue())

	g.Expect(e2).NotTo(HaveOccurred())
	g.Expect(w.Header()).To(HaveLen(2))
	g.Expect(w.Header().Get(LastModified)).To(Equal("Wed, 01 Jan 2020 01:01:01 UTC"))
	g.Expect(w.Header().Get(ETag)).To(Equal(`"abcdef"`))
}

func TestLazyValue_attaching_lazy_metadata(t *testing.T) {
	g := NewGomegaWithT(t)

	// Given ...
	d := Lazy(func(template, language string) (interface{}, error) {
		return "foo", nil
	}).
		ETagUsing(func(template, language string) (string, error) {
			return "abcdef", nil
		}).
		LastModifiedUsing(func(template, language string) (time.Time, error) {
			return t1, nil
		})

	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	// When ...
	send, e1 := ConditionalRequest(w, req, d, "home.html", "en")
	_, _, e2 := d.Content("home.html", "en")

	// Then ...
	g.Expect(e1).NotTo(HaveOccurred())
	g.Expect(send).To(BeTrue())

	g.Expect(e2).NotTo(HaveOccurred())
	g.Expect(w.Header()).To(HaveLen(2))
	g.Expect(w.Header().Get(LastModified)).To(Equal("Wed, 01 Jan 2020 01:01:01 UTC"))
	g.Expect(w.Header().Get(ETag)).To(Equal(`"abcdef"`))
}

func TestLazyValue_returning_error(t *testing.T) {
	g := NewGomegaWithT(t)

	for i := 1; i <= 2; i++ {
		// Given ...
		d := Lazy(func(template, language string) (interface{}, error) {
			return nil, errors.New("expected error")
		})

		req := &http.Request{}
		w := httptest.NewRecorder()

		// When ...
		_, e1 := ConditionalRequest(w, req, d, "home.html", "en")
		_, _, e2 := d.Content("home.html", "en")

		// Then ...
		g.Expect(e1).NotTo(HaveOccurred())
		g.Expect(e2).To(HaveOccurred())
		g.Expect(e2.Error()).To(Equal("expected error"))
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
	send, e1 := ConditionalRequest(w, req, d, "home.html", "en")
	_, _, e2 := d.Content("home.html", "en")

	// Then ...
	g.Expect(e1).NotTo(HaveOccurred())
	g.Expect(send).To(BeTrue())

	g.Expect(e2).NotTo(HaveOccurred())
	g.Expect(w.Header()).To(HaveLen(4))
	g.Expect(w.Header().Get(CacheControl)).To(Equal("max-age=10"))
	g.Expect(w.Header().Get(Expires)).To(Equal("Thu, 02 Jan 2020 03:04:05 UTC"))
	g.Expect(w.Header().Get(LastModified)).To(Equal("Wed, 01 Jan 2020 01:01:01 UTC"))
	g.Expect(w.Header().Get(ETag)).To(Equal(`"abcdef"`))
}

func TestValue_no_cache_and_additional_headers(t *testing.T) {
	g := NewGomegaWithT(t)

	// Given ...
	d := Of("foo").NoCache().With("Abc", "1", "Def", "true")

	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	// When ...
	send, e1 := ConditionalRequest(w, req, d, "home.html", "en")
	_, _, e2 := d.Content("home.html", "en")

	// Then ...
	g.Expect(e1).NotTo(HaveOccurred())
	g.Expect(send).To(BeTrue())

	g.Expect(e2).NotTo(HaveOccurred())
	g.Expect(w.Header()).To(HaveLen(4))
	g.Expect(w.Header().Get(CacheControl)).To(Equal("no-cache, must-revalidate"))
	g.Expect(w.Header().Get(Pragma)).To(Equal("no-cache"))
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
		req.Header.Set(IfNoneMatch, `"foo", "hash123", "bar"`)
		w := httptest.NewRecorder()

		// When ...
		send, e1 := ConditionalRequest(w, req, d, "home.html", "en")
		_, _, e2 := d.Content("home.html", "en")

		// Then ...
		g.Expect(e1).NotTo(HaveOccurred())
		g.Expect(send).To(BeFalse())

		g.Expect(e2).NotTo(HaveOccurred())
		g.Expect(w.Code).To(Equal(304))
		g.Expect(w.Header()).To(HaveLen(5))
		g.Expect(w.Header().Get(ETag)).To(Equal(`"hash123"`))
		g.Expect(w.Header().Get(LastModified)).To(Equal(`Thu, 02 Jan 2020 03:04:05 UTC`))
		g.Expect(w.Header().Get(CacheControl)).To(Equal("max-age=10"))
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
		req.Header.Set(IfModifiedSince, `Wed, 01 Jan 2020 00:00:00 UTC`)
		w := httptest.NewRecorder()

		// When ...
		send, e1 := ConditionalRequest(w, req, d, "home.html", "en")
		_, _, e2 := d.Content("home.html", "en")

		// Then ...
		g.Expect(e1).NotTo(HaveOccurred())
		g.Expect(send).To(BeFalse())

		g.Expect(e2).NotTo(HaveOccurred())
		g.Expect(w.Code).To(Equal(304))
		g.Expect(w.Header()).To(HaveLen(5))
		g.Expect(w.Header().Get(ETag)).To(Equal(`"hash123"`))
		g.Expect(w.Header().Get(LastModified)).To(Equal(`Thu, 02 Jan 2020 03:04:05 UTC`))
		g.Expect(w.Header().Get(CacheControl)).To(Equal("max-age=10"))
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
		req.Header.Set(IfNoneMatch, `"foo", "hash123", "bar"`)
		w := httptest.NewRecorder()

		// When ...
		send, e1 := ConditionalRequest(w, req, d, "home.html", "en")
		_, _, e2 := d.Content("home.html", "en")

		// Then ...
		g.Expect(e1).NotTo(HaveOccurred())
		g.Expect(send).To(BeTrue())

		g.Expect(e2).NotTo(HaveOccurred())
		g.Expect(w.Code).To(Equal(200))
		g.Expect(w.Header()).To(HaveLen(4))
		g.Expect(w.Header().Get(CacheControl)).To(Equal("no-cache, must-revalidate"))
		g.Expect(w.Header().Get(Pragma)).To(Equal("no-cache"))
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
	send, err := ConditionalRequest(w, req, nil, "home.html", "en")

	// Then ...
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(send).To(BeFalse())
}

func conditional(predicate bool, value interface{}) interface{} {
	if predicate {
		return value
	}
	return nil
}
