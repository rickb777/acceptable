package header_test

import (
	"io"
	"net/http"
	"testing"

	"github.com/rickb777/acceptable/data"
	. "github.com/rickb777/acceptable/header"
	"github.com/rickb777/acceptable/offer"
	"github.com/rickb777/expect"
)

func TestMediaRange_String(t *testing.T) {
	ct := ParseContentType("text/html;charset=utf-8;level=1").AsMediaRange(0.5)

	expect.String(ct.String()).ToBe(t, "text/html;charset=utf-8;level=1;q=0.5")
}

func TestPrecedenceValues_String(t *testing.T) {
	vv := PrecedenceValues{{Value: "iso-8859-5", Quality: DefaultQuality}, {Value: "unicode-1-1", Quality: 0.8}}

	expect.String(vv.String()).ToBe(t, "iso-8859-5, unicode-1-1;q=0.8")
}

func TestOffer_String(t *testing.T) {
	p := func(_ io.Writer, _ *http.Request, _ data.Data, _, _ string) error {
		return nil
	}
	o := offer.Of(p, "text/html").With(nil, "en")

	expect.String(o.String()).ToBe(t, "Accept: text/html. Accept-Language: en")
}
