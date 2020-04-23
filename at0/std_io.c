#pragma once
#include "metaleap.c"


void printChr(U8 const chr) {
    fwrite(&chr, 1, 1, stderr);
}

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

Str readFile(CStr const file_path) {
    FILE* const file = fopen(file_path, "rb");
    if (file == NULL)
        ·fail(str2(str("failed to open "), str(file_path)));
    Str const file_bytes = readUntilEof(file);
    fclose(file);
    return file_bytes;
}
