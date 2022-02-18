package offer

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/rickb777/acceptable/contenttype"
	datapkg "github.com/rickb777/acceptable/data"
)

// ImageJPEGPNG is an Offer for image/jpeg content using BinaryProcessor.
func ImageJPEG() Offer { return Of(BinaryProcessor(), contenttype.ImageJPEG) }

// ImagePNG is an Offer for image/png content using BinaryProcessor.
func ImagePNG() Offer { return Of(BinaryProcessor(), contenttype.ImagePNG) }

// BinaryProcessor creates an output processor that outputs binary data in a form suitable for image/* and similar responses.
// Model values should be one of the following:
//
// * []byte
// * io.WriterTo
// * io.Reader
// * nil
//
// Because it handles io.Reader and io.WriterTo, BinaryProcessor can be used to stream large responses (without any
// further encoding).
func BinaryProcessor() Processor {
	return func(w io.Writer, _ *http.Request, data datapkg.Data, template, language string) (err error) {
		more := data != nil

		for more {
			var d interface{}
			d, more, err = data.Content(template, language)
			if err != nil {
				return err
			}

			switch v := d.(type) {
			case []byte:
				_, err = io.Copy(w, bytes.NewBuffer(v))
			case io.WriterTo:
				_, err = v.WriteTo(w)
			case io.Reader:
				_, err = io.Copy(w, v)
			case nil:
				// no-op
			default:
				info := fmt.Sprintf("%T: unsupported binary data", d)
				panic(info)
			}
		}

		return err
	}
}
