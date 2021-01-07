// Package processor contains flexible implementations for rendering JSON, XML, CSV etc.
// A Render is defined as
//
//    type Render func(w http.ResponseWriter, match Match, template string, dataModel interface{}) error
package processor
