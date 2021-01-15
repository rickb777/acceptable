package processor_test

import (
	"net/http/httptest"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/rickb777/acceptable"
	"github.com/rickb777/acceptable/data"
	"github.com/rickb777/acceptable/processor"
)

func TestTXTShouldWriteResponseBody(t *testing.T) {
	g := NewGomegaWithT(t)
	models := []struct {
		stuff    data.Data
		expected string
	}{
		{data.Of("Joe Bloggs"), "Joe Bloggs\n"},
		{data.Of("Joe Bloggs\n"), "Joe Bloggs\n"},
		{data.Lazy(func(string, string) (interface{}, error) { return "Joe Bloggs", nil }), "Joe Bloggs\n"},
		{data.Of(hidden{tt(2001, 10, 31)}), "(2001-10-31)\n"},
		{data.Of(tm{"Joe Bloggs"}), "Joe Bloggs\n"},
	}

	p := processor.TXT()

	for _, m := range models {
		w := httptest.NewRecorder()
		p(w, acceptable.Match{Data: m.stuff}, "")
		g.Expect(w.Body.String()).To(Equal(m.expected))
	}
}

func TestTXTShouldNotReturnError(t *testing.T) {
	g := NewGomegaWithT(t)
	w := httptest.NewRecorder()

	p := processor.TXT()

	err := p(w, acceptable.Match{}, "")

	g.Expect(err).NotTo(HaveOccurred())
}

func tt(y, m, d int) time.Time {
	return time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
}

type tm struct {
	s string
}

func (tm tm) MarshalText() (text []byte, err error) {
	return []byte(tm.s), nil
}

// has hidden fields
type hidden struct {
	d time.Time
}

func (h hidden) String() string {
	return "(" + h.d.Format("2006-01-02") + ")"
}
