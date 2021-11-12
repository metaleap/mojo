#pragma once
#include <execinfo.h>
#include <limits.h>
#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>
#include "utils_std_basics.c"


#ifndef mem_bss_max
#define mem_bss_max (1 * 1024 * 1024)
#endif
struct {
    U8 buf[mem_bss_max];
    UInt pos;
} mem_bss = {.pos = 0};


typedef enum MemHeapKind {
    mem_heap_fixed_size,
    mem_heap_pages_malloc,
} MemHeapKind;

typedef struct MemHeap {
    U8* ptr;
    UInt cap;
    UInt len;
    MemHeapKind kind;
} MemHeap;

// itoa for uints
Str uIntToStr(MemHeap*, UInt const, UInt const, UInt const);

UInt memHeapSize(MemHeap* mem_heap, Bool const cap) {
    if (mem_heap == NULL)
        return mem_bss_max;
    switch (mem_heap->kind) {
        case mem_heap_fixed_size: return cap ? mem_heap->cap : mem_heap->len;
        case mem_heap_pages_malloc: {
            UInt ret_size = 0;
            Any ptr = mem_heap->ptr;
            while (ptr != NULL) {
                Any ptr_next = *((Any*)ptr);
                ret_size += (cap ? mem_heap->cap : mem_heap->len);
                ptr = ptr_next;
            }
            return ret_size;
        }
        default: ·fail(uIntToStr(NULL, mem_heap->kind, 1, 10));
    }
}

void memHeapFree(MemHeap* mem_heap) {
    if (mem_heap == NULL) {
        mem_bss.pos = 0;
        return;
    }
    switch (mem_heap->kind) {
        case mem_heap_fixed_size: {
            mem_heap->len = 0;
        } break;
        case mem_heap_pages_malloc: {
            Any ptr = mem_heap->ptr;
            mem_heap->ptr = NULL;
            mem_heap->len = 0;
            while (ptr != NULL) {
                Any const ptr_next = *((Any*)ptr);
                free(ptr);
                ptr = ptr_next;
            }
        } break;
        default: ·fail(uIntToStr(NULL, mem_heap->kind, 1, 10));
    }
}

Any memHeapAlloc(MemHeap* mem_heap, UInt const size) {
    if (size == 0)
        return NULL;
    MemHeap bss;
    Bool const is_bss = (mem_heap == NULL);
    if (is_bss) {
        bss = (MemHeap) {.ptr = &mem_bss.buf[0],
                         .cap = mem_bss_max,
                         .len = mem_bss.pos,
                         .kind = mem_heap_fixed_size};
        mem_heap = &bss;
    }
    UInt idx = mem_heap->len;
    UInt new_len = mem_heap->len + size;

    if (mem_heap->ptr == NULL || new_len > mem_heap->cap)
        switch (mem_heap->kind) {
            case mem_heap_fixed_size: {
                ·fail(str((!is_bss) ? "memAlloc: fixed-size allocator out of memory"
                                    : "TODO: increase mem_bss_max"));
            } break;
            case mem_heap_pages_malloc: {
                if ((sizeof(Any) + size) >= mem_heap->cap)
                    ·fail(str("memAlloc: insufficient .cap for mem_heap_pages_malloc"));
                ·assert(mem_heap->cap > 0);
                Any next_page = malloc(mem_heap->cap);
                if (next_page == NULL)
                    ·fail(str("memAlloc: heap allocator out of memory"));
                *((Any*)next_page) = mem_heap->ptr;
                mem_heap->ptr = next_page;
                idx = sizeof(Any);
                new_len = idx + size;
            } break;
            default: ·fail(uIntToStr(NULL, mem_heap->kind, 1, 10));
        }

    mem_heap->len = new_len;
    if (is_bss)
        mem_bss.pos = new_len;
    return mem_heap->ptr + idx;
}

