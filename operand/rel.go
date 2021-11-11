package operand

type (
	RelParam uint16
	Label    string
)

func (r RelParam) Size() Size {
	return Size(r & sizeMask)
}

func (_ Label) Kind() Kind           { return KindMem }
func (_ Label) Validate() error      { return nil }
func (l Label) String() string       { return string(l) }
func (_ Label) Matches(p Param) bool { return false }
