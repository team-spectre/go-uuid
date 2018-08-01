package uuid

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestUUID(t *testing.T) {
	type testrow struct {
		name          string
		dense         string
		version       Version
		variant       Variant
		isTimeBased   bool
		standardBytes [ByteLength]byte
		denseBytes    [ByteLength]byte
	}
	data := []testrow{
		{
			name:        "00000000-0000-0000-0000-000000000000",
			dense:       "@AAAAAAAAAAAAAAAAAAAAAA",
			version:     0,
			variant:     0,
			isTimeBased: false,
			standardBytes: [...]byte{
				0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00,
			},
			denseBytes: [...]byte{
				0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00,
			},
		},
		{
			name:        "77b99cea-8ab4-11e8-96a8-185e0fad6335",
			dense:       "@EeiKtHe5nOqWqBheD61jNQ",
			version:     V1,
			variant:     VariantRFC4122,
			isTimeBased: true,
			standardBytes: [...]byte{
				0x77, 0xb9, 0x9c, 0xea,
				0x8a, 0xb4, 0x11, 0xe8,
				0x96, 0xa8, 0x18, 0x5e,
				0x0f, 0xad, 0x63, 0x35,
			},
			denseBytes: [...]byte{
				0x11, 0xe8, 0x8a, 0xb4,
				0x77, 0xb9, 0x9c, 0xea,
				0x96, 0xa8, 0x18, 0x5e,
				0x0f, 0xad, 0x63, 0x35,
			},
		},
		{
			name:        "4baaf498-5125-4934-9822-483fd50b16c0",
			dense:       "@STRRJUuq9JiYIkg/1QsWwA",
			version:     V4,
			variant:     VariantRFC4122,
			isTimeBased: false,
			standardBytes: [...]byte{
				0x4b, 0xaa, 0xf4, 0x98,
				0x51, 0x25, 0x49, 0x34,
				0x98, 0x22, 0x48, 0x3f,
				0xd5, 0x0b, 0x16, 0xc0,
			},
			denseBytes: [...]byte{
				0x49, 0x34, 0x51, 0x25,
				0x4b, 0xaa, 0xf4, 0x98,
				0x98, 0x22, 0x48, 0x3f,
				0xd5, 0x0b, 0x16, 0xc0,
			},
		},
	}

	type testsubrow0 struct {
		name string
		f0   func(_, _ string) string
		f1   func(_ string) string
	}
	subdata0 := []testsubrow0{
		{"standard", firstString, identity},
		{"hashlike", firstString, dehyphened},
		{"bracketed", firstString, bracketed},
		{"urned", firstString, urned},
		{"dense", secondString, identity},
		{"dense-with-padding", secondString, denseWithPad},
		{"dense-wrong-padding", secondString, denseFakePad},
	}

	type testsubrow1 struct {
		vm ValueMode
		bm BinaryMode
		tm TextMode
		f0 func(_, _ interface{}) interface{}
		f1 func(_, _ []byte) []byte
		f2 func(_, _ string) string
		f3 func(_ string) string
	}
	subdata1 := []testsubrow1{
		{Text, StandardOnly, Canonical, firstVariant, firstBytes, firstString, identity},
		{Text, StandardOnly, HashLike, firstVariant, firstBytes, firstString, dehyphened},
		{Text, StandardOnly, Bracketed, firstVariant, firstBytes, firstString, bracketed},
		{Text, StandardOnly, URN, firstVariant, firstBytes, firstString, urned},
		{Text, StandardOnly, Dense, firstVariant, firstBytes, secondString, identity},

		{Text, DenseOnly, Canonical, firstVariant, secondBytes, firstString, identity},
		{Text, DenseOnly, HashLike, firstVariant, secondBytes, firstString, dehyphened},
		{Text, DenseOnly, Bracketed, firstVariant, secondBytes, firstString, bracketed},
		{Text, DenseOnly, URN, firstVariant, secondBytes, firstString, urned},
		{Text, DenseOnly, Dense, firstVariant, secondBytes, secondString, identity},

		{Binary, StandardOnly, Canonical, secondVariant, firstBytes, firstString, identity},
		{Binary, StandardOnly, HashLike, secondVariant, firstBytes, firstString, dehyphened},
		{Binary, StandardOnly, Bracketed, secondVariant, firstBytes, firstString, bracketed},
		{Binary, StandardOnly, URN, secondVariant, firstBytes, firstString, urned},
		{Binary, StandardOnly, Dense, secondVariant, firstBytes, secondString, identity},

		{Binary, DenseOnly, Canonical, secondVariant, secondBytes, firstString, identity},
		{Binary, DenseOnly, HashLike, secondVariant, secondBytes, firstString, dehyphened},
		{Binary, DenseOnly, Bracketed, secondVariant, secondBytes, firstString, bracketed},
		{Binary, DenseOnly, URN, secondVariant, secondBytes, firstString, urned},
		{Binary, DenseOnly, Dense, secondVariant, secondBytes, secondString, identity},
	}

	for _, row := range data {
		t.Run(row.name, func(t *testing.T) {
			var uuid UUID
			if err := uuid.FromString(row.name); err != nil {
				t.Errorf("failed to FromString %q: %v", row.name, err)
				return
			}

			version, variant := uuid.VersionAndVariant()
			if row.version != version {
				t.Errorf("wrong Version: expected %v, got %v", row.version, version)
			}
			if row.variant != variant {
				t.Errorf("wrong Variant: expected %v, got %v", row.variant, variant)
			}

			isTimeBased := uuid.IsV1()
			if row.isTimeBased != isTimeBased {
				t.Errorf("wrong IsV1: expected %v, got %v", row.isTimeBased, isTimeBased)
			}

			for _, subrow := range subdata0 {
				subname := fmt.Sprintf("FromString-%s", subrow.name)
				t.Run(subname, func(t *testing.T) {
					str := subrow.f1(subrow.f0(row.name, row.dense))
					var alt UUID
					if err := alt.FromString(str); err != nil {
						t.Errorf("failed to FromString %q: %v", str, err)
						return
					}
					if !uuid.Equal(alt) {
						t.Errorf("flubbed FromString %q: got %s", str, alt.CanonicalString())
					}
				})
			}

			for _, subrow := range subdata1 {
				pref := Preferences{
					Value:  subrow.vm,
					Binary: subrow.bm,
					Text:   subrow.tm,
				}
				subname := fmt.Sprintf("Marshal-%s", pref)
				t.Run(subname, func(t *testing.T) {
					dupe := uuid
					dupe.SetPreferences(pref)

					textRep := subrow.f3(subrow.f2(row.name, row.dense))
					byteRep := subrow.f1(row.standardBytes[:], row.denseBytes[:])
					valueRep := subrow.f0(textRep, byteRep)
					debugRep := debugging(row.name, row.dense)

					checkString(t, "CanonicalString", row.name, dupe.CanonicalString())
					checkString(t, "DenseString", row.dense, dupe.DenseString())
					checkString(t, "GoString", gostr(textRep), dupe.GoString())
					checkString(t, "String", textRep, dupe.String())
					checkText(t, "MarshalText", textRep, justBytes(dupe.MarshalText()))
					checkText(t, "MarshalJSON", quoted(textRep), justBytes(dupe.MarshalJSON()))
					checkBinary(t, "StandardBytes", row.standardBytes[:], dupe.StandardBytes())
					checkBinary(t, "DenseBytes", row.denseBytes[:], dupe.DenseBytes())
					checkBinary(t, "MarshalBinary", byteRep, justBytes(dupe.MarshalBinary()))
					checkVariant(t, "Value", valueRep, justVariant(dupe.Value()))

					checkString(t, "Format %d", row.dense, fmt.Sprintf("%d", dupe))
					checkString(t, "Format %s", textRep, fmt.Sprintf("%s", dupe))
					checkString(t, "Format %#s", row.name, fmt.Sprintf("%#s", dupe))
					checkString(t, "Format %+s", row.dense, fmt.Sprintf("%+s", dupe))
					checkString(t, "Format %q", quoted(textRep), fmt.Sprintf("%q", dupe))
					checkString(t, "Format %#q", quoted(row.name), fmt.Sprintf("%#q", dupe))
					checkString(t, "Format %+q", quoted(row.dense), fmt.Sprintf("%+q", dupe))
					checkString(t, "Format %v", textRep, fmt.Sprintf("%v", dupe))
					checkString(t, "Format %#v", row.name, fmt.Sprintf("%#v", dupe))
					checkString(t, "Format %+v", debugRep, fmt.Sprintf("%+v", dupe))
					checkString(t, "Format %30d", leftPad(30, row.dense), fmt.Sprintf("%30d", dupe))
					checkString(t, "Format %-30d", rightPad(30, row.dense), fmt.Sprintf("%-30d", dupe))
					checkString(t, "Format %+80v", leftPad(80, debugRep), fmt.Sprintf("%+80v", dupe))
					checkString(t, "Format %+-80v", rightPad(80, debugRep), fmt.Sprintf("%+-80v", dupe))
				})
			}
		})
	}
}

