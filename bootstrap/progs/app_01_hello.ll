
target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64-unknown-linux-gnu"

@stdin = external global i8*
@stdout = external global i8*
@stderr = external global i8*
@msg = constant [13 x i8] c"Hello World!\0A"
@.str_1 = constant [5 x i8] c"stdin"
@.str_2 = constant [6 x i8] c"stdout"
@.str_3 = constant [6 x i8] c"stderr"
@.str_4 = constant [5 x i8] c"fread"
@.str_5 = constant [6 x i8] c"fwrite"
@.str_6 = constant [6 x i8] c"ferror"
@.str_7 = constant [4 x i8] c"exit"

declare i64 @fread(i8*, i64, i64, i8*)
declare i64 @fwrite(i8*, i64, i64, i8*)
declare i32 @ferror(i8*)
declare void @exit(i16)

define i32 @main() {
print_msg:
ret_err:
ret_ok:
done:
}

