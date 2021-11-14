@.str = constant [13 x i8] c"hello world\0A\00"
declare i32 @puts(i8*)

define i32 @main() {
  %ptr = getelementptr [13 x i8], [13 x i8]* @.str, i64 0, i64 0
  call i32 @puts(i8* %ptr)
  ret i32 0
}
