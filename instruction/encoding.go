package instruction

import (
	"encoding/xml"
	"math"
	"reflect"

	"github.com/kalamay/x86/operand"
)

type Format struct {
	Val [15]byte
	Len uint8
}

func (f *Format) Bytes() []byte {
	return f.Val[:f.Len]
}

type Encoding struct {
	Prefix       PrefixList    `xml:"Prefix"`
	REX          REX           `xml:"REX"`
	VEX          VEX           `xml:"VEX"`
	EVEX         EVEX          `xml:"EVEX"`
	Opcode       OpcodeList    `xml:"Opcode"`
	ModRM        ModRM         `xml:"ModRM"`
	RegisterByte RegisterByte  `xml:"RegisterByte"`
	Immediate    ImmediateList `xml:"Immediate"`
	CodeOffset   CodeOffset    `xml:"CodeOffset"`
	DataOffset   DataOffset    `xml:"DataOffset"`
}

func (e *Encoding) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	tmp := Encoding{}
	v := reflect.ValueOf(&tmp).Elem()

	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}
		switch tok := tok.(type) {
		case xml.StartElement:
			skip := true
			for i := 0; i < v.NumField(); i++ {
				t := v.Type().Field(i)
				if name := t.Tag.Get("xml"); name == tok.Name.Local {
					if err = d.DecodeElement(v.Field(i).Addr().Interface(), &tok); err != nil {
						return err
					}
					skip = false
					break
				}
			}
			if skip {
				if err = d.Skip(); err != nil {
					return err
				}
			}
		case xml.EndElement:
			if e.Opcode.Len == 0 || tmp.score() < e.score() {
				*e = tmp
			}
			return nil
		}
	}
}

func (e *Encoding) score() int {
	s := int(e.Prefix.Len) +
		e.REX.score() +
		e.VEX.score() +
		e.EVEX.score() +
		e.ModRM.score() +
		e.CodeOffset.Size.Bytes()
	for i := uint8(0); i < e.Opcode.Len; i++ {
		if !e.Opcode.Val[i].AddEnd.IsSet() {
			s++
		}
	}
	for i := uint8(0); i < e.Immediate.Len; i++ {
		s += e.Immediate.Val[i].Size.Bytes()
	}
	return s
}

type Opcode struct {
	// Byte is operation code as a byte integer.
	Byte HexByte `xml:"byte,attr"`
	// AddEnd opcode merging behavior, None or a reference to an instruction
	// operand. If addend is a reference to an instruction operand, the operand
	// is of register type and the three lowest bits of its number must be ORed
	// with Byte to produce the final opcode value.
	AddEnd OptRef `xml:"addend,attr"`
}

type OpcodeList struct {
	Val [3]Opcode
	Len uint8
}

func (ol *OpcodeList) Encode(f *Format, args []operand.Arg) {
	for i := uint8(0); i < ol.Len; i++ {
		f.Val[f.Len] = byte(ol.Val[i].Byte)
		if v, mode := ol.Val[i].AddEnd.Value(); mode == OptModeRef {
			r := args[v].(operand.Reg)
			f.Val[f.Len] |= (r.ID() & 0b111)
		}
		f.Len++
	}
}

func (ol *OpcodeList) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var oc Opcode
	if err := d.DecodeElement(&oc, &start); err != nil {
		return err
	}
	ol.Val[ol.Len] = oc
	ol.Len++
	return nil
}

// ModRM defines the addressing-form specifier byte.
type ModRM struct {
	// Mode describes the addressing mode. Possible values are 0b11 or a reference
	// to an instruction operand. If mode value is 0b11, the Mod R/M encodes two
	// register operands or a register operand and an opcode extension. If mode
	// is a reference to an instruction operand, the operand has memory type and
	// its addressing mode must be coded instruction the Mod R/M mode field.
	Mode OptRefBits `xml:"mode,attr"`
	// Reg is a register or an opcode extension. Possible values are an int value,
	// or a reference to an instruction operand. If reg is an int value, this value
	// extends the opcode and must be directly coded in the reg field. If reg is a
	// reference to an instruction operand, the operand is of register type, and
	// the reg field specifies bits 0-2 of the register number.
	Reg OptRefByte `xml:"reg,attr"`
	// Rm a register or memory operand. Must be a reference to an instruction
	// operand. If rm is a reference to a operand, rm specifies bits 0-2 of the
	// register number. If the operand is of memory type, rm specifies bits 0-2
	// of the base register number unless a SIB byte is used.
	RM OptRef `xml:"rm,attr"`
}

