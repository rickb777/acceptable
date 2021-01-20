package data

import (
	"fmt"
	"net/http"
	"time"

	"github.com/rickb777/acceptable/header"
)

// Data provides a source for response content. It is optimised for lazy evaluation, avoiding
// wasted processing.
//
// When preparing to render, Content will be called once or twice. The first time, the
// dataRequired flag is false; this gives an opportunity to obtain the entity tag
// with or without the data. At this stage, data may be returned only if it is convenient.
//
// If necessary, Content will be called a second time, this time with dataRequired=true. The
// data must always be returned in this case. However the metadata will be ignored.
//
// The metadata can be nil if not needed.
type Data interface {
	// Content returns the data as a value that can be processed by encoders such as "encoding/json"
	// The returned values are the data itself, a hash that will be used as the entity tag (if required),
	// and an error if arising.
	Content(template, language string, dataRequired bool) (data interface{}, meta *Metadata, err error)

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

// Lazy wraps a function that supplies a data value, but only fetches te data when it is needed.
//
// If an entity tag is already known, the ETag method should be used. Likewise, if a last-modified
// timestamp is already known, the LastModified method should also be used. Otherwise, metadata
// can be returned by the supplier function.
func Lazy(fn func(template, language string, dataRequired bool) (interface{}, *Metadata, error)) *Value {
	return &Value{supplier: fn}
}

// Value is a simple implementation of Data.
type Value struct {
	supplier func(template, language string, dataRequired bool) (interface{}, *Metadata, error)
	value    interface{}
	meta     *Metadata
	hdrs     map[string]string
}

func (v *Value) Content(template, language string, dataRequired bool) (result interface{}, meta *Metadata, err error) {
	if v.value != nil {
		return v.value, v.meta, nil
	}

	if v.supplier != nil {
		oldMeta := v.meta

		v.value, v.meta, err = v.supplier(template, language, dataRequired)

		// preserve the oldMeta values unless they were overwritten
		if v.meta == nil {
			v.meta = oldMeta
		} else if oldMeta != nil {
			if v.meta.Hash == "" {
				v.meta.Hash = oldMeta.Hash
			}
			if v.meta.LastModified.IsZero() {
				v.meta.LastModified = oldMeta.LastModified
			}
		}
	}

	return v.value, v.meta, err
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
	if v.meta == nil {
		v.meta = &Metadata{}
	}
	v.meta.Hash = hash
	return &v //.With("Last-Modified", at.Format(time.RFC1123))
}

// LastModified sets the time at which the content was last modified. This allows for conditional
// requests, possibly avoiding network traffic, although ETag takes precedence. This is not
// necessary if Lazy was used and the function returns metadata.
func (v Value) LastModified(at time.Time) *Value {
	if v.meta == nil {
		v.meta = &Metadata{}
	}
	v.meta.LastModified = at
	return &v //.With("Last-Modified", at.Format(time.RFC1123))
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
//
// Along with Match.ApplyHeaders, this function handles the response preparation needed by response
// processors (e.g. acceptable.JSON).
//
// If the returned result value is nil, the response has been set to 304-Not Modified, so the
// response processor does not need to do anything further.
func GetContentAndApplyExtraHeaders(rw http.ResponseWriter, req *http.Request, d Data, template, language string) (interface{}, error) {
	if d == nil {
		return nil, nil
	}

	v, meta, err := d.Content(template, language, false)
	if err != nil {
		return nil, err
	}

	for hn, hv := range d.Headers() {
		rw.Header().Set(hn, hv)
	}

	isGetOrHeadMethod := req.Method == http.MethodGet || req.Method == http.MethodHead

	sendContent := true

	if isGetOrHeadMethod && meta != nil && meta.Hash != "" {
		rw.Header().Set("ETag", fmt.Sprintf("%q", meta.Hash))

		ifNoneMatch := header.ETagsOf(req.Header.Get(header.IfNoneMatch))
		if ifNoneMatch.WeaklyMatches(meta.Hash) {
			rw.WriteHeader(http.StatusNotModified)
			sendContent = false
			v = nil
		}
	}

	if isGetOrHeadMethod && meta != nil && !meta.LastModified.IsZero() {
		rw.Header().Set("Last-Modified", meta.LastModified.Format(time.RFC1123))

		if sendContent {
			ifModifiedSince, e2 := time.Parse(time.RFC1123, req.Header.Get(header.IfModifiedSince))
			if e2 == nil {
				if meta.LastModified.After(ifModifiedSince) {
					rw.WriteHeader(http.StatusNotModified)
					sendContent = false
					v = nil
				}
			}
		}
	}

	if sendContent && v == nil {
		v, _, err = d.Content(template, language, true)
	}

	return v, err
}
