package emit

import (
	"io"

	. "github.com/kalamay/x86/asm/amd64"
	. "github.com/kalamay/x86/asm/amd64/inst"
)

func New(w io.Writer) *Emitter {
	return &Emitter{w: w}
}

type Emitter struct {
	Errors []error

	w io.Writer
}

func (e *Emitter) Emit(in *InstSet, ops []Op) (n int, err error) {
	b := [15]byte{}
	if n, err = in.Encode(b[:], ops); err != nil {
		n = 0
		e.Errors = append(e.Errors, err)
	} else if len(e.Errors) == 0 {
		_, err = e.w.Write(b[:n])
	}
	return
}

func (e *Emitter) ADD(ops ...Op) (n int, err error) { return e.Emit(ADD, ops) }
func (e *Emitter) MOV(ops ...Op) (n int, err error) { return e.Emit(MOV, ops) }
