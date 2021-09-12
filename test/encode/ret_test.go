package encode

import (
	"testing"

	"github.com/kalamay/x86/asm/amd64/inst"
	"github.com/kalamay/x86/test"
)

func TestRET(t *testing.T) {
	test.Compare(t, inst.RET, expectRET[:])
}

var expectRET = [...]test.Expect{
	{0, 0, []byte{0xc3}, ""},
	{16, 0, []byte{0xc2, 0x0, 0x80}, ""},
	{16, 1, []byte{0xc2, 0x0, 0xc0}, ""},
	{16, 2, []byte{0xc2, 0xff, 0x3f}, ""},
	{16, 3, []byte{0xc2, 0xff, 0x7f}, ""},
	{16, 4, []byte{0xc2, 0x0, 0x0}, ""},
	{16, 5, []byte{0xc2, 0xff, 0x7f}, ""},
	{16, 6, []byte{0xc2, 0xff, 0xff}, ""},
}
