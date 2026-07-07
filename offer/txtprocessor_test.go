package offer_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	dpkg "github.com/rickb777/acceptable/data"
	"github.com/rickb777/acceptable/offer"
	"github.com/rickb777/expect"
)

func TestTXTShouldWriteResponseBody(t *testing.T) {
	req := &http.Request{}
	names1 := []string{"Alice\n", "Bob\n", "Charles\n"}
	names2 := []string{"Alice ", "Bob ", "Charles"}

	models := []struct {
		stuff    dpkg.Data
		expected string
	}{
		{dpkg.Of("Joe Bloggs"), "Joe Bloggs\n"},
		{dpkg.Of("Joe Bloggs\n"), "Joe Bloggs\n"},
		{dpkg.Of([]byte("Joe Bloggs")), "Joe Bloggs\n"},
		{dpkg.Lazy(func(dpkg.Chosen) (any, error) { return "Joe Bloggs", nil }), "Joe Bloggs\n"},
		{dpkg.Sequence(
			stringSequence(names1)),
			"Alice\nBob\nCharles\n",
		},
		{dpkg.Sequence(
			stringSequence(names2)),
			"Alice Bob Charles\n",
		},
		{dpkg.Of(hidden{tt(2001, 10, 31)}), "(2001-10-31)\n"},
		{dpkg.Of(tm{"Joe Bloggs"}), "Joe Bloggs\n"},
		{dpkg.Of(nil), ""},
	}

	p := offer.TXTProcessor()

	for _, m := range models {
		w := httptest.NewRecorder()
		err := p(w, req, m.stuff, dpkg.Chosen{})
		expect.String(w.Body.String(), err).ToBe(t, m.expected)
	}
}

func stringSequence(names []string) func(params dpkg.Chosen) (any, error) {
	return func(params dpkg.Chosen) (any, error) {
		if len(names) == 0 {
			return nil, nil
		}
		n := names[0]
		names = names[1:]
		return n, nil
	}
}

func TestTXTShouldNotReturnError(t *testing.T) {
	req := &http.Request{}
	w := httptest.NewRecorder()

	p := offer.TXTProcessor()

	err := p(w, req, nil, dpkg.Chosen{})

	expect.Error(err).Not().ToHaveOccurred(t)
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
