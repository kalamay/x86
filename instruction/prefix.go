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
		if r, ok := args[v].(operand.Reg); ok && r.Extended() {
			rex, enc = rex|(1<<2), true
		}
	}

	switch v, m := r.FX.Value(); m {
	case OptModeValue:
		if v == 1 {
			rex, enc = rex|(1<<1), true
		}
	case OptModeRef:
		if m, ok := args[v].(operand.Mem); ok && m.Index.Extended() {
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
			if arg.Base.Extended() {
				rex, enc = rex|1, true
			}
		case operand.Reg:
			if arg.Extended() {
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
		if reg, ok := args[b].(operand.Reg); ok && reg.Extended() {
			vex &= ^uint32(1 << 15)
		}
	}

	switch b, m := v.FX.Value(); m {
	case OptModeValue:
		vex = (vex & ^uint32(1<<14)) | (uint32(^b&1) << 14)
	case OptModeRef:
		if mem, ok := args[b].(operand.Mem); ok && mem.Index.Extended() {
			vex &= ^uint32(1 << 14)
		}
	}

	switch b, m := v.FB.Value(); m {
	case OptModeValue:
		vex = (vex & ^uint32(1<<13)) | (uint32(^b&1) << 13)
	case OptModeRef:
		switch arg := args[b].(type) {
		case operand.Mem:
			if arg.Base.Extended() {
				vex &= ^uint32(1 << 13)
			}
		case operand.Reg:
			if arg.Extended() {
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

type EVEX struct {
	// Fmm describes the EVEX mm (compressed legacy escape) field. Identical to
	// two low bits of VEX.m-mmmm field. Possible values are:
	//     0b01: Implies 0x0F leading opcode byte.
	//     0b10: Implies 0x0F 0x38 leading opcode bytes.
	//     0b11: Implies 0x0F 0x3A leading opcode bytes.
	Fmm OptBits `xml:"mm,attr"`
	// Fpp describes the EVEX pp (compressed legacy prefix) field. Possible values are:
	//     0b00: No implied prefix.
	//     0b01: Implied 0x66 prefix.
	//     0b10: Implied 0xF3 prefix.
	//     0b11: Implied 0xF2 prefix.
	Fpp OptBits `xml:"pp,attr"`
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
	// FW describes the EVEX.W bit. Possible values are 0, 1, and None. None
	// indicates that the bit is ignored.
	FW OptBits `xml:"W,attr"`
	// Fvvvv describes the EVEX vvvv field. Possible values are 0b0000 or a
	// reference to one of the instruction operands. The value 0b0000 indicates
	// that this field is not used. If vvvv is a reference to an instruction
	// operand, the operand is of register type and EVEX.vvvv field specifies
	// the register number.
	Fvvvv OptRefBits `xml:"vvvv,attr"`
	// FV describes the EVEX V field. Possible values are 0, or a reference to
	// one of the instruction operands. The value 0 indicates that this field is
	// not used (EVEX.vvvv is not used or encodes a general-purpose register).
	FV OptRefBits `xml:"V,attr"`
	// FRR describes the EVEX.R'R bits. Possible values are None, or a reference
	// to an register-type instruction operand. None indicates that the field is
	// ignored. The R' bit specifies bit 4 of the register number and the R bit
	// specifies bit 3 of the register number.
	FRR OptRefBits `xml:"RR,attr"`
	// FB describes the EVEX.B bit. Possible values are None, or a reference to
	// one of the instruction operands. None indicates that this bit is ignored.
	// If R is a reference to an instruction operand, the operand can be of
	// register or memory type. If the operand is of register type, the EVEX.R
	// bit specifies the high bit (bit 3) of the register number, and the EVEX.X
	// bit is ignored. If the operand is of memory type, the EVEX.R bit specifies
	// the high bit (bit 3) of the base register number, and the X instance
	// variable refers to the same operand.
	FB OptRefBits `xml:"B,attr"`
	// FX describes the EVEX.X bit. Possible values are None, or a reference to
	// one of the instruction operands. The value None indicates that this bit is
	// ignored. If X is a reference to an instruction operand, the operand is of
	// memory type and the EVEX.X bit specifies the high bit (bit 3) of the index
	// register number, and the B instance variable refers to the same operand.
	FX OptRefBits `xml:"X,attr"`
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
	// Faaa describes the EVEX aaa (embedded opmask register specifier) field. Possible
	// values are 0 or a reference to one of the instruction operands. The value 0
	// indicates that this field is not used. If aaa is a reference to an instruction
	// operand, the operand supports register mask, and EVEX.aaa encodes the mask register.
	Faaa OptRefBits `xml:"aaa,attr"`
	// Fz describes the EVEX z bit. Possible values are None, 0 or a reference
	// to one of the instruction operands. None indicates that the bit is ignored.
	// The value 0 indicates that the bit is not used. If z is a reference to an
	// instruction operand, the operand supports zero-masking with register mask,
	// and EVEX.z indicates whether zero-masking is used.
	Fz OptRefBits `xml:"z,attr"`
	// Disp8xN describes the N value used for encoding compressed 8-bit displacement.
	// Possible values are powers of 2 in [1, 64] range or None. None indicates that
	// this instruction form does not use displacement (the form has no memory operands).
	Disp8xN uint8 `xml:"disp8xN,attr"`
}

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
