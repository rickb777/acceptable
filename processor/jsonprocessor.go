package processor

import (
	"encoding/json"
	"net/http"

	"github.com/rickb777/acceptable"
)

const defaultJSONContentType = "application/json; charset=utf-8"

// JSON creates a new processor for JSON with a specified indentation.
// It handles all requests except Ajax requests.
func JSON(indent ...string) acceptable.Processor {
	if len(indent) == 0 || len(indent[0]) == 0 {
		return func(w http.ResponseWriter, match acceptable.Match, template string, dataModel interface{}) error {
			match.ApplyHeaders(w)

			return json.NewEncoder(w).Encode(dataModel)
		}
	}
	return func(w http.ResponseWriter, match acceptable.Match, template string, dataModel interface{}) error {
		match.ApplyHeaders(w)

		js, err := json.MarshalIndent(dataModel, "", indent[0])
		if err != nil {
			return err
		}

		return WriteWithNewline(w, js)
	}
}
