package data

import (
	"fmt"
	"net/http"
	"time"

	"github.com/rickb777/acceptable/header"
	. "github.com/rickb777/acceptable/headername"
)

type Chosen struct {
	Template string
	Language string
}

// A Supplier supplies data.
type Supplier func(params Chosen) (any, error)

// Data provides a source for response content. It is optimised for lazy evaluation, avoiding
// wasted processing as much as possible.
type Data interface {
	// Meta returns the metadata that will be used to set response headers automatically.
	// The headers are ETag and Last-Modified.
	Meta(params Chosen) (meta *Metadata, err error)

	// Content returns the data as a value that can be processed by encoders such as "encoding/json"
	// The returned values are
	//   - the data itself,
	//   - a boolean that is true if the data is in chunks and there is more data to follow, and
	//   - an error if one occurs.
	// For chunked data, this method will be called repeatedly until the boolean yields false
	// or an error arises.
	Content(params Chosen) (any, bool, error)

	// Headers returns response headers relating to the data (optional)
	Headers() map[string]string
}

// Metadata provides optional entity tag and last modified information about some data. This
// can be sent with a response such thath the client can make conditional requests in future.
type Metadata struct {
	Hash         string    // used as entity tag; blank if not required
	LastModified time.Time // used for Last-Modified header; zero if not required
}

// Of wraps a data value.
//
// If an entity tag is known, the [Value.ETag] method should be used on the result. Likewise,
// if a last-modified timestamp is known, the [Value.LastModified] method should also be used.
func Of(v any) *Value {
	return Lazy(func(Chosen) (any, error) { return v, nil })
}

// Lazy wraps a function that supplies a data value, but only fetches the data when it is needed.
//
// If an entity tag is known, the [Value.ETag] method should be used on the result. Likewise,
// if a last-modified timestamp is known, the [Value.LastModified] method should also be used.
func Lazy(supplier Supplier) *Value {
	return &Value{supplier: supplier, chunked: false}
}

// Sequence wraps a function that supplies data values in a sequence chunk by chunk. This
// function will not be not called until it is needed. When it is called, it will be called
// repeatedly until the returned value is nil or an error arises.
//
// Typical use might be where a response contains many database records that are obtained
// one by one to avoid the need to cache all results in memory before rendering.
//
// If an entity tag is known, the [Value.ETag] method should be used on the result. Likewise,
// if a last-modified timestamp is known, the [Value.LastModified] method should also be used.
func Sequence(supplier Supplier) *Value {
	return &Value{supplier: supplier, chunked: true}
}

//-------------------------------------------------------------------------------------------------

// Value is a simple implementation of Data.
type Value struct {
	supplier     Supplier
	chunked      bool
	next         any // used for sequence behaviour
	etagFn       func(params Chosen) (string, error)
	lastModFn    func(params Chosen) (time.Time, error)
	etag         string
	lastModified time.Time
	hdrs         map[string]string
}

func (v *Value) Meta(params Chosen) (meta *Metadata, err error) {
	meta = &Metadata{
		Hash:         v.etag,
		LastModified: v.lastModified,
	}

	if v.etagFn != nil {
		meta.Hash, err = v.etagFn(params)
		if err != nil {
			return meta, err
		}
	}

	if v.lastModFn != nil {
		meta.LastModified, err = v.lastModFn(params)
	}

	return meta, err
}

func (v *Value) Content(params Chosen) (result any, more bool, err error) {
	if v.chunked {
		return v.chunkedContent(params)
	}

	r, err := v.supplier(params)
	return r, false, err
}

func (v *Value) chunkedContent(params Chosen) (result any, more bool, err error) {
	if v.next != nil {
		result = v.next
		v.next, err = v.supplier(params)
		return result, v.next != nil, err
	}

	result, err = v.supplier(params)
	if result != nil {
		// lookahead
		v.next, err = v.supplier(params)
	}

	return result, result != nil && v.next != nil, err
}

func (v Value) Headers() map[string]string {
	return v.hdrs
}

// With returns a copy of v with extra headers attached. These are passed in as key+value pairs.
// The header names should be in normal form, e.g. "Last-Modified" instead of "last-modified",
// but this is not mandatory. The values are simple strings, numbers etc.
// The others contain more key/value pairs; there should be an even number of them.
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
// requests, possibly avoiding network traffic, although [Value.ETag] takes precedence. This is not
// necessary if Lazy was used and the function returns metadata.
func (v Value) LastModified(at time.Time) *Value {
	v.lastModified = at
	return &v
}

// ETagUsing lazily sets the entity tag for the content. This allows for conditional requests,
// possibly avoiding some network traffic.
func (v Value) ETagUsing(fn func(params Chosen) (string, error)) *Value {
	v.etagFn = fn
	return &v
}

// LastModifiedUsing lazily sets the time at which the content was last modified. This allows
// for conditional requests, possibly avoiding network traffic, although ETag takes precedence.
func (v Value) LastModifiedUsing(fn func(params Chosen) (time.Time, error)) *Value {
	v.lastModFn = fn
	return &v
}

// Expires sets the time at which the response becomes stale. MaxAge takes precedence.
func (v Value) Expires(at time.Time) *Value {
	return v.With(Expires, header.FormatHTTPDateTime(at))
}

// MaxAge sets the max-age header on the response. This is used to allow caches to avoid repeating
// the request until the max age has expired, after which time the resource is considered stale.
func (v Value) MaxAge(max time.Duration) *Value {
	return v.With(CacheControl, fmt.Sprintf("max-age=%d", max/time.Second))
}

// NoCache sets cache control headers to prevent the response being cached.
func (v Value) NoCache() *Value {
	return v.With(CacheControl, "no-cache, must-revalidate", Pragma, "no-cache")
}

// ConditionalRequest checks the headers for conditional requests and returns a flag indicating whether
// content should be rendered or skipped.
//
// If the returned result value is false, the response has been set to 304-Not Modified, so the
// response processor does not need to do anything further.
//
// Data d must not be nil.
func ConditionalRequest(rw http.ResponseWriter, req *http.Request, d Data, params Chosen) (sendContent bool, err error) {
	meta, err := d.Meta(params)
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
		rw.Header().Set(ETag, fmt.Sprintf("%q", meta.Hash))

		ifNoneMatch := header.ETagsOf(req.Header.Get(IfNoneMatch))
		if ifNoneMatch.WeaklyMatches(meta.Hash) {
			rw.WriteHeader(http.StatusNotModified)
			sendContent = false
		}
	}

	if !meta.LastModified.IsZero() {
		rw.Header().Set(LastModified, header.FormatHTTPDateTime(meta.LastModified))

		if sendContent {
			ifModifiedSince, e2 := header.ParseHTTPDateTime(req.Header.Get(IfModifiedSince))
			if e2 == nil && !ifModifiedSince.IsZero() {
				if meta.LastModified.After(ifModifiedSince) {
					rw.WriteHeader(http.StatusNotModified)
					sendContent = false
				}
			}
		}
	}

	return sendContent, nil
}
