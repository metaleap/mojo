#pragma once
#include "utils_std_mem.c"


Str readUntilEof(FILE* const stream) {
    const UInt buf_size = 4096;
    UInt const orig_mem_pos = mem_bss.pos;
    Str ret_str = {.len = 0, .at = memHeapAlloc(NULL, buf_size)};
    UInt n_bytes_over_allocated = 0;
    for (Any dst_addr = ret_str.at; true;) {
        UInt const n_read = fread(dst_addr, 1, buf_size, stream);
        ret_str.len += n_read;

        if (n_read != buf_size) {
            if (ferror(stream)) {
                mem_bss.pos = orig_mem_pos;
                return ·len0(U8);
            } else if (feof(stream)) { // reading is done
                n_bytes_over_allocated = buf_size - n_read;
                break;
            }
        }
        if ((ret_str.len % buf_size) == 0) // ret_str is "full"?
            dst_addr = memHeapAlloc(
                NULL, buf_size); // but: reading is not done, so expand ret_str
    }
    mem_bss.pos -= n_bytes_over_allocated;
    return ret_str;
}

Str readFile(Str const file_path) {
    FILE* const file = fopen(strZ(file_path), "rb");
    if (file == NULL)
        return ·len0(U8);
    Str const file_bytes = readUntilEof(file);
    fclose(file);
    return file_bytes;
}

Str readStdinUntilSuffix(Str const suffix) {
#define buf_size (1 * 1024 * 1024)
    static U8 buf[buf_size];
    UInt len = 0;

    while (true) {
        int chr = fgetc(stdin);
        if (chr <= 0 || chr > 255)
            exit((chr < 0) ? 0 : 1); // eof for our purposes

        buf[len] = chr;
        len += 1;
        if (len == buf_size)
            ·fail(str("malicious counterparty"));

        buf[len] = 0;
        if (strSuff((Str) {.at = &buf[0], .len = len}, suffix))
            break;
    }

    Str ret_str = (Str) {.at = &buf[0], .len = len};
    return ret_str;
}

Str pathJoin(Str const prefix, Str const suffix) {
    return (prefix.len == 0 && suffix.len == 0)
               ? ·len0(U8)
               : (prefix.len == 0)
                     ? suffix
                     : (suffix.len == 0) ? prefix
                                         : str3(NULL, prefix, strL("/", 1), suffix);
}

Str pathParent(Str const path) {
    UInt idx = 0;
    for (UInt i = path.len - 1; i > 0; i -= 1)
        if (path.at[i] == '/') {
            idx = i;
            break;
        }
    return strSub(path, 0, idx);
}

Str relPathFromRelPath(Str const path_own, Str const path_other) {
    Str const dir_path = pathParent(path_own);
    return pathJoin(dir_path, path_other);
}
