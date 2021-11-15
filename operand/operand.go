package operand

import (
	"encoding/xml"
	"fmt"
	"strings"
	"sync"
)

//                    3               2               1               0
//      7             0 7 6 5 4 3 2 1 0 7             0               0
//     ╭───────────────┬───────────────┬───────────────────────────────╮
//     │ ╎ ╎ ╎ ╎C╎I╎W╎R│Z╎K╎ESIZE╎KIND │          KIND BITS            │
//     ╰───────────────┴───────────────┴───────────────────────────────╯
//
// KIND BITS hold the 16-bits of kind-specific values.
//
// KIND is a three-bit kind value
//
// ESIZE is the extension size, if any.
//
// K is the masked flag.
//
// Z is the merged flag. This is generally used with K.
//
// R is the input (read) flag.
//
// W is the output (write) flag.
//
// I is the implicit flag. Implicit params define operands that are used
// without being passed to the instruction. (i.e. eax, ebx, etc. for CPUID)
//
// C is the const flag.
type Param uint32

type Arg interface {
	fmt.Stringer
	Kind() Kind
	Validate() error
	Matches(p Param) bool
}

const (
	pMasked = 1 << (iota + pExtSizeShift + sizeBits)
	pMerged
	pInput
	pOutput
	pImplicit
	pConst

	pMergeMasked  = pMerged | pMasked
	pExtSizeShift = kindShift + kindBits
	pShift        = 16
	pMask         = 0b1111111111111111 << pShift
	pKeyMask      = ^Param(pImplicit | pInput | pOutput | (sizeMask << pExtSizeShift))
)

func ParamOf(id string) (p Param, ok bool) {
	p, ok = paramTypes[id]
	return
}

func (p Param) Kind() Kind         { return Kind(p & kindMask) }
func (p Param) Implicit() bool     { return (p & pImplicit) == pImplicit }
func (p Param) Input() bool        { return (p & pInput) == pInput }
func (p Param) Output() bool       { return (p & pOutput) == pOutput }
func (p Param) Const() bool        { return (p & pConst) == pConst }
func (p Param) Masked() bool       { return (p & pMasked) == pMasked }
func (p Param) MergeMasked() bool  { return (p & pMergeMasked) == pMergeMasked }
func (p Param) ExtendedSize() Size { return Size((p >> pExtSizeShift) & sizeMask) }
func (p Param) ImmConst() bool     { return (p & (pConst | kindMask)) == (pConst | KindImm) }

func (p Param) String() string {
	paramNamesOnce.Do(paramNamesInit)
	m := p & pKeyMask
	switch p & (kindMask | mTypeMask) {
	case KindMem | MemTypeVector32, KindMem | MemTypeVector64:
		m &= ^Param(mElemMask)
	}
	if name, ok := paramNames[m]; ok {
		return name
	}
	return p.DetailString()
}

func (p Param) DetailString() string {
	flags := [4]byte{'I', 'R', 'W', 'C'}
	for i, f := 0, pImplicit; i < len(flags); i++ {
		if (p & Param(f)) == 0 {
			flags[i] |= 32
		}
		f <<= 1
	}
	tail := ""
	switch p & pMergeMasked {
	case pMasked:
		tail = "{k}{z}"
	case pMergeMasked:
		tail = "{k}"
	}
	return fmt.Sprintf("operand.Param(%s, %s%s, %s)", p.Kind(), flags, tail, p.ExtendedSize().ByteString())
}

func (p *Param) setValue(v uint16) {
	*p = (*p & pMask) | Param(v)
}

func (p *Param) ApplyMnemonic(n string) {
	switch *p & (kindMask | mTypeMask) {
	case KindMem | MemTypeVector32, KindMem | MemTypeVector64:
	default:
		return
	}

	mem := MemParam(*p)

	if strings.HasPrefix(n, "VGATHER") || strings.HasPrefix(n, "VSCATTER") {
		if strings.HasSuffix(n, "PS") {
			mem.setElemSize(Size32)
		} else if strings.HasSuffix(n, "PD") {
			mem.setElemSize(Size64)
		}
	} else if strings.HasPrefix(n, "VPGATHER") || strings.HasPrefix(n, "VPSCATTER") {
		if strings.HasSuffix(n, "D") {
			mem.setElemSize(Size32)
		} else if strings.HasSuffix(n, "Q") {
			mem.setElemSize(Size64)
		}
	}

	p.setValue(uint16(mem))
}

