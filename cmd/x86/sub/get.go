package sub

import (
	"bufio"
	"fmt"
	"os"

	"github.com/kalamay/x86/instruction"
)

type GetCmd struct {
	Full bool `short:"f" help:"Output full YAML."`

	Names []string `arg:"" name:"name" help:"Names to show information for."`
}

func (cli *GetCmd) Run(data *instruction.Set) error {
	insts := make([]*instruction.Instruction, len(cli.Names))

	for i, name := range cli.Names {
		if in := data.Lookup(name); in != nil {
			insts[i] = in
		} else {
			return fmt.Errorf("unknown mnemonic %q", name)
		}
	}

	buf := bufio.NewWriter(os.Stdout)
	for _, in := range insts {
		if cli.Full {
			getFull(buf, in)
		} else {
			getDefault(buf, in)
		}
	}
	buf.Flush()

	return nil
}

func getFull(buf *bufio.Writer, in *instruction.Instruction) {
	fmt.Fprintf(buf, "---\nName: %s\nSummary: %s\nForms:\n", in.Name, in.Summary)
	for _, f := range in.Forms {
		fmt.Fprintf(buf, "- GasName: %s\n", f.GasName)
		if len(f.GoName) > 0 {
			fmt.Fprintf(buf, "  GoName: %s\n", f.GoName)
		}
		if f.Encoding.Prefix.Len > 0 {
			buf.WriteString(`  Prefix: "`)
			for i := uint8(0); i < f.Encoding.Prefix.Len; i++ {
				fmt.Fprintf(buf, "%02x", f.Encoding.Prefix.Val[i])
			}
			buf.WriteString("\"\n")
		}
		buf.WriteString(`  Opcode: "`)
		for i := uint8(0); i < f.Encoding.Opcode.Len; i++ {
			fmt.Fprintf(buf, "%02x", f.Encoding.Opcode.Val[i].Byte)
		}
		buf.WriteString("\"\n")
		if f.Operands.Len > 0 {
			buf.WriteString("  Operands: [")
			for i := uint8(0); i < f.Operands.Len; i++ {
				if i > 0 {
					buf.WriteString(", ")
				}
				buf.WriteString(f.Operands.Val[i].String())
			}
			buf.WriteString("]\n")
		}
		if f.ISA != 0 {
			buf.WriteString("  ISA: [")
			for i, n := range f.ISA.Names() {
				if i > 0 {
					buf.WriteString(", ")
				}
				buf.WriteString(n)
			}
			buf.WriteString("]\n")
		}
	}
}

func getDefault(buf *bufio.Writer, in *instruction.Instruction) {
	max, maxop, maxgas := [6]int{}, uint8(0), 0
	for _, f := range in.Forms {
		for i := uint8(0); i < f.Operands.Len; i++ {
			if n := len(f.Operands.Val[i].String()); n > max[i] {
				max[i] = n
			}
		}
		if f.Encoding.Opcode.Len > maxop {
			maxop = f.Encoding.Opcode.Len
		}
		if len(f.GasName) > maxgas {
			maxgas = len(f.GasName)
		}
	}

	full := len(in.Name) + 2
	for i := uint8(0); i < 6 && max[i] > 0; i++ {
		full += max[i] + 1
	}

	fmt.Fprintf(buf, "---\nName: %s\nSummary: %s\n%-*s  # %-*s  %-*s  ISA\n",
		in.Name,
		in.Summary,
		full, "Forms:",
		maxgas, "GAS",
		maxop*2, "OP",
	)
	for _, f := range in.Forms {
		enc := &f.Encoding
		fmt.Fprintf(buf, "- %s", in.Name)
		for i := uint8(0); i < f.Operands.Len; i++ {
			fmt.Fprintf(buf, " %-*s", max[i], f.Operands.Val[i])
		}
		fmt.Fprintf(buf, "  # %-*s  ", maxgas, f.GasName)
		for i := uint8(0); i < enc.Opcode.Len; i++ {
			fmt.Fprintf(buf, "%02x", enc.Opcode.Val[i].Byte)
		}
		if f.ISA != 0 {
			buf.WriteString("  ")
			for i, n := range f.ISA.Names() {
				if i > 0 {
					buf.WriteByte(',')
				}
				buf.WriteString(n)
			}
		}
		buf.WriteByte('\n')
	}
}
