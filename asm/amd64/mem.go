package amd64

import (
	"errors"
	"strconv"
	"strings"
)

var ErrMemBase = errors.New("memory base register invalid")

type Mem struct {
	size  Size
	scale Size
	index Reg
	base  Reg
	disp  uint32
}

func MakeMem(base Reg) Mem {
	if base.Validate() != nil {
		panic("base must be set")
	}
	return Mem{base: base}
}

func (m Mem) WithSize(size Size) Mem {
	m.size = size
	return m
}

func (m Mem) WithIndex(index Reg, scale Size) Mem {
	if index.Size() == S0 {
		if scale != S0 {
			panic("scale cannot be set without index")
		}
	} else if scale == S0 || scale > S64 {
		panic("scale must be 1, 2, 4, or 8")
	}
	m.scale = scale
	m.index = index
	return m
}

func (m Mem) WithDisplacement(disp uint32) Mem {
	m.disp = disp
	return m
}

func (_ Mem) Kind() Kind {
	return KindMem
}

func (m Mem) Size() Size {
	return m.size
}

func (m Mem) Match(t Type) bool {
	// TODO: check other memory flags to support 32-bit
	if !t.IsMem() {
		return false
	}
	s, ms := t.MemSize(), m.size
	if ms == S0 {
		return true
	}
	if s > S0 {
		return s == ms
	}
	return S16 <= ms && ms <= S64
}

func (m Mem) Validate() error {
	if m.base.Validate() != nil {
		return ErrMemBase
	}
	return nil
}

func (m Mem) Name() string {
	return memNames[m.size]
}

func (m Mem) String() string {
	n, parts := 0, [11]string{}

	if m.size > S0 {
		parts[n] = m.Name()
		parts[n+1] = " "
		n += 2
	}

	parts[n] = "["
	n += 1

	if m.disp > 0 {
		parts[n] = strconv.FormatUint(uint64(m.disp), 10)
		parts[n+1] = " + "
		n += 2
	}

	parts[n] = m.base.String()
	n += 1

	if m.index > 0 {
		parts[n] = " + "
		parts[n+1] = m.index.Name()
		if m.scale > 1 {
			parts[n+2] = "*"
			parts[n+3] = m.scale.ByteString()
		}
		n += 4
	}

	parts[n] = "]"
	n += 1

	return strings.Join(parts[:n], "")
}

var memScale = [...]string{"0", "1", "2", "x", "4", "x", "x", "x", "8"}
var memNames = [...]string{"", "BYTE PTR", "WORD PTR", "DWORD PTR", "QWORD PTR", "m128", "m256"}
