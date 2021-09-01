package main

import (
	"encoding/hex"
	"fmt"

	. "github.com/kalamay/x86/asm/amd64"
	. "github.com/kalamay/x86/asm/amd64/inst"
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

	enc := [15]byte{}
	n, err := MOV.Encode(enc[:], ops[:2])
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("%s\n", hex.EncodeToString(enc[:n]))
	}
}
