package offer_test

import (
	"encoding/xml"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/rickb777/acceptable/data"
	"github.com/rickb777/acceptable/header"
	"github.com/rickb777/acceptable/headername"
	"github.com/rickb777/acceptable/offer"
)

func TestXMLShouldWriteLazyResponseBody(t *testing.T) {
	g := NewGomegaWithT(t)
	req := &http.Request{}
	rw := httptest.NewRecorder()

	model := &ValidXMLUser{
		"Joe Bloggs",
	}

	match := offer.Match{
		ContentType: header.ContentType{Type: "application", Subtype: "json"},
		Language:    "en",
		Charset:     "utf-8",
		Data:        data.Lazy(func(string, string) (interface{}, error) { return model, nil }),
	}

	p := offer.XMLProcessor("xml")

	w := match.ApplyHeaders(rw)
	err := p(w, req, match.Data, "template", match.Language)

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(rw.Header().Get(headername.ContentType)).To(Equal("application/json;charset=utf-8"))
	g.Expect(rw.Header().Get(headername.ContentLanguage)).To(Equal("en"))
	g.Expect(rw.Body.String()).To(Equal("<ValidXMLUser><Name>Joe Bloggs</Name></ValidXMLUser>\n"))
}

func TestXMLShouldWriteSequenceResponseBody(t *testing.T) {
	g := NewGomegaWithT(t)
	req := &http.Request{}
	rw := httptest.NewRecorder()

	model := []interface{}{
		&ValidXMLUser{
			"Ann Bollin",
		},
		&ValidXMLUser{
			"Bob Peel",
		},
		&ValidXMLUser{
			"Charles Dickens",
		},
	}

	match := offer.Match{
		ContentType: header.ContentType{Type: "application", Subtype: "json"},
		Language:    "en",
		Charset:     "utf-8",
		Data: data.Sequence(func(string, string) (interface{}, error) {
			if len(model) == 0 {
				return nil, nil
			}
			m := model[0]
			model = model[1:]
			return m, nil
		}),
	}

	p := offer.XMLProcessor("xml", "  ")

	w := match.ApplyHeaders(rw)
	err := p(w, req, match.Data, "template", match.Language)

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(rw.Header().Get(headername.ContentType)).To(Equal("application/json;charset=utf-8"))
	g.Expect(rw.Header().Get(headername.ContentLanguage)).To(Equal("en"))
	g.Expect(rw.Body.String()).To(Equal(
		"<xml>\n"+
			"  <ValidXMLUser>\n"+
			"    <Name>Ann Bollin</Name>\n"+
			"  </ValidXMLUser>\n\n"+
			"  <ValidXMLUser>\n"+
			"    <Name>Bob Peel</Name>\n"+
			"  </ValidXMLUser>\n\n"+
			"  <ValidXMLUser>\n"+
			"    <Name>Charles Dickens</Name>\n"+
			"  </ValidXMLUser>\n"+
			"</xml>\n"), rw.Body.String())
}

func TestXMlShouldWriteResponseBodyWithIndentation_utf_16be(t *testing.T) {
	g := NewGomegaWithT(t)
	req := &http.Request{}

	model := &ValidXMLUser{Name: "名称"}
	cases := []string{"utf-16be"} // unsupported: "unicodefffe"

	for _, enc := range cases {
		match := offer.Match{
			ContentType: header.ContentType{Type: "application", Subtype: "json"},
			Language:    "cn",
			Charset:     enc,
			Data:        data.Of(model),
		}

		p := offer.XMLProcessor("xml", "  ")
		rw := httptest.NewRecorder()

		w := match.ApplyHeaders(rw)
		err := p(w, req, match.Data, "template", match.Language)

		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(rw.Header().Get(headername.ContentType)).To(Equal("application/json;charset=utf-16be"), enc)
		g.Expect(rw.Header().Get(headername.ContentLanguage)).To(Equal("cn"), enc)
		g.Expect(rw.Body.Bytes()).To(Equal([]byte{
			0, '<', 0, 'V', 0, 'a', 0, 'l', 0, 'i', 0, 'd', 0, 'X', 0, 'M', 0, 'L', 0, 'U', 0, 's', 0, 'e', 0, 'r', 0, '>', 0, '\n',
			0, ' ', 0, ' ', 0, '<', 0, 'N', 0, 'a', 0, 'm', 0, 'e', 0, '>', 84, 13, 121, 240,
			0, '<', 0, '/', 0, 'N', 0, 'a', 0, 'm', 0, 'e', 0, '>', 0, '\n', 0,
			'<', 0, '/', 0, 'V', 0, 'a', 0, 'l', 0, 'i', 0, 'd', 0, 'X', 0, 'M', 0, 'L', 0, 'U', 0, 's', 0, 'e', 0, 'r', 0, '>', 0, '\n',
		}), rw.Body.String(), enc)
	}
}

func TestXMlShouldWriteResponseBodyWithIndentation_utf_16le(t *testing.T) {
	g := NewGomegaWithT(t)
	req := &http.Request{}

	model := &ValidXMLUser{Name: "名称"}
	cases := []string{"utf-16le", "utf-16"} // unsupported "unicode"

	for _, enc := range cases {
		match := offer.Match{
			ContentType: header.ContentType{Type: "application", Subtype: "json"},
			Language:    "cn",
			Charset:     enc,
			Data:        data.Of(model),
		}

		p := offer.XMLProcessor("xml", "  ")
		rw := httptest.NewRecorder()

		w := match.ApplyHeaders(rw)
		err := p(w, req, match.Data, "template", match.Language)

		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(rw.Header().Get(headername.ContentType)).To(Equal("application/json;charset=utf-16le"), enc)
		g.Expect(rw.Header().Get(headername.ContentLanguage)).To(Equal("cn"), enc)
		g.Expect(rw.Body.Bytes()).To(Equal([]byte{
			'<', 0, 'V', 0, 'a', 0, 'l', 0, 'i', 0, 'd', 0, 'X', 0, 'M', 0, 'L', 0, 'U', 0, 's', 0, 'e', 0, 'r', 0, '>', 0, '\n', 0,
			' ', 0, ' ', 0, '<', 0, 'N', 0, 'a', 0, 'm', 0, 'e', 0, '>', 0, 13, 84, 240, 121,
			'<', 0, '/', 0, 'N', 0, 'a', 0, 'm', 0, 'e', 0, '>', 0, '\n', 0,
			'<', 0, '/', 0, 'V', 0, 'a', 0, 'l', 0, 'i', 0, 'd', 0, 'X', 0, 'M', 0, 'L', 0, 'U', 0, 's', 0, 'e', 0, 'r', 0, '>', 0, '\n', 0,
		}), rw.Body.String(), enc)
	}
}

func TestXMLShouldReturnError(t *testing.T) {
	g := NewGomegaWithT(t)
	req := &http.Request{}
	w := httptest.NewRecorder()

	model := &ErrXMLUser{Msg: "oops"}

	p := offer.XMLProcessor("xml", "  ")

	err := p(w, req, data.Of(model), "template", "en")

	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("oops"))
}

type ValidXMLUser struct {
	Name string
}

type ErrXMLUser struct {
	Msg string
}

func (u *ErrXMLUser) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return errors.New(u.Msg)
}

//func xmltestErrorHandler(w http.ResponseWriter, err error) {
//	w.WriteHeader(500)
//	w.Write([]byte(err.Error()))
//}
