package uuid

import (
	"testing"
)

func TestBitsDefaulted(t *testing.T) {
	type testrow struct {
		in  bits
		out bits
	}
	data := []testrow{
		{0x00, bitsDefault},
		{0x7f, bitsDefault},
		{0x80, 0x80},
		{0xff, 0xff},
		{bitsDefault, bitsDefault},
	}
	for _, row := range data {
		name := row.in.String()
		t.Run(name, func(t *testing.T) {
			actual := row.in.defaulted()
			expected := row.out
			if expected != actual {
				t.Errorf("wrong defaulted: expected %08b, got %08b", expected, actual)
			}
		})
	}
}
