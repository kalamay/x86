package operand

import (
	"testing"
)

func TestReg(t *testing.T) {
	tests := []struct {
		Reg      Reg
		String   string
		Type     RegType
		Size     Size
		HighByte bool
		Next8    bool
		Next16   bool
	}{
		{AL, "al", RegTypeGeneral, Size8, false, false, false},
		{CL, "cl", RegTypeGeneral, Size8, false, false, false},
		{DL, "dl", RegTypeGeneral, Size8, false, false, false},
		{BL, "bl", RegTypeGeneral, Size8, false, false, false},
		{SPL, "spl", RegTypeGeneral, Size8, false, false, false},
		{BPL, "bpl", RegTypeGeneral, Size8, false, false, false},
		{SIL, "sil", RegTypeGeneral, Size8, false, false, false},
		{DIL, "dil", RegTypeGeneral, Size8, false, false, false},
		{R8B, "r8b", RegTypeGeneral, Size8, false, true, false},
		{R9B, "r9b", RegTypeGeneral, Size8, false, true, false},
		{R10B, "r10b", RegTypeGeneral, Size8, false, true, false},
		{R11B, "r11b", RegTypeGeneral, Size8, false, true, false},
		{R12B, "r12b", RegTypeGeneral, Size8, false, true, false},
		{R13B, "r13b", RegTypeGeneral, Size8, false, true, false},
		{R14B, "r14b", RegTypeGeneral, Size8, false, true, false},
		{R15B, "r15b", RegTypeGeneral, Size8, false, true, false},
		{AH, "ah", RegTypeGeneral, Size8, true, false, false},
		{CH, "ch", RegTypeGeneral, Size8, true, false, false},
		{DH, "dh", RegTypeGeneral, Size8, true, false, false},
		{BH, "bh", RegTypeGeneral, Size8, true, false, false},
		{AX, "ax", RegTypeGeneral, Size16, false, false, false},
		{CX, "cx", RegTypeGeneral, Size16, false, false, false},
		{DX, "dx", RegTypeGeneral, Size16, false, false, false},
		{BX, "bx", RegTypeGeneral, Size16, false, false, false},
		{SP, "sp", RegTypeGeneral, Size16, false, false, false},
		{BP, "bp", RegTypeGeneral, Size16, false, false, false},
		{SI, "si", RegTypeGeneral, Size16, false, false, false},
		{DI, "di", RegTypeGeneral, Size16, false, false, false},
		{R8W, "r8w", RegTypeGeneral, Size16, false, true, false},
		{R9W, "r9w", RegTypeGeneral, Size16, false, true, false},
		{R10W, "r10w", RegTypeGeneral, Size16, false, true, false},
		{R11W, "r11w", RegTypeGeneral, Size16, false, true, false},
		{R12W, "r12w", RegTypeGeneral, Size16, false, true, false},
		{R13W, "r13w", RegTypeGeneral, Size16, false, true, false},
		{R14W, "r14w", RegTypeGeneral, Size16, false, true, false},
		{R15W, "r15w", RegTypeGeneral, Size16, false, true, false},
		{EAX, "eax", RegTypeGeneral, Size32, false, false, false},
		{ECX, "ecx", RegTypeGeneral, Size32, false, false, false},
		{EDX, "edx", RegTypeGeneral, Size32, false, false, false},
		{EBX, "ebx", RegTypeGeneral, Size32, false, false, false},
		{ESP, "esp", RegTypeGeneral, Size32, false, false, false},
		{EBP, "ebp", RegTypeGeneral, Size32, false, false, false},
		{ESI, "esi", RegTypeGeneral, Size32, false, false, false},
		{EDI, "edi", RegTypeGeneral, Size32, false, false, false},
		{R8D, "r8d", RegTypeGeneral, Size32, false, true, false},
		{R9D, "r9d", RegTypeGeneral, Size32, false, true, false},
		{R10D, "r10d", RegTypeGeneral, Size32, false, true, false},
		{R11D, "r11d", RegTypeGeneral, Size32, false, true, false},
		{R12D, "r12d", RegTypeGeneral, Size32, false, true, false},
		{R13D, "r13d", RegTypeGeneral, Size32, false, true, false},
		{R14D, "r14d", RegTypeGeneral, Size32, false, true, false},
		{R15D, "r15d", RegTypeGeneral, Size32, false, true, false},
		{RAX, "rax", RegTypeGeneral, Size64, false, false, false},
		{RCX, "rcx", RegTypeGeneral, Size64, false, false, false},
		{RDX, "rdx", RegTypeGeneral, Size64, false, false, false},
		{RBX, "rbx", RegTypeGeneral, Size64, false, false, false},
		{RSP, "rsp", RegTypeGeneral, Size64, false, false, false},
		{RBP, "rbp", RegTypeGeneral, Size64, false, false, false},
		{RSI, "rsi", RegTypeGeneral, Size64, false, false, false},
		{RDI, "rdi", RegTypeGeneral, Size64, false, false, false},
		{R8, "r8", RegTypeGeneral, Size64, false, true, false},
		{R9, "r9", RegTypeGeneral, Size64, false, true, false},
		{R10, "r10", RegTypeGeneral, Size64, false, true, false},
		{R11, "r11", RegTypeGeneral, Size64, false, true, false},
		{R12, "r12", RegTypeGeneral, Size64, false, true, false},
		{R13, "r13", RegTypeGeneral, Size64, false, true, false},
		{R14, "r14", RegTypeGeneral, Size64, false, true, false},
		{R15, "r15", RegTypeGeneral, Size64, false, true, false},
		{MM0, "mm0", RegTypeVector, Size64, false, false, false},
		{MM1, "mm1", RegTypeVector, Size64, false, false, false},
		{MM2, "mm2", RegTypeVector, Size64, false, false, false},
		{MM3, "mm3", RegTypeVector, Size64, false, false, false},
		{MM4, "mm4", RegTypeVector, Size64, false, false, false},
		{MM5, "mm5", RegTypeVector, Size64, false, false, false},
		{MM6, "mm6", RegTypeVector, Size64, false, false, false},
		{MM7, "mm7", RegTypeVector, Size64, false, false, false},
		{XMM0, "xmm0", RegTypeVector, Size128, false, false, false},
		{XMM1, "xmm1", RegTypeVector, Size128, false, false, false},
		{XMM2, "xmm2", RegTypeVector, Size128, false, false, false},
		{XMM3, "xmm3", RegTypeVector, Size128, false, false, false},
		{XMM4, "xmm4", RegTypeVector, Size128, false, false, false},
		{XMM5, "xmm5", RegTypeVector, Size128, false, false, false},
		{XMM6, "xmm6", RegTypeVector, Size128, false, false, false},
		{XMM7, "xmm7", RegTypeVector, Size128, false, false, false},
		{XMM8, "xmm8", RegTypeVector, Size128, false, true, false},
		{XMM9, "xmm9", RegTypeVector, Size128, false, true, false},
		{XMM10, "xmm10", RegTypeVector, Size128, false, true, false},
		{XMM11, "xmm11", RegTypeVector, Size128, false, true, false},
		{XMM12, "xmm12", RegTypeVector, Size128, false, true, false},
		{XMM13, "xmm13", RegTypeVector, Size128, false, true, false},
		{XMM14, "xmm14", RegTypeVector, Size128, false, true, false},
		{XMM15, "xmm15", RegTypeVector, Size128, false, true, false},
		{YMM0, "ymm0", RegTypeVector, Size256, false, false, false},
		{YMM1, "ymm1", RegTypeVector, Size256, false, false, false},
		{YMM2, "ymm2", RegTypeVector, Size256, false, false, false},
		{YMM3, "ymm3", RegTypeVector, Size256, false, false, false},
		{YMM4, "ymm4", RegTypeVector, Size256, false, false, false},
		{YMM5, "ymm5", RegTypeVector, Size256, false, false, false},
		{YMM6, "ymm6", RegTypeVector, Size256, false, false, false},
		{YMM7, "ymm7", RegTypeVector, Size256, false, false, false},
		{YMM8, "ymm8", RegTypeVector, Size256, false, true, false},
		{YMM9, "ymm9", RegTypeVector, Size256, false, true, false},
		{YMM10, "ymm10", RegTypeVector, Size256, false, true, false},
		{YMM11, "ymm11", RegTypeVector, Size256, false, true, false},
		{YMM12, "ymm12", RegTypeVector, Size256, false, true, false},
		{YMM13, "ymm13", RegTypeVector, Size256, false, true, false},
		{YMM14, "ymm14", RegTypeVector, Size256, false, true, false},
		{YMM15, "ymm15", RegTypeVector, Size256, false, true, false},
		{ZMM0, "zmm0", RegTypeVector, Size512, false, false, false},
		{ZMM1, "zmm1", RegTypeVector, Size512, false, false, false},
		{ZMM2, "zmm2", RegTypeVector, Size512, false, false, false},
		{ZMM3, "zmm3", RegTypeVector, Size512, false, false, false},
		{ZMM4, "zmm4", RegTypeVector, Size512, false, false, false},
		{ZMM5, "zmm5", RegTypeVector, Size512, false, false, false},
		{ZMM6, "zmm6", RegTypeVector, Size512, false, false, false},
		{ZMM7, "zmm7", RegTypeVector, Size512, false, false, false},
		{ZMM8, "zmm8", RegTypeVector, Size512, false, true, false},
		{ZMM9, "zmm9", RegTypeVector, Size512, false, true, false},
		{ZMM10, "zmm10", RegTypeVector, Size512, false, true, false},
		{ZMM11, "zmm11", RegTypeVector, Size512, false, true, false},
		{ZMM12, "zmm12", RegTypeVector, Size512, false, true, false},
		{ZMM13, "zmm13", RegTypeVector, Size512, false, true, false},
		{ZMM14, "zmm14", RegTypeVector, Size512, false, true, false},
		{ZMM15, "zmm15", RegTypeVector, Size512, false, true, false},
		{ZMM16, "zmm16", RegTypeVector, Size512, false, false, true},
		{ZMM17, "zmm17", RegTypeVector, Size512, false, false, true},
		{ZMM18, "zmm18", RegTypeVector, Size512, false, false, true},
		{ZMM19, "zmm19", RegTypeVector, Size512, false, false, true},
		{ZMM20, "zmm20", RegTypeVector, Size512, false, false, true},
		{ZMM21, "zmm21", RegTypeVector, Size512, false, false, true},
		{ZMM22, "zmm22", RegTypeVector, Size512, false, false, true},
		{ZMM23, "zmm23", RegTypeVector, Size512, false, false, true},
		{ZMM24, "zmm24", RegTypeVector, Size512, false, true, true},
		{ZMM25, "zmm25", RegTypeVector, Size512, false, true, true},
		{ZMM26, "zmm26", RegTypeVector, Size512, false, true, true},
		{ZMM27, "zmm27", RegTypeVector, Size512, false, true, true},
		{ZMM28, "zmm28", RegTypeVector, Size512, false, true, true},
		{ZMM29, "zmm29", RegTypeVector, Size512, false, true, true},
		{ZMM30, "zmm30", RegTypeVector, Size512, false, true, true},
		{ZMM31, "zmm31", RegTypeVector, Size512, false, true, true},
		{K0, "k0", RegTypeMask, Size64, false, false, false},
		{K1, "k1", RegTypeMask, Size64, false, false, false},
		{K2, "k2", RegTypeMask, Size64, false, false, false},
		{K3, "k3", RegTypeMask, Size64, false, false, false},
		{K4, "k4", RegTypeMask, Size64, false, false, false},
		{K5, "k5", RegTypeMask, Size64, false, false, false},
		{K6, "k6", RegTypeMask, Size64, false, false, false},
		{K7, "k7", RegTypeMask, Size64, false, false, false},
		{SS, "ss", RegTypeSegment, Size16, false, false, false},
		{CS, "cs", RegTypeSegment, Size16, false, false, false},
		{DS, "ds", RegTypeSegment, Size16, false, false, false},
		{ES, "es", RegTypeSegment, Size16, false, false, false},
		{FS, "fs", RegTypeSegment, Size16, false, false, false},
		{GS, "gs", RegTypeSegment, Size16, false, false, false},
		{IP, "ip", RegTypeIP, Size16, false, false, false},
		{EIP, "eip", RegTypeIP, Size32, false, false, false},
		{RIP, "rip", RegTypeIP, Size64, false, false, false},
		{FLAGS, "flags", RegTypeStatus, Size16, false, false, false},
		{EFLAGS, "eflags", RegTypeStatus, Size32, false, false, false},
		{RFLAGS, "rflags", RegTypeStatus, Size64, false, false, false},
	}

	for _, test := range tests {
		if r, ok := RegOf(test.String); !ok {
			t.Errorf("failed to lookup %q", test.String)
		} else if test.Reg != r {
			t.Errorf("incorrect lookup for %q: expect=%s, actual=%s", test.String, test.Reg, r)
		}
		if test.String != test.Reg.String() {
			t.Errorf("incorrect string: expect=%q, actual=%q", test.String, test.Reg.String())
		}
		if test.Type != test.Reg.Type() {
			t.Errorf("incorrect type for %q: expect=%s, actual=%s", test.Reg, test.Type, test.Reg.Type())
		}
		if test.Size != test.Reg.Size() {
			t.Errorf("incorrect size for %q: expect=%s, actual=%s", test.Reg, test.Size, test.Reg.Size())
		}
		if test.HighByte != test.Reg.HighByte() {
			t.Errorf("incorrect high byte for %q: expect=%v, actual=%v", test.Reg, test.HighByte, test.Reg.HighByte())
		}
		if test.Next8 != test.Reg.Next8() {
			t.Errorf("incorrect next 8 for %q: expect=%v, actual=%v", test.Reg, test.Next8, test.Reg.Next8())
		}
		if test.Next16 != test.Reg.Next16() {
			t.Errorf("incorrect next 16 for %q: expect=%v, actual=%v", test.Reg, test.Next16, test.Reg.Next16())
		}

		masks := [...]Reg{K0, K1, K2, K3, K4, K5, K6, K7}
		for _, k := range masks {
			if test.Reg.Type() != RegTypeVector {
				continue
			}

			r := test.Reg
			if r.Masked() {
				t.Errorf("%q should not be masked", r)
			}
			if r.MergeMasked() {
				t.Errorf("%q should not be merge-masked", r)
			}
			if r.MaskReg() != 0 {
				t.Errorf("%q should not have a mask", r)
			}

			r = r.MergeMask(k)
			if !r.Masked() {
				t.Errorf("%q should be masked", r)
			}
			if !r.MergeMasked() {
				t.Errorf("%q should be merge-masked", r)
			}
			if r.MaskReg() != k {
				t.Errorf("%q should have a mask", r)
			}

			r = r.Mask(k)
			if !r.Masked() {
				t.Errorf("%q should be masked", r)
			}
			if r.MergeMasked() {
				t.Errorf("%q should not be merge-masked", r)
			}
			if r.MaskReg() != k {
				t.Errorf("%q should have a mask", r)
			}

			r = r.Unmask()
			if r != test.Reg {
				t.Errorf("%q should be restored", r)
			}
		}
	}
}
