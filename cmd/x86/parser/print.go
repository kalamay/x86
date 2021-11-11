package parser

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"syscall"
	"unsafe"

	"github.com/kalamay/x86/operand"
	"github.com/kalamay/x86/x64"
	"github.com/kalamay/x86/x64/jit"
)

const printName = "print"

var PrintNop = NewDirective(printName, func(p *Parser) error {
	_, _, err := parsePrint(p)
	return err
})

var ErrPrintFmtUnknown = errors.New("unknown print format")

type Print struct {
	emit *x64.Emit
	ptr  operand.Reg
	buf  []byte
}

func NewPrint(emit *x64.Emit, ptr operand.Reg) *Print {
	return &Print{emit: emit, ptr: ptr}
}

func (_ *Print) Name() string { return printName }

func (pr *Print) Before(p *Parser) (err error) {
	const (
		pg    = 4096
		full  = pg * 16
		size  = full - pg
		rw    = syscall.PROT_READ | syscall.PROT_WRITE
		flags = syscall.MAP_ANON | syscall.MAP_PRIVATE
	)

	if pr.buf == nil {
		if pr.buf, err = syscall.Mmap(-1, 0, full, rw, flags); err == nil {
			syscall.Mprotect(pr.buf[size:], syscall.PROT_NONE)
			p.Reserved.Add(pr.ptr)
			pr.emit.MOV(pr.ptr, operand.Uint(uintptr(unsafe.Pointer(&pr.buf[0]))))
		}
	} else {
		jit.Memset(pr.buf[:size], 0)
	}

	return
}

func (pr *Print) After(p *Parser) error { return nil }

func (pr *Print) Parse(p *Parser) error {
	reg, pf, err := parsePrint(p)
	if err != nil {
		return err
	}

	mem := operand.Mem{Base: pr.ptr, Size: operand.Size32}
	pr.emit.MOV(mem, operand.Int((int(pf)<<16)|int(reg)))

	mem.Size = reg.Size()
	mem.Disp = 4

	switch reg.Type() {
	case operand.RegTypeGeneral:
		pr.emit.MOV(mem, reg)
	case operand.RegTypeVector:
		pr.emit.VMOVDQU(mem, reg)
	}
	pr.emit.LEA(pr.ptr, operand.Mem{Base: pr.ptr, Disp: int32(4 + reg.Size().Bytes())})

	return nil
}

func (pr *Print) WriteTo(w io.Writer) (n int64, err error) {
	b := pr.buf
	if b == nil {
		return
	}

	for {
		next := binary.LittleEndian.Uint32(b)
		if next == 0 {
			break
		}

		pf, reg := printFmt(next>>16), operand.Reg(next)
		val := b[4 : 4+reg.Size().Bytes()]
		b = b[4+reg.Size().Bytes():]

		rn := 0
		rn, err = printReg(w, pf, reg, val)
		n += int64(rn)

		if err != nil {
			break
		}
	}
	return
}

func parsePrint(p *Parser) (reg operand.Reg, pf printFmt, err error) {
	if reg, err = p.Reg(); err != nil {
		return
	}

	if p.Maybe(rune(',')) {
		name, ok := "", false
		if name, err = p.Ident(); err == nil {
			if pf, ok = printFmts[name]; !ok {
				err = p.NewError(ErrPrintFmtUnknown)
			}
		}
	}

	return
}

func printReg(w io.Writer, pf printFmt, r operand.Reg, value []byte) (int, error) {
	switch pf {
	case pfBits:
		return printBits(w, r, value)
	case pfInt8:
		return print8(w, r, value, pmInt)
	case pfUint8:
		return print8(w, r, value, pmUint)
	case pfInt16:
		return print16(w, r, value, pmInt)
	case pfUint16:
		return print16(w, r, value, pmUint)
	case pfInt32:
		return print32(w, r, value, pmInt)
	case pfUint32:
		return print32(w, r, value, pmUint)
	case pfInt64:
		return print64(w, r, value, pmInt)
	case pfUint64:
		return print64(w, r, value, pmUint)
	case pfFloat32:
		return print32(w, r, value, pmFloat)
	case pfFloat64:
		return print64(w, r, value, pmFloat)
	}
	return printHex(w, r, value)
}