const (
	ModRMIndirect       byte = 0b00000000
	ModRMIndirectDisp8       = 0b01000000
	ModRMIndirectDisp32      = 0b10000000
	ModRMDirect              = 0b11000000
	ModRMSIB                 = 0b00000100
	ModRMRelative            = 0b00000101
	ModRMMask                = 0b00000111
	SIBNoBase                = 0b00000101
	SIBNoIndex               = 0b00100000
)

// Encode appends the ModR/M byte, and possible SIB byte, to f.
//
// Many instructions that refer to an operand in memory have an addressing-form
// specifier byte (called the ModR/M byte) following the primary opcode. The
// ModR/M byte contains three fields of information:
//
//      7    6 5       3 2       0
//     ╭──────────────────────────╮
//     │ MOD  ╎   REG   ╎   R/M   │
//     ╰──────────────────────────╯
//
// MOD:
//     The mod field combines with the r/m field to form 32 possible values:
//     eight registers and 24 addressing modes.
// REG:
//     The reg/opcode field specifies either a register number or three more
//     bits of opcode information. The purpose of the reg/opcode field is
//     specified in the primary opcode.
// R/M:
//     The r/m field can specify a register as an operand or it can be
//     combined with the mod field to encode an addressing mode. Sometimes,
//     certain combinations of the mod field and the r/m field are used to
//     express opcode information for some instructions.
//
// SIB defintes the secondary addressing-form specifier byte.
//
// Certain encodings of the ModR/M byte require a second addressing byte
// (the SIB byte). The base-plus-index and scale-plus-index forms of 32-bit
// addressing require the SIB byte. The SIB byte includes the following fields:
//
//      7    6 5       3 2       0
//     ╭──────────────────────────╮
//     │SCALE ╎  INDEX  ╎  BASE   │
//     ╰──────────────────────────╯
//
// SCALE:
//     The scale field specifies the scale factor.
// INDEX:
//     The index field specifies the register number of the index register.
// BASE:
//     The base field specifies the register number of the base register.
func (m *ModRM) Encode(f *Format, args []operand.Arg) {
	v, mode := m.RM.Value()
	if mode != OptModeRef {
		return
	}

	var (
		disp       int32
		dispsize   operand.Size
		modrm, sib byte
	)

	n := 1

	switch arg := args[v].(type) {
	case operand.Reg:
		modrm = ModRMDirect | (arg.ID() & ModRMMask)
	case operand.Mem:
		disp = arg.Disp
		switch arg.Base.Type() {
		case operand.RegTypeIP:
			modrm = ModRMRelative
			dispsize = operand.Size32
		case operand.RegTypeGeneral:
			switch {
			case disp < math.MinInt8 || math.MaxInt8 < disp:
				dispsize = operand.Size32
				modrm = ModRMIndirectDisp32
			case disp != 0:
				dispsize = operand.Size8
				modrm = ModRMIndirectDisp8
			default:
				modrm = ModRMIndirect
			}

			if arg.Base == 0 {
				n = 2
				sib = SIBNoBase
			} else {
				id := arg.Base.ID() & ModRMMask
				if arg.Index != 0 || id == ModRMSIB || id == ModRMRelative {
					n = 2
					modrm |= ModRMSIB
					sib = id
				} else {
					modrm |= id
				}
			}

			if n == 2 {
				if arg.Index == 0 {
					sib |= SIBNoIndex
				} else {
					sib |= (uint8(arg.Scale-1) << 6) | ((arg.Index.ID() & ModRMMask) << 3)
				}
			}
		}
	}

	switch v, mode = m.Reg.Value(); mode {
	case OptModeValue:
		modrm |= (v & ModRMMask) << 3
	case OptModeRef:
		modrm |= (args[v].(operand.Reg).ID() & ModRMMask) << 3
	}

	f.Val[f.Len] = modrm
	f.Val[f.Len+1] = sib
	f.Len += uint8(n)
	if dispsize > operand.Size0 {
		f.Len += uint8(operand.Int(disp).Encode(f.Val[f.Len:], dispsize))
	}
}

