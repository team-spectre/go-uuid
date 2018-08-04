package uuid

import (
	"encoding/hex"
	"fmt"
	"io"
	"math/rand"
	"net"
	"sync"
	"testing"
	"time"
)

type order int8

const (
	LT order = -1
	EQ order = 0
	GT order = 1
)

func TestTickCompare(t *testing.T) {
	const (
		a   = 0
		ap1 = 1
		ap2 = 2

		bm2 = tickCentury - 2
		bm1 = tickCentury - 1
		b   = tickCentury
		bp1 = tickCentury + 1
		bp2 = tickCentury + 2

		cm2 = tickEpoch - tickCentury - 2
		cm1 = tickEpoch - tickCentury - 1
		c   = tickEpoch - tickCentury
		cp1 = tickEpoch - tickCentury + 1
		cp2 = tickEpoch - tickCentury + 2

		dm2 = tickEpoch - 2
		dm1 = tickEpoch - 1
		d   = tickEpoch
		dp1 = tickEpoch + 1
		dp2 = tickEpoch + 2

		em2 = tickEpoch + tickCentury - 2
		em1 = tickEpoch + tickCentury - 1
		e   = tickEpoch + tickCentury
		ep1 = tickEpoch + tickCentury + 1
		ep2 = tickEpoch + tickCentury + 2

		fm2 = tickMask - tickCentury - 2
		fm1 = tickMask - tickCentury - 1
		f   = tickMask - tickCentury
		fp1 = tickMask - tickCentury + 1
		fp2 = tickMask - tickCentury + 2

		gm2 = tickMask - 2
		gm1 = tickMask - 1
		g   = tickMask
	)
	var names = map[uint64]string{
		a:   "a",
		ap1: "ap1",
		ap2: "ap2",

		bm2: "bm2",
		bm1: "bm1",
		b:   "b",
		bp1: "bp1",
		bp2: "bp2",

		cm2: "cm2",
		cm1: "cm1",
		c:   "c",
		cp1: "cp1",
		cp2: "cp2",

		dm2: "dm2",
		dm1: "dm1",
		d:   "d",
		dp1: "dp1",
		dp2: "dp2",

		em2: "em2",
		em1: "em1",
		e:   "e",
		ep1: "ep1",
		ep2: "ep2",

		fm2: "fm2",
		fm1: "fm1",
		f:   "f",
		fp1: "fp1",
		fp2: "fp2",

		gm2: "gm2",
		gm1: "gm1",
		g:   "g",
	}
	name := func(t uint64) string {
		if s, found := names[t]; found {
			return s
		}
		return fmt.Sprintf("%d", t)
	}

	type testrow struct {
		t0 uint64
		t1 uint64
		o  order
	}
	data := []testrow{
		{a, g, GT},
		{a, a, EQ},
		{a, ap1, LT},
		{a, b, LT},
		{a, c, LT},
		{a, dm2, LT},
		{a, dm1, LT},
		{a, d, GT},

		{cm1, cm2, GT},
		{cm1, cm1, EQ},
		{cm1, c, LT},
		{cm1, dm1, LT},
		{cm1, d, GT},

		{c, cm1, GT},
		{c, c, EQ},
		{c, cp1, LT},
		{c, dm1, LT},
		{c, d, LT},
		{c, dp1, GT},

		{cp1, c, GT},
		{cp1, cp1, EQ},
		{cp1, cp2, LT},
		{cp1, d, LT},
		{cp1, dp1, LT},
		{cp1, dp2, GT},

		{dm1, dm2, GT},
		{dm1, dm1, EQ},
		{dm1, d, LT},
		{dm1, dp1, LT},
		{dm1, em1, LT},
		{dm1, e, GT},

		{d, dm1, GT},
		{d, d, EQ},
		{d, dp1, LT},
		{d, e, LT},
		{d, g, LT},
		{d, a, LT},
		{d, cm1, LT},
		{d, c, GT},

		{dp1, d, GT},
		{dp1, dp1, EQ},
		{dp1, dp2, LT},
		{dp1, e, LT},
		{dp1, f, LT},
		{dp1, g, LT},
		{dp1, a, LT},
		{dp1, b, LT},
		{dp1, c, LT},
		{dp1, cp1, GT},

		{em1, em2, GT},
		{em1, em1, EQ},
		{em1, e, LT},
		{em1, f, LT},
		{em1, g, LT},
		{em1, a, LT},
		{em1, b, LT},
		{em1, c, LT},
		{em1, dm2, LT},
		{em1, dm1, GT},

		{e, em1, GT},
		{e, e, EQ},
		{e, ep1, LT},
		{e, f, LT},
		{e, g, LT},
		{e, a, LT},
		{e, b, LT},
		{e, c, LT},
		{e, dm1, LT},
		{e, d, GT},

		{ep1, e, GT},
		{ep1, ep1, EQ},
		{ep1, ep2, LT},
		{ep1, f, LT},
		{ep1, g, LT},
		{ep1, a, LT},
		{ep1, b, LT},
		{ep1, c, LT},
		{ep1, dm1, LT},
		{ep1, d, GT},

		{f, fm1, GT},
		{f, f, EQ},
		{f, fp1, LT},
		{f, g, LT},
		{f, a, LT},
		{f, b, LT},
		{f, c, LT},
		{f, dm1, LT},
		{f, d, GT},

		{g, gm1, GT},
		{g, g, EQ},
		{g, a, LT},
		{g, ap1, LT},
		{g, b, LT},
		{g, c, LT},
		{g, dm1, LT},
		{g, d, GT},
	}

	for _, row := range data {
		t0, t1 := row.t0, row.t1
		n0, n1 := name(t0), name(t1)
		t.Run(fmt.Sprintf("%s/%s", n0, n1), func(t *testing.T) {
			switch row.o {
			case EQ:
				if t0 != t1 {
					panic(fmt.Errorf("bad data! {%s, %s, EQ} but t0 != t1", n0, n1))
				}
				if isLess(t0, t1) {
					t.Error("!(t0 < t1) but isLess returns true")
				}
				if isLess(t1, t0) {
					t.Error("!(t1 < t0) but isLess returns true")
				}

			case LT:
				if t0 == t1 {
					panic(fmt.Errorf("bad data! {%s, %s, LT} but t0 == t1", n0, n1))
				}
				if !isLess(t0, t1) {
					t.Error("t0 < t1 but isLess returns false")
				}
				if isLess(t1, t0) {
					t.Error("!(t1 < t0) but isLess returns true")
				}

			case GT:
				if t0 == t1 {
					panic(fmt.Errorf("bad data! {%s, %s, GT} but t0 == t1", n0, n1))
				}
				if isLess(t0, t1) {
					t.Error("!(t0 < t1) but isLess returns true")
				}
				if !isLess(t1, t0) {
					t.Error("t1 < t0 but isLess returns false")
				}
			}
		})
	}
}

