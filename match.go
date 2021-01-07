package acceptable

import (
	"fmt"
	"net/http"
)

type Match struct {
	Type     string
	Subtype  string
	Language string
	Charset  string
	Render   Processor
}

func (r *Match) ApplyHeaders(w http.ResponseWriter) {
	ct := fmt.Sprintf("%s/%s;charset=%s", r.Type, r.Subtype, orDefault(r.Charset, "utf-8"))
	w.Header().Set("Content-Type", ct)

	if r.Language != "" && r.Language != "*" {
		w.Header().Set("Content-Language", r.Language)
		w.Header().Set("Vary", "accept, accept-language")
	} else {
		w.Header().Set("Vary", "accept")
	}

}

func orDefault(s, d string) string {
	if s == "" {
		return d
	}
	return s
}
