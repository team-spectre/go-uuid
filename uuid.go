package uuid

import (
	"database/sql"
	"database/sql/driver"
	"encoding"
	"encoding/json"
	"fmt"
)

// ByteLength is the number of bytes in the binary representation of a UUID.
const ByteLength = 16

// UUID holds a single Universally Unique Identifier.
type UUID struct {
	a [ByteLength]byte
	b bits
}

var _ fmt.Stringer = UUID{}
var _ fmt.Stringer = (*UUID)(nil)

var _ fmt.GoStringer = UUID{}
var _ fmt.GoStringer = (*UUID)(nil)

var _ fmt.Formatter = UUID{}
var _ fmt.Formatter = (*UUID)(nil)

var _ encoding.BinaryMarshaler = UUID{}
var _ encoding.BinaryMarshaler = (*UUID)(nil)
var _ encoding.BinaryUnmarshaler = (*UUID)(nil)

var _ encoding.TextMarshaler = UUID{}
var _ encoding.TextMarshaler = (*UUID)(nil)
var _ encoding.TextUnmarshaler = (*UUID)(nil)

var _ json.Marshaler = UUID{}
var _ json.Marshaler = (*UUID)(nil)
var _ json.Unmarshaler = (*UUID)(nil)

var _ driver.Valuer = UUID{}
var _ driver.Valuer = (*UUID)(nil)
var _ sql.Scanner = (*UUID)(nil)

// Nil returns a Nil-valued UUID.
func Nil() UUID {
	var uuid UUID
	return uuid
}

// New returns a newly generated V1 UUID.
func New() UUID {
	var uuid UUID
	uuid.SetNew()
	return uuid
}

// FromBytes attempts to parse a binary UUID representation.
func FromBytes(in []byte) (UUID, error) {
	var uuid UUID
	err := unmarshalBinary("", "FromBytes", uuid.a[:], in, uuid.getBits())
	return uuid, err
}

// MustFromBytes parses a binary UUID representation, or panics if it cannot.
func MustFromBytes(in []byte) UUID {
	var uuid UUID
	err := unmarshalBinary("", "MustFromBytes", uuid.a[:], in, uuid.getBits())
	if err != nil {
		panic(err)
	}
	return uuid
}

// FromString attempts to parse a textual UUID representation.
func FromString(in string) (UUID, error) {
	var uuid UUID
	err := unmarshalText("", "FromString", uuid.a[:], []byte(in))
	return uuid, err
}

// MustFromString parses a textual UUID representation, or panics if it cannot.
func MustFromString(in string) UUID {
	var uuid UUID
	err := unmarshalText("", "MustFromString", uuid.a[:], []byte(in))
	if err != nil {
		panic(err)
	}
	return uuid
}

// Equal returns true iff its two arguments hold the same UUID.
func Equal(u1, u2 UUID) bool {
	return u1.Equal(u2)
}

func (uuid UUID) getBits() bits {
	return uuid.b.defaulted()
}

// Preferences returns the preference knobs for this object.
func (uuid UUID) Preferences() Preferences {
	return uuid.b.defaulted().expand()
}

// SetPreferences updates the preference knobs for this object.
func (uuid *UUID) SetPreferences(pref Preferences) {
	uuid.b = pref.collapse(uuid.getBits())
}

// SetNil updates this UUID to hold the Nil UUID.
func (uuid *UUID) SetNil() {
	zeroBytes(uuid.a[:])
}

// SetNew updates this UUID to hold a newly generated V1 UUID.
func (uuid *UUID) SetNew() {
	globalState().generate(uuid.a[:])
}

// IsNil returns true iff this object holds the Nil UUID.
func (uuid UUID) IsNil() bool {
	var zero [ByteLength]byte
	return uuid.a == zero
}

// Equal returns true iff this object and the argument hold the same UUID.
func (uuid UUID) Equal(other UUID) bool {
	return uuid.a == other.a
}

// VersionAndVariant returns this UUID's Version and Variant.  This method is
// slightly more efficient than calling Version() and Variant() separately.
func (uuid UUID) VersionAndVariant() (version Version, variant Variant) {
	var tmp [ByteLength]byte
	marshalBinaryStandard(tmp[:], uuid.a[:])
	version, _, variant = extract(tmp[:])
	return
}

// Version returns this UUID's Version.
func (uuid UUID) Version() Version {
	version, _ := uuid.VersionAndVariant()
	return version
}

