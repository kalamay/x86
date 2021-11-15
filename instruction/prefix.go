package instruction

import (
	"encoding/xml"
	"fmt"
	"math/bits"

	"github.com/kalamay/x86/operand"
)

const (
	rexDefault    = 0b01000000
	vexDefault    = 0b11000100_11100001_01111000
	xopDefault    = 0b10001111_11100000_01111000
	vexMask       = 0b11111111_01111111_10000000
	vex2Byte      = 0b11000101
	vex2MaskUpper = 0b10000000_00000000
	vex2MaskLower = 0b00000000_011111111
)

type Prefix struct {
	Byte      HexByte `xml:"byte,attr"`
	Mandatory bool    `xml:"mandatory,attr"`
}

type PrefixList struct {
	Val       [2]byte
	Len       uint8
	mandatory uint8
}

func (p PrefixList) MandatoryLen() int {
	return bits.OnesCount8(p.mandatory)
}

func (pl *PrefixList) Encode(f *Format, args []operand.Arg) {
	for i := range args {
		if mem, ok := args[i].(operand.Mem); ok {
			if mem.Base.Size() == operand.Size32 {
				f.Val[f.Len] = 0x67
				f.Len++
			}
			if p, ok := mem.Segment.Prefix(); ok {
				f.Val[f.Len] = p
				f.Len++
			}
			break
		}
	}
	f.Len += uint8(copy(f.Val[f.Len:], pl.Val[:pl.Len]))
}

func (pl *PrefixList) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var p Prefix
	if err := d.DecodeElement(&p, &start); err != nil {
		return err
	}
	pl.Val[pl.Len] = byte(p.Byte)
	if p.Mandatory {
		pl.mandatory |= 1 << pl.Len
	}
	pl.Len++
	return nil
}

// REX prefixes are instruction-prefix bytes used in 64-bit mode.
type REX struct {
	// FW describes the REX.W bit. Possible values are "0", "1", and "". An empty
	// value indicates that the bit is ignored.
	FW OptBits `xml:"W,attr"`
	// FR describes the REX.R bit. Possible values are "0", "1", "", or a
	// reference to one of the instruction operands. An empty value indicates that
	// this bit is ignored. If R is a reference to an instruction operand, the
	// operand is of register type and REX.R bit specifies the high bit (bit 3) of
	// the register number.
	FR OptRefBits `xml:"R,attr"`
	// FX describes the REX.X bit. Possible values are "0", "1", "", or a
	// reference to one of the instruction operands. An empty value indicates that
	// this bit is ignored. If X is a reference to an instruction operand, the
	// operand is of memory type and the REX.X bit specifies the high bit (bit 3) of
	// the index register number, and the B instance variable refers to the same operand.
	FX OptRefBits `xml:"X,attr"`
	// FB describes the REX.B bit. Possible values are "0", "1", "", or a
	// reference to one of the instruction operands. An empty value indicates that
	// this bit is ignored. If R is a reference to an instruction operand, the
	// operand can be of register or memory type. If the operand is of register
	// type, the REX.R bit specifies the high bit (bit 3) of the register number,
	// and the REX.X bit is ignored. If the operand is of memory type, the REX.R
	// bit specifies the high bit (bit 3) of the base register number, and the X
	// instance variable refers to the same operand.
	FB OptRefBits `xml:"B,attr"`
}