func TestGenerate(t *testing.T) {
	var clock uint64 = tickEpoch + 15306624000000000
	f := func() uint64 { return clock }
	s := uint16(0x3ffe)
	a := [6]byte{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}

	advance := func(n uint64) {
		clock = (clock + n) & tickMask
	}

	rewind := func(n uint64) {
		tt := clock
		for n > tt {
			tt += (1 << 60)
		}
		tt -= n
		clock = tt & tickMask
	}

	format := func(in [ByteLength]byte) string {
		var out [36]byte
		w := sliceWriter{slice: out[:], i: 0, j: 36}
		marshalTextCanonical(&w, in[:])
		return w.String()
	}

	state := newState(f, s, a)

	test := func(expected string) {
		var buf [ByteLength]byte
		state.generate(buf[:])
		actual := format(buf)
		if expected != actual {
			t.Errorf("expected %s, got %s", expected, actual)
		}
	}

	test("31bc4000-7f1d-11e8-bffe-aabbccddeeff")
	test("31bc4000-7f1d-11e8-bfff-aabbccddeeff")
	test("31bc4000-7f1d-11e8-8000-aabbccddeeff")
	test("31bc4000-7f1d-11e8-8001-aabbccddeeff")
	advance(1)
	test("31bc4001-7f1d-11e8-8001-aabbccddeeff")
	test("31bc4001-7f1d-11e8-8002-aabbccddeeff")
	advance(1)
	test("31bc4002-7f1d-11e8-8002-aabbccddeeff")
	test("31bc4002-7f1d-11e8-8003-aabbccddeeff")
	rewind(2)
	test("31bc4002-7f1d-11e8-8004-aabbccddeeff")
	test("31bc4002-7f1d-11e8-8005-aabbccddeeff")
	state.sequence = 0x3ffc
	test("31bc4002-7f1d-11e8-bffd-aabbccddeeff")
	test("31bc4003-7f1d-11e8-bffe-aabbccddeeff")
	test("31bc4003-7f1d-11e8-bfff-aabbccddeeff")
}

