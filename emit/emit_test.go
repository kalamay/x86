package emit

import (
	"bytes"
	. "math"
	"os/exec"
	"strconv"
	"strings"
	"testing"

	. "github.com/kalamay/x86/asm/amd64"
	. "github.com/kalamay/x86/asm/amd64/inst"
)

func TestMOV(t *testing.T) { testEmit(t, MOV) }

func TestADD(t *testing.T) { testEmit(t, ADD) }

var testImms = [...][]Op{
	S8:  []Op{Int(MinInt8), Int(MinInt8 / 2), Int(MaxInt8 / 2), Int(MaxInt8), Uint(0), Uint(MaxUint8 / 2), Uint(MaxUint8)},
	S16: []Op{Int(MinInt16), Int(MinInt16 / 2), Int(MaxInt16 / 2), Int(MaxInt16), Uint(0), Uint(MaxUint16 / 2), Uint(MaxUint16)},
	S32: []Op{Int(MinInt32), Int(MinInt32 / 2), Int(MaxInt32 / 2), Int(MaxInt32), Uint(0), Uint(MaxUint32 / 2), Uint(MaxUint32)},
	S64: []Op{Int(MinInt64), Int(MinInt64 / 2), Int(MaxInt64 / 2), Int(MaxInt64), Uint(0), Uint(MaxUint64 / 2), Uint(MaxUint64)},
}

var testRegs = [...][]Op{
	S8:  []Op{AL, BL, BH, R9B, SPL},
	S16: []Op{AX, CX, DI, R10W},
	S32: []Op{EAX, EDX, ESI, R11D},
	S64: []Op{RAX, RDX, RBP, R12},
}

var testMems = [...][]Op{
	S8:  collectMems(S8),
	S16: collectMems(S16),
	S32: collectMems(S32),
	S64: collectMems(S64),
}

func collectMems(s Size) []Op {
	return []Op{
		MakeMem(RAX).WithSize(s),
		MakeMem(RBX).WithSize(s).WithIndex(RDX, S32),
		MakeMem(RCX).WithSize(s).WithIndex(RBP, S64).WithDisplacement(16),
		MakeMem(R11).WithSize(s),
		MakeMem(R12).WithSize(s).WithIndex(R14, S32),
		MakeMem(R13).WithSize(s).WithIndex(R15, S64).WithDisplacement(24),
	}
}

type permute struct {
	at, n int
	ops   [][]Op
}

func makePermute(ops [][]Op) permute {
	n := len(ops[0])
	for i := 1; i < len(ops); i++ {
		n *= len(ops[i])
	}
	return permute{0, n, ops}
}

func (p *permute) HasNext() bool {
	return p.at < p.n
}

func (p *permute) Next() []Op {
	ops, at := p.ops, p.at
	out := make([]Op, len(ops))
	for i := 0; i < len(ops); i++ {
		out[i] = ops[i][at%len(ops[i])]
		at /= len(ops[i])
	}
	p.at++
	return out
}

func testEmit(t *testing.T, in *InstSet) {
	for i := 0; i < len(in.Inst); i++ {
		j, ops := 0, [4][]Op{}
		ty, ts := in.Inst[i].Types.Next()

		for ; ty > 0; j++ {
			if ty.IsImm() {
				ops[j] = append(ops[j], testImms[ty.ImmSize()]...)
			}
			if ty.IsReg() {
				ops[j] = append(ops[j], testRegs[ty.RegSize()]...)
			}
			if ty.IsMem() {
				ops[j] = append(ops[j], testMems[ty.MemSize()]...)
			}
			ty, ts = ts.Next()
		}

		for perm := makePermute(ops[:j]); perm.HasNext(); {
			args := perm.Next()

			t.Run(strconv.Itoa(perm.at), func(t *testing.T) {
				args := args
				t.Parallel()

				n, s := 0, [4]string{}
				for _, op := range args {
					s[n] = op.String()
					n++
				}

				t.Logf("%s %s", in.Name, strings.Join(s[:n], ", "))
				testComapre(t, func(e *Emit) { e.Emit(in, args) })
			})
		}
	}
}

