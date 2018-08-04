package uuid

import "fmt"

// ValueMode selects the output behavior of the SQL-oriented Value method.
type ValueMode byte

// ValueMode enum constants.
const (
	_ ValueMode = iota

	// Text: output is a string in the current TextMode.
	Text

	// Binary: output is a byte slice in the current BinaryMode.
	Binary
)

func (vm ValueMode) asByte() byte {
	if ch, found := vmMap[vm]; found {
		return ch
	}
	return '!'
}

func (vm ValueMode) String() string {
	return string(vm.asByte())
}

// BinaryMode selects the input/output format for parsing/producing binary represenatations.
type BinaryMode byte

// BinaryMode enum constants.
const (
	_ BinaryMode = iota

	// StandardOnly: produce RFC 4122 bytes, parse RFC 4122 bytes.
	StandardOnly

	// StandardFirst: produce RFC 4122 bytes, parse either byte ordering.
	// Ambiguous inputs will be interpreted as RFC 4122.
	StandardFirst

	// DenseOnly: produce "dense" bytes, parse "dense" bytes.
	DenseOnly

	// DenseFirst: produce "dense" bytes, parse either byte ordering.
	// Ambiguous inputs will be interpreted as "dense".
	DenseFirst
)

func (bm BinaryMode) asByte() byte {
	if ch, found := bmMap[bm]; found {
		return ch
	}
	return '!'
}

func (bm BinaryMode) String() string {
	return string(bm.asByte())
}

// TextMode selects the output format for producing textual representations.
type TextMode byte

// TextMode enum constants.
const (
	_ TextMode = iota

	// Dense: "@<base64>", e.g. "@EeiKtHe5nOqWqBheD61jNQ"
	Dense

	// Canonical: "8-4-4-4-12", e.g. "77b99cea-8ab4-11e8-96a8-185e0fad6335"
	Canonical

	// HashLike: raw hex digits, e.g. "77b99cea8ab411e896a8185e0fad6335"
	HashLike

	// Bracketed: "{8-4-4-4-12}", e.g. "{77b99cea-8ab4-11e8-96a8-185e0fad6335}"
	Bracketed

	// URN: URN format, e.g. "urn:uuid:77b99cea-8ab4-11e8-96a8-185e0fad6335"
	URN
)

func (tm TextMode) asByte() byte {
	if ch, found := tmMap[tm]; found {
		return ch
	}
	return '!'
}

func (tm TextMode) String() string {
	return string(tm.asByte())
}

// Preferences represents a combination of modes to apply to a UUID.
type Preferences struct {
	Value  ValueMode
	Binary BinaryMode
	Text   TextMode
}

// Nil returns a Nil-valued UUID with these preferences.
func (pref Preferences) Nil() UUID {
	var uuid UUID
	uuid.SetPreferences(pref)
	return uuid
}

// New returns a newly generated V1 UUID with these preferences.
func (pref Preferences) New() UUID {
	var uuid UUID
	uuid.SetPreferences(pref)
	uuid.SetNew()
	return uuid
}

// FromBytes attempts to parse a binary UUID representation using these preferences.
func (pref Preferences) FromBytes(in []byte) (UUID, error) {
	var uuid UUID
	uuid.SetPreferences(pref)
	err := unmarshalBinary("Preferences", "FromBytes", uuid.a[:], in, uuid.getBits())
	return uuid, err
}

// MustFromBytes parses a binary UUID representation using these preferences, or panics if it cannot.
func (pref Preferences) MustFromBytes(in []byte) UUID {
	var uuid UUID
	uuid.SetPreferences(pref)
	err := unmarshalBinary("Preferences", "MustFromBytes", uuid.a[:], in, uuid.getBits())
	if err != nil {
		panic(err)
	}
	return uuid
}

// FromString attempts to parse a textual UUID representation with these preferences.
func (pref Preferences) FromString(in string) (UUID, error) {
	var uuid UUID
	uuid.SetPreferences(pref)
	err := unmarshalText("Preferences", "FromString", uuid.a[:], []byte(in))
	return uuid, err
}

// MustFromString parses a textual UUID representation with these preferences, or panics if it cannot.
func (pref Preferences) MustFromString(in string) UUID {
	var uuid UUID
	uuid.SetPreferences(pref)
	err := unmarshalText("Preferences", "MustFromString", uuid.a[:], []byte(in))
	if err != nil {
		panic(err)
	}
	return uuid
}

func (pref Preferences) String() string {
	var buf [3]byte
	buf[0] = pref.Value.asByte()
	buf[1] = pref.Binary.asByte()
	buf[2] = pref.Text.asByte()
	return string(buf[:])
}

