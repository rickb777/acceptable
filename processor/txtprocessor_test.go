package processor_test

import (
	"net/http/httptest"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/rickb777/negotiator/processor"
)

func TestTXTShouldProcessAcceptHeader(t *testing.T) {
	g := NewGomegaWithT(t)
	var acceptTests = []struct {
		acceptheader string
		expected     bool
	}{
		{"text/plain", true},
		{"text/*", true},
		{"text/csv", false},
	}

	p := processor.TXT()

	for _, tt := range acceptTests {
		result := p.CanProcess(tt.acceptheader, "")
		g.Expect(result).To(Equal(tt.expected), "Should process "+tt.acceptheader)
	}
}

func TestTXTShouldSetContentTypeHeader(t *testing.T) {
	g := NewGomegaWithT(t)

	p := processor.TXT().(processor.ContentTypeSettable).WithContentType("text/foo")

	g.Expect(p.ContentType()).To(Equal("text/foo"))
}

func TestTXTShouldSetResponseBody(t *testing.T) {
	g := NewGomegaWithT(t)
	models := []struct {
		stuff    interface{}
		expected string
	}{
		{"Joe Bloggs", "Joe Bloggs\n"},
		{hidden{tt(2001, 10, 31)}, "(2001-10-31)\n"},
		{tm{"Joe Bloggs"}, "Joe Bloggs\n"},
	}

	p := processor.TXT()

	for _, m := range models {
		recorder := httptest.NewRecorder()
		p.Process(recorder, "", m.stuff)
		g.Expect(recorder.Body.String()).To(Equal(m.expected))
	}
}

func TestTXTShouldReturnError(t *testing.T) {
	g := NewGomegaWithT(t)
	recorder := httptest.NewRecorder()

	p := processor.TXT()

	err := p.Process(recorder, "", make(chan int, 0))

	g.Expect(err).To(HaveOccurred())
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
