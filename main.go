package main

import (
	"bytes"
	"fmt"
	"os"

	. "github.com/kalamay/x86/asm/amd64"
	. "github.com/kalamay/x86/asm/amd64/inst"
	"github.com/kalamay/x86/emit"
	"github.com/kalamay/x86/jit"
)

func main() {
	ops := []Op{RAX, MakeMem(RSP).WithDisplacement(8)}

	fmt.Printf("operands:\n")
	for i, op := range ops {
		fmt.Printf("  %d: %s\n", i, op)
	}

	if inst, err := MOV.Select(ops); err == nil {
		fmt.Printf("match: %s\n", inst.Types)
	} else {
		fmt.Printf("error: %v\n", err)
	}

	w := bytes.Buffer{}

	e := emit.New(&w)
	e.MOV(RAX, MakeMem(RSP).WithSize(S64).WithDisplacement(8))
	e.MOV(RBX, Int(-123))
	e.ADD(RBX, Int(456))
	//e.MOV(MakeMem(RAX), RBX)
	e.RET()

	for _, err := range e.Errors {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}

	p := jit.NewPool(jit.PoolConfig{
		MinSize:  16,
		MaxSize:  1024,
		MapPages: 4,
		Fill:     "none",
	})

	alloc, err := p.Alloc(w.Bytes())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	} else {
		fmt.Print(alloc)
		Invoke(alloc)
	}

	p.Free(&alloc)
	fmt.Printf("pool = %s\n", p)
}

func Invoke(alloc jit.Alloc) {
	var out int64
	var f func(*int64)

	jit.SetFunc(&f, alloc)
	f(&out)

	fmt.Printf("out(%p) = %d\n", &out, out)
}
