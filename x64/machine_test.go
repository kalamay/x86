package x64

import (
	"bytes"
	"testing"

	. "github.com/kalamay/x86/operand"
)

func TestMachine(t *testing.T) {
	buf := bytes.Buffer{}
	e := Emit{}

	e.Open(NewMachine(), &buf)
	e.JMP(Label("a"))
	e.MOV(RBX, Int(123))
	e.Label("a")
	e.MOV(BX, Int(123))
	e.JMP(Label("a"))
	e.VPAND(XMM0, XMM1, XMM2)
	e.VPAND(XMM0, XMM12, XMM2)
	e.VPAND(XMM0, XMM1, XMM12)
	e.VPAND(YMM12, YMM13, YMM14)
	for _, err := range e.Close() {
		t.Error(err)
	}

	expect := [...]byte{
		0xeb, 0x07,
		0x48, 0xc7, 0xc3, 0x7b, 0x00, 0x00, 0x00,
		0x66, 0xbb, 0x7b, 0x00,
		0xeb, 0xfa,
		0xc5, 0xf1, 0xdb, 0xc2,
		0xc5, 0x99, 0xdb, 0xc2,
		0xc4, 0xc1, 0x71, 0xdb, 0xc4,
		0xc4, 0x41, 0x15, 0xdb, 0xe6,
	}
	if !bytes.Equal(expect[:], buf.Bytes()) {
		t.Errorf("failed to encode:\n\texpect = %#v\n\tactual = %#v", expect[:], buf.Bytes())
	}
}

func BenchmarkMachine(b *testing.B) {
	buf := bytes.Buffer{}
	e := Emit{}
	e.Open(NewMachine(), &buf)

	for i := 0; i < b.N; i++ {
		buf.Reset()
		e.JMP(Label("a"))
		e.MOV(EBX, Int(123))
		e.Label("a")
		e.MOV(BX, Int(123))
		e.JMP(Label("a"))
	}
}
