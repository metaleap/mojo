	.text
	.file	"app_04_atmo_tokenize.ll"
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
	.globl	swapBytes               # -- Begin function swapBytes
	.p2align	4, 0x90
	.type	swapBytes,@function
swapBytes:                              # @swapBytes
	.cfi_startproc
# %bb.0:                                # %b.5
	movb	(%rdi), %al
	movb	(%rsi), %cl
	movb	%cl, (%rdi)
	movb	%al, (%rsi)
	retq
.Lfunc_end7:
	.size	swapBytes, .Lfunc_end7-swapBytes
	.cfi_endproc
                                        # -- End function
	.globl	main                    # -- Begin function main
	.p2align	4, 0x90
	.type	main,@function
main:                                   # @main
	.cfi_startproc
# %bb.0:                                # %b.6
	pushq	%rbp
	.cfi_def_cfa_offset 16
	.cfi_offset %rbp, -16
	movq	%rsp, %rbp
	.cfi_def_cfa_register %rbp
	movq	%rsp, %rdi
	addq	$-1073674240, %rdi      # imm = 0xC0010800
	movq	%rdi, %rsp
	movl	$1073674240, %esi       # imm = 0x3FFEF800
	callq	readInOrDie@PLT
	xorl	%eax, %eax
	movq	%rbp, %rsp
	popq	%rbp
	.cfi_def_cfa %rsp, 8
	retq
.Lfunc_end8:
	.size	main, .Lfunc_end8-main
	.cfi_endproc
                                        # -- End function
	.type	tok_kind_comment,@object # @tok_kind_comment
	.section	.rodata,"a",@progbits
	.globl	tok_kind_comment
tok_kind_comment:
	.ascii	"comment"
	.size	tok_kind_comment, 7

	.type	tok_kind_ident,@object  # @tok_kind_ident
	.globl	tok_kind_ident
tok_kind_ident:
	.ascii	"ident"
	.size	tok_kind_ident, 5

	.type	tok_kind_lit_int,@object # @tok_kind_lit_int
	.globl	tok_kind_lit_int
tok_kind_lit_int:
	.ascii	"lit_int"
	.size	tok_kind_lit_int, 7

	.type	tok_kind_lit_str,@object # @tok_kind_lit_str
	.globl	tok_kind_lit_str
tok_kind_lit_str:
	.ascii	"lit_str"
	.size	tok_kind_lit_str, 7

	.type	tok_kind_sep_bparen_open,@object # @tok_kind_sep_bparen_open
	.globl	tok_kind_sep_bparen_open
tok_kind_sep_bparen_open:
	.ascii	"sep_bparen_open"
	.size	tok_kind_sep_bparen_open, 15

	.type	tok_kind_sep_bparen_close,@object # @tok_kind_sep_bparen_close
	.globl	tok_kind_sep_bparen_close
tok_kind_sep_bparen_close:
	.ascii	"sep_bparen_close"
	.size	tok_kind_sep_bparen_close, 16

	.type	tok_kind_sep_bcurly_open,@object # @tok_kind_sep_bcurly_open
	.globl	tok_kind_sep_bcurly_open
tok_kind_sep_bcurly_open:
	.ascii	"sep_bcurly_open"
	.size	tok_kind_sep_bcurly_open, 15

	.type	tok_kind_sep_bcurly_close,@object # @tok_kind_sep_bcurly_close
	.globl	tok_kind_sep_bcurly_close
tok_kind_sep_bcurly_close:
	.ascii	"sep_bcurly_close"
	.size	tok_kind_sep_bcurly_close, 16

	.type	tok_kind_sep_bsquare_open,@object # @tok_kind_sep_bsquare_open
	.globl	tok_kind_sep_bsquare_open
tok_kind_sep_bsquare_open:
	.ascii	"sep_bsquare_open"
	.size	tok_kind_sep_bsquare_open, 16

	.type	tok_kind_sep_bsquare_close,@object # @tok_kind_sep_bsquare_close
	.globl	tok_kind_sep_bsquare_close
	.p2align	4
tok_kind_sep_bsquare_close:
	.ascii	"sep_bsquare_close"
	.size	tok_kind_sep_bsquare_close, 17

	.type	tok_kind_sep_comma,@object # @tok_kind_sep_comma
	.globl	tok_kind_sep_comma
tok_kind_sep_comma:
	.ascii	"sep_comma"
	.size	tok_kind_sep_comma, 9


	.section	".note.GNU-stack","",@progbits
