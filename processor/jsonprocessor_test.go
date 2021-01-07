package processor_test

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/rickb777/acceptable"
	"github.com/rickb777/acceptable/processor"

	. "github.com/onsi/gomega"
)

func TestJSONShouldWriteResponseBody(t *testing.T) {
	g := NewGomegaWithT(t)
	w := httptest.NewRecorder()

	model := struct {
		Name string
	}{
		"Joe Bloggs",
	}

	match := &acceptable.Match{
		Type:     "application",
		Subtype:  "json",
		Language: "en",
		Charset:  "utf-8",
	}

	p := processor.JSON()

	p(w, match, "template", model)

	g.Expect(w.Header().Get("Content-Type")).To(Equal("application/json;charset=utf-8"))
	g.Expect(w.Header().Get("Content-Language")).To(Equal("en"))
	g.Expect(w.Body.String()).To(Equal("{\"Name\":\"Joe Bloggs\"}\n"))
}

func TestJSONShouldWriteResponseBodyIndented(t *testing.T) {
	g := NewGomegaWithT(t)
	w := httptest.NewRecorder()

	model := struct {
		Name string
	}{
		"名称",
	}
	match := &acceptable.Match{
		Type:     "application",
		Subtype:  "json",
		Language: "cn",
		Charset:  "utf-16",
	}

	p := processor.JSON("  ")

	p(w, match, "template", model)

	g.Expect(w.Header().Get("Content-Type")).To(Equal("application/json;charset=utf-16"))
	g.Expect(w.Header().Get("Content-Language")).To(Equal("cn"))
	g.Expect(w.Body.String()).To(Equal("{\n  \"Name\": \"名称\"\n}\n"))
}

func TestJSONShouldReturnError(t *testing.T) {
	g := NewGomegaWithT(t)
	w := httptest.NewRecorder()

	model := &User{
		"Joe Bloggs",
	}
	match := &acceptable.Match{}

	p := processor.JSON()

	err := p(w, match, "template", model)

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

//func jsontestErrorHandler(w http.ResponseWriter, err error) {
//	w.WriteHeader(500)
//	w.Write([]byte(err.Error()))
//}
