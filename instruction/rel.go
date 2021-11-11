package instruction

import (
	"math"
	"strconv"

	"github.com/kalamay/x86/operand"
)

type (
	relFwd int64
	relRwd int64
)

const (
	minRel8  = math.MinInt8 + 2
	maxRel8  = math.MaxInt8
	minRel32 = math.MinInt32 + 6
	maxRel32 = math.MaxInt32
)

func (_ relFwd) Kind() operand.Kind { return operand.KindMem }
func (_ relFwd) Validate() error    { return nil }
func (r relFwd) String() string     { return strconv.FormatInt(int64(r), 10) }

func (r relFwd) Matches(p operand.Param) bool {
	if p.Kind() == operand.KindRel {
		switch operand.RelParam(p).Size() {
		case operand.Size8:
			return 0 <= r && r <= maxRel8
		case operand.Size32:
			return 0 <= r && r <= maxRel32
		}
	}
	return false
}

func (_ relRwd) Kind() operand.Kind { return operand.KindMem }
func (_ relRwd) Validate() error    { return nil }
func (r relRwd) String() string     { return strconv.FormatInt(int64(r), 10) }

func (r relRwd) Matches(p operand.Param) bool {
	if p.Kind() == operand.KindRel {
		switch operand.RelParam(p).Size() {
		case operand.Size8:
			return minRel8 <= r && r <= 0
		case operand.Size32:
			return minRel32 <= r && r <= 0
		}
	}
	return false
}

func ResolveRel(from, to int, enc []Format) (operand.Arg, bool) {
	if from >= to {
		rel := relRwd(0)
		for from--; from >= to; from-- {
			l := enc[from].Len
			if l == 0 {
				return nil, false
			}
			rel -= relRwd(l)
		}
		return rel, true
	}

	rel := relFwd(0)
	for from++; from < to; from++ {
		l := enc[from].Len
		if l == 0 {
			return nil, false
		}
		rel += relFwd(l)
	}
	return rel, true
}
