package operand

import (
	"fmt"
	"testing"
)

func TestOperandNames(t *testing.T) {
	fmt.Printf("mask: %032b\n", pKeyMask&kindMask)
	for name, param := range paramTypes {
		str := param.String()
		if name != str {
			t.Errorf("string match failed: expect=%q, actual=%q", name, str)
		}
	}
}
