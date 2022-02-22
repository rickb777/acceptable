package offer

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	datapkg "github.com/rickb777/acceptable/data"
	"github.com/rickb777/acceptable/header"
	"github.com/rickb777/acceptable/internal"
)

// Processor is a function that renders content according to the matched result.
type Processor func(w io.Writer, req *http.Request, data datapkg.Data, template, language string) error

// Offer holds information about one particular resource representation that can potentially
// provide an acceptable response.
type Offer struct {
	// ContentType is the content type that is to be matched.
	// Wildcard values may be used.
	header.ContentType

	// processor is an optional function you can use to apply the offer if it is selected.
	// How this is used is entirely at the discretion of the call site.
	processor Processor

	// Langs lists the language(s) provided by this offer.
	Langs []string

	// data has optional responses, keyed by language, to be rendered if this offer is selected.
	data map[string]datapkg.Data

	// Handle406As enables this offer to be a handler for any 406-not-acceptable case that arises.
	// Normally, this field will be left zero. However, if non-zero, the offer can be rendered
	// even when no acceptable match has been found. This overrides the acceptable.NoMatchAccepted
	// handler, providing a means to supply bespoke error responses.
	// The value will be the required status code (e.g. 400 for Bad Request).
	Handle406As int
}

// Of constructs an Offer easily, given a content type.
// The contentType can be a partial wildcard "type/*".
//
// Also, if the content type is blank, it is assumed to be the full wildcard "*/*".
// However, in this catch-all situation, the best matching MIME type will be determined from
// the Accept header and the response Content-Type will be set to this value, even if it is
// inappropriate for the actual content. Therefore this should be used sparingly or not at all.
// The correct behaviour is a 406 when no match can be made.
func Of(processor Processor, contentType string) Offer {
	return Offer{
		ContentType: header.ParseContentType(contentType),
		processor:   processor,
		Langs:       []string{"*"},
		data:        make(map[string]datapkg.Data),
	}
}

// clone makes a defensive deep copy of the original offer.
func (o Offer) clone() Offer {
	c := Of(o.processor, o.ContentType.String())

	c.Langs = make([]string, len(o.Langs))
	for i, s := range o.Langs {
		c.Langs[i] = s
	}

	for l, d := range o.data {
		c.data[l] = d
	}

	return c
}

// With attaches response data to an offer.
// The returned offer is a clone of the original offer, which is unchanged. This
// allows base offers to be derived from.
//
// The data can be a value (struct, slice, etc) or a data.Data. It may also be
// nil, which means the method merely serves to add the language to the Offer's
// supported languages.
//
// The language parameter specifies what language (or language group such as "en-GB")
// the data represents and that the offer will therefore match. It can be "*" to
// match every language.
//
// The method panics if language is blank. Other languages can also be specified, but these
// must not be "*" (or blank). Duplicates are not allowed.
//
// Language matching is described further in IETF BCP 47.
func (o Offer) With(data interface{}, language string, otherLanguages ...string) Offer {
	o.checkForBlanks(language, otherLanguages)
	o.checkForDuplicates(language, otherLanguages)

	if data == nil {
		if language == "*" {
			return o // no-op
		}
		data = emptyValue
	}

	c := o.clone()

	// clear pre-existing wildcard
	if len(c.data) == 0 && len(c.Langs) == 1 && c.Langs[0] == "*" {
		c.Langs = nil
	}

	c.Langs = append(c.Langs, language)

	if s, ok := data.(datapkg.Data); ok {
		c.data[language] = s
	} else {
		c.data[language] = datapkg.Of(data)
	}

	for _, l := range otherLanguages {
		c.Langs = append(c.Langs, l)

		if s, ok := data.(datapkg.Data); ok {
			c.data[l] = s
		} else {
			c.data[l] = datapkg.Of(data)
		}
	}
	return c
}

// 'With' parameters must be reasonable
func (o Offer) checkForBlanks(language string, otherLanguages []string) {
	if language == "" {
		panic("language must not be blank")
	}
	for i, l := range otherLanguages {
		if l == "" {
			panic(fmt.Sprintf("other language %d must not be blank", i))
		}
		if l == "*" {
			panic(fmt.Sprintf("other language %d must not be * wildcard", i))
		}
	}
}

