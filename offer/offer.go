package offer

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/rickb777/acceptable/data"
	"github.com/rickb777/acceptable/header"
	"github.com/rickb777/acceptable/internal"
)

// Processor is a function that renders content according to the matched result.
type Processor func(w http.ResponseWriter, req *http.Request, match Match, template string) error

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

	// data is an optional response to be rendered if this offer is selected.
	data map[string]data.Data
}

// Of constructs an Offer easily, given a content type.
// If the content type is blank, it is assumed to be the full wildcard "*/*".
// Also, contentType can be a partial wildcard "type/*".
func Of(processor Processor, contentType string) Offer {
	t, s := internal.Split1(contentType, '/')
	ct := header.ContentTypeOf(t, s)

	return Offer{
		ContentType: ct,
		processor:   processor,
		Langs:       []string{"*"},
		data:        make(map[string]data.Data),
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
// The language parameter specifies what language (or language group) the offer
// will match. It can be "*" to match any. The method panics if it is blank.
// Other languages can also be specified, but these must not be "*" (or blank).
//
// The data can be a value (struct, slice, etc) or a data.Data. It may also be
// nil, which means the method merely serves to add the language to the Offer's
// supported languages.
//
// The original offer is unchanged.
func (o Offer) With(d interface{}, language string, otherLanguages ...string) Offer {
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

	if d == nil {
		if language == "*" {
			return o
		}
		d = emptyValue
	}

	c := o.clone()

	// clear pre-existing wildcard
	if len(c.data) == 0 && len(c.Langs) == 1 && c.Langs[0] == "*" {
		c.Langs = nil
	}

	c.Langs = append(c.Langs, language)

	if s, ok := d.(data.Data); ok {
		c.data[language] = s
	} else {
		c.data[language] = data.Of(d)
	}

	for _, l := range otherLanguages {
		c.Langs = append(c.Langs, l)

		if s, ok := d.(data.Data); ok {
			c.data[l] = s
		} else {
			c.data[l] = data.Of(d)
		}
	}
	return c
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
func (o Offer) BuildMatch(lang string, acceptedCT header.MediaRange) *Match {
	t, s := o.resolvedType(acceptedCT)

	return &Match{
		Type:     t,
		Subtype:  s,
		Language: lang,
		Data:     o.Data(lang),
		Render:   o.processor,
	}
}

func (o Offer) resolvedType(acceptedCT header.MediaRange) (string, string) {
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

	return t, s
}

func (o Offer) Data(lang string) data.Data {
	d := o.data[lang]
	if d == emptyValue {
		d = nil
	}
	return d
}

//-------------------------------------------------------------------------------------------------

type empty struct{}

func (e empty) Content(string, string, bool) (interface{}, *data.Metadata, error) {
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