// Encode conditionally appends the REX prefix to f.
//
// A prefix is necessary only if an instruction references one of the extended
// registers or uses a 64-bit operand. If a REX prefix is used when it has no
// meaning, it is ignored. If used, the REX prefix byte must immediately precede
// the opcode byte or the escape opcode byte.
//
//                    0
//      7     4 3 2 1 0
//     ╭───────────────╮
//     │ 0100  ╎W╎R╎X╎B│
//     ╰───────────────╯
//
// REX.W can be used to determine the operand size but does not solely determine
// operand width. Like the 66H size prefix, 64-bit operand size override has no
// effect on byte-specific operations.
//
// REX.R modifies the ModR/M reg field when that field encodes a GPR, SSE, control
// or debug register. REX.R is ignored when ModR/M specifies other registers or
// defines an extended opcode.
//
// REX.X bit modifies the SIB index field.
//
// REX.B either modifies the base in the ModR/M r/m field or SIB base field; or
// it modifies the opcode reg field used for accessing GPRs.
func (r *REX) Encode(f *Format, args []operand.Arg) bool {
	rex, enc := byte(rexDefault), false

	switch v, m := r.FW.Value(); m {
	case OptModeValue:
		if v == 1 {
			rex, enc = rex|(1<<3), true
		}
	}

	switch v, m := r.FR.Value(); m {
	case OptModeValue:
		if v == 1 {
			rex, enc = rex|(1<<2), true
		}
	case OptModeRef:
		if r, ok := args[v].(operand.Reg); ok && r.Extended8() {
			rex, enc = rex|(1<<2), true
		}
	}

	switch v, m := r.FX.Value(); m {
	case OptModeValue:
		if v == 1 {
			rex, enc = rex|(1<<1), true
		}
	case OptModeRef:
		if m, ok := args[v].(operand.Mem); ok && m.Index.Extended8() {
			rex, enc = rex|(1<<1), true
		}
	}

	switch v, m := r.FB.Value(); m {
	case OptModeValue:
		if v == 1 {
			rex, enc = rex|1, true
		}
	case OptModeRef:
		switch arg := args[v].(type) {
		case operand.Mem:
			if arg.Base.Extended8() {
				rex, enc = rex|1, true
			}
		case operand.Reg:
			if arg.Extended8() {
				rex, enc = rex|1, true
			}
		}
	}

	if enc {
		f.Val[f.Len] = rex
		f.Len++
	}

	return enc
}

func (r *REX) score() int {
	if !r.FW.IsSet() {
		return 0
	}
	return 1
}

// VEX prefixes are used to encode instructions for 128- and 256-bit instructions.
//
// The VEX prefix is required to be the last prefix and immediately precedes the
// opcode bytes. It must follow any other prefixes. If VEX prefix is present a
// REX prefix is not supported.
type VEX struct {
	// Type is the type of the leading byte for VEX encoding.
	Type VexType `xml:"type,attr"`
	// FR describes the VEX.R bit. Possible values are 0, 1, None, or a reference
	// to one of the instruction operands. The value None indicates that this bit
	// is ignored. If R is a reference to an instruction operand, the operand is
	// of register type and VEX.R bit specifies the high bit (bit 3) of the
	// register number.
	FR OptRefBits `xml:"R,attr"`
	// FX describes the VEX.X bit. Possible values are 0, 1, None, or a reference
	// to one of the instruction operands. The value None indicates that this bit
	// is ignored. If X is a reference to an instruction operand, the operand is
	// of memory type and the VEX.X bit specifies the high bit (bit 3) of the index
	// register number, and the B instance variable refers to the same operand.
	FX OptRefBits `xml:"X,attr"`
	// FB describes the VEX.B bit. Possible values are 0, 1, None, or a reference
	// to one of the instruction operands. The value None indicates that this bit
	// is ignored. If R is a reference to an instruction operand, the operand can
	// be of register or memory type. If the operand is of register type, the
	// VEX.R bit specifies the high bit (bit 3) of the register number, and the
	// VEX.X bit is ignored. If the operand is of memory type, the VEX.R bit
	// specifies the high bit (bit 3) of the base register number, and the X
	// instance variable refers to the same operand.
	FB OptRefBits `xml:"B,attr"`
	// Fmmmmm describes the VEX m-mmmm (implied leading opcode bytes) field. In
	// AMD documentation this field is called map_select. Possible values are:
	//     0b00001: Implies 0x0F leading opcode byte.
	//     0b00010: Implies 0x0F 0x38 leading opcode bytes.
	//     0b00011: Implies 0x0F 0x3A leading opcode bytes.
	//     0b01000: This value does not have opcode byte interpretation.
	//              Only XOP instructions use this value.
	//     0b01001: This value does not have opcode byte interpretation.
	//              Only XOP and TBM instructions use this value.
	//     0b01010: This value does not have opcode byte interpretation.
	//              Only TBM instructions use this value.
	//
	// Only VEX prefix with m-mmmm equal to 0b00001 could be encoded in two bytes.
	Fmmmmm OptBits `xml:"m-mmmm,attr"`
	// FW describes the VEX.W bit. Possible values are 0, 1, and None. None indicates
	// that the bit is ignored.
	FW OptBits `xml:"W,attr"`
	// Fvvvv describes the VEX vvvv field. Possible values are 0b0000 or a
	// reference to one of the instruction operands. The value 0b0000 indicates
	// that this field is not used. If vvvv is a reference to an instruction
	// operand, the operand is of register type and VEX.vvvv field specifies its number.
	Fvvvv OptRefBits `xml:"vvvv,attr"`
	// FL describes the VEX.L bit. Possible values are 0, 1, and None. None indicates
	// that the bit is ignored.
	FL OptBits `xml:"L,attr"`
	// Fpp describes the VEX pp (implied legacy prefix) field. Possible values are:
	//     0b00: No implied prefix.
	//     0b01: Implied 0x66 prefix.
	//     0b10: Implied 0xF3 prefix.
	//     0b11: Implied 0xF2 prefix.
	Fpp OptBits `xml:"pp,attr"`
}

