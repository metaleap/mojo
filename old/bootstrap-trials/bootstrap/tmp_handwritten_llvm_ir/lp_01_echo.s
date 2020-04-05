	.text
	.file	"lp_01_echo.ll"
	.globl	main                    # -- Begin function main
	.p2align	4, 0x90
	.type	main,@function
main:                                   # @main
	.cfi_startproc
# %bb.0:                                # %output_prompt
	pushq	%rbp
	.cfi_def_cfa_offset 16
	.cfi_offset %rbp, -16
	movq	%rsp, %rbp
	.cfi_def_cfa_register %rbp
	pushq	%r15
	pushq	%r14
	pushq	%r13
	pushq	%r12
	pushq	%rbx
	pushq	%rax
	.cfi_offset %rbx, -56
	.cfi_offset %r12, -48
	.cfi_offset %r13, -40
	.cfi_offset %r14, -32
	.cfi_offset %r15, -24
	movq	stdout@GOTPCREL(%rip), %rax
	movq	(%rax), %r15
	movq	str.1@GOTPCREL(%rip), %rdi
	movl	$1, %esi
	movl	$26, %edx
	movq	%r15, %rcx
	callq	fwrite@PLT
	movl	$1, %r14d
	cmpq	$26, %rax
	jne	.LBB0_4
# %bb.1:                                # %read_input
	movq	stdin@GOTPCREL(%rip), %rax
	movq	(%rax), %r13
	movq	%rsp, %rbx
	addq	$-512, %rbx             # imm = 0xFE00
	movq	%rbx, %rsp
	movl	$1, %esi
	movl	$512, %edx              # imm = 0x200
	movq	%rbx, %rdi
	movq	%r13, %rcx
	callq	fread@PLT
	movq	%rax, %r12
	movq	%r13, %rdi
	callq	ferror@PLT
	testl	%eax, %eax
	jne	.LBB0_4
# %bb.2:                                # %output_result
	movl	$1, %esi
	movq	%rbx, %rdi
	movq	%r12, %rdx
	movq	%r15, %rcx
	callq	fwrite@PLT
	cmpq	%r12, %rax
	jne	.LBB0_4
# %bb.3:                                # %ret_ok
	xorl	%r14d, %r14d
.LBB0_4:                                # %return
	movl	%r14d, %eax
	leaq	-40(%rbp), %rsp
	popq	%rbx
	popq	%r12
	popq	%r13
	popq	%r14
	popq	%r15
	popq	%rbp
	.cfi_def_cfa %rsp, 8
	retq
.Lfunc_end0:
	.size	main, .Lfunc_end0-main
	.cfi_endproc
                                        # -- End function
	.type	str.1,@object           # @str.1
	.section	.rodata,"a",@progbits
	.globl	str.1
	.p2align	4
str.1:
	.ascii	"I'll echo what you enter: "
	.size	str.1, 26


	.section	".note.GNU-stack","",@progbits
