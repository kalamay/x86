package inst

import . "github.com/kalamay/x86/asm/amd64"

var ADD = &InstSet{"add", S32, []Inst{
	{M8_I8, 0, 0, C(0x80)},          // Add imm8 to r/m8.
	{M8_I8s, 0, Rex, C(0x80)},       // Add sign-extended imm8 to r/m8.
	{M16_I16, OpSize, 0, C(0x81)},   // Add imm16 to r/m16.
	{M32_I32, 0, 0, C(0x81)},        // Add imm32 to r/m32.
	{M64_I32s, 0, RexW, C(0x81)},    // Add imm32 sign-extended to 64-bits to r/m64.
	{M16_I8s, AddrSize, 0, C(0x83)}, // Add sign-extended imm8 to r/m16.
	{M32_I8s, 0, 0, C(0x83)},        // Add sign-extended imm8 to r/m32.
	{M64_I8s, 0, RexW, C(0x83)},     // Add sign-extended imm8 to r/m64.
	{M8_R8, 0, 0, C(0x00)},          // Add r8 to r/m8.
	{M8_R8, 0, Rex, C(0x00)},        // Add r8 to r/m8.
	// 01 /r	ADD r/m16, r16	MR	Valid	Valid	Add r16 to r/m16.
	// 01 /r	ADD r/m32, r32	MR	Valid	Valid	Add r32 to r/m32.
	// REX.W + 01 /r	ADD r/m64, r64	MR	Valid	N.E.	Add r64 to r/m64.
	// 02 /r	ADD r8, r/m8	RM	Valid	Valid	Add r/m8 to r8.
	// REX + 02 /r	ADD r8*, r/m8*	RM	Valid	N.E.	Add r/m8 to r8.
	// 03 /r	ADD r16, r/m16	RM	Valid	Valid	Add r/m16 to r16.
	// 03 /r	ADD r32, r/m32	RM	Valid	Valid	Add r/m32 to r32.
	// REX.W + 03 /r	ADD r64, r/m64	RM	Valid	N.E.	Add r/m64 to r64.
}}
