package amd64

import "errors"

var ErrRegSizeInvalid = errors.New("reg size invalid")

type Reg uint8

func (_ Reg) Kind() Kind {
	return KindReg
}

func (r Reg) Size() Size {
	return Size(r & SizeMask)
}

func (r Reg) ID() int {
	return int(r >> SizeBits)
}

func (r Reg) Index() uint8 {
	return uint8((r >> SizeBits) & 0b111)
}

func (r Reg) Extended() bool {
	return r.ID() > 7
}

func (r Reg) Match(t Type) bool {
	// TODO: check other register flags to support 32-bit
	if !t.IsReg() {
		return false
	}
	s, rs := t.RegSize(), r.Size()
	if s > S0 {
		return s == rs
	}
	return rs > S0 && S16 <= rs && rs <= S64
}

func (r Reg) Validate() error {
	s := r.Size()
	if s < S8 || S256 < s {
		return ErrRegSizeInvalid
	}
	return nil
}

func (r Reg) Name() string {
	return regNames[r.Size()-1][r.ID()]
}

func (r Reg) String() string {
	return r.Name()
}

// TODO: add AH, CH, etc.

const (
	AL = Reg(iota<<SizeBits | S8)
	CL
	DL
	BL
	SPL
	BPL
	SIL
	DIL
	R8L
	R9L
	R10L
	R11L
	R12L
	R13L
	R14L
	R15L
)

const (
	AX = Reg(iota<<SizeBits | S16)
	CX
	DX
	BX
	SP
	BP
	SI
	DI
	R8W
	R9W
	R10W
	R11W
	R12W
	R13W
	R14W
	R15W
)

const (
	EAX = Reg(iota<<SizeBits | S32)
	ECX
	EDX
	EBX
	ESP
	EBP
	ESI
	EDI
	R8D
	R9D
	R10D
	R11D
	R12D
	R13D
	R14D
	R15D
)

const (
	RAX = Reg(iota<<SizeBits | S64)
	RCX
	RDX
	RBX
	RSP
	RBP
	RSI
	RDI
	R8
	R9
	R10
	R11
	R12
	R13
	R14
	R15
)

const (
	XMM0 = Reg(iota<<SizeBits | S128)
	XMM1
	XMM2
	XMM3
	XMM4
	XMM5
	XMM6
	XMM7
	XMM8
	XMM9
	XMM10
	XMM11
	XMM12
	XMM13
	XMM14
	XMM15
)

const (
	YMM0 = Reg(iota<<SizeBits | S256)
	YMM1
	YMM2
	YMM3
	YMM4
	YMM5
	YMM6
	YMM7
	YMM8
	YMM9
	YMM10
	YMM11
	YMM12
	YMM13
	YMM14
	YMM15
)

var regNames = [...][]string{
	{"al", "cl", "dl", "bl", "spl", "bpl", "sil", "dil", "r8l", "r9l", "r10l", "r11l", "r12l", "r13l", "r14l", "r15l"},
	{"ax", "cx", "dx", "bx", "sp", "bp", "si", "di", "r8w", "r9w", "r10w", "r11w", "r12w", "r13w", "r14w", "r15w"},
	{"eax", "ecx", "edx", "ebx", "esp", "ebp", "esi", "edi", "r8d", "r9d", "r10d", "r11d", "r12d", "r13d", "r14d", "r15d"},
	{"rax", "rcx", "rdx", "rbx", "rsp", "rbp", "rsi", "rdi", "r8", "r9", "r10", "r11", "r12", "r13", "r14", "r15"},
	{"xmm0", "xmm1", "xmm2", "xmm3", "xmm4", "xmm5", "xmm6", "xmm7", "xmm8", "xmm9", "xmm10", "xmm11", "xmm12", "xmm13", "xmm14", "xmm15"},
	{"ymm0", "ymm1", "ymm2", "ymm3", "ymm4", "ymm5", "ymm6", "ymm7", "ymm8", "ymm9", "ymm10", "ymm11", "ymm12", "ymm13", "ymm14", "ymm15"},
	{"zmm0", "zmm1", "zmm2", "zmm3", "zmm4", "zmm5", "zmm6", "zmm7", "zmm8", "zmm9", "zmm10", "zmm11", "zmm12", "zmm13", "zmm14", "zmm15", "zmm16", "zmm17", "zmm18", "zmm19", "zmm20", "zmm21", "zmm22", "zmm23", "zmm24", "zmm25", "zmm26", "zmm27", "zmm28", "zmm29", "zmm30", "zmm31"},
}
