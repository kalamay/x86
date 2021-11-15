package operand

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

var ErrRegSizeInvalid = errors.New("reg size invalid")

type (
	Reg      uint32
	RegParam uint16
	RegType  uint8
)

const (
	RegTypeGeneral = iota << rTypeShift
	RegTypeVector
	RegTypeMask
	RegTypeIP
	RegTypeStatus
	RegTypeSegment

	rTypeShift   = sizeBits
	rTypeMask    = 0b111 << rTypeShift
	rIPID        = 0b101
	rMatchMask   = rTypeMask | sizeMask
	rIDShift     = 8
	regInv       = "%!"
	rNx8Mask     = (8 << rIDShift) | (0b110 << rTypeShift)
	rNx8Cmp      = (8 << rIDShift) // (8 <= ID <= 15 || 24 <= ID <= 31) && (Type == (General|Vector))
	rNx16Mask    = (16 << rIDShift) | rTypeMask
	rNx16Cmp     = (16 << rIDShift) | RegTypeVector // 16 <= ID <= 31 && Type == Vector
	rHiMask      = (0b11111100 << rIDShift) | rTypeMask | sizeMask
	rHiCmp       = (20 << rIDShift) | Size8 // 20 <= ID <= 23 && Type == General && Size == 8
	rMasked      = pMasked
	rMergeMasked = pMergeMasked
	rMaskShift   = pShift
	rMaskMask    = 0b111 << rMaskShift
)

func (rt RegType) String() string {
	switch rt {
	case RegTypeGeneral:
		return "#type=general"
	case RegTypeVector:
		return "#type=vector"
	case RegTypeMask:
		return "#type=mask"
	case RegTypeIP:
		return "#type=ip"
	case RegTypeStatus:
		return "#type=status"
	case RegTypeSegment:
		return "#type=segment"
	}
	return fmt.Sprintf("#type=%d", rt)
}

func MakeReg(id uint8, typ RegType, size Size) Reg {
	return (Reg(id) << rIDShift) | Reg(typ&rTypeMask) | Reg(size&sizeMask)
}

const (
	AL = Reg(iota<<rIDShift | RegTypeGeneral | Size8)
	CL
	DL
	BL
	SPL
	BPL
	SIL
	DIL
	R8B
	R9B
	R10B
	R11B
	R12B
	R13B
	R14B
	R15B
	_
	_
	_
	_
	AH
	CH
	DH
	BH
)

const (
	AX = Reg(iota<<rIDShift | RegTypeGeneral | Size16)
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
	EAX = Reg(iota<<rIDShift | RegTypeGeneral | Size32)
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
	RAX = Reg(iota<<rIDShift | RegTypeGeneral | Size64)
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
	MM0 = Reg(iota<<rIDShift | RegTypeVector | Size64)
	MM1
	MM2
	MM3
	MM4
	MM5
	MM6
	MM7
)

