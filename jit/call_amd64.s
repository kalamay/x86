#include "textflag.h"

TEXT	·call(SB),NOSPLIT,$0
	LEAQ f+0(FP), DI
	MOVQ 8(DX), AX
	JMP AX
