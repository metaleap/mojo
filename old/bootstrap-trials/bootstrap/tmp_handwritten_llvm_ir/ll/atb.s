	.text
	.file	"atb.ll"
	.globl	writeTo                 # -- Begin function writeTo
	.p2align	4, 0x90
	.type	writeTo,@function
writeTo:                                # @writeTo
	.cfi_startproc
# %bb.0:
	pushq	%rbx
	.cfi_def_cfa_offset 16
	.cfi_offset %rbx, -16
	movq	%rsi, %rax
	movq	(%rdx), %rbx
	movl	$1, %esi
	movq	%rax, %rdx
	movq	%rbx, %rcx
	callq	fwrite@PLT
	movq	%rbx, %rdi
	callq	ferror@PLT
	cmpl	$1, %eax
	jne	.LBB0_2
# %bb.1:                                # %exit_on_err
	movl	$1, %edi
	callq	exit@PLT
.LBB0_2:                                # %return
	popq	%rbx
	.cfi_def_cfa_offset 8
	retq
.Lfunc_end0:
	.size	writeTo, .Lfunc_end0-writeTo
	.cfi_endproc
                                        # -- End function
	.globl	writeErr                # -- Begin function writeErr
	.p2align	4, 0x90
	.type	writeErr,@function
writeErr:                               # @writeErr
	.cfi_startproc
# %bb.0:
	pushq	%rax
	.cfi_def_cfa_offset 16
	movq	stderr@GOTPCREL(%rip), %rdx
	callq	writeTo@PLT
	popq	%rax
	.cfi_def_cfa_offset 8
	retq
.Lfunc_end1:
	.size	writeErr, .Lfunc_end1-writeErr
	.cfi_endproc
                                        # -- End function
	.globl	writeOut                # -- Begin function writeOut
	.p2align	4, 0x90
	.type	writeOut,@function
writeOut:                               # @writeOut
	.cfi_startproc
# %bb.0:
	pushq	%rax
	.cfi_def_cfa_offset 16
	movq	stdout@GOTPCREL(%rip), %rdx
	callq	writeTo@PLT
	popq	%rax
	.cfi_def_cfa_offset 8
	retq
.Lfunc_end2:
	.size	writeOut, .Lfunc_end2-writeOut
	.cfi_endproc
                                        # -- End function
	.globl	main                    # -- Begin function main
	.p2align	4, 0x90
	.type	main,@function
main:                                   # @main
	.cfi_startproc
# %bb.0:                                # %output_prompt
	pushq	%r14
	.cfi_def_cfa_offset 16
	pushq	%rbx
	.cfi_def_cfa_offset 24
	subq	$520, %rsp              # imm = 0x208
	.cfi_def_cfa_offset 544
	.cfi_offset %rbx, -24
	.cfi_offset %r14, -16
	movq	str.1@GOTPCREL(%rip), %rdi
	movl	$26, %esi
	callq	writeOut@PLT
	movq	stdin@GOTPCREL(%rip), %rax
	movq	(%rax), %rbx
	leaq	8(%rsp), %rdi
	movl	$1, %esi
	movl	$512, %edx              # imm = 0x200
	movq	%rbx, %rcx
	callq	fread@PLT
	movq	%rax, %r14
	movq	%rbx, %rdi
	callq	ferror@PLT
	movl	$1, %ebx
	testl	%eax, %eax
	jne	.LBB3_3
# %bb.1:                                # %output_result
	movq	stdout@GOTPCREL(%rip), %rax
	movq	(%rax), %rcx
	leaq	8(%rsp), %rdi
	movl	$1, %esi
	movq	%r14, %rdx
	callq	fwrite@PLT
	cmpq	%r14, %rax
	jne	.LBB3_3
# %bb.2:                                # %ret_ok
	xorl	%ebx, %ebx
.LBB3_3:                                # %return
	movl	%ebx, %eax
	addq	$520, %rsp              # imm = 0x208
	.cfi_def_cfa_offset 24
	popq	%rbx
	.cfi_def_cfa_offset 16
	popq	%r14
	.cfi_def_cfa_offset 8
	retq
.Lfunc_end3:
	.size	main, .Lfunc_end3-main
	.cfi_endproc
                                        # -- End function
	.type	str.1,@object           # @str.1
	.section	.rodata,"a",@progbits
	.globl	str.1
	.p2align	4
str.1:
	.ascii	"I'll echo what you enter. "
	.size	str.1, 26


	.section	".note.GNU-stack","",@progbits
