package acceptable

import (
	"fmt"
	"net/http"
)

// Supplier supplies data in the form of a struct, a slice, etc.
// This allows for evaluation on demand ('lazy'), e.g. fetching from a database.
type Supplier func() (interface{}, error)

// Processor is a function that renders content according to the matched result.
// The data can be a struct, slice etc or a Supplier.
type Processor func(w http.ResponseWriter, match *Match, template string, data interface{}) error

// Offer holds information about one particular resource representation that can potentially
// provide an acceptable response.
type Offer struct {
	// ContentType is the content type that is to be matched.
	// Wildcard values may be used.
	ContentType

	// Language defines which language will be provided if this offer is matched.
	// Can also be blank or "*" - both indicate that this is not used.
	Language string

	// Charset returns the preferred character set for the response, if any.
	// This is set on return from the BestRequestMatch function.
	Charset string

	// Processor is an optional function you can use to apply the offer if it is selected.
	// How this is used is entirely at the discretion of the call site.
	Processor Processor

	// Data is an optional response to be rendered if this offer is selected.
	// If Data is a Supplier function, the data can be sourced lazily.
	Data interface{}
}

// OfferOf constructs an Offer easily.
// If the language is absent, it is assumed to be the wildcard "*".
// If the content type is blank, it is assumed to be the full wildcard "*/*".
// Also, contentType can be a partial wildcard "type/*".
func OfferOf(contentType string, language ...string) Offer {
	t, s, l := "*", "*", "*"
	if contentType != "" {
		t, s = split(contentType, '/')
	}
	if len(language) > 0 {
		l = language[0]
	}
	return Offer{
		ContentType: ContentType{
			Type:    t,
			Subtype: s,
		},
		Language: l,
	}
}

// Using attaches a processor function to an offer and returns the modified offer.
// The original offer is unchanged.
func (o Offer) Using(processor Processor) Offer {
	o.Processor = processor
	return o
}

// With attaches response data to an offer.
// The original offer is unchanged.
func (o Offer) With(data interface{}) Offer {
	o.Data = data
	return o
}

func (o Offer) String() string {
	return fmt.Sprintf("Accept: %s. Accept-Language: %s", o.ContentType, o.Language)
}

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
