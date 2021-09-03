package amd64

import (
	"math/bits"
	"strings"
)

type Kind uint16

const (
	KindReg = 1 << (iota + (2 * SizeBits))
	KindMem
	KindImm = 0
)

// Type is a compact representation of an allowed operand type.
//
// Memory Layout in bits:
//
//       7  6        3        0  F  E  D  C  B  A  9  8
//     ╭──┬──┬────────┬────────┬──┬──┬──┬──┬──┬──┬──┬──╮
//     │ A╎ R╎ SIZE2  ╎ SIZE1  │  ╎AP╎SP╎SR╎RA╎FP╎SX╎OR│
//     ╰──┴──┴────────┴────────┴──┴──┴──┴──┴──┴──┴──┴──╯
//
// SIZE1 - Size for KindMem
// SIZE2 - Size for KindImm, KindReg, or a secondary memory size.
//  R - Flag enabling KindReg and disabling KindImm
//  A - Flag enabling KindMem and disabling KindImm
// OR - Flag specifying opcode merge behavior of KindReg
// SX - Flag specifying sign-extension mode
// FP - Flag specifying that the operand is a floating point value.
// RA - Flag specifying that KindMem must be a relative address.
// SR - Flag specifying that KindReg must be a segment register.
// SP - Flag specifying a segment far pointer. For KindReg, this requires a
//      16-bit selector for the segment register, and a SIZE2-sized offset
//      within the destination segment. For KindMem, this indicates the memory
//      operand contains a far pointer with a SIZE1-sized offset.
// AP - Flag specifying that KindMem must be an address pair indicating bounds.
//      When set, the R flag must be unset, and the SIZE2 is used for the upper
//      bound size.
type Type uint16

const (
	TOR = 1 << (iota + 2 + (2 * SizeBits)) // opcode-register flag
	TSX                                    // sign-extension flag
	TFP                                    // floating-point flag
	TRA                                    // relative-address flag
	TSR                                    // segment-register flag
	TSP                                    // segment-pointer flag
	TAP                                    // address-pair flag

	typeBits = 16
)

const (
	// Immediate operand
	TI8  = KindImm | Type(S8<<SizeBits)  // imm8
	TI16 = KindImm | Type(S16<<SizeBits) // imm16
	TI32 = KindImm | Type(S32<<SizeBits) // imm32
	TI64 = KindImm | Type(S64<<SizeBits) // imm64

	// Register operand
	TR8   = KindReg | Type(S8<<SizeBits)   // r8
	TR16  = KindReg | Type(S16<<SizeBits)  // r16
	TR32  = KindReg | Type(S32<<SizeBits)  // r32
	TR64  = KindReg | Type(S64<<SizeBits)  // r64
	TR128 = KindReg | Type(S128<<SizeBits) // xmm
	TR256 = KindReg | Type(S256<<SizeBits) // ymm
	TR512 = KindReg | Type(S512<<SizeBits) // zmm

	// Memory operand
	TA8   = KindMem | Type(S8)   // m8
	TA16  = KindMem | Type(S16)  // m16
	TA32  = KindMem | Type(S32)  // m32
	TA64  = KindMem | Type(S64)  // m64
	TA128 = KindMem | Type(S128) // m128
	TA256 = KindMem | Type(S256) // m256
	TA512 = KindMem | Type(S512) // m512

	// Operand that is either a register or memory
	TM8   = TR8 | TA8     // r/m8
	TM16  = TR16 | TA16   // r/m16
	TM32  = TR32 | TA32   // r/m32
	TM64  = TR64 | TA64   // r/m64
	TM128 = TR128 | TA128 // xmm/m128
	TM256 = TR256 | TA256 // ymm/m256
	TM512 = TR512 | TA512 // zmm/m512
)

func (t Type) ImmSize() Size { return t.RegSize() }
func (t Type) RegSize() Size { return Size((t >> SizeBits) & SizeMask) }
func (t Type) MemSize() Size { return Size(t & SizeMask) }

