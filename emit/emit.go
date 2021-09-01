package emit

import (
	"io"

	"github.com/kalamay/x86/asm/amd64"
	"github.com/kalamay/x86/asm/amd64/inst"
)

func New(w io.Writer) *Emitter {
	return &Emitter{w: w}
}

type Emitter struct {
	Errors []error

	w io.Writer
}

func (e *Emitter) MOV(ops ...amd64.Op) (n int, err error) {
	b := [15]byte{}
	if n, err = inst.MOV.Encode(b[:], ops); err != nil {
		n = 0
		e.Errors = append(e.Errors, err)
	} else if len(e.Errors) == 0 {
		_, err = e.w.Write(b[:n])
	}
	return
}
