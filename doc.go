// Package acceptable is a library that handles headers for content negotiation in web applications written in Go.
// Content negotiation is specified by RFC (http://tools.ietf.org/html/rfc7231) and, less formally, by
// Ajax (https://en.wikipedia.org/wiki/XMLHttpRequest).
//
// Accept
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
// Accept-Language
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
package acceptable
