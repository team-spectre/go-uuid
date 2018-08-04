package uuid

import (
	"database/sql/driver"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
)

var (
	atSign       = []byte(`@`)
	urnPrefix    = []byte(`urn:uuid:`)
	openBracket  = []byte(`{`)
	closeBracket = []byte(`}`)
)

var allZeroes [ByteLength]byte

type pair struct{ i, j uint }

var hashlikePairs = []pair{
	{0, 16},
}

var canonicalPairs = []pair{
	{0, 4},
	{4, 6},
	{6, 8},
	{8, 10},
	{10, 16},
}

var spaceSet = [8]uint32{
	0x007c0000, // 0x00..0x1f -> false except 0x09..0x0d
	0x80000000, // 0x20..0x3f -> false except 0x20
	0x00000000, // 0x40..0x5f -> false
	0x00000000, // 0x60..0x7f -> false
	0x00000000, // 0x80..0x9f -> false
	0x00000000, // 0xa0..0xbf -> false
	0x00000000, // 0xc0..0xdf -> false
	0x00000000, // 0xe0..0xff -> false
}

var badSet = [8]uint32{
	0xff83ffff, // 0x00..0x1f -> true except 0x09..0x0d
	0x00000000, // 0x20..0x3f -> false
	0x00000000, // 0x40..0x5f -> false
	0x00000001, // 0x60..0x7f -> false except 0x7f
	0xffffffff, // 0x80..0x9f -> true
	0xffffffff, // 0xa0..0xbf -> true
	0xffffffff, // 0xc0..0xdf -> true
	0xffffffff, // 0xe0..0xff -> true
}

func lookupByte(a [8]uint32, ch byte) bool {
	i := uint(ch >> 5)
	j := uint32(1) << (31 - (ch & 0x1f))
	return (a[i] & j) == j
}

func isSpace(ch byte) bool {
	return lookupByte(spaceSet, ch)
}

func isText(in []byte) bool {
	i, j := 0, len(in)
	for i < j {
		ch := in[i]
		i++
		if lookupByte(badSet, ch) {
			return false
		}
	}
	return true
}

func quoteString(in []byte) string {
	return fmt.Sprintf("%q", string(in))
}

func formatBytes(in []byte) string {
	w := makeSliceWriter(3*uint(len(in)) + 2)
	defer w.release()

	w.WriteByte('[')
	for i := range in {
		if i > 0 {
			w.WriteByte(' ')
		}
		slice := w.Grab(2)
		hex.Encode(slice, in[i:i+1])
	}
	w.WriteByte(']')
	return w.String()
}

func zeroBytes(out []byte) {
	for i := range out {
		out[i] = 0
	}
}

func equalBytes(p, q []byte) bool {
	if len(p) != len(q) {
		return false
	}
	for i := range p {
		if p[i] != q[i] {
			return false
		}
	}
	return true
}

func copyBytes(in []byte) []byte {
	out := make([]byte, len(in))
	copy(out, in)
	return out
}

func isZero(in []byte) bool {
	for _, b := range in {
		if b != 0 {
			return false
		}
	}
	return true
}

func importStandard(out, in []byte) {
	copy(out, in)
}

func exportStandard(out, in []byte) {
	copy(out, in)
}

func importDense(out, in []byte) {
	out[0] = in[4]
	out[1] = in[5]
	out[2] = in[6]
	out[3] = in[7]
	out[4] = in[2]
	out[5] = in[3]
	out[6] = in[0]
	out[7] = in[1]
	copy(out[8:16], in[8:16])
}

func exportDense(out, in []byte) {
	out[0] = in[6]
	out[1] = in[7]
	out[2] = in[4]
	out[3] = in[5]
	out[4] = in[0]
	out[5] = in[1]
	out[6] = in[2]
	out[7] = in[3]
	copy(out[8:16], in[8:16])
}

func extract(in []byte) (versionStandard Version, versionDense Version, variant Variant) {
	if len(in) == ByteLength && !isZero(in) {
		versionStandard = Version((in[6] & 0xf0) >> 4)
		versionDense = Version((in[0] & 0xf0) >> 4)
		switch (in[8] & 0xe0) >> 5 {
		case 0x00, 0x01, 0x02, 0x03:
			// 0xxz zzzz -> 0xx -> 000 001 010 011
			variant = VariantNCS
		case 0x04, 0x05:
			// 10xz zzzz -> 10x -> 100 101
			variant = VariantRFC4122
		case 0x06:
			// 110z zzzz -> 110
			variant = VariantMicrosoft
		case 0x07:
			// 111z zzzz -> 111
			variant = VariantFuture
		}
	}
	return
}

func pickStandard(version, _ Version) Version {
	return version
}

func pickDense(_, version Version) Version {
	return version
}

