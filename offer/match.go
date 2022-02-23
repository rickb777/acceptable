package offer

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/rickb777/acceptable/data"
	"github.com/rickb777/acceptable/header"
	"github.com/rickb777/acceptable/headername"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/htmlindex"
)

// Match holds the best-matched offer after content negotiation and is used for response rendering.
type Match struct {
	header.ContentType
	Language           string
	Charset            string
	Vary               []string
	Data               data.Data
	Render             Processor
	StatusCodeOverride int
}

//-------------------------------------------------------------------------------------------------

// ApplyHeaders sets response headers so that the user agent is notified of the content
// negotiation decisions made. Four headers may be set, depending on context.
//
//   * Content-Type is always set.
//   * Content-Language is set when a language was selected.
//   * Content-Encoding is set when the character set is being transcoded
//   * Vary is set to list the accept headers that led to the three decisions above.
//
func (m Match) ApplyHeaders(rw http.ResponseWriter) io.Writer {
	charset := "utf-8"

	var enc encoding.Encoding
	if m.Charset != "" && m.Charset != "utf-8" {
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

	if m.IsTextual() {
		ct := fmt.Sprintf("%s/%s;charset=%s", m.Type, m.Subtype, charset)
		rw.Header().Set(headername.ContentType, ct)

		if m.Language != "" && m.Language != "*" {
			rw.Header().Set(headername.ContentLanguage, m.Language)
		}
	} else {
		ct := fmt.Sprintf("%s/%s", m.Type, m.Subtype)
		rw.Header().Set(headername.ContentType, ct)
	}

	if len(m.Vary) > 0 {
		rw.Header().Set(headername.Vary, strings.Join(m.Vary, ", "))
	}

	if enc != nil {
		return enc.NewEncoder().Writer(rw)
	}

	return rw
}

func (m Match) String() string {
	d := ""
	if m.Data == nil {
		d = "; no data"
	}
	r := ""
	if m.Render == nil {
		r = "; no renderer"
	}
	a := ""
	if m.StatusCodeOverride != 0 {
		a = "; not accepted"
	}
	return fmt.Sprintf("%s/%s; charset=%s; lang=%s vary=%v%s%s%s", m.Type, m.Subtype, m.Charset, m.Language, m.Vary, d, r, a)
}
