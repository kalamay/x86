package main

import (
	"bytes"
	"fmt"
	"os"
	"unsafe"

	. "github.com/kalamay/x86/asm/amd64"
	. "github.com/kalamay/x86/asm/amd64/inst"
	"github.com/kalamay/x86/emit"
	"github.com/kalamay/x86/jit"
)

func main() {
	ops := []Op{MakeMem(RAX).WithSize(S32), Int(-1073741824)}

	fmt.Printf("operands:\n")
	for i, op := range ops {
		fmt.Printf("  %d: %s\n", i, op)
	}

	if inst, err := ADD.Select(ops); err == nil {
		fmt.Printf("match: %s\n", inst.Types)
	} else {
		fmt.Printf("error: %v\n", err)
	}

	w := bytes.Buffer{}

	var out int64
	e := emit.New(&w)
	e.MOV(RBX, Int(-123))
	e.ADD(RBX, Int(456))
	e.MOV(RAX, Uint(uintptr(unsafe.Pointer(&out))))
	e.MOV(MakeMem(RAX).WithSize(S64), RBX)
	e.RET()

	for _, err := range e.Errors {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}

	p := jit.NewPool(jit.PoolConfig{
		MinSize:  32,
		MaxSize:  1024,
		MapPages: 4,
	})

	alloc, err := p.Alloc(w.Bytes())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	} else {
		fmt.Printf("%spool = %s\n", alloc, p)
		jit.FuncOf(alloc)()
		fmt.Printf("out = %d\n", out)
	}
}
