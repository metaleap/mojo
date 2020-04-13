	.text
	.file	"app_03_reverse.ll"
	.globl	readFrom                # -- Begin function readFrom
	.p2align	4, 0x90
	.type	readFrom,@function
readFrom:                               # @readFrom
	.cfi_startproc
# %bb.0:                                # %b.2
	pushq	%rax
	.cfi_def_cfa_offset 16
	movq	%rdx, %rcx
	movq	%rsi, %rdx
	movl	$1, %esi
	callq	fread@PLT
	popq	%rcx
	.cfi_def_cfa_offset 8
	retq
.Lfunc_end0:
	.size	readFrom, .Lfunc_end0-readFrom
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
	jne	.LBB1_2
# %bb.1:                                # %end
	movq	%r14, %rax
	addq	$8, %rsp
	.cfi_def_cfa_offset 24
	popq	%rbx
	.cfi_def_cfa_offset 16
	popq	%r14
	.cfi_def_cfa_offset 8
	retq
.LBB1_2:                                # %die
	.cfi_def_cfa_offset 32
	movl	$1, %edi
	callq	exit@PLT
.Lfunc_end1:
	.size	readInOrDie, .Lfunc_end1-readInOrDie
	.cfi_endproc
                                        # -- End function
	.globl	ptrIncr                 # -- Begin function ptrIncr
	.p2align	4, 0x90
	.type	ptrIncr,@function
ptrIncr:                                # @ptrIncr
	.cfi_startproc
# %bb.0:                                # %b.3
	leaq	(%rdi,%rsi), %rax
	retq
.Lfunc_end2:
	.size	ptrIncr, .Lfunc_end2-ptrIncr
	.cfi_endproc
                                        # -- End function
	.globl	swapBytes               # -- Begin function swapBytes
	.p2align	4, 0x90
	.type	swapBytes,@function
swapBytes:                              # @swapBytes
	.cfi_startproc
# %bb.0:                                # %b.4
	movb	(%rdi), %al
	movb	(%rsi), %cl
	movb	%cl, (%rdi)
	movb	%al, (%rsi)
	retq
.Lfunc_end3:
	.size	swapBytes, .Lfunc_end3-swapBytes
	.cfi_endproc
                                        # -- End function
	.globl	reverseBytes            # -- Begin function reverseBytes
	.p2align	4, 0x90
	.type	reverseBytes,@function
reverseBytes:                           # @reverseBytes
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
	cmpq	$2, %rsi
	jb	.LBB4_3
# %bb.1:                                # %prep
	movq	%rsi, %r15
	movq	%rdi, %r14
	movq	%rsi, %r13
	shrq	%r13
	decq	%r15
	xorl	%ebx, %ebx
	.p2align	4, 0x90
.LBB4_2:                                # %loop
                                        # =>This Inner Loop Header: Depth=1
	movq	%r14, %rdi
	movq	%rbx, %rsi
	callq	ptrIncr@PLT
	movq	%rax, %r12
	movq	%r14, %rdi
	movq	%r15, %rsi
	callq	ptrIncr@PLT
	movq	%r12, %rdi
	movq	%rax, %rsi
	callq	swapBytes@PLT
	incq	%rbx
	decq	%r15
	cmpq	%r13, %rbx
	jb	.LBB4_2
.LBB4_3:                                # %end
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
.Lfunc_end4:
	.size	reverseBytes, .Lfunc_end4-reverseBytes
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
	jne	.LBB5_2
# %bb.1:
	xorl	%eax, %eax
	jmp	.LBB5_3
.LBB5_2:                                # %err_case
	movq	%r14, %rdi
	callq	ferror@PLT
                                        # kill: def $ax killed $ax def $eax
.LBB5_3:                                # %end
                                        # kill: def $ax killed $ax killed $eax
	addq	$8, %rsp
	.cfi_def_cfa_offset 24
	popq	%rbx
	.cfi_def_cfa_offset 16
	popq	%r14
	.cfi_def_cfa_offset 8
	retq
.Lfunc_end5:
	.size	writeTo, .Lfunc_end5-writeTo
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
	je	.LBB6_2
# %bb.1:                                # %end
	popq	%rax
	.cfi_def_cfa_offset 8
	retq
.LBB6_2:                                # %exit_on_err
	.cfi_def_cfa_offset 16
	movl	$1, %edi
	callq	exit@PLT
.Lfunc_end6:
	.size	writeToStd, .Lfunc_end6-writeToStd
	.cfi_endproc
                                        # -- End function
	.globl	writeOut                # -- Begin function writeOut
	.p2align	4, 0x90
	.type	writeOut,@function
writeOut:                               # @writeOut
	.cfi_startproc
# %bb.0:                                # %b.5
	pushq	%rax
	.cfi_def_cfa_offset 16
	movq	stdout@GOTPCREL(%rip), %rdx
	callq	writeToStd@PLT
	popq	%rax
	.cfi_def_cfa_offset 8
	retq
.Lfunc_end7:
	.size	writeOut, .Lfunc_end7-writeOut
	.cfi_endproc
                                        # -- End function
	.globl	main                    # -- Begin function main
	.p2align	4, 0x90
	.type	main,@function
main:                                   # @main
	.cfi_startproc
# %bb.0:                                # %b.1
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
	callq	reverseBytes@PLT
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
.Lfunc_end8:
	.size	main, .Lfunc_end8-main
	.cfi_endproc
                                        # -- End function

	.section	".note.GNU-stack","",@progbits
