	.text
	.file	"app_01_hello.ll"
	.globl	strPtrOf                # -- Begin function strPtrOf
	.p2align	4, 0x90
	.type	strPtrOf,@function
strPtrOf:                               # @strPtrOf
	.cfi_startproc
# %bb.0:                                # %b.2
	movq	%rdi, %rax
	retq
.Lfunc_end0:
	.size	strPtrOf, .Lfunc_end0-strPtrOf
	.cfi_endproc
                                        # -- End function
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
	jne	.LBB1_2
# %bb.1:
	xorl	%eax, %eax
	jmp	.LBB1_3
.LBB1_2:                                # %err_case
	movq	%r14, %rdi
	callq	ferror@PLT
                                        # kill: def $ax killed $ax def $eax
.LBB1_3:                                # %end
                                        # kill: def $ax killed $ax killed $eax
	addq	$8, %rsp
	.cfi_def_cfa_offset 24
	popq	%rbx
	.cfi_def_cfa_offset 16
	popq	%r14
	.cfi_def_cfa_offset 8
	retq
.Lfunc_end1:
	.size	writeTo, .Lfunc_end1-writeTo
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
	je	.LBB2_2
# %bb.1:                                # %end
	popq	%rax
	.cfi_def_cfa_offset 8
	retq
.LBB2_2:                                # %exit_on_err
	.cfi_def_cfa_offset 16
	movl	$1, %edi
	callq	exit@PLT
.Lfunc_end2:
	.size	writeToStd, .Lfunc_end2-writeToStd
	.cfi_endproc
                                        # -- End function
	.globl	writeOut                # -- Begin function writeOut
	.p2align	4, 0x90
	.type	writeOut,@function
writeOut:                               # @writeOut
	.cfi_startproc
# %bb.0:                                # %b.3
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
	.globl	main                    # -- Begin function main
	.p2align	4, 0x90
	.type	main,@function
main:                                   # @main
	.cfi_startproc
# %bb.0:                                # %b.1
	pushq	%rax
	.cfi_def_cfa_offset 16
	movq	msg@GOTPCREL(%rip), %rdi
	callq	strPtrOf@PLT
	movl	$11, %esi
	movq	%rax, %rdi
	callq	writeOut@PLT
	xorl	%eax, %eax
	popq	%rcx
	.cfi_def_cfa_offset 8
	retq
.Lfunc_end4:
	.size	main, .Lfunc_end4-main
	.cfi_endproc
                                        # -- End function
	.type	msg,@object             # @msg
	.section	.rodata,"a",@progbits
	.globl	msg
msg:
	.ascii	"Hola Welt.\n"
	.size	msg, 11


	.section	".note.GNU-stack","",@progbits
