package amd64

import (
	"encoding/binary"
	"fmt"
)

type InstSet struct {
	Name string
	Size Size
	Inst []Inst
}

func (s *InstSet) Select(ops []Op) (*Inst, error) {
	m, sized := TypeSetMaskOf(ops)
	if !sized {
		return nil, fmt.Errorf("ambiguous operand size for %q", s.Name)
	}
	for i := 0; i < len(s.Inst); i++ {
		if (s.Inst[i].Types&m) == m && s.Inst[i].Types.Match(ops) {
			return &s.Inst[i], nil
		}
	}
	return nil, fmt.Errorf("unsupported instruction %q", s.Name)
}

func (s *InstSet) Encode(b []byte, ops []Op) (int, error) {
	in, err := s.Select(ops)
	if err != nil {
		return 0, err
	}
	return in.Encode(b, ops), nil
}

type Inst struct {
	Types  TypeSet
	Prefix Prefix
	Ex     Ex
	Code   Code
}

func (in *Inst) Encode(b []byte, ops []Op) int {
	n := 0
	n += in.Prefix.Encode(b[n:])
	n += in.Ex.Encode(b[n:]) // TODO handle register extension bits
	n += in.Code.Encode(b[n:])

	imt := Type(0)
	imm := Op(nil)

	switch in.Types.Kinds() {
	default:
		panic("TODO: check for vector instructions")
	case MR:
		panic("TODO: handle MR")
	case RM:
		panic("TODO: handle RM")
	case RI:
		if in.Types.At(0).IsOpcode() {
			b[n-1] |= ops[0].(Reg).Index()
		} else {
			panic("TODO: handle RI")
		}
		imt, imm = in.Types.At(1), ops[1]
	case MI:
		b[n] = byte(ModRM(0).WithRegs(ops[0].(Reg)))
		n++
		imt, imm = in.Types.At(1), ops[1]
	}

	if imt != 0 {
		switch imt.ImmSize() {
		default:
			panic("invalid type ize")
		case S8:
			b[n] = byte(uint8(ImmValue(imm)))
		case S16:
			binary.LittleEndian.PutUint16(b[n:], uint16(ImmValue(imm)))
		case S32:
			binary.LittleEndian.PutUint32(b[n:], uint32(ImmValue(imm)))
		case S64:
			binary.LittleEndian.PutUint64(b[n:], uint64(ImmValue(imm)))
		}
		n += imt.ImmSize().Bytes()
	}

	return n
}
