# Acceptable 

[![GoDoc](https://img.shields.io/badge/api-Godoc-blue.svg)](http://pkg.go.dev/github.com/rickb777/acceptable)
[![Build Status](https://travis-ci.org/rickb777/acceptable.svg?branch=master)](https://travis-ci.org/rickb777/acceptable/builds)
[![Issues](https://img.shields.io/github/issues/rickb777/acceptable.svg)](https://github.com/rickb777/acceptable/issues)

This is a library that handles `Accept` headers, which form the basis of content negotiation in HTTP server applications written in Go. It provides an implementation of the content negotiation algorithm specified in RFC-7231.

## Accept parsing

The `Accept` header is parsed using `ParseMediaRanges(hdr)`, which returns the slice of media ranges, e.g.

```go
    // handle Accept-Language
    mediaRanges := header.ParseMediaRanges("application/json;q=0.8, application/xml, application/*;q=0.1")
```

The resulting slice is sorted according to precedence and quality rules, so in this example the order is `{"application/xml", "application/json", "application/*"}` because the middle item has an implied quality of 1, whereas the first item has a lower quality.

## Accept-Language and Accept-Charset parsing

The other important content-negotiation headers, `Accept-Language` and `Accept-Charset`, are handled by the `header.Parse` method, e.g.

```go
    // handle Accept-Language
    acceptLanguages := header.Parse("en-GB,fr;q=0.5,en;q=0.8")
```

This will contain `{"en-GB", "en", "fr"}` in a `header.PrecedenceValues` slice, sorted according to precedence rules with the most preferred first.

The `acceptable.Parse` function can be used for `Accept-Encoding` as well as `Accept-Language` and `Accept-Charset`. However, the Go standard library deals with `Accept-Encoding`, so you won't need to.

## Putting it together - simple content negotiation

Finding the best match depends on you listing your available response representations. This is all rolled up into a simple-to-use function `acceptable.RenderBestMatch`

```go
    en := ... obtain some content in English
    fr := ... obtain some content in English

    // long-hand construction of an offer for indented JSON
    offer1 := acceptable.OfferOf("application/json").Using(processor.JSON("  ")).With("en", en).With("fr", fr)

    // short-hand construction of an XML offer
    offer2 := processor.DefaultXMLOffer.With("en", en).With("fr", fr)
    // equivalent to
    //offer2 := acceptable.OfferOf("application/xml").Using(processor.XML()).With("en", en).With("fr", fr)

    // a catch-all offer is optional
    catchAll := acceptable.OfferOf("*/*").Using(processor.TXT()).With("en", en).With("fr", fr)
    
    err := acceptable.RenderBestMatch(request, offer1, offer2, catchAll)
```

The best result will be the one that best matches the request headers. If none match, the response will be 406-Not Acceptable. If you need to have a catch-all case, include `acceptable.OfferOf("*/*")` last in the list.

The offers will usually hold a suitable rendering function. This is attached with the `Using` method. Sub-packages `processor` and `templates` provide useful renderers but you can also provide your own.

The offers can also be restricted by language matching. This is done either using `OfferOf` or with the `With` method, This matches `Accept-Language` using the basic prefix algorithm, which means for example that if you specify "en" it will match "en", "en-GB" and everything else beginning with "en-".

### Providing response data

The response data (`en` and `fr` above) can be structs, slices, maps, or other values. Alternatively they can be functions that return values. These functions are either

```go
    en := func() (interface{}, error) {
        return ...
    }
```

or for template/web content data

```go
    en := func(template, language string) (interface{}, error) {
        return ...
    }
```

and they can even be nested, returning another such function.

The selected response processor will render the actual response using the data provided, for example a struct will become JSON text if `processor.JSON` renders it.
