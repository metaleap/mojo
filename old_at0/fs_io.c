#pragma once
#include "utils_and_libc_deps.c"


Str readUntilEof(FILE* const stream) {
    const UInt buf_size = 4096;
    Str ret_str = {.len = 0, .at = memAlloc(buf_size)};
    UInt n_bytes_over_allocated = 0;
    for (PtrAny dst_addr = ret_str.at; true;) {
        UInt const n_read = fread(dst_addr, 1, buf_size, stream);
        ret_str.len += n_read;

        if (n_read != buf_size) {
            failIf(ferror(stream));
            if (feof(stream)) { // reading is done
                n_bytes_over_allocated = buf_size - n_read;
                break;
            }
        }
        if ((ret_str.len % buf_size) == 0) // ret_str is "full"?
            dst_addr = memAlloc(buf_size); // but: reading is not done, so expand ret_str
    }
    mem.pos -= n_bytes_over_allocated;
    return ret_str;
}

Str readFile(Str const file_path) {
    FILE* const file = fopen(strZ(file_path), "rb");
    if (file == NULL)
        ·fail(str2(str("failed to open "), file_path));
    Str const file_bytes = readUntilEof(file);
    fclose(file);
    return file_bytes;
}

Str pathJoin(Str const prefix, Str const suffix) {
    return (prefix.len == 0 && suffix.len == 0)
               ? ·len0(U8)
               : (prefix.len == 0) ? suffix : (suffix.len == 0) ? prefix : str3(prefix, strL("/", 1), suffix);
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
