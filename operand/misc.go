package operand

import "fmt"

type (
	Misc      uint16
	MiscParam = Misc
)

const (
	SAE = iota + 1
	ER
)

func (m Misc) Matches(p Param) bool {
	return p.Kind() == KindMisc && MiscParam(p) == m
}

func (m Misc) String() string {
	return fmt.Sprintf("operand.Misc(%d)", m)
}