Any memHeapPut(MemHeap* dst, Any src, UInt num_bytes) {
    Any ret_ptr = memHeapAlloc(dst, num_bytes);
    for (UInt i = 0; i < num_bytes; i += 1)
        ((U8*)ret_ptr)[i] = ((U8*)src)[i];
    return ret_ptr;
}

UInt memHeapCopy(MemHeap* src, U8* dst) {
    UInt ret_len = 0;
    if (src->ptr != NULL) {
        Bool const is_pages_malloc = (src->kind == mem_heap_pages_malloc);
        if (is_pages_malloc) {
            Any ptr_next = *((Any*)src->ptr);
            if (ptr_next != NULL) {
                MemHeap next = (MemHeap) {.cap = src->cap,
                                          .len = src->cap,
                                          .kind = mem_heap_pages_malloc,
                                          .ptr = ptr_next};
                ret_len += memHeapCopy(&next, dst);
            }
        }
        UInt const offset = is_pages_malloc ? sizeof(Any) : 0;
        for (UInt i = 0; i < src->len - offset; i += 1)
            dst[ret_len + i] = src->ptr[i + offset];
        ret_len += src->len - offset;
    }
    return ret_len;
}



// macro names prefixed with '·' instead of all upper-case (avoids SCREAM_CODE)

#define ·new(T, ¹mem_heap__) ((T*)memHeapAlloc((¹mem_heap__), sizeof(T)))

#define ·keep(T, ¹mem_heap__, ¹the_value__)                                            \
    ((T*)memHeapPut(¹mem_heap__, ¹the_value__, sizeof(T)))

#define ·sliceOf(T, ¹mem_heap__, ³initial_len__, ²max_capacity__)                      \
    ((T##s) {                                                                          \
        .len = (³initial_len__),                                                       \
        .at = (T*)(memHeapAlloc((¹mem_heap__), (((²max_capacity__) < (³initial_len__)) \
                                                    ? (³initial_len__)                 \
                                                    : (²max_capacity__))               \
                                                   * (sizeof(T))))})

#define ·sliceOfPtrs(T, ¹mem_heap__, ³initial_len__, ²max_capacity__)                  \
    {                                                                                  \
        .len = (³initial_len__),                                                       \
        .at =                                                                          \
            (T**)(memHeapAlloc((¹mem_heap__), (((²max_capacity__) < (³initial_len__))  \
                                                   ? (³initial_len__)                  \
                                                   : (²max_capacity__))                \
                                                  * (sizeof(T*))))                     \
    }

#define ·listOf(T, ¹mem_heap__, ⁵initial_len__, ⁴max_capacity__)                       \
    ((T##s) {                                                                          \
        .len = (⁵initial_len__),                                                       \
        .cap = (((⁴max_capacity__) < (⁵initial_len__)) ? (⁵initial_len__)              \
                                                       : (⁴max_capacity__)),           \
        .at = (T*)(memHeapAlloc((¹mem_heap__), (((⁴max_capacity__) < (⁵initial_len__)) \
                                                    ? (⁵initial_len__)                 \
                                                    : (⁴max_capacity__))               \
                                                   * (sizeof(T))))})

#define ·listOfPtrs(T, ¹mem_heap__, ⁵initial_len__, ⁴max_capacity__)                   \
    {                                                                                  \
        .len = (⁵initial_len__),                                                       \
        .cap = (((⁴max_capacity__) < (⁵initial_len__)) ? (⁵initial_len__)              \
                                                       : (⁴max_capacity__)),           \
        .at =                                                                          \
            (T**)(memHeapAlloc((¹mem_heap__), (((⁴max_capacity__) < (⁵initial_len__))  \
                                                   ? (⁵initial_len__)                  \
                                                   : (⁴max_capacity__))                \
                                                  * (sizeof(T*))))                     \
    }




