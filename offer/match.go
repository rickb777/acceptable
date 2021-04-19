package offer

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/rickb777/acceptable/data"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/htmlindex"
)

// Match holds the best-matched offer after content negotiation and is used for response rendering.
type Match struct {
	Type     string
	Subtype  string
	Language string
	Charset  string
	Vary     []string
	Data     data.Data
	Render   Processor
}

//-------------------------------------------------------------------------------------------------

// ApplyHeaders sets response headers so that the user agent is notified of the content
// negotiation decisons made. Four headers may be set, depending on context.
//
//   * Content-Type is always set.
//   * Content-Language is set when a language was selected.
//   * Content-Encoding is set when the character set is being transcoded
//   * Vary is set to list the accept headers that led to the three decisions above.
//
func (m Match) ApplyHeaders(rw http.ResponseWriter) io.Writer {
	charset := "utf-8"

	var enc encoding.Encoding
	if m.Charset != "" {
		var err error
		enc, err = htmlindex.Get(m.Charset)
		if err == nil {
			// get the canonical name of the encoding
			charset, _ = htmlindex.Name(enc)
			if charset == "utf-8" {
				enc = nil // not needed
			}
		}
	}

	if m.isTextual() {
		ct := fmt.Sprintf("%s/%s;charset=%s", m.Type, m.Subtype, charset)
		rw.Header().Set("Content-Type", ct)

		if m.Language != "" && m.Language != "*" {
			rw.Header().Set("Content-Language", m.Language)
		}
	} else {
		ct := fmt.Sprintf("%s/%s", m.Type, m.Subtype)
		rw.Header().Set("Content-Type", ct)
	}

	if len(m.Vary) > 0 {
		rw.Header().Set("Vary", strings.Join(m.Vary, ", "))
	}

	if enc != nil {
		return enc.NewEncoder().Writer(rw)
	}

	return rw
}

func (m Match) isTextual() bool {
	if m.Type == "text" {
		return true
	}

	if m.Type == "application" {
		return m.Subtype == "json" ||
			m.Subtype == "xml" ||
			strings.HasSuffix(m.Subtype, "+xml") ||
			strings.HasSuffix(m.Subtype, "+json")
	}

	if m.Type == "image" {
		return strings.HasSuffix(m.Subtype, "+xml")
	}

	return false
}

func (m Match) String() string {
	d := "; no data"
	if m.Data != nil {
		d = ""
	}
	r := "; no renderer"
	if m.Render != nil {
		r = ""
	}
	return fmt.Sprintf("%s/%s; charset=%s; lang=%s vary=%v%s%s", m.Type, m.Subtype, m.Charset, m.Language, m.Vary, d, r)
}