func TestIsSuitable(t *testing.T) {
	type testrow struct {
		input    []byte
		expected bool
	}
	data := []testrow{
		{
			input:    []byte{0x54, 0xee, 0x75, 0x81, 0x2f, 0xc9},
			expected: true,
		},
		{
			input:    []byte{0x54, 0xee, 0x75, 0x81, 0x2f},
			expected: false,
		},
		{
			input:    []byte{0x54, 0xee, 0x75, 0x81, 0x2f, 0xc9, 0x00},
			expected: false,
		},
		{
			input:    []byte{0x41, 0x01, 0x0a, 0x80, 0x00, 0x02},
			expected: false,
		},
		{
			input:    []byte{0x42, 0x01, 0x0a, 0x80, 0x00, 0x02},
			expected: false,
		},
		{
			input:    []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			expected: false,
		},
	}

	for _, row := range data {
		name := hex.EncodeToString(row.input)
		t.Run(name, func(t *testing.T) {
			actual := isSuitable(row.input)
			if row.expected != actual {
				t.Errorf("expected isSuitable=%v, got %v", row.expected, actual)
			}
		})
	}
}

func TestSystemGenerate(t *testing.T) {
	g := globalState()
	p := make([]byte, ByteLength)
	g.generate(p)
	if !isSystemGenerated(p) {
		t.Errorf("flubbed generate: got %#02x", p)
	}
}

func isSystemGenerated(actual []byte) bool {
	expect := [ByteLength]byte{
		0xb5, 0xba, 0x40, 0x00,
		0xee, 0x86, 0x11, 0xe7,
		0xff, 0xff, 0x54, 0xee,
		0x75, 0x81, 0x2f, 0xc9,
	}
	if len(expect) != len(actual) {
		return false
	}
	for i, x := range expect {
		y := actual[i]
		if x != y && x != 0xff {
			return false
		}
	}
	return true
}

type mockRandomReader struct {
	mu  sync.Mutex
	rnd *rand.Rand
}

func newMockRandomReader(seed int64) io.Reader {
	return &mockRandomReader{
		rnd: rand.New(rand.NewSource(seed)),
	}
}

func (r *mockRandomReader) Read(p []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.rnd.Read(p)
}

func mockInterfaces() ([]net.Interface, error) {
	return []net.Interface{
		{
			Index:        1,
			MTU:          65536,
			Name:         "lo",
			HardwareAddr: net.HardwareAddr{},
			Flags:        net.FlagUp | net.FlagLoopback,
		},
		{
			Index:        2,
			MTU:          1500,
			Name:         "eth0",
			HardwareAddr: net.HardwareAddr{0x54, 0xee, 0x75, 0x81, 0x2f, 0xc9},
			Flags:        net.FlagUp | net.FlagBroadcast | net.FlagMulticast,
		},
	}, nil
}

func mockNow() time.Time {
	return time.Unix(1514764800, 0) // 2018-01-01 00:00:00Z
}

func init() {
	gInterfaces = ifaceFunc(mockInterfaces)
	gNow = nowFunc(mockNow)
	gReader = newMockRandomReader(42)
	globalState()
}
