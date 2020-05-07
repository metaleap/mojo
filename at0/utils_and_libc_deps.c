#pragma once
#include "../../misc/metaleap.c"
#include <execinfo.h>
#include <limits.h>
#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>
#if CHAR_BIT != 8
#error unsupported 'CHAR_BIT', need 8
#endif

// macro names prefixed with '·' instead of all upper-case (avoids SCREAM_CODE)



UInt counter = 0;

struct Mem {
#define mem_max (1 * 1024 * 1024)
    U8 buf[mem_max];
    UInt pos;
} mem = {.pos = 0};




#define ·new(T) ((T*)memAlloc(sizeof(T)))

#define ·sliceOf(T, ³initial_len__, ²max_capacity__)                                                                                         \
    ((T##s) {.len = (³initial_len__),                                                                                                        \
             .at = (T*)(memAlloc((((²max_capacity__) < (³initial_len__)) ? (³initial_len__) : (²max_capacity__)) * (sizeof(T))))})

#define ·sliceOfPtrs(T, ³initial_len__, ²max_capacity__)                                                                                     \
    {                                                                                                                                        \
        .len = (³initial_len__),                                                                                                             \
        .at = (T**)(memAlloc((((²max_capacity__) < (³initial_len__)) ? (³initial_len__) : (²max_capacity__)) * (sizeof(T*))))                \
    }

#define ·listOf(T, ⁵initial_len__, ⁴max_capacity__)                                                                                          \
    ((T##s) {.len = (⁵initial_len__),                                                                                                        \
             .cap = (((⁴max_capacity__) < (⁵initial_len__)) ? (⁵initial_len__) : (⁴max_capacity__)),                                         \
             .at = (T*)(memAlloc((((⁴max_capacity__) < (⁵initial_len__)) ? (⁵initial_len__) : (⁴max_capacity__)) * (sizeof(T))))})

#define ·listOfPtrs(T, ⁵initial_len__, ⁴max_capacity__)                                                                                      \
    {                                                                                                                                        \
        .len = (⁵initial_len__), .cap = (((⁴max_capacity__) < (⁵initial_len__)) ? (⁵initial_len__) : (⁴max_capacity__)),                     \
        .at = (T**)(memAlloc((((⁴max_capacity__) < (⁵initial_len__)) ? (⁵initial_len__) : (⁴max_capacity__)) * (sizeof(T*))))                \
    }



U8* memAlloc(UInt const num_bytes) {
    UInt const new_pos = mem.pos + num_bytes;
    if (new_pos >= mem_max - 1)
        ·fail(str("out of memory: increase mem_max!"));
    U8* const mem_ptr = &mem.buf[mem.pos];
    mem.pos = new_pos;
    return mem_ptr;
}

Str newStr(UInt const initial_len, UInt const max_capacity) {
    Str ret_str = (Str) {.len = initial_len, .at = memAlloc(max_capacity)};
    return ret_str;
}

