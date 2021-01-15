package processor_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/rickb777/acceptable"
	"github.com/rickb777/acceptable/data"
	"github.com/rickb777/acceptable/processor"
)

func TestBinaryShouldWriteResponseBody(t *testing.T) {
	g := NewGomegaWithT(t)
	models := []struct {
		stuff    data.Data
		expected string
	}{
		{data.Of([]byte("Joe Bloggs")), "Joe Bloggs"},
		{data.Lazy(func(string, string) (interface{}, string, error) { return []byte("Joe Bloggs"), "", nil }), "Joe Bloggs"},
		{data.Of(strings.NewReader("Joe Bloggs")), "Joe Bloggs"},
		{nil, ""},
	}

	req := &http.Request{}
	p := processor.Binary()

	for _, m := range models {
		w := httptest.NewRecorder()
		p(w, req, acceptable.Match{Data: m.stuff}, "")
		g.Expect(w.Body.String()).To(Equal(m.expected))
	}
}

func TestBinaryShouldNotReturnError(t *testing.T) {
	g := NewGomegaWithT(t)
	w := httptest.NewRecorder()

	req := &http.Request{}
	p := processor.Binary()

	err := p(w, req, acceptable.Match{}, "")

	g.Expect(err).NotTo(HaveOccurred())
}
