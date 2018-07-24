package uuid

import (
	"testing"
)

func TestVersion(t *testing.T) {
	type testrow struct {
		in    Version
		name  string
		valid bool
	}
	data := []testrow{
		{V1, "V1", true},
		{V2, "V2", true},
		{V3, "V3", true},
		{V4, "V4", true},
		{V5, "V5", true},
		{0, "V0", false},
		{42, "V42", false},
		{255, "V255", false},
	}
	for _, row := range data {
		t.Run(row.name, func(t *testing.T) {
			if name := row.in.String(); row.name != name {
				t.Errorf("wrong String: expected %q, got %q", row.name, name)
			}

			if valid := row.in.IsValid(); row.valid != valid {
				t.Errorf("wrong IsValid: expected %v, got %v", row.valid, valid)
			}

			var parsed Version
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
		"V256",
		"V",
		"V@",
		"V@1",
		"V1@",
	}
	for _, str := range faildata {
		var bogus Version
		if err := bogus.FromString(str); err == nil {
			t.Errorf("unexpected success at FromString %q", str)
		}
	}
}
