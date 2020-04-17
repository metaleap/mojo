#pragma once
#include <execinfo.h>
#include <limits.h>
#include <stdbool.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <sys/types.h>
#if CHAR_BIT != 8
#error unsupported 'CHAR_BIT', need 8
#endif



#define nameOf(ident) #ident

#define SliceOf(T)                                                                                \
    struct {                                                                                      \
        T *_;                                                                                     \
        Uint len;                                                                                 \
    }

#define slice(T, slice, start, num_items)                                                         \
    ((T##s) {.len = num_items, ._ = slice._ + (start * sizeof(T))})


typedef bool Bool;
typedef u_int8_t U8;
typedef u_int16_t U16;
typedef u_int32_t U32;
typedef u_int64_t U64;
typedef ssize_t Int;
typedef size_t Uint;
typedef int8_t I8;
typedef int16_t I16;
typedef int32_t I32;
typedef int64_t I64;
typedef void *Any;
typedef SliceOf(U8) Str;
typedef const char *String;



void panic(String format, ...) {
    Any callstack[16];
    Uint n_frames = backtrace(callstack, 16);
    backtrace_symbols_fd(callstack, n_frames, 2); // 2 being stderr

    va_list args;
    va_start(args, format);
    vfprintf(stderr, format, args);
    va_end(args);

    exit(1);
}

void panicIf(int err) {
    if (err)
        panic("error %d", err);
}

void assert(Bool pred) {
    panicIf(!pred);
}

void unreachable() {
    panic("unreachable");
}



// "heap" without syscalls
#define alloc(T, num_items)                                                                       \
    ((T##s) {.len = num_items, ._ = (T *)(memAlloc(num_items * (sizeof(T))))})

#define mem_max (2 * 1024 * 1024)
U8 mem_buf[mem_max];
Uint mem_pos = 0;

U8 *memAlloc(Uint const num_bytes) {
    Uint const new_pos = mem_pos + num_bytes;
    if (new_pos >= mem_max)
        panic("out of memory: increase mem_max!\n");
    U8 *mem_ptr = &mem_buf[mem_pos];
    mem_pos = new_pos;
    return mem_ptr;
}

Str newStr(Uint str_len) {
    return (Str) {.len = str_len, ._ = memAlloc(str_len)};
}

Str str(String from) {
    Uint str_len = 0;
    for (Uint i = 0; from[i] != 0; i += 1)
        str_len += 1;
    return (Str) {.len = str_len, ._ = (U8 *)from};
}

Bool strEql(Str one, Str two) {
    if (one.len == two.len) {
        for (Uint i = 0; i < one.len; i += 1)
            if (one._[i] != two._[i])
                return false;
        return true;
    }
    return false;
}

Bool strEq(Str one, String two) {
    return strEql(one, str(two));
}

Uint uintParse(Str str) {
    assert(str.len > 0);
    Uint ret_uint = 0;
    Uint mult = 1;
    for (Uint i = str.len; i > 0;) {
        i -= 1;
        if (str._[i] < '0' || str._[i] > '9')
            panic("bad Uint literal: %s\n", str);
        ret_uint += mult * (str._[i] - 48);
        mult *= 10;
    }
    return ret_uint;
}
