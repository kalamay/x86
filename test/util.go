package test

import (
	"bytes"
	"fmt"
	"math"
	"strings"
	"testing"

	. "github.com/kalamay/x86/asm/amd64"
	"github.com/kalamay/x86/emit"
)

type Expect struct {
	Types       TypeSet
	Permutation int
	Code        []byte
	Error       string
}

func FailInst(t *testing.T, in *InstSet, args []Op, format string, a ...interface{}) {
	n, s := 0, [4]string{}
	for _, op := range args {
		s[n] = op.String()
		n++
	}
	t.Errorf("%s %s\n%s", in.Name, strings.Join(s[:n], ", "), fmt.Sprintf(format, a...))
}

func Compare(t *testing.T, in *InstSet, expect []Expect) {
	var (
		ts  TypeSet
		ops [][]Op
		buf bytes.Buffer
	)

	for _, expect := range expect {
		if expect.Types != ts {
			ts = expect.Types
			ops, _ = OperandsOf(ts)
		}

		args := OperandsAt(ops, expect.Permutation)

		buf.Reset()
		em := emit.Emit{Emitter: emit.X86{}, Writer: &buf}
		em.Emit(in, args)

		switch {
		case len(expect.Error) == 0 && len(em.Errors) > 0:
			for _, err := range em.Errors {
				FailInst(t, in, args, "failed emit: %v", err)
			}
		case len(expect.Error) > 0 && len(em.Errors) == 0:
			FailInst(t, in, args, "expected error: %s", expect.Error)
		default:
			if !bytes.Equal(expect.Code, buf.Bytes()) {
				t.Errorf("incorrect code generated:\n    expect = %#v\n    actual = %#v\n", expect.Code, buf.Bytes())
			}
		}
	}
}

func OperandsOf(types TypeSet) ([][]Op, int) {
	ops, perms, i := [4][]Op{}, 1, 0
	ty, ts := types.Next()
	for ; ty > 0; i++ {
		if ty.IsImm() {
			ops[i] = append(ops[i], imm[ty.ImmSize()]...)
		}
		if ty.IsReg() {
			ops[i] = append(ops[i], reg[ty.RegSize()]...)
		}
		if ty.IsMem() {
			ops[i] = append(ops[i], mem[ty.MemSize()]...)
		}
		perms *= len(ops[i])
		ty, ts = ts.Next()
	}
	return ops[:i], perms
}

func OperandsAt(ops [][]Op, at int) []Op {
	out := make([]Op, len(ops))
	for i := 0; i < len(ops); i++ {
		out[i] = ops[i][at%len(ops[i])]
		at /= len(ops[i])
	}
	return out
}

var imm = [...][]Op{
	S8: []Op{
		Int(math.MinInt8),
		Int(math.MinInt8 / 2),
		Int(math.MaxInt8 / 2),
		Int(math.MaxInt8),
		Uint(0),
		Uint(math.MaxUint8 / 2),
		Uint(math.MaxUint8),
	},
	S16: []Op{
		Int(math.MinInt16),
		Int(math.MinInt16 / 2),
		Int(math.MaxInt16 / 2),
		Int(math.MaxInt16),
		Uint(0),
		Uint(math.MaxUint16 / 2),
		Uint(math.MaxUint16),
	},
	S32: []Op{
		Int(math.MinInt32),
		Int(math.MinInt32 / 2),
		Int(math.MaxInt32 / 2),
		Int(math.MaxInt32),
		Uint(0),
		Uint(math.MaxUint32 / 2),
		Uint(math.MaxUint32),
	},
	S64: []Op{
		Int(math.MinInt64),
		Int(math.MinInt64 / 2),
		Int(math.MaxInt64 / 2),
		Int(math.MaxInt64),
		Uint(0),
		Uint(math.MaxUint64 / 2),
		Uint(math.MaxUint64),
	},
}

var reg = [...][]Op{
	S8:  []Op{AL, BL, BH, R9B, SPL},
	S16: []Op{AX, CX, DI, R10W},
	S32: []Op{EAX, EDX, ESI, R11D},
	S64: []Op{RAX, RDX, RBP, R12},
}

var mem = [...][]Op{
	S8:  collectMem(S8),
	S16: collectMem(S16),
	S32: collectMem(S32),
	S64: collectMem(S64),
}

func collectMem(s Size) []Op {
	return []Op{
		MakeMem(RDX),
		MakeMem(RAX).WithSize(s),
		MakeMem(RBX).WithSize(s).WithIndex(RDX, S32),
		MakeMem(RCX).WithSize(s).WithIndex(RBP, S64).WithDisplacement(16),
		MakeMem(R11).WithSize(s),
		MakeMem(R12).WithSize(s).WithIndex(R14, S32),
		MakeMem(R13).WithSize(s).WithIndex(R15, S64).WithDisplacement(24),
	}
}
