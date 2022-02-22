package header_test

import (
	"io"
	"net/http"
	"testing"

	"github.com/onsi/gomega"
	"github.com/rickb777/acceptable/data"
	. "github.com/rickb777/acceptable/header"
	"github.com/rickb777/acceptable/offer"
)

func TestMediaRange_String(t *testing.T) {
	g := gomega.NewWithT(t)

	ct := ParseContentType("text/html;charset=utf-8;level=1").AsMediaRange(0.5)

	g.Expect(ct.String()).To(gomega.Equal("text/html;charset=utf-8;level=1;q=0.5"))
}

func TestPrecedenceValues_String(t *testing.T) {
	g := gomega.NewWithT(t)

	vv := PrecedenceValues{{Value: "iso-8859-5", Quality: DefaultQuality}, {Value: "unicode-1-1", Quality: 0.8}}

	g.Expect(vv.String()).To(gomega.Equal("iso-8859-5, unicode-1-1;q=0.8"))
}

func TestOffer_String(t *testing.T) {
	g := gomega.NewWithT(t)

	p := func(_ io.Writer, _ *http.Request, _ data.Data, _, _ string) error {
		return nil
	}
	o := offer.Of(p, "text/html").With(nil, "en")

	g.Expect(o.String()).To(gomega.Equal("Accept: text/html. Accept-Language: en"))
}
