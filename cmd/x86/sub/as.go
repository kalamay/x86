package sub

import (
	"bytes"
	"os"
	"strings"

	"github.com/kalamay/x86/cmd/x86/parser"
	"github.com/kalamay/x86/instruction"
	"github.com/kalamay/x86/x64"
)

type AsCmd struct {
	File *os.File `short:"f" help:"Load assembly from specified file."`

	Input string `arg:"" optional:"" help:"Input to assemble instead of stdin."`
}

func (cli *AsCmd) Run(data *instruction.Set) error {
	p := parser.Parser{}
	switch {
	case cli.File != nil:
		p.Init(cli.File.Name(), cli.File)
	case len(cli.Input) > 0:
		p.Init("<input>", strings.NewReader(cli.Input))
	default:
		p.Init("<stdin>", os.Stdin)
	}

	p.SetDirectives(parser.PrintNop, parser.BreakNop)

	buf := bytes.Buffer{}

	e := x64.Emit{}
	e.Open(x64.NewMachine(), &buf)
	if err := p.Eval(data, &e); err != nil {
		return err
	}
	for _, err := range e.Close() {
		return err
	}

	os.Stdout.Write(buf.Bytes())
	return nil
}
