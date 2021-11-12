package instruction

import (
	"encoding/xml"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/kalamay/x86/operand"
)

type Set struct {
	XMLName      xml.Name      `xml:"InstructionSet"`
	Name         string        `xml:"name,attr"`
	Instructions []Instruction `xml:"Instruction"`
}

func (is *Set) Load(r io.Reader) (err error) {
	v := Set{}

	dec := xml.NewDecoder(r)
	if err = dec.Decode(&v); err != nil {
		return
	}

	sort.Slice(v.Instructions, func(i, j int) bool {
		return v.Instructions[i].Name < v.Instructions[j].Name
	})

	for i := range v.Instructions {
		inst := &v.Instructions[i]
		for f := range inst.Forms {
			form := &inst.Forms[f]
			for o := uint8(0); o < form.Operands.Len; o++ {
				form.Operands.Val[o].ApplyMnemonic(inst.Name)
			}
		}
		sort.SliceStable(inst.Forms, func(i, j int) bool {
			return inst.Forms[i].score() < inst.Forms[j].score()
		})
		inst.Summary = strings.TrimSpace(inst.Summary)
	}

	*is = v
	return nil
}

func (is *Set) Lookup(name string) *Instruction {
	name = strings.ToUpper(name)
	for lo, hi := 0, len(is.Instructions); lo < hi; {
		mid := int(uint(lo+hi) >> 1)
		switch {
		case name == is.Instructions[mid].Name:
			return &is.Instructions[mid]
		case name < is.Instructions[mid].Name:
			hi = mid
		default:
			lo = mid + 1
		}
	}
	return nil
}

type Instruction struct {
	// Name is the instruction name in Intel-style assembly (PeachPy, NASM and YASM assemblers).
	Name string `xml:"name,attr"`
	// Summary describes the instruction name.
	Summary string `xml:"summary,attr"`
	// Forms is a list of InstructionForm values representing the instruction forms.
	Forms []Form `xml:"InstructionForm"`
}

type Form struct {
	// GasName is the instruction form name in GNU assembler (gas).
	GasName string `xml:"gas-name,attr"`
	// GoName is the instruction form name in Go/Plan 9 assembler (8a). An empty
	// value means it is not supported by the Go/Plan 9 assembler.
	GoName string `xml:"go-name,attr"`
	// ISA is the set of instruction set architectures required for the form.
	ISA ISA `xml:"ISA"`
	// MmxMode indicated the MMX technology state required or forced by this instruction.
	MmxMode MmxMode `xml:"mmx-mode,attr"`
	// XmmMode indicates XMM registers are accessed by this instruction.
	XmmMode XmmMode `xml:"xmm-mode,attr"`
	// CancelingInputs indicates that the instruction form has no dependency on
	// the values of input operands when they refer to the same register. For
	// example, "VPXOR xmm1, xmm0, xmm0" does not depend on "xmm0". Instruction
	// forms with cancelling inputs have only two input operands, which have the
	// same register type.
	CancelingInputs bool `xml:"cancelling-inputs,attr"`
	// Operands are the list of constraints required for the arguments used to encode
	// this form.
	Operands operand.ParamList `xml:"Operand"`
	Encoding Encoding          `xml:"Encoding"`
}

func (f *Form) score() int {
	return f.Encoding.score()
}

type ISA uint64

const (
	RDTSC           ISA = 1 << iota // The RDTSC instruction.
	RDTSCP                          // The RDTSCP instruction.
	CPUID                           // The CPUID instruction.
	FEMMS                           // The FEMMS instruction.
	MOVBE                           // The MOVBE instruction.
	POPCNT                          // The POPCNT instruction.
	LZCNT                           // The LZCNT instruction.
	PCLMULQDQ                       // The PCLMULQDQ instruction.
	RDRAND                          // The RDRAND instruction.
	RDSEED                          // The RDSEED instruction.
	CLFLUSH                         // The CLFLUSH instruction.
	CLFLUSHOPT                      // The CLFLUSHOPT instruction.
	CLWB                            // The CLWB instruction.
	CLZERO                          // The CLZERO instruction.
	PREFETCH                        // The PREFETCH instruction (3dnow! Prefetch).
	PREFETCHW                       // The PREFETCHW instruction (3dnow! Prefetch/Intel PRFCHW).
	PREFETCHWT1                     // The PREFETCHWT1 instruction.
	MONITOR                         // The MONITOR and MWAIT instructions.
	MONITORX                        // The MONITORX and MWAITX instructions.
	CMOV                            // Conditional MOVe instructions.
	MMX                             // MultiMedia eXtension.
	MMXPlus                         // AMD MMX+ extension / Integer SSE (Intel).
	ThreeDNow                       // AMD 3dnow! extension.
	ThreeDNowPlus                   // AMD 3dnow!+ extension.
	ThreeDNowGeode                  // AMD 3dnow! Geode extension.
	SSE                             // Streaming SIMD Extension.
	SSE2                            // Streaming SIMD Extension 2.
	SSE3                            // Streaming SIMD Extension 3.
	SSSE3                           // Supplemental Streaming SIMD Extension 3.
	SSE4_1                          // Streaming SIMD Extension 4.1.
	SSE4_2                          // Streaming SIMD Extension 4.2.
	SSE4A                           // Streaming SIMD Extension 4a.
	AVX                             // Advanced Vector eXtension.
	AVX2                            // Advanced Vector eXtension 2.
	AVX512F                         // AVX-512 Foundation instructions.
	AVX512BW                        // AVX-512 Byte and Word instructions.
	AVX512DQ                        // AVX-512 Doubleword and Quadword instructions.
	AVX512VL                        // AVX-512 Vector Length extension (EVEX-encoded XMM/YMM operations).
	AVX512PF                        // AVX-512 Prefetch instructions.
	AVX512ER                        // AVX-512 Exponential and Reciprocal instructions.
	AVX512CD                        // AVX-512 Conflict Detection instructions.
	AVX512VBMI                      // AVX-512 Vector Bit Manipulation instructions.
	AVX512IFMA                      // AVX-512 Integer 52-bit Multiply-Accumulate instructions.
	AVX512VPOPCNTDQ                 // AVX-512 Vector Population Count instructions.
	XOP                             // eXtended OPerations extension.
	F16C                            // Half-Precision (F16) Conversion instructions.
	FMA3                            // Fused Multiply-Add instructions (3-operand).
	FMA4                            // Fused Multiply-Add instructions (4-operand).
	BMI                             // Bit Manipulation Instructions.
	BMI2                            // Bit Manipulation Instructions 2.
	TBM                             // Trailing Bit Manipulation instructions.
	ADX                             // The ADCX and ADOX instructions.
	AES                             // AES instruction set.
	SHA                             // SHA instruction set.
)

