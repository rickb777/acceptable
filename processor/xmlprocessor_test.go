package processor_test

import (
	"encoding/xml"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/rickb777/acceptable"
	"github.com/rickb777/acceptable/processor"

	. "github.com/onsi/gomega"
)

func TestXMLShouldWriteResponseBody(t *testing.T) {
	g := NewGomegaWithT(t)
	w := httptest.NewRecorder()

	model := &ValidXMLUser{
		"Joe Bloggs",
	}

	match := &acceptable.Match{
		Type:     "application",
		Subtype:  "json",
		Language: "en",
		Charset:  "utf-8",
	}

	p := processor.XML()

	p(w, match, "template", model)

	g.Expect(w.Header().Get("Content-Type")).To(Equal("application/json;charset=utf-8"))
	g.Expect(w.Header().Get("Content-Language")).To(Equal("en"))
	g.Expect(w.Body.String()).To(Equal("<ValidXMLUser><Name>Joe Bloggs</Name></ValidXMLUser>\n"))
}

func TestXMlShouldWriteResponseBodyWithIndentation(t *testing.T) {
	g := NewGomegaWithT(t)
	w := httptest.NewRecorder()

	model := &ValidXMLUser{Name: "Joe Bloggs"}
	match := &acceptable.Match{
		Type:     "application",
		Subtype:  "json",
		Language: "cn",
		Charset:  "utf-16",
	}

	p := processor.XML("  ")

	p(w, match, "template", model)

	g.Expect(w.Header().Get("Content-Type")).To(Equal("application/json;charset=utf-16"))
	g.Expect(w.Header().Get("Content-Language")).To(Equal("cn"))
	g.Expect(w.Body.String()).To(Equal("<ValidXMLUser>\n  <Name>Joe Bloggs</Name>\n</ValidXMLUser>\n"))
}

func TestXMLShouldRPanicOnError(t *testing.T) {
	g := NewGomegaWithT(t)
	w := httptest.NewRecorder()

	model := &XMLUser{Name: "Joe Bloggs"}
	match := &acceptable.Match{}

	p := processor.XML("  ")

	err := p(w, match, "template", model)

	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("oops"))
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

//func xmltestErrorHandler(w http.ResponseWriter, err error) {
//	w.WriteHeader(500)
//	w.Write([]byte(err.Error()))
//}
