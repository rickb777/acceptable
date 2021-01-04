package acceptable

import (
	"fmt"
)

// Offer holds information about one particular resource representation that can potentially
// provide an acceptable response.
type Offer struct {
	ContentType
	Language  string
	Processor func() error
}

// OfferOf constructs an Offer easily.
// If the language is absent, it is assumed to be the wildcard "*".
// If the content type is blank, it is assumed to be the full wildcard "*/*".
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
func (o Offer) With(processor func() error) Offer {
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
