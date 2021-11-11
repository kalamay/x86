package x64

import (
	"bytes"
	"fmt"
)

type Assembly struct {
	buf bytes.Buffer
}

func NewAssembly() *Assembly {
	return &Assembly{}
}

func (a *Assembly) Open() {
}

func (a *Assembly) Emit(e *Emit, call *EmitCall) {
	a.buf.Reset()
	a.buf.WriteString(call.Instruction.Name)
	for i, op := range call.Args {
		if i > 0 {
			a.buf.WriteByte(',')
		}
		a.buf.WriteByte(' ')
		a.buf.WriteString(op.String())
	}
	a.buf.WriteByte('\n')
	e.Write(a.buf.Bytes())
}

func (a *Assembly) Label(e *Emit, label *EmitLabel) {
	fmt.Fprintf(e, "%s:\n", label.Name())
}

func (a *Assembly) Close(e *Emit) {
}