Str uIntToStr(UInt const uInt_value, UInt const str_min_len, UInt const base) {
    UInt num_digits = 1;
    UInt n = uInt_value;
    while (n >= base) {
        num_digits += 1;
        n /= base;
    }
    n = uInt_value;

    UInt const str_len = (num_digits > str_min_len) ? num_digits : str_min_len;
    Str const ret_str = newStr(str_len, str_len + 1);
    ret_str.at[str_len] = 0;
    for (UInt i = 0; i < str_len - num_digits; i += 1)
        ret_str.at[i] = '0';

    Bool done = false;
    for (UInt i = ret_str.len; i > 0 && !done;) {
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

Str strCopy(CStr const c_str) {
    Str const ptr_ref = str(c_str);
    Str copy = newStr(ptr_ref.len, ptr_ref.len);
    for (UInt i = 0; i < copy.len; i += 1)
        copy.at[i] = ptr_ref.at[i];
    return copy;
}

// unused in principle, but kept around for the occasional temporary printf.
CStr strZ(Str const str) {
    U8* buf = memAlloc(1 + str.len);
    buf[str.len] = 0;
    for (UInt i = 0; i < str.len; i += 1)
        buf[i] = str.at[i];
    return (CStr)buf;
}

Str strQuot(Str const str) {
    Str ret_str = newStr(1, 3 + (3 * str.len));
    ret_str.at[0] = '\"';
    for (UInt i = 0; i < str.len; i += 1) {
        U8 const chr = str.at[i];
        if (chr >= 32 && chr < 127) {
            ret_str.at[ret_str.len] = chr;
            ret_str.len += 1;
        } else {
            ret_str.at[ret_str.len] = '\\';
            const Str esc_num_str = uIntToStr(chr, 3, 10);
            for (UInt c = 0; c < esc_num_str.len; c += 1)
                ret_str.at[1 + c + ret_str.len] = esc_num_str.at[c];
            ret_str.len += 1 + esc_num_str.len;
        }
    }
    ret_str.at[ret_str.len] = '\"';
    ret_str.len += 1;
    ret_str.at[ret_str.len] = 0;
    return ret_str;
}

Str strConcat(Strs const strs, U8 const sep) {
    UInt str_len = 0;
    ·forEach(Str, str, strs, { str_len += (sep == 0 ? 0 : 1) + str->len; });

    Str ret_str = newStr(0, str_len);
    ·forEach(Str, str, strs, {
        if (iˇstr != 0 && sep != 0) {
            ret_str.at[ret_str.len] = sep;
            ret_str.len += 1;
        }
        for (UInt i = 0; i < str->len; i += 1)
            ret_str.at[i + ret_str.len] = str->at[i];
        ret_str.len += str->len;
    });
    return ret_str;
}

Str str2(Str const s1, Str const s2) {
    return strConcat((Strs) {.len = 2, .at = ((Str[]) {s1, s2})}, 0);
}

Str str3(Str const s1, Str const s2, Str const s3) {
    return strConcat((Strs) {.len = 3, .at = ((Str[]) {s1, s2, s3})}, 0);
}

Str str4(Str const s1, Str const s2, Str const s3, Str const s4) {
    return strConcat((Strs) {.len = 4, .at = ((Str[]) {s1, s2, s3, s4})}, 0);
}

Str str5(Str const s1, Str const s2, Str const s3, Str const s4, Str const s5) {
    return strConcat((Strs) {.len = 5, .at = ((Str[]) {s1, s2, s3, s4, s5})}, 0);
}

Str str6(Str const s1, Str const s2, Str const s3, Str const s4, Str const s5, Str const s6) {
    return strConcat((Strs) {.len = 6, .at = ((Str[]) {s1, s2, s3, s4, s5, s6})}, 0);
}

Str str7(Str const s1, Str const s2, Str const s3, Str const s4, Str const s5, Str const s6, Str const s7) {
    return strConcat((Strs) {.len = 7, .at = ((Str[]) {s1, s2, s3, s4, s5, s6, s7})}, 0);
}

Str str8(Str const s1, Str const s2, Str const s3, Str const s4, Str const s5, Str const s6, Str const s7, Str const s8) {
    return strConcat((Strs) {.len = 8, .at = ((Str[]) {s1, s2, s3, s4, s5, s6, s7, s8})}, 0);
}



void printChr(U8 const chr) {
    fwrite(&chr, 1, 1, stderr);
}

void failIf(int some_err) {
    if (some_err)
        ·fail(str2(str("error code: "), uIntToStr(some_err, 1, 10)));
}

Str ident(Str const str) {
    Str ret_ident = newStr(0, 4 * str.len);
    Bool all_chars_ok = true;
    for (UInt i = 0; i < str.len; i += 1) {
        U8 c = str.at[i];
        if ((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '.' || c == '-' || c == '_' || c == '$')
            ·push(ret_ident, c);
        else {
            all_chars_ok = false;
            Str const hex = uIntToStr(c, 1, 16);
            ·push(ret_ident, '-');
            for (UInt j = 0; j < hex.len; j += 1)
                ·push(ret_ident, hex.at[j]);
            ·push(ret_ident, '-');
        }
    }
    if (all_chars_ok) {
        mem.pos -= str.len;
        ret_ident = str;
    } else
        mem.pos -= (2 * str.len) - ret_ident.len;
    return ret_ident;
}
