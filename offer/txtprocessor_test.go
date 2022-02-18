package offer_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/rickb777/acceptable/data"
	"github.com/rickb777/acceptable/offer"
)

func TestTXTShouldWriteResponseBody(t *testing.T) {
	g := NewGomegaWithT(t)
	req := &http.Request{}
	names1 := []string{"Alice\n", "Bob\n", "Charles\n"}
	names2 := []string{"Alice ", "Bob ", "Charles"}

	models := []struct {
		stuff    data.Data
		expected string
	}{
		{data.Of("Joe Bloggs"), "Joe Bloggs\n"},
		{data.Of("Joe Bloggs\n"), "Joe Bloggs\n"},
		{data.Of([]byte("Joe Bloggs")), "Joe Bloggs\n"},
		{data.Lazy(func(string, string) (interface{}, error) { return "Joe Bloggs", nil }), "Joe Bloggs\n"},
		{data.Sequence(
			stringSequence(names1)),
			"Alice\nBob\nCharles\n",
		},
		{data.Sequence(
			stringSequence(names2)),
			"Alice Bob Charles\n",
		},
		{data.Of(hidden{tt(2001, 10, 31)}), "(2001-10-31)\n"},
		{data.Of(tm{"Joe Bloggs"}), "Joe Bloggs\n"},
		{data.Of(nil), "\n"},
	}

	p := offer.TXTProcessor()

	for _, m := range models {
		w := httptest.NewRecorder()
		err := p(w, req, m.stuff, "", "")
		g.Expect(w.Body.String(), err).To(Equal(m.expected))
	}
}

func stringSequence(names []string) func(string, string) (interface{}, error) {
	return func(string, string) (interface{}, error) {
		if len(names) == 0 {
			return nil, nil
		}
		n := names[0]
		names = names[1:]
		return n, nil
	}
}

func TestTXTShouldNotReturnError(t *testing.T) {
	g := NewGomegaWithT(t)
	req := &http.Request{}
	w := httptest.NewRecorder()

	p := offer.TXTProcessor()

	err := p(w, req, nil, "", "")

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
