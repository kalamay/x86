package inst

import . "github.com/kalamay/x86/asm/amd64"

var MOV = &InstSet{"MOV", S32, []Inst{
	{M8_R8, 0, 0, C1 | 0x88},       // Move r8 to r/m8.
	{M16_R16, 0, 0, C1 | 0x89},     // Move r16 to r/m16.
	{M32_R32, 0, 0, C1 | 0x89},     // Move r32 to r/m32.
	{M64_R64, 0, RexW, C1 | 0x89},  // Move r64 to r/m64.
	{R8_M8, 0, 0, C1 | 0x8A},       // Move r/m8 to r8.
	{R16_M16, 0, 0, C1 | 0x8B},     // Move r/m16 to r16.
	{R32_M32, 0, 0, C1 | 0x8B},     // Move r/m32 to r32.
	{R64_M64, 0, RexW, C1 | 0x8B},  // Move r/m64 to r64.
	{O8_I8, 0, 0, C1 | 0xB0},       // Move imm8 to r8.
	{M8_I8, 0, 0, C1 | 0xC6},       // Move imm8 to r/m8.
	{O16_I16, 0, 0, C1 | 0xB8},     // Move imm16 to r16.
	{M16_I16, 0, 0, C1 | 0xC7},     // Move imm16 to r/m16.
	{O32_I32, 0, 0, C1 | 0xB8},     // Move imm32 to r32.
	{M32_I32, 0, 0, C1 | 0xC7},     // Move imm32 to r/m32.
	{M64_I32s, 0, RexW, C1 | 0xC7}, // Move imm32 sign extended to 64-bits to r/m64.
	{O64_I64, 0, RexW, C1 | 0xB8},  // Move imm64 to r64.
}}