func (p *Param) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var elem struct {
		Type         string `xml:"type,attr"`
		Id           string `xml:"id,attr"`
		ExtendedSize Size   `xml:"extended-size,attr"`
		Input        bool   `xml:"input,attr"`
		Output       bool   `xml:"output,attr"`
	}
	if err := d.DecodeElement(&elem, &start); err != nil {
		return err
	}

	val, ok := Param(0), false
	if start.Name.Local == "ImplicitOperand" {
		if val, ok = paramTypes[elem.Id]; !ok {
			return fmt.Errorf("ImplicitOperand: unknown id %q", elem.Id)
		}
		val |= pImplicit
	} else {
		if val, ok = paramTypes[elem.Type]; !ok {
			return fmt.Errorf("Operand: unknown type %q", elem.Type)
		}
	}
	val |= Param(elem.ExtendedSize) << pExtSizeShift
	if elem.Input {
		val |= pInput
	}
	if elem.Output {
		val |= pOutput
	}

	*p = val
	return nil
}

type ParamList struct {
	Val [6]Param
	Len uint8
}

func (pl *ParamList) Add(p Param) {
	pl.Val[pl.Len] = p
	pl.Len++
}

func (pl *ParamList) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var p Param
	if err := p.UnmarshalXML(d, start); err != nil {
		return err
	}
	pl.Add(p)
	return nil
}