func printHex(w io.Writer, r operand.Reg, value []byte) (int, error) {
	const table = "0123456789abcdef"

	buf, n := [64]byte{}, len(value)*2
	i := n - 1
	for _, v := range value {
		buf[i-1] = table[v>>4]
		buf[i] = table[v&0x0f]
		i -= 2
	}

	return fmt.Fprintf(w, "%6s = %s\n", r, buf[:n])
}

func printBits(w io.Writer, r operand.Reg, value []byte) (n int, err error) {
	var rn int
	n, err = fmt.Fprintf(w, "%6s =", r)
	for i := len(value) - 1; err == nil && i >= 0; i-- {
		rn, err = fmt.Fprintf(w, " %08b", value[i])
		n += rn
	}
	rn, err = fmt.Fprintln(w)
	n += rn
	return
}

func print8(w io.Writer, r operand.Reg, value []byte, mode byte) (n int, err error) {
	var rn int
	n, err = fmt.Fprintf(w, "%6s =", r)
	for i := len(value) - 1; err == nil && i >= 0; i-- {
		switch step := value[i]; mode {
		case pmInt:
			rn, err = fmt.Fprintf(w, " %d", int8(step))
		case pmUint:
			rn, err = fmt.Fprintf(w, " %d", step)
		}
		n += rn
	}
	rn, err = fmt.Fprintln(w)
	n += rn
	return
}

func print16(w io.Writer, r operand.Reg, value []byte, mode byte) (n int, err error) {
	var rn int
	n, err = fmt.Fprintf(w, "%6s =", r)
	for i := len(value); err == nil && i >= 1; i -= 2 {
		switch step := binary.LittleEndian.Uint16(value[i-2 : i]); mode {
		case pmInt:
			rn, err = fmt.Fprintf(w, " %d", int16(step))
		case pmUint:
			rn, err = fmt.Fprintf(w, " %d", step)
		}
		n += rn
	}
	rn, err = fmt.Fprintln(w)
	n += rn
	return
}

func print32(w io.Writer, r operand.Reg, value []byte, mode byte) (n int, err error) {
	var rn int
	n, err = fmt.Fprintf(w, "%6s =", r)
	for i := len(value); err == nil && i >= 3; i -= 4 {
		switch step := binary.LittleEndian.Uint32(value[i-4 : i]); mode {
		case pmInt:
			rn, err = fmt.Fprintf(w, " %d", int32(step))
		case pmUint:
			rn, err = fmt.Fprintf(w, " %d", step)
		case pmFloat:
			rn, err = fmt.Fprintf(w, " %g", math.Float32frombits(step))
		}
		n += rn
	}
	rn, err = fmt.Fprintln(w)
	n += rn
	return
}

func print64(w io.Writer, r operand.Reg, value []byte, mode byte) (n int, err error) {
	var rn int
	n, err = fmt.Fprintf(w, "%6s =", r)
	for i := len(value); err == nil && i >= 7; i -= 8 {
		switch step := binary.LittleEndian.Uint64(value[i-8 : i]); mode {
		case pmInt:
			rn, err = fmt.Fprintf(w, " %d", int64(step))
		case pmUint:
			rn, err = fmt.Fprintf(w, " %d", step)
		case pmFloat:
			rn, err = fmt.Fprintf(w, " %g", math.Float64frombits(step))
		}
		n += rn
	}
	rn, err = fmt.Fprintln(w)
	n += rn
	return
}

type printFmt byte

const (
	pfHex printFmt = iota
	pfBits
	pfInt8
	pfUint8
	pfInt16
	pfUint16
	pfInt32
	pfUint32
	pfInt64
	pfUint64
	pfFloat32
	pfFloat64
)

const (
	pmInt byte = iota
	pmUint
	pmFloat
)

var printFmts = map[string]printFmt{
	"hex":     pfHex,
	"bits":    pfBits,
	"int8":    pfInt8,
	"uint8":   pfUint8,
	"int16":   pfInt16,
	"uint16":  pfUint16,
	"int32":   pfInt32,
	"uint32":  pfUint32,
	"int64":   pfInt64,
	"uint64":  pfUint64,
	"float32": pfFloat32,
	"float64": pfFloat64,
}
