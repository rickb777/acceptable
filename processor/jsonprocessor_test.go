package processor_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rickb777/acceptable"
	"github.com/rickb777/acceptable/data"
	"github.com/rickb777/acceptable/processor"

	. "github.com/onsi/gomega"
)

func TestJSONShouldWriteResponseBody(t *testing.T) {
	g := NewGomegaWithT(t)
	req := &http.Request{}
	w := httptest.NewRecorder()

	model := struct {
		Name string
	}{
		"Joe Bloggs",
	}

	match := acceptable.Match{
		Type:     "application",
		Subtype:  "json",
		Language: "en",
		Charset:  "utf-8",
		Data:     data.Lazy(func(string, string, bool) (interface{}, *data.Metadata, error) { return model, nil, nil }),
	}

	p := processor.JSON()

	p(w, req, match, "template")

	g.Expect(w.Header().Get("Content-Type")).To(Equal("application/json;charset=utf-8"))
	g.Expect(w.Header().Get("Content-Language")).To(Equal("en"))
	g.Expect(w.Body.String()).To(Equal("{\"Name\":\"Joe Bloggs\"}\n"))
}

func TestJSONShouldWriteResponseBodyIndented_utf16le(t *testing.T) {
	g := NewGomegaWithT(t)
	req := &http.Request{}

	model := struct {
		Name string
	}{
		"名称",
	}

	cases := []string{"utf-16le", "utf-16"} // unsupported "unicode"

	for _, enc := range cases {
		match := acceptable.Match{
			Type:     "application",
			Subtype:  "json",
			Language: "cn",
			Charset:  enc,
			Data:     data.Of(model),
		}

		p := processor.JSON("")
		w := httptest.NewRecorder()

		p(w, req, match, "template")

		g.Expect(w.Header().Get("Content-Type")).To(Equal("application/json;charset=utf-16le"))
		g.Expect(w.Header().Get("Content-Language")).To(Equal("cn"))
		g.Expect(w.Body.Bytes()).To(Equal([]byte{
			'{', 0, '"', 0, 'N', 0, 'a', 0, 'm', 0, 'e', 0, '"', 0,
			':', 0, '"', 0, 13, 84, 240, 121, '"', 0, '}', 0, '\n', 0}))
	}
}

func TestJSONShouldReturnError(t *testing.T) {
	g := NewGomegaWithT(t)
	req := &http.Request{}
	w := httptest.NewRecorder()

	model := &User{"Joe Bloggs"}
	match := acceptable.Match{Data: data.Of(model)}

	p := processor.JSON()

	err := p(w, req, match, "template")

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
