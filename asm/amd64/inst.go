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
	m, sized := TypeSet(0), false
	for i, op := range ops {
		if err := op.Validate(); err != nil {
			return nil, err
		}
		if op.Size() > S0 {
			sized = true
		}
		m |= TypeSet(op.Kind()) << (i * typeBits)
	}
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
	return in.Encode(b, ops)
}

type Inst struct {
	Types  TypeSet
	Prefix Prefix
	Ex     Ex
	Code   Code
}

func (in *Inst) Encode(b []byte, ops []Op) (int, error) {
	prefix, ex, code := in.Prefix, in.Ex, in.Code
	var (
		addr  Addr
		disp  uint32
		imm   [8]byte
		nimm  int
		high  Reg
		wop   bool
		waddr bool
	)

	t, ts := in.Types.Next()
	for i := 0; t > 0; i++ {
		switch v := ops[i].(type) {
		case Int:
			wop = wop || t.ImmSize() == S16
			nimm = v.Encode(imm[:], t.ImmSize())
		case Uint:
			wop = wop || t.ImmSize() == S16
			nimm = v.Encode(imm[:], t.ImmSize())
		case Reg:
			wop = wop || t.RegSize() == S16
			if high == 0 && v.IsHighByte() {
				high = v
			}
			if !t.IsExplicit() {
				if t.IsOpcode() {
					n := code.Len() - 1
					code.Set(n, code.At(n)|v.Index())
					ex.Extend(v, ExtendOpcode)
				} else if t.Kind() == KindReg {
					addr.SetReg(v)
					ex.Extend(v, ExtendReg)
				} else {
					addr.SetDirect(v)
					ex.Extend(v, ExtendBase)
				}
			}
		case Mem:
			wop = wop || t.MemSize() == S16
			waddr = waddr || v.base.Size() == S32
			addr.SetIndirect(v)
			ex.Extend(v.index, ExtendIndex)
			ex.Extend(v.base, ExtendBase)
			disp = v.disp
		}

		t, ts = ts.Next()
	}

	if ex.IsRex() && high != 0 {
		return 0, fmt.Errorf("cannot encode register '%s' in REX-prefixed instruction", high.Name())
	}

	if waddr {
		prefix.Insert(0x67)
	}
	if wop {
		prefix.Insert(0x66)
	}

	n := 0
	n += prefix.Encode(b[n:])
	n += ex.Encode(b[n:])
	n += code.Encode(b[n:])
	n += addr.Encode(b[n:])
	switch addr.DispSize() {
	case S8:
		b[n] = byte(disp)
		n += 1
	case S32:
		binary.LittleEndian.PutUint32(b[n:], disp)
		n += 4
	}
	if nimm > 0 {
		n += copy(b[n:], imm[:nimm])
	}

	return n, nil
}
