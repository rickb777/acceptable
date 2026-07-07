package offer

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/rickb777/acceptable/contenttype"
	dpkg "github.com/rickb777/acceptable/data"
)

// ImageJPEG is an Offer for image/jpeg content using BinaryProcessor.
func ImageJPEG() Offer { return Of(BinaryProcessor(0), contenttype.ImageJPEG) }

// ImagePNG is an Offer for image/png content using BinaryProcessor.
func ImagePNG() Offer { return Of(BinaryProcessor(0), contenttype.ImagePNG) }

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
//
// GZIP compression-on-demand is enabled when gzipLevel is non-zero.
func BinaryProcessor(gzipLevel int) Processor {
	return GZIPProcessor(gzipLevel, binaryProcessor())
}

func binaryProcessor() Processor {
	return func(w io.Writer, _ *http.Request, data dpkg.Data, chosen dpkg.Chosen) (err error) {
		more := data != nil

		for more {
			var d any
			d, more, err = data.Content(chosen)
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
