# Acceptable 

[![GoDoc](https://img.shields.io/badge/api-Godoc-blue.svg)](http://pkg.go.dev/github.com/rickb777/acceptable)
[![Go Report Card](https://goreportcard.com/badge/github.com/rickb777/acceptable)](https://goreportcard.com/report/github.com/rickb777/acceptable)
[![Build](https://github.com/rickb777/acceptable/actions/workflows/go.yml/badge.svg)](https://github.com/rickb777/acceptable/actions)
[![Coverage](https://coveralls.io/repos/github/rickb777/acceptable/badge.svg?branch=main)](https://coveralls.io/github/rickb777/acceptable?branch=main)
[![Issues](https://img.shields.io/github/issues/rickb777/acceptable.svg)](https://github.com/rickb777/acceptable/issues)

This is a library that handles `Accept` headers, which form the basis of content negotiation in HTTP server applications written in Go. It provides an implementation of the proactive server-driven content negotiation algorithm specified in [RFC-7231 section 5.3](https://tools.ietf.org/html/rfc7231#section-5.3).

There is also support for conditional requests ([RFC-7232](https://tools.ietf.org/html/rfc7232)) using entity tags and last-modified timestamps.

Bring your favourite router and framework - this library can be used with [Gin](https://github.com/gin-gonic/gin), [Echo](https://echo.labstack.com/), etc.

Please see the documentation for more info.

## Status

This API is well-tested and known to work but not yet fully released because it may yet require breaking API changes.
