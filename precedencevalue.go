package acceptable

import (
	"fmt"
	"strings"
)

// KV holds a parameter with a key and optional value.
type KV struct {
	Key, Value string
}

// PrecedenceValue is a value and associate quality between 0.0 and 1.0
type PrecedenceValue struct {
	Value   string
	Quality float64
}

// PrecedenceValues holds a slice of precedence values.
type PrecedenceValues []PrecedenceValue

// wvByPrecedence implements sort.Interface for []PrecedenceValue based
// on the precedence rules. The data will be returned sorted decending
type wvByPrecedence []PrecedenceValue

func (a wvByPrecedence) Len() int      { return len(a) }
func (a wvByPrecedence) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

func (a wvByPrecedence) Less(i, j int) bool {
	// qualities are floats so we don't use == directly
	if a[i].Quality > a[j].Quality {
		return true
	} else if a[i].Quality < a[j].Quality {
		return false
	}
	return false
}

func (pvs PrecedenceValues) WithDefault() PrecedenceValues {
	if len(pvs) == 0 {
		return []PrecedenceValue{{Value: "*", Quality: DefaultQuality}}
	}
	return pvs
}

func (pvs PrecedenceValues) String() string {
	buf := &strings.Builder{}
	comma := ""
	for _, pv := range pvs {
		buf.WriteString(comma)
		buf.WriteString(pv.String())
		comma = ", "
	}
	return buf.String()
}

func (pv PrecedenceValue) String() string {
	buf := &strings.Builder{}
	buf.WriteString(pv.Value)
	if pv.Quality < DefaultQuality {
		fmt.Fprintf(buf, ";q=%g", pv.Quality)
	}
	return buf.String()
}
