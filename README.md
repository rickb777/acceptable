# Acceptable 

[![GoDoc](https://img.shields.io/badge/api-Godoc-blue.svg)](http://pkg.go.dev/github.com/rickb777/acceptable)
[![Build Status](https://travis-ci.org/rickb777/acceptable.svg?branch=master)](https://travis-ci.org/rickb777/acceptable/builds)
[![Issues](https://img.shields.io/github/issues/rickb777/acceptable.svg)](https://github.com/rickb777/acceptable/issues)

This is a library that handles `Accept` headers, which form the basis of content negotiation in HTTP server applications written in Go.

## Accept

The `Accept` header is parsed using `ParseMediaRanges(hdr)`, which returns the slice of media ranges, e.g.

```
    // handle Accept-Language
    mediaRanges := acceptable.ParseMediaRanges("application/json;q=0.8, application/xml, application/*;q=0.1")
```

The resulting slice is sorted according to precedence and quality rules, so in this example the order is `{"application/xml", "application/json", "application/*"}` because the middle item has an implied quality of 1, whereas the first item has a lower quality.

## Accept-Language and Accept-Charset

The other important content-negotiation headers, `Accept-Language` and `Accept-Charset`, are handled by the `header.Parse` method, e.g.

```
    // handle Accept-Language
    acceptLanguages := acceptable.Parse("en-GB,fr;q=0.5,en;q=0.8")
```

This will contain `{"en-GB", "en", "fr"}` in a `header.PrecedenceValues` slice, sorted according to precedence rules.

The `acceptable.Parse` function can be used for `Accept-Encoding` as well as `Accept-Language` and `Accept-Charset`. However, the Go standard library deals with `Accept-Encoding`, so you won't need to.