// Encode conditionally appends the VEX prefix to f.
//
// The VEX prefix is encoded in either the two-byte form (the first byte must
// be C5H) or in the three-byte form (the first byte must be C4H). The two-byte
// VEX is used mainly for 128-bit, scalar, and the most common 256-bit AVX
// instructions; while the three-byte VEX provides a compact replacement of REX
// and 3-byte opcode instructions (including AVX and FMA instructions). Beyond
// the first byte of the VEX prefix, it consists of a number of bit fields
// providing specific capability
//
// Three-byte format:
//
//                    2               1               0
//      7             0 7 6 5 4       0 7 6     3 2 1 0
//     ╭───────────────┬───────────────┬───────────────╮
//     │   11000100    │R╎X╎B╎ m-mmmm  │W╎ vvvv  ╎L╎pp │
//     ╰───────────────┴───────────────┴───────────────╯
//
// Two-byte format:
//
//                    1               0
//      7             0 7 6 5 4 3 2 1 0
//     ╭───────────────┬───────────────╮
//     │   11000101    │R╎ vvvv  ╎L╎pp │
//     ╰───────────────┴───────────────╯
//
// VEX.R is equivalent to REX.R in 1’s complement (inverted) form.
//     1: Same as REX.R=0 (must be 1 in 32-bit mode)
//     0: Same as REX.R=1 (64-bit mode only)
//
// VEX.X is equivalent to REX.X in 1’s complement (inverted) form.
//     1: Same as REX.X=0 (must be 1 in 32-bit mode)
//     0: Same as REX.X=1 (64-bit mode only)
//
// VEX.B is equivalent to REX.B in 1’s complement (inverted) form.
//     1: Same as REX.B=0 (Ignored in 32-bit mode).
//     0: Same as REX.B=1 (64-bit mode only)
//
// VEX.W is opcode specific (use like REX.W, or used for opcode extension, or
// ignored, depending on the opcode byte).
//
// VEX.m-mmmm provides compaction to allow many legacy instruction to be encoded
// without the constant byte sequence.
//     00000: Reserved for future use (will #UD)
//     00001: implied 0F leading opcode byte
//     00010: implied 0F 38 leading opcode bytes
//     00011: implied 0F 3A leading opcode bytes
//     00100-11111: Reserved for future use (will #UD)
//
// VEX.vvvv is a register specifier (in 1’s complement form) or 1111 if unused.
//
// VEX.L sets the vector length.
//     0: scalar or 128-bit vector
//     1: 256-bit vector
//
// VEX.pp sets the opcode extension providing equivalent functionality of a SIMD prefix.
//     00: None
//     01: 66
//     10: F3
//     11: F2
func (v *VEX) Encode(f *Format, args []operand.Arg) bool {
	var vex uint32

	switch v.Type {
	default:
		return false
	case VexTypeVEX:
		vex = vexDefault
	case VexTypeXOP:
		vex = xopDefault
	}

	switch b, m := v.FR.Value(); m {
	case OptModeValue:
		vex = (vex & ^uint32(1<<15)) | (uint32(^b&1) << 15)
	case OptModeRef:
		if reg, ok := args[b].(operand.Reg); ok && reg.Extended8() {
			vex &= ^uint32(1 << 15)
		}
	}

	switch b, m := v.FX.Value(); m {
	case OptModeValue:
		vex = (vex & ^uint32(1<<14)) | (uint32(^b&1) << 14)
	case OptModeRef:
		if mem, ok := args[b].(operand.Mem); ok && mem.Index.Extended8() {
			vex &= ^uint32(1 << 14)
		}
	}

	switch b, m := v.FB.Value(); m {
	case OptModeValue:
		vex = (vex & ^uint32(1<<13)) | (uint32(^b&1) << 13)
	case OptModeRef:
		switch arg := args[b].(type) {
		case operand.Mem:
			if arg.Base.Extended8() {
				vex &= ^uint32(1 << 13)
			}
		case operand.Reg:
			if arg.Extended8() {
				vex &= ^uint32(1 << 13)
			}
		}
	}

	if b, m := v.Fmmmmm.Value(); m == OptModeValue {
		vex = (vex & ^uint32(0b11111<<8)) | (uint32(b&0b11111) << 8)
	}

	switch b, m := v.FW.Value(); m {
	case OptModeValue:
		vex = (vex & ^uint32(1<<7)) | (uint32(b&1) << 7)
	}

	switch b, m := v.Fvvvv.Value(); m {
	case OptModeValue:
		vex = (vex & ^uint32(0b1111<<3)) | (uint32(^b&0b1111) << 3)
	case OptModeRef:
		id := ^(uint32(args[b].(operand.Reg).ID())) & 0b1111
		vex = (vex & ^uint32(0b1111<<3)) | (id << 3)
	}

	if b, m := v.FL.Value(); m == OptModeValue {
		vex = (vex & ^uint32(1<<2)) | (uint32(b&1) << 2)
	}

	if b, m := v.Fpp.Value(); m == OptModeValue {
		vex = (vex & ^uint32(0b11)) | uint32(b&0b11)
	}

	if (vex & vexMask) == (vexDefault & vexMask) {
		f.Val[f.Len] = vex2Byte
		f.Val[f.Len+1] = byte(((vex & vex2MaskUpper) >> 8) | (vex & vex2MaskLower))
		f.Len += 2
	} else {
		f.Val[f.Len] = byte(vex >> 16)
		f.Val[f.Len+1] = byte(vex >> 8)
		f.Val[f.Len+2] = byte(vex)
		f.Len += 3
	}

	return true
}

