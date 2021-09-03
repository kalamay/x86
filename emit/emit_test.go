package emit

import (
	"bytes"
	"os"
	"os/exec"
	"testing"

	. "github.com/kalamay/x86/asm/amd64"
)

var tests = []struct {
	name string
	fn   func(e *Emit)
}{
	{"mov1", func(e *Emit) {
		e.MOV(RBX, Int(-123))
	}},
	{"mov2", func(e *Emit) {
		e.MOV(EAX, Int(123))
	}},
	{"mov3", func(e *Emit) {
		e.MOV(MakeMem(RBX).WithSize(S64), Int(123))
	}},
	{"mov3", func(e *Emit) {
		e.MOV(MakeMem(RBX).WithSize(S64).WithIndex(RCX, S64), Int(123))
	}},
	{"mov3", func(e *Emit) {
		e.MOV(MakeMem(RBX).WithSize(S64).WithIndex(RCX, S64).WithDisplacement(4), Int(123))
	}},
}

func TestEmit(t *testing.T) {
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
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
			test.fn(&asm)
			test.fn(&x86)
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
		})
	}
}
