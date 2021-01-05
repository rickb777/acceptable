package processor_test

import (
	"encoding/xml"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/rickb777/negotiator/processor"
)

func TestXMLShouldProcessAcceptHeader(t *testing.T) {
	g := NewGomegaWithT(t)
	var acceptTests = []struct {
		acceptheader string
		expected     bool
	}{
		{"application/xml", true},
		{"application/xml-dtd", true},
		{"application/CEA", false},
		{"image/svg+xml", true},
	}

	p := processor.XML()

	for _, tt := range acceptTests {
		result := p.CanProcess(tt.acceptheader, "")
		g.Expect(result).To(Equal(tt.expected), "Should process "+tt.acceptheader)
	}
}

func TestXMLShouldSetContentTypeHeader(t *testing.T) {
	g := NewGomegaWithT(t)

	p := processor.XML().(processor.ContentTypeSettable).WithContentType("application/my+xml")

	g.Expect(p.ContentType()).To(Equal("application/my+xml"))
}

func TestXMLShouldSetResponseBody(t *testing.T) {
	g := NewGomegaWithT(t)
	recorder := httptest.NewRecorder()

	model := &ValidXMLUser{
		"Joe Bloggs",
	}

	p := processor.XML()

	p.Process(recorder, "", model)

	g.Expect(recorder.Body.String()).To(Equal("<ValidXMLUser><Name>Joe Bloggs</Name></ValidXMLUser>"))
}

func TestXMlShouldSetResponseBodyWithIndentation(t *testing.T) {
	g := NewGomegaWithT(t)
	recorder := httptest.NewRecorder()

	model := &ValidXMLUser{Name: "Joe Bloggs"}

	p := processor.IndentedXML("  ")

	p.Process(recorder, "", model)

	g.Expect(recorder.Body.String()).To(Equal("<ValidXMLUser>\n  <Name>Joe Bloggs</Name>\n</ValidXMLUser>\n"))
}

func TestXMLShouldRPanicOnError(t *testing.T) {
	g := NewGomegaWithT(t)
	recorder := httptest.NewRecorder()

	model := &XMLUser{Name: "Joe Bloggs"}

	p := processor.IndentedXML("  ")

	err := p.Process(recorder, "", model)

	g.Expect(err).To(HaveOccurred())
}

type ValidXMLUser struct {
	Name string
}

type XMLUser struct {
	Name string
}

func (u *XMLUser) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return errors.New("oops")
}

func xmltestErrorHandler(w http.ResponseWriter, err error) {
	w.WriteHeader(500)
	w.Write([]byte(err.Error()))
}
