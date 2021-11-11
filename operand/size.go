package operand

import (
	"fmt"
	"math"
	"math/bits"
	"strconv"
)

type Size uint8

const (
	Size0 = iota
	Size8
	Size16
	Size32
	Size64
	Size128
	Size256
	Size512

	// Size4 is a special-case zero size used when working with immedates.
	// Immediate operands must have a size, so a zero-size is interpreted
	// as an imm4 to use with VPERMIL2PD and VPERMIL2PS encodings.
	Size4 = Size0

	sizeBits = 3
	sizeMask = 0b111
)

func (s Size) ImmBits() int {
	if s == Size4 {
		return 4
	}
	return s.Bits()
}

func (s Size) Bits() int  { return s.Bytes() * 8 }
func (s Size) Bytes() int { return (1 << int(s&sizeMask)) >> 1 }

func (s Size) MaxUint() uint64 {
	return ^uint64(math.MaxUint64 << s.ImmBits())
}

func (s *Size) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		*s = Size0
		return nil
	}

	v, err := strconv.ParseUint(string(text), 10, 10)
	if err != nil {
		return err
	}

	if bits.OnesCount64(v) != 1 {
		return fmt.Errorf("Size: not a power-of-2 %q", text)
	}

	*s = Size(bits.TrailingZeros64(v) + 1)
	return nil
}

func (s Size) ByteString() string {
	return byteNames[s]
}

var byteNames = [...]string{"0", "1", "2", "4", "8", "16", "32", "64"}
