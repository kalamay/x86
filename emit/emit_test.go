package emit

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	. "github.com/kalamay/x86/asm/amd64"
	. "github.com/kalamay/x86/asm/amd64/inst"
)

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

			testEmit(t, func(e *Emit) {
				e.Emit(test.inst, test.ops)
			})
		})
	}
}

func testEmit(t *testing.T, fn func(e *Emit)) {
	var obj, raw bytes.Buffer

	cmd := exec.Command("gcc-11", "-c", "-x", "assembler", "-o", "-", "-")
	cmd.Stdout = &obj
	cmd.Stderr = os.Stderr

	w, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("pipe failed: %v", err)
	}

	if err = cmd.Start(); err != nil {
		t.Fatalf("command failed: %v", err)
	}

	w.Write([]byte(".intel_syntax noprefix\n"))
	asm := Emit{Emitter: Assembly{}, w: w}
	x86 := Emit{Emitter: X86{}, w: &raw}
	fn(&asm)
	fn(&x86)
	w.Close()

	if err = cmd.Wait(); err != nil {
		t.Fatalf("command failed: %v", err)
	}

	if len(asm.Errors) > 0 || len(x86.Errors) > 0 {
		for _, err = range asm.Errors {
			t.Errorf("failed emit: %v", err)
		}
		for _, err = range x86.Errors {
			t.Errorf("failed emit: %v", err)
		}
		return
	}

	expect := extractText(t, obj.Bytes())
	actual := raw.Bytes()
	if !bytes.Equal(expect, actual) {
		t.Errorf("gas comparison failed:\n    expect = %#v\n    actual = %#v\n", expect, actual)
	}
}
