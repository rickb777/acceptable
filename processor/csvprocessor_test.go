package processor_test

import (
	"net/http/httptest"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/rickb777/acceptable"
	"github.com/rickb777/acceptable/processor"
)

func TestCSVShouldWriteResponseBody(t *testing.T) {
	g := NewGomegaWithT(t)
	models := []struct {
		stuff    interface{}
		expected string
	}{
		{"Joe Bloggs", "Joe Bloggs\n"},
		{[]string{"Red", "Green", "Blue"}, "Red,Green,Blue\n"},
		{[][]string{{"Red", "Green", "Blue"}, {"Cyan", "Magenta", "Yellow"}}, "Red,Green,Blue\nCyan,Magenta,Yellow\n"},
		{[]int{101, -5, 42}, "101,-5,42\n"},
		{[]int8{101, -5, 42}, "101,-5,42\n"},
		{[]uint{101, 42}, "101,42\n"},
		{[]uint8{101, 42}, "101,42\n"},
		{[][]int{{101, 42}, {39, 7}}, "101,42\n39,7\n"},
		{[][]uint{{101, 42}, {39, 7}}, "101,42\n39,7\n"},
		{Data{"x,y", 9, 4, true}, "\"x,y\",9,4,true\n"},
		{[]Data{{"x", 9, 4, true}, {"y", 7, 1, false}}, "x,9,4,true\ny,7,1,false\n"},
		{[]hidden{{tt(2001, 11, 29)}, {tt(2001, 11, 30)}}, "(2001-11-29),(2001-11-30)\n"},
		{[][]hidden{{{tt(2001, 12, 30)}, {tt(2001, 12, 31)}}}, "(2001-12-30),(2001-12-31)\n"},
		{[]*hidden{{tt(2001, 11, 29)}, {tt(2001, 11, 30)}}, "(2001-11-29),(2001-11-30)\n"},
		{[][]*hidden{{{tt(2001, 12, 30)}, {tt(2001, 12, 31)}}}, "(2001-12-30),(2001-12-31)\n"},
	}

	p := processor.CSV()

	for _, m := range models {
		w := httptest.NewRecorder()
		p(w, acceptable.Match{Data: m.stuff}, "")
		g.Expect(w.Body.String()).To(Equal(m.expected))
	}
}

func TestCSVShouldWriteResponseBodyWithTabs(t *testing.T) {
	g := NewGomegaWithT(t)
	models := []struct {
		stuff    interface{}
		expected string
	}{
		{"Joe Bloggs", "Joe Bloggs\n"},
		{[]string{"Red", "Green", "Blue"}, "Red\tGreen\tBlue\n"},
		{[][]string{{"Red", "Green", "Blue"}, {"Cyan", "Magenta", "Yellow"}}, "Red\tGreen\tBlue\nCyan\tMagenta\tYellow\n"},
		{[]int{101, -5, 42}, "101\t-5\t42\n"},
		{[]int8{101, -5, 42}, "101\t-5\t42\n"},
		{[]uint{101, 42}, "101\t42\n"},
		{[]uint8{101, 42}, "101\t42\n"},
		{[][]int{{101, 42}, {39, 7}}, "101\t42\n39\t7\n"},
		{[][]uint{{101, 42}, {39, 7}}, "101\t42\n39\t7\n"},
		{Data{"x", 9, 4, true}, "x\t9\t4\ttrue\n"},
		{[]Data{{"x", 9, 4, true}, {"y", 7, 1, false}}, "x\t9\t4\ttrue\ny\t7\t1\tfalse\n"},
	}

	p := processor.CSV('\t')

	for _, m := range models {
		w := httptest.NewRecorder()
		p(w, acceptable.Match{Data: m.stuff}, "")
		g.Expect(w.Body.String()).To(Equal(m.expected))
	}
}

func TestCSVShouldReturnError(t *testing.T) {
	g := NewGomegaWithT(t)
	w := httptest.NewRecorder()

	p := processor.CSV()

	err := p(w, acceptable.Match{}, "")

	g.Expect(err).To(HaveOccurred())
}

type Data struct {
	F1 string
	F2 int
	F3 uint
	F4 bool
}
