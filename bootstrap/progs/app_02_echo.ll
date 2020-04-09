
target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64-unknown-linux-gnu"

@stdin = external global i8*
@stdout = external global i8*
@stderr = external global i8*


define i16 @writeTo(i8* %str_ptr, i64 %str_len, i8* %out_file) {
begin:
  %n = call i64 @fwrite(i8* %str_ptr, i64 1, i64 %str_len, i8* %out_file)
  %ok = icmp eq i64 %n, %str_len
  br i1 %ok, label %end, label %err_case
err_case:
  %err_code = call i16 @ferror(i8* %out_file)
  br label %end
end:
  %ret_val = phi i16 [0, %begin], [%err_code, %err_case]
  ret i16 %ret_val
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
b.1:
  call void @writeToStd(i8* %str_ptr, i64 %str_len, i8** @stderr)
  ret void
}


define void @writeOut(i8* %str_ptr, i64 %str_len) {
b.2:
  call void @writeToStd(i8* %str_ptr, i64 %str_len, i8** @stdout)
  ret void
}


define i64 @readFrom(i8* %buf_ptr, i64 %buf_len, i8* %in_file) {
b.3:
  %n = call i64 @fread(i8* %buf_ptr, i64 1, i64 %buf_len, i8* %in_file)
  ret i64 %n
}


define i64 @readInOrDie(i8* %buf_ptr, i64 %buf_len) {
begin:
  %in_file = load i8*, i8** @stdin
  %n = call i64 @readFrom(i8* %buf_ptr, i64 %buf_len, i8* %in_file)
  %err = call i16 @ferror(i8* %in_file)
  %ok = icmp eq i16 %err, 0
  br i1 %ok, label %end, label %die
die:
  call void @exit(i16 1)
  unreachable
end:
  %n_ok = phi i64 [%n, %begin]
  ret i64 %n_ok
}

declare i64 @fread(i8*, i64, i64, i8*)
declare i64 @fwrite(i8*, i64, i64, i8*)
declare i16 @ferror(i8*)
declare void @exit(i16)

define i32 @main() {
b.4:
  %buf = alloca i8, i32 1024
  %n = call i64 @readInOrDie(i8* %buf, i64 1024)
  call void @writeOut(i8* %buf, i64 %n)
  ret i32 0
}

