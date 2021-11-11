package x64

import (
	"fmt"
	"io"
	"path"
	"runtime"
	"strings"

	"github.com/kalamay/x86/instruction"
	"github.com/kalamay/x86/operand"
)

type InstructionID uint16

func (i InstructionID) String() string {
	return Instructions.Instructions[i].Name
}

type Emitter interface {
	Open()
	Emit(e *Emit, call *EmitCall)
	Label(e *Emit, label *EmitLabel)
	Close(e *Emit)
}

type EmitValue interface {
	Name() string
	Position() EmitPosition
}

type EmitPosition struct {
	Filename string
	Line     int
	Column   int
}

func (pos *EmitPosition) SetFrame(f runtime.Frame) {
	if f.Line > 0 {
		pos.Filename = path.Base(f.File)
		pos.Line = f.Line
	} else {
		pos.Filename = ""
		pos.Line = -1
	}
	pos.Column = 0
}

func (pos *EmitPosition) IsValid() bool {
	return pos.Line > 0
}

func (pos *EmitPosition) String() string {
	if !pos.IsValid() {
		return ""
	}
	if pos.Column <= 0 {
		return fmt.Sprintf("%s:%d", pos.Filename, pos.Line)
	}
	return fmt.Sprintf("%s:%d:%d", pos.Filename, pos.Line, pos.Column)
}

type Emit struct {
	emitter Emitter
	errors  []error
	w       io.Writer
}

func (e *Emit) Write(p []byte) (int, error) {
	if len(e.errors) > 0 {
		return len(p), nil
	}
	n, err := e.w.Write(p)
	if err != nil {
		e.AddError(err, nil)
	}
	return n, err
}

func (e *Emit) AddError(err error, val EmitValue) {
	if val != nil {
		err = &Error{Value: val, Err: err}
	}
	e.errors = append(e.errors, err)
}

func (e *Emit) Open(em Emitter, w io.Writer) {
	em.Open()
	e.emitter = em
	e.errors = nil
	e.w = w
}

func (e *Emit) Emit(id InstructionID, args []operand.Arg) {
	call := &EmitCall{
		Instruction: &Instructions.Instructions[id],
		Args:        args,
	}
	runtime.Callers(2, call.pc[:])
	e.emitter.Emit(e, call)
}

func (e *Emit) EmitCall(call *EmitCall) {
	if call.Line == 0 {
		runtime.Callers(2, call.pc[:])
	}
	e.emitter.Emit(e, call)
}

func (e *Emit) Lock()     { e.Write([]byte{0xf0}) }
func (e *Emit) Likely()   { e.Write([]byte{0x3e}) }
func (e *Emit) Unlikely() { e.Write([]byte{0x2e}) }

func (e *Emit) Label(name string) {
	label := &EmitLabel{
		Value: name,
	}
	runtime.Callers(2, label.pc[:])
	e.emitter.Label(e, label)
}

func (e *Emit) Close() []error {
	e.emitter.Close(e)
	errs := e.errors
	e.emitter = nil
	e.errors = nil
	e.w = nil
	return errs
}

type EmitCall struct {
	Instruction *instruction.Instruction
	Args        []operand.Arg
	EmitPosition

	pc [2]uintptr
}

func (c *EmitCall) Name() string {
	return c.Instruction.Name
}

func (c *EmitCall) Position() EmitPosition {
	if c.Line == 0 {
		frames := runtime.CallersFrames(c.pc[:])
		frame, more := frames.Next()
		// TODO: check package too?
		if more && strings.HasSuffix(frame.Function, "."+c.Instruction.Name) {
			frame, more = frames.Next()
		}
		c.SetFrame(frame)
	}
	return c.EmitPosition
}

type EmitLabel struct {
	Value string
	EmitPosition

	pc [1]uintptr
}

func (l *EmitLabel) Name() string {
	return l.Value
}

func (l *EmitLabel) Position() EmitPosition {
	if l.Line == 0 {
		frames := runtime.CallersFrames(l.pc[:])
		frame, _ := frames.Next()
		l.SetFrame(frame)
	}
	return l.EmitPosition
}

type Error struct {
	Value EmitValue
	Err   error
}

func (e *Error) Unwrap() error {
	return e.Err
}

func (e *Error) Error() string {
	name, pos := e.Value.Name(), e.Value.Position()
	if !pos.IsValid() {
		return fmt.Sprintf("%q failed: %v", name, e.Err)
	}
	return fmt.Sprintf("%s: %q failed: %v", &pos, name, e.Err)
}
