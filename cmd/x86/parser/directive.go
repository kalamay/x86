package parser

type Directive interface {
	Name() string
	Before(p *Parser) error
	After(p *Parser) error
	Parse(p *Parser) error
}

func NewDirective(name string, f ...func(p *Parser) error) Directive {
	fn := parseNone
	if len(f) > 0 {
		fn = f[0]
	}
	return &directive{name: name, parse: fn}
}

type directive struct {
	name  string
	parse func(p *Parser) error
}

func (d *directive) Name() string           { return d.name }
func (_ *directive) Before(p *Parser) error { return nil }
func (_ *directive) After(p *Parser) error  { return nil }

func (d *directive) Parse(p *Parser) error {
	return d.parse(p)
}

func parseNone(p *Parser) error {
	return nil
}
