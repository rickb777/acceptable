package acceptable

import (
	"fmt"
	"net/http"
)

type Match struct {
	Type      string
	Subtype   string
	Language  string
	Charset   string
	Processor Processor
}

func (r *Match) ApplyHeaders(w http.ResponseWriter) {
	ct := fmt.Sprintf("%s/%s;charset=%s", r.Type, r.Subtype, orDefault(r.Charset, "utf-8"))
	w.Header().Set("Content-Type", ct)

	if r.Language != "" {
		w.Header().Set("Content-Language", r.Language)
	}
}

func orDefault(s, d string) string {
	if s == "" {
		return d
	}
	return s
}
