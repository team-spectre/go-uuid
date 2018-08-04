package uuid

import (
	"fmt"
	"io"
	"testing"
)

func TestSliceReader(t *testing.T) {
	expectThrow := func(name string, fn func()) {
		var panicValue interface{}
		(func() {
			defer func() {
				panicValue = recover()
			}()
			fn()
		}())
		if panicValue == nil {
			t.Errorf("expected panic at %s, but did not panic", name)
		}
	}

	input := []byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit. Mauris hendrerit.")
	r := makeSliceReader(input)
	n := uint(len(input))

	checkCanRead := func(k uint, expect bool) {
		actual := r.CanRead(k)
		if expect != actual {
			t.Errorf("expect CanRead(%d) = %v, got %v", k, expect, actual)
		}
	}

	checkOffsets := func(i, j, k uint) {
		actualI := r.CurrentOffset()
		actualJ := r.EndOffset()
		actualR := r.Remain()
		if i != actualI {
			t.Errorf("expect CurrentOffset %d, got %d", i, actualI)
		}
		if j != actualJ {
			t.Errorf("expect EndOffset %d, got %d", j, actualJ)
		}
		if k != actualR {
			t.Errorf("expect Remain %d, got %d", k, actualR)
		}
		checkCanRead(0, true)
		checkCanRead(k, true)
		checkCanRead(k+1, false)
	}

	checkEqual := func(actual, expect []byte) {
		if !equalBytes(expect, actual) {
			t.Errorf("wrong result:\n\texpected: %#02x\n\t  actual: %#02x", expect, actual)
		}
	}

	checkHasPrefix := func(p []byte, expect bool) {
		actual := r.HasPrefix(p)
		if expect != actual {
			t.Errorf("expect HasPrefix(%#02x) = %v, got %v", p, expect, actual)
		}
	}

	checkRead := func(k, kk uint, expect []byte) {
		var expecterr error
		if kk < k {
			expecterr = io.EOF
		}

		actual := make([]byte, k)
		n, err := r.Read(actual)
		if n < 0 || uint(n) != kk || err != expecterr {
			x := fmt.Sprintf("%#v", expecterr)
			y := fmt.Sprintf("%#v", err)
			t.Errorf("expect Read(%d) = (%d, %s), got (%d, %s)", k, kk, x, n, y)
			return
		}
		actual = actual[:n]
		checkEqual(actual, expect)
	}

	checkOffsets(0, n, n)
	checkEqual(r.Grab(6), []byte("Lorem "))
	checkOffsets(6, n, n-6)
	expectThrow("Unread(7)", func() { r.Unread(7) })
	r.Unread(6)
	checkOffsets(0, n, n)
	r.Grab(6)
	checkOffsets(6, n, n-6)
	checkHasPrefix([]byte("Lorem "), false)
	checkHasPrefix([]byte("ipsum "), true)
	checkRead(6, 6, []byte("ipsum "))
	r.TrimPrefix(45)
	expectThrow("Grab(18)", func() { r.Grab(18) })
	checkRead(18, 17, []byte("Mauris hendrerit."))

	r = makeSliceReader([]byte("   foo"))
	r.TrimLeading(isSpace)
	checkRead(3, 3, []byte("foo"))
}
