package amd64

import (
	"fmt"
	"math"
	"math/bits"
)

type Code uint64

const (
	C1 = 1 << (iota + cMaskOff)
	C2
	C3
	C4
	C5
	C6
	C7

	cFinal   = 0x7fffffffffffffff
	cValues  = 0x00ffffffffffffff
	cMask    = 0xff00000000000000
	cMaskOff = 56
)

func C(b ...byte) (c Code) {
	for i, b := range b {
		c.Set(i, b)
	}
	return
}

func (c Code) Len() int       { return 8 - bits.LeadingZeros8(c.At(c.Cap())) }
func (c Code) Cap() int       { return 7 }
func (c Code) Has(i int) bool { return (c>>(cMaskOff+i))&1 == 1 }
func (c Code) At(i int) byte  { return byte(c >> (i << 3)) }

func (c *Code) Set(i int, b byte) {
	if i >= c.Cap() {
		panic("invalid code index")
	}
	*c = (*c & ^(0xff << (i * 8))) | Code(b)<<(i*8) | 1<<(cMaskOff+i)
}

func (c *Code) Clear(i int) {
	if i >= c.Cap() {
		panic("invalid code index")
	}
	*c &= ^(0xff<<(i*8) | 1<<(cMaskOff+i))
}

func (c *Code) Insert(b byte) {
	*c = (((*c & cMask) << 1) | (1 << cMaskOff) | ((*c << 8) & cValues) | Code(b)) & cFinal
}

func (c Code) Encode(b []byte) int {
	n := c.Len()
	for i := 0; i < n; i++ {
		b[i] = byte(c)
		c >>= 8
	}
	return n
}

func (c Code) String() string {
	n := c.Len()
	return fmt.Sprintf("#%0*x", n*2, uint64(c)&cValues)
}

type Prefix Code

const (
	OpSize     Prefix = 0x66 | C1
	AddrSize   Prefix = 0x67 | C1
	AddrOpSize Prefix = (0x67 << 8) | 0x66 | C2
)

func (p Prefix) Len() int            { return Code(p).Len() }
func (p Prefix) Cap() int            { return Code(p).Cap() }
func (p *Prefix) Set(i int, b byte)  { (*Code)(p).Set(i, b) }
func (p *Prefix) Clear(i int)        { (*Code)(p).Clear(i) }
func (p Prefix) Encode(b []byte) int { return Code(p).Encode(b) }

type Ex Code

const (
	Rex  Ex = 0b01000000 | C1
	Vex2 Ex = 0b10000000_11000101 | C2
	Vex3 Ex = 0b00000000_11100000_11000100 | C3

	RexW = 0b00001000 | Rex
	RexR = 0b00000100 | Rex
	RexX = 0b00000010 | Rex
	RexB = 0b00000001 | Rex

	Vex3_0F   = 0b00000000_00000001_00000000 | Vex3
	Vex3_0F38 = 0b00000000_00000010_00000000 | Vex3
	Vex3_0F3A = 0b00000000_00000011_00000000 | Vex3

	Vex256 Ex = 0b00000100

	Vex3W = 0b10000000_00000000_00000000 | Vex3

	/*
		EVex Ex = 0b01100010_00000000_00000000_00000000
		eVexMask = 0b00000000_00000000_00000000_00000000 | EVex
	*/

	vex2R Ex = 0b10000000_00000000
	vex3R Ex = 0b00000000_10000000_00000000
	vex3X Ex = 0b00000000_01000000_00000000
	vex3B Ex = 0b00000000_00100000_00000000
)

func (e Ex) Len() int            { return Code(e).Len() }
func (e Ex) Encode(b []byte) int { return Code(e).Encode(b) }

func (e Ex) IsRex() bool  { return (e & 0xff) == Rex }
func (e Ex) IsVex2() bool { return (e & 0xff) == Vex2 }
func (e Ex) IsVex3() bool { return (e & 0xff) == Vex3 }

func (e Ex) ExpandDest() Ex {
	switch e & 0xff {
	case Rex:
		return e | RexR
	case Vex2:
		return e & ^vex2R
	case Vex3:
		return e & ^vex3R
	}
	return e
}

// Addr represents the ModR/M and SIB addressing-form specifier bytes.
//
// Many instructions that refer to an operand in memory have an addressing-form
// specifier byte (called the ModR/M byte) following the primary opcode. Certain
// encodings of the ModR/M byte require a second addressing byte (the SIB byte).
// The base-plus-index and scale-plus-index forms of 32-bit addressing require
// the SIB byte.
//
//           6         3         0     14        11         8
//     ╭──────┬─────────┬─────────┬──────┬─────────┬─────────╮
//     │ MOD  ╎   REG   ╎   R/M   │SCALE ╎  INDEX  ╎  BASE   │
//     ╰──────┴─────────┴─────────┴──────┴─────────┴─────────╯
//
// MOD   - The mod field combines with the r/m field to form 32 possible values:
//         eight registers and 24 addressing modes.
// REG   - The reg/opcode field specifies either a register number or three more
//         bits of opcode information. The purpose of the reg/opcode field is
//         specified in the primary opcode.
// R/M   - The r/m field can specify a register as an operand or it can be
//         combined with the mod field to encode an addressing mode. Sometimes,
//         certain combinations of the mod field and the r/m field are used to
//         express opcode information for some instructions.
// SCALE - The scale field specifies the scale factor.
// INDEX - The index field specifies the register number of the index register.
// BASE  - The base field specifies the register number of the base register.
type Addr Code

