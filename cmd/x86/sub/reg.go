package sub

import (
	"bufio"
	"fmt"
	"os"

	"github.com/kalamay/x86/instruction"
	"github.com/kalamay/x86/operand"
)

type RegCmd struct {
	Names []string `arg:"" name:"name" help:"Names to show information for."`
}

func (cli *RegCmd) Run(data *instruction.Set) error {
	regs := make([]operand.Reg, len(cli.Names))

	for i, name := range cli.Names {
		if reg, ok := operand.RegOf(name); ok {
			regs[i] = reg
		} else {
			return fmt.Errorf("unknown register %q", name)
		}
	}

	buf := bufio.NewWriter(os.Stdout)
	buf.WriteString("NAME   ID  SIZE HI 8  16 TYPE\n")
	for _, reg := range regs {
		hi, ex8, ex16 := "✗", "✗", "✗"
		if reg.HighByte() {
			hi = "✓"
		}
		if reg.Next8() {
			ex8 = "✓"
		}
		if reg.Next16() {
			ex16 = "✓"
		}
		fmt.Fprintf(buf, "%-6s %-3d %-4d %-2s %-2s %-2s ", reg, reg.ID(), reg.Size().Bits(), hi, ex8, ex16)
		switch reg.Type() {
		case operand.RegTypeGeneral:
			buf.WriteString("general\n")
		case operand.RegTypeVector:
			buf.WriteString("vector\n")
		case operand.RegTypeMask:
			buf.WriteString("mask\n")
		case operand.RegTypeIP:
			buf.WriteString("instruction pointer\n")
		case operand.RegTypeStatus:
			buf.WriteString("status\n")
		case operand.RegTypeSegment:
			buf.WriteString("segment\n")
		}
	}
	buf.Flush()

	return nil
}
