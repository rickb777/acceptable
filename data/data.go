// package data provides wrappers for response data, optionally including response headers
// such as ETag and Cache-Control.
package data

import (
	"fmt"
	"net/http"
	"time"
)

type Data interface {
	// Content returns the data as a value that can be processed by encoders such as "encoding/json"
	Content(template, language string) (interface{}, error)

	// Headers returns response headers relating to the data (optional)
	Headers() map[string]interface{}
}

// Of wraps a data value.
func Of(v interface{}) Value {
	return Value{v: v}
}

// Lazy wraps a function that supplies a data value, but only when it is needed.
func Lazy(fn func(template, language string) (interface{}, error)) Value {
	return Value{v: fn}
}

// Value is a simple implementation of Data.
type Value struct {
	v interface{}
	h map[string]interface{}
}

func (v Value) Content(template, language string) (result interface{}, err error) {
	result = v.v
loop:
	for {
		switch fn := result.(type) {
		case func(string, string) (interface{}, error):
			result, err = fn(template, language)
		default:
			break loop
		}
		if err != nil {
			return nil, err
		}
	}
	return result, err
}

func (v Value) Headers() map[string]interface{} {
	return v.h
}

// With returns a copy of v with extra headers attached. These are passed in as key+value pairs.
// The header names should be in normal form, e.g. "Last-Modified" instead of "last-modified",
// but this is not mandatory. The values are simple strings, numbers etc. Or they can be
// func(interface{}) string, in which case they will be called using the result of Content.
func (v Value) With(hdr string, value interface{}, others ...interface{}) Value {
	if v.h == nil {
		v.h = make(map[string]interface{})
	}
	v.h[hdr] = value
	for i := 1; i < len(others); i += 2 {
		v.h[others[i-1].(string)] = others[i]
	}
	return v
}

// ETag computes and sets the entity tag header on the response using some hash value. This is used for
// efficient conditional requests, possibly avoiding network traffic. The parameter fn evaluates the
// hash lazily based on the content of this Data (which may also be evaluated lazily).
func (v Value) ETag(fn func(interface{}) string, weak ...bool) Value {
	fn2 := func(d interface{}) string {
		if len(weak) > 0 && weak[0] {
			return fmt.Sprintf("W/%q", fn(d))
		} else {
			return fmt.Sprintf("%q", fn(d))
		}
	}
	return v.With("ETag", fn2)
}

// LastModified sets the time at which the content was last modified. This allows for conditional
// requests, possibly avoiding network traffic. ETag takes precedence.
func (v Value) LastModified(at time.Time) Value {
	return v.With("Last-Modified", at.Format(time.RFC1123))
}

// Expires sets the time at which the response becomes stale. MaxAge takes precedence.
func (v Value) Expires(at time.Time) Value {
	return v.With("Expires", at.Format(time.RFC1123))
}

// MaxAge sets the max-age header on the response. This is used to allow caches to avoid repeating
// the request until the max age has expired, after which time the resource is considered stale.
func (v Value) MaxAge(max time.Duration) Value {
	return v.With("Cache-Control", fmt.Sprintf("max-age=%d", max/time.Second))
}

// NoCache sets cache control headers to prevent the response being cached.
func (v Value) NoCache() Value {
	return v.With("Cache-Control", "no-cache, must-revalidate", "Pragma", "no-cache")
}

// GetContentAndApplyExtraHeaders applies all lazy functions to produce the resulting content to be
// rendered; this value is returned. It also sets any extra response headers.
func GetContentAndApplyExtraHeaders(rw http.ResponseWriter, d Data, template, language string) (interface{}, error) {
	if d == nil {
		return nil, nil
	}

	v, err := d.Content(template, language)
	if err != nil {
		return nil, err
	}

	for hn, hv := range d.Headers() {
		var s string
		switch k := hv.(type) {
		case func(interface{}) string:
			s = k(v)
		case string:
			s = k
		default:
			s = fmt.Sprintf("%v", hv)
		}
		rw.Header().Set(hn, s)
	}

	return v, nil
}