/*
func TestEmitInst(t *testing.T) {
	var (
		m8        = MakeMem(RBX).WithSize(S8)
		m64       = MakeMem(RBX).WithSize(S64)
		m64i64    = m64.WithIndex(RCX, S64)
		m64i64d4  = m64i64.WithDisplacement(4)
		me64      = MakeMem(R11).WithSize(S64)
		me64i64   = m64.WithIndex(R12, S64)
		me64i64d4 = m64i64.WithDisplacement(4)
	)

	var tests = []struct {
		inst *InstSet
		ops  []Op
	}{
		{MOV, []Op{RBX, Int(-123)}},
		{MOV, []Op{R12, Int(-123)}},
		{MOV, []Op{RSI, Int(-123)}},
		{MOV, []Op{EBX, Int(123)}},
		{MOV, []Op{ESI, Int(123)}},
		{MOV, []Op{BX, Int(123)}},
		{MOV, []Op{BL, Int(123)}},
		{MOV, []Op{SI, Int(123)}},
		{MOV, []Op{BH, Int(123)}},
		{MOV, []Op{SIL, Int(123)}},
		{MOV, []Op{m8, BL}},
		{MOV, []Op{m8, BH}},
		{MOV, []Op{m8, SIL}},
		{MOV, []Op{m64, Int(123)}},
		{MOV, []Op{m64i64, Int(123)}},
		{MOV, []Op{m64i64d4, Int(123)}},
		{MOV, []Op{me64, Uint(123456789)}},
		{MOV, []Op{me64i64, Int(123456789)}},
		{MOV, []Op{me64i64d4, Int(123456789)}},
		{ADD, []Op{RBX, Int(-123)}},
		{ADD, []Op{R12, Int(-123)}},
		{ADD, []Op{RSI, Int(-123)}},
		{ADD, []Op{EBX, Int(123)}},
		{ADD, []Op{ESI, Int(123)}},
		{ADD, []Op{BX, Int(123)}},
		{ADD, []Op{BL, Int(123)}},
		{ADD, []Op{SI, Int(123)}},
		{ADD, []Op{BH, Int(123)}},
		{ADD, []Op{SIL, Int(123)}},
		{ADD, []Op{m8, BL}},
		{ADD, []Op{m8, BH}},
		{ADD, []Op{m8, SIL}},
		{ADD, []Op{m64, Int(123)}},
		{ADD, []Op{m64i64, Int(123)}},
		{ADD, []Op{m64i64d4, Int(123)}},
		{ADD, []Op{me64, Uint(123456789)}},
		{ADD, []Op{me64i64, Int(123456789)}},
		{ADD, []Op{me64i64d4, Int(123456789)}},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("emit-%d", i), func(t *testing.T) {
			test := test
			t.Parallel()

			n, s := 0, [4]string{}
			for _, op := range test.ops {
				s[n] = op.String()
				n++
			}

			t.Logf("%s %s", test.inst.Name, strings.Join(s[:n], ", "))

			testComapre(t, func(e *Emit) {
				e.Emit(test.inst, test.ops)
			})
		})
	}
}
*/

func testComapre(t *testing.T, fn func(e *Emit)) {
	var sout, serr, raw bytes.Buffer

	cmd := exec.Command("as", "-o", "-", "--")
	cmd.Stdout = &sout
	cmd.Stderr = &serr

	w, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("pipe failed: %v", err)
	}

	if err = cmd.Start(); err != nil {
		t.Fatalf("command failed: %v", err)
	}

	w.Write([]byte(".intel_syntax noprefix\n"))
	asm := Emit{Emitter: Assembly{}, Writer: w}
	x86 := Emit{Emitter: X86{}, Writer: &raw}
	fn(&asm)
	fn(&x86)
	w.Close()

	if err = cmd.Wait(); err != nil {
		if len(x86.Errors) == 0 {
			t.Fatalf("command failed: %v\n%s\n", err, serr.Bytes())
		} else {
			t.Logf("command failed successfully: %v\n%s\n", err, serr.Bytes())
		}
		return
	}

	for _, x86Err := range x86.Errors {
		t.Fatalf("failed emit: %v", x86Err)
	}

	expect := extractText(t, sout.Bytes())
	actual := raw.Bytes()
	if !bytes.Equal(expect, actual) {
		t.Errorf("gas comparison failed:\n    expect = %#v\n    actual = %#v\n", expect, actual)
	}
}
