// package data provides wrappers for response data, optionally including response headers
// such as ETag and Cache-Control. Type Data provides the means to wrap data with its metadata
// and to obtain these lazily when required. When the response data is provided lazily, this
// can be either a single item or a sequence of items. In both cases, a supplier function
// provided by the caller is used.
package data