// Variant returns this UUID's Variant.
func (uuid UUID) Variant() Variant {
	_, variant := uuid.VersionAndVariant()
	return variant
}

// IsV1 returns true iff this UUID has a Version of V1.
func (uuid UUID) IsV1() bool {
	return uuid.Version() == V1
}

// IsValid returns true iff this UUID has a valid Version and a valid Variant.
func (uuid UUID) IsValid() bool {
	version, variant := uuid.VersionAndVariant()
	return version.IsValid() && variant.IsValid()
}

// StandardBytes returns the binary representation of this UUID in RFC 4122 byte order.
func (uuid UUID) StandardBytes() []byte {
	var out [ByteLength]byte
	marshalBinaryStandard(out[:], uuid.a[:])
	return out[:]
}

// DenseBytes returns the binary representation of this UUID in "dense" byte order.
func (uuid UUID) DenseBytes() []byte {
	var out [ByteLength]byte
	marshalBinaryDense(out[:], uuid.a[:])
	return out[:]
}

// Bytes returns the binary representation of this UUID.
func (uuid UUID) Bytes() []byte {
	var out [ByteLength]byte
	marshalBinary(out[:], uuid.a[:], uuid.getBits())
	return out[:]
}

// MarshalBinary fulfills the "encoding".BinaryMarshaler interface.
// It produces a binary representation of this UUID.
func (uuid UUID) MarshalBinary() ([]byte, error) {
	var out [ByteLength]byte
	marshalBinary(out[:], uuid.a[:], uuid.getBits())
	return out[:], nil
}

// CanonicalString returns the textual representation of this UUID in canonical "8-4-4-4-12" format.
func (uuid UUID) CanonicalString() string {
	w := makeSliceWriter(bufferLength)
	defer w.release()
	marshalTextCanonical(&w, uuid.a[:])
	return w.String()
}

// HashLikeString returns the textual representation of this UUID as a 32-character hex string.
func (uuid UUID) HashLikeString() string {
	w := makeSliceWriter(bufferLength)
	defer w.release()
	marshalTextHashLike(&w, uuid.a[:])
	return w.String()
}

// BracketedString returns the textual representation of this UUID in "{8-4-4-4-12}" format.
func (uuid UUID) BracketedString() string {
	w := makeSliceWriter(bufferLength)
	defer w.release()
	marshalTextBracketed(&w, uuid.a[:])
	return w.String()
}

// URNString returns the textual representation of this UUID in "urn:uuid:8-4-4-4-12" format.
func (uuid UUID) URNString() string {
	w := makeSliceWriter(bufferLength)
	defer w.release()
	marshalTextURN(&w, uuid.a[:])
	return w.String()
}

// DenseString returns the textual representation of this UUID in "@<base64>" format.
func (uuid UUID) DenseString() string {
	w := makeSliceWriter(bufferLength)
	defer w.release()
	marshalTextDense(&w, uuid.a[:])
	return w.String()
}

// String fulfills the "fmt".Stringer interface.
// It produces a textual representation of this UUID.
func (uuid UUID) String() string {
	w := makeSliceWriter(bufferLength)
	defer w.release()
	marshalText(&w, uuid.a[:], uuid.getBits())
	return w.String()
}

// GoString fulfills the "fmt".GoStringer interface.
// It produces a textual representation of this UUID as a snippet of Go pseudocode.
func (uuid UUID) GoString() string {
	w := makeSliceWriter(bufferLength)
	defer w.release()
	w.WriteString(`UUID("`)
	marshalText(&w, uuid.a[:], uuid.getBits())
	w.WriteString(`")`)
	return w.String()
}

// MarshalText fulfills the "encoding".TextMarshaler interface.
// It produces a textual representation of this UUID.
func (uuid UUID) MarshalText() ([]byte, error) {
	w := makeSliceWriter(bufferLength)
	defer w.release()
	marshalText(&w, uuid.a[:], uuid.getBits())
	return w.CopyBytes(), nil
}

// MarshalJSON fulfills the "encoding/json".Marshaler interface.
// It produces a textual representation of this UUID as a JSON string.
func (uuid UUID) MarshalJSON() ([]byte, error) {
	w := makeSliceWriter(bufferLength)
	defer w.release()
	w.WriteByte('"')
	marshalText(&w, uuid.a[:], uuid.getBits())
	w.WriteByte('"')
	return w.CopyBytes(), nil
}

