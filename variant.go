package uuid

import (
	"fmt"
)

// Variant indicates the RFC 4122-defined UUID variant.
type Variant byte

// Variant enum constants.
const (
	VariantNCS       Variant = 1
	VariantRFC4122   Variant = 2
	VariantMicrosoft Variant = 3
	VariantFuture    Variant = 4
)

var variantMap = map[Variant]string{
	VariantNCS:       "VariantNCS",
	VariantRFC4122:   "VariantRFC4122",
	VariantMicrosoft: "VariantMicrosoft",
	VariantFuture:    "VariantFuture",
}

// IsValid returns true iff the Variant is VariantRFC4122.
func (variant Variant) IsValid() bool {
	return (variant == VariantRFC4122)
}

func (variant Variant) String() string {
	if str, found := variantMap[variant]; found {
		return str
	}
	return fmt.Sprintf("Variant(%d)", variant)
}

// FromString attempts to parse the string representation of a Variant.
func (variant *Variant) FromString(in string) error {
	for k, v := range variantMap {
		if v == in {
			*variant = k
			return nil
		}
	}
	var b byte
	n, err := fmt.Sscanf(in+"|", "Variant(%d)|", &b)
	if n == 1 && err == nil {
		*variant = Variant(b)
		return nil
	}
	return makeParseError("Variant", "FromString", []byte(in), true)
}
