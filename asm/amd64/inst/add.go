package inst

import . "github.com/kalamay/x86/asm/amd64"

var ADD = NewInst("ADD", S32, []Inst{
	{A8_I8, 0, 0, C(0x04)},       // Add imm8 to AL.
	{M8_I8, 0, 0, C(0x80)},       // Add imm8 to r/m8.
	{M16_I8s, 0, 0, C(0x83)},     // Add sign-extended imm8 to r/m16.
	{M32_I8s, 0, 0, C(0x83)},     // Add sign-extended imm8 to r/m32.
	{M64_I8s, 0, RexW, C(0x83)},  // Add sign-extended imm8 to r/m64.
	{A16_I16, 0, 0, C(0x05)},     // Add imm16 to AX.
	{M16_I16, 0, 0, C(0x81)},     // Add imm16 to r/m16.
	{A32_I32, 0, 0, C(0x05)},     // Add imm32 to EAX.
	{A64_I32s, 0, RexW, C(0x05)}, // Add imm32 sign-extended to 64-bits to RAX.
	{M32_I32, 0, 0, C(0x81)},     // Add imm32 to r/m32.
	{M64_I32s, 0, RexW, C(0x81)}, // Add imm32 sign-extended to 64-bits to r/m64.
	{M8_R8, 0, 0, C(0x00)},       // Add r8 to r/m8.
	{M16_R16, 0, 0, C(0x01)},     // Add r16 to r/m16.
	{M32_R32, 0, 0, C(0x01)},     // Add r32 to r/m32.
	{M64_R64, 0, RexW, C(0x01)},  // Add r64 to r/m64.
	{R8_M8, 0, 0, C(0x02)},       // Add r/m8 to r8.
	{R16_M16, 0, 0, C(0x03)},     // Add r/m16 to r16.
	{R32_M32, 0, 0, C(0x03)},     // Add r/m32 to r32.
	{R64_M64, 0, RexW, C(0x03)},  // Add r/m64 to r64.
})
