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

	match := acceptable.Match{
		Type:     "application",
		Subtype:  "json",
		Language: "en",
		Charset:  "utf-8",
		Data:     model,
	}

	p := processor.XML()

	p(w, match, "template")

	g.Expect(w.Header().Get("Content-Type")).To(Equal("application/json;charset=utf-8"))
	g.Expect(w.Header().Get("Content-Language")).To(Equal("en"))
	g.Expect(w.Body.String()).To(Equal("<ValidXMLUser><Name>Joe Bloggs</Name></ValidXMLUser>\n"))
}

func TestXMlShouldWriteResponseBodyWithIndentation(t *testing.T) {
	g := NewGomegaWithT(t)
	w := httptest.NewRecorder()

	model := &ValidXMLUser{Name: "名称"}
	match := acceptable.Match{
		Type:     "application",
		Subtype:  "json",
		Language: "cn",
		Charset:  "utf-16be",
		Data:     model,
	}

	p := processor.XML("  ")

	p(w, match, "template")

	g.Expect(w.Header().Get("Content-Type")).To(Equal("application/json;charset=utf-16be"))
	g.Expect(w.Header().Get("Content-Language")).To(Equal("cn"))
	g.Expect(w.Body.Bytes()).To(Equal([]byte{
		0, '<', 0, 'V', 0, 'a', 0, 'l', 0, 'i', 0, 'd', 0, 'X', 0, 'M', 0, 'L', 0, 'U', 0, 's', 0, 'e', 0, 'r', 0, '>', 0, '\n',
		0, ' ', 0, ' ', 0, '<', 0, 'N', 0, 'a', 0, 'm', 0, 'e', 0, '>', 84, 13, 121, 240,
		0, '<', 0, '/', 0, 'N', 0, 'a', 0, 'm', 0, 'e', 0, '>', 0, '\n', 0,
		'<', 0, '/', 0, 'V', 0, 'a', 0, 'l', 0, 'i', 0, 'd', 0, 'X', 0, 'M', 0, 'L', 0, 'U', 0, 's', 0, 'e', 0, 'r', 0, '>', 0, '\n',
	}), w.Body.String())
}

func TestXMLShouldRPanicOnError(t *testing.T) {
	g := NewGomegaWithT(t)
	w := httptest.NewRecorder()

	model := &XMLUser{Name: "Joe Bloggs"}
	match := acceptable.Match{Data: model}

	p := processor.XML("  ")

	err := p(w, match, "template")

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
