package amd64

import (
	"encoding/binary"
	"fmt"
	"sync"
)

var lock sync.RWMutex
var insts = map[string]*InstSet{}

func NewInst(name string, size Size, inst []Inst) *InstSet {
	s := &InstSet{name, size, inst}
	lock.Lock()
	insts[name] = s
	lock.Unlock()
	return s
}

func LookupInst(name string) (*InstSet, bool) {
	lock.RLock()
	s, ok := insts[name]
	lock.RUnlock()
	return s, ok
}

type InstSet struct {
	Name string
	Size Size
	Inst []Inst
}

type Inst struct {
	Types  TypeSet
	Prefix Prefix
	Ex     Ex
	Code   Code
}

func (s *InstSet) OperandPrefix(op Size) bool {
	switch op {
	case S16:
		return s.Size == S32
	case S32:
		return s.Size == S16
	}
	return false
}

func (s *InstSet) Select(ops []Op) (*Inst, error) {
	m, mems, sized := TypeSet(0), 0, false
	for i, op := range ops {
		if err := op.Validate(); err != nil {
			return nil, err
		}
		if op.Kind() == KindMem {
			mems++
		}
		if op.Size() > S0 {
			sized = true
		}
		m |= TypeSet(op.Kind()) << (i * typeBits)
	}
	if mems > 0 && !sized {
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
			wop = wop || s.OperandPrefix(t.ImmSize())
			nimm = v.Encode(imm[:], t.ImmSize())
		case Uint:
			wop = wop || s.OperandPrefix(t.ImmSize())
			nimm = v.Encode(imm[:], t.ImmSize())
		case Reg:
			wop = wop || s.OperandPrefix(t.RegSize())
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
			wop = wop || s.OperandPrefix(t.MemSize())
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
