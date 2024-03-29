// Package acceptable is a library that handles headers for content negotiation and conditional requests in
// web applications written in Go.
// Content negotiation is specified by RFC (http://tools.ietf.org/html/rfc7231) and, less formally, by
// Ajax (https://en.wikipedia.org/wiki/XMLHttpRequest).
//
// # Subpackages
//
// * contenttype, headername - bundles of useful constants
//
// * data - for holding response data & metadata prior to rendering the response, also allowing lazy evaluation
//
// * header - for parsing and representing certain HTTP headers
//
// * offer - for enumerating offers to be matched against requests
//
// * templates - for rendering Go templates
//
// # Easy content negotiation
//
// Server-based content negotiation is essentially simple: the user agent sends a request including some preferences
// (accept headers), then the server selects one of several possible ways of sending the response. Finding the best
// match depends on you listing your available response representations. This is all rolled up into a simple-to-use
// function `acceptable.RenderBestMatch`. What this does is described in detail in
// [RFC-7231](https://tools.ietf.org/html/rfc7231#section-5.3), but it's easy to use in practice.
//
// For example
//
//	en := ... obtain some content in English
//	fr := ... obtain some content in French
//
//	// a JSON offer with pretty output
//	offer1 := offer.JSON("  ").With(en, "en").With(fr, "fr")
//
//	// an XML offer
//	offer2 := offer.XML().With(en, "en").With(fr, "fr")
//
//	// a catch-all offer is optional
//	catchAll := offer.Of(acceptable.TXTProcessor(), contenttype.Any).With(en, "en").With(fr, "fr")
//
//	err := acceptable.RenderBestMatch(response, request, 200, "", offer1, offer2, catchAll)
//
// The RenderBestMatch function searches for the offer that best matches the request headers. If none match,
// the response will be 406-Not Acceptable. If you need to have a catch-all case, include
// offer.Of(p, contenttype.TextAny) or offer.Of(p, contenttype.Any) last in the list.
// Note that contenttype.TextAny is "text/*" and will typically return "text/plain"; contenttype.Any is "*/*"
// and will likewise return "application/octet-stream".
//
// Each offer will (usually) have a suitable offer.Processor, which is a rendering function. Several are
// provided (for JSON, XML etc), but you can also provide your own.
//
// Also, the templates sub-package provides Go template support.
//
// Offers are restricted both by content-type matching and by language matching. The `With` method provides data
// and specifies its content language. Use it as many times as you need to.
//
// The language(s) is matched against the Accept-Language header using the basic prefix algorithm. This means
// for example that if you specify "en" it will match "en", "en-GB" and everything else beginning with "en-", but
// if you specify "en-GB", it only matches "en-GB" and "en-GB-*", but won't match "en-US" or even "en".
// (This implements the basic filtering language matching algorithm defined in https://tools.ietf.org/html/rfc4647.)
//
// If your data doesn't need to specify a language, the With method should simply use the "*" wildcard instead. For
// example, myOffer.With(data, "*") attaches data to myOffer and doesn't restrict the offer to any particular language.
//
// The language wildcard could also be used as a catch-all case if it comes after one or more With with a specified
// language. However, the standard (RFC-7231) advises that a response should be returned even when language matching has
// failed; RenderBestMatch will do this by picking the first language listed as a fallback, so the catch-all case
// is only necessary if its data is different to that of the first case.
//
// # Providing response data
//
// The response data (en and fr above) can be structs, slices, maps, or other values that the rendering processors
// accept. They will be wrapped as data.Data values, which you can provid explicitly. These allow for lazy
// evaluation of the content and also support conditional requests. This comes into its own when there are
// several offers each with their own data model - if these were all to be read from the database before selection
// of the best match, all but one would be wasted. Lazy evaluation of the selected data easily overcomes this problem.
//
//	en := data.Lazy(func(template, language string) (value interface{}, err error) {
//	    return ...
//	})
//
// Besides the data and error returned values, some metadata can optionally be returned. This is the basis for easy
// support for conditional requests (see [RFC-7232](https://tools.ietf.org/html/rfc7232)).
//
// If the metadata is nil, it is simply ignored. However, if it contains a hash of the data (e.g. via MD5) known as the
// entity tag or etag, then the response will have an ETag header. User agents that recognise this will later repeat
// the request along with an If-None-Match header. If present, If-None-Match is recognised before rendering starts
// and a successful match will avoid the need for any rendering. Due to the lazy content fetching, it can reduce
// unnecessary database traffic etc.
//
// The metadata can also carry the last-modified timestamp of the data, if this is known. When present, this becomes the
// Last-Modified header and is checked on subsequent requests using the If-Modified-Since.
//
// The template and language parameters are used for templated/web content data; otherwise they are ignored.
//
// Sequences of data can also be produced. This is done with data.Sequence() and this takes the same supplier function
// as used by data.Lazy(). The difference is that, in a sequence, the supplier function will be called repeatedly
// until its result value is nil. All the values will be streamed in the response (how this is done depends on
// the rendering processor.
//
// # Character set transcoding
//
// Most responses will be UTF-8, sometimes UTF-16. All other character sets (e.g. Windows-1252) are now strongly deprecated.
//
// However, legacy support for other character sets is provided. Transcoding is implemented by Match.ApplyHeaders so
// that the Accept-Charset content negotiation can be implemented. This depends on finding an encoder in
// golang.org/x/text/encoding/htmlindex (this has an extensive list, however no other encoders are supported).
//
// Whenever possible, responses will be UTF-8. Not only is this strongly recommended, it also avoids any transcoding
// processing overhead. It means for example that "Accept-Charset: iso-8859-1, utf-8" will ignore the iso-8859-1
// preference because it can use UTF-8. Conversely, "Accept-Charset: iso-8859-1" will always have to transcode into
// ISO-8859-1 because there is no UTF-8 option.
package acceptable
