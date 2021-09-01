package amd64

import "math/bits"

type Prefix prefix

const (
	OpSize     Prefix = 0x66
	AddrSize   Prefix = 0x67
	AddrOpSize        = AddrSize | (OpSize << 8)
)

func (p Prefix) Encode(b []byte) int {
	return prefix(p).Encode(b)
}

func (p Prefix) Len() int {
	return prefix(p).Len()
}

type Ex prefix

const (
	Rex  Ex = 0b01000000
	Vex2 Ex = 0b11000101_10000000
	Vex3 Ex = 0b11000100_11100000_00000000

	RexW = 0b00001000 | Rex
	RexR = 0b00000100 | Rex
	RexX = 0b00000010 | Rex
	RexB = 0b00000001 | Rex

	Vex3_0F   = 0b00000000_00000001_00000000 | Vex3
	Vex3_0F38 = 0b00000000_00000010_00000000 | Vex3
	Vex3_0F3A = 0b00000000_00000011_00000000 | Vex3

	Vex256 Ex = 0b00000100

	Vex3W = 0b00000000_00000000_10000000 | Vex3

	/*
		EVex Ex = 0b01100010_00000000_00000000_00000000
		eVexMask = 0b00000000_00000000_00000000_00000000 | EVex
	*/

	vex2R Ex = 0b00000000_10000000
	vex3R Ex = 0b00000000_10000000_00000000
	vex3X Ex = 0b00000000_01000000_00000000
	vex3B Ex = 0b00000000_00100000_00000000

	rexMask  = 0b11111111_11111111_11111111_00000000 | Rex
	vex2Mask = 0b11111111_11111111_00000000_00000000 | Vex2
	vex3Mask = 0b11111111_00000000_00000000_00000000 | Vex3
)

func (e Ex) Encode(b []byte) int {
	return prefix(e).Encode(b)
}

func (e Ex) Len() int {
	return prefix(e).Len()
}

func (e Ex) ExpandDest() Ex {
	switch {
	case e.IsRex():
		return e | RexR
	case e.IsVex2():
		return e & ^vex2R
	case e.IsVex3():
		return e & ^vex3R
	}
	return e
}

func (e Ex) IsRex() bool {
	return (e & rexMask) == Rex
}

func (e Ex) IsVex2() bool {
	return (e & vex2Mask) == Vex2
}

func (e Ex) IsVex3() bool {
	return (e & vex3Mask) == Vex3
}

// ModRM defintes the addressing-form specifier byte.
//
// Many instructions that refer to an operand in memory have an addressing-form
// specifier byte (called the ModR/M byte) following the primary opcode. The
// ModR/M byte contains three fields of information:
//
//     ╭─────┬────────┬────────╮
//     │ MOD ╎  REG   ╎  R/M   │
//     ╰─────┴────────┴────────╯
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

func (m ModRM) WithMod(v uint8) ModRM { return ModRM(triple(m).set3(v)) }
func (m ModRM) WithReg(r Reg) ModRM   { return ModRM(triple(m).set2(r.Index())) }
func (m ModRM) WithRM(v uint8) ModRM  { return ModRM(triple(m).set1(v)) }

func (m ModRM) WithRegs(r1 Reg, r2 ...Reg) ModRM {
	m = m.WithMod(0b11)
	m = ModRM(triple(m).set1(r1.Index()))
	if len(r2) > 0 {
		m = ModRM(triple(m).set2(r2[0].Index()))
	}
	return m
}

// SIB defintes the secondary addressing-form specifier byte.
//
// Certain encodings of the ModR/M byte require a second addressing byte
// (the SIB byte). The base-plus-index and scale-plus-index forms of 32-bit
// addressing require the SIB byte. The SIB byte includes the following fields:
// Memory Layout:
//
//          6        3        0
//     ╭─────┬────────┬────────╮
//     │SCALE╎ INDEX  ╎  BASE  │
//     ╰─────┴────────┴────────╯
//
// SCALE - The scale field specifies the scale factor.
// INDEX - The index field specifies the register number of the index register.
// BASE  - The base field specifies the register number of the base register.
type SIB triple

func (s SIB) WithScale(sz Size) SIB { return SIB(triple(s).set3(sz.Factor())) }
func (s SIB) WithIndex(r Reg) SIB   { return SIB(triple(s).set2(r.Index())) }
func (s SIB) WithBase(r Reg) SIB    { return SIB(triple(s).set1(r.Index())) }

type prefix uint32

func (p prefix) Encode(b []byte) int {
	n := 0
	for s := (p.Len() - 1) * 8; s >= 0; s -= 8 {
		b[n] = byte(p >> s)
		n++
	}
	return n
}

func (p prefix) Len() int {
	return 4 - bits.LeadingZeros32(uint32(p))/8
}

type triple byte

const (
	triM1 triple = 0b00000111
	triM2 triple = 0b00111000
	triM3 triple = 0b11000000
	triS2 triple = 3
	triS3 triple = 6
)

func (t triple) set1(v uint8) triple { return (t & ^triM1) | (triple(v) & triM1) }
func (t triple) set2(v uint8) triple { return (t & ^triM2) | ((triple(v) << triS2) & triM2) }
func (t triple) set3(v uint8) triple { return (t & ^triM3) | ((triple(v) << triS3) & triM3) }
