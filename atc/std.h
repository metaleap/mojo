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



#define SliceOf(T)                                                                                                                             \
    struct {                                                                                                                                   \
        T *at;                                                                                                                                 \
        Uint len;                                                                                                                              \
    }

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
typedef void *Ptr;
typedef SliceOf(U8) Str;
typedef const char *String;



#define nameOf(ident) #ident

#define slice(T, the_slice, idx_start, idx_end) ((T##s) {.len = idx_end - idx_start, .at = the_slice.at + (idx_start * sizeof(T))})



void panic(String format, ...) {
    Ptr callstack[16];
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

void assert(Bool pred, String msg) {
    if (!pred)
        panic("%s", msg);
}



// pre-allocated fixed-size "heap"
#define mem_max (2 * 1024 * 1024)
U8 mem_buf[mem_max];
Uint mem_pos = 0;

#define alloc(T, num_items) ((T##s) {.len = num_items, .at = (T *)(memAlloc(num_items * (sizeof(T))))})

U8 *memAlloc(Uint const num_bytes) {
    Uint const new_pos = mem_pos + num_bytes;
    if (new_pos >= mem_max)
        panic("out of memory: increase mem_max!\n");
    U8 *mem_ptr = &mem_buf[mem_pos];
    mem_pos = new_pos;
    return mem_ptr;
}

Str newStr(Uint str_len) {
    return (Str) {.len = str_len, .at = memAlloc(str_len)};
}

Str str(String from) {
    Uint str_len = 0;
    for (Uint i = 0; from[i] != 0; i += 1)
        str_len += 1;
    return (Str) {.len = str_len, .at = (U8 *)from};
}

Bool strEql(Str one, Str two) {
    if (one.len == two.len) {
        for (Uint i = 0; i < one.len; i += 1)
            if (one.at[i] != two.at[i])
                return false;
        return true;
    }
    return false;
}

Bool strEq(Str one, String two) {
    return strEql(one, str(two));
}

Str strSub(Str str, Uint idx_start, Uint idx_end) {
    return (Str) {.len = idx_end - idx_start, .at = str.at + idx_start};
}

String strZ(Str str) {
    U8 *buf = memAlloc(1 + str.len);
    buf[str.len] = 0;
    for (Uint i = 0; i < str.len; i++)
        buf[i] = str.at[i];
    return (String)buf;
}

Uint uintParse(Str str) {
    assert(str.len > 0, "empty string passed to uintParse");
    Uint ret_uint = 0;
    Uint mult = 1;
    for (Uint i = str.len; i > 0;) {
        i -= 1;
        if (str.at[i] < '0' || str.at[i] > '9')
            panic("bad Uint literal: %s\n", strZ(str));
        ret_uint += mult * (str.at[i] - 48);
        mult *= 10;
    }
    return ret_uint;
}

Str uintToStr(Uint uint_value, Uint base) {
    Uint num_digits = 1;
    Uint n = uint_value;
    while (n >= base) {
        num_digits += 1;
        n /= base;
    }
    n = uint_value;

    Str ret_str = newStr(num_digits);
    for (Uint i = ret_str.len; i > 0;) {
        i -= 1;
        if (n < base) {
            ret_str.at[i] = 48 + n;
            break;
        } else {
            ret_str.at[i] = 48 + (n % base);
            n /= base;
        }
        if (base > 10 && ret_str.at[i] > '9')
            ret_str.at[i] += 7;
    }
    return ret_str;
}

Bool strHasChar(String s, U8 c) {
    for (Uint i = 0; s[i] != 0; i += 1)
        if (s[i] == c)
            return true;
    return false;
}