func TestFromString(t *testing.T) {
	standardBytes := []byte{
		0x77, 0xb9, 0x9c, 0xea,
		0x8a, 0xb4, 0x11, 0xe8,
		0x96, 0xa8, 0x18, 0x5e,
		0x0f, 0xad, 0x63, 0x35,
	}

	type testrow struct {
		input   string
		output  []byte
		failure bool
	}
	data := []testrow{
		{
			input:  "77b99cea-8ab4-11e8-96a8-185e0fad6335",
			output: standardBytes,
		},
		{
			input:  " 77b99cea 8ab4 11e8 96a8 185e0fad6335 ",
			output: standardBytes,
		},
		{
			input:  " { 77b99cea 8ab4 11e8 96a8 185e0fad6335 } ",
			output: standardBytes,
		},
		{
			input:  " urn:uuid: 77b99cea 8ab4 11e8 96a8 185e0fad6335 ",
			output: standardBytes,
		},
		{
			input:  " urn:uuid: { 77b99cea 8ab4 11e8 96a8 185e0fad6335 } ",
			output: standardBytes,
		},

		{
			input:   "77b99cea-8ab4-11e8-96a8-185e0fad633", // delete '5'
			failure: true,
		},
		{
			input:   "77b99cea-8ab4-11e8-96a8-185e0fad63350", // add '0'
			failure: true,
		},
		{
			input:   "77b99cea-8ab4-11e8-96a8-185e0fad63357", // add '7'
			failure: true,
		},
		{
			input:   "x77b99cea-8ab4-11e8-96a8-185e0fad6335", // add 'x' at start
			failure: true,
		},
		{
			input:   "77b99xcea-8ab4-11e8-96a8-185e0fad6335", // add 'x' at odd offset
			failure: true,
		},
		{
			input:   "77b99ceax-8ab4-11e8-96a8-185e0fad6335", // add 'x' at even offset
			failure: true,
		},
		{
			input:   "77b99cea-8ab4-11e8-96a8-185e0fad6335x", // add 'x' at end
			failure: true,
		},

		{
			input:  "@EeiKtHe5nOqWqBheD61jNQ",
			output: standardBytes,
		},
		{
			input:  "@EeiKtHe5nOqWqBheD61jNQ==",
			output: standardBytes,
		},
		{
			input:  "@EeiKtHe5nOqWqBheD61jNQA=",
			output: standardBytes,
		},
		{
			input:  "@EeiKtHe5nOqWqBheD61jNQAA",
			output: standardBytes,
		},
		{
			input:  " @ EeiKtHe5n OqWqBheD61jNQ ",
			output: standardBytes,
		},

		{
			input:   "@EeiKtHe5nOqWqBheD61j", // delete 'NQ'
			failure: true,
		},
		{
			input:   "@EeiKtHe5nOqWqBheD61jNQB", // append 'B'
			failure: true,
		},
		{
			input:   "@EeiKtHe5nOqWqBheD61jNQBA", // append 'BA'
			failure: true,
		},
		{
			input:   "@EeiKtHe5nOqWqBheD61jNQB=", // append 'B='
			failure: true,
		},
		{
			input:   "@EeiKtHe5nOqWqBheD61jNQAB", // append 'AB'
			failure: true,
		},
		{
			input:   "@EeiKtHe5nOqWqBheD61jNQ=A", // append '=A'
			failure: true,
		},
		{
			input:   "@EeiKtHe5nOqWqBheD61jNQAA=", // append 'AA='
			failure: true,
		},
		{
			input:   "@EeiKtHe5nOqWqBheD61jNQ===", // append '==='
			failure: true,
		},
	}
	for _, row := range data {
		t.Run(row.input, func(t *testing.T) {
			var expected, parsed UUID
			var err error

			err = parsed.FromString(row.input)
			if err != nil {
				if !row.failure {
					t.Errorf("unexpected failure at FromString %q: %v", row.input, err)
				}
			} else {
				if row.failure {
					t.Errorf("unexpected success at FromString %q: %s", row.input, parsed.CanonicalString())
				}
				if row.output != nil {
					copy(expected.a[:], row.output[0:ByteLength])
					if !parsed.Equal(expected) {
						t.Errorf("flubbed FromString %q: expected %s, got %s", row.input, expected.CanonicalString(), parsed.CanonicalString())
					}
				}
			}

			parsed.SetNil()
			err = parsed.UnmarshalText([]byte(row.input))
			if err != nil {
				if !row.failure {
					t.Errorf("unexpected failure at UnmarshalText %q: %v", row.input, err)
				}
			} else {
				if row.failure {
					t.Errorf("unexpected success at UnmarshalText %q: %s", row.input, parsed.CanonicalString())
				}
				if row.output != nil {
					copy(expected.a[:], row.output[0:ByteLength])
					if !parsed.Equal(expected) {
						t.Errorf("flubbed UnmarshalText %q: expected %s, got %s", row.input, expected.CanonicalString(), parsed.CanonicalString())
					}
				}
			}
		})
	}
}

