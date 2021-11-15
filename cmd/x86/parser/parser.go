package parser

import (
	"errors"
	"fmt"
	"io"
	"math"
	"sort"
	"strconv"
	"strings"
	"text/scanner"

	"github.com/kalamay/x86/instruction"
	"github.com/kalamay/x86/operand"
	"github.com/kalamay/x86/x64"
)

var (
	ErrDirectiveUnknown       = errors.New("unknown directive")
	ErrDisplacementInvalid    = errors.New("invalid displacement size")
	ErrIdentifierExpected     = errors.New("identifier expected")
	ErrInputUnexpected        = errors.New("unexpected input")
	ErrIntegerExpected        = errors.New("integer expected")
	ErrIntegerOverflow        = errors.New("integer overflow")
	ErrMnemonicUnknown        = errors.New("unknown mnemonic")
	ErrOperandExpected        = errors.New("operand expected")
	ErrPTRExpected            = errors.New(`"PTR" expected`)
	ErrRegisterExpected       = errors.New("register expected")
	ErrRegisterTypeUnexpected = errors.New("unexpected register type")
	ErrRegisterUnavailable    = errors.New("register is unavailable")
	ErrSIBInvalid             = errors.New("invalid base/index expression")
	ErrScaleInvalid           = errors.New("invalid scale")
)

type Error struct {
	scanner.Position
	Err error
}

func NewError(err error, at scanner.Position) *Error {
	return &Error{Position: at, Err: err}
}

func (e *Error) Unwrap() error {
	return e.Err
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %v", e.Position, e.Err)
}

type RegisterSet map[operand.Reg]struct{}

func (rs RegisterSet) Add(r operand.Reg) {
	rs[r] = struct{}{}
}

func (rs RegisterSet) Contains(r operand.Reg) bool {
	_, ok := rs[r]
	return ok
}

func (rs RegisterSet) Copy() RegisterSet {
	c := make(RegisterSet, len(rs))
	for reg := range rs {
		c[reg] = struct{}{}
	}
	return c
}

func (rs RegisterSet) Values() []operand.Reg {
	regs := make([]operand.Reg, 0, len(rs))
	for reg := range rs {
		regs = append(regs, reg)
	}
	sort.Slice(regs, func(i, j int) bool {
		return regs[i] < regs[j]
	})
	return regs
}

type Parser struct {
	Registers RegisterSet
	Reserved  RegisterSet

	dirs     []Directive
	dirnames map[string]int

	scan   scanner.Scanner
	val    interface{}
	err    error
	peeked bool
}

func (p *Parser) NewError(err error) *Error {
	return NewError(err, p.scan.Position)
}

func (p *Parser) SetDirectives(directives ...Directive) {
	p.dirs = make([]Directive, 0, len(directives))
	p.dirnames = make(map[string]int, len(directives))
	for _, dir := range directives {
		if _, ok := p.dirnames[dir.Name()]; !ok {
			p.dirnames[dir.Name()] = len(p.dirs)
			p.dirs = append(p.dirs, dir)
		}
	}
}

func (p *Parser) Init(name string, src io.Reader) {
	p.Registers = RegisterSet{}
	p.Reserved = RegisterSet{}
	p.scan.Mode = scanner.ScanIdents | scanner.ScanInts | scanner.ScanComments | scanner.SkipComments
	p.scan.Filename = name
	p.scan.Init(src)

	if p.scan.Peek() == '#' {
		for {
			ch := p.scan.Next()
			if ch < 0 || ch == '\n' {
				break
			}
		}
	}
}

func (p *Parser) Eval(data *instruction.Set, e *x64.Emit) error {
	for _, d := range p.dirs {
		if err := d.Before(p); err != nil {
			return err
		}
	}

	for {
		pos := p.scan.Position
		val, err := p.Next()
		if err != nil {
			return err
		}
		if val == nil {
			break
		}

		switch val := val.(type) {
		case string:
			switch {
			case val == ".":
				if err := p.directive(); err != nil {
					return err
				}
			case p.Maybe(rune(':')):
				e.Label(val)
			default:
				inst := data.Lookup(val)
				if inst == nil {
					return NewError(ErrMnemonicUnknown, pos)
				}
				args, err := p.Args(inst)
				if err != nil {
					return err
				}
				e.EmitCall(&x64.EmitCall{
					Instruction: inst,
					Args:        args,
					EmitPosition: x64.EmitPosition{
						Filename: pos.Filename,
						Line:     pos.Line,
						Column:   pos.Column,
					},
				})
			}
		default:
			return p.NewError(ErrIdentifierExpected)
		}
	}

	for i := len(p.dirs) - 1; i >= 0; i-- {
		if err := p.dirs[i].After(p); err != nil {
			return err
		}
	}

	return nil
}

