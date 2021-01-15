package processor_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/rickb777/acceptable"
	"github.com/rickb777/acceptable/data"
	"github.com/rickb777/acceptable/processor"
)

func TestCSVShouldWriteResponseBody(t *testing.T) {
	g := NewGomegaWithT(t)
	models := []struct {
		stuff    data.Data
		expected string
	}{
		{data.Of("Joe Bloggs"), "Joe Bloggs\n"},
		{data.Lazy(func(string, string) (interface{}, string, error) { return "Joe Bloggs", "", nil }), "Joe Bloggs\n"},
		{data.Of([]string{"Red", "Green", "Blue"}), "Red,Green,Blue\n"},
		{data.Of([][]string{{"Red", "Green", "Blue"}, {"Cyan", "Magenta", "Yellow"}}), "Red,Green,Blue\nCyan,Magenta,Yellow\n"},
		{data.Of([]int{101, -5, 42}), "101,-5,42\n"},
		{data.Of([]int8{101, -5, 42}), "101,-5,42\n"},
		{data.Of([]uint{101, 42}), "101,42\n"},
		{data.Of([]uint8{101, 42}), "101,42\n"},
		{data.Of([][]int{{101, 42}, {39, 7}}), "101,42\n39,7\n"},
		{data.Of([][]uint{{101, 42}, {39, 7}}), "101,42\n39,7\n"},
		{data.Of(Data{"x,y", 9, 4, true}), "\"x,y\",9,4,true\n"},
		{data.Of([]Data{{"x", 9, 4, true}, {"y", 7, 1, false}}), "x,9,4,true\ny,7,1,false\n"},
		{data.Of([]hidden{{tt(2001, 11, 29)}, {tt(2001, 11, 30)}}), "(2001-11-29),(2001-11-30)\n"},
		{data.Of([][]hidden{{{tt(2001, 12, 30)}, {tt(2001, 12, 31)}}}), "(2001-12-30),(2001-12-31)\n"},
		{data.Of([]*hidden{{tt(2001, 11, 29)}, {tt(2001, 11, 30)}}), "(2001-11-29),(2001-11-30)\n"},
		{data.Of([][]*hidden{{{tt(2001, 12, 30)}, {tt(2001, 12, 31)}}}), "(2001-12-30),(2001-12-31)\n"},
	}

	req := &http.Request{}
	p := processor.CSV()

	for _, m := range models {
		w := httptest.NewRecorder()
		p(w, req, acceptable.Match{Data: m.stuff}, "")
		g.Expect(w.Body.String()).To(Equal(m.expected))
	}
}

func TestCSVShouldWriteResponseBodyWithTabs(t *testing.T) {
	g := NewGomegaWithT(t)
	models := []struct {
		stuff    data.Data
		expected string
	}{
		{data.Of("Joe Bloggs"), "Joe Bloggs\n"},
		{data.Of([]string{"Red", "Green", "Blue"}), "Red\tGreen\tBlue\n"},
		{data.Of([][]string{{"Red", "Green", "Blue"}, {"Cyan", "Magenta", "Yellow"}}), "Red\tGreen\tBlue\nCyan\tMagenta\tYellow\n"},
		{data.Of([]int{101, -5, 42}), "101\t-5\t42\n"},
		{data.Of([]int8{101, -5, 42}), "101\t-5\t42\n"},
		{data.Of([]uint{101, 42}), "101\t42\n"},
		{data.Of([]uint8{101, 42}), "101\t42\n"},
		{data.Of([][]int{{101, 42}, {39, 7}}), "101\t42\n39\t7\n"},
		{data.Of([][]uint{{101, 42}, {39, 7}}), "101\t42\n39\t7\n"},
		{data.Of(Data{"x", 9, 4, true}), "x\t9\t4\ttrue\n"},
		{data.Of([]Data{{"x", 9, 4, true}, {"y", 7, 1, false}}), "x\t9\t4\ttrue\ny\t7\t1\tfalse\n"},
	}

	req := &http.Request{}
	p := processor.CSV('\t')

	for _, m := range models {
		w := httptest.NewRecorder()
		p(w, req, acceptable.Match{Data: m.stuff}, "")
		g.Expect(w.Body.String()).To(Equal(m.expected))
	}
}

type Data struct {
	F1 string
	F2 int
	F3 uint
	F4 bool
}
