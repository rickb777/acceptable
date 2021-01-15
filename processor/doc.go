// Package processor contains flexible implementations for rendering JSON, XML, CSV and plain text.
//
// An acceptable.Processor is defined as
//
//    type Processor func(w http.ResponseWriter, match Match, template string) error
package processor