func (p *Parser) Next() (interface{}, error) {
	if p.peeked {
		p.peeked = false
	} else {
		p.advance()
	}
	return p.val, p.err
}

func (p *Parser) Peek() (interface{}, error) {
	if !p.peeked {
		p.advance()
		p.peeked = true
	}
	return p.val, p.err
}

func (p *Parser) Expect(exp interface{}) error {
	val, err := p.Next()
	if err != nil {
		return err
	}
	if val != exp {
		if r, ok := exp.(rune); ok {
			return p.NewError(fmt.Errorf("'%c' expected", r))
		}
		return p.NewError(fmt.Errorf("%v expected", exp))
	}
	return nil
}

func (p *Parser) Maybe(exp interface{}) bool {
	val, _ := p.Peek()
	if val != exp {
		return false
	}
	p.Next()
	return true
}

func (p *Parser) advance() {
	if p.err != nil {
		return
	}

	tok, txt := p.scan.Scan(), p.scan.TokenText()
	p.peeked = false

	switch tok {

	case scanner.EOF:
		p.val = nil

	case scanner.Int:
		var n uint64
		if n, p.err = strconv.ParseUint(txt, 0, 64); p.err != nil {
			p.val = nil
		} else {
			p.val = n
		}

	case scanner.Ident:
		if sz, ok := parseSize(txt); ok {
			if p.scan.Scan() != scanner.Ident || !strings.EqualFold(p.scan.TokenText(), "PTR") {
				p.err = p.NewError(ErrPTRExpected)
			}
			p.val = sz
		} else {
			p.val = txt
		}

	case ',', '[', ']', '*', '+', '-', ':':
		p.val = tok

	default:
		p.val = txt
	}
}

func (p *Parser) directive() error {
	name, err := p.Ident()
	if err != nil {
		return err
	}
	if d, ok := p.dirnames[name]; ok {
		return p.dirs[d].Parse(p)
	}
	return p.NewError(ErrDirectiveUnknown)
}

func (p *Parser) Args(inst *instruction.Instruction) ([]operand.Arg, error) {
	args := []operand.Arg{}

	if expectRel(inst) {
		id, err := p.Ident()
		if err != nil {
			return nil, err
		}
		args = append(args, operand.Label(id))
	} else {
		arg, err := p.Arg(false)

		if err != nil || arg == nil {
			return nil, err
		}

		args = append(args, arg)
		for p.Maybe(rune(',')) {
			if arg, err = p.Arg(true); err != nil {
				return nil, err
			}
			args = append(args, arg)
		}
	}

	return args, nil
}

func (p *Parser) getReg(name string) (r operand.Reg, ok bool, err error) {
	if r, ok = operand.RegOf(name); ok {
		if p.Reserved.Contains(r) {
			r, ok, err = 0, false, p.NewError(ErrRegisterUnavailable)
		} else {
			p.Registers.Add(r)
		}
	}
	return
}

func (p *Parser) Arg(expect bool) (arg operand.Arg, err error) {
	var val interface{}
	if val, err = p.Peek(); err != nil {
		return
	}

	switch v := val.(type) {
	case uint64:
		p.Next()
		return operand.Uint(v), nil
	case string:
		if reg, ok, rerr := p.getReg(v); rerr != nil {
			err = rerr
			return
		} else if ok {
			p.Next()
			return reg, nil
		}
	case rune:
		switch v {
		case '-':
			p.Next()
			var n int64
			if n, err = p.neg(); err != nil {
				return
			}
			return operand.Int(n), nil
		case '[':
			return p.Mem(operand.Size0)
		}
	case operand.Size:
		p.Next()
		return p.Mem(v)
	}

	if expect {
		err = p.NewError(ErrOperandExpected)
	}
	return
}

func (p *Parser) neg() (int64, error) {
	val, err := p.Next()
	if err != nil {
		return 0, err
	}
	if n, ok := val.(uint64); ok {
		if n > math.MaxInt64+1 {
			return 0, p.NewError(ErrIntegerOverflow)
		}
		return -int64(n), nil
	} else {
		return 0, p.NewError(ErrIntegerExpected)
	}
}

