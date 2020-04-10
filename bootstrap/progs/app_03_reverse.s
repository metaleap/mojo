	.text
	.file	"app_03_reverse.ll"
	.globl	writeTo                 # -- Begin function writeTo
	.p2align	4, 0x90
	.type	writeTo,@function
writeTo:                                # @writeTo
	.cfi_startproc
# %bb.0:                                # %begin
	pushq	%r14
	.cfi_def_cfa_offset 16
	pushq	%rbx
	.cfi_def_cfa_offset 24
	pushq	%rax
	.cfi_def_cfa_offset 32
	.cfi_offset %rbx, -24
	.cfi_offset %r14, -16
	movq	%rdx, %r14
	movq	%rsi, %rbx
	movl	$1, %esi
	movq	%rbx, %rdx
	movq	%r14, %rcx
	callq	fwrite@PLT
	cmpq	%rbx, %rax
	jne	.LBB0_2
# %bb.1:
	xorl	%eax, %eax
	jmp	.LBB0_3
.LBB0_2:                                # %err_case
	movq	%r14, %rdi
	callq	ferror@PLT
                                        # kill: def $ax killed $ax def $eax
.LBB0_3:                                # %end
                                        # kill: def $ax killed $ax killed $eax
	addq	$8, %rsp
	.cfi_def_cfa_offset 24
	popq	%rbx
	.cfi_def_cfa_offset 16
	popq	%r14
	.cfi_def_cfa_offset 8
	retq
.Lfunc_end0:
	.size	writeTo, .Lfunc_end0-writeTo
	.cfi_endproc
                                        # -- End function
	.globl	writeToStd              # -- Begin function writeToStd
	.p2align	4, 0x90
	.type	writeToStd,@function
writeToStd:                             # @writeToStd
	.cfi_startproc
# %bb.0:                                # %begin
	pushq	%rax
	.cfi_def_cfa_offset 16
	movq	(%rdx), %rdx
	callq	writeTo@PLT
	cmpw	$1, %ax
	je	.LBB1_2
# %bb.1:                                # %end
	popq	%rax
	.cfi_def_cfa_offset 8
	retq
.LBB1_2:                                # %exit_on_err
	.cfi_def_cfa_offset 16
	movl	$1, %edi
	callq	exit@PLT
.Lfunc_end1:
	.size	writeToStd, .Lfunc_end1-writeToStd
	.cfi_endproc
                                        # -- End function
	.globl	writeErr                # -- Begin function writeErr
	.p2align	4, 0x90
	.type	writeErr,@function
writeErr:                               # @writeErr
	.cfi_startproc
# %bb.0:                                # %b.1
	pushq	%rax
	.cfi_def_cfa_offset 16
	movq	stderr@GOTPCREL(%rip), %rdx
	callq	writeToStd@PLT
	popq	%rax
	.cfi_def_cfa_offset 8
	retq
.Lfunc_end2:
	.size	writeErr, .Lfunc_end2-writeErr
	.cfi_endproc
                                        # -- End function
	.globl	writeOut                # -- Begin function writeOut
	.p2align	4, 0x90
	.type	writeOut,@function
writeOut:                               # @writeOut
	.cfi_startproc
# %bb.0:                                # %b.2
	pushq	%rax
	.cfi_def_cfa_offset 16
	movq	stdout@GOTPCREL(%rip), %rdx
	callq	writeToStd@PLT
	popq	%rax
	.cfi_def_cfa_offset 8
	retq
.Lfunc_end3:
	.size	writeOut, .Lfunc_end3-writeOut
	.cfi_endproc
                                        # -- End function
	.globl	readFrom                # -- Begin function readFrom
	.p2align	4, 0x90
	.type	readFrom,@function
readFrom:                               # @readFrom
	.cfi_startproc
# %bb.0:                                # %b.3
	pushq	%rax
	.cfi_def_cfa_offset 16
	movq	%rdx, %rcx
	movq	%rsi, %rdx
	movl	$1, %esi
	callq	fread@PLT
	popq	%rcx
	.cfi_def_cfa_offset 8
	retq
.Lfunc_end4:
	.size	readFrom, .Lfunc_end4-readFrom
	.cfi_endproc
                                        # -- End function
	.globl	readInOrDie             # -- Begin function readInOrDie
	.p2align	4, 0x90
	.type	readInOrDie,@function
readInOrDie:                            # @readInOrDie
	.cfi_startproc
# %bb.0:                                # %begin
	pushq	%r14
	.cfi_def_cfa_offset 16
	pushq	%rbx
	.cfi_def_cfa_offset 24
	pushq	%rax
	.cfi_def_cfa_offset 32
	.cfi_offset %rbx, -24
	.cfi_offset %r14, -16
	movq	stdin@GOTPCREL(%rip), %rax
	movq	(%rax), %rbx
	movq	%rbx, %rdx
	callq	readFrom@PLT
	movq	%rax, %r14
	movq	%rbx, %rdi
	callq	ferror@PLT
	testw	%ax, %ax
	jne	.LBB5_2
