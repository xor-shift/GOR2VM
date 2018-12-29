mov sp, 0
jmp main

BFROM: 
dw "++++[->-[-]<]"
dw 0x0000

interpret:
	mov r0, 0;BFIP
	mov r1, 0;INSTR
	mov r2, 0;DPTR
	mov r3, 0;TEMP
	mov r4, 0;IOPORT(constant)
	mov r5, 0;STATUS
	mov r7, 0; used for [r2+BFRAM] *reads*, not writes

	.while:
	cmp r5, 2
	je .while_end
		mov r1, [r0+BFROM]
		;mov r7, [r2+BFRAM] ;moved and optimised
		cmp r1, 0
		jne .statusCheck
			ret
		.statusCheck:
			cmp r5, 1
			jne .statusAndNullCheck_cont
			cmp r1, 0x005D
			jne .statusAndNullCheck_cont
			mov r5, 0
			add r0, 1
			jmp .while
		.statusAndNullCheck_cont: ;switch start
		cmp r1, 0x002B
		je .instr_add
		cmp r1, 0x002D
		je .instr_sub
		cmp r1, 0x003E
		je .instr_right
		cmp r1, 0x003C
		je .instr_left
		cmp r1, 0x005B
		je .instr_label
		cmp r1, 0x005D
		je .instr_jump
		cmp r1, 0x002E
		je .instr_out
		;default:
		ret

		.instr_add:
			add r7, 1
			and r7, 0x00FF ;8bit
			mov [r2+BFRAM], r7
			jmp .switchEnd
		.instr_sub:
			sub r7, 1
			and r7, 0x00FF ;8bit
			mov [r2+BFRAM], r7
			jmp .switchEnd
		.instr_right:
			add r2, 1
			mov r7, [r2+BFRAM]
			cmp r2, 255
			jna .switchEnd
				mov r2, 0
				mov r7, [r2+BFRAM]
			jmp .switchEnd
		.instr_left:
			sub r2, 1
			mov r7, [r2+BFRAM]
			cmp r2, 255
			jna .switchEnd
				mov r2, 255
				mov r7, [r2+BFRAM]
			jmp .switchEnd
		.instr_label:
			cmp r7, 0
			jna .instr_label_else
				push r0
				jmp .switchEnd
			.instr_label_else:
				mov r5, 0
				jmp .switchEnd
		.instr_jump:
			pop r3
			cmp r7, 0
			jna .instr_jump_else
				mov r0, r3
				;hlt
				jmp .while
			.instr_jump_else:
				jmp .switchEnd
		.instr_out:
			;hlt
			send r4, r7
			jmp .switchEnd
		.instr_in:
			;NYI
			jmp .switchEnd
		.switchEnd:
		add r0, 1
		jmp .while
	.while_end:

	ret

main:
	send sp, 0x200F
	send sp, 0x1000
	call interpret
	hlt
	
BFRAM:
	dw 0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0
	dw 0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0
	dw 0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0
	dw 0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0
	;256 byte ram
