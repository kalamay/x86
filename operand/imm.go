package operand

import (
	"encoding/binary"
	"fmt"
	"math"
)

type (
	Int      int64
	Uint     uint64
	ImmParam int16
)

func (i ImmParam) Size() Size {
	return Size(i & sizeMask)
}

func (_ Int) Kind() Kind      { return KindImm }
func (i Int) Validate() error { return nil }
func (i Int) String() string  { return fmt.Sprintf("%d", i) }

func (i Int) Matches(p Param) bool {
	if p.Kind() != KindImm {
		return false
	}
	s := ImmParam(p).Size()
	if p.ExtendedSize() == Size0 && i >= 0 {
		return s >= Uint(i).MinSize()
	}
	return s >= i.MinSize()
}

func (i Int) MinSize() Size {
	switch {
	case 0 <= i && i <= 0b1111:
		return Size4
	case math.MinInt8 <= i && i <= math.MaxInt8:
		return Size8
	case math.MinInt16 <= i && i <= math.MaxInt16:
		return Size16
	case math.MinInt32 <= i && i <= math.MaxInt32:
		return Size32
	}
	return Size64
}

func (i Int) Encode(b []byte, s Size) int {
	return encodeInt(b, uint64(i), s)
}

func (_ Uint) Kind() Kind      { return KindImm }
func (i Uint) Validate() error { return nil }
func (i Uint) String() string  { return fmt.Sprintf("%d", i) }

func (i Uint) Matches(p Param) bool {
	if p.Kind() != KindImm {
		return false
	}
	s, es := ImmParam(p).Size(), p.ExtendedSize()
	if es > s {
		b := s.ImmBits() - 1
		n := i >> b
		return n == 0 || n == Uint(es.MaxUint()>>b)
	}
	return s >= i.MinSize()
}

func (i Uint) MinSize() Size {
	switch {
	case i <= 0b1111:
		return Size4
	case i <= math.MaxUint8:
		return Size8
	case i <= math.MaxUint16:
		return Size16
	case i <= math.MaxUint32:
		return Size32
	}
	return Size64
}

func (i Uint) Encode(b []byte, s Size) int {
	return encodeInt(b, uint64(i), s)
}

func encodeInt(b []byte, v uint64, s Size) int {
	switch s {
	default:
		panic("invalid type size")
	case Size4:
		b[0] = (byte(v) << 4) | (b[0] & 0b1111)
	case Size8:
		b[0] = byte(v)
	case Size16:
		binary.LittleEndian.PutUint16(b, uint16(v))
	case Size32:
		binary.LittleEndian.PutUint32(b, uint32(v))
	case Size64:
		binary.LittleEndian.PutUint64(b, uint64(v))
	}
	return s.Bytes()
}
