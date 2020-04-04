target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64-unknown-linux-gnu"

@stdin = external global i8*
@stdout = external global i8*
declare i64 @fread(i8*, i64, i64, i8*)
declare i64 @fwrite(i8*, i64, i64, i8*)
declare i32 @ferror(i8*)

@str.1 = constant [26 x i8] c"I'll echo what you enter: "

define i32 @main() {
output_prompt:
  %stdout = load i8*, i8** @stdout
  %str.1 = getelementptr [26 x i8], [26 x i8]* @str.1, i64 0, i64 0
  %n_out_prompt = call i64 @fwrite(i8* %str.1, i64 1, i64 26, i8* %stdout)
  %n_out_prompt.eq.26 = icmp eq i64 %n_out_prompt, 26
  switch i1 %n_out_prompt.eq.26, label %ret_err [ i1 1, label %read_input ]
read_input:
  %stdin = load i8*, i8** @stdin
  %buf = alloca i8, i32 512
  %n_input_len = call i64 @fread(i8* %buf, i64 1, i64 512, i8* %stdin)
  %err_input = call i32 @ferror(i8* %stdin)
  switch i32 %err_input, label %ret_err [ i32 0, label %output_result ]
output_result:
  %n_out_echo = call i64 @fwrite(i8* %buf, i64 1, i64 %n_input_len, i8* %stdout)
  %n_out_echo.eq.n_input = icmp eq i64 %n_out_echo, %n_input_len
  switch i1 %n_out_echo.eq.n_input, label %ret_err [ i1 1, label %ret_ok ]
ret_err:
  br label %return
ret_ok:
  br label %return
return:
  %ret = phi i32 [ 0, %ret_ok ] , [ 1, %ret_err ]
  ret i32 %ret
}
