package uuid

import (
	"strings"
	"testing"
)

func TestIsSpace(t *testing.T) {
	for i := uint(0); i < 256; i++ {
		ch := byte(i)
		expect := (strings.IndexByte("\t\n\v\f\r ", ch) >= 0)
		actual := isSpace(ch)
		if expect != actual {
			t.Errorf("%#02x: expected %v, got %v", ch, expect, actual)
		}
	}
}

func TestIsText(t *testing.T) {
	var buf [40]byte

	oneTest := func(ba []byte, expect bool) {
		actual := isText(ba)
		if expect != actual {
			t.Errorf("%v: expected %v, got %v", ba, expect, actual)
		}
	}

	allTests := func(ch byte, expect bool) {
		for i := range buf {
			buf[i] = ch
		}
		oneTest(buf[:1], expect)
		oneTest(buf[:], expect)
		buf[len(buf)-1] = 0
		oneTest(buf[:], false)
	}

	oneTest(nil, true)
	for i := uint(0); i < 256; i++ {
		ch := byte(i)
		expect := (ch >= 0x20 && ch < 0x7f) || isSpace(ch)
		allTests(ch, expect)
	}
}

func TestEqualBytes(t *testing.T) {
	var buf [8]byte
	for i := range buf {
		buf[i] = byte(0x40 + i)
	}

	oneTest := func(p, q []byte, expect bool) {
		actual := equalBytes(p, q)
		if expect != actual {
			t.Errorf("%v vs %v: expected %v, got %v", p, q, expect, actual)
		}
	}

	oneTest(nil, nil, true)
	oneTest(buf[:], buf[:], true)
	oneTest(buf[0:4], buf[4:8], false)
	oneTest(buf[0:8], buf[1:8], false)
	oneTest(buf[0:8], buf[0:7], false)
}
