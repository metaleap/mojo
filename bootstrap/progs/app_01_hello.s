	.text
	.file	"app_01_hello.ll"
	.globl	main                    # -- Begin function main
	.p2align	4, 0x90
	.type	main,@function
main:                                   # @main
	.cfi_startproc
# %bb.0:                                # %done
	xorl	%eax, %eax
	retq
.Lfunc_end0:
	.size	main, .Lfunc_end0-main
	.cfi_endproc
                                        # -- End function
	.type	msg,@object             # @msg
	.section	.rodata,"a",@progbits
	.globl	msg
msg:
	.ascii	"Hello World!\n"
	.size	msg, 13

	.type	.str_1,@object          # @.str_1
	.globl	.str_1
.str_1:
	.ascii	"stdin"
	.size	.str_1, 5

	.type	.str_2,@object          # @.str_2
	.globl	.str_2
.str_2:
	.ascii	"stdout"
	.size	.str_2, 6

	.type	.str_3,@object          # @.str_3
	.globl	.str_3
.str_3:
	.ascii	"stderr"
	.size	.str_3, 6

	.type	.str_4,@object          # @.str_4
	.globl	.str_4
.str_4:
	.ascii	"fread"
	.size	.str_4, 5

	.type	.str_5,@object          # @.str_5
	.globl	.str_5
.str_5:
	.ascii	"fwrite"
	.size	.str_5, 6

	.type	.str_6,@object          # @.str_6
	.globl	.str_6
.str_6:
	.ascii	"ferror"
	.size	.str_6, 6

	.type	.str_7,@object          # @.str_7
	.globl	.str_7
.str_7:
	.ascii	"exit"
	.size	.str_7, 4


	.section	".note.GNU-stack","",@progbits
