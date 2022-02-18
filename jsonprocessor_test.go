package acceptable_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/rickb777/acceptable"
	datapkg "github.com/rickb777/acceptable/data"
	"github.com/rickb777/acceptable/header"
	. "github.com/rickb777/acceptable/header/headername"
	"github.com/rickb777/acceptable/offer"
)

func TestJSONShouldWriteResponseBody_lazy_indented(t *testing.T) {
	g := NewGomegaWithT(t)
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

	p := acceptable.JSON("  ")

	w := match.ApplyHeaders(rw)
	err := p(w, req, match.Data, "template", match.Language)

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(rw.Header().Get(ContentType)).To(Equal("application/json;charset=utf-8"))
	g.Expect(rw.Header().Get(ContentLanguage)).To(Equal("en"))
	g.Expect(rw.Body.String()).To(Equal("{\n  \"Name\": \"Joe Bloggs\"\n}\n"))
}

func TestJSONShouldWriteResponseBody_sequence(t *testing.T) {
	g := NewGomegaWithT(t)
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

	p := acceptable.JSON("  ")

	w := match.ApplyHeaders(rw)
	err := p(w, req, match.Data, "template", match.Language)

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(rw.Header().Get(ContentType)).To(Equal("application/json;charset=utf-8"))
	g.Expect(rw.Header().Get(ContentLanguage)).To(Equal("en"))
	g.Expect(rw.Body.String()).To(Equal(
		"[\n{\n    \"Name\": \"Ann Bollin\"\n  }\n,\n{\n    \"Name\": \"Joe Bloggs\"\n  }\n,\n{\n    \"Name\": \"Jane Hays\"\n  }\n\n]\n",
	))
}

func TestJSONShouldWriteResponseBodyIndented_utf16le(t *testing.T) {
	g := NewGomegaWithT(t)
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

		p := acceptable.JSON("")
		rw := httptest.NewRecorder()
		w := match.ApplyHeaders(rw)

		err := p(w, req, match.Data, "template", "cn")

		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(rw.Header().Get(ContentType)).To(Equal("application/json;charset=utf-16le"))
		g.Expect(rw.Header().Get(ContentLanguage)).To(Equal("cn"))
		g.Expect(rw.Body.Bytes()).To(Equal([]byte{
			'{', 0, '"', 0, 'N', 0, 'a', 0, 'm', 0, 'e', 0, '"', 0,
			':', 0, '"', 0, 13, 84, 240, 121, '"', 0, '}', 0, '\n', 0}))
	}
}

func TestJSONShouldReturnError(t *testing.T) {
	g := NewGomegaWithT(t)
	req := &http.Request{}
	w := httptest.NewRecorder()

	model := &User{"Joe Bloggs"}

	p := acceptable.JSON()

	err := p(w, req, datapkg.Of(model), "template", "en")

	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("error calling MarshalJSON for type"))
	g.Expect(err.Error()).To(ContainSubstring("oops"))
}

type User struct {
	Name string
}

func (u *User) MarshalJSON() ([]byte, error) {
	return nil, errors.New("oops")
}
