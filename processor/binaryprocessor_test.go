package processor_test

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rickb777/acceptable"

	. "github.com/onsi/gomega"
	"github.com/rickb777/acceptable/processor"
)

func TestBinaryShouldWriteResponseBody(t *testing.T) {
	g := NewGomegaWithT(t)
	models := []struct {
		stuff    interface{}
		expected string
	}{
		{[]byte("Joe Bloggs"), "Joe Bloggs"},
		{func() (interface{}, error) { return []byte("Joe Bloggs"), nil }, "Joe Bloggs"},
		{strings.NewReader("Joe Bloggs"), "Joe Bloggs"},
		{nil, ""},
	}

	p := processor.Binary()

	for _, m := range models {
		w := httptest.NewRecorder()
		p(w, acceptable.Match{Data: m.stuff}, "")
		g.Expect(w.Body.String()).To(Equal(m.expected))
	}
}

func TestBinaryShouldNotReturnError(t *testing.T) {
	g := NewGomegaWithT(t)
	w := httptest.NewRecorder()

	p := processor.Binary()

	err := p(w, acceptable.Match{}, "")

	g.Expect(err).NotTo(HaveOccurred())
}