Str newStr(MemHeap* mem_heap, UInt const initial_len, UInt const max_capacity) {
    Str ret_str = (Str) {
        .len = initial_len,
        .at = memHeapAlloc(
            mem_heap, 1 + ((max_capacity > initial_len) ? max_capacity : initial_len))};
    ret_str.at[max_capacity] = 0;
    ret_str.at[initial_len] = 0;
    return ret_str;
}

Str uintToBuf(Any str_buf, UInt const uint_value, UInt const str_min_len,
              UInt const base, UInt num_digits) {
    if (num_digits == 0) {
        num_digits = 1;
        UInt n = uint_value;
        while (n >= base) {
            num_digits += 1;
            n /= base;
        }
    }
    UInt const str_len = (num_digits > str_min_len) ? num_digits : str_min_len;
    Str const ret_str = (Str) {.at = str_buf, .len = str_len};
    ret_str.at[str_len] = 0;
    for (UInt i = 0; i < str_len - num_digits; i += 1)
        ret_str.at[i] = '0';

    Bool done = false;
    UInt n = uint_value;
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

Str uIntToStr(MemHeap* mem_heap, UInt const uint_value, UInt const str_min_len,
              UInt const base) {
    UInt num_digits = 1;
    UInt n = uint_value;
    while (n >= base) {
        num_digits += 1;
        n /= base;
    }

    UInt const str_len = (num_digits > str_min_len) ? num_digits : str_min_len;
    return uintToBuf(newStr(mem_heap, str_len, str_len + 1).at, uint_value, str_min_len,
                     base, num_digits);
}

CStr strZ(Str const str) {
    if (str.at[str.len] == 0)
        return (CStr)str.at;
    U8* buf = memHeapAlloc(NULL, 1 + str.len);
    buf[str.len] = 0;
    for (UInt i = 0; i < str.len; i += 1)
        buf[i] = str.at[i];
    return (CStr)buf;
}

Str strParse(MemHeap* mem_heap, Str const tok) {
    Str ret_str = ·len0(U8);
    if (tok.len >= 2 && tok.at[0] == '\"' && tok.at[tok.len - 1] == '\"') {
        ret_str = newStr(mem_heap, 0, tok.len - 1);

        for (UInt i = 1; i < tok.len - 1; i += 1) {
            U8 byte = tok.at[i];
            if (byte == '\\') {
                switch (tok.at[i + 1]) {
                    case 't': byte = '\t'; break;
                    case 'b': byte = '\b'; break;
                    case 'f': byte = '\f'; break;
                    case 'n': byte = '\n'; break;
                    case 'r': byte = '\r'; break;
                    case '\"': byte = '\"'; break;
                    case '\\': byte = '\\'; break;
                    default: return ·len0(U8);
                }
                i += 1;
            }
            ret_str.at[ret_str.len] = byte;
            ret_str.len += 1;
        }
        ret_str.at[ret_str.len] = 0;
    }
    return ret_str;
}

Str strQuotTo(Str buf, UInt buf_cap, Str const str) {
    buf.at[0] = '\"';
    buf.len += 1;
    ·assert(buf.len < buf_cap);
    for (UInt i = 0; (buf.len < buf_cap) && (i < str.len); i += 1) {
        U8 const chr = str.at[i];
        if (chr >= 32 && chr < 127 && chr != '\\' && chr != '\"')
            buf.at[buf.len] = chr;
        else {
            buf.at[buf.len] = '\\';
            buf.len += 1;
            if (buf.len < buf_cap)
                switch (chr) {
                    case '\t': buf.at[buf.len] = 't'; break;
                    case '\b': buf.at[buf.len] = 'b'; break;
                    case '\f': buf.at[buf.len] = 'f'; break;
                    case '\n': buf.at[buf.len] = 'n'; break;
                    case '\r': buf.at[buf.len] = 'r'; break;
                    case '\"': buf.at[buf.len] = '\"'; break;
                    case '\\': buf.at[buf.len] = '\\'; break;
                    default: {
                        buf.at[buf.len] = 'u';
                        buf.len += 1;
                        if ((buf.len + 4) < buf_cap)
                            buf = uintToBuf(buf.at, chr, 4, 16, 0);
                    } break;
                }
        }
        buf.len += 1;
    }
    ·assert(buf.len < buf_cap);
    buf.at[buf.len] = '\"';
    buf.len += 1;
    ·assert(buf.len < buf_cap);
    buf.at[buf.len] = 0;
    return buf;
}

Str strQuot(MemHeap* mem_heap, Str const str) {
    UInt const cap = 3 + (6 * str.len);
    return strQuotTo(newStr(mem_heap, 0, cap), cap, str);
}

Str strConcat(MemHeap* mem_heap, Strs const strs, U8 const sep) {
    UInt str_len = 0;
    UInt const sep_len = ((sep == 0) ? 0 : 1);
    ·forEach(Str, str, strs, { str_len += (sep_len + str->len); });

    Str ret_str = newStr(mem_heap, 0, 1 + str_len);
    ret_str.at[str_len] = 0;
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

Str str2(MemHeap* mem_heap, Str const s1, Str const s2) {
    return strConcat(mem_heap, (Strs) {.len = 2, .at = ((Str[]) {s1, s2})}, 0);
}

Str str3(MemHeap* mem_heap, Str const s1, Str const s2, Str const s3) {
    return strConcat(mem_heap, (Strs) {.len = 3, .at = ((Str[]) {s1, s2, s3})}, 0);
}

Str str4(MemHeap* mem_heap, Str const s1, Str const s2, Str const s3, Str const s4) {
    return strConcat(mem_heap, (Strs) {.len = 4, .at = ((Str[]) {s1, s2, s3, s4})}, 0);
}

Str str5(MemHeap* mem_heap, Str const s1, Str const s2, Str const s3, Str const s4,
         Str const s5) {
    return strConcat(mem_heap, (Strs) {.len = 5, .at = ((Str[]) {s1, s2, s3, s4, s5})},
                     0);
}

Str str6(MemHeap* mem_heap, Str const s1, Str const s2, Str const s3, Str const s4,
         Str const s5, Str const s6) {
    return strConcat(mem_heap,
                     (Strs) {.len = 6, .at = ((Str[]) {s1, s2, s3, s4, s5, s6})}, 0);
}

Str str7(MemHeap* mem_heap, Str const s1, Str const s2, Str const s3, Str const s4,
         Str const s5, Str const s6, Str const s7) {
    return strConcat(
        mem_heap, (Strs) {.len = 7, .at = ((Str[]) {s1, s2, s3, s4, s5, s6, s7})}, 0);
}

Str str8(MemHeap* mem_heap, Str const s1, Str const s2, Str const s3, Str const s4,
         Str const s5, Str const s6, Str const s7, Str const s8) {
    return strConcat(
        mem_heap, (Strs) {.len = 8, .at = ((Str[]) {s1, s2, s3, s4, s5, s6, s7, s8})},
        0);
}



void printChr(U8 const chr) {
    fwrite(&chr, 1, 1, stderr);
}

Str ident(MemHeap* mem_heap, Str const str) {
    Str ret_ident = newStr(mem_heap, 0, 4 * str.len);
    Bool all_chars_ok = true;
    for (UInt i = 0; i < str.len; i += 1) {
        U8 c = str.at[i];
        if ((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
            || c == '.' || c == '-' || c == '_' || c == '$')
            ·push(ret_ident, c);
        else {
            all_chars_ok = false;
            Str const hex = uIntToStr(mem_heap, c, 1, 16);
            ·push(ret_ident, '-');
            for (UInt j = 0; j < hex.len; j += 1)
                ·push(ret_ident, hex.at[j]);
            ·push(ret_ident, '-');
        }
    }
    if (all_chars_ok) {
        mem_bss.pos -= str.len;
        ret_ident = str;
    } else
        mem_bss.pos -= ((4 * str.len) - ret_ident.len);
    return ret_ident;
}
