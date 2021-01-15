package acceptable

import (
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

	langs []string

	// data is an optional response to be rendered if this offer is selected.
	data map[string]data.Data
}

// OfferOf constructs an Offer easily. Zero or more languages can be specified.
// If the language is absent, it is assumed to be the wildcard "*".
// If the content type is blank, it is assumed to be the full wildcard "*/*".
// Also, contentType can be a partial wildcard "type/*".
func OfferOf(contentType string, language ...string) Offer {
	t, s := "*", "*"
	if contentType != "" {
		t, s = internal.Split1(contentType, '/')
	}

	ct := header.ContentType{
		Type:    t,
		Subtype: s,
	}

	offer := Offer{
		ContentType: ct,
		data:        make(map[string]data.Data),
	}

	if len(language) == 0 {
		offer.langs = []string{"*"}
		return offer
	}

	offer.langs = language

	for _, l := range language {
		offer.data[l] = emptyValue
	}

	return offer
}

// Using attaches a processor function to an offer and returns the modified offer.
// The original offer is unchanged.
func (o Offer) Using(processor Processor) Offer {
	o.processor = processor
	return o
}

// With attaches response data to an offer. The data can be a value (struct, slice, etc) or
// a data.Data. It may also be nil, which serves to add the language to the Offer's supported
// languages.
//
//The original offer is unchanged.
func (o Offer) With(language string, d interface{}) Offer {
	if d == nil {
		d = emptyValue
	}
	if len(o.data) == 0 && len(o.langs) == 1 && o.langs[0] == "*" {
		o.langs = nil
	}
	o.langs = append(o.langs, language)
	if s, ok := d.(data.Data); ok {
		o.data[language] = s
	} else {
		o.data[language] = data.Of(d)
	}
	return o
}

func (o Offer) String() string {
	buf := &strings.Builder{}
	buf.WriteString("Accept: ")
	buf.WriteString(o.ContentType.String())
	if len(o.data) > 0 {
		buf.WriteString(". Accept-Language: ")
		comma := ""
		for k := range o.data {
			buf.WriteString(comma)
			buf.WriteString(k)
			comma = ", "
		}
	}
	return buf.String()
}

//-------------------------------------------------------------------------------------------------

type empty struct{}

func (e empty) Content(string, string, bool) (interface{}, string, error) {
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
		if equalOrWildcard(mr.Type, typ) && equalOrWildcard(mr.Subtype, subtype) {
			allowed = append(allowed, mr)
		}
	}

	return allowed
}
