package uuid

import (
	"fmt"
	"sync"
)

const bufferLength = 256

var bufferPool = &sync.Pool{
	New: func() interface{} {
		return make([]byte, bufferLength)
	},
}

type sliceWriter struct {
	alloc []byte
	slice []byte
	i, j  uint
}

func makeSliceWriter(n uint) sliceWriter {
	var p, q []byte
	if n <= bufferLength {
		p = bufferPool.Get().([]byte)
		q = p[0:n]
	} else {
		q = make([]byte, n)
	}
	return sliceWriter{
		alloc: p,
		slice: q,
		i:     0,
		j:     n,
	}
}

func (w *sliceWriter) release() {
	alloc := w.alloc
	*w = sliceWriter{}
	if alloc != nil {
		zeroBytes(alloc)
		bufferPool.Put(alloc)
	}
}

func (w sliceWriter) Remain() uint {
	return w.j - w.i
}

func (w sliceWriter) CanWrite(n uint) bool {
	return w.Remain() >= n
}

func (w *sliceWriter) Grab(n uint) []byte {
	if w.i+n > w.j {
		panic(fmt.Errorf("write past end: i+n>j, i=%d j=%d n=%d", w.i, w.j, n))
	}
	slice := w.slice[w.i : w.i+n]
	w.i += n
	return slice
}

func (w *sliceWriter) Unwrite(n uint) {
	if n > w.i {
		panic(fmt.Errorf("unwrite past start: i-n<0, i=%d n=%d", w.i, n))
	}
	w.i -= n
}

func (w *sliceWriter) Fill(ch byte, n uint) {
	slice := w.Grab(n)
	for i := range slice {
		slice[i] = ch
	}
}

func (w *sliceWriter) Write(p []byte) {
	slice := w.Grab(uint(len(p)))
	copy(slice, p)
}

func (w *sliceWriter) WriteByte(ch byte) {
	slice := w.Grab(1)
	slice[0] = ch
}

func (w *sliceWriter) WriteString(s string) {
	slice := w.Grab(uint(len(s)))
	copy(slice, []byte(s))
}

func (w sliceWriter) Bytes() []byte {
	return w.slice[0:w.i]
}

func (w sliceWriter) CopyBytes() []byte {
	return copyBytes(w.slice[0:w.i])
}

func (w sliceWriter) String() string {
	return string(w.slice[0:w.i])
}
