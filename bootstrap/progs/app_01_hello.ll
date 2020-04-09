
target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64-unknown-linux-gnu"

@stdin = external global i8*
@stdout = external global i8*
@stderr = external global i8*
@msg = constant [11 x i8] c"Hola Welt.\0A"

declare i64 @fread(i8*, i64, i64, i8*)
declare i64 @fwrite(i8*, i64, i64, i8*)
declare i16 @ferror(i8*)
declare void @exit(i16)

define i16 @writeTo(i8* %str_ptr, i64 %str_len, i8* %out_file) {
b.1:
  call i64 @fwrite(i8* %str_ptr, i64 1, i64 %str_len, i8* %out_file)
  %err = call i16 @ferror(i8* %out_file)
  ret i16 %err
}


define void @writeToStd(i8* %str_ptr, i64 %str_len, i8** %std_file) {
begin:
  %out_file = load i8*, i8** %std_file
  %err = call i16 @writeTo(i8* %str_ptr, i64 %str_len, i8* %out_file)
  switch i16 %err, label %end [i16 1, label %exit_on_err]
exit_on_err:
  call void @exit(i16 1)
  unreachable
end:
  ret void
}


define void @writeErr(i8* %str_ptr, i64 %str_len) {
b.2:
  call void @writeToStd(i8* %str_ptr, i64 %str_len, i8** @stderr)
  ret void
}


define void @writeOut(i8* %str_ptr, i64 %str_len) {
b.3:
  call void @writeToStd(i8* %str_ptr, i64 %str_len, i8** @stdout)
  ret void
}


define i32 @main() {
b.4:
  %msg = getelementptr [11 x i8], [11 x i8]* @msg, i64 0, i64 0
  call void @writeOut(i8* %msg, i64 11)
  ret i32 0
}

