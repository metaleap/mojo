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

#define ·slice(TSlice__, ¹the_slice_to_reslice__, ²idx_start_reslice_from__, ¹idx_end_to_reslice_until__)                                      \
    ((TSlice__##s) {.len = (¹idx_end_to_reslice_until__) - (²idx_start_reslice_from__),                                                        \
                    .at = &((¹the_slice_to_reslice__).at[²idx_start_reslice_from__])})

#define ·append(³the_slice_to_append_to__, ¹the_item_to_append__)                                                                              \
    do {                                                                                                                                       \
        (³the_slice_to_append_to__).at[(³the_slice_to_append_to__).len] = (¹the_item_to_append__);                                             \
        (³the_slice_to_append_to__).len += 1;                                                                                                  \
    } while (0)

#define ·forEach(TItem, iteree_ident__, ²the_slice_to_iter__, ¹do_block__)                                                                     \
    do {                                                                                                                                       \
        for (Uint iˇ##iteree_ident__ = 0; iˇ##iteree_ident__ < (²the_slice_to_iter__).len; iˇ##iteree_ident__ += 1) {                          \
            TItem* const iteree_ident__ = &((²the_slice_to_iter__).at[iˇ##iteree_ident__]);                                                    \
            ¹do_block__                                                                                                                       \
        }                                                                                                                                      \
    } while (0)

#define ·ok(T, ¹the_value__) ((º##T) {.ok = true, .it = (¹the_value__)})

#define ·none(T) ((º##T) {.ok = false})


#define ·fail(¹the_msg)                                                                                                                        \
    do {                                                                                                                                       \
        fprintf(stderr, "\npanicked at: %s:%d\n", __FILE__, __LINE__);                                                                         \
        fail(¹the_msg);                                                                                                                        \
    } while (0)


#if DEBUG
#define ·assert(¹the_predicate)                                                                                                                \
    do {                                                                                                                                       \
        if (!(¹the_predicate)) {                                                                                                               \
            fprintf(stderr, "\n>>>>>>>>>>>>>>>>>>>>>>\n\nassert violation `%s` triggered in: %s:%d\n\n", #¹the_predicate, __FILE__, __LINE__); \
            exit(1);                                                                                                                           \
        }                                                                                                                                      \
    } while (0)
#else
#define ·assert(¹the_predicate)
#endif




void printStr(Str const str) {
    fwrite(&str.at[0], 1, str.len, stderr);
}
void writeStr(Str const str) {
    fwrite(&str.at[0], 1, str.len, stdout);
}
void fail(Str const str) {
    PtrAny callstack[16];
    Uint const n_frames = backtrace(callstack, 16);
    backtrace_symbols_fd(callstack, n_frames, 2); // 2 being stderr

    printStr(str);
    fwrite("\n", 1, 1, stderr);
    exit(1);
}

Str str(String const);
Str uintToStr(Uint const, Uint const, Uint const);
Str str2(Str const, Str const);
void failIf(int err_code) {
    if (err_code)
        ·fail(str2(str("error code: "), uintToStr(err_code, 1, 10)));
}



// pre-allocated fixed-size "heap"
#define mem_max (1 * 1024 * 1024) // would `const` but triggers -Wgnu-folding-constant
U8 mem_buf[mem_max];
Uint mem_pos = 0;

#define ·make(T, ³initial_len__, ²max_capacity__)                                                                                              \
    ((T##s) {.len = (³initial_len__),                                                                                                          \
             .at = (T*)(memAlloc((((²max_capacity__) < (³initial_len__)) ? (³initial_len__) : (²max_capacity__)) * (sizeof(T))))})

U8* memAlloc(Uint const num_bytes) {
    Uint const new_pos = mem_pos + num_bytes;
    if (new_pos >= mem_max - 1)
        ·fail(str("out of memory: increase mem_max!"));
    U8* const mem_ptr = &mem_buf[mem_pos];
    mem_pos = new_pos;
    return mem_ptr;
}

Str newStr(Uint const initial_len, Uint const max_capacity) {
    Str ret_str = (Str) {.len = initial_len, .at = memAlloc(max_capacity)};
    return ret_str;
}

ºU64 uintParse(Str const str) {
    ·assert(str.len > 0);
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

Str uintToStr(Uint const uint_value, Uint const str_min_len, Uint const base) {
    Uint num_digits = 1;
    Uint n = uint_value;
    while (n >= base) {
        num_digits += 1;
        n /= base;
    }
    n = uint_value;

    Uint const str_len = (num_digits > str_min_len) ? num_digits : str_min_len;
    Str const ret_str = newStr(str_len, str_len + 1);
    ret_str.at[str_len] = 0;
    for (Uint i = 0; i < str_len - num_digits; i += 1)
        ret_str.at[i] = '0';

    Bool done = false;
    for (Uint i = ret_str.len; i > 0 && !done;) {
        i -= 1;
        if (n < base) {
            ret_str.at[i] = 48 + n;
            done = true;
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

// unused in principle, but kept around just in case we really do need a printf
// occasionally while debugging. result for immediate consumption! not for keeping.
String strZ(Str const str) {
    if (str.at[str.len] == 0)
        return (String)str.at;
    U8* buf = memAlloc(1 + str.len);
    buf[str.len] = 0;
    for (Uint i = 0; i < str.len; i += 1)
        buf[i] = str.at[i];
    return (String)buf;
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

Str strQuot(Str const str) {
    Str ret_str = newStr(1, 3 + (3 * str.len));
    ret_str.at[0] = '\"';
    for (Uint i = 0; i < str.len; i += 1) {
        U8 const chr = str.at[i];
        if (chr >= 32 && chr < 127) {
            ret_str.at[ret_str.len] = chr;
            ret_str.len += 1;
        } else {
            ret_str.at[ret_str.len] = '\\';
            const Str esc_num_str = uintToStr(chr, 3, 10);
            for (Uint c = 0; c < esc_num_str.len; c += 1)
                ret_str.at[1 + c + ret_str.len] = esc_num_str.at[c];
            ret_str.len += 1 + esc_num_str.len;
        }
    }
    ret_str.at[ret_str.len] = '\"';
    ret_str.len += 1;
    ret_str.at[ret_str.len] = 0;
    return ret_str;
}

Str strSub(Str const str, Uint const idx_start, Uint const idx_end) {
    return (Str) {.len = idx_end - idx_start, .at = str.at + idx_start};
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
