package header_test

import (
	"net/http"
	"testing"

	"github.com/onsi/gomega"
	"github.com/rickb777/acceptable"
	. "github.com/rickb777/acceptable/header"
)

func TestContentType_String(t *testing.T) {
	g := gomega.NewWithT(t)

	ct := ContentTypeOf("text", "html", "charset=utf-8")
	ct.Extensions = append(ct.Extensions, KV{"level", "1"})

	g.Expect(ct.String()).To(gomega.Equal("text/html;charset=utf-8;level=1"))
}

func TestContentType_Wildcards(t *testing.T) {
	g := gomega.NewWithT(t)

	ct := ContentTypeOf("", "")

	g.Expect(ct.String()).To(gomega.Equal("*/*"))
}

func TestMediaRange_String(t *testing.T) {
	g := gomega.NewWithT(t)

	ct := MediaRange{
		ContentType: ContentTypeOf("text", "html", "charset=utf-8"),
		Quality:     0.5,
	}
	ct.Extensions = append(ct.Extensions, KV{"level", "1"})

	g.Expect(ct.String()).To(gomega.Equal("text/html;charset=utf-8;q=0.5;level=1"))
}

func TestPrecedenceValues_String(t *testing.T) {
	g := gomega.NewWithT(t)

	vv := PrecedenceValues{{Value: "iso-8859-5", Quality: DefaultQuality}, {Value: "unicode-1-1", Quality: 0.8}}

	g.Expect(vv.String()).To(gomega.Equal("iso-8859-5, unicode-1-1;q=0.8"))
}

func TestOffer_String(t *testing.T) {
	g := gomega.NewWithT(t)

	p := func(w http.ResponseWriter, req *http.Request, match acceptable.Match, template string) error {
		return nil
	}
	o := acceptable.OfferOf(p, "text/html").With(nil, "en")

	g.Expect(o.String()).To(gomega.Equal("Accept: text/html. Accept-Language: en"))
}
