package parser

import (
	"github.com/kalamay/x86/operand"
	"github.com/kalamay/x86/x64"
)

const breakName = "break"

var BreakNop = NewDirective(breakName)

type Break struct {
	emit *x64.Emit
	set  bool
}

func NewBreak(emit *x64.Emit) *Break {
	return &Break{emit: emit}
}

func (b *Break) IsSet() bool            { return b.set }
func (b *Break) Name() string           { return breakName }
func (b *Break) Before(p *Parser) error { return nil }
func (b *Break) After(p *Parser) error  { return nil }

func (b *Break) Parse(p *Parser) error {
	b.emit.INT(operand.Int(3))
	b.set = true
	return nil
}