func (p *Parser) Ident() (string, error) {
	val, err := p.Next()
	if err != nil {
		return "", err
	}

	if str, ok := val.(string); ok {
		return str, nil
	}

	return "", p.NewError(ErrIdentifierExpected)
}

func (p *Parser) Scale() (operand.Size, error) {
	val, err := p.Next()
	if err != nil {
		return 0, err
	}

	if n, ok := val.(uint64); ok {
		switch n {
		case 1:
			return operand.Size8, nil
		case 2:
			return operand.Size16, nil
		case 4:
			return operand.Size32, nil
		case 8:
			return operand.Size64, nil
		}
	}

	return 0, p.NewError(ErrScaleInvalid)
}

func (p *Parser) Reg() (reg operand.Reg, err error) {
	name := ""
	if name, err = p.Ident(); err == nil {
		ok := false
		if reg, ok, err = p.getReg(name); err == nil && !ok {
			err = p.NewError(ErrRegisterExpected)
		}
	}
	return
}

func (p *Parser) Mem(sz operand.Size) (mem operand.Mem, err error) {
	if err = p.Expect(rune('[')); err != nil {
		return
	}

	mem.Size = sz
	if mem.Base, err = p.Reg(); err != nil {
		return
	}

	switch mem.Base.Type() {
	case operand.RegTypeGeneral:
	case operand.RegTypeSegment:
		mem.Segment = mem.Base
		if err = p.Expect(rune(':')); err != nil {
			return
		}
		if mem.Base, err = p.Reg(); err != nil {
			return
		}
	default:
		err = p.NewError(ErrRegisterTypeUnexpected)
		return
	}

	for {
		if p.Maybe(rune('+')) {
			if err = p.memadd(&mem); err != nil {
				return
			}
		} else if p.Maybe(rune('-')) {
			if err = p.memsub(&mem); err != nil {
				return
			}
		} else {
			break
		}
	}

	err = p.Expect(rune(']'))
	return
}

func (p *Parser) memadd(mem *operand.Mem) (err error) {
	var op interface{}
	if op, err = p.Peek(); err != nil {
		return
	}

	switch op := op.(type) {
	case string:
		if mem.Index != 0 {
			err = p.NewError(ErrSIBInvalid)
			return
		}
		if mem.Index, err = p.Reg(); err == nil {
			if p.Maybe(rune('*')) {
				mem.Scale, err = p.Scale()
			} else {
				mem.Scale = operand.Size8
			}
		}
	case uint64:
		if val, ok := addUint(mem.Disp, op); ok {
			mem.Disp = val
			p.Next()
		} else {
			err = p.NewError(ErrDisplacementInvalid)
		}
	case rune:
		if op == '-' {
			p.Next()
			err = p.memsub(mem)
		} else {
			err = p.NewError(ErrInputUnexpected)
		}
	default:
		err = p.NewError(ErrInputUnexpected)
	}
	return
}

func (p *Parser) memsub(mem *operand.Mem) (err error) {
	var n int64
	if n, err = p.neg(); err == nil {
		if val, ok := addInt(mem.Disp, n); ok {
			mem.Disp = val
		} else {
			err = p.NewError(ErrDisplacementInvalid)
		}
	}
	return
}

func expectRel(inst *instruction.Instruction) bool {
	op := inst.Forms[0].Operands
	return op.Len == 1 && op.Val[0].Kind() == operand.KindRel
}
func addInt(d int32, val int64) (int32, bool) {
	if val < math.MinInt32 || val > math.MaxInt32 {
		return d, false
	}
	n := int64(d) + val
	if n < math.MinInt32 || n > math.MaxInt32 {
		return d, false
	}
	return int32(n), true
}

func addUint(d int32, val uint64) (int32, bool) {
	if val > math.MaxInt32 {
		return d, false
	}
	return addInt(d, int64(val))
}

func parseSize(t string) (operand.Size, bool) {
	for i, name := range sizeNames {
		if strings.EqualFold(t, name) {
			return operand.Size(i + 1), true
		}
	}
	return 0, false

}

var sizeNames = [...]string{"BYTE", "WORD", "DWORD", "QWORD", "XMMWORD", "YMMWORD", "ZMMWORD"}