var paramTypes = map[string]Param{
	"1": KindImm | pConst | 1, //  The constant value 1.
	"3": KindImm | pConst | 3, //  The constant value 3.

	"imm4":  KindImm,          // A 4-bit immediate value.
	"imm8":  KindImm | Size8,  // An 8-bit immediate value.
	"imm16": KindImm | Size16, // A 16-bit immediate value.
	"imm32": KindImm | Size32, // A 32-bit immediate value.
	"imm64": KindImm | Size64, // A 64-bit immediate value.

	"al":   KindReg | pConst | Param(AL),   // The al register.
	"cl":   KindReg | pConst | Param(CL),   // The cl register.
	"ax":   KindReg | pConst | Param(AX),   // The ax register.
	"dx":   KindReg | pConst | Param(DX),   // The dx register.
	"eax":  KindReg | pConst | Param(EAX),  // The eax register.
	"ecx":  KindReg | pConst | Param(ECX),  // The ecx register.
	"edx":  KindReg | pConst | Param(EDX),  // The edx register.
	"ebx":  KindReg | pConst | Param(EBX),  // The ebx register.
	"rax":  KindReg | pConst | Param(RAX),  // The rax register.
	"rcx":  KindReg | pConst | Param(RCX),  // The rcx register.
	"rdx":  KindReg | pConst | Param(RDX),  // The rdx register.
	"rbx":  KindReg | pConst | Param(RBX),  // The rbx register.
	"rdi":  KindReg | pConst | Param(RDI),  // The rdi register.
	"r11":  KindReg | pConst | Param(R11),  // The r11 register.
	"xmm0": KindReg | pConst | Param(XMM0), // The xmm0 register.

	"r8":        KindReg | RegTypeGeneral | Size8,                 // An 8-bit general-purpose register (al, bl, cl, dl, sil, dil, bpl, spl, r8b-r15b).
	"r16":       KindReg | RegTypeGeneral | Size16,                // A 16-bit general-purpose register (ax, bx, cx, dx, si, di, bp, sp, r8w-r15w).
	"r32":       KindReg | RegTypeGeneral | Size32,                // A 32-bit general-purpose register (eax, ebx, ecx, edx, esi, edi, ebp, esp, r8d-r15d).
	"r64":       KindReg | RegTypeGeneral | Size64,                // A 64-bit general-purpose register (rax, rbx, rcx, rdx, rsi, rdi, rbp, rsp, r8-r15).
	"mm":        KindReg | RegTypeVector | Size64,                 // A 64-bit MMX SIMD register (mm0-mm7).
	"xmm":       KindReg | RegTypeVector | Size128,                // A 128-bit XMM SIMD register (xmm0-xmm31).
	"xmm{k}":    KindReg | RegTypeVector | Size128 | pMergeMasked, // A 128-bit XMM SIMD register (xmm0-xmm31), optionally merge-masked by an AVX-512 mask register (k1-k7).
	"xmm{k}{z}": KindReg | RegTypeVector | Size128 | pMasked,      // A 128-bit XMM SIMD register (xmm0-xmm31), optionally masked by an AVX-512 mask register (k1-k7).
	"ymm":       KindReg | RegTypeVector | Size256,                // A 256-bit YMM SIMD register (ymm0-ymm31).
	"ymm{k}":    KindReg | RegTypeVector | Size256 | pMergeMasked, // A 256-bit YMM SIMD register (ymm0-ymm31), optionally merge-masked by an AVX-512 mask register (k1-k7).
	"ymm{k}{z}": KindReg | RegTypeVector | Size256 | pMasked,      // A 256-bit YMM SIMD register (ymm0-ymm31), optionally masked by an AVX-512 mask register (k1-k7).
	"zmm":       KindReg | RegTypeVector | Size512,                // A 512-bit ZMM SIMD register (zmm0-zmm31).
	"zmm{k}":    KindReg | RegTypeVector | Size512 | pMergeMasked, // A 512-bit ZMM SIMD register (zmm0-zmm31), optionally merge-masked by an AVX-512 mask register (k1-k7).
	"zmm{k}{z}": KindReg | RegTypeVector | Size512 | pMasked,      // A 512-bit ZMM SIMD register (zmm0-zmm31), optionally masked by an AVX-512 mask register (k1-k7).
	"k":         KindReg | RegTypeMask | Size64,                   // An AVX-512 mask register (k0-k7).
	"k{k}":      KindReg | RegTypeMask | Size64 | pMergeMasked,    // An AVX-512 mask register (k0-k7), optionally merge-masked by an AVX-512 mask register (k1-k7).

	"m":          KindMem | MemTypeGeneral,                         // A memory operand of any size.
	"m8":         KindMem | MemTypeGeneral | Size8,                 // An 8-bit memory operand.
	"m16":        KindMem | MemTypeGeneral | Size16,                // A 16-bit memory operand.
	"m16{k}{z}":  KindMem | MemTypeGeneral | Size16 | pMasked,      // A 16-bit memory operand, optionally masked by an AVX-512 mask register (k1-k7).
	"m32":        KindMem | MemTypeGeneral | Size32,                // A 32-bit memory operand.
	"m32{k}":     KindMem | MemTypeGeneral | Size32 | pMergeMasked, // A 32-bit memory operand, optionally merge-masked by an AVX-512 mask register (k1-k7).
	"m32{k}{z}":  KindMem | MemTypeGeneral | Size32 | pMasked,      // A 32-bit memory operand, optionally masked by an AVX-512 mask register (k1-k7).
	"m64":        KindMem | MemTypeGeneral | Size64,                // A 64-bit memory operand.
	"m64{k}":     KindMem | MemTypeGeneral | Size64 | pMergeMasked, // A 64-bit memory operand, optionally merge-masked by an AVX-512 mask register (k1-k7).
	"m64{k}{z}":  KindMem | MemTypeGeneral | Size64 | pMasked,      // A 64-bit memory operand, optionally masked by an AVX-512 mask register (k1-k7).
	"m80":        KindMem | MemTypeFloat80,                         // An 80-bit memory operand.
	"m128":       KindMem | MemTypeGeneral | Size128,               // A 128-bit memory operand.
	"m128{k}{z}": KindMem | MemTypeGeneral | Size128 | pMasked,     // A 128-bit memory operand, optionally masked by an AVX-512 mask register (k1-k7).
	"m256":       KindMem | MemTypeGeneral | Size256,               // A 256-bit memory operand.
	"m256{k}{z}": KindMem | MemTypeGeneral | Size256 | pMasked,     // A 256-bit memory operand, optionally masked by an AVX-512 mask register (k1-k7).
	"m512":       KindMem | MemTypeGeneral | Size512,               // A 512-bit memory operand.
	"m512{k}{z}": KindMem | MemTypeGeneral | Size512 | pMasked,     // A 512-bit memory operand, optionally masked by an AVX-512 mask register (k1-k7).

	"m64/m32bcst":  KindMem | MemTypeBroadcast | Size64 | (Size32 << mElemShift),  // A 64-bit memory operand or a 32-bit memory operand broadcasted to 64 bits {1to2}.
	"m128/m32bcst": KindMem | MemTypeBroadcast | Size128 | (Size32 << mElemShift), // A 128-bit memory operand or a 32-bit memory operand broadcasted to 128 bits {1to4}.
	"m256/m32bcst": KindMem | MemTypeBroadcast | Size256 | (Size32 << mElemShift), // A 256-bit memory operand or a 32-bit memory operand broadcasted to 256 bits {1to8}.
	"m512/m32bcst": KindMem | MemTypeBroadcast | Size512 | (Size32 << mElemShift), // A 512-bit memory operand or a 32-bit memory operand broadcasted to 512 bits {1to16}.
	"m128/m64bcst": KindMem | MemTypeBroadcast | Size128 | (Size64 << mElemShift), // A 128-bit memory operand or a 64-bit memory operand broadcasted to 128 bits {1to2}.
	"m256/m64bcst": KindMem | MemTypeBroadcast | Size256 | (Size64 << mElemShift), // A 256-bit memory operand or a 64-bit memory operand broadcasted to 256 bits {1to4}.
	"m512/m64bcst": KindMem | MemTypeBroadcast | Size512 | (Size64 << mElemShift), // A 512-bit memory operand or a 64-bit memory operand broadcasted to 512 bits {1to8}.

	"moffs8":  KindMem | MemTypeOffset | Size8,  // An 8-bit memory offset from the segment register.
	"moffs16": KindMem | MemTypeOffset | Size16, // A 16-bit memory offset from the segment register.
	"moffs32": KindMem | MemTypeOffset | Size32, // A 32-bit memory offset from the segment register.
	"moffs64": KindMem | MemTypeOffset | Size64, // A 64-bit memory offset from the segment register.

	"vm32x":    KindMem | MemTypeVector32 | (Size128 << mTargetShift),                // A vector of memory addresses using VSIB with 32-bit indices in XMM register.
	"vm32x{k}": KindMem | MemTypeVector32 | (Size128 << mTargetShift) | pMergeMasked, // A vector of memory addresses using VSIB with 32-bit indices in XMM register merge-masked by an AVX-512 mask register (k1-k7).
	"vm32y":    KindMem | MemTypeVector32 | (Size256 << mTargetShift),                // A vector of memory addresses using VSIB with 32-bit indices in YMM register.
	"vm32y{k}": KindMem | MemTypeVector32 | (Size256 << mTargetShift) | pMergeMasked, // A vector of memory addresses using VSIB with 32-bit indices in YMM register merge-masked by an AVX-512 mask register (k1-k7).
	"vm32z":    KindMem | MemTypeVector32 | (Size512 << mTargetShift),                // A vector of memory addresses using VSIB with 32-bit indices in ZMM register.
	"vm32z{k}": KindMem | MemTypeVector32 | (Size512 << mTargetShift) | pMergeMasked, // A vector of memory addresses using VSIB with 32-bit indices in ZMM register merge-masked by an AVX-512 mask register (k1-k7).
	"vm64x":    KindMem | MemTypeVector64 | (Size128 << mTargetShift),                // A vector of memory addresses using VSIB with 64-bit indices in XMM register.
	"vm64x{k}": KindMem | MemTypeVector64 | (Size128 << mTargetShift) | pMergeMasked, // A vector of memory addresses using VSIB with 64-bit indices in XMM register merge-masked by an AVX-512 mask register (k1-k7).
	"vm64y":    KindMem | MemTypeVector64 | (Size256 << mTargetShift),                // A vector of memory addresses using VSIB with 64-bit indices in YMM register.
	"vm64y{k}": KindMem | MemTypeVector64 | (Size256 << mTargetShift) | pMergeMasked, // A vector of memory addresses using VSIB with 64-bit indices in YMM register merge-masked by an AVX-512 mask register (k1-k7).
	"vm64z":    KindMem | MemTypeVector64 | (Size512 << mTargetShift),                // A vector of memory addresses using VSIB with 64-bit indices in ZMM register.
	"vm64z{k}": KindMem | MemTypeVector64 | (Size512 << mTargetShift) | pMergeMasked, // A vector of memory addresses using VSIB with 64-bit indices in ZMM register merge-masked by an AVX-512 mask register (k1-k7).

	"rel8":  KindRel | Size8,  // An 8-bit signed offset relative to the address of instruction end.
	"rel32": KindRel | Size32, // A 32-bit signed offset relative to the address of instruction end.

	"{sae}": KindMisc | SAE, // Suppress-all-exceptions modifier. This operand is optional and can be omitted.
	"{er}":  KindMisc | ER,  // Embedded rounding control. This operand is optional and can be omitted.
}

var paramNames = map[Param]string{}
var paramNamesOnce sync.Once

func paramNamesInit() {
	for n, v := range paramTypes {
		paramNames[v] = n
	}
}
