// Package headername provides constants for HTTP header names.
package headername

const (
	Accept              = "Accept"
	AcceptCharset       = "Accept-Charset"
	AcceptEncoding      = "Accept-Encoding" // AcceptEncoding is handled effectively by net/http and can be disregarded here
	AcceptLanguage      = "Accept-Language"
	Allow               = "Allow"
	Authorization       = "Authorization"
	CacheControl        = "Cache-Control"
	ContentDisposition  = "Content-Disposition"
	ContentEncoding     = "Content-Encoding"
	ContentLanguage     = "Content-Language"
	ContentLength       = "Content-Length"
	ContentType         = "Content-Type"
	Cookie              = "Cookie" // Cookie and Set-Cookie are handled effectively by the standard library APIs
	ETag                = "ETag"
	Expires             = "Expires"
	IfModifiedSince     = "If-Modified-Since"
	IfNoneMatch         = "If-None-Match"
	LastModified        = "Last-Modified"
	Location            = "Location"
	Origin              = "Origin"
	Pragma              = "Pragma"
	Server              = "Server"
	SetCookie           = "Set-Cookie"
	Upgrade             = "Upgrade"
	UserAgent           = "User-Agent"
	Vary                = "Vary"
	WWWAuthenticate     = "WWW-Authenticate"
	XCorrelationID      = "X-Correlation-ID"
	XForwardedFor       = "X-Forwarded-For"
	XForwardedProto     = "X-Forwarded-Proto"
	XForwardedProtocol  = "X-Forwarded-Protocol"
	XForwardedSsl       = "X-Forwarded-Ssl"
	XHTTPMethodOverride = "X-HTTP-Method-Override"
	XRealIP             = "X-Real-IP"
	XRequestID          = "X-Request-ID"
	XRequestedWith      = "X-Requested-With" // XRequestedWith defines the header strings used for XHR.
	XUrlScheme          = "X-Url-Scheme"
)
