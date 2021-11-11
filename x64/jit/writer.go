package jit

/*
import (
	"errors"
	"io"
)

var ErrWriteClosed = errors.New("write on closed writer")

var _ io.WriteCloser = (*Writer)(nil)

type Writer struct {
	pg   []byte
	pos  int
	done bool
}

func (w *Writer) Bytes() []byte {
	if !w.done {
		return nil
	}
	return w.pg[:w.pos]
}

func (w *Writer) Write(p []byte) (n int, err error) {
	if w.done {
		return 0, ErrWriteClosed
	}
	need := w.pos + len(p)
	if need > len(w.pg) {
		if w.pg, err = Realloc(w.pg, need); err != nil {
			return
		}
	}
	n = copy(w.pg[w.pos:], p)
	w.pos += n
	return
}

func (w *Writer) Close() error {
	if !w.done {
		if err := Exec(w.pg); err != nil {
			return err
		}
		w.done = true
	}
	return nil
}
*/
