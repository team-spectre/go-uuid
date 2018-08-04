package uuid

import (
	"fmt"
	"io"
)

type sliceReader struct {
	slice []byte
	i, j  uint
}

func makeSliceReader(p []byte) sliceReader {
	n := uint(len(p))
	return sliceReader{
		slice: p,
		i:     0,
		j:     n,
	}
}

func (r sliceReader) checkAdvanceI(n uint) (uint, uint) {
	if r.i+n > r.j {
		panic(fmt.Errorf("read past end: i+n>j, i=%d j=%d n=%d", r.i, r.j, n))
	}
	return r.i, r.i + n
}

func (r sliceReader) checkRewindI(n uint) uint {
	if n > r.i {
		panic(fmt.Errorf("unread past start: i-n<0, i=%d n=%d", r.i, n))
	}
	return r.i - n
}

func (r sliceReader) IsEOF() bool {
	return r.i >= r.j
}

func (r sliceReader) CurrentOffset() uint {
	return r.i
}

func (r sliceReader) EndOffset() uint {
	return r.j
}

func (r sliceReader) Remain() uint {
	return r.j - r.i
}

func (r sliceReader) CanRead(n uint) bool {
	return r.Remain() >= n
}

func (r *sliceReader) Grab(n uint) []byte {
	oldI, newI := r.checkAdvanceI(n)
	r.i = newI
	return r.slice[oldI:newI]
}

func (r *sliceReader) Unread(n uint) {
	r.i = r.checkRewindI(n)
}

func (r *sliceReader) Read(p []byte) (int, error) {
	var err error
	k := uint(len(p))
	if r.i+k > r.j {
		err = io.EOF
		k = r.j - r.i
	}
	copy(p[:k], r.Grab(k))
	return int(k), err
}

func (r *sliceReader) ReadByte() byte {
	slice := r.Grab(1)
	return slice[0]
}

func (r sliceReader) HasPrefix(p []byte) bool {
	n := uint(len(p))
	if r.CanRead(n) {
		slice := r.slice[r.i : r.i+n]
		return equalBytes(p, slice)
	}
	return false
}

func (r *sliceReader) TrimPrefix(n uint) {
	_, r.i = r.checkAdvanceI(n)
}

func (r *sliceReader) TrimLeading(pred func(byte) bool) bool {
	old := r.i
	for r.i < r.j && pred(r.slice[r.i]) {
		r.i++
	}
	return (r.i > old)
}
