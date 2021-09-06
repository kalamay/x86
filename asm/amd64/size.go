package amd64

import "math"

type Size uint8

const (
	S0 Size = iota
	S8
	S16
	S32
	S64
	S128
	S256
	S512

	SizeBits = 3
	SizeMask = 0b111
)

func (s Size) Factor() uint8 {
	if s == S0 {
		panic("size 0 has no factor")
	}
	return uint8(s - 1)
}

func (s Size) Bits() int {
	return s.Bytes() << 3
}

func (s Size) Bytes() int {
	return 1 << int(s) >> 1
}

func (s Size) String() string {
	return sizeNames[s]
}

func (s Size) ByteString() string {
	return byteNames[s]
}

func (s Size) MaxUint() uint64 {
	return ^uint64(math.MaxUint64 << s.Bits())
}

var sizeNames = [...]string{"S0", "S8", "S16", "S32", "S64", "S128", "S256", "S512"}
var byteNames = [...]string{"0", "1", "2", "4", "8", "16", "32", "64"}