func (t Type) Kind() Kind       { return Kind(t & (KindReg | KindMem)) }
func (t Type) IsImm() bool      { return t.Kind() == KindImm }
func (t Type) IsReg() bool      { return (t & KindReg) != 0 }
func (t Type) IsMem() bool      { return (t & KindMem) != 0 }
func (t Type) IsOpcode() bool   { return (t & TOR) != 0 }
func (t Type) IsSignExt() bool  { return (t & TSX) != 0 }
func (t Type) IsZeroExt() bool  { return (t & TSX) == 0 }
func (t Type) IsFloat() bool    { return (t & TFP) != 0 }
func (t Type) IsRelAddr() bool  { return (t & TRA) != 0 }
func (t Type) IsSegReg() bool   { return (t & TSR) != 0 }
func (t Type) IsSegPtr() bool   { return (t & TSP) != 0 }
func (t Type) IsAddrPair() bool { return (t & TAP) != 0 }

func (t Type) String() string {
	ms, rs := t.MemSize(), t.RegSize()

	switch t.Kind() {
	case KindImm:
		if t.IsZeroExt() {
			return immzTypes[rs]
		}
		return immsTypes[rs]
	case KindReg:
		return regTypes[rs]
	case KindMem:
		return memTypes[ms]
	case KindReg | KindMem:
		if rs == ms {
			return rmTypes[ms]
		}
		return regTypes[rs] + "/" + memTypes[ms][1:]
	}

	return badType
}

// TypeSet represents upto 4 Type values.
//
// Memory Layout in bytes:
//      0       1        2       3        4       5        6       7
//     ╭────────────────┬────────────────┬────────────────┬────────────────╮
//     │     TYPE1      │     TYPE2      │     TYPE3      │     TYPE4      │
//     ╰────────────────┴────────────────┴────────────────┴────────────────╯
type TypeSet uint64

const (
	T1 = iota * typeBits
	T2
	T3
	T4

	MR = ((KindMem | KindReg) << T1) | (KindReg << T2)
	RM = (KindReg << T1) | ((KindMem | KindReg) << T2)
	RI = (KindReg << T1) | (KindImm << T2)
	MI = ((KindMem | KindReg) << T1) | (KindImm << T2)
)

