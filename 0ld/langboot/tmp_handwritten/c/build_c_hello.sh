#!/bin/sh

clear

# C to LLVM:

clang -O0 -emit-llvm hello.c -S -o hello0.ll
clang -O3 -emit-llvm hello.c -S -o hello3.ll
clang -O0 -emit-llvm hello.c -c -o hello0.bc
clang -O3 -emit-llvm hello.c -c -o hello3.bc

# LLVM to ASM to executable binary:

llc --relocation-model=pic hello0.bc -o hello0.s
llc --relocation-model=pic hello3.bc -o hello3.s

gcc hello0.s -o ~/.local/bin/hello0.exe
gcc hello3.s -o ~/.local/bin/hello3.exe
