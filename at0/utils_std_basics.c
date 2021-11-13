#pragma once
#include <execinfo.h>
#include <limits.h>
#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>


// #pragma clang diagnostic ignored "-Wlanguage-extension-token"
// #define ·dbgBrk() asm volatile("int3"); // gdb myprog -ex run

#define ·SliceOf(T)                                                                    \
    struct {                                                                           \
        T* at;                                                                         \
        UInt len;                                                                      \
    }

#define ·SliceOfPtrs(T)                                                                \
    struct {                                                                           \
        T** at;                                                                        \
        UInt len;                                                                      \
    }

#define ·ListOf(T)                                                                     \
    struct {                                                                           \
        T* at;                                                                         \
        UInt len;                                                                      \
        UInt cap;                                                                      \
    }

#define ·ListOfPtrs(T)                                                                 \
    struct {                                                                           \
        T** at;                                                                        \
        UInt len;                                                                      \
        UInt cap;                                                                      \
    }

#define ·Maybe(T)                                                                      \
    struct {                                                                           \
        T it;                                                                          \
        Bool got;                                                                      \
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
typedef size_t UInt;
typedef void* Any;
typedef char* CStr;

typedef ·Maybe(Bool) ºBool;
typedef ·Maybe(UInt) ºUInt;
typedef ·Maybe(U64) ºU64;
typedef ·Maybe(I64) ºI64;
typedef ·SliceOf(Bool) Bools;
typedef ·SliceOf(UInt) UInts;
typedef ·SliceOf(U8) U8s;
typedef U8s Str;
typedef ·Maybe(Str) ºStr;
typedef ·SliceOf(Str) Strs;


#define ·len0(T) ((T##s) {.len = 0, .at = NULL})

#define ·len1(T, ptr_to_the_value_at_0__)                                              \
    ((T##s) {.len = 1, .at = ptr_to_the_value_at_0__})

#define ·slice(TSlice__, ¹the_slice_to_reslice__, ²idx_start_reslice_from__,           \
               ¹idx_end_to_reslice_until__)                                            \
    ((TSlice__##s) {.len =                                                             \
                        (¹idx_end_to_reslice_until__) - (²idx_start_reslice_from__),   \
                    .at = &((¹the_slice_to_reslice__).at[²idx_start_reslice_from__])})

#define ·last(²the_slice__) (&(²the_slice__.at[²the_slice__.len - 1]))

#define ·push(³the_slice_to_append_to__, ¹the_item_to_append__)                        \
    do {                                                                               \
        (³the_slice_to_append_to__).at[(³the_slice_to_append_to__).len] =              \
            (¹the_item_to_append__);                                                   \
        (³the_slice_to_append_to__).len += 1;                                          \
    } while (0)

#define ·append(⁵the_list_to_append_to__, ¹the_item_to_append__)                       \
    do {                                                                               \
        ·assert((⁵the_list_to_append_to__).len < (⁵the_list_to_append_to__).cap);      \
        (⁵the_list_to_append_to__).at[(⁵the_list_to_append_to__).len] =                \
            (¹the_item_to_append__);                                                   \
        (⁵the_list_to_append_to__).len += 1;                                           \
    } while (0)

#define ·forEach(TItem, the_ident__, ²the_slice_to_iter__, ¹do_block__)                \
    do {                                                                               \
        for (UInt iˇ##the_ident__ = 0; iˇ##the_ident__ < (²the_slice_to_iter__).len;   \
             iˇ##the_ident__ += 1) {                                                   \
            TItem* const the_ident__ = &((²the_slice_to_iter__).at[iˇ##the_ident__]);  \
            { ¹do_block__ }                                                            \
        }                                                                              \
    } while (0)

#define ·got(T, ¹the_value__) ((º##T) {.got = true, .it = (¹the_value__)})

#define ·none(T) ((º##T) {.got = false})

#define ·as(T, ¹the_expr__) ((T*)(¹the_expr__))

#define ·fYay(¹the_enum__, ¹the_flag__) (¹the_enum__) | (¹the_flag__)

#define ·fNay(¹the_enum__, ¹the_flag__) (¹the_enum__) & (~(¹the_flag__))

#define ·fHas(¹the_enum__, ²the_flag__)                                                \
    (((¹the_enum__) & (²the_flag__)) == (²the_flag__))

#define ·flag(²the_enum__, ²the_flag__, ¹the_bool__)                                   \
    ((¹the_bool__) ? ((²the_enum__) | (²the_flag__))                                   \
                   : ((²the_enum__) & (~(²the_flag__))))

#define ·swap(T, the_ptr_lhs, the_ptr_rhs)                                             \
    do {                                                                               \
        T* tmp_rhs = the_ptr_rhs;                                                      \
        the_ptr_rhs = the_ptr_lhs;                                                     \
        the_ptr_lhs = tmp_rhs;                                                         \
    } while (0)

#define ·fail(¹the_msg)                                                                \
    do {                                                                               \
        fprintf(stderr,                                                                \
                "\n——————————————————————————————————————————\npanicked at: %s:%d\n",  \
                __FILE__, __LINE__);                                                   \
        abortWithBacktraceAndMsg(¹the_msg);                                            \
    } while (0)

#ifdef NDEBUG
#define ·assert(¹the_predicate)
#else
#define ·assert(¹the_predicate)                                                                                                                                         \
    do {                                                                                                                                                                \
        if (!(¹the_predicate)) {                                                                                                                                        \
            fprintf(stderr,                                                                                                                                             \
                    "\n——————————————————————————————————————————\ncondition `%s` " \
                    "violated in: %s:%d\n\n",                                                                                                                           \
                    #¹the_predicate, __FILE__, __LINE__);                                                                                                               \
            abortWithBacktraceAndMsg(·len0(U8));                                                                                                                        \
        }                                                                                                                                                               \
    } while (0)
#endif



void printStr(Str const str) {
    fwrite(&str.at[0], 1, str.len, stderr);
}
void writeStr(Str const str) {
    fwrite(&str.at[0], 1, str.len, stdout);
}

void abortWithBacktraceAndMsg(Str const msg) {
    Any callstack[16];
    UInt const n_frames = backtrace(callstack, 16);
    backtrace_symbols_fd(callstack, n_frames, 2); // 2 = stderr

    fwrite("\n", 1, 1, stderr);
    printStr(msg);
    fwrite("\n", 1, 1, stderr);
    fflush(stderr);
    abort();
}

Str strL(Any const c_str, UInt str_len) {
    if (c_str == NULL)
        return ·len0(U8);
    if (str_len == 0)
        for (UInt i = 0; ((U8*)c_str)[i] != 0; i += 1)
            str_len += 1;
    return (Str) {.len = str_len, .at = c_str};
}

Str str(Any const c_str) {
    return strL(c_str, 0);
}

ºU64 uInt64Parse(Str const str) {
    ·assert(str.len > 0);
    U64 ret_uInt = 0;
    U64 mult = 1;
    for (UInt i = str.len; i > 0;) {
        i -= 1;
        if (str.at[i] < '0' || str.at[i] > '9')
            return ·none(U64);
        ret_uInt += mult * (str.at[i] - 48);
        mult *= 10;
    }
    return ·got(U64, ret_uInt);
}

UInt uIntMinSize(UInt const max_incl, UInt const bits) {
    UInt r = 1;
    for (UInt m = max_incl >> bits; m > 0; m >>= bits)
        r += 1;
    return r;
}

Bool strEql(Str const one, Str const two) {
    if (one.at == NULL || two.at == NULL)
        return (one.at == two.at);
    if (one.len != two.len)
        return false;
    if (one.len > 0 && one.at != two.at) {
        UInt i_middle = 1 + (one.len / 2);
        for (UInt i = 0, j = one.len - 1; i < i_middle; i += 1, j -= 1)
            if (one.at[i] != two.at[i] || one.at[j] != two.at[j])
                return false;
    }
    return true;
}

Bool strEq(Any const one, Str const two, UInt const str_len) {
    return strEql(strL(one, str_len), two);
}

Bool strIn(Str const needle, Strs const haystack) {
    for (UInt i = 0; i < haystack.len; i += 1)
        if (strEql(needle, haystack.at[i]))
            return true;
    return false;
}

Str strPrefSuff(Str const str, Str const prefix) {
    Str ret_str = ·len0(U8);
    if (str.len >= prefix.len)
        if (strEql(prefix, ·slice(U8, str, 0, prefix.len)))
            return ·slice(U8, str, prefix.len, str.len);
    return ret_str;
}

Str strSub(Str const str, UInt const idx_start, UInt const idx_end) {
    return (Str) {.len = idx_end - idx_start, .at = str.at + idx_start};
}

Bool cStrHasChar(CStr const s, U8 const c) {
    for (UInt i = 0; s[i] != 0; i += 1)
        if (s[i] == c)
            return true;
    return false;
}

Bool strPref(Str const haystack, Str const prefix) {
    if (haystack.len >= prefix.len && prefix.len > 0)
        return (prefix.len == 1) ? (haystack.at[0] == prefix.at[0])
                                 : strEql(prefix, strSub(haystack, 0, prefix.len));
    return false;
}

Bool strSuff(Str const haystack, Str const suffix) {
    if (haystack.len >= suffix.len && suffix.len > 0)
        return (suffix.len == 1)
                   ? (haystack.at[haystack.len - 1] == suffix.at[0])
                   : strEql(suffix,
                            strSub(haystack, haystack.len - suffix.len, haystack.len));
    return false;
}

ºUInt strIndexOf(Str const haystack, Str const needle) {
    if (haystack.at != NULL && haystack.len != 0 && needle.len != 0
        && haystack.len >= needle.len)
        for (UInt i = 0; i < 1 + (haystack.len - needle.len); i += 1) {
            Str const h_sub = ·slice(U8, haystack, i, i + needle.len);
            if (strEql(needle, h_sub))
                return ·got(UInt, i);
        }
    return ·none(UInt);
}

Str strPush(Str buf, CStr const str_z) {
    UInt len;
    for (len = 0; str_z != NULL && str_z[len] != 0; len += 1)
        buf.at[buf.len + len] = str_z[len];
    buf.len += len;
    buf.at[buf.len] = 0;
    return buf;
}

Str strCopyTo(Str buf, Str const str) {
    for (UInt i = 0; i < str.len; i += 1)
        buf.at[i + buf.len] = str.at[i];
    buf.len += str.len;
    return buf;
}

UInt uIntMin(UInt const uint1, UInt const uint2) {
    return (uint1 > uint2) ? uint2 : uint1;
}

UInt uIntMax(UInt const uint1, UInt const uint2) {
    return (uint1 < uint2) ? uint2 : uint1;
}
