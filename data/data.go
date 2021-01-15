// package data provides wrappers for response data, optionally including response headers
// such as ETag.
package data

import "github.com/rickb777/acceptable/internal"

type Data interface {
	// Content returns the data as a value that can be processed by encoders such as "encoding/json"
	Content(template, language string) (interface{}, error)

	// Headers returns response headers relating to the data (optional)
	Headers() map[string]string
}

// Of wraps a data value.
func Of(v interface{}) Data {
	return Value{v: v}
}

// Lazy wraps a function that supplies a data value, but only when it is needed.
func Lazy(fn func(template, language string) (interface{}, error)) Data {
	return Value{v: fn}
}

// Value is a simple implementation of Data.
type Value struct {
	v interface{}
	h map[string]string
}

func (v Value) Content(template, language string) (interface{}, error) {
	return internal.CallDataSuppliers2(v.v, template, language)
}

func (v Value) Headers() map[string]string {
	return v.h
}

// With returns a copy of v with extra headers attached. These are passed in as key+value pairs.
// The header names should be in normal form, e.g. "Last-Modified" instead of "last-modified",
// but this is not mandatory.
func (v Value) With(hdr, value string, others ...string) Value {
	if v.h == nil {
		v.h = make(map[string]string)
	}
	v.h[hdr] = value
	for i := 1; i < len(others); i += 2 {
		v.h[others[i-1]] = others[i]
	}
	return v
}
