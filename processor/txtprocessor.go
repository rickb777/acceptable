package processor

import (
	"encoding"
	"fmt"
	"net/http"

	"github.com/rickb777/acceptable"
)

const defaultTxtContentType = "text/plain; charset=utf-8"

// TXT creates an output processor that serialises strings in text/plain form.
// Model values should be one of the following:
//
// * string
//
// * fmt.Stringer
//
// * encoding.TextMarshaler
func TXT() acceptable.Processor {
	return func(w http.ResponseWriter, match acceptable.Match, template string, dataModel interface{}) error {
		match.ApplyHeaders(w)

		s, ok := dataModel.(string)
		if ok {
			return WriteWithNewline(w, []byte(s))
		}

		st, ok := dataModel.(fmt.Stringer)
		if ok {
			return WriteWithNewline(w, []byte(st.String()))
		}

		tm, ok := dataModel.(encoding.TextMarshaler)
		if ok {
			b, err := tm.MarshalText()
			if err != nil {
				return err
			}
			return WriteWithNewline(w, b)
		}

		return fmt.Errorf("Unsupported type for TXT: %T", dataModel)
	}
}