# %bb.1:                                # %end
	movq	%r14, %rax
	addq	$8, %rsp
	.cfi_def_cfa_offset 24
	popq	%rbx
	.cfi_def_cfa_offset 16
	popq	%r14
	.cfi_def_cfa_offset 8
	retq
.LBB5_2:                                # %die
	.cfi_def_cfa_offset 32
	movl	$1, %edi
	callq	exit@PLT
.Lfunc_end5:
	.size	readInOrDie, .Lfunc_end5-readInOrDie
	.cfi_endproc
                                        # -- End function
	.globl	ptrIncr                 # -- Begin function ptrIncr
	.p2align	4, 0x90
	.type	ptrIncr,@function
ptrIncr:                                # @ptrIncr
	.cfi_startproc
# %bb.0:                                # %b.4
	leaq	(%rdi,%rsi), %rax
	retq
.Lfunc_end6:
	.size	ptrIncr, .Lfunc_end6-ptrIncr
	.cfi_endproc
                                        # -- End function
	.globl	swapByte                # -- Begin function swapByte
	.p2align	4, 0x90
	.type	swapByte,@function
swapByte:                               # @swapByte
	.cfi_startproc
# %bb.0:                                # %b.5
	movb	(%rdi), %al
	movb	(%rsi), %cl
	movb	%cl, (%rdi)
	movb	%al, (%rsi)
	retq
.Lfunc_end7:
	.size	swapByte, .Lfunc_end7-swapByte
	.cfi_endproc
                                        # -- End function
	.globl	reverse                 # -- Begin function reverse
	.p2align	4, 0x90
	.type	reverse,@function
reverse:                                # @reverse
	.cfi_startproc
# %bb.0:                                # %begin
	pushq	%r15
	.cfi_def_cfa_offset 16
	pushq	%r14
	.cfi_def_cfa_offset 24
	pushq	%r13
	.cfi_def_cfa_offset 32
	pushq	%r12
	.cfi_def_cfa_offset 40
	pushq	%rbx
	.cfi_def_cfa_offset 48
	.cfi_offset %rbx, -48
	.cfi_offset %r12, -40
	.cfi_offset %r13, -32
	.cfi_offset %r14, -24
	.cfi_offset %r15, -16
	movq	%rsi, %r12
	subq	$1, %r12
	jb	.LBB8_3
# %bb.1:                                # %loop.preheader
	movq	%rsi, %r14
	movq	%rdi, %r15
	shrq	%r14
	xorl	%ebx, %ebx
	.p2align	4, 0x90
.LBB8_2:                                # %loop
                                        # =>This Inner Loop Header: Depth=1
	movq	%r15, %rdi
	movq	%rbx, %rsi
	callq	ptrIncr@PLT
	movq	%rax, %r13
	movq	%r15, %rdi
	movq	%r12, %rsi
	callq	ptrIncr@PLT
	movq	%r13, %rdi
	movq	%rax, %rsi
	callq	swapByte@PLT
	incq	%rbx
	decq	%r12
	cmpq	%r14, %rbx
	jb	.LBB8_2
.LBB8_3:                                # %end
	popq	%rbx
	.cfi_def_cfa_offset 40
	popq	%r12
	.cfi_def_cfa_offset 32
	popq	%r13
	.cfi_def_cfa_offset 24
	popq	%r14
	.cfi_def_cfa_offset 16
	popq	%r15
	.cfi_def_cfa_offset 8
	retq
.Lfunc_end8:
	.size	reverse, .Lfunc_end8-reverse
	.cfi_endproc
                                        # -- End function
	.globl	main                    # -- Begin function main
	.p2align	4, 0x90
	.type	main,@function
main:                                   # @main
	.cfi_startproc
# %bb.0:                                # %b.6
	pushq	%r14
	.cfi_def_cfa_offset 16
	pushq	%rbx
	.cfi_def_cfa_offset 24
	subq	$1032, %rsp             # imm = 0x408
	.cfi_def_cfa_offset 1056
	.cfi_offset %rbx, -24
	.cfi_offset %r14, -16
	leaq	8(%rsp), %r14
	movl	$1024, %esi             # imm = 0x400
	movq	%r14, %rdi
	callq	readInOrDie@PLT
	movq	%rax, %rbx
	movq	%r14, %rdi
	movq	%rax, %rsi
	callq	reverse@PLT
	movq	%r14, %rdi
	movq	%rbx, %rsi
	callq	writeOut@PLT
	xorl	%eax, %eax
	addq	$1032, %rsp             # imm = 0x408
	.cfi_def_cfa_offset 24
	popq	%rbx
	.cfi_def_cfa_offset 16
	popq	%r14
	.cfi_def_cfa_offset 8
	retq
.Lfunc_end9:
	.size	main, .Lfunc_end9-main
	.cfi_endproc
                                        # -- End function

	.section	".note.GNU-stack","",@progbits
