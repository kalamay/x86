package sub

import (
	"bufio"
	"os"

	"github.com/kalamay/x86/instruction"
)

type ListCmd struct {
	Mnemonic string `short:"n" help:"Mnemonic format." default:"intel" enum:"intel,gas,go"`
}

func (cli *ListCmd) Run(data *instruction.Set) error {
	buf := bufio.NewWriter(os.Stdout)
	if cli.Mnemonic == "intel" {
		for _, inst := range data.Instructions {
			buf.WriteString(inst.Name)
			buf.WriteByte('\n')
		}
	} else {
		prev, name := "", ""
		for _, inst := range data.Instructions {
			for _, f := range inst.Forms {
				switch cli.Mnemonic {
				case "go":
					name = f.GoName
				case "gas":
					name = f.GasName
				}
				if name != "" && name != prev {
					buf.WriteString(name)
					buf.WriteByte('\n')
					prev = name
				}
			}
		}
	}
	buf.Flush()
	return nil
}
