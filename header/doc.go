// Package header provides parsing rules for content negotiation & conditional requires headers according
// to RFC-7231 & RFC-7232.
//
// For "Accept-Language", "Accept-Encoding" or "Accept-Charset" use the ParsePrecedenceValues function.
//
// For "Accept" use the ParseMediaRanges function. This has more complex attributes and rules.
//
// For "If-None-Match" use the ETagsOf function (also useful for "If-Match").
//
// # Accept
//
// The Accept header is parsed using ParseMediaRanges(hdr), which returns the slice of media ranges, e.g.
//
//	// handle Accept-Language
//	mediaRanges := header.ParseMediaRanges("application/json;q=0.8, application/xml, application/*;q=0.1")
//
// The resulting slice is ready-sorted according to precedence and quality rules, so in this example the order
// is {"application/xml", "application/json", "application/*"} because the middle item has an implied
// quality of 1, whereas the first item has a lower quality.
//
// from https://tools.ietf.org/html/rfc7231#section-5.3.2:
//
// The "Accept" header field can be used by user agents to specify
// response media types that are acceptable.  Accept header fields can
// be used to indicate that the request is specifically limited to a
// small set of desired types, as in the case of a request for an
// in-line image.
//
// A request without any Accept header field implies that the user agent
// will accept any media type in response.
//
// If the header field is present in a request and none of the available
// representations for the response have a media type that is listed as
// acceptable, the origin server can either honor the header field by
// sending a 406 (Not Acceptable) response, or disregard the header field
// by treating the response as if it is not subject to content negotiation.
//
// Example header
//
//	Accept: audio/*; q=0.2, audio/basic
//
// # Accept-Language
//
// The other important content-negotiation headers, Accept-Language and Accept-Charset, are handled
// by the header.Parse method, e.g.
//
//	// handle Accept-Language
//	acceptLanguages := header.ParsePrecedenceValues("en-GB,fr;q=0.5,en;q=0.8")
//
// This will contain {"en-GB", "en", "fr"} in a header.PrecedenceValues slice, sorted according to
// precedence rules with the most preferred first.
//
// The acceptable.Parse function can be used for Accept-Encoding as well as Accept-Language and
// Accept-Charset. However, the Go standard library deals with Accept-Encoding, so you won't need to.
//
// from https://tools.ietf.org/html/rfc7231#section-5.3.5:
//
// The "Accept-Language" header field can be used by user agents to
// indicate the set of natural languages that are preferred in the
// response.
//
// A request without any Accept-Language header field implies that the
// user agent will accept any language in response.
//
// If the header field is present in a request and none of the available
// representations for the response have a matching language tag, the origin
// server can either disregard the header field by treating the response as if it
// is not subject to content negotiation or honor the header field by
// sending a 406 (Not Acceptable) response.  However, the latter is not
// encouraged, as doing so can prevent users from accessing content that
// they might be able to use (with translation software, for example).
//
// Example header
//
//	Accept-Language: da, en-gb;q=0.8, en;q=0.7
//
// # If-None-Match
//
// This header is used for conditional requests where large responses can be avoided
// when they are already present in caches. Its use is closely related to that of
// If-Modified-Since, which uses a timestamp (in RFC1123 format), whilst If-None-Match
// uses entity tags.
//
// from https://tools.ietf.org/html/rfc7232#section-3.2
//
// The "If-None-Match" header field makes the request method conditional
// on a recipient cache or origin server either not having any current
// representation of the target resource, when the field-value is "*",
// or having a selected representation with an entity-tag that does not
// match any of those listed in the field-value.
//
// Example header
//
//	If-None-Match: "xyzzy", "r2d2xxxx", "c3piozzzz"
package header
