package uuid

import (
	"testing"
)

func TestVariant(t *testing.T) {
	type testrow struct {
		in    Variant
		name  string
		valid bool
	}
	data := []testrow{
		{VariantNCS, "VariantNCS", false},
		{VariantRFC4122, "VariantRFC4122", true},
		{VariantMicrosoft, "VariantMicrosoft", false},
		{VariantFuture, "VariantFuture", false},
		{0, "Variant(0)", false},
		{42, "Variant(42)", false},
		{255, "Variant(255)", false},
	}
	for _, row := range data {
		t.Run(row.name, func(t *testing.T) {
			if name := row.in.String(); row.name != name {
				t.Errorf("wrong String: expected %q, got %q", row.name, name)
			}

			if valid := row.in.IsValid(); row.valid != valid {
				t.Errorf("wrong IsValid: expected %v, got %v", row.valid, valid)
			}

			var parsed Variant
			err := parsed.FromString(row.name)
			if err != nil {
				t.Errorf("failed to FromString: %v", err)
			}
			if err == nil && row.in != parsed {
				t.Errorf("wrong FromString: expected %d, got %d", row.in, parsed)
			}
		})
	}

	faildata := []string{
		"Variant(256)",
		"Variant(0",
		"Variant(0)@",
		"Variant(0@)",
		"Variant(@0)",
	}
	for _, str := range faildata {
		var bogus Variant
		if err := bogus.FromString(str); err == nil {
			t.Errorf("unexpected success at FromString %q", str)
		}
	}
}