func TestUnmarshalBinary(t *testing.T) {
	text := "77b99cea-8ab4-11e8-96a8-185e0fad6335"
	standardBytes := []byte{
		0x77, 0xb9, 0x9c, 0xea,
		0x8a, 0xb4, 0x11, 0xe8,
		0x96, 0xa8, 0x18, 0x5e,
		0x0f, 0xad, 0x63, 0x35,
	}
	denseBytes := []byte{
		0x11, 0xe8, 0x8a, 0xb4,
		0x77, 0xb9, 0x9c, 0xea,
		0x96, 0xa8, 0x18, 0x5e,
		0x0f, 0xad, 0x63, 0x35,
	}

	text0 := "11d3c015-c015-11d3-96a8-185e0fad6335"
	text1 := "c01511d3-c015-11d3-96a8-185e0fad6335"
	ambiguousBytes := []byte{
		0x11, 0xd3, 0xc0, 0x15,
		0xc0, 0x15, 0x11, 0xd3,
		0x96, 0xa8, 0x18, 0x5e,
		0x0f, 0xad, 0x63, 0x35,
	}

	type testrow struct {
		input   []byte
		order   BinaryMode
		success bool
		output  string
	}
	data := []testrow{
		{standardBytes, StandardOnly, true, text},
		{standardBytes, DenseOnly, false, ""},
		{standardBytes, StandardFirst, true, text},
		{standardBytes, DenseFirst, true, text},

		{denseBytes, StandardOnly, false, ""},
		{denseBytes, DenseOnly, true, text},
		{denseBytes, StandardFirst, true, text},
		{denseBytes, DenseFirst, true, text},

		{ambiguousBytes, StandardOnly, true, text0},
		{ambiguousBytes, StandardFirst, true, text0},
		{ambiguousBytes, DenseOnly, true, text1},
		{ambiguousBytes, DenseFirst, true, text1},
	}
	for _, row := range data {
		pref := bitsDefault.expand()
		pref.Binary = row.order
		name := fmt.Sprintf("%x-%s", row.input, pref)
		t.Run(name, func(t *testing.T) {
			var uuid UUID
			uuid.SetPreferences(pref)
			err := uuid.UnmarshalBinary(row.input)
			x := formatBytes(row.input)

			switch {
			case err != nil && row.success:
				t.Errorf("failed to UnmarshalBinary %s: %v", x, err)

			case err == nil && !row.success:
				t.Errorf("unexpected success at UnmarshalBinary %s: %s", x, uuid.CanonicalString())

			case err == nil:
				if output := uuid.CanonicalString(); row.output != output {
					t.Errorf("flubbed UnmarshalBinary %s: expected %s, got %s", x, row.output, output)
				}
			}
		})
	}
}