func (i ISA) Names() (names []string) {
	for name, v := range isaNames {
		if (i & v) == v {
			names = append(names, name)
		}
	}
	return
}

func (i *ISA) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	for _, attr := range start.Attr {
		if attr.Name.Local == "id" {
			if v, ok := isaNames[attr.Value]; ok {
				*i |= v
				break
			}
			return fmt.Errorf("ISA: unknown id %q", attr.Value)
		}
	}
	return d.Skip()
}

var isaNames = map[string]ISA{
	"RDTSC":           RDTSC,
	"RDTSCP":          RDTSCP,
	"CPUID":           CPUID,
	"FEMMS":           FEMMS,
	"MOVBE":           MOVBE,
	"POPCNT":          POPCNT,
	"LZCNT":           LZCNT,
	"PCLMULQDQ":       PCLMULQDQ,
	"RDRAND":          RDRAND,
	"RDSEED":          RDSEED,
	"CLFLUSH":         CLFLUSH,
	"CLFLUSHOPT":      CLFLUSHOPT,
	"CLWB":            CLWB,
	"CLZERO":          CLZERO,
	"PREFETCH":        PREFETCH,
	"PREFETCHW":       PREFETCHW,
	"PREFETCHWT1":     PREFETCHWT1,
	"MONITOR":         MONITOR,
	"MONITORX":        MONITORX,
	"CMOV":            CMOV,
	"MMX":             MMX,
	"MMX+":            MMXPlus,
	"3dnow!":          ThreeDNow,
	"3dnow!+":         ThreeDNowPlus,
	"3dnow! Geode":    ThreeDNowGeode,
	"SSE":             SSE,
	"SSE2":            SSE2,
	"SSE3":            SSE3,
	"SSSE3":           SSSE3,
	"SSE4.1":          SSE4_1,
	"SSE4.2":          SSE4_2,
	"SSE4A":           SSE4A,
	"AVX":             AVX,
	"AVX2":            AVX2,
	"AVX512F":         AVX512F,
	"AVX512BW":        AVX512BW,
	"AVX512DQ":        AVX512DQ,
	"AVX512VL":        AVX512VL,
	"AVX512PF":        AVX512PF,
	"AVX512ER":        AVX512ER,
	"AVX512CD":        AVX512CD,
	"AVX512VBMI":      AVX512VBMI,
	"AVX512IFMA":      AVX512IFMA,
	"AVX512VPOPCNTDQ": AVX512VPOPCNTDQ,
	"XOP":             XOP,
	"F16C":            F16C,
	"FMA3":            FMA3,
	"FMA4":            FMA4,
	"BMI":             BMI,
	"BMI2":            BMI2,
	"TBM":             TBM,
	"ADX":             ADX,
	"AES":             AES,
	"SHA":             SHA,
}

type MmxMode uint8

const (
	MmxModeNone MmxMode = iota // Instruction neither affects nor cares about the MMX technology state.
	MmxModeFPU                 // Instruction requires the MMX technology state to be clear.
	MmxModeMMX                 // Instruction causes transition to MMX technology state.
)

func (m *MmxMode) UnmarshalText(text []byte) error {
	if v, ok := mmxModes[string(text)]; ok {
		*m = v
		return nil
	}
	return fmt.Errorf("MmxMode: unknown type %q", text)
}

var mmxModes = map[string]MmxMode{
	"":    MmxModeNone,
	"FPU": MmxModeFPU,
	"MMX": MmxModeMMX,
}

type XmmMode uint8

const (
	XmmModeNone XmmMode = iota // Instruction does not affect XMM registers and does not change XMM registers access mode.
	XmmModeSSE                 // Instruction accesses XMM registers in legacy SSE mode.
	XmmModeAVX                 // Instruction accesses XMM registers in AVX mode.
)

func (m *XmmMode) UnmarshalText(text []byte) error {
	if v, ok := xmmModes[string(text)]; ok {
		*m = v
		return nil
	}
	return fmt.Errorf("XmmMode: unknown type %q", text)
}

var xmmModes = map[string]XmmMode{
	"":    XmmModeNone,
	"SSE": XmmModeSSE,
	"AVX": XmmModeAVX,
}