func (v *VEX) score() int {
	if v.Type == VexTypeNone {
		return 0
	}
	if b, m := v.Fmmmmm.Value(); m == OptModeValue && b == 1 && !v.FW.IsSet() && !v.FB.IsSet() && !v.FX.IsSet() {
		return 2
	}
	return 3
}

// EVEX prefixes encode the majority of the AVX-512 family of instructions
// operating on 512/256/128-bit vector register operands.
type EVEX struct {
	// FRR describes the EVEX.R'R bits. Possible values are None, or a reference
	// to an register-type instruction operand. None indicates that the field is
	// ignored. The R' bit specifies bit 4 of the register number and the R bit
	// specifies bit 3 of the register number.
	FRR OptRefBits `xml:"RR,attr"`
	// FX describes the EVEX.X bit. Possible values are None, or a reference to
	// one of the instruction operands. The value None indicates that this bit is
	// ignored. If X is a reference to an instruction operand, the operand is of
	// memory type and the EVEX.X bit specifies the high bit (bit 3) of the index
	// register number, and the B instance variable refers to the same operand.
	FX OptRefBits `xml:"X,attr"`
	// FB describes the EVEX.B bit. Possible values are None, or a reference to
	// one of the instruction operands. None indicates that this bit is ignored.
	// If R is a reference to an instruction operand, the operand can be of
	// register or memory type. If the operand is of register type, the EVEX.R
	// bit specifies the high bit (bit 3) of the register number, and the EVEX.X
	// bit is ignored. If the operand is of memory type, the EVEX.R bit specifies
	// the high bit (bit 3) of the base register number, and the X instance
	// variable refers to the same operand.
	FB OptRefBits `xml:"B,attr"`
	// Fmm describes the EVEX mm (compressed legacy escape) field. Identical to
	// two low bits of VEX.m-mmmm field. Possible values are:
	//     0b01: Implies 0x0F leading opcode byte.
	//     0b10: Implies 0x0F 0x38 leading opcode bytes.
	//     0b11: Implies 0x0F 0x3A leading opcode bytes.
	Fmm OptBits `xml:"mm,attr"`
	// FW describes the EVEX.W bit. Possible values are 0, 1, and None. None
	// indicates that the bit is ignored.
	FW OptBits `xml:"W,attr"`
	// Fvvvv describes the EVEX vvvv field. Possible values are 0b0000 or a
	// reference to one of the instruction operands. The value 0b0000 indicates
	// that this field is not used. If vvvv is a reference to an instruction
	// operand, the operand is of register type and EVEX.vvvv field specifies
	// the register number.
	Fvvvv OptRefBits `xml:"vvvv,attr"`
	// Fpp describes the EVEX pp (compressed legacy prefix) field. Possible values are:
	//     0b00: No implied prefix.
	//     0b01: Implied 0x66 prefix.
	//     0b10: Implied 0xF3 prefix.
	//     0b11: Implied 0xF2 prefix.
	Fpp OptBits `xml:"pp,attr"`
	// Fz describes the EVEX z bit. Possible values are None, 0 or a reference
	// to one of the instruction operands. None indicates that the bit is ignored.
	// The value 0 indicates that the bit is not used. If z is a reference to an
	// instruction operand, the operand supports zero-masking with register mask,
	// and EVEX.z indicates whether zero-masking is used.
	Fz OptRefBits `xml:"z,attr"`
	// FLL describes the EVEX.L'L bits. Specify either vector length for the operation,
	// or explicit rounding control (in which case operation is 512 bits wide).
	// Possible values:
	//     None: Indicates that the EVEX.L'L field is ignored.
	//     0b00: 128-bits wide operation.
	//     0b01: 256-bits wide operation.
	//     0b10: 512-bits wide operation.
	//     Reference to the last instruction operand:
	//         EVEX.L'L are interpreted as rounding control and set to the value
	//         specified by the operand. If the rounding control operand is omitted,
	//         EVEX.L'L is set to 0b10 (embedded rounding control is only supported
	//         for 512-bit wide operations).
	FLL OptRefBits `xml:"LL,attr"`
	// Fb describes the EVEX b (broadcast/rounding control/suppress all exceptions
	// context) bit. Possible values are 0 or a reference to one of the instruction
	// operands. The value 0 indicates that this field is not used. If b is a
	// reference to an instruction operand, the operand can be a memory operand with
	// optional broadcasting, an optional rounding specification, or an optional
	// Suppress-all-exceptions specification. If b is a reference to a memory operand,
	// EVEX.b encodes whether broadcasting is used to the operand. If b is a reference
	// to a optional rounding control specification, EVEX.b encodes whether explicit
	// rounding control is used. If b is a reference to a suppress-all-exceptions
	// specification, EVEX.b encodes whether suppress-all-exceptions is enabled.
	Fb OptRefBits `xml:"b,attr"`
	// FV describes the EVEX V field. Possible values are 0, or a reference to
	// one of the instruction operands. The value 0 indicates that this field is
	// not used (EVEX.vvvv is not used or encodes a general-purpose register).
	FV OptRefBits `xml:"V,attr"`
	// Faaa describes the EVEX aaa (embedded opmask register specifier) field. Possible
	// values are 0 or a reference to one of the instruction operands. The value 0
	// indicates that this field is not used. If aaa is a reference to an instruction
	// operand, the operand supports register mask, and EVEX.aaa encodes the mask register.
	Faaa OptRefBits `xml:"aaa,attr"`
	// Disp8xN describes the N value used for encoding compressed 8-bit displacement.
	// Possible values are powers of 2 in [1, 64] range or None. None indicates that
	// this instruction form does not use displacement (the form has no memory operands).
	Disp8xN uint8 `xml:"disp8xN,attr"`
}

