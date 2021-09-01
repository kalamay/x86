package main

import (
	"encoding/hex"
	"fmt"
	"os"

	. "github.com/kalamay/x86/asm/amd64"
	. "github.com/kalamay/x86/asm/amd64/inst"
	"github.com/kalamay/x86/emit"
)

func main() {
	ops := []Op{
		RBX,
		Int(-123),
		MakeMem(RBX).WithSize(S64),
		MakeMem(RCX).WithSize(S32).WithIndex(EAX, S32).WithDisplacement(Int(12)),
	}

	fmt.Printf("operands:\n")
	for i, op := range ops {
		fmt.Printf("  %d: %s\n", i, op)
	}

	if inst, err := MOV.Select(ops[:2]); err == nil {
		fmt.Printf("match: %s\n", inst.Types)
	}

	e := emit.New(hex.NewEncoder(os.Stdout))
	e.MOV(RBX, Int(-123))
	e.MOV(EAX, Int(123))

	for _, err := range e.Errors {
		fmt.Printf("Error: %v\n", err)
	}
	fmt.Printf("\n")
}