func marshalBinary(out, in []byte, x bits) {
	if x.has(bitBinaryIsDense) {
		marshalBinaryDense(out, in)
		return
	}
	marshalBinaryStandard(out, in)
}

func marshalBinaryStandard(out, in []byte) {
	exportStandard(out, in)
}

func marshalBinaryDense(out, in []byte) {
	exportDense(out, in)
}

func marshalText(w *sliceWriter, in []byte, x bits) {
	if x.has(bitTextIsDense) {
		marshalTextDense(w, in)
		return
	}
	switch x.just(bitTextIsModeX | bitTextIsModeY) {
	case textModeCanonical:
		marshalTextCanonical(w, in)
	case textModeHashLike:
		marshalTextHashLike(w, in)
	case textModeBracketed:
		marshalTextBracketed(w, in)
	case textModeURN:
		marshalTextURN(w, in)
	}
}

func marshalTextCanonical(w *sliceWriter, in []byte) {
	marshalTextHelper(w, in, "", "", canonicalPairs)
}

func marshalTextHashLike(w *sliceWriter, in []byte) {
	marshalTextHelper(w, in, "", "", hashlikePairs)
}

func marshalTextBracketed(w *sliceWriter, in []byte) {
	marshalTextHelper(w, in, "{", "}", canonicalPairs)
}

func marshalTextURN(w *sliceWriter, in []byte) {
	marshalTextHelper(w, in, "urn:uuid:", "", canonicalPairs)
}

func marshalTextHelper(w *sliceWriter, in []byte, pre, post string, pairs []pair) {
	var tmp [ByteLength]byte
	exportStandard(tmp[:], in)

	w.WriteString(pre)
	for i, p := range pairs {
		if i > 0 {
			w.WriteByte('-')
		}
		slice := w.Grab(2 * (p.j - p.i))
		hex.Encode(slice, tmp[p.i:p.j])
	}
	w.WriteString(post)
}

func marshalTextDense(w *sliceWriter, in []byte) {
	var tmp [ByteLength]byte
	exportDense(tmp[:], in)

	// ceil(16 / 3) * 3 -> 18  for padding
	// 18 * (4/3)       -> 24  for base-64 encoded length
	// 24 + 1           -> 25  for '@' prefix
	slice := w.Grab(25)
	slice[0] = '@'
	base64.StdEncoding.Encode(slice[1:], tmp[:])
	w.Unwrite(2) // trim unnecessary "==" suffix
}

func valueImpl(in []byte, x bits) driver.Value {
	if x.has(bitValueIsBinary) {
		var out [ByteLength]byte
		marshalBinary(out[:], in, x)
		return out[:]
	}
	w := makeSliceWriter(bufferLength)
	defer w.release()
	marshalText(&w, in, x)
	return w.String()
}

func unmarshalBinary(typeName, methodName string, out, in []byte, x bits) error {
	preferDense := false
	switch x.just(bitsBinary) {
	case 0:
		f := pickStandard
		g := importStandard
		return unmarshalBinaryHelper(typeName, methodName, out, in, f, g)

	case bitBinaryIsDense:
		f := pickDense
		g := importDense
		return unmarshalBinaryHelper(typeName, methodName, out, in, f, g)

	case bitBinaryIsDense | bitBinaryIsLoose:
		preferDense = true
	}

	versionStandard, versionDense, _ := extract(in)
	validStandard := versionStandard.IsValid()
	validDense := versionDense.IsValid()

	// +-----+---------------+------------+-------------+----------------------+
	// | row | validStandard | validDense | preferDense | outcome              |
	// +-----+---------------+------------+-------------+----------------------+
	// | A   | false         | false      | false       | standard (forced)    |
	// | B   | false         | false      | true        | dense    (forced)    |
	// | C   | false         | true       | false       | dense                |
	// | D   | false         | true       | true        | dense                |
	// | E   | true          | false      | false       | standard             |
	// | F   | true          | false      | true        | standard             |
	// | G   | true          | true       | false       | standard if p0 >= p1 |
	// | H   | true          | true       | true        | dense    if p1 >= p0 |
	// +-----+---------------+------------+-------------+----------------------+

	if !validStandard && (validDense || preferDense) {
		// B, C, D
		f := pickDense
		g := importDense
		return unmarshalBinaryHelper(typeName, methodName, out, in, f, g)
	}

	if !validDense {
		// A, E, F [B already covered]
		f := pickStandard
		g := importStandard
		return unmarshalBinaryHelper(typeName, methodName, out, in, f, g)
	}

	// Only G and H remain

	timeStandard := binary.BigEndian.Uint64([]byte{
		in[6],
		in[7],
		in[4],
		in[5],
		in[0],
		in[1],
		in[2],
		in[3],
	})
	timeStandard &= tickMask

	timeDense := binary.BigEndian.Uint64(in[0:8])
	timeDense &= tickMask

	// Timestamp was generated between [1970-01-01] and [now + 5 years]? 0 < p < 1
	// Timestamp out of range? p = 0
	now := systemTick()
	t0 := uint64(tickEpoch)
	t1 := now + (5 * tickYear)

	var probStandard, probDense float64
	if timeStandard >= t0 && timeStandard <= t1 {
		probStandard = tickProbability(timeStandard, now)
	}
	if timeDense >= t0 && timeDense <= t1 {
		probDense = tickProbability(timeDense, now)
	}

	// Bias the preferred interpretation very slightly.
	// Only swings things when the larger is no more than 17/16ths of the smaller.
	bias := math.Min(probStandard, probDense) * 0.0625
	if preferDense {
		probDense += bias
	} else {
		probStandard += bias
	}

	if probStandard < probDense || (probStandard == probDense && preferDense) {
		f := pickDense
		g := importDense
		return unmarshalBinaryHelper(typeName, methodName, out, in, f, g)
	}

	f := pickStandard
	g := importStandard
	return unmarshalBinaryHelper(typeName, methodName, out, in, f, g)
}