// Format fulfills the "fmt".Formatter interface.
func (uuid UUID) Format(s fmt.State, verb rune) {
	hasPlus := s.Flag('+')
	hasSharp := s.Flag('#')
	hasMinus := s.Flag('-')
	width, hasWidth := s.Width()
	if hasWidth && width < 0 {
		panic(fmt.Errorf("width is negative: %d < 0", width))
	}
	d := formatStudy(verb, hasPlus, hasSharp, hasMinus, hasWidth, uint(width), uuid.getBits())

	w := makeSliceWriter(d.peakCount)
	defer w.release()
	formatApply(&w, uuid.a[:], d)
	s.Write(w.Bytes())
}

// Value fulfills the "database/sql/driver".Valuer interface.
// It produces a SQL representation of the UUID.
func (uuid UUID) Value() (driver.Value, error) {
	value := valueImpl(uuid.a[:], uuid.getBits())
	return value, nil
}

// FromStandardBytes attempts to parse a binary UUID representation in RFC 4122 byte order.
func (uuid *UUID) FromStandardBytes(in []byte) error {
	return unmarshalBinaryStandard("UUID", "FromStandardBytes", uuid.a[:], in)
}

// FromDenseBytes attempts to parse a binary UUID representation in "dense" byte order.
func (uuid *UUID) FromDenseBytes(in []byte) error {
	return unmarshalBinaryDense("UUID", "FromDenseBytes", uuid.a[:], in)
}

// FromBytes attempts to parse a binary UUID representation.
func (uuid *UUID) FromBytes(in []byte) error {
	return unmarshalBinary("UUID", "FromBytes", uuid.a[:], in, uuid.getBits())
}

// MustFromBytes parses a binary UUID representation, or panics if it cannot.
func (uuid *UUID) MustFromBytes(in []byte) {
	if err := unmarshalBinary("UUID", "MustFromBytes", uuid.a[:], in, uuid.getBits()); err != nil {
		panic(err)
	}
}

// UnmarshalBinary fulfills the "encoding".BinaryUnmarshaler interface.
// It attempts to parse a binary UUID representation.
func (uuid *UUID) UnmarshalBinary(in []byte) error {
	return unmarshalBinary("UUID", "UnmarshalBinary", uuid.a[:], in, uuid.getBits())
}

// FromString attempts to parse a textual UUID representation.
func (uuid *UUID) FromString(in string) error {
	return unmarshalText("UUID", "FromString", uuid.a[:], []byte(in))
}

// MustFromString parses a textual UUID representation, or panics if it cannot.
func (uuid *UUID) MustFromString(in string) {
	if err := unmarshalText("UUID", "MustFromString", uuid.a[:], []byte(in)); err != nil {
		panic(err)
	}
}

// UnmarshalText fulfills the "encoding".TextUnmarshaler interface.
// It attempts to parse a textual UUID representation.
func (uuid *UUID) UnmarshalText(in []byte) error {
	return unmarshalText("UUID", "UnmarshalText", uuid.a[:], in)
}

// UnmarshalJSON fulfills the "encoding/json".Unmarshaler interface.
// It attempts to parse a JSON value as a textual UUID representation.
func (uuid *UUID) UnmarshalJSON(in []byte) error {
	var ptr *string
	if len(in) != 0 {
		if err := json.Unmarshal(in, &ptr); err != nil {
			return makeParseError("UUID", "UnmarshalJSON", in, true).detailf("json.Unmarshal: %v", err)
		}
	}
	var str string
	if ptr != nil {
		str = *ptr
	}
	return unmarshalText("UUID", "UnmarshalJSON", uuid.a[:], []byte(str))
}

// Scan fulfills the "database/sql".Scanner interface.
// It attempts to interpret a SQL value as a UUID representation of some kind.
func (uuid *UUID) Scan(value interface{}) error {
	switch v := value.(type) {
	case nil:
		zeroBytes(uuid.a[:])
		return nil
	case []byte:
		return scanImpl("UUID", "Scan", uuid.a[:], v, false, uuid.getBits())
	case string:
		return scanImpl("UUID", "Scan", uuid.a[:], []byte(v), true, uuid.getBits())
	}
	return makeTypeError("UUID", "Scan", value, nil, []byte(nil), "")
}
