package x64

import (
	"errors"

	"github.com/kalamay/x86/instruction"
	"github.com/kalamay/x86/operand"
)

var (
	ErrUnsupportedInstruction = errors.New("unsupported instruction")
	ErrAmbiguousOperandSize   = errors.New("ambiguous operand size")
)

func Select(in *instruction.Instruction, args []operand.Arg) (*instruction.Encoding, error) {
	for _, arg := range args {
		if err := arg.Validate(); err != nil {
			return nil, err
		}
	}

	for i := 0; i < len(in.Forms); i++ {
		if matchOperands(in.Forms[i].Operands, args) {
			return &in.Forms[i].Encoding, nil
		}
	}

	return nil, ErrUnsupportedInstruction
}

func matchOperands(params operand.ParamList, args []operand.Arg) bool {
	for a, i := 0, uint8(0); i < params.Len; i++ {
		p := params.Val[i]
		if p.ImmConst() {
			continue
		}
		if a >= len(args) || !args[i].Matches(p) {
			return false
		}
		a++
	}
	return true
}