// 'With' languages cannot duplicate earlier ones because that would break the
// invariant that o.Langs is in the order they were added
func (o Offer) checkForDuplicates(language string, otherLanguages []string) {
	if _, existsAlready := o.data[language]; existsAlready {
		panic(fmt.Sprintf("language %s is a duplicate", language))
	}
	for _, l := range otherLanguages {
		if _, existsAlready := o.data[l]; existsAlready {
			panic(fmt.Sprintf("other language %s is a duplicate", l))
		}
	}
}

// CanHandle406As sets the Handle406As status code.
func (o Offer) CanHandle406As(statusCode int) Offer {
	o.Handle406As = statusCode
	return o
}

// ToSlice returns the offer as a single-item slice.
func (o Offer) ToSlice() Offers {
	return Offers{o}
}

// String is merely for information purposes.
func (o Offer) String() string {
	buf := &strings.Builder{}
	buf.WriteString("Accept: ")
	buf.WriteString(o.ContentType.String())
	if len(o.data) > 0 {
		buf.WriteString(". Accept-Language: ")
		comma := ""
		for _, l := range o.Langs {
			buf.WriteString(comma)
			buf.WriteString(l)
			comma = ","
		}
	}
	return buf.String()
}

//-------------------------------------------------------------------------------------------------

// BuildMatch implements the transition between a selected Offer and the resulting Match.
// The result is based on the best-matched media type and language.
func (o Offer) BuildMatch(acceptedCT header.ContentType, lang string, statusCodeOverride int) *Match {
	resolved := o.resolvedType(acceptedCT)

	return &Match{
		ContentType:        resolved,
		Language:           lang,
		Data:               o.Data(lang),
		Render:             o.processor,
		StatusCodeOverride: statusCodeOverride,
	}
}

func (o Offer) BuildFallbackMatch() *Match {
	return o.BuildMatch(o.ContentType, o.Langs[0], o.Handle406As)
}

func (o Offer) resolvedType(acceptedCT header.ContentType) header.ContentType {
	t := o.Type
	s := o.Subtype

	if o.Subtype == "*" && acceptedCT.Subtype != "*" {
		s = acceptedCT.Subtype
		if o.Type == "*" && acceptedCT.Type != "*" {
			t = acceptedCT.Type
		}
	}

	if t == "text" && s == "*" {
		s = "plain"
	} else if t == "*" || s == "*" {
		t = "application"
		s = "octet-stream"
		// Ideally this should choose text/plain when the content is purely textual,
		// allowing for the encoding of the selected character set. This is hard to do
		// without knowledge of the response content; the standard library sniffs the
		// first 512 bytes but there is no attempt to do that here.
	}

	return header.ContentType{Type: t, Subtype: s}
}

func (o Offer) Data(lang string) datapkg.Data {
	d := emptyToNil(o.data[lang])

	// When the only data matches the wildcard "*", that should be the
	// result for all languages.
	if d == nil && len(o.data) == 1 {
		if d2, exists := o.data["*"]; exists {
			return emptyToNil(d2)
		}
	}

	return d
}

func emptyToNil(d datapkg.Data) datapkg.Data {
	if d == emptyValue {
		return nil
	}
	return d
}

//-------------------------------------------------------------------------------------------------

type empty struct{}

func (e empty) Meta(_, _ string) (*datapkg.Metadata, error) {
	panic("not reachable")
}

func (e empty) Content(_, _ string) (interface{}, bool, error) {
	panic("not reachable")
}

func (e empty) Headers() map[string]string {
	panic("not reachable")
}

var emptyValue = empty{}

//-------------------------------------------------------------------------------------------------

// Offers holds a slice of Offer.
type Offers []Offer

// Filter returns only the offers that match specified type and subtype.
// The type and subtype parameters can be a wildcard, "*".
func (offers Offers) Filter(typ, subtype string) Offers {
	if len(offers) == 0 {
		return nil
	}

	allowed := make(Offers, 0, len(offers))
	for _, mr := range offers {
		if internal.EqualOrWildcard(mr.Type, typ) && internal.EqualOrWildcard(mr.Subtype, subtype) {
			allowed = append(allowed, mr)
		}
	}

	return allowed
}

// CanHandle406 filters offers to retunr only those with non-zero status codes in the
// Handle406As field.
func (offers Offers) CanHandle406() Offers {
	if len(offers) == 0 {
		return nil
	}

	n := 0
	for _, mr := range offers {
		if mr.Handle406As != 0 {
			n++
		}
	}

	allowed := make(Offers, 0, n)
	for _, mr := range offers {
		if mr.Handle406As != 0 {
			allowed = append(allowed, mr)
		}
	}

	return allowed
}