func (pref Preferences) collapse(old bits) (x bits) {
	x = bitValid

	switch pref.Value {
	case 0:
		x |= old.just(bitValueIsBinary)
	case Text:
		// pass
	case Binary:
		x |= bitValueIsBinary
	default:
		panic(fmt.Errorf("unknown value ValueMode(%d)", pref.Value))
	}

	switch pref.Binary {
	case 0:
		x |= old.just(bitsBinary)
	case StandardOnly:
		// pass
	case StandardFirst:
		x |= bitBinaryIsLoose
	case DenseOnly:
		x |= bitBinaryIsDense
	case DenseFirst:
		x |= bitBinaryIsDense | bitBinaryIsLoose
	default:
		panic(fmt.Errorf("unknown value BinaryMode(%d)", pref.Binary))
	}

	switch pref.Text {
	case 0:
		x |= old.just(bitsText)
	case Canonical:
		x |= textModeCanonical
	case HashLike:
		x |= textModeHashLike
	case Bracketed:
		x |= textModeBracketed
	case URN:
		x |= textModeURN
	case Dense:
		x |= bitTextIsDense
	default:
		panic(fmt.Errorf("unknown value TextMode(%d)", pref.Text))
	}

	return
}

type bits byte

const (
	bitValid         bits = 0x80
	bitValueIsBinary bits = 0x40
	bitBinaryIsDense bits = 0x20
	bitBinaryIsLoose bits = 0x10
	bitTextIsDense   bits = 0x08
	bitTextIsModeX   bits = 0x04
	bitTextIsModeY   bits = 0x02
	bitReserved      bits = 0x01
)

const (
	bitsDefault    = bitValid | bitBinaryIsDense | bitBinaryIsLoose | bitTextIsDense
	bitsBinary     = bitBinaryIsDense | bitBinaryIsLoose
	bitsText       = bitTextIsDense | bitTextIsModeX | bitTextIsModeY
	bitsTextIsMode = bitTextIsModeX | bitTextIsModeY
)

const (
	textModeCanonical bits = 0
	textModeHashLike  bits = bitTextIsModeX
	textModeBracketed bits = bitTextIsModeY
	textModeURN       bits = bitTextIsModeX | bitTextIsModeY
)

func (x bits) String() string {
	return x.expand().String()
}

func (x bits) GoString() string {
	return fmt.Sprintf("%08b", byte(x))
}

func (x bits) just(y bits) bits {
	return (x & y)
}

func (x bits) has(y bits) bool {
	return (x & y) == y
}

func (x bits) defaulted() bits {
	if x.has(bitValid) {
		return x
	}
	return bitsDefault
}

func (x bits) expand() (pref Preferences) {
	switch x.just(bitValueIsBinary) {
	case 0:
		pref.Value = Text
	case bitValueIsBinary:
		pref.Value = Binary
	}

	switch x.just(bitsBinary) {
	case 0:
		pref.Binary = StandardOnly
	case bitBinaryIsLoose:
		pref.Binary = StandardFirst
	case bitBinaryIsDense:
		pref.Binary = DenseOnly
	case bitBinaryIsDense | bitBinaryIsLoose:
		pref.Binary = DenseFirst
	}

	switch x.just(bitsText) {
	case textModeCanonical:
		pref.Text = Canonical
	case textModeHashLike:
		pref.Text = HashLike
	case textModeBracketed:
		pref.Text = Bracketed
	case textModeURN:
		pref.Text = URN

	case bitTextIsDense | textModeCanonical:
		fallthrough
	case bitTextIsDense | textModeHashLike:
		fallthrough
	case bitTextIsDense | textModeBracketed:
		fallthrough
	case bitTextIsDense | textModeURN:
		pref.Text = Dense
	}

	return
}

func (x bits) textLength() (used uint, peak uint) {
	switch x.just(bitsText) {
	case textModeCanonical:
		return 36, 36

	case textModeHashLike:
		return 32, 32

	case textModeBracketed:
		return 38, 38

	case textModeURN:
		return 45, 45

	default:
		return 23, 25
	}
}

var vmMap = map[ValueMode]byte{
	0:      '-',
	Text:   'T',
	Binary: 'B',
}

var bmMap = map[BinaryMode]byte{
	0:             '-',
	StandardOnly:  'S',
	StandardFirst: 's',
	DenseOnly:     'D',
	DenseFirst:    'd',
}

var tmMap = map[TextMode]byte{
	0:         '-',
	Dense:     'D',
	Canonical: 'C',
	HashLike:  'H',
	Bracketed: 'B',
	URN:       'U',
}