const (
	M8_R8   = (TypeSet(TM8) << T1) | (TypeSet(TR8) << T2)   // r/m8 ← r8
	M16_R16 = (TypeSet(TM16) << T1) | (TypeSet(TR16) << T2) // r/m16 ← r16
	M32_R32 = (TypeSet(TM32) << T1) | (TypeSet(TR32) << T2) // r/m32 ← r32
	M64_R64 = (TypeSet(TM64) << T1) | (TypeSet(TR64) << T2) // r/m64 ← r64

	R8_M8    = (TypeSet(TR8) << T1) | (TypeSet(TM8) << T2)       // r8 ← r/m8
	R16_M16  = (TypeSet(TR16) << T1) | (TypeSet(TM16) << T2)     // r16 ← r/m16
	R16_M16s = (TypeSet(TR16) << T1) | (TypeSet(TM16|TSX) << T2) // r16 ← r/m16 (sign extended)
	R32_M32  = (TypeSet(TR32) << T1) | (TypeSet(TM32) << T2)     // r32 ← r/m32
	R32_M32s = (TypeSet(TR32) << T1) | (TypeSet(TM32|TSX) << T2) // r32 ← r/m32 (sign extended)
	R64_M64  = (TypeSet(TR64) << T1) | (TypeSet(TM64) << T2)     // r64 ← r/m64
	R16_M8   = (TypeSet(TR16) << T1) | (TypeSet(TM8) << T2)      // r16 ← r/m8
	R16_M8s  = (TypeSet(TR16) << T1) | (TypeSet(TM8|TSX) << T2)  // r16 ← r/m8 (sign extended)
	R32_M8   = (TypeSet(TR32) << T1) | (TypeSet(TM8) << T2)      // r32 ← r/m8
	R32_M8s  = (TypeSet(TR32) << T1) | (TypeSet(TM8|TSX) << T2)  // r32 ← r/m8 (sign extended)
	R32_M16  = (TypeSet(TR32) << T1) | (TypeSet(TM16) << T2)     // r32 ← r/m16
	R32_M16s = (TypeSet(TR32) << T1) | (TypeSet(TM16|TSX) << T2) // r32 ← r/m16 (sign extended)
	R64_M8   = (TypeSet(TR64) << T1) | (TypeSet(TM8) << T2)      // r64 ← r/m8
	R64_M8s  = (TypeSet(TR64) << T1) | (TypeSet(TM8|TSX) << T2)  // r64 ← r/m8 (sign extended)
	R64_M16  = (TypeSet(TR64) << T1) | (TypeSet(TM16) << T2)     // r64 ← r/m16
	R64_M16s = (TypeSet(TR64) << T1) | (TypeSet(TM16|TSX) << T2) // r64 ← r/m16 (sign extended)
	R64_M32  = (TypeSet(TR64) << T1) | (TypeSet(TM32) << T2)     // r64 ← r/m32
	R64_M32s = (TypeSet(TR64) << T1) | (TypeSet(TM32|TSX) << T2) // r64 ← r/m32 (sign extended)

	R8_I8   = (TypeSet(TR8) << T1) | (TypeSet(TI8) << T2)   // r8 ← imm8
	R16_I16 = (TypeSet(TR16) << T1) | (TypeSet(TI16) << T2) // r16 ← imm16
	R32_I32 = (TypeSet(TR32) << T1) | (TypeSet(TI32) << T2) // r32 ← imm32

	O8_I8   = (TypeSet(TR8|TOR) << T1) | (TypeSet(TI8) << T2)   // r8 ← imm8 (opcode merged)
	O16_I16 = (TypeSet(TR16|TOR) << T1) | (TypeSet(TI16) << T2) // r16 ← imm16 (opcode merged)
	O32_I32 = (TypeSet(TR32|TOR) << T1) | (TypeSet(TI32) << T2) // r32 ← imm32 (opcode merged)
	O64_I64 = (TypeSet(TR64|TOR) << T1) | (TypeSet(TI64) << T2) // r64 ← imm64 (opcode merged)

	M8_I8    = (TypeSet(TM8) << T1) | (TypeSet(TI8) << T2)       // r/m8 ← r8
	M8_I8s   = (TypeSet(TM8) << T1) | (TypeSet(TI8|TSX) << T2)   // r/m8 ← r8 (sign extended)
	M16_I8s  = (TypeSet(TM16) << T1) | (TypeSet(TI8|TSX) << T2)  // r/m16 ← r8 (sign extended)
	M16_I16  = (TypeSet(TM16) << T1) | (TypeSet(TI16) << T2)     // r/m16 ← r16
	M32_I8s  = (TypeSet(TM32) << T1) | (TypeSet(TI8|TSX) << T2)  // r/m32 ← r8
	M32_I32  = (TypeSet(TM32) << T1) | (TypeSet(TI32) << T2)     // r/m32 ← r32
	M64_I8s  = (TypeSet(TM64) << T1) | (TypeSet(TI8|TSX) << T2)  // r/m64 ← r8
	M64_I32s = (TypeSet(TM64) << T1) | (TypeSet(TI32|TSX) << T2) // r/m64 ← r32 (sign extended)
)

func T(t ...Type) (ts TypeSet) {
	for i, t := range t {
		ts |= TypeSet(t << i * typeBits)
	}
	return
}

func (ts TypeSet) Next() (Type, TypeSet) {
	return Type(ts), ts >> typeBits
}

func (ts TypeSet) Len() int {
	return 4 - bits.LeadingZeros32(uint32(ts))/typeBits
}

func (ts TypeSet) At(i int) Type {
	return Type(ts >> (i * typeBits))
}

func (ts TypeSet) Kinds() TypeSet {
	return ts & 0xc0c0c0c0
}

func (ts TypeSet) Match(op []Op) bool {
	t, ts := ts.Next()
	for _, op := range op {
		if !op.Match(t) {
			return false
		}
		t, ts = ts.Next()
	}
	return ts == 0
}

func (ts TypeSet) String() string {
	s := [4]string{}
	n := 0
	t, ts := ts.Next()
	for t > 0 {
		s[n] = t.String()
		n++
		t, ts = ts.Next()
	}
	return "(" + strings.Join(s[:n], ",") + ")"
}

const badType = "@!"

var immsTypes = [...]string{badType, "@imm8s", "@imm16s", "@imm32s", "@imm64s"}
var immzTypes = [...]string{badType, "@imm8z", "@imm16z", "@imm32z", "@imm64z"}
var regTypes = [...]string{"@r", "@r8", "@r16", "@r32", "@r64", "@xmm", "@ymm", "@zmm"}
var memTypes = [...]string{"@m", "@m8", "@m16", "@m32", "@m64", "@m128", "@m256", "@m512"}
var rmTypes = [...]string{"@r/m", "@r/m8", "@r/m16", "@r/m32", "@r/m64", "@xmm/m128", "@ymm/m256", "@zmm/m512"}
