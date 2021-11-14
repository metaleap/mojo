@.str = constant [12 x i8] c"hello\0Aworld\00"
declare i32 @puts(i8*)

define i32 @main() {
  %ptr = getelementptr [12 x i8], [12 x i8]* @.str, i64 0, i64 0
  call i32 @puts(i8* %ptr)
  ret i32 0
}
