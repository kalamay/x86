package operand

import (
	"testing"
)

func TestOperandNames(t *testing.T) {
	for name, param := range paramTypes {
		str := param.String()
		if name != str {
			t.Errorf("string match failed: expect=%q, actual=%q", name, str)
		}
	}
}

/*
func TestParamValues(t *testing.T) {
	tests := struct {
		Kind         Kind
		Implicit     bool
		Input        bool
		Output       bool
		Const        bool
		Masked       bool
		MergeMasked  bool
		ExtendedSize Size
	}{}
}
*/