func TestUnmarshalBinaryFailure(t *testing.T) {
	data := [][]byte{
		// Missing last byte
		// [... ad 63 35] -> [... ad 63]
		{0x77, 0xb9, 0x9c, 0xea, 0x8a, 0xb4, 0x11, 0xe8, 0x96, 0xa8, 0x18, 0x5e, 0x0f, 0xad, 0x63},

		// Extra byte at end
		// [... ad 63 35] -> [... ad 63 35 42]
		{0x77, 0xb9, 0x9c, 0xea, 0x8a, 0xb4, 0x11, 0xe8, 0x96, 0xa8, 0x18, 0x5e, 0x0f, 0xad, 0x63, 0x35, 0x42},

		// Version 0
		// [... b4 11 e8 ...] -> [... b4 01 e8 ...]
		{0x77, 0xb9, 0x9c, 0xea, 0x8a, 0xb4, 0x01, 0xe8, 0x96, 0xa8, 0x18, 0x5e, 0x0f, 0xad, 0x63, 0x35},

		// Version 6
		// [... b4 11 e8 ...] -> [... b4 61 e8 ...]
		{0x77, 0xb9, 0x9c, 0xea, 0x8a, 0xb4, 0x61, 0xe8, 0x96, 0xa8, 0x18, 0x5e, 0x0f, 0xad, 0x63, 0x35},

		// Variant NCS
		// [... e8 96 a8 ...] -> [... e8 76 a8 ...]
		{0x77, 0xb9, 0x9c, 0xea, 0x8a, 0xb4, 0x11, 0xe8, 0x76, 0xa8, 0x18, 0x5e, 0x0f, 0xad, 0x63, 0x35},

		// Variant Microsoft
		// [... e8 96 a8 ...] -> [... e8 c6 a8 ...]
		{0x77, 0xb9, 0x9c, 0xea, 0x8a, 0xb4, 0x11, 0xe8, 0xc6, 0xa8, 0x18, 0x5e, 0x0f, 0xad, 0x63, 0x35},

		// Variant future
		// [... e8 96 a8 ...] -> [... e8 e6 a8 ...]
		{0x77, 0xb9, 0x9c, 0xea, 0x8a, 0xb4, 0x11, 0xe8, 0xe6, 0xa8, 0x18, 0x5e, 0x0f, 0xad, 0x63, 0x35},
	}
	for _, item := range data {
		name := fmt.Sprintf("%x", item)
		t.Run(name, func(t *testing.T) {
			var uuid UUID
			uuid.SetPreferences(Preferences{
				Value:  Binary,
				Binary: StandardOnly,
				Text:   Canonical,
			})
			if err := uuid.UnmarshalBinary(item); err == nil {
				x := formatBytes(item)
				t.Errorf("unexpected success at UnmarshalBinary %s: %s", x, uuid.CanonicalString())
			}
		})
	}
}

