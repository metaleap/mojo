
target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64-unknown-linux-gnu"

@stdout = external global i8*
@stdin = external global i8*
@intro_msg = constant [53 x i8] c"I'll echo what you enter, up until EOF aka Ctrl+D...\0A"

declare i64 @fread(i8*, i64, i64, i8*)
declare i64 @fwrite(i8*, i64, i64, i8*)
declare i32 @ferror(i8*)


define i32 @main() {
.c_ok.c_maybe.c:
  %.c_ok.c_maybe.c_n_in_read_buf = alloca i8, i32 512
  %.c_ok.c_maybe.c_n_in_read.a3 = load i8*, i8** @stdin
  %.c_ok.c_maybe.c_n_in_read = call i64 @fread(i8* %.c_ok.c_maybe.c_n_in_read_buf,
    i64 1,
    i64 512,
    i8* %.c_ok.c_maybe.c_n_in_read.a3)
  %.c_ok.c_maybe.c.scrut = icmp eq i64 %.c_ok.c_maybe.c_n_in_read, 0
  switch i1 %.c_ok.c_maybe.c.scrut, label %.c_ok.c_maybe.c.1 [i1 0, label %.c_ok.c_maybe.c.0]
.c_ok.c_maybe.c.1:
  br label %.c_ok.c.1
.c_ok.c_maybe.c.0:
  %.c_ok.c_maybe.c_n_out_echo.a3 = load i8*, i8** @stdout
  %.c_ok.c_maybe.c_n_out_echo = call i64 @fwrite(i8* %.c_ok.c_maybe.c_n_in_read_buf,
    i64 1,
    i64 %.c_ok.c_maybe.c_n_in_read,
    i8* %.c_ok.c_maybe.c_n_out_echo.a3)
  %.c_ok.c_maybe.c.0.result = icmp eq i64 %.c_ok.c_maybe.c_n_out_echo, %.c_ok.c_maybe.c_n_in_read
  br label %.c_ok.c.1
.c_ok.c:
  %.c_ok.c_n_out_intro.a0 = getelementptr [53 x i8], [53 x i8]* @intro_msg, i64 0, i64 0
  %.c_ok.c_n_out_intro.a3 = load i8*, i8** @stdout
  %.c_ok.c_n_out_intro = call i64 @fwrite(i8* %.c_ok.c_n_out_intro.a0,
    i64 1,
    i64 53,
    i8* %.c_ok.c_n_out_intro.a3)
  %.c_ok.c.scrut = icmp ne i64 %.c_ok.c_n_out_intro, 53
  switch i1 %.c_ok.c.scrut, label %.c_ok.c.1 [i1 1, label %.c_ok.c.0]
.c_ok.c.1:
  %.c_ok.c_maybe = phi i1 [%.c_ok.c_maybe.c.0.result, %.c_ok.c_maybe.c.0], [1, %.c_ok.c_maybe.c.1]
  br label %.c
.c_ok.c.0:
  br label %.c
.c:
  %.c_ok = phi i1 [0, %.c_ok.c.0], [%.c_ok.c_maybe, %.c_ok.c.1]
  switch i1 %.c_ok, label %.c.1 [i1 1, label %.c.0]
.c.1:
  br label %.return
.c.0:
  br label %.return
.return:
  %ret = phi i32 [0, %.c.0], [1, %.c.1]
  ret i32 %ret
}

