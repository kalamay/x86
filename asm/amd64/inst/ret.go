package inst

import . "github.com/kalamay/x86/asm/amd64"

var RET = NewInst("RET", S16, []Inst{
	{0, 0, 0, C(0xC3)},
	{T(TI16), 0, 0, C(0xC2)},
})
