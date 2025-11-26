package offer_test

import (
	"encoding/xml"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rickb777/acceptable/data"
	"github.com/rickb777/acceptable/header"
	"github.com/rickb777/acceptable/headername"
	"github.com/rickb777/acceptable/offer"
	"github.com/rickb777/expect"
)

func TestXMLShouldWriteLazyResponseBody(t *testing.T) {
	req := &http.Request{}
	rw := httptest.NewRecorder()

	model := &ValidXMLUser{
		"Joe Bloggs",
	}

	match := offer.Match{
		ContentType: header.ContentType{MediaType: "application/json"},
		Language:    "en",
		Charset:     "utf-8",
		Data:        data.Lazy(func(string, string) (interface{}, error) { return model, nil }),
	}

	p := offer.XMLProcessor("xml")

	w := match.ApplyHeaders(rw)
	err := p(w, req, match.Data, "template", match.Language)

	expect.Error(err).Not().ToHaveOccurred(t)
	expect.String(rw.Header().Get(headername.ContentType)).ToBe(t, "application/json")
	expect.String(rw.Header().Get(headername.ContentLanguage)).ToBe(t, "en")
	expect.String(rw.Body.String()).ToBe(t, "<ValidXMLUser><Name>Joe Bloggs</Name></ValidXMLUser>\n")
}

func TestXMLShouldWriteSequenceResponseBody(t *testing.T) {
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
		ContentType: header.ContentType{MediaType: "application/json"},
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

	expect.Error(err).Not().ToHaveOccurred(t)
	expect.String(rw.Header().Get(headername.ContentType)).ToBe(t, "application/json")
	expect.String(rw.Header().Get(headername.ContentLanguage)).ToBe(t, "en")
	expect.String(rw.Body.String()).Info(rw.Body.String()).ToBe(t,
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
			"</xml>\n")
}

func TestXMlShouldWriteResponseBodyWithIndentation_utf_16be(t *testing.T) {
	req := &http.Request{}

	model := &ValidXMLUser{Name: "名称"}
	cases := []string{"utf-16be"} // unsupported: "unicodefffe"

	for _, enc := range cases {
		match := offer.Match{
			ContentType: header.ContentType{MediaType: "application/json"},
			Language:    "cn",
			Charset:     enc,
			Data:        data.Of(model),
		}

		p := offer.XMLProcessor("xml", "  ")
		rw := httptest.NewRecorder()

		w := match.ApplyHeaders(rw)
		err := p(w, req, match.Data, "template", match.Language)

		expect.Error(err).Not().ToHaveOccurred(t)
		expect.String(rw.Header().Get(headername.ContentType)).I(enc).ToBe(t, "application/json")
		expect.String(rw.Header().Get(headername.ContentLanguage)).I(enc).ToBe(t, "cn")
		expect.String(rw.Body.Bytes()).I(enc).ToBe(t, []byte{
			0, '<', 0, 'V', 0, 'a', 0, 'l', 0, 'i', 0, 'd', 0, 'X', 0, 'M', 0, 'L', 0, 'U', 0, 's', 0, 'e', 0, 'r', 0, '>', 0, '\n',
			0, ' ', 0, ' ', 0, '<', 0, 'N', 0, 'a', 0, 'm', 0, 'e', 0, '>', 84, 13, 121, 240,
			0, '<', 0, '/', 0, 'N', 0, 'a', 0, 'm', 0, 'e', 0, '>', 0, '\n', 0,
			'<', 0, '/', 0, 'V', 0, 'a', 0, 'l', 0, 'i', 0, 'd', 0, 'X', 0, 'M', 0, 'L', 0, 'U', 0, 's', 0, 'e', 0, 'r', 0, '>', 0, '\n',
		})
	}
}

func TestXMlShouldWriteResponseBodyWithIndentation_utf_16le(t *testing.T) {
	req := &http.Request{}

	model := &ValidXMLUser{Name: "名称"}
	cases := []string{"utf-16le", "utf-16"} // unsupported "unicode"

	for _, enc := range cases {
		match := offer.Match{
			ContentType: header.ContentType{MediaType: "application/json"},
			Language:    "cn",
			Charset:     enc,
			Data:        data.Of(model),
		}

		p := offer.XMLProcessor("xml", "  ")
		rw := httptest.NewRecorder()

		w := match.ApplyHeaders(rw)
		err := p(w, req, match.Data, "template", match.Language)

		expect.Error(err).Not().ToHaveOccurred(t)
		expect.String(rw.Header().Get(headername.ContentType)).I(enc).ToBe(t, "application/json")
		expect.String(rw.Header().Get(headername.ContentLanguage)).I(enc).ToBe(t, "cn")
		expect.String(rw.Body.Bytes()).I(enc).ToBe(t, []byte{
			'<', 0, 'V', 0, 'a', 0, 'l', 0, 'i', 0, 'd', 0, 'X', 0, 'M', 0, 'L', 0, 'U', 0, 's', 0, 'e', 0, 'r', 0, '>', 0, '\n', 0,
			' ', 0, ' ', 0, '<', 0, 'N', 0, 'a', 0, 'm', 0, 'e', 0, '>', 0, 13, 84, 240, 121,
			'<', 0, '/', 0, 'N', 0, 'a', 0, 'm', 0, 'e', 0, '>', 0, '\n', 0,
			'<', 0, '/', 0, 'V', 0, 'a', 0, 'l', 0, 'i', 0, 'd', 0, 'X', 0, 'M', 0, 'L', 0, 'U', 0, 's', 0, 'e', 0, 'r', 0, '>', 0, '\n', 0,
		})
	}
}

func TestXMLShouldReturnError(t *testing.T) {
	req := &http.Request{}
	w := httptest.NewRecorder()

	model := &ErrXMLUser{Msg: "oops"}

	p := offer.XMLProcessor("xml", "  ")

	err := p(w, req, data.Of(model), "template", "en")

	expect.Error(err).ToHaveOccurred(t)
	expect.String(err.Error()).ToContain(t, "oops")
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
