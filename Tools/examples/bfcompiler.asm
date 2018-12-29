mov sp, 0
mov r13, 0 ;screen port
bump r13
send r13, 0x200F
send r13, 0x1000
jmp main
hlt

bfrom:
	dw "++++++++[>++++[>++>+++>+++>+<<<<-]>+>+>->>+[<]<-]>>.>---.+++++++..+++.>>.<-.<.+++.------.--------.>>+.>++."
	dw 0x0000
;compiles into r2asm:
;
;first bytes to reset stuff:
;	mov sp, 0   ;reset stack             0x2020000E
;	mov r13, 0  ;set port                0x2020000D
;	mov r12, 0  ;set ram pointer         0x2020000C
;	mov r11, n  ;set ram index           0x202....B
;   mov r10, m  ;set the getchar address 0x202....A
;extended:
;	mov r9, 0   ;set the storage         0x20200009
;
;the instructions:
;
;	+: add [r11+r12], #+ ;0x24EB000C | #+<<4
;		mask: 0x00E00000 | r11<<16 = B0000 | r12<<0 = C | #+<<4
;		add:  0x24000000
;
;	-: sub [r11+r12], #- ;0x26EB000C | #-<<4
;		mask: 0x00E00000 | r11<<16  = B0000 | r12<<0 = C | #-<<4
;		sub:  0x26000000
;
;   >: add r12, #> ;0x2420000C | #><<4
;		mask: 0x00200000 | r12<<0 = C | #><<4
;		add:  0x24000000
;      and r12, 0xFF ;0x21200FFC
;		mask: 0x00200000 | r12 = C | 0xFF<<4 = 0xFF0
;		and:  0x21000000
;
;   <: sub r12, #< ;0x2620000C | #<<<4
;		mask: 0x00200000 | r12<<0 = C | #<<<4
;		sub:  0x26000000
;      and r12, 0xFF ;0x21200FFC
;		mask: 0x00200000 | r12 = C | 0xFF<<4 = 0xFF0
;		and:  0x21000000
;
;   .: and [r11+r12], 0x00FF ;0x21EB0FFC
;		mask: 0x00E00000 | r11<<16 = B0000 | r12<<0 = C | 0xFF<<4 = FF0
;		and:  0x21000000
;	   send r13, [r11+r12] ;0x3A9B00CD
;		mask: 0x00900000 | r13<<0 = D | r12<<4 = C0 | r11<<16 = B0000
;       send: 0x3A000000
;
;   ,: call r10 ;0x3E0000A0
;		mask: 0 | r10<<4 = A0
;		call: 0x3E000000
;
;totally didnt forget these when testing for the first time:
;
;	[: cmp [r11+r12], 0 ;0x2EEB000C
;		mask: 0x00E00000 | r12<<0 = C | r11<<16 = B0000
;		cmp:  0x2E000000
;	   jz <pairing ], inserted compiletime> ;0x31200..8
;		mask: 0x00200000 | $]<<4
;		jz:   0x31000008
;	]: cmp [r11+r12], 0 ;0x2EEB000C
;		mask: 0x00E00000 | r12<<0 = C | r11<<16 = B0000
;		cmp:  0x2E000000
;      jnz <pairing [, inserted compiletime> ;0x31200..9
;		mask: 0x00200000 | $]<<4
;		jnz:  0x31000009

