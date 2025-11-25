package contenttype

const (
	Any = "*/*"

	TextAny   = "text/*"
	TextHTML  = "text/html"
	TextCSV   = "text/csv"
	TextPlain = "text/plain"

	ApplicationAny = "application/*"

	// application MIME types are sent without charset (since RFC-7231 - see Appendix B)

	ApplicationJSON   = "application/json"
	ApplicationPDF    = "application/pdf"
	ApplicationXML    = "application/xml"
	ApplicationXHTML  = "application/xhtml+xml"
	ApplicationBinary = "application/octet-stream"

	// ApplicationForm is for POSTed forms. If you have binary (non-alphanumeric) data
	// (or a significantly sized payload) to transmit, use multipart/form-data. Otherwise,
	// use application/x-www-form-urlencoded.
	ApplicationForm   = "application/x-www-form-urlencoded"
	MultipartFormData = "multipart/form-data"

	ImageAny  = "image/*"
	ImageJPEG = "image/jpeg"
	ImageGIF  = "image/gif"
	ImagePNG  = "image/png"
	ImageSVG  = "image/svg+xml"

	CharsetUTF8 = "charset=utf-8"
)
