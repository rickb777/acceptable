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
	// The returned values are the data itself, a hash that will be used as the entity tag (if required),
	// and an error if arising. The hash should be blank if not needed.
	Content(template, language string) (interface{}, string, error)

	// Headers returns response headers relating to the data (optional)
	Headers() map[string]string
}

// Of wraps a data value. An optional entity tag can also be passed in. This is often the MD5 sum
// of the content, or something similar. If this is non-blank, the ETag response header will be sent.
func Of(v interface{}, etag ...string) *Value {
	if len(etag) == 0 {
		return &Value{value: v}
	}
	return &Value{value: v, etag: etag[0]}
}

// Lazy wraps a function that supplies a data value, but only when it is needed.
func Lazy(fn func(template, language string) (interface{}, string, error)) *Value {
	return &Value{supplier: fn}
}

// Value is a simple implementation of Data.
type Value struct {
	supplier func(template, language string) (interface{}, string, error)
	value    interface{}
	etag     string
	hdrs     map[string]string
}

func (v *Value) Content(template, language string) (result interface{}, etag string, err error) {
	if v.value != nil {
		return v.value, v.etag, nil
	}

	if v.supplier != nil {
		v.value, v.etag, err = v.supplier(template, language)
	}

	return v.value, v.etag, err
}

func (v Value) Headers() map[string]string {
	return v.hdrs
}

// With returns a copy of v with extra headers attached. These are passed in as key+value pairs.
// The header names should be in normal form, e.g. "Last-Modified" instead of "last-modified",
// but this is not mandatory. The values are simple strings, numbers etc. Or they can be
// func(interface{}) string, in which case they will be called using the result of Content.
func (v Value) With(hdr string, value string, others ...string) *Value {
	if v.hdrs == nil {
		v.hdrs = make(map[string]string)
	}
	v.hdrs[hdr] = value
	for i := 1; i < len(others); i += 2 {
		v.hdrs[others[i-1]] = others[i]
	}
	return &v
}

// LastModified sets the time at which the content was last modified. This allows for conditional
// requests, possibly avoiding network traffic. ETag takes precedence.
func (v Value) LastModified(at time.Time) *Value {
	return v.With("Last-Modified", at.Format(time.RFC1123))
}

// Expires sets the time at which the response becomes stale. MaxAge takes precedence.
func (v Value) Expires(at time.Time) *Value {
	return v.With("Expires", at.Format(time.RFC1123))
}

// MaxAge sets the max-age header on the response. This is used to allow caches to avoid repeating
// the request until the max age has expired, after which time the resource is considered stale.
func (v Value) MaxAge(max time.Duration) *Value {
	return v.With("Cache-Control", fmt.Sprintf("max-age=%d", max/time.Second))
}

// NoCache sets cache control headers to prevent the response being cached.
func (v Value) NoCache() *Value {
	return v.With("Cache-Control", "no-cache, must-revalidate", "Pragma", "no-cache")
}

// GetContentAndApplyExtraHeaders applies all lazy functions to produce the resulting content to be
// rendered; this value is returned. It also sets any extra response headers.
func GetContentAndApplyExtraHeaders(rw http.ResponseWriter, d Data, template, language string) (interface{}, error) {
	if d == nil {
		return nil, nil
	}

	v, etag, err := d.Content(template, language)
	if err != nil {
		return nil, err
	}

	for hn, hv := range d.Headers() {
		rw.Header().Set(hn, hv)
	}

	if etag != "" {
		rw.Header().Set("ETag", fmt.Sprintf("%q", etag))
	}

	return v, nil
}
