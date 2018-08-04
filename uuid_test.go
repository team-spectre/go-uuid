package uuid

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestUUID_RoundTripping(t *testing.T) {
	type testrow struct {
		name          string
		dense         string
		version       Version
		variant       Variant
		isValid       bool
		isV1          bool
		standardBytes [ByteLength]byte
		denseBytes    [ByteLength]byte
	}
	data := []testrow{
		{
			name:    "00000000-0000-0000-0000-000000000000",
			dense:   "@AAAAAAAAAAAAAAAAAAAAAA",
			version: 0,
			variant: 0,
			isValid: false,
			isV1:    false,
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
			name:    "77b99cea-8ab4-11e8-96a8-185e0fad6335",
			dense:   "@EeiKtHe5nOqWqBheD61jNQ",
			version: V1,
			variant: VariantRFC4122,
			isValid: true,
			isV1:    true,
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
			name:    "4baaf498-5125-4934-9822-483fd50b16c0",
			dense:   "@STRRJUuq9JiYIkg/1QsWwA",
			version: V4,
			variant: VariantRFC4122,
			isValid: true,
			isV1:    false,
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

			checkVersion(t, "", row.version, uuid)
			checkVariant(t, "", row.variant, uuid)
			checkValid(t, "", row.isValid, uuid)
			checkV1(t, "", row.isV1, uuid)

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
					checkValue(t, "Value", valueRep, justValue(dupe.Value()))

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

func TestUUID_BasicOps(t *testing.T) {
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

	checkBinary(t, "Nil()", allZeroes[:], Nil().StandardBytes())
	checkGenerated(t, "New()", New().StandardBytes())
	checkBinary(t, "MustFromBytes(S)", standardBytes, MustFromBytes(standardBytes).StandardBytes())
	checkBinary(t, "MustFromBytes(D)", standardBytes, MustFromBytes(denseBytes).StandardBytes())
	checkBinary(t, "MustFromString()", standardBytes, MustFromString(text).StandardBytes())

	var u0, u1, u2, u3, u4 UUID
	var e0, e1, e2, e3, e4 error

	u0, e0 = FromBytes(standardBytes)
	u1, e1 = FromBytes(denseBytes)
	u2, e2 = FromString(text)
	checkOp(t, "FromBytes(S)", standardBytes, u0, e0)
	checkOp(t, "FromBytes(D)", standardBytes, u1, e1)
	checkOp(t, "FromString()", standardBytes, u2, e2)

	u0 = UUID{}
	u1 = UUID{}
	u2 = UUID{}

	u0.MustFromBytes(standardBytes)
	u1.MustFromBytes(denseBytes)
	u2.MustFromString(text)
	checkBinary(t, "UUID.MustFromBytes(S)", standardBytes, u0.StandardBytes())
	checkBinary(t, "UUID.MustFromBytes(D)", standardBytes, u1.StandardBytes())
	checkBinary(t, "UUID.MustFromString()", standardBytes, u2.StandardBytes())

	u0 = UUID{}
	u1 = UUID{}
	u2 = UUID{}

	e0 = u0.FromStandardBytes(standardBytes)
	e1 = u1.FromDenseBytes(denseBytes)
	e2 = u2.FromBytes(standardBytes)
	e3 = u3.FromBytes(denseBytes)
	e4 = u4.FromString(text)
	checkOp(t, "UUID.FromStandardBytes()", standardBytes, u0, e0)
	checkOp(t, "UUID.FromDenseBytes()", standardBytes, u1, e1)
	checkOp(t, "UUID.FromBytes(S)", standardBytes, u2, e2)
	checkOp(t, "UUID.FromBytes(D)", standardBytes, u3, e3)
	checkOp(t, "UUID.FromString()", standardBytes, u4, e4)

	u0 = UUID{}
	u1 = UUID{}
	u2 = UUID{}
	u3 = UUID{}
	u4 = UUID{}

	u1.SetPreferences(Preferences{Binary, StandardFirst, Bracketed})
	u3.SetPreferences(u1.Preferences())

	u0.MustFromString(text)
	u0.SetNil()
	u1.MustFromString(text)
	u1.SetNil()
	checkBinary(t, "UUID.SetNil()", allZeroes[:], u0.StandardBytes())
	checkBinary(t, "UUID.SetNil()", allZeroes[:], u1.StandardBytes())
	checkVersion(t, "u0 after UUID.SetNil()", 0, u0)
	checkVariant(t, "u0 after UUID.SetNil()", 0, u0)
	checkValid(t, "u0 after UUID.SetNil()", false, u0)
	checkPrefs(t, "u0 after UUID.SetNil()", Text, DenseFirst, Dense, u0)
	checkPrefs(t, "u1 after UUID.SetNil()", Binary, StandardFirst, Bracketed, u1)
	checkEqual(t, "Equal", true, u0, u1)

	u2.MustFromString(text)
	u3.MustFromString(text)
	checkBinary(t, "UUID.MustFromString()", standardBytes, u2.StandardBytes())
	checkBinary(t, "UUID.MustFromString()", standardBytes, u3.StandardBytes())
	checkVersion(t, "u2 after UUID.MustFromString()", V1, u2)
	checkVariant(t, "u2 after UUID.MustFromString()", VariantRFC4122, u2)
	checkValid(t, "u2 after UUID.MustFromString()", true, u2)
	checkPrefs(t, "u2 after UUID.MustFromString()", Text, DenseFirst, Dense, u2)
	checkPrefs(t, "u3 after UUID.MustFromString()", Binary, StandardFirst, Bracketed, u3)
	checkEqual(t, "Equal", true, u2, u3)
	checkEqual(t, "Equal", false, u0, u2)

	u0 = UUID{}
	u1 = UUID{}
	u2 = UUID{}
	u3 = UUID{}

	u0.SetNew()
	u1.SetNew()
	checkGenerated(t, "SetNew()", u0.StandardBytes())
	checkGenerated(t, "SetNew()", u1.StandardBytes())
}

func TestUUID_FromString(t *testing.T) {
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
			input:  "",
			output: allZeroes[:],
		},
		{
			input:  "00000000-0000-0000-0000-000000000000",
			output: allZeroes[:],
		},
		{
			input:  "00000000 0000 0000 0000 000000000000",
			output: allZeroes[:],
		},
		{
			input:  "00000000000000000000000000000000",
			output: allZeroes[:],
		},
		{
			input:  "{00000000000000000000000000000000}",
			output: allZeroes[:],
		},
		{
			input:  "urn:uuid:00000000000000000000000000000000",
			output: allZeroes[:],
		},
		{
			input:  "urn:uuid:{00000000000000000000000000000000}",
			output: allZeroes[:],
		},

		{
			input:  "77b99cea-8ab4-11e8-96a8-185e0fad6335",
			output: standardBytes,
		},
		{
			input:  "{77b99cea-8ab4-11e8-96a8-185e0fad6335}",
			output: standardBytes,
		},
		{
			input:  "urn:uuid:77b99cea-8ab4-11e8-96a8-185e0fad6335",
			output: standardBytes,
		},
		{
			input:  "urn:uuid:{77b99cea-8ab4-11e8-96a8-185e0fad6335}",
			output: standardBytes,
		},
		{
			input:  "77b99cea8ab411e896a8185e0fad6335",
			output: standardBytes,
		},
		{
			input:  "77B99CEA8AB411E896A8185E0FAD6335",
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
			input:   "{77b99cea-8ab4-11e8-96a8-185e0fad6335", // '{' without '}'
			failure: true,
		},
		{
			input:   "77b99cea-{8ab4-11e8-96a8-185e0fad6335}", // '{' in weird place
			failure: true,
		},
		{
			input:   "{77b99cea-8ab4-11e8-96a8}-185e0fad6335", // '}' in weird place
			failure: true,
		},

		{
			input:  "@AZaz09+/-_AAAAAAAAAAAA",
			output: nil,
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
			input:   "@EeiKtHe5!nOqWqBheD61jNQ", // add '!'
			failure: true,
		},
		{
			input:   "@EeiKtHe5nOqWqBheD61j", // delete 'NQ'
			failure: true,
		},
		{
			input:   "@EeiKtHe5nOqWqBheD61jN/", // 'Q' -> '/'
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

func TestUUID_UnmarshalBinary(t *testing.T) {
	zero := "00000000-0000-0000-0000-000000000000"

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

	ambiguousText0A := "11d3c015-c015-11d3-96a8-185e0fad6335"
	ambiguousText0B := "c01511d3-c015-11d3-96a8-185e0fad6335"
	ambiguousBytes0 := []byte{
		0x11, 0xd3, 0xc0, 0x15,
		0xc0, 0x15, 0x11, 0xd3,
		0x96, 0xa8, 0x18, 0x5e,
		0x0f, 0xad, 0x63, 0x35,
	}

	ambiguousText1A := "11ec6a8f-6a8f-11ec-96a8-185e0fad6335"
	ambiguousText1B := "6a8f11ec-6a8f-11ec-96a8-185e0fad6335"
	ambiguousBytes1 := []byte{
		0x11, 0xec, 0x6a, 0x8f,
		0x6a, 0x8f, 0x11, 0xec,
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
		{nil, StandardOnly, true, zero},
		{nil, DenseOnly, true, zero},
		{nil, StandardFirst, true, zero},
		{nil, DenseFirst, true, zero},

		{allZeroes[:], StandardOnly, true, zero},
		{allZeroes[:], DenseOnly, true, zero},
		{allZeroes[:], StandardFirst, true, zero},
		{allZeroes[:], DenseFirst, true, zero},

		{standardBytes, StandardOnly, true, text},
		{standardBytes, DenseOnly, false, ""},
		{standardBytes, StandardFirst, true, text},
		{standardBytes, DenseFirst, true, text},

		{denseBytes, StandardOnly, false, ""},
		{denseBytes, DenseOnly, true, text},
		{denseBytes, StandardFirst, true, text},
		{denseBytes, DenseFirst, true, text},

		{ambiguousBytes0, StandardOnly, true, ambiguousText0A},
		{ambiguousBytes0, StandardFirst, true, ambiguousText0A},
		{ambiguousBytes0, DenseOnly, true, ambiguousText0B},
		{ambiguousBytes0, DenseFirst, true, ambiguousText0B},

		{ambiguousBytes1, StandardOnly, true, ambiguousText1A},
		{ambiguousBytes1, StandardFirst, true, ambiguousText1A},
		{ambiguousBytes1, DenseOnly, true, ambiguousText1B},
		{ambiguousBytes1, DenseFirst, true, ambiguousText1B},
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

func TestUUID_UnmarshalBinary_Failure(t *testing.T) {
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
			uuid.SetPreferences(Preferences{Binary, StandardOnly, Canonical})
			if err := uuid.UnmarshalBinary(item); err == nil {
				x := formatBytes(item)
				t.Errorf("unexpected success at UnmarshalBinary %s: %s", x, uuid.CanonicalString())
			}
		})
	}
}

func TestUUID_JSON(t *testing.T) {
	standardBytes := []byte{
		0x77, 0xb9, 0x9c, 0xea,
		0x8a, 0xb4, 0x11, 0xe8,
		0x96, 0xa8, 0x18, 0x5e,
		0x0f, 0xad, 0x63, 0x35,
	}

	zero := "00000000-0000-0000-0000-000000000000"
	zeroDense := "@AAAAAAAAAAAAAAAAAAAAAA"
	text := "77b99cea-8ab4-11e8-96a8-185e0fad6335"
	textDense := "@EeiKtHe5nOqWqBheD61jNQ"

	zeroJSON := `"` + zero + `"`
	textJSON := `"` + text + `"`
	textDenseJSON := `"` + textDense + `"`
	zeroDenseJSON := `"` + zeroDense + `"`

	var u UUID
	u.SetPreferences(Preferences{0, 0, Canonical})
	checkMarshalJSON(t, "canon nil", zeroJSON, u)
	u.MustFromString(text)
	checkMarshalJSON(t, "canon real", textJSON, u)
	u.SetPreferences(Preferences{0, 0, Dense})
	checkMarshalJSON(t, "dense real", textDenseJSON, u)
	u.SetNil()
	checkMarshalJSON(t, "dense nil", zeroDenseJSON, u)

	checkUnmarshalJSON(t, "canon nil", allZeroes[:], zeroJSON)
	checkUnmarshalJSON(t, "canon real", standardBytes, textJSON)
	checkUnmarshalJSON(t, "dense nil", allZeroes[:], zeroDenseJSON)
	checkUnmarshalJSON(t, "dense real", standardBytes, textDenseJSON)
	checkUnmarshalJSON(t, "JSON empty string", allZeroes[:], `""`)
	checkUnmarshalJSON(t, "JSON null", allZeroes[:], `null`)
	checkUnmarshalJSON(t, "JSON omit", allZeroes[:], ``)
}

func TestUUID_Scan(t *testing.T) {
	zero := "00000000-0000-0000-0000-000000000000"
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

	type testrow struct {
		name    string
		input   interface{}
		success bool
		output  string
	}
	data := []testrow{
		{"nil", nil, true, zero},
		{"bytes-nil", []byte(nil), true, zero},
		{"bytes-zero", allZeroes[:], true, zero},
		{"bytes-std", standardBytes, true, text},
		{"bytes-dense", denseBytes, true, text},
		{"string-empty", "", true, zero},
		{"string-zero", zero, true, zero},
		{"string-text", text, true, text},
	}
	for _, row := range data {
		pref := bitsDefault.expand()
		pref.Binary = DenseFirst
		t.Run(row.name, func(t *testing.T) {
			var uuid UUID
			uuid.SetPreferences(pref)
			err := uuid.Scan(row.input)
			output := uuid.CanonicalString()

			switch {
			case err != nil && row.success:
				t.Errorf("failed to Scan %[1]T %[1]v: %v", row.input, err)

			case err == nil && !row.success:
				t.Errorf("unexpected success at Scan %[1]T %[1]v: %q", row.input, output)

			case err == nil:
				if row.output != output {
					t.Errorf("flubbed Scan %[1]T %[1]v: expected %q, got %q", row.input, row.output, output)
				}
			}
		})
	}
}

func justBytes(ba []byte, _ error) []byte {
	return ba
}

func justValue(v interface{}, _ error) interface{} {
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

func checkGenerated(t *testing.T, opName string, ba []byte) {
	if !isSystemGenerated(ba) {
		x := formatBytes(ba)
		t.Errorf("flubbed %s:\n\t  actual: %s", opName, x)
	}
}

func checkOp(t *testing.T, opName string, expect []byte, u UUID, err error) {
	if err != nil {
		t.Errorf("failed to %s: got %#v", opName, err)
		return
	}
	actual := u.StandardBytes()
	if !equalBytes(expect, actual) {
		x := formatBytes(expect)
		y := formatBytes(actual)
		t.Errorf("flubbed %s: expected %q, got %q", opName, x, y)
	}
}

func checkValue(t *testing.T, opName string, v0, v1 interface{}) {
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

func checkEqual(t *testing.T, opName string, expect bool, u0, u1 UUID) {
	actual := Equal(u0, u1)
	if expect != actual {
		x := formatBytes(u0.StandardBytes())
		y := formatBytes(u1.StandardBytes())
		t.Errorf("flubbed %s: expected %v, got %v\n\t%s\n\t%s", opName, expect, actual, x, y)
	}
}

func checkVersion(t *testing.T, context string, expect Version, u UUID) {
	actual := u.Version()
	if expect != actual {
		t.Errorf("%s: expected Version() = %v, got %v", context, expect, actual)
	}
}

func checkVariant(t *testing.T, context string, expect Variant, u UUID) {
	actual := u.Variant()
	if expect != actual {
		t.Errorf("%s: expected Variant() = %v, got %v", context, expect, actual)
	}
}

func checkValid(t *testing.T, context string, expect bool, u UUID) {
	actual := u.IsValid()
	if expect != actual {
		t.Errorf("%s: expected IsValid() = %v, got %v", context, expect, actual)
	}
}

func checkV1(t *testing.T, context string, expect bool, u UUID) {
	actual := u.IsV1()
	if expect != actual {
		t.Errorf("%s: expected IsV1() = %v, got %v", context, expect, actual)
	}
}

func checkPrefs(t *testing.T, context string, vm ValueMode, bm BinaryMode, tm TextMode, u UUID) {
	p := u.Preferences()
	if vm != 0 && p.Value != vm {
		t.Errorf("%s: wrong Preferences.Value: expected %q, got %q", context, vm, p.Value)
	}
	if bm != 0 && p.Binary != bm {
		t.Errorf("%s: wrong Preferences.Binary: expected %q, got %q", context, bm, p.Binary)
	}
	if tm != 0 && p.Text != tm {
		t.Errorf("%s: wrong Preferences.Text: expected %q, got %q", context, tm, p.Text)
	}
}

func checkMarshalJSON(t *testing.T, context string, expect string, u UUID) {
	raw, err := u.MarshalJSON()
	if err != nil {
		t.Errorf("%s: MarshalJSON failed: %#v", context, err)
		return
	}
	actual := string(raw)
	if expect != actual {
		x := formatBytes(u.StandardBytes())
		t.Errorf("%s: flubbed MarshalJSON: expected %q, got %q\n\t%s", context, expect, actual, x)
	}
}

func checkUnmarshalJSON(t *testing.T, context string, expect []byte, input string) {
	var u UUID
	u.MustFromString("77b99cea-8ab4-11e8-96a8-185e0fad6335")
	err := u.UnmarshalJSON([]byte(input))
	if err != nil {
		t.Errorf("%s: UnmarshalJSON failed: %#v", context, err)
		return
	}
	actual := u.StandardBytes()
	if !equalBytes(expect, actual) {
		x := formatBytes(expect)
		y := formatBytes(actual)
		t.Errorf("%s: flubbed UnmarshalJSON: expected %s, got %s\n\t%s", context, x, y, input)
	}
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
