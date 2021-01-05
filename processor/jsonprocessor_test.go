package processor_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/rickb777/negotiator/processor"
)

func TestJSONShouldProcessAcceptHeader(t *testing.T) {
	g := NewGomegaWithT(t)
	var acceptTests = []struct {
		acceptheader string
		expected     bool
	}{
		{"application/json", true},
		{"application/json-", true},
		{"application/CEA", false},
		{"+json", true},
	}

	p := processor.JSON()

	for _, tt := range acceptTests {
		result := p.CanProcess(tt.acceptheader, "")

		g.Expect(result).To(Equal(tt.expected), "Should process "+tt.acceptheader)
	}
}

func TestJSONShouldSetContentTypeHeader(t *testing.T) {
	g := NewGomegaWithT(t)

	p := processor.JSON().(processor.ContentTypeSettable).WithContentType("application/foo")

	g.Expect(p.ContentType()).To(Equal("application/foo"))
}

func TestJSONShouldWriteResponseBody(t *testing.T) {
	g := NewGomegaWithT(t)
	recorder := httptest.NewRecorder()

	model := struct {
		Name string
	}{
		"Joe Bloggs",
	}

	p := processor.JSON()

	p.Process(recorder, "", model)

	g.Expect(recorder.Body.String()).To(Equal("{\"Name\":\"Joe Bloggs\"}\n"))
}

func TestJSONShouldWriteResponseBodyIndented(t *testing.T) {
	g := NewGomegaWithT(t)
	recorder := httptest.NewRecorder()

	model := struct {
		Name string
	}{
		"Joe Bloggs",
	}

	p := processor.JSON("  ")

	p.Process(recorder, "", model)

	g.Expect(recorder.Body.String()).To(Equal("{\n  \"Name\": \"Joe Bloggs\"\n}\n"))
}

func TestJSONShouldReturnError(t *testing.T) {
	g := NewGomegaWithT(t)
	recorder := httptest.NewRecorder()

	model := &User{
		"Joe Bloggs",
	}

	p := processor.JSON()

	err := p.Process(recorder, "", model)

	g.Expect(err).To(HaveOccurred())
}

type User struct {
	Name string
}

func (u *User) MarshalJSON() ([]byte, error) {
	return nil, errors.New("oops")
}

func jsontestErrorHandler(w http.ResponseWriter, err error) {
	w.WriteHeader(500)
	w.Write([]byte(err.Error()))
}
