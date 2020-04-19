#pragma once
#include <execinfo.h>
#include <limits.h>
#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>
#if CHAR_BIT != 8
#error unsupported 'CHAR_BIT', need 8
#endif


// macro names prefixed with '·' instead of all upper-case (I abhor SCREAM_CODE!)
// exceptions: param-less atomic-expression macros


#define ·SliceOf(T)                                                                                                                            \
    struct {                                                                                                                                   \
        T* at;                                                                                                                                 \
        Uint len;                                                                                                                              \
    }

#define ·Maybe(T)                                                                                                                              \
    struct {                                                                                                                                   \
        T it;                                                                                                                                  \
        Bool ok;                                                                                                                               \
    }

#define ·Tup2(T0, T1)                                                                                                                          \
    struct {                                                                                                                                   \
        T0 _0;                                                                                                                                 \
        T1 _1;                                                                                                                                 \
    }

#define ·Tup3(T0, T1, T2)                                                                                                                      \
    struct {                                                                                                                                   \
        T0 _0;                                                                                                                                 \
        T1 _1;                                                                                                                                 \
        T2 _2;                                                                                                                                 \
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
typedef size_t Uint;
typedef void* PtrAny;
typedef const char* String;

typedef ·Maybe(Int) ºInt;
typedef ·Maybe(Uint) ºUint;
typedef ·Maybe(U64) ºU64;
typedef ·SliceOf(Uint) Uints;
typedef ·SliceOf(U8) U8s;
typedef U8s Str;
typedef ·SliceOf(Str) Strs;



#define null NULL

#define ·nameOf(ident) (#ident)

#define ·slice(TSlice__, ¹the_slice_to_reslice__, idx_start_reslice_from__, ¹idx_end_to_reslice_until__)                                       \
    ((TSlice__##s) {.len = (¹idx_end_to_reslice_until__) - (idx_start_reslice_from__),                                                         \
                    .at = &((¹the_slice_to_reslice__).at[idx_start_reslice_from__])})

#define ·append(the_slice_to_append_to__, ¹the_item_to_append__)                                                                               \
    do {                                                                                                                                       \
        (the_slice_to_append_to__).at[(the_slice_to_append_to__).len] = (¹the_item_to_append__);                                               \
        (the_slice_to_append_to__).len += 1;                                                                                                   \
    } while (0)

#define ·forEach(TItem, iteree_ident__, the_slice_to_iter__, ¹do_block__)                                                                      \
    do {                                                                                                                                       \
        for (Uint iˇ##iteree_ident__ = 0; iˇ##iteree_ident__ < (the_slice_to_iter__).len; iˇ##iteree_ident__ += 1) {                           \
            TItem* const iteree_ident__ = &((the_slice_to_iter__).at[iˇ##iteree_ident__]);                                                     \
            ¹do_block__                                                                                                                       \
        }                                                                                                                                      \
    } while (0)

#define ·ok(T, ¹the_value__) ((º##T) {.ok = true, .it = (¹the_value__)})

#define ·none(T) ((º##T) {.ok = false})




void fail(Str const str) {
    PtrAny callstack[16];
    Uint const n_frames = backtrace(callstack, 16);
    backtrace_symbols_fd(callstack, n_frames, 2); // 2 being stderr

    fwrite(str.at, 1, str.len, stderr);
    fwrite("\n", 1, 1, stderr);
    exit(1);
}

Str str(String const);
Str uintToStr(Uint const, Uint const);
Str str2(Str const, Str const);
void failIf(int err_code) {
    if (err_code)
        fail(str2(str("error code: "), uintToStr(err_code, 10)));
}

void assert(Bool const pred) {
#ifdef DEBUG
    if (!pred)
        fail(str("assertion failure"));
#endif
}




// pre-allocated fixed-size "heap"
#define mem_max (1 * 1024 * 1024) // would `const` but triggers -Wgnu-folding-constant
U8 mem_buf[mem_max];
Uint mem_pos = 0;

#define ·make(T, initial_len__, max_capacity__)                                                                                                \
    ((T##s) {.len = (initial_len__),                                                                                                           \
             .at = (T*)(memAlloc((((max_capacity__) < (initial_len__)) ? (initial_len__) : (max_capacity__)) * (sizeof(T))))})

U8* memAlloc(Uint const num_bytes) {
    Uint const new_pos = mem_pos + num_bytes;
    if (new_pos >= mem_max - 1)
        fail(str("out of memory: increase mem_max!"));
    U8* const mem_ptr = &mem_buf[mem_pos];
    mem_pos = new_pos;
    return mem_ptr;
}

Str newStr(Uint const initial_len, Uint const max_capacity) {
    Str ret_str = (Str) {.len = initial_len, .at = memAlloc(max_capacity)};
    return ret_str;
}

ºU64 uintParse(Str const str) {
    assert(str.len > 0);
    U64 ret_uint = 0;
    U64 mult = 1;
    for (Uint i = str.len; i > 0;) {
        i -= 1;
        if (str.at[i] < '0' || str.at[i] > '9')
            return ·none(U64);
        ret_uint += mult * (str.at[i] - 48);
        mult *= 10;
    }
    return ·ok(U64, ret_uint);
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

// for immediate consumption! not for keeping around
String strZ(Str const str) {
    if (str.at[str.len] == 0)
        return (String)str.at;
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
    ·forEach(Str, str, strs, { str_len += str->len; });

    Str ret_str = newStr(0, 1 + str_len);
    ret_str.at[str_len] = 0;
    ·forEach(Str, str, strs, {
        for (Uint i = 0; i < str->len; i += 1)
            ret_str.at[i + ret_str.len] = str->at[i];
        ret_str.len += str->len;
    });

    return ret_str;
}

Str str1(Str const s1) {
    return strConcat((Strs) {.len = 1, .at = ((Str[]) {s1})});
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
