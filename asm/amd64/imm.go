package amd64

import (
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
func (i Int) Name() string    { return immsTypes[i.MinSize()][1:] }
func (i Int) String() string  { return fmt.Sprintf("%d", i) }

func (i Int) Match(t Type) bool {
	return t.IsImm() && i.MinSize() <= t.ImmSize()
	/*
		// This is the strict sign extention match, disable this for now.
		if !t.IsImm() {
			return false
		}
		s, is := t.ImmSize(), i.MinSize()
		if is > s {
			return false
		}
		return t.IsSignExt() || i >= 0
	*/
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

func (_ Uint) Kind() Kind      { return KindImm }
func (_ Uint) Size() Size      { return S0 }
func (i Uint) Validate() error { return nil }
func (i Uint) Name() string    { return immzTypes[i.MinSize()][1:] }
func (i Uint) String() string  { return fmt.Sprintf("%d", i) }

func (i Uint) Match(t Type) bool {
	return t.IsImm() && i.MinSize() <= t.ImmSize()
	/*
		// This is the strict sign extention match, disable this for now.
		if !t.IsImm() {
			return false
		}
		s, is := t.ImmSize(), i.MinSize()
		if is > s {
			return false
		}
		return t.IsZeroExt() || i < Uint(1)<<(s.Bits()-1)
	*/
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

func ImmValue(op Op) uint64 {
	switch v := op.(type) {
	default:
		panic("invalid operand type")
	case Int:
		return uint64(v)
	case Uint:
		return uint64(v)
	}
}