func justBytes(ba []byte, _ error) []byte {
	return ba
}

func justVariant(v interface{}, _ error) interface{} {
	return v
}

func checkString(t *testing.T, opName, x, y string) {
	if x != y {
		t.Errorf("flubbed %s:\n\texpected: %q\n\t  actual: %q", opName, x, y)
	}
}

func checkText(t *testing.T, opName, x string, ba []byte) {
	y := string(ba)
	if x != y {
		t.Errorf("flubbed %s:\n\texpected: %q\n\t  actual: %q", opName, x, y)
	}
}

func checkBinary(t *testing.T, opName string, ba0, ba1 []byte) {
	if !equalBytes(ba0, ba1) {
		x := formatBytes(ba0)
		y := formatBytes(ba1)
		t.Errorf("flubbed %s:\n\texpected: %s\n\t  actual: %s", opName, x, y)
	}
}

func checkVariant(t *testing.T, opName string, v0, v1 interface{}) {
	s0, ok0 := v0.(string)
	s1, ok1 := v1.(string)
	if ok0 && ok1 {
		checkString(t, opName, s0, s1)
		return
	}

	ba0, ok0 := v0.([]byte)
	ba1, ok1 := v1.([]byte)
	if ok0 && ok1 {
		checkBinary(t, opName, ba0, ba1)
		return
	}

	t.Errorf("flubbed %s: expected type %T, got type %T", opName, v0, v1)
}

func compose(list ...func(string) string) func(string) string {
	return func(str string) string {
		i := len(list)
		for i > 0 {
			i--
			fn := list[i]
			str = fn(str)
		}
		return str
	}
}

func firstBytes(a, _ []byte) []byte  { return a }
func secondBytes(_, b []byte) []byte { return b }

func firstString(a, _ string) string  { return a }
func secondString(_, b string) string { return b }

func firstVariant(a, _ interface{}) interface{}  { return a }
func secondVariant(_, b interface{}) interface{} { return b }

func identity(str string) string     { return str }
func quoted(str string) string       { return `"` + str + `"` }
func gostr(str string) string        { return `UUID("` + str + `")` }
func dehyphened(str string) string   { return strings.Replace(str, "-", "", -1) }
func bracketed(str string) string    { return `{` + str + `}` }
func urned(str string) string        { return `urn:uuid:` + str }
func denseWithPad(str string) string { return str + `==` }
func denseFakePad(str string) string { return str + `AA` }

func leftPad(n int, str string) string {
	if n < len(str) {
		n = len(str)
	}
	var buf bytes.Buffer
	buf.Grow(n)
	for pad := n - len(str); pad > 0; pad-- {
		buf.WriteByte(' ')
	}
	buf.WriteString(str)
	return buf.String()
}

func rightPad(n int, str string) string {
	if n < len(str) {
		n = len(str)
	}
	var buf bytes.Buffer
	buf.Grow(n)
	buf.WriteString(str)
	for pad := n - len(str); pad > 0; pad-- {
		buf.WriteByte(' ')
	}
	return buf.String()
}

// This is not a stable API, so feel free to modify this function.
func debugging(a, b string) string { return quoted(b) + " " + bracketed(a) }
