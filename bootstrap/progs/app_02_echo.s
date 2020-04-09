	.text
	.file	"app_02_echo.ll"
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
	.globl	main                    # -- Begin function main
	.p2align	4, 0x90
	.type	main,@function
main:                                   # @main
	.cfi_startproc
# %bb.0:                                # %b.4
	pushq	%rbx
	.cfi_def_cfa_offset 16
	subq	$1024, %rsp             # imm = 0x400
	.cfi_def_cfa_offset 1040
	.cfi_offset %rbx, -16
	movq	%rsp, %rbx
	movl	$1024, %esi             # imm = 0x400
	movq	%rbx, %rdi
	callq	readInOrDie@PLT
	movq	%rbx, %rdi
	movq	%rax, %rsi
	callq	writeOut@PLT
	xorl	%eax, %eax
	addq	$1024, %rsp             # imm = 0x400
	.cfi_def_cfa_offset 16
	popq	%rbx
	.cfi_def_cfa_offset 8
	retq
.Lfunc_end6:
	.size	main, .Lfunc_end6-main
	.cfi_endproc
                                        # -- End function

	.section	".note.GNU-stack","",@progbits
