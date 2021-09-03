package emit

import (
	"bytes"
	"io"

	. "github.com/kalamay/x86/asm/amd64"
	. "github.com/kalamay/x86/asm/amd64/inst"
)

type Emitter interface {
	Emit(w io.Writer, in *InstSet, ops []Op) error
}

type Emit struct {
	Errors  []error
	Emitter Emitter

	w io.Writer
}

func New(w io.Writer) *Emit {
	return &Emit{Emitter: X86{}, w: w}
}

func (e *Emit) Emit(in *InstSet, ops []Op) (err error) {
	if err = e.Emitter.Emit(e.w, in, ops); err != nil {
		e.Errors = append(e.Errors, err)
		e.w = io.Discard
	}
	return
}

func (e *Emit) ADD(ops ...Op) error { return e.Emit(ADD, ops) }
func (e *Emit) MOV(ops ...Op) error { return e.Emit(MOV, ops) }

type X86 struct{}

func (x X86) Emit(w io.Writer, in *InstSet, ops []Op) (err error) {
	n, b := 0, [15]byte{}
	if n, err = in.Encode(b[:], ops); err == nil {
		_, err = w.Write(b[:n])
	}
	return
}

type Assembly struct {
	buf bytes.Buffer
}

func (a Assembly) Emit(w io.Writer, in *InstSet, ops []Op) (err error) {
	a.buf.Reset()
	a.buf.WriteString(in.Name)
	for i, op := range ops {
		if i > 0 {
			a.buf.WriteByte(',')
		}
		a.buf.WriteByte(' ')
		a.buf.WriteString(op.String())
	}
	a.buf.WriteByte('\n')
	_, err = w.Write(a.buf.Bytes())
	return
}