func (a Addr) Len() int            { return Code(a).Len() }
func (a Addr) Encode(b []byte) int { return Code(a).Encode(b) }

func (a Addr) ModRM() (ModRM, bool) {
	if Code(a).Has(0) {
		return ModRM(Code(a).At(0)), true
	}
	return ModRM(0), false
}

func (a Addr) SIB() (SIB, bool) {
	if Code(a).Has(1) {
		return SIB(Code(a).At(1)), true
	}
	return SIB(0), false
}

func (a Addr) DispSize() Size {
	if m, ok := a.ModRM(); ok {
		return m.DispSize()
	}
	return S0
}

func (a *Addr) SetReg(r Reg) {
	m, ok := a.ModRM()
	if !ok {
		m = ModRMDirect
	}
	(*Code)(a).Set(0, byte(m.WithReg(r)))
}

func (a *Addr) SetDirect(r Reg) {
	m, _ := a.ModRM()
	(*Code)(a).Set(0, byte(m.WithDirect(r)))
}

func (a *Addr) SetIndirect(mem Mem) {
	m, _ := a.ModRM()

	d := uint8(ModDisp0)
	switch {
	case mem.disp > math.MaxUint8:
		d = ModDisp32
	case mem.disp > 0:
		d = ModDisp8
	}

	if mem.scale == S0 {
		(*Code)(a).Set(0, byte(m.WithIndirectReg(d, mem.base)))
	} else {
		(*Code)(a).Set(0, byte(m.WithIndirectMem(d)))
		(*Code)(a).Set(1, byte(SIBOf(mem)))
	}
}

// ModRM defintes the addressing-form specifier byte.
//
// Many instructions that refer to an operand in memory have an addressing-form
// specifier byte (called the ModR/M byte) following the primary opcode. The
// ModR/M byte contains three fields of information:
//
//           6         3         0
//     ╭──────┬─────────┬─────────╮
//     │ MOD  ╎   REG   ╎   R/M   │
//     ╰──────┴─────────┴─────────╯
//
// MOD - The mod field combines with the r/m field to form 32 possible values:
//       eight registers and 24 addressing modes.
// REG - The reg/opcode field specifies either a register number or three more
//       bits of opcode information. The purpose of the reg/opcode field is
//       specified in the primary opcode.
// R/M - The r/m field can specify a register as an operand or it can be
//       combined with the mod field to encode an addressing mode. Sometimes,
//       certain combinations of the mod field and the r/m field are used to
//       express opcode information for some instructions.
type ModRM triple

const (
	ModDisp0 = iota
	ModDisp8
	ModDisp32
	ModDirect

	ModRMDirect = ModDirect << 6

	rmIndirect = 0b100
	rmDisp32   = 0b101
)

func (m ModRM) WithReg(r Reg) ModRM {
	return ModRM(triple(m).set2(r.Index()))
}

func (m ModRM) WithDirect(r Reg) ModRM {
	return ModRM(triple(m).set3(ModDirect).set1(r.Index()))
}

func (m ModRM) WithIndirectReg(disp uint8, r Reg) ModRM {
	i := r.Index()
	if i == rmIndirect || (disp == 0 && i == rmDisp32) {
		panic("R/M field cannot hold register index")
	}
	return ModRM(triple(m).set3(disp).set1(i))
}

func (m ModRM) WithIndirectMem(disp uint8) ModRM {
	return ModRM(triple(m).set3(disp).set1(rmIndirect))
}

func (m ModRM) Mod() uint8 { return triple(m).v3() }
func (m ModRM) Reg() uint8 { return triple(m).v2() }

func (m ModRM) DispSize() Size {
	switch m.Mod() {
	case ModDisp8:
		return S8
	case ModDisp32:
		return S32
	}
	return S0
}

// SIB defintes the secondary addressing-form specifier byte.
//
// Certain encodings of the ModR/M byte require a second addressing byte
// (the SIB byte). The base-plus-index and scale-plus-index forms of 32-bit
// addressing require the SIB byte. The SIB byte includes the following fields:
//
//           6         3         0
//     ╭──────┬─────────┬─────────╮
//     │SCALE ╎  INDEX  ╎  BASE   │
//     ╰──────┴─────────┴─────────╯
//
// SCALE - The scale field specifies the scale factor.
// INDEX - The index field specifies the register number of the index register.
// BASE  - The base field specifies the register number of the base register.
type SIB triple

func SIBOf(m Mem) SIB {
	return SIB(triple(0).set1(m.base.Index()).set2(m.index.Index()).set3(uint8(m.scale) - 1))
}

type triple byte

const (
	// TODO remove types?
	triM1 triple = 0b00000111
	triM2 triple = 0b00111000
	triM3 triple = 0b11000000
	triS1 triple = 0
	triS2 triple = 3
	triS3 triple = 6
)

func (t triple) v1() uint8 { return uint8((t & triM1) >> triS1) }
func (t triple) v2() uint8 { return uint8((t & triM2) >> triS2) }
func (t triple) v3() uint8 { return uint8((t & triM3) >> triS3) }

func (t triple) set1(v uint8) triple { return (t & ^triM1) | (triple(v) & triM1) }
func (t triple) set2(v uint8) triple { return (t & ^triM2) | ((triple(v) << triS2) & triM2) }
func (t triple) set3(v uint8) triple { return (t & ^triM3) | ((triple(v) << triS3) & triM3) }
