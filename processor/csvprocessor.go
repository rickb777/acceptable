package processor

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"reflect"

	"github.com/rickb777/acceptable"
	"github.com/rickb777/acceptable/data"
)

const TextCsv = "text/csv"

// DefaultCSVOffer is an Offer for text/plain content using the CSV() processor.
var DefaultCSVOffer = acceptable.OfferOf(TextCsv).Using(CSV())

// CSV creates an output processor that serialises a dataModel in CSV form. With no arguments, the default
// format is comma-separated; you can supply any rune to be used as an alternative separator.
//
// Model values should be one of the following:
//
// * string or []string, or [][]string
//
// * fmt.Stringer or []fmt.Stringer, or [][]fmt.Stringer
//
// * []int or similar (bool, int8, int16, int32, int64, uint8, uint16, uint32, uint63, float32, float64, complex)
//
// * [][]int or similar (bool, int8, int16, int32, int64, uint8, uint16, uint32, uint63, float32, float64, complex)
//
// * struct for some struct in which all the fields are exported and of simple types (as above).
//
// * []struct for some struct in which all the fields are exported and of simple types (as above).
func CSV(comma ...rune) acceptable.Processor {

	in := ','
	if len(comma) > 0 {
		in = comma[0]
	}

	return func(rw http.ResponseWriter, match acceptable.Match, template string) (err error) {
		w := match.ApplyHeaders(rw)

		d, err := data.GetContentAndApplyExtraHeaders(rw, match.Data, template, match.Language)
		if err != nil || d == nil {
			return err
		}

		writer := csv.NewWriter(w)
		writer.Comma = in

		err = writeCSV(writer, d)
		if err != nil {
			return err
		}

		writer.Flush()
		return writer.Error()
	}
}

func writeCSV(writer *csv.Writer, data interface{}) error {
	debug("csvProcessor.process %T\n", data)

	switch v := data.(type) {
	case string:
		return writer.Write([]string{v})
	case []string:
		return writer.Write(v)
	case [][]string:
		return writer.WriteAll(v)
	}

	value := reflect.Indirect(reflect.ValueOf(data))
	debug("  is %v\n", value.Kind())

	switch value.Kind() {
	case reflect.Struct:
		if value.NumField() == 0 {
			return nil // nothing to write
		}

		return writeStructFields(writer, value, data)

	case reflect.Array, reflect.Slice:
		if value.Len() == 0 {
			return nil // nothing to write
		}

		v0 := reflect.Indirect(value.Index(0))
		k0 := v0.Kind()

		if reflect.Bool <= k0 && k0 <= reflect.Complex128 {
			debug("    -- containing scalars\n")
			return writeArrayOfScalars(writer, value)
		}

		switch k0 {
		case reflect.Struct:
			if v0.NumField() == 0 {
				return nil // nothing to write
			}
			debug("    -- v0 is Struct\n")

			_, ok := v0.Interface().(fmt.Stringer)
			if ok {
				return writeArrayOfStringers(writer, value)
			}

			return writeArrayOfStructFields(writer, value, data)

		case reflect.Array, reflect.Slice:
			if v0.Len() == 0 {
				return nil // nothing to write
			}
			debug("    -- v0 is Array/Slice\n")

			v00 := reflect.Indirect(v0.Index(0))
			k00 := v00.Kind()
			debug("      -- v00 is %v\n", k00)

			if reflect.Bool <= k00 && k00 <= reflect.Complex128 {
				return write2DArrayOfScalars(writer, value)

			} else if k00 == reflect.Struct {
				_, ok := v00.Interface().(fmt.Stringer)
				if ok {
					return write2DArrayOfStringers(writer, value)
				}
			}

		default:
			debug("  -- v0 is %v -- ignored\n", k0)
		}
	}

	s, ok := data.(fmt.Stringer)
	if ok {
		return writer.Write([]string{s.String()})
	}

	return fmt.Errorf("Unsupported type for CSV: %T", data)
}

func writeArrayOfStructFields(writer *csv.Writer, value reflect.Value, dataModel interface{}) error {
	for j := 0; j < value.Len(); j++ {
		err := writeStructFields(writer, reflect.Indirect(value.Index(j)), dataModel)
		if err != nil {
			return err
		}
	}
	return nil
}

func writeStructFields(writer *csv.Writer, str reflect.Value, dataModel interface{}) error {
	sa := make([]string, str.NumField())
	for i := 0; i < str.NumField(); i++ {
		sa[i] = fmt.Sprintf("%v", reflect.Indirect(str.Field(i)))
	}
	return writer.Write(sa)
}

func write2DArrayOfStringers(writer *csv.Writer, value reflect.Value) error {
	debug("        -- write2DArrayOfStringers %d\n", value.Len())
	for j := 0; j < value.Len(); j++ {
		err := writeArrayOfStringers(writer, reflect.Indirect(value.Index(j)))
		if err != nil {
			return err
		}
	}
	return nil
}

func writeArrayOfStringers(writer *csv.Writer, value reflect.Value) error {
	debug("        -- writeArrayOfStringers %d\n", value.Len())
	sa := make([]string, value.Len())
	for i := 0; i < value.Len(); i++ {
		sa[i] = fmt.Sprintf("%v", reflect.Indirect(value.Index(i)).Interface().(fmt.Stringer))
	}
	return writer.Write(sa)
}

func write2DArrayOfScalars(writer *csv.Writer, value reflect.Value) error {
	for j := 0; j < value.Len(); j++ {
		err := writeArrayOfScalars(writer, reflect.Indirect(value.Index(j)))
		if err != nil {
			return err
		}
	}
	return nil
}

func writeArrayOfScalars(writer *csv.Writer, vj reflect.Value) error {
	sa := make([]string, vj.Len())
	for i := 0; i < vj.Len(); i++ {
		sa[i] = fmt.Sprintf("%v", reflect.Indirect(vj.Index(i)))
	}
	return writer.Write(sa)
}

var debug = func(msg string, args ...interface{}) {}

//var debug = fmt.Printf