compile:
	mov r9, 0 ;rom counter
	mov r8, 0 ;temporary instruction storage
	mov r7, 0 ;+/-/>/< counter
	mov r6, 0 ;instruction upper (SWM)
	mov r5, 0 ;instruction lower (MOV)
	mov r4, 0 ;instruction counter
	mov r3, 0 ;old bfrom address counter for ]
	mov r2, 0 ;popped instrrom address counter for ]

	.init:
		;	mov sp, 0   ;reset stack             0x2020000E
		;	mov r13, 0  ;set port                0x2020000D
		;	mov r12, 0  ;set ram pointer         0x2020000C
		;	mov r11, n  ;set ram index           0x202....B
		;   mov r10, m  ;set the getchar address 0x202....A
		mov r6, 0x2020		;pretty much the only MSB set
		swm r6              ;load em' up
		mov r5, 0x000E      ;load sp's instruction first as we'll
							;decrement for the rest (except for) r11
		mov [r4+code], r5   ;encode mov sp(r14), 0
		add r4, 1
		sub r5, 1
		mov [r4+code], r5	;encode mov r13, 0
		add r4, 1
		sub r5, 1
		mov [r4+code], r5	;encode mov r12, 0
		add r4, 1
		sub r5, 1
		;now we need to get the RAM's index address
		mov r8, ram			;using r8 as its safe
		shl r8, 4           ;shift it by 4 to fit the class 2
		or r5, r8           ;or it with LSBs of our instruction
		mov [r4+code], r5	;encode mov r11, ram
		add r4, 1

		mov r8, read_character
		shl r8, 4
		mov r5, 0x000A
		or r5, r8
		mov [r4+code], r5
		add r4, 1
		
	.loop:
		mov r8, [bfrom+r9]  ;fetch the next bf instruction

		cmp r8, '+'         ;check if it is +
		je .instr_add		;^

		cmp r8, '-'			;check if it is -
		je .instr_sub       ;^

		cmp r8, '>'         ;check if it is >
		je .instr_right     ;^

		cmp r8, '<'         ;check if it is <
		je .instr_left      ;^

		cmp r8, '['         ;check if it is [
		je .instr_bropen     ;^

		cmp r8, ']'         ;check if it is ]
		je .instr_brclose    ;^

		cmp r8, '.'         ;check if it is .
		je .instr_out

		cmp r8, ','         ;check if its ,
		je .instr_in		;^

		jmp .end            ;if none match, we hit EOF

		.instr_add:
			add r7, 1           ;add 1 to r7 since we found a +
			add r9, 1           ;set index for the next cell
			mov r8, [bfrom+r9]  ;check the next cell
			cmp r8, '+'         ;check if it is + aswell
			je .instr_add       ;if so, loop
			sub r9, 1           ;else, revert the incerased index, loop
							    ;will take care of r8
			shl r7, 4           ;shifting r7 by 4 to fit the instructi-
							    ;on, doesent matter if the MSBs get lo-
							    ;st as this is an 8 bit bf machine.
			mov r5, r7			;move the number of +s into our instru-
								;ction register         
			mov r7, 0           ;set it to 0 for future use
			or r5, 0x000C       ;load up half of our instruction
			mov r6, 0x24EB      ;load up the rest
			swm r6              ;store MSBs in the 13 bit WO register
			mov [r4+code], r5   ;write the instuction
			add r4, 1           ;move on to the next instruction
			jmp .loopiter       ;continue compilation

		.instr_sub:
			add r7, 1           ;add 1 to r7 since we found a -
			add r9, 1           ;set index for the next cell
			mov r8, [bfrom+r9]  ;check the next cell
			cmp r8, '-'         ;check if it is - aswell
			je .instr_sub       ;if so, loop
			sub r9, 1           ;else, revert the incerased index, loop
							    ;will take care of r8
			shl r7, 4           ;shifting r7 by 4 to fit the instructi-
							    ;on, doesent matter if the MSBs get lo-
							    ;st as this is an 8 bit bf machine.
			mov r5, r7			;move the number of -s into our instru-
								;ction register         
			mov r7, 0           ;set it to 0 for future use
			or r5, 0x000C       ;load up half of our instruction
			mov r6, 0x26EB      ;load up the rest
			swm r6              ;store MSBs in the 13 bit WO register
			mov [r4+code], r5   ;write the instuction
			add r4, 1           ;move on to the next instruction
			jmp .loopiter       ;continue compilation

		.instr_right:
			add r7, 1           ;add 1 to r7 since we found a >
			add r9, 1           ;set index for the next cell
			mov r8, [bfrom+r9]  ;check the next cell
			cmp r8, '>'         ;check if it is > aswell
			je .instr_right     ;if so, loop
			sub r9, 1           ;else, revert the incerased index, loop
							    ;will take care of r8
			shl r7, 4           ;shifting r7 by 4 to fit the instructi-
							    ;on, doesent matter if the MSBs get lo-
							    ;st as this is an 8 bit bf machine.
			mov r5, r7			;move the number of >s into our instru-
								;ction register         
			mov r7, 0           ;set it to 0 for future use
			or r5, 0x000C       ;load up half of our instruction
			mov r6, 0x2420      ;load up the rest
			swm r6              ;store MSBs in the 13 bit WO register
			mov [r4+code], r5   ;write the instuction
			add r4, 1           ;move on to the next instruction

			mov r5, 0x3FFC      ;we need to keep the number in 10 bit
								;bounds soo we encode `and r12, 0x3FF`
			mov r6, 0x2120      ;second part
			swm r6              ;load up the MSBs
			mov [r4+code], r5   ;encode it
			add r4, 1           ;move on to the next instruction

			jmp .loopiter       ;continue compilation

		.instr_left:
			add r7, 1           ;add 1 to r7 since we found a <
			add r9, 1           ;set index for the next cell
			mov r8, [bfrom+r9]  ;check the next cell
			cmp r8, '<'         ;check if it is < aswell
			je .instr_left       ;if so, loop
			sub r9, 1           ;else, revert the incerased index, loop
							    ;will take care of r8
			shl r7, 4           ;shifting r7 by 4 to fit the instructi-
							    ;on, doesent matter if the MSBs get lo-
							    ;st as this is an 8 bit bf machine.
			mov r5, r7			;move the number of <s into our instru-
								;ction register         
			mov r7, 0           ;set it to 0 for future use
			or r5, 0x000C       ;load up half of our instruction
			mov r6, 0x2620      ;load up the rest
			swm r6              ;store MSBs in the 13 bit WO register
			mov [r4+code], r5   ;write the instuction
			add r4, 1           ;move on to the next instruction

			mov r5, 0x3FFC      ;we need to keep the number in 10 bit
								;bounds soo we encode `and r12, 0x3FF`
			mov r6, 0x2120      ;second part
			swm r6              ;load up the MSBs
			mov [r4+code], r5   ;encode it
			add r4, 1           ;move on to the next instruction

			jmp .loopiter       ;continue compilation
		.instr_bropen:
			mov r6, 0x2EEB      ;the first instruction is static
			mov r5, 0x000C      ;which is 0x2EEB000C
			swm r6              ;load 0x2E2B
			mov [r4+code], r5   ;encode it
			add r4, 1           ;incerment our PC
			push r4				;push the current instruction as the
								;paired ] will use it
			mov r6, 0           ;placeholder since we cant OR a memory
								;location without resetting the upper
								;13 bits, going to be placed
			swm r6				;in .instr_brclose
			mov r5, 0           ;
			mov [r4+code], r5   ;encode it
			add r4, 1           ;incerment our PC
			jmp .loopiter       ;continue

		.instr_brclose:
			mov r6, 0x2EEB      ;the first instruction is static
			mov r5, 0x000C      ;which is 0x2EEB000C
			swm r6              ;load 0x2E2B
			mov [r4+code], r5   ;encode it
			add r4, 1           ;incerment our PC

			pop r2              ;get the pairing [s index

			;edit [r2+code]:
			;	OR it with [r4+code] shifted by 4
			mov r5, r4          ;move r4 into r5 since it is used by
								;the instruction
			add r5, code        ;the instr. uses [r4+code]
			shl r5, 4           ;shift it by 4 to fit the mask
			or r5, 0x0008       ;add mask
			mov r6, 0x3120      ;upper mask bits
			swm r6              ;set write mask
			mov [r2+code], r5   ;write the pairing ['s 2nd instruction
			;[r2+code] is now 0x312(r4+code(current instruction))8

			;edit [r4+code]:
			;	encode: jnz [r2+code]
			add r2, code 		;add the offset to r2
			shl r2, 4			;shift it by 4 to fit the mask
			mov r5, r2			;move it to the instruction upper reg.
			or r5, 0x0009		;or it with for the mask
			mov r6, 0x3120		;load the upper bits
			mov [r4+code], r5	;encode
			add r4, 1           ;incerment our PC

			;limitations:
			;	* we can't exceed 12 bit addressing though it won't be
			;	  an issue since R216 can at max reach 4096 bytes of
			;	  RAM (12 bit addressed)
			jmp .loopiter

		.instr_out: ;note: fuck yes no optimisation needed			  
			mov r5, 0x0FFC		;we need to keep the values in 8 bit
								;bounds and the best way to do it is
								;when it is output (for speed)
			mov r6, 0x21EB      ;
			swm r6              ;load up the MSBs
			mov [r4+code], r5   ;encode it
			add r4, 1           ;move on to the next instruction

			mov r5, 0x00CD      ;and now for sending, we encode
								;send r13, [r11+r12]
			mov r6, 0x3A9B      ;MSBs
			swm r6              ;load the MSBs
			mov [r4+code], r5   ;encode it
			add r4, 1           ;move on to the next instruction
			jmp .loopiter       ;continue compilation

		.instr_in:
			mov r5, 0x00A0		;first part of the call
			mov r6, 0x3E00      ;2nd part
			swm r6 				;set write mask
			mov [r4+code], r5	;encode the instruction
			add r4, 1			;incerment the PC
			jmp .loopiter		;continue


	.loopiter:
		add r9, 1
		;mov r6, 0 ;reset SWM data since it messes up memory writes
		cmp r9, 256 ;257th instruction, out of rom bounds
		jne .loop   ;if we are in bounds, continue
	.end:
	mov r6, 0x3000
	swm r6
	mov r5, 0x0000
	mov [r4+code], r5 ;append a HLT
	ret

main:
	call compile
	mov r6, 0 ;reset SWM data since it messes up memory writes
	swm r6
	;jn 1
	jmp code
	hlt

;;;CREDIT TO LBPHACKER FOR THIS FUNCTION!!
;;;MODIFIED TO FIT THIS PROGRAM
; * Reads a single character from the terminal.
; * Character code is returned in [r11+r12].
; * r13 is terminal port address.
read_character:
push r0
.wait_loop:
    wait r3                   ; * Wait for a bump. r3 should be checked but
                              ;   as in this demo there's no other peripheral,
                              ;   it's fine this way.
    js .wait_loop
    bump r13                  ; * Ask for character code.
.recv_loop:
    recv r0, r13       ; * Receive character code.
    jnc .recv_loop            ; * The carry bit it set if something is received.
    mov [r11+r12], r0
    pop r0
    ret

    dw 0xDEAD
    dw 0xBEEF

ram:
	dw 0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0
	dw 0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0
	dw 0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0
	dw 0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0
	dw 0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0
	dw 0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0
	dw 0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0
	dw 0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0
	dw 0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0
	dw 0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0
	dw 0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0
	dw 0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0
	dw 0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0
	dw 0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0
	dw 0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0
	dw 0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0
	;256byte ram

org 640
code:
	dw 0x0000