func unmarshalBinaryStandard(typeName, methodName string, out, in []byte) error {
	f := pickStandard
	g := importStandard
	return unmarshalBinaryHelper(typeName, methodName, out, in, f, g)
}

func unmarshalBinaryDense(typeName, methodName string, out, in []byte) error {
	f := pickDense
	g := importDense
	return unmarshalBinaryHelper(typeName, methodName, out, in, f, g)
}

func unmarshalBinaryHelper(
	typeName string,
	methodName string,
	out []byte,
	in []byte,
	f func(_, _ Version) Version,
	g func(_, _ []byte),
) error {
	if len(in) == 0 || equalBytes(allZeroes[:], in) {
		zeroBytes(out)
		return nil
	}

	versionStandard, versionDense, variant := extract(in)
	version := f(versionStandard, versionDense)
	if !version.IsValid() {
		return makeParseError(typeName, methodName, in, false).detailf("expected version V1-V5, got %s", version)
	}
	if !variant.IsValid() {
		return makeParseError(typeName, methodName, in, false).detailf("expected VariantRFC4122, got %s", variant)
	}
	g(out, in)
	return nil
}

func unmarshalText(typeName, methodName string, out, in []byte) error {
	r := makeSliceReader(in)
	if r.IsEOF() {
		zeroBytes(out)
		return nil
	}
	r.TrimLeading(isSpace)
	if r.HasPrefix(atSign) {
		r.TrimPrefix(uint(len(atSign)))
		r.TrimLeading(isSpace)
		return unmarshalTextDense(typeName, methodName, out, &r)
	}
	if r.HasPrefix(urnPrefix) {
		r.TrimPrefix(uint(len(urnPrefix)))
		r.TrimLeading(isSpace)
	}
	needClose := false
	if r.HasPrefix(openBracket) {
		r.TrimPrefix(uint(len(openBracket)))
		r.TrimLeading(isSpace)
		needClose = true
	}
	return unmarshalTextStandard(typeName, methodName, out, &r, needClose)
}

func unmarshalTextStandard(typeName, methodName string, out []byte, r *sliceReader, needClose bool) error {
	// Why did we roll our own here? Better error messages.

	var tmp [ByteLength]byte
	w := sliceWriter{slice: tmp[:], i: 0, j: ByteLength}

	partialValue := byte(0)
	partialCount := uint(0)
	absorb := func(next byte) {
		partialValue = (partialValue << 4) | next
		partialCount++
		if partialCount > 1 {
			w.WriteByte(partialValue)
			partialValue = 0
			partialCount = 0
		}
	}

	allowAny := true
	allowBracket := needClose
	for !r.IsEOF() {
		if w.Remain() <= 0 {
			allowAny = false
		}

		ch := r.ReadByte()
		if allowAny && ch >= '0' && ch <= '9' {
			absorb(ch - '0' + 0x00)
			continue
		}
		if allowAny && ch >= 'A' && ch <= 'F' {
			absorb(ch - 'A' + 0x0a)
			continue
		}
		if allowAny && ch >= 'a' && ch <= 'f' {
			absorb(ch - 'a' + 0x0a)
			continue
		}
		if allowAny && ch == '-' {
			continue
		}
		if allowBracket && ch == '}' {
			allowAny = false
			allowBracket = false
			continue
		}
		if isSpace(ch) {
			r.TrimLeading(isSpace)
			continue
		}

		var expected string
		if allowAny {
			expected = "0-9, A-F, a-f, or -"
		} else if allowBracket {
			expected = "}"
		} else {
			expected = "end of input"
		}
		i := r.CurrentOffset() - 1
		in := r.slice[0:r.j]
		return makeParseError(typeName, methodName, in, true).detailf("unexpected byte %q %#02x at position %d, expected %s", ch, ch, i, expected)
	}

	if w.Remain() > 0 {
		i := r.CurrentOffset()
		in := r.slice[0:r.j]
		more := w.Remain()*2 - partialCount
		return makeParseError(typeName, methodName, in, true).detailf("unexpected end of input at position %d, expected %d more hex digits", i, more)
	}

	if allowBracket {
		i := r.CurrentOffset()
		in := r.slice[0:r.j]
		return makeParseError(typeName, methodName, in, true).detailf("unexpected end of input at position %d, expected '}'", i)
	}

	importStandard(out, w.Bytes())
	return nil
}