// Encode conditionally appends the EVEX prefix to f.
//
//                    3               2               1               0
//      7             0 7 6 5 4 3 2 1 0 7 6     3 2 1 0 7 6 5 4 3 2   0
//     ╭───────────────┬───────────────┬───────────────┬───────────────╮
//     │   01100010    │R╎X╎B╎Ṛ╎0╎0╎mm │W╎ vvvv  ╎1╎pp │z╎ḶL ╎b╎Ṿ╎ aaa │
//     ╰───────────────┴───────────────┴───────────────┴───────────────╯
//
// EVEX.R is equivalent to REX.R in 1’s complement (inverted) form
//     1: Same as REX.R=0 (must be 1 in 32-bit mode)
//     0: Same as REX.R=1 (64-bit mode only)
//
// EVEX.X is equivalent to REX.X in 1’s complement (inverted) form
//     1: Same as REX.X=0 (must be 1 in 32-bit mode)
//     0: Same as REX.X=1 (64-bit mode only)
//
// EVEX.B is equivalent to REX.B in 1’s complement (inverted) form
//     1: Same as REX.B=0 (Ignored in 32-bit mode).
//     0: Same as REX.B=1 (64-bit mode only)
//
// EVEX.Ṛ is the high-16 register specifier modifier. Combine with EVEX.R and
// ModR/M.reg. This bit is stored in inverted format.
//
// EVEX.mm is the compressed legacy escape. Identical to low two bits of VEX.mmmmm.
//
// EVEX.W sets the Osize promotion/Opcode extension.
//
// EVEX.vvvv is the register specifier. Same as VEX.vvvv. This field is encoded in bit
// inverted format.
//
// EVEX.pp is the ompressed legacy prefix. Identical to VEX.pp.
//
// EVEX.z sets zeroing/merging behavior.
//
// EVEX.ḶL sets vector length/RC.
//
// EVEX.b sets broadcast/RC/SAE context.
//
// EVEX.Ṿ is the high-16 VVVV/VIDX register specifier.
//
// EVEX.aaa sets the embedded opmask register specifier.
func (e *EVEX) Encode(f *Format, args []operand.Arg) bool {
	if !e.Fmm.IsSet() {
		return false
	}
	panic("TODO: encode EVEX")
}

func (e *EVEX) score() int {
	if !e.Fmm.IsSet() {
		return 0
	}
	return 4
}

type VexType uint8

const (
	VexTypeNone VexType = iota
	VexTypeVEX          // The VEX prefix (0xC4 or 0xC5) is used.
	VexTypeXOP          // The XOP prefix (0x8F) is used.
)

var vexTypes = map[string]VexType{
	"VEX": VexTypeVEX,
	"XOP": VexTypeXOP,
}

func (t *VexType) UnmarshalText(text []byte) error {
	if v, ok := vexTypes[string(text)]; ok {
		*t = v
		return nil
	}
	return fmt.Errorf("VexType: unknown type %q", text)
}
