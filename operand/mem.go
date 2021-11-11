package operand

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var (
	ErrMemBase          = errors.New("base register not provided")
	ErrNoIndexScale     = errors.New("scale provided without index")
	ErrInvalidScale     = errors.New("unsupported scale for index")
	ErrUnsupportedIndex = errors.New("unsupported index")
)

type (
	Mem struct {
		Disp    int32
		Index   Reg
		Base    Reg
		Segment Reg
		Size    Size
		Scale   Size
		Type    MemType
	}
	MemParam uint16
	MemType  uint16
)

const (
	MemTypeGeneral = iota << mTypeShift
	MemTypeOffset
	MemTypeFloat80
	MemTypeBroadcast
	MemTypeVector32
	MemTypeVector64

	mElemShift   = sizeBits
	mElemMask    = 0b111 << mElemShift
	mTargetShift = mElemShift + sizeBits
	mTypeShift   = mTargetShift + sizeBits
	mTypeBits    = 3
	mTypeMask    = 0b111 << mTypeShift
)

func (m Mem) Kind() Kind { return KindMem }

func (m Mem) Matches(p Param) bool {
	if p.Kind() != KindMem {
		return false
	}
	mp := MemParam(p)

	// TODO verify all the memory fields
	if mp.Type() == MemTypeOffset && (m.Segment == 0 || m.Segment.Type() != RegTypeSegment) {
		return false
	}

	s, ms := mp.Size(), m.Size
	if ms == Size0 {
		return true
	}
	if s > Size0 {
		return s == ms
	}
	return Size16 <= ms && ms <= Size64
}

func (m Mem) Validate() error {
	if m.Base.Validate() != nil {
		return ErrMemBase
	}

	bs, is := m.Base.Size(), m.Index.Size()
	if bs != Size64 && bs != Size32 {
		return fmt.Errorf("invalid %d-bit base register", bs.Bits())
	}
	if is != Size0 && is != bs {
		return fmt.Errorf("base register is %d-bit, but index is %d-bit", bs.Bits(), is.Bits())
	}

	if m.Index == 0 {
		if m.Scale != Size0 {
			return ErrNoIndexScale
		}
	} else {
		if m.Scale < Size8 || Size64 < m.Scale {
			return ErrInvalidScale
		}
		if m.Scale == Size8 && (m.Index.ID()&0b111) == 0b100 {
			return ErrUnsupportedIndex
		}
	}

	return nil
}

func (m Mem) String() string {
	n, parts := 0, [11]string{}

	if m.Size > Size0 {
		parts[n] = memNames[m.Size]
		parts[n+1] = " "
		n += 2
	}

	parts[n] = "["
	n += 1

	parts[n] = m.Base.String()
	n += 1

	if m.Index > 0 {
		parts[n] = " + "
		parts[n+1] = m.Index.String()
		if m.Scale > 1 {
			parts[n+2] = "*"
			parts[n+3] = m.Scale.ByteString()
		}
		n += 4
	}

	if m.Disp != 0 {
		var d uint64
		if m.Disp > 0 {
			parts[n] = " + "
			d = uint64(m.Disp)
		} else {
			parts[n] = " - "
			d = uint64(-int64(m.Disp))
		}
		parts[n+1] = strconv.FormatUint(d, 10)
		n += 2
	}

	parts[n] = "]"
	n += 1

	return strings.Join(parts[:n], "")
}

func (m MemParam) Size() Size       { return Size(m & sizeMask) }
func (m MemParam) ElemSize() Size   { return Size((m >> mElemShift) & sizeMask) }
func (m MemParam) TargetSize() Size { return Size((m >> mTargetShift) & sizeMask) }
func (m MemParam) Type() MemType    { return MemType(m & mTypeMask) }

func (m *MemParam) setElemSize(s Size) {
	*m = (*m & ^MemParam(mElemMask)) | (MemParam(s) << mElemShift)
}

var memScale = [...]string{"0", "1", "2", "x", "4", "x", "x", "x", "8"}
var memNames = [...]string{"", "BYTE PTR", "WORD PTR", "DWORD PTR", "QWORD PTR", "XMMWORD PTR", "YMMWORD PTR", "ZMMWORD PTR"}
