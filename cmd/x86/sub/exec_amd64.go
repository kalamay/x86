package sub

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/kalamay/x86/cmd/x86/parser"
	"github.com/kalamay/x86/instruction"
	"github.com/kalamay/x86/operand"
	"github.com/kalamay/x86/x64"
	"github.com/kalamay/x86/x64/jit"
)

func (cli *ExecCmd) Run(data *instruction.Set) error {
	if cli.Debug {
		lldb, err := exec.LookPath("lldb")
		if err != nil {
			return err
		}

		me, err := os.Executable()
		if err != nil {
			return err
		}

		args := []string{lldb, me, "-O", "settings set target.x86-disassembly-flavor intel", "-o", "run", "--", "exec"}

		switch {
		case cli.File != nil:
			args = append(args, "-f")
			args = append(args, cli.File.Name())
		case len(cli.Input) > 0:
			args = append(args, cli.Input)
		default:
			return errors.New("stdin not supported")
		}

		return syscall.Exec(lldb, args, os.Environ())
	}

	buf := bytes.Buffer{}

	emit := x64.Emit{}
	emit.Open(x64.NewMachine(), &buf)

	pr := parser.NewPrint(&emit, operand.R15)

	p := parser.Parser{}
	p.SetDirectives(pr, parser.NewBreak(&emit))

	switch {
	case cli.File != nil:
		p.Init(cli.File.Name(), cli.File)
	case len(cli.Input) > 0:
		p.Init("<input>", strings.NewReader(cli.Input))
	default:
		p.Init("<stdin>", os.Stdin)
	}

	if err := p.Eval(data, &emit); err != nil {
		return err
	}

	for _, err := range emit.Close() {
		return err
	}

	pool := jit.NewPool(jit.PoolConfig{
		MinSize:  32,
		MaxSize:  4096,
		MapPages: 1,
	})

	alloc, err := pool.Alloc(buf.Bytes())
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	jit.FuncOf(alloc)()

	out := bufio.NewWriter(os.Stderr)
	pr.WriteTo(out)
	out.Flush()

	return nil
}
