package acceptable_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/rickb777/acceptable"
	"github.com/rickb777/acceptable/data"
)

func TestBinaryShouldWriteResponseBody(t *testing.T) {
	g := NewGomegaWithT(t)
	names := []string{"Alice\n", "Bob\n", "Charles\n"}

	models := []struct {
		stuff    data.Data
		expected string
	}{
		{data.Of([]byte("Joe Bloggs")), "Joe Bloggs"},
		{data.Lazy(func(string, string) (interface{}, error) { return []byte("Joe Bloggs"), nil }), "Joe Bloggs"},
		{data.Sequence(
			func(string, string) (interface{}, error) {
				if len(names) == 0 {
					return nil, nil
				}
				n := []byte(names[0])
				names = names[1:]
				return n, nil
			}),
			"Alice\nBob\nCharles\n",
		},
		{data.Of(strings.NewReader("Joe Bloggs")), "Joe Bloggs"},
		{data.Of(&simpleReader{s: "Joe Bloggs"}), "Joe Bloggs"},
		{data.Of(nil), ""},
		{nil, ""},
	}

	req := &http.Request{}
	p := acceptable.Binary()

	for _, m := range models {
		w := httptest.NewRecorder()
		err := p(w, req, m.stuff, "", "")
		g.Expect(w.Body.String(), err).To(Equal(m.expected))
	}
}

func TestBinaryShouldNotReturnError(t *testing.T) {
	g := NewGomegaWithT(t)
	w := httptest.NewRecorder()

	req := &http.Request{}
	p := acceptable.Binary()

	err := p(w, req, nil, "", "")

	g.Expect(err).NotTo(HaveOccurred())
}

type simpleReader struct {
	s string
}

func (s *simpleReader) Read(p []byte) (n int, err error) {
	l := len(s.s)
	if l == 0 {
		return 0, io.EOF
	}

	if l > len(p) {
		copy(p, s.s[:len(p)])
		s.s = s.s[len(p):]
		return len(p), nil
	}

	copy(p, s.s)
	s.s = ""
	return l, nil
}
