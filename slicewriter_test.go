package uuid

import (
	"testing"
)

func TestSliceWriter(t *testing.T) {
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

	w := makeSliceWriter(2 * bufferLength)
	if !w.CanWrite(2 * bufferLength) {
		t.Errorf("cannot write %d bytes to empty buffer", 2*bufferLength)
	}

	w.Grab(2 * bufferLength)
	w.Unwrite(2 * bufferLength)

	expectThrow("Grab(2*n+1)", func() { w.Grab(2*bufferLength + 1) })

	copy(w.Grab(3), []byte{1, 2, 3})
	expectThrow("Unwrite(4)", func() { w.Unwrite(4) })
	w.Unwrite(1)
	w.Write([]byte{4, 5, 6})

	expected := []byte{1, 2, 4, 5, 6}
	actual := w.Bytes()
	if !equalBytes(expected, actual) {
		t.Errorf("wrong result:\n\texpected: %v\n\t  actual: %v", expected, actual)
	}
}
