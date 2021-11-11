package instruction

import (
	"fmt"
	"strconv"
)

type HexByte byte

func (hb *HexByte) UnmarshalText(text []byte) error {
	n, err := strconv.ParseUint(string(text), 16, 8)
	if err == nil {
		*hb = HexByte(n)
	}
	return err
}

type (
	OptBits    uint8
	OptByte    uint8
	OptRef     uint8
	OptRefBits uint8
	OptRefByte uint8
	OptMode    uint8
)

const (
	OptModeNone = iota << oModeShift
	OptModeValue
	OptModeRef

	oModeShift = 6
	oModeMask  = 0b11 << oModeShift
	oValueMask = ^uint8(oModeMask)
)

func isOptRef(v []byte) bool {
	return len(v) == 0 || v[0] == '#'
}

func parseRef(v []byte) (OptRef, error) {
	if len(v) == 0 {
		return 0, nil
	}
	if len(v) == 2 && v[0] == '#' && '0' <= v[1] && v[1] <= '9' {
		return OptRef(OptModeRef | v[1] - '0'), nil
	}
	return 0, fmt.Errorf("invalid reference format %q", v)
}

func (o *OptBits) UnmarshalText(text []byte) error {
	n, err := strconv.ParseUint(string(text), 2, oModeShift)
	if err == nil {
		*o = OptBits(OptModeValue | n)
	}
	return err
}

func (o *OptByte) UnmarshalText(text []byte) error {
	n, err := strconv.ParseUint(string(text), 10, oModeShift)
	if err == nil {
		*o = OptByte(OptModeValue | n)
	}
	return err
}

func (o *OptRef) UnmarshalText(text []byte) error {
	r, err := parseRef(text)
	if err == nil {
		*o = r
	}
	return err
}

func (o *OptRefBits) UnmarshalText(text []byte) error {
	if isOptRef(text) {
		return (*OptRef)(o).UnmarshalText(text)
	}
	return (*OptBits)(o).UnmarshalText(text)
}

func (o *OptRefByte) UnmarshalText(text []byte) error {
	if isOptRef(text) {
		return (*OptRef)(o).UnmarshalText(text)
	}
	return (*OptByte)(o).UnmarshalText(text)
}

func (o OptBits) Mode() OptMode    { return OptMode(o & oModeMask) }
func (o OptByte) Mode() OptMode    { return OptMode(o & oModeMask) }
func (o OptRef) Mode() OptMode     { return OptMode(o & oModeMask) }
func (o OptRefBits) Mode() OptMode { return OptMode(o & oModeMask) }
func (o OptRefByte) Mode() OptMode { return OptMode(o & oModeMask) }

func (o OptBits) IsSet() bool    { return o.Mode() != 0 }
func (o OptByte) IsSet() bool    { return o.Mode() != 0 }
func (o OptRef) IsSet() bool     { return o.Mode() != 0 }
func (o OptRefBits) IsSet() bool { return o.Mode() != 0 }
func (o OptRefByte) IsSet() bool { return o.Mode() != 0 }

func (o OptBits) Value() (uint8, OptMode)    { return (uint8(o) & oValueMask), o.Mode() }
func (o OptByte) Value() (uint8, OptMode)    { return (uint8(o) & oValueMask), o.Mode() }
func (o OptRef) Value() (uint8, OptMode)     { return (uint8(o) & oValueMask), o.Mode() }
func (o OptRefBits) Value() (uint8, OptMode) { return (uint8(o) & oValueMask), o.Mode() }
func (o OptRefByte) Value() (uint8, OptMode) { return (uint8(o) & oValueMask), o.Mode() }
