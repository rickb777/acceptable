package data

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	. "github.com/rickb777/acceptable/headername"
	"github.com/rickb777/expect"
)

var t1 = time.Date(2020, 1, 1, 1, 1, 1, 0, time.UTC)
var t2 = time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)

func TestLazyValue_should_pass_template_and_language(t *testing.T) {
	for i := 1; i <= 2; i++ {
		// Given ...
		count := 0
		expectedTemplate := fmt.Sprintf("p%d.html", i)
		expectedLanguage := fmt.Sprintf("en-x%d", i)

		d := Lazy(func(template, language string) (interface{}, error) {
			// Then ...
			count++
			expect.String(template).ToBe(t, expectedTemplate)
			expect.String(language).ToBe(t, expectedLanguage)
			return "foo", nil
		})

		req, _ := http.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		// When ...
		send, e1 := ConditionalRequest(w, req, d, expectedTemplate, expectedLanguage)
		c, more, e2 := d.Content(expectedTemplate, expectedLanguage)

		// Then ...
		expect.Error(e1).Not().ToHaveOccurred(t)
		expect.Bool(send).ToBeTrue(t)

		expect.Error(e2).Not().ToHaveOccurred(t)
		expect.Bool(more).ToBeFalse(t)
		expect.Any(c).ToBe(t, "foo")
		expect.Number(count).ToBe(t, 1)
	}
}

func TestLazyValue_attaching_eager_metadata(t *testing.T) {
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
	expect.Error(e1).Not().ToHaveOccurred(t)
	expect.Bool(send).ToBeTrue(t)

	expect.Error(e2).Not().ToHaveOccurred(t)
	expect.Map(w.Header()).ToHaveLength(t, 2)
	expect.String(w.Header().Get(LastModified)).ToBe(t, "Wed, 01 Jan 2020 01:01:01 GMT")
	expect.String(w.Header().Get(ETag)).ToBe(t, `"abcdef"`)
}

func TestLazyValue_attaching_lazy_metadata(t *testing.T) {
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
	expect.Error(e1).Not().ToHaveOccurred(t)
	expect.Bool(send).ToBeTrue(t)

	expect.Error(e2).Not().ToHaveOccurred(t)
	expect.Map(w.Header()).ToHaveLength(t, 2)
	expect.String(w.Header().Get(LastModified)).ToBe(t, "Wed, 01 Jan 2020 01:01:01 GMT")
	expect.String(w.Header().Get(ETag)).ToBe(t, `"abcdef"`)
}

func TestLazyValue_returning_error(t *testing.T) {
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
		expect.Error(e1).Not().ToHaveOccurred(t)
		expect.Error(e2).ToHaveOccurred(t)
		expect.Error(e2).ToContain(t, "expected error")
	}
}

//-------------------------------------------------------------------------------------------------

func TestValue_future_expiry(t *testing.T) {
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
	expect.Error(e1).Not().ToHaveOccurred(t)
	expect.Bool(send).ToBeTrue(t)

	expect.Error(e2).Not().ToHaveOccurred(t)
	expect.Map(w.Header()).ToHaveLength(t, 4)
	expect.String(w.Header().Get(CacheControl)).ToBe(t, "max-age=10")
	expect.String(w.Header().Get(Expires)).ToBe(t, "Thu, 02 Jan 2020 03:04:05 GMT")
	expect.String(w.Header().Get(LastModified)).ToBe(t, "Wed, 01 Jan 2020 01:01:01 GMT")
	expect.String(w.Header().Get(ETag)).ToBe(t, `"abcdef"`)
}

func TestValue_no_cache_and_additional_headers(t *testing.T) {
	// Given ...
	d := Of("foo").NoCache().With("Abc", "1", "Def", "true")

	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	// When ...
	send, e1 := ConditionalRequest(w, req, d, "home.html", "en")
	_, _, e2 := d.Content("home.html", "en")

	// Then ...
	expect.Error(e1).Not().ToHaveOccurred(t)
	expect.Bool(send).ToBeTrue(t)

	expect.Error(e2).Not().ToHaveOccurred(t)
	expect.Map(w.Header()).ToHaveLength(t, 4)
	expect.String(w.Header().Get(CacheControl)).ToBe(t, "no-cache, must-revalidate")
	expect.String(w.Header().Get(Pragma)).ToBe(t, "no-cache")
	expect.String(w.Header().Get("Abc")).ToBe(t, "1")
	expect.String(w.Header().Get("Def")).ToBe(t, "true")
}

func TestValue_if_none_match_not_modified_get_request(t *testing.T) {
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
		expect.Error(e1).Not().ToHaveOccurred(t)
		expect.Bool(send).ToBeFalse(t)

		expect.Error(e2).Not().ToHaveOccurred(t)
		expect.Number(w.Code).ToBe(t, 304)
		expect.Map(w.Header()).ToHaveLength(t, 5)
		expect.String(w.Header().Get(ETag)).ToBe(t, `"hash123"`)
		expect.String(w.Header().Get(LastModified)).ToBe(t, `Thu, 02 Jan 2020 03:04:05 GMT`)
		expect.String(w.Header().Get(CacheControl)).ToBe(t, "max-age=10")
		expect.String(w.Header().Get("Abc")).ToBe(t, "1")
		expect.String(w.Header().Get("Def")).ToBe(t, "true")
	}
}

func TestValue_if_modified_since_not_modified_get_request(t *testing.T) {
	// Given ...
	d := Of("foo").ETag("hash123").With("Abc", "1", "Def", "true").MaxAge(10 * time.Second).
		LastModified(t2)

	for _, method := range []string{"GET", "HEAD"} {
		req, _ := http.NewRequest(method, "/", nil)
		req.Header.Set(IfModifiedSince, `Wed, 01 Jan 2020 00:00:00 GMT`)
		w := httptest.NewRecorder()

		// When ...
		send, e1 := ConditionalRequest(w, req, d, "home.html", "en")
		_, _, e2 := d.Content("home.html", "en")

		// Then ...
		expect.Error(e1).Not().ToHaveOccurred(t)
		expect.Bool(send).ToBeFalse(t)

		expect.Error(e2).Not().ToHaveOccurred(t)
		expect.Number(w.Code).ToBe(t, 304)
		expect.Map(w.Header()).ToHaveLength(t, 5)
		expect.String(w.Header().Get(ETag)).ToBe(t, `"hash123"`)
		expect.String(w.Header().Get(LastModified)).ToBe(t, `Thu, 02 Jan 2020 03:04:05 GMT`)
		expect.String(w.Header().Get(CacheControl)).ToBe(t, "max-age=10")
		expect.String(w.Header().Get("Abc")).ToBe(t, "1")
		expect.String(w.Header().Get("Def")).ToBe(t, "true")
	}
}

func TestValue_not_modified_put_request(t *testing.T) {
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
		expect.Error(e1).Not().ToHaveOccurred(t)
		expect.Bool(send).ToBeTrue(t)

		expect.Error(e2).Not().ToHaveOccurred(t)
		expect.Number(w.Code).ToBe(t, 200)
		expect.Map(w.Header()).ToHaveLength(t, 4)
		expect.String(w.Header().Get(CacheControl)).ToBe(t, "no-cache, must-revalidate")
		expect.String(w.Header().Get(Pragma)).ToBe(t, "no-cache")
		expect.String(w.Header().Get("Abc")).ToBe(t, "1")
		expect.String(w.Header().Get("Def")).ToBe(t, "true")
	}
}
