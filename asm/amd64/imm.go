package amd64

import (
	"encoding/binary"
	"fmt"
	"math"
)

type (
	Int  int64
	Uint uint64
)

func (_ Int) Kind() Kind      { return KindImm }
func (_ Int) Size() Size      { return S0 }
func (i Int) Validate() error { return nil }
func (i Int) Name() string    { return "imm" }
func (i Int) String() string  { return fmt.Sprintf("%d", i) }

func (i Int) Match(t Type, _ Size) bool {
	if !t.IsImm() || (t.IsZeroExt() && i < 0) {
		return false
	}
	if !t.IsSignExt() && i >= 0 {
		return t.ImmSize() >= Uint(i).MinSize()
	}
	return t.ImmSize() >= i.MinSize()
}

func (i Int) MinSize() Size {
	switch {
	case math.MinInt8 <= i && i <= math.MaxInt8:
		return S8
	case math.MinInt16 <= i && i <= math.MaxInt16:
		return S16
	case math.MinInt32 <= i && i <= math.MaxInt32:
		return S32
	}
	return S64
}

func (i Int) Encode(b []byte, s Size) int {
	return encodeInt(b, uint64(i), s)
}

func (_ Uint) Kind() Kind      { return KindImm }
func (_ Uint) Size() Size      { return S0 }
func (i Uint) Validate() error { return nil }
func (i Uint) Name() string    { return "imm" }
func (i Uint) String() string  { return fmt.Sprintf("%d", i) }

func (i Uint) Match(t Type, dst Size) bool {
	if !t.IsImm() {
		return false
	}
	s := t.ImmSize()
	if t.IsSignExt() {
		b := s.Bits() - 1
		n := i >> b
		return n == 0 || n == Uint(dst.MaxUint()>>b)
	}
	return s >= i.MinSize()
}

func (i Uint) MinSize() Size {
	switch {
	case i <= math.MaxUint8:
		return S8
	case i <= math.MaxUint16:
		return S16
	case i <= math.MaxUint32:
		return S32
	}
	return S64
}

func (i Uint) Encode(b []byte, s Size) int {
	return encodeInt(b, uint64(i), s)
}

func encodeInt(b []byte, v uint64, s Size) int {
	switch s {
	default:
		panic("invalid type size")
	case S8:
		b[0] = byte(uint8(v))
	case S16:
		binary.LittleEndian.PutUint16(b, uint16(v))
	case S32:
		binary.LittleEndian.PutUint32(b, uint32(v))
	case S64:
		binary.LittleEndian.PutUint64(b, uint64(v))
	}
	return s.Bytes()
}
