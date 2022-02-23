package data

import (
	"fmt"
	"net/http"
	"time"

	"github.com/rickb777/acceptable/header"
	"github.com/rickb777/acceptable/headername"
)

// Data provides a source for response content. It is optimised for lazy evaluation, avoiding
// wasted processing.
//
// If necessary, Content will be called a second time, this time with dataRequired=true. The
// data must always be returned in this case. However the metadata will be ignored.
//
// The metadata can be nil if not needed.
type Data interface {
	// Meta returns the metadata that will be used to set response headers automatically.
	// The headers are ETag and Last-Modified.
	Meta(template, language string) (meta *Metadata, err error)

	// Content returns the data as a value that can be processed by encoders such as "encoding/json"
	// The returned values are the data itself, a boolean that is true if the data is in chunks and
	// there is more data to follow, and an error if arising. For chunked data, this method
	// will be called repeatedly until the boolean yields false or an error arises.
	Content(template, language string) (interface{}, bool, error)

	// Headers returns response headers relating to the data (optional)
	Headers() map[string]string
}

// Metadata provides optional entity tag and last modified information about some data.
type Metadata struct {
	Hash         string    // used as entity tag; blank if not required
	LastModified time.Time // used for Last-Modified header; zero if not required
}

// Of wraps a data value.
//
// If an entity tag is known, the ETag method should be used. Likewise, if a last-modified
// timestamp is known, the LastModified method should also be used.
func Of(v interface{}) *Value {
	return &Value{value: v}
}

// Lazy wraps a function that supplies a data value, but only fetches the data when it is needed.
//
// If an entity tag is known, the ETag method should be used. Likewise, if a last-modified
// timestamp is known, the LastModified method should also be used.
func Lazy(supplier func(template, language string) (interface{}, error)) *Value {
	return &Value{supplier: supplier, chunked: false}
}

// Sequence wraps a function that supplies data values in a sequence chunk by chunk. This
// function will not be not called until it is needed. When it is called, it will be called
// repeatedly until the returned value is nil or an error arises.
//
// Typical use might be where a response contains many database records that are obtained
// one by one to avoid the need to cache all results in memory before rendering.
//
// If an entity tag is known, the ETag method should be used. Likewise, if a last-modified
// timestamp is known, the LastModified method should also be used.
func Sequence(supplier func(template, language string) (interface{}, error)) *Value {
	return &Value{supplier: supplier, chunked: true}
}

// Value is a simple implementation of Data.
type Value struct {
	supplier     func(template, language string) (interface{}, error)
	chunked      bool
	etagFn       func(template, language string) (string, error)
	lastModFn    func(template, language string) (time.Time, error)
	value        interface{}
	etag         string
	lastModified time.Time
	hdrs         map[string]string
}

func (v *Value) Meta(template, language string) (meta *Metadata, err error) {
	meta = &Metadata{
		Hash:         v.etag,
		LastModified: v.lastModified,
	}

	if v.etagFn != nil {
		meta.Hash, err = v.etagFn(template, language)
		if err != nil {
			return meta, err
		}
	}

	if v.lastModFn != nil {
		meta.LastModified, err = v.lastModFn(template, language)
	}

	return meta, err
}

func (v *Value) Content(template, language string) (result interface{}, more bool, err error) {
	if v.supplier == nil {
		return v.value, false, nil
	}

	if v.chunked {
		return v.chunkedContent(template, language)
	}

	return v.lazyContent(template, language)
}

func (v *Value) lazyContent(template, language string) (result interface{}, more bool, err error) {
	r, err := v.supplier(template, language)
	return r, false, err
}

func (v *Value) chunkedContent(template, language string) (result interface{}, more bool, err error) {
	if v.value != nil {
		result = v.value
		v.value, err = v.supplier(template, language)
		return result, v.value != nil, err
	}

	result, err = v.supplier(template, language)
	if result != nil {
		// lookahead
		v.value, err = v.supplier(template, language)
	}

	return result, result != nil && v.value != nil, err
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

// ETag sets the entity tag for the content. This allows for conditional requests, possibly
// avoiding network traffic. This is not necessary if Lazy was used and the function
// returns metadata.
func (v Value) ETag(hash string) *Value {
	v.etag = hash
	return &v
}

// LastModified sets the time at which the content was last modified. This allows for conditional
// requests, possibly avoiding network traffic, although ETag takes precedence. This is not
// necessary if Lazy was used and the function returns metadata.
func (v Value) LastModified(at time.Time) *Value {
	v.lastModified = at
	return &v
}

// ETag lazily sets the entity tag for the content. This allows for conditional requests,
// possibly avoiding network traffic.
func (v Value) ETagUsing(fn func(template, language string) (string, error)) *Value {
	v.etagFn = fn
	return &v
}

// LastModifiedUsing lazily sets the time at which the content was last modified. This allows
// for conditional requests, possibly avoiding network traffic, although ETag takes precedence.
func (v Value) LastModifiedUsing(fn func(template, language string) (time.Time, error)) *Value {
	v.lastModFn = fn
	return &v
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

// ConditionalRequest checks the headers for conditional requests and returns a flag indicating whether
// content should be rendered or skipped.
//
// If the returned result value is false, the response has been set to 304-Not Modified, so the
// response processor does not need to do anything further.
//
// Data d must not be nil.
func ConditionalRequest(rw http.ResponseWriter, req *http.Request, d Data, template, language string) (sendContent bool, err error) {
	meta, err := d.Meta(template, language)
	if err != nil {
		return false, err
	}

	for hn, hv := range d.Headers() {
		rw.Header().Set(hn, hv)
	}

	if meta == nil || (req.Method != http.MethodGet && req.Method != http.MethodHead) {
		return true, nil
	}

	sendContent = true

	if meta.Hash != "" {
		rw.Header().Set("ETag", fmt.Sprintf("%q", meta.Hash))

		ifNoneMatch := header.ETagsOf(req.Header.Get(headername.IfNoneMatch))
		if ifNoneMatch.WeaklyMatches(meta.Hash) {
			rw.WriteHeader(http.StatusNotModified)
			sendContent = false
		}
	}

	if !meta.LastModified.IsZero() {
		rw.Header().Set("Last-Modified", meta.LastModified.Format(time.RFC1123))

		if sendContent {
			ifModifiedSince, e2 := time.Parse(time.RFC1123, req.Header.Get(headername.IfModifiedSince))
			if e2 == nil {
				if meta.LastModified.After(ifModifiedSince) {
					rw.WriteHeader(http.StatusNotModified)
					sendContent = false
				}
			}
		}
	}

	return sendContent, nil
}
