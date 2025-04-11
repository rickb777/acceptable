package offer_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	datapkg "github.com/rickb777/acceptable/data"
	"github.com/rickb777/acceptable/header"
	. "github.com/rickb777/acceptable/headername"
	"github.com/rickb777/acceptable/offer"
	"github.com/rickb777/expect"
)

func TestJSONShouldWriteResponseBody_lazy_indented(t *testing.T) {
	req := &http.Request{}
	rw := httptest.NewRecorder()

	model := struct {
		Name string
	}{
		"Joe Bloggs",
	}

	match := offer.Match{
		ContentType: header.ContentType{Type: "application", Subtype: "json"},
		Language:    "en",
		Charset:     "utf-8",
		Data:        datapkg.Lazy(func(string, string) (interface{}, error) { return model, nil }),
	}

	p := offer.JSONProcessor("  ")

	w := match.ApplyHeaders(rw)
	err := p(w, req, match.Data, "template", match.Language)

	expect.Error(err).Not().ToHaveOccurred(t)
	expect.String(rw.Header().Get(ContentType)).ToBe(t, "application/json;charset=utf-8")
	expect.String(rw.Header().Get(ContentLanguage)).ToBe(t, "en")
	expect.String(rw.Body.String()).ToBe(t, "{\n  \"Name\": \"Joe Bloggs\"\n}\n")
}

func TestJSONShouldWriteResponseBody_sequence(t *testing.T) {
	req := &http.Request{}
	rw := httptest.NewRecorder()

	model := []interface{}{User{Name: "Ann Bollin"}, User{Name: "Joe Bloggs"}, User{Name: "Jane Hays"}}

	match := offer.Match{
		ContentType: header.ContentType{Type: "application", Subtype: "json"},
		Language:    "en",
		Charset:     "utf-8",
		Data: datapkg.Sequence(func(string, string) (interface{}, error) {
			if len(model) == 0 {
				return nil, nil
			}
			m := model[0]
			model = model[1:]
			return m, nil
		}),
	}

	p := offer.JSONProcessor("  ")

	w := match.ApplyHeaders(rw)
	err := p(w, req, match.Data, "template", match.Language)

	expect.Error(err).Not().ToHaveOccurred(t)
	expect.String(rw.Header().Get(ContentType)).ToBe(t, "application/json;charset=utf-8")
	expect.String(rw.Header().Get(ContentLanguage)).ToBe(t, "en")
	expect.String(rw.Body.String()).ToBe(t,
		"[\n{\n    \"Name\": \"Ann Bollin\"\n  }\n,\n{\n    \"Name\": \"Joe Bloggs\"\n  }\n,\n{\n    \"Name\": \"Jane Hays\"\n  }\n\n]\n",
	)
}

func TestJSONShouldWriteResponseBodyIndented_utf16le(t *testing.T) {
	req := &http.Request{}

	model := struct {
		Name string
	}{
		"名称", // "name"
	}

	cases := []string{"utf-16le", "utf-16"} // unsupported "unicode"

	for _, enc := range cases {
		match := offer.Match{
			ContentType: header.ContentType{Type: "application", Subtype: "json"},
			Language:    "cn",
			Charset:     enc,
			Data:        datapkg.Of(model),
		}

		p := offer.JSONProcessor("")
		rw := httptest.NewRecorder()
		w := match.ApplyHeaders(rw)

		err := p(w, req, match.Data, "template", "cn")

		expect.Error(err).Not().ToHaveOccurred(t)
		expect.String(rw.Header().Get(ContentType)).ToBe(t, "application/json;charset=utf-16le")
		expect.String(rw.Header().Get(ContentLanguage)).ToBe(t, "cn")
		expect.String(rw.Body.Bytes()).ToBe(t, []byte{
			'{', 0, '"', 0, 'N', 0, 'a', 0, 'm', 0, 'e', 0, '"', 0,
			':', 0, '"', 0, 13, 84, 240, 121, '"', 0, '}', 0, '\n', 0})
	}
}

func TestJSONShouldReturnError(t *testing.T) {
	req := &http.Request{}
	w := httptest.NewRecorder()

	model := &User{"Joe Bloggs"}

	p := offer.JSONProcessor()

	err := p(w, req, datapkg.Of(model), "template", "en")

	expect.Error(err).ToHaveOccurred(t)
	expect.Error(err).ToContain(t, "error calling MarshalJSON for type")
	expect.Error(err).ToContain(t, "oops")
}

type User struct {
	Name string
}

func (u *User) MarshalJSON() ([]byte, error) {
	return nil, errors.New("oops")
}
