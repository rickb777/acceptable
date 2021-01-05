// Package processor contains flexible implementations for rendering JSON, XML, CSV etc.
// A Processor is defined as
//
//    type Processor func(w http.ResponseWriter, match Match, template string, dataModel interface{}) error
package processor
