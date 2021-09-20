#include "textflag.h"

TEXT	·memset16(SB), NOSPLIT, $0-25
	MOVQ b_base+0(FP), SI
	MOVQ b_len+8(FP), DI
	MOVB c+24(FP), AL
	MOVD AX, X0
	VPBROADCASTB X0, Y0

	CMPQ DI, $128
	JAE set128
	CMPQ DI, $64
	JE set64
	CMPQ DI, $32
	JE set32
	JMP set16

set128:
	ADDQ SI, DI
loop128:
	VMOVDQU Y0, 0(SI)
	VMOVDQU Y0, 32(SI)
	VMOVDQU Y0, 64(SI)
	VMOVDQU Y0, 96(SI)
	ADDQ $128, SI
	CMPQ SI, DI
	JB loop128
	RET

set64:
	VMOVDQU Y0, 0(SI)
	VMOVDQU Y0, 32(SI)
	RET

set32:
	VMOVDQU Y0, 0(SI)
	RET

set16:
	VMOVDQU X0, 0(SI)
	RET
