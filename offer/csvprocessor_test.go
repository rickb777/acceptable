package offer_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	dpkg "github.com/rickb777/acceptable/data"
	"github.com/rickb777/acceptable/offer"
	"github.com/rickb777/expect"
)

func TestCSVShouldWriteResponseBody(t *testing.T) {
	models := []struct {
		stuff    dpkg.Data
		expected string
	}{
		{dpkg.Of("Joe Bloggs"), "Joe Bloggs\n"},
		{dpkg.Lazy(func(dpkg.Chosen) (any, error) { return "Joe Bloggs", nil }), "Joe Bloggs\n"},
		{dpkg.Of([]string{"Red", "Green", "Blue"}), "Red,Green,Blue\n"},
		{dpkg.Of([][]string{{"Red", "Green", "Blue"}, {"Cyan", "Magenta", "Yellow"}}), "Red,Green,Blue\nCyan,Magenta,Yellow\n"},
		{dpkg.Of([]int{101, -5, 42}), "101,-5,42\n"},
		{dpkg.Of([]int8{101, -5, 42}), "101,-5,42\n"},
		{dpkg.Of([]uint{101, 42}), "101,42\n"},
		{dpkg.Of([]uint8{101, 42}), "101,42\n"},
		{dpkg.Of([][]int{{101, 42}, {39, 7}}), "101,42\n39,7\n"},
		{dpkg.Of([][]uint{{101, 42}, {39, 7}}), "101,42\n39,7\n"},
		{dpkg.Of(Data{"x,y", 9, 4, true}), "\"x,y\",9,4,true\n"},
		{dpkg.Of([]Data{{"x", 9, 4, true}, {"y", 7, 1, false}}), "x,9,4,true\ny,7,1,false\n"},
		{dpkg.Of([]hidden{{tt(2001, 11, 29)}, {tt(2001, 11, 30)}}), "(2001-11-29),(2001-11-30)\n"},
		{dpkg.Of([][]hidden{{{tt(2001, 12, 30)}, {tt(2001, 12, 31)}}}), "(2001-12-30),(2001-12-31)\n"},
		{dpkg.Of([]*hidden{{tt(2001, 11, 29)}, {tt(2001, 11, 30)}}), "(2001-11-29),(2001-11-30)\n"},
		{dpkg.Of([][]*hidden{{{tt(2001, 12, 30)}, {tt(2001, 12, 31)}}}), "(2001-12-30),(2001-12-31)\n"},
	}

	req := &http.Request{}
	p := offer.CSVProcessor(0)

	for _, m := range models {
		w := httptest.NewRecorder()
		err := p(w, req, m.stuff, dpkg.Chosen{})
		expect.String(w.Body.String(), err).ToBe(t, m.expected)
	}
}

func TestCSVShouldWriteResponseBodyWithTabs(t *testing.T) {
	models := []struct {
		stuff    dpkg.Data
		expected string
	}{
		{dpkg.Of("Joe Bloggs"), "Joe Bloggs\n"},
		{dpkg.Of([]string{"Red", "Green", "Blue"}), "Red\tGreen\tBlue\n"},
		{dpkg.Of([][]string{{"Red", "Green", "Blue"}, {"Cyan", "Magenta", "Yellow"}}), "Red\tGreen\tBlue\nCyan\tMagenta\tYellow\n"},
		{dpkg.Of([]int{101, -5, 42}), "101\t-5\t42\n"},
		{dpkg.Of([]int8{101, -5, 42}), "101\t-5\t42\n"},
		{dpkg.Of([]uint{101, 42}), "101\t42\n"},
		{dpkg.Of([]uint8{101, 42}), "101\t42\n"},
		{dpkg.Of([][]int{{101, 42}, {39, 7}}), "101\t42\n39\t7\n"},
		{dpkg.Of([][]uint{{101, 42}, {39, 7}}), "101\t42\n39\t7\n"},
		{dpkg.Of(Data{"x", 9, 4, true}), "x\t9\t4\ttrue\n"},
		{dpkg.Of([]Data{{"x", 9, 4, true}, {"y", 7, 1, false}}), "x\t9\t4\ttrue\ny\t7\t1\tfalse\n"},
	}

	req := &http.Request{}
	p := offer.CSVProcessor(0, '\t')

	for _, m := range models {
		w := httptest.NewRecorder()
		err := p(w, req, m.stuff, dpkg.Chosen{})
		expect.String(w.Body.String(), err).ToBe(t, m.expected)
	}
}

type Data struct {
	F1 string
	F2 int
	F3 uint
	F4 bool
}
