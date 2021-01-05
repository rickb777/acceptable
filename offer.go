package acceptable

import (
	"fmt"
	"net/http"
)

// Processor is a function that renders content according to the matched result.
type Processor func(w http.ResponseWriter, match Match, template string, dataModel interface{}) error

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
}

// OfferOf constructs an Offer easily.
// If the language is absent, it is assumed to be the wildcard "*".
// If the content type is blank, it is assumed to be the full wildcard "*/*".
// If the content subtype is blank, it is assumed to be the partial wildcard "type/*".
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

// With attaches a processor function to an offer and returns the modified offer.
// The original offer is unchanged.
func (o Offer) With(processor Processor) Offer {
	o.Processor = processor
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
