package header

import (
	"fmt"
	"strings"

	"github.com/rickb777/acceptable/internal"
)

// ETag is an entity tag used for content matching comparisons.
// See https://tools.ietf.org/html/rfc7232
type ETag struct {
	Hash string
	Weak bool
}

// ETag is a slice of ETag.
type ETags []ETag

// WeaklyMatches finds whether the tags match weakly.
// See https://tools.ietf.org/html/rfc7232#section-2.3.2
func (es ETags) WeaklyMatches(strongHash string) bool {
	for _, e := range es {
		if strongHash == e.Hash {
			return true
		}
	}
	return false
}

// StronglyMatches finds whether the tags match strongly.
// This ignores all weak etags in es.
func (es ETags) StronglyMatches(strongHash string) bool {
	for _, e := range es {
		// strong hash never matches a weak ETag
		if !e.Weak && strongHash == e.Hash {
			return true
		}
	}
	return false
}

func ETagsOf(s string) ETags {
	if s == "" {
		return nil
	}
	parts := internal.Split(s, ",").TrimSpace()
	es := make(ETags, len(parts))
	for i, p := range parts {
		es[i] = eTagOf(p)
	}
	return es
}

func eTagOf(s string) ETag {
	if s == "*" {
		return ETag{Hash: "*"}
	}

	var e ETag
	if strings.HasPrefix(s, "W/") {
		e.Weak = true
		e.Hash = s[3 : len(s)-1]
	} else {
		e = ETag{Hash: s[1 : len(s)-1]}
	}
	return e
}

func (etag ETag) String() string {
	if etag.Weak {
		return fmt.Sprintf("W/%q", etag.Hash)
	}
	return fmt.Sprintf("%q", etag.Hash)
}
