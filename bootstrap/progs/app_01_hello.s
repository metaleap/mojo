	.text
	.file	"app_01_hello.ll"
	.globl	writeTo                 # -- Begin function writeTo
	.p2align	4, 0x90
	.type	writeTo,@function
writeTo:                                # @writeTo
	.cfi_startproc
# %bb.0:                                # %b.1
	pushq	%rbx
	.cfi_def_cfa_offset 16
	.cfi_offset %rbx, -16
	movq	%rdx, %rbx
	movq	%rsi, %rdx
	movl	$1, %esi
	movq	%rbx, %rcx
	callq	fwrite@PLT
	movq	%rbx, %rdi
	callq	ferror@PLT
	popq	%rbx
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
# %bb.0:                                # %b.2
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
# %bb.0:                                # %b.4
	pushq	%rax
	.cfi_def_cfa_offset 16
	movq	msg@GOTPCREL(%rip), %rdi
	movl	$11, %esi
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
