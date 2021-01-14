package acceptable

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"golang.org/x/text/encoding/htmlindex"
)

// Match holds the best-matched offer after content negotiation and is used for response rendering.
type Match struct {
	Type     string
	Subtype  string
	Language string
	Charset  string
	Vary     []string
	Data     interface{}
	Render   Processor
}

// ApplyHeaders sets response headers so that the user agent is notified of the content
// negotiation decisons made. Four headers may be set, depending on context.
//
//   * Content-Type is always set.
//   * Content-Language is set when a language was selected.
//   * Content-Encoding is set when the character set is being transcoded
//   * Vary is set to list the accept headers that led to the three decisions above.
//
func (m Match) ApplyHeaders(rw http.ResponseWriter) (w io.Writer) {
	w = rw
	cs := "utf-8"
	vary := m.Vary

	if m.Charset != "" {
		enc, err := htmlindex.Get(m.Charset)
		if err == nil {
			// get the canonical name of the encoding
			cs, _ = htmlindex.Name(enc)
			rw.Header().Set("Content-Encoding", cs)
			vary = append(vary, "accept-charset")
			w = enc.NewEncoder().Writer(w)
		} else {
			Debug("%v\n", err)
		}
	}

	ct := fmt.Sprintf("%s/%s;charset=%s", m.Type, m.Subtype, cs)
	rw.Header().Set("Content-Type", ct)

	if m.Language != "" && m.Language != "*" {
		rw.Header().Set("Content-Language", m.Language)
	}

	rw.Header().Set("Vary", strings.Join(vary, ", "))

	return w
}
