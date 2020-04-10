
target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64-unknown-linux-gnu"

@stdin = external global i8*
@stdout = external global i8*
@stderr = external global i8*
@tok_kind_comment = constant [7 x i8] c"comment"
@tok_kind_ident = constant [5 x i8] c"ident"
@tok_kind_lit_int = constant [7 x i8] c"lit_int"
@tok_kind_lit_str = constant [7 x i8] c"lit_str"
@tok_kind_sep_bparen_open = constant [15 x i8] c"sep_bparen_open"
@tok_kind_sep_bparen_close = constant [16 x i8] c"sep_bparen_close"
@tok_kind_sep_bcurly_open = constant [15 x i8] c"sep_bcurly_open"
@tok_kind_sep_bcurly_close = constant [16 x i8] c"sep_bcurly_close"
@tok_kind_sep_bsquare_open = constant [16 x i8] c"sep_bsquare_open"
@tok_kind_sep_bsquare_close = constant [17 x i8] c"sep_bsquare_close"
@tok_kind_sep_comma = constant [9 x i8] c"sep_comma"


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
  %num_bytes_read = call i64 @fread(i8* %buf_ptr, i64 1, i64 %buf_len, i8* %in_file)
  ret i64 %num_bytes_read
}


define i64 @readInOrDie(i8* %buf_ptr, i64 %buf_len) {
begin:
  %in_file = load i8*, i8** @stdin
  %num_bytes_read = call i64 @readFrom(i8* %buf_ptr, i64 %buf_len, i8* %in_file)
  %err = call i16 @ferror(i8* %in_file)
  %ok = icmp eq i16 %err, 0
  br i1 %ok, label %end, label %die
die:
  call void @exit(i16 1)
  unreachable
end:
  %n_ok = phi i64 [%num_bytes_read, %begin]
  ret i64 %n_ok
}

declare i64 @fread(i8*, i64, i64, i8*)
declare i64 @fwrite(i8*, i64, i64, i8*)
declare i16 @ferror(i8*)
declare void @exit(i16)

define i8* @ptrIncr(i8* %ptr, i64 %incr_by_bytes) {
b.4:
  %ptr_as_int = ptrtoint i8* %ptr to i64
  %ptr_int_incr = add i64 %incr_by_bytes, %ptr_as_int
  %int_as_ptr = inttoptr i64 %ptr_int_incr to i8*
  ret i8* %int_as_ptr
}


define void @swapBytes(i8* %ptr_l, i8* %ptr_r) {
b.5:
  %byte_l = load i8, i8* %ptr_l
  %byte_r = load i8, i8* %ptr_r
  store i8 %byte_r, i8* %ptr_l
  store i8 %byte_l, i8* %ptr_r
  ret void
}


define i32 @main() {
b.6:
  %buf_in_src_file_size = mul i64 1048510, 1024
  %buf_in_src_file = alloca i8, i64 %buf_in_src_file_size
  %buf_in_src_file_len = call i64 @readInOrDie(i8* %buf_in_src_file, i64 %buf_in_src_file_size)
  ret i32 0
}

