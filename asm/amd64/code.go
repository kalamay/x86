package amd64

import "fmt"

type RawCode [3]uint8

type Code struct {
	n   uint8
	raw [3]byte
}

func C(b ...byte) (c Code) {
	for _, b := range b {
		c.raw[c.n] = b
		c.n++
	}
	return
}

func (c Code) Len() int {
	return int(c.n)
}

func (c Code) Encode(dst []byte) int {
	dst[c.n-1] = 0
	return copy(dst, c.raw[:c.n])
}

func (c Code) String() string {
	switch c.n {
	case 1:
		return fmt.Sprintf("#%02x", c.raw[0])
	case 2:
		return fmt.Sprintf("#%02x%02x", c.raw[0], c.raw[1])
	case 3:
		return fmt.Sprintf("#%02x%02x%02x", c.raw[0], c.raw[1], c.raw[2])
	}
	return ""
}
