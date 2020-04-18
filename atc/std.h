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
        T* at;                                                                                                                                 \
        Uint len;                                                                                                                              \
    }

#define Maybe(T)                                                                                                                               \
    struct {                                                                                                                                   \
        T it;                                                                                                                                  \
        Bool ok;                                                                                                                               \
    }

typedef bool Bool;
typedef u_int8_t U8;
typedef u_int16_t U16;
typedef u_int32_t U32;
typedef u_int64_t U64;
typedef int8_t I8;
typedef int16_t I16;
typedef int32_t I32;
typedef int64_t I64;
typedef ssize_t Int;
typedef Maybe(Int) ˇInt;
typedef size_t Uint;
typedef Maybe(Uint) ºUint;
typedef SliceOf(Uint) Uints;
typedef void* Ptr;
typedef SliceOf(U8) Str;
typedef SliceOf(Str) Strs;
typedef const char* String;



#define nameOf(ident) #ident

#define slice(T, the_slice, idx_start, idx_end) ((T##s) {.len = idx_end - idx_start, .at = the_slice.at + (idx_start * sizeof(T))})

#define append(the_slice, item)                                                                                                                \
    do {                                                                                                                                       \
        the_slice.at[the_slice.len] = (item);                                                                                                  \
        the_slice.len += 1;                                                                                                                    \
    } while (0)

#define forEach(T, iteree_ident, the_slice, do_block)                                                                                          \
    for (Uint iteree_ident##ˇidx = 0; iteree_ident##ˇidx < the_slice.len; iteree_ident##ˇidx += 1) {                                           \
        T* const iteree_ident = &the_slice.at[iteree_ident##ˇidx];                                                                             \
        do_block                                                                                                                               \
    }

#define ok(T, the_it)                                                                                                                          \
    (º##T) {                                                                                                                                   \
        .ok = true, .it = the_it                                                                                                               \
    }

#define none(T)                                                                                                                                \
    (º##T) {                                                                                                                                   \
        .ok = false                                                                                                                            \
    }


void panic(String const format, ...) {
    Ptr callstack[16];
    Uint const n_frames = backtrace(callstack, 16);
    backtrace_symbols_fd(callstack, n_frames, 2); // 2 being stderr

    va_list args;
    va_start(args, format);
    vfprintf(stderr, format, args);
    va_end(args);
    fwrite("\n", 1, 1, stderr);

    exit(1);
}

void panicIf(int const err) {
    if (err)
        panic("error %d", err);
}

void assert(Bool const pred) {
#ifdef DEBUG
    if (!pred)
        panic("assertion failure");
#endif
}



// pre-allocated fixed-size "heap"
#define mem_max (2 * 1024 * 1024)
U8 mem_buf[mem_max];
Uint mem_pos = 0;

#define make(T, initial_len, max_capacity)                                                                                                     \
    ((T##s) {.len = initial_len, .at = (T*)(memAlloc(((max_capacity < initial_len) ? initial_len : max_capacity) * (sizeof(T))))})

U8* memAlloc(Uint const num_bytes) {
    Uint const new_pos = mem_pos + num_bytes;
    if (new_pos >= mem_max)
        panic("out of memory: increase mem_max!");
    U8* const mem_ptr = &mem_buf[mem_pos];
    mem_pos = new_pos;
    return mem_ptr;
}

Str newStr(Uint const str_len, Uint const str_cap) {
    Str ret_str = (Str) {.len = str_len, .at = memAlloc(str_cap)};
    return ret_str;
}

ºUint uintParse(Str const str) {
    assert(str.len > 0);
    Uint ret_uint = 0;
    Uint mult = 1;
    for (Uint i = str.len; i > 0;) {
        i -= 1;
        if (str.at[i] < '0' || str.at[i] > '9')
            return none(Uint);
        ret_uint += mult * (str.at[i] - 48);
        mult *= 10;
    }
    return ok(Uint, ret_uint);
}

Str uintToStr(Uint const uint_value, Uint const base) {
    Uint num_digits = 1;
    Uint n = uint_value;
    while (n >= base) {
        num_digits += 1;
        n /= base;
    }
    n = uint_value;

    Str const ret_str = newStr(num_digits, num_digits);
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

Str str(String const from) {
    Uint str_len = 0;
    for (Uint i = 0; from[i] != 0; i += 1)
        str_len += 1;
    return (Str) {.len = str_len, .at = (U8*)from};
}

Bool strEql(Str const one, Str const two) {
    if (one.len == two.len) {
        for (Uint i = 0; i < one.len; i += 1)
            if (one.at[i] != two.at[i])
                return false;
        return true;
    }
    return false;
}

Bool strEq(Str const one, String const two, ºUint const str_len) {
    Str const s2 = (!str_len.ok) ? str(two) : ((Str) {.len = str_len.it, .at = (U8*)two});
    return strEql(one, s2);
}

Str strSub(Str const str, Uint const idx_start, Uint const idx_end) {
    return (Str) {.len = idx_end - idx_start, .at = str.at + idx_start};
}

String strZ(Str const str) {
    U8* buf = memAlloc(1 + str.len);
    buf[str.len] = 0;
    for (Uint i = 0; i < str.len; i++)
        buf[i] = str.at[i];
    return (String)buf;
}

Bool strHasChar(String const s, U8 const c) {
    for (Uint i = 0; s[i] != 0; i += 1)
        if (s[i] == c)
            return true;
    return false;
}

Str strConcat(Strs const strs) {
    Uint str_len = 0;
    forEach(Str, str, strs, { str_len += str->len; });

    Str ret_str = newStr(0, 1 + str_len);
    ret_str.at[str_len] = 0;
    forEach(Str, str, strs, {
        for (Uint i = 0; i < str->len; i += 1)
            ret_str.at[i + ret_str.len] = str->at[i];
        ret_str.len += str->len;
    });

    return ret_str;
}

Str str2(Str const s1, Str const s2) {
    return strConcat((Strs) {.len = 2, .at = ((Str[]) {s1, s2})});
}

Str str3(Str const s1, Str const s2, Str const s3) {
    return strConcat((Strs) {.len = 3, .at = ((Str[]) {s1, s2, s3})});
}

Str str4(Str const s1, Str const s2, Str const s3, Str const s4) {
    return strConcat((Strs) {.len = 4, .at = ((Str[]) {s1, s2, s3, s4})});
}

Str str5(Str const s1, Str const s2, Str const s3, Str const s4, Str const s5) {
    return strConcat((Strs) {.len = 5, .at = ((Str[]) {s1, s2, s3, s4, s5})});
}
