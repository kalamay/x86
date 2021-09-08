// +build ignore

package main

import (
	"bytes"
	"debug/macho"
	"errors"
	"flag"
	"fmt"
	"go/format"
	"os"
	"os/exec"
	"strings"
	"sync"
	"text/template"

	"github.com/kalamay/x86/emit"
	"github.com/kalamay/x86/test"

	. "github.com/kalamay/x86/asm/amd64"
	_ "github.com/kalamay/x86/asm/amd64/inst"
)

var gen = template.Must(template.New("").Funcs(template.FuncMap{
	"typeset": func(ts TypeSet) uint64 { return uint64(ts) },
	"bytes": func(b []byte) string {
		if b == nil {
			return "nil"
		}
		return fmt.Sprintf("%#v", b)
	},
}).Parse(`package encode

import (
	"testing"

	"github.com/kalamay/x86/test"
	"github.com/kalamay/x86/asm/amd64/inst"
)

func Test{{.Inst.Name}}(t *testing.T) {
	test.Compare(t, inst.{{.Inst.Name}}, expect{{.Inst.Name}}[:])
}

var expect{{.Inst.Name}} = [...]test.Expect {
	{{range .Tests}} { {{typeset .Types}}, {{.Permutation}}, {{bytes .Code}}, {{printf "%q" .Error}} },
	{{end}}
}
`))

func main() {
	name, out := "", "-"
	flag.StringVar(&name, "n", name, "instruction name")
	flag.StringVar(&out, "o", out, "output file")
	flag.Parse()

	ins, ok := LookupInst(strings.ToUpper(name))
	if !ok {
		fmt.Fprintf(os.Stderr, "instruction not defined: %q\n", name)
		os.Exit(1)
	}

	tests := []test.Expect{}
	for _, in := range ins.Inst {
		ops, perms := test.OperandsOf(in.Types)
		wg := sync.WaitGroup{}
		tmp := make([]test.Expect, perms)

		for i := 0; i < perms; i++ {
			wg.Add(1)

			go func(i int) {
				code, err := assemble(ins, test.OperandsAt(ops, i))
				errmsg := ""
				if err != nil {
					errmsg = err.Error()
				}
				tmp[i] = test.Expect{
					Types:       in.Types,
					Permutation: i,
					Code:        code,
					Error:       errmsg,
				}
				wg.Done()
			}(i)

			if (i & 15) == 15 {
				wg.Wait()
			}
		}

		wg.Wait()

		tests = append(tests, tmp...)
	}

	data := struct {
		Inst  *InstSet
		Tests []test.Expect
	}{
		Inst:  ins,
		Tests: tests,
	}

	buf := bytes.Buffer{}
	err := gen.Execute(&buf, &data)
	if err != nil {
		panic(err)
	}

	src, err := format.Source(buf.Bytes())
	if err != nil {
		panic(err)
	}

	if out == "-" {
		os.Stdout.Write(src)
	} else {
		f, err := os.Create(out)
		if err != nil {
			panic(err)
		}
		f.Write(src)
	}
}

func assemble(ins *InstSet, ops []Op) ([]byte, error) {
	var sout, serr bytes.Buffer

	cmd := exec.Command("as", "-o", "-", "--")
	cmd.Stdout = &sout
	cmd.Stderr = &serr

	w, err := cmd.StdinPipe()
	if err != nil {
		fatal("pipe failed: %v", err)
	}

	if err = cmd.Start(); err != nil {
		fatal("command failed: %v", err)
	}

	w.Write([]byte(".intel_syntax noprefix\n"))
	asm := emit.Emit{Emitter: emit.Assembly{}, Writer: w}
	asm.Emit(ins, ops)
	w.Close()

	if err = cmd.Wait(); err != nil {
		if serr.Len() > 0 {
			return nil, errors.New(string(serr.Bytes()))
		}
		return nil, fmt.Errorf("command failed: %v", err)
	}

	return extractText(sout.Bytes()), nil
}

func extractText(b []byte) []byte {
	f, err := macho.NewFile(bytes.NewReader(b))
	if err != nil {
		fatal("macho failed: %v", err)
	}

	s := f.Section("__text")
	if s == nil {
		fatal("macho missing __text")
	}

	d, err := s.Data()
	if err != nil {
		fatal("macho data failed: %v", err)
	}

	return d
}

func fatal(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
	os.Exit(1)
}