const (
	XMM0 = Reg(iota<<rIDShift | RegTypeVector | Size128)
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
	YMM0 = Reg(iota<<rIDShift | RegTypeVector | Size256)
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

const (
	ZMM0 = Reg(iota<<rIDShift | RegTypeVector | Size512)
	ZMM1
	ZMM2
	ZMM3
	ZMM4
	ZMM5
	ZMM6
	ZMM7
	ZMM8
	ZMM9
	ZMM10
	ZMM11
	ZMM12
	ZMM13
	ZMM14
	ZMM15
	ZMM16
	ZMM17
	ZMM18
	ZMM19
	ZMM20
	ZMM21
	ZMM22
	ZMM23
	ZMM24
	ZMM25
	ZMM26
	ZMM27
	ZMM28
	ZMM29
	ZMM30
	ZMM31
)

const (
	K0 = Reg(iota<<rIDShift | RegTypeMask | Size64)
	K1
	K2
	K3
	K4
	K5
	K6
	K7
)

const (
	SS = Reg(iota<<rIDShift | RegTypeSegment | Size16)
	CS
	DS
	ES
	FS
	GS

	IP     = Reg(rIPID<<rIDShift | RegTypeIP | Size16)
	EIP    = Reg(rIPID<<rIDShift | RegTypeIP | Size32)
	RIP    = Reg(rIPID<<rIDShift | RegTypeIP | Size64)
	FLAGS  = Reg(RegTypeStatus | Size16)
	EFLAGS = Reg(RegTypeStatus | Size32)
	RFLAGS = Reg(RegTypeStatus | Size64)
)

func (r Reg) Unmask() Reg {
	return r & ^Reg(rMergeMasked|rMaskMask)
}

func (r Reg) Mask(z Reg) Reg {
	switch {
	case r.Type() != RegTypeVector:
		panic("vector register required")
	case z.Type() != RegTypeMask:
		panic("mask register required")
	}
	return r.Unmask() | rMasked | ((Reg(z.ID()) << rMaskShift) & rMaskMask)
}

func (r Reg) MergeMask(k Reg) Reg {
	switch {
	case r.Type() != RegTypeVector:
		panic("vector register required")
	case k.Type() != RegTypeMask:
		panic("mask register required")
	}
	return r.Unmask() | rMergeMasked | ((Reg(k.ID()) << rMaskShift) & rMaskMask)
}

func (r Reg) Kind() Kind        { return KindReg }
func (r Reg) Size() Size        { return Size(r & sizeMask) }
func (r Reg) Type() RegType     { return RegType(r & rTypeMask) }
func (r Reg) ID() uint8         { return uint8(r >> rIDShift) }
func (r Reg) MMX() bool         { return r&(rTypeMask|sizeMask) == RegTypeVector|Size64 }
func (r Reg) Masked() bool      { return (r & rMasked) == rMasked }
func (r Reg) MergeMasked() bool { return (r & rMergeMasked) == rMergeMasked }

func (r Reg) MaskReg() Reg {
	if r&rMasked == 0 {
		return 0
	}
	return MakeReg(uint8((r&rMaskMask)>>rMaskShift), RegTypeMask, Size64)
}

func (r Reg) Prefix() (byte, bool) {
	switch r.Type() {
	case RegTypeSegment:
		return segmentPrefix[r.ID()], true
	default:
		return 0, false
	}
}

func (r Reg) Validate() error {
	s := r.Size()
	if s < Size8 || Size512 < s {
		return ErrRegSizeInvalid
	}
	return nil
}

func (r Reg) String() string {
	switch r.Type() {
	case RegTypeGeneral:
		return genNames[r.Size()-1][r.ID()]
	case RegTypeVector:
		return vecNames[r.Size()-4][r.ID()]
	case RegTypeMask:
		return maskNames[r.ID()]
	case RegTypeIP:
		return ipNames[r.Size()-2]
	case RegTypeStatus:
		return statusNames[r.Size()-2]
	case RegTypeSegment:
		return segmentNames[r.ID()]
	}
	return regInv
}

func (r Reg) HighByte() bool { return (r & rHiMask) == rHiCmp }
func (r Reg) Next8() bool    { return (r & rNx8Mask) == rNx8Cmp }
func (r Reg) Next16() bool   { return (r & rNx16Mask) == rNx16Cmp }

func (r Reg) Matches(p Param) bool {
	if p.Kind() != KindReg {
		return false
	}
	if p.Const() {
		return RegParam(r) == RegParam(p)
	}
	// r.Type() == RegParam(p).Type() && r.Size() == RegParam(p).Size()
	return ((RegParam(r) ^ RegParam(p)) & rMatchMask) == 0
}

var regNames = map[string]Reg{}
var regNamesOnce sync.Once

func RegOf(name string) (Reg, bool) {
	regNamesOnce.Do(regNamesInit)
	r, ok := regNames[strings.ToLower(name)]
	return r, ok
}

func regNamesInit() {
	for i, names := range genNames {
		for j, name := range names {
			if name == regInv {
				continue
			}
			regNames[name] = MakeReg(uint8(j), RegTypeGeneral, Size(i+1))
		}
	}
	for i, names := range vecNames {
		for j, name := range names {
			regNames[name] = MakeReg(uint8(j), RegTypeVector, Size(i+4))
		}
	}
	for i, name := range maskNames {
		regNames[name] = MakeReg(uint8(i), RegTypeMask, Size64)
	}
	for i, name := range ipNames {
		regNames[name] = MakeReg(rIPID, RegTypeIP, Size(i+2))
	}
	for i, name := range statusNames {
		regNames[name] = MakeReg(0, RegTypeStatus, Size(i+2))
	}
	for i, name := range segmentNames {
		regNames[name] = MakeReg(uint8(i), RegTypeSegment, Size16)
	}
}

var (
	genNames = [...][]string{
		{"al", "cl", "dl", "bl", "spl", "bpl", "sil", "dil", "r8b", "r9b", "r10b", "r11b", "r12b", "r13b", "r14b", "r15b", regInv, regInv, regInv, regInv, "ah", "ch", "dh", "bh"},
		{"ax", "cx", "dx", "bx", "sp", "bp", "si", "di", "r8w", "r9w", "r10w", "r11w", "r12w", "r13w", "r14w", "r15w"},
		{"eax", "ecx", "edx", "ebx", "esp", "ebp", "esi", "edi", "r8d", "r9d", "r10d", "r11d", "r12d", "r13d", "r14d", "r15d"},
		{"rax", "rcx", "rdx", "rbx", "rsp", "rbp", "rsi", "rdi", "r8", "r9", "r10", "r11", "r12", "r13", "r14", "r15"},
	}
	vecNames = [...][]string{
		{"mm0", "mm1", "mm2", "mm3", "mm4", "mm5", "mm6", "mm7"},
		{"xmm0", "xmm1", "xmm2", "xmm3", "xmm4", "xmm5", "xmm6", "xmm7", "xmm8", "xmm9", "xmm10", "xmm11", "xmm12", "xmm13", "xmm14", "xmm15"},
		{"ymm0", "ymm1", "ymm2", "ymm3", "ymm4", "ymm5", "ymm6", "ymm7", "ymm8", "ymm9", "ymm10", "ymm11", "ymm12", "ymm13", "ymm14", "ymm15"},
		{"zmm0", "zmm1", "zmm2", "zmm3", "zmm4", "zmm5", "zmm6", "zmm7", "zmm8", "zmm9", "zmm10", "zmm11", "zmm12", "zmm13", "zmm14", "zmm15", "zmm16", "zmm17", "zmm18", "zmm19", "zmm20", "zmm21", "zmm22", "zmm23", "zmm24", "zmm25", "zmm26", "zmm27", "zmm28", "zmm29", "zmm30", "zmm31"},
	}
	maskNames     = [...]string{"k0", "k1", "k2", "k3", "k4", "k5", "k6", "k7"}
	ipNames       = [...]string{"ip", "eip", "rip"}
	statusNames   = [...]string{"flags", "eflags", "rflags"}
	segmentNames  = [...]string{"ss", "cs", "ds", "es", "fs", "gs"}
	segmentPrefix = [...]byte{0x2E, 0x36, 0x3E, 0x26, 0x64, 0x65}
)