func (m *ModRM) score() int {
	switch m.Mode.Mode() {
	default:
		return 0
	case OptModeValue:
		// For now, any value bits are always 0b11 (i.e. direct mode), so there won't
		// be any SIB byte or displacement.
		return 1
	case OptModeRef:
		// A ref could point to a memory operand that requires an SIB byte and/or a
		// displacement, but we only know that when we have the arguments. We'll just
		// estimate with both for now, but this is pretty ugly.
		return 2
	}
}

type RegisterByte struct {
	// Register is a reference to an instruction operand of register type. The
	// register number is encoded in the low 4 bits of the byte (register number
	// is in 0..15 for all instructions which use this encoding component).
	Register OptRef `xml:"register,attr"`
	// Payload is the optional value of the high 4 bits of the byte. Can be None
	// or a reference to an instruction operand of imm4 type. None indicates that
	// this high 4 bits are not used. The only instructions that use the payload
	// are VPERMIL2PD and VPERMIL2PS from XOP instruction set.
	Payload OptRefByte `xml:"payload,attr"`
}

func (rb *RegisterByte) Encode(f *Format, args []operand.Arg) {
	if _, m := rb.Register.Value(); m == OptModeRef {
		panic("TODO: encode RegisterByte")
	}
}

type Immediate struct {
	// Size of the constant in bytes. Possible values are 1, 2, 4, or 8.
	Size operand.Size `xml:"size,attr"`
	// Value of the constant. Can be an int value or a reference to an instruction operand.
	Value OptRefByte `xml:"value,attr"`
}

type ImmediateList struct {
	Val [2]Immediate
	Len uint8
}

func (il *ImmediateList) Encode(f *Format, args []operand.Arg) {
	for i := uint8(0); i < il.Len; i++ {
		switch v, mode := il.Val[i].Value.Value(); mode {
		case OptModeRef:
			switch arg := args[v].(type) {
			case operand.Int:
				f.Len += uint8(arg.Encode(f.Val[f.Len:], il.Val[i].Size))
			case operand.Uint:
				f.Len += uint8(arg.Encode(f.Val[f.Len:], il.Val[i].Size))
			}
		case OptModeValue:
			f.Len += uint8(operand.Uint(v).Encode(f.Val[f.Len:], il.Val[i].Size))
		}
	}
}

func (il *ImmediateList) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var i Immediate
	if err := d.DecodeElement(&i, &start); err != nil {
		return err
	}
	il.Val[il.Len] = i
	il.Len++
	return nil
}

type CodeOffset struct {
	// Size of the offset in bytes. Possible values are 1 or 4.
	Size operand.Size `xml:"size,attr"`
	// Value of the offset. Must be a reference to an instruction operand.
	Value OptRef `xml:"value,attr"`
}

func (co *CodeOffset) Encode(f *Format, args []operand.Arg) {
	if v, m := co.Value.Value(); m == OptModeRef && co.Size >= operand.Size8 {
		switch rel := args[v].(type) {
		case relFwd:
			f.Len += uint8(operand.Int(rel).Encode(f.Val[f.Len:], co.Size))
		case relRwd:
			rel -= relRwd(f.Len) + relRwd(co.Size.Bytes())
			f.Len += uint8(operand.Int(rel).Encode(f.Val[f.Len:], co.Size))
		}
	}
}

type DataOffset struct {
	// Size of the offset in bytes. Possible values are 4 or 8.
	Size operand.Size `xml:"size,attr"`
	// Value of the offset. Must be a reference to an instruction operand. The
	// instruction operand has "moffs" type of the matching size.
	Value OptRef `xml:"value,attr"`
}

func (do *DataOffset) Encode(f *Format, args []operand.Arg) {
	if _, m := do.Value.Value(); m == OptModeRef && do.Size >= operand.Size8 {
		panic("TODO: encode DataOffset")
	}
}
