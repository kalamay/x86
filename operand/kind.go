package operand

import "fmt"

type Kind uint32

const (
	KindMisc = (iota + 1) << kindShift
	KindImm
	KindReg
	KindMem
	KindRel

	kindShift = 16
	kindBits  = 3
	kindMask  = 0b111 << kindShift
)

func (k Kind) String() string {
	switch k {
	case KindMisc:
		return "misc"
	case KindImm:
		return "imm"
	case KindReg:
		return "reg"
	case KindMem:
		return "mem"
	case KindRel:
		return "rel"
	}
	return fmt.Sprintf("Kind(%d)", uint32(k))
}
