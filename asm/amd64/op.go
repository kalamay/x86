package amd64

import "fmt"

type Op interface {
	fmt.Stringer
	Name() string
	Kind() Kind
	Size() Size
	Match(t Type, dst Size) bool
	Validate() error
}

var (
	_ Op = Reg(0)
	_ Op = Mem{}
	_ Op = Int(0)
	_ Op = Uint(0)
)
