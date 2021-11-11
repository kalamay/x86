package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"

	"github.com/kalamay/x86/cmd/x86/sub"
	"github.com/kalamay/x86/x64"
)

type Cmd struct {
	XML *os.File `short:"x" help:"XML source file."`

	As   sub.AsCmd   `cmd:"" help:"Parse and assemble instructions."`
	Exec sub.ExecCmd `cmd:"" help:"Assemble and execute instructions."`
	List sub.ListCmd `cmd:"" help:"List command names."`
	Get  sub.GetCmd  `cmd:"" help:"Show instruction information."`
	Reg  sub.RegCmd  `cmd:"" help:"Show register information."`
	Gen  sub.GenCmd  `cmd:"" help:"Generate instructions."`
}

func main() {
	var err error
	data := x64.Instructions

	cli := Cmd{}
	ctx := kong.Parse(&cli)

	if cli.XML != nil {
		err = data.Load(cli.XML)
		cli.XML.Close()
	}

	if err == nil {
		err = ctx.Run(&data)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
