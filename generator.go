package uuid

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type ifaceFunc func() ([]net.Interface, error)
type nowFunc func() time.Time

var gInterfaces ifaceFunc = net.Interfaces
var gNow nowFunc = time.Now
var gReader io.Reader = rand.Reader

var gStateOnce sync.Once
var gState *state

func globalState() *state {
	gStateOnce.Do(func() {
		s := systemSequence()
		a := systemAddress()
		gState = newState(systemTick, s, a)
	})
	return gState
}

// Number of 100ns ticks per year
const tickYear = 315569520000000

// Number of 100ns ticks per 100 years
const tickCentury = 100 * tickYear

// Number of 100ns ticks between UUID epoch (Oct 15, 1582) and Unix epoch (Jan 1, 1970)
const tickEpoch = 122192928000000000

const tickMask = (1 << 60) - 1

const sequenceMask = (1 << 14) - 1

type state struct {
	mu       sync.Mutex
	tickFunc func() uint64
	lastTick uint64
	sequence uint16
	address  [6]byte
	seqStart uint16
}

func newState(f func() uint64, s uint16, a [6]byte) *state {
	return &state{
		tickFunc: f,
		lastTick: (1 << 60),
		sequence: s,
		address:  a,
		seqStart: s,
	}
}

func (state *state) generate(out []byte) {
	state.mu.Lock()
	t := state.tickFunc()
	t0 := state.lastTick
	s := state.sequence
	a := state.address
	if t0 <= tickMask {
		if isLess(t, t0) {
			t = t0
		}
		if t == t0 {
			s = (s + 1) & sequenceMask
			if s == state.seqStart {
				t0 = (t0 + 1) & tickMask
				t = t0
			}
		}
	}
	state.lastTick = t
	state.sequence = s
	state.mu.Unlock()

	var u64 [8]byte
	binary.BigEndian.PutUint64(u64[:], t)

	var u16 [2]byte
	binary.BigEndian.PutUint16(u16[:], s)

	out[0] = u64[4]
	out[1] = u64[5]
	out[2] = u64[6]
	out[3] = u64[7]
	out[4] = u64[2]
	out[5] = u64[3]
	out[6] = u64[0] | 0x10 // force V1
	out[7] = u64[1]
	out[8] = u16[0] | 0x80 // force VariantRFC4122
	out[9] = u16[1]
	out[10] = a[0]
	out[11] = a[1]
	out[12] = a[2]
	out[13] = a[3]
	out[14] = a[4]
	out[15] = a[5]
}

func mustReadRandom(p []byte) {
	n, err := io.ReadFull(gReader, p)
	if n == len(p) && err == nil {
		return
	}
	panic(fmt.Errorf("io.ReadFull %d: %d, %v", len(p), n, err))
}

func systemTick() uint64 {
	value := uint64(gNow().UnixNano()/100) + tickEpoch
	return value & tickMask
}

func systemSequence() uint16 {
	var out [2]byte
	mustReadRandom(out[:])
	value := binary.BigEndian.Uint16(out[0:2])
	return value & sequenceMask
}

func systemAddress() (out [6]byte) {
	ifaces, err := gInterfaces()
	if err == nil {
		// Search for a suitable MAC address
		for _, iface := range ifaces {
			if (iface.Flags & (net.FlagLoopback | net.FlagPointToPoint)) == 0 {
				if isSuitable(iface.HardwareAddr) {
					copy(out[0:6], iface.HardwareAddr)
					return
				}
			}
		}
	}

	// No suitable MAC address was found; pick something at random
	mustReadRandom(out[0:6])
	out[0] |= 0x01 // force multicast bit per RFC 4122
	return
}

func isSuitable(in []byte) bool {
	if len(in) != 6 {
		return false
	}
	if (in[0] & 0x01) == 0x01 {
		// Multicast bit is true -> not a real hardware address
		return false
	}
	if (in[0] & 0x02) == 0x02 {
		// Local bit is true -> not globally unique
		return false
	}
	if in[0] == 0x00 && in[1] == 0x00 && in[2] == 0x00 && in[3] == 0x00 && in[4] == 0x00 && in[5] == 0x00 {
		// All-zeroes MAC
		return false
	}
	return true
}

func isLess(t0, t1 uint64) bool {
	if t0 < tickEpoch && t1 >= tickEpoch {
		return (t1 - t0) <= tickCentury
	}
	if t0 >= tickEpoch && t1 < tickEpoch {
		return (t0 - t1) > tickCentury
	}
	return t0 < t1
}