func unmarshalTextDense(typeName, methodName string, out []byte, r *sliceReader) error {
	// Why did we roll our own here? Better error messages.

	var tmp [18]byte
	w := sliceWriter{slice: tmp[:], i: 0, j: 18}

	partialValue := uint32(0)
	partialCount := uint(0)
	absorb := func(next byte) {
		partialValue = (partialValue << 6) | uint32(next)
		partialCount++
		if partialCount >= 4 {
			slice := w.Grab(3)
			slice[0] = byte(partialValue >> 16)
			slice[1] = byte(partialValue >> 8)
			slice[2] = byte(partialValue >> 0)
			partialValue = 0
			partialCount = 0
		}
	}

	allowAny := true
	allowEqual := true
	for !r.IsEOF() {
		if w.Remain() <= 0 {
			allowAny = false
			allowEqual = false
		}

		ch := r.ReadByte()
		if allowAny && ch >= 'A' && ch <= 'Z' {
			absorb(ch - 'A' + 0x0000)
			continue
		}
		if allowAny && ch >= 'a' && ch <= 'z' {
			absorb(ch - 'a' + 0x001a)
			continue
		}
		if allowAny && ch >= '0' && ch <= '9' {
			absorb(ch - '0' + 0x0034)
			continue
		}
		if allowAny && (ch == '+' || ch == '-') {
			absorb(0x3e)
			continue
		}
		if allowAny && (ch == '/' || ch == '_') {
			absorb(0x3f)
			continue
		}
		if allowEqual && ch == '=' {
			absorb(0x00)
			allowAny = false
			if partialCount == 0 {
				allowEqual = false
			}
			continue
		}
		if isSpace(ch) {
			r.TrimLeading(isSpace)
			continue
		}

		var expected string
		if allowAny {
			expected = "A-Z, a-z, 0-9, +, /, -, _, or ="
		} else if allowEqual {
			expected = "="
		} else {
			expected = "end of input"
		}
		i := r.CurrentOffset() - 1
		in := r.slice[0:r.j]
		return makeParseError(typeName, methodName, in, true).detailf("unexpected byte %q %#02x at position %d, expected %s", ch, ch, i, expected)
	}

	// Allow incomplete base-64 sequences
	for partialCount > 0 {
		absorb(0x00)
	}

	if w.Remain() > 0 {
		i := r.CurrentOffset()
		in := r.slice[0:r.j]
		more := (w.Remain()/3)*4 - partialCount
		return makeParseError(typeName, methodName, in, true).detailf("unexpected end of input at position %d, expected %d more base-64 digits", i, more)
	}

	slice := w.Bytes()
	if slice[16] != 0 || slice[17] != 0 {
		got := base64.StdEncoding.EncodeToString(slice[15:18])
		slice[16] = 0
		slice[17] = 0
		expect := base64.StdEncoding.EncodeToString(slice[15:18])
		in := r.slice[0:r.j]
		return makeParseError(typeName, methodName, in, true).detailf("unexpected data at end of input, expected %q but got %q", expect, got)
	}

	importDense(out, slice[0:ByteLength])
	return nil
}

func scanImpl(typeName, methodName string, out, in []byte, isDefinitelyText bool, x bits) error {
	if len(in) == 0 {
		zeroBytes(out)
		return nil
	}
	if isDefinitelyText || (len(in) != ByteLength && isText(in)) {
		return unmarshalText(typeName, methodName, out, in)
	}
	return unmarshalBinary(typeName, methodName, out, in, x)
}

func tickProbability(value, now uint64) float64 {
	x := float64(now - value)
	if value > now {
		x = float64(value - now)
	}
	x /= tickCentury

	const a float64 = 1.0 / (math.Sqrt2 * math.SqrtPi)
	b := -0.5 * math.Pow(x, 2.0)
	return a * math.Exp(b)
}
