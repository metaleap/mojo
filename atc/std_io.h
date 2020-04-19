#pragma once
#include "metaleap.h"

void printChr(U8 const chr) {
    fwrite(&chr, 1, 1, stderr);
}
void writeChr(U8 const chr) {
    fwrite(&chr, 1, 1, stdout);
}

Str readUntilEof(FILE* const stream) {
    const Uint buf_size = 4096;
    Str ret_str = {.len = 0, .at = memAlloc(buf_size)};
    for (PtrAny dst_addr = ret_str.at; true;) {
        Uint const n_read = fread(dst_addr, 1, buf_size, stream);
        ret_str.len += n_read;

        if (n_read != buf_size) {
            failIf(ferror(stream));
            if (feof(stream)) // reading is done
                break;
        }

        if ((ret_str.len % buf_size) == 0) // ret_str is "full"?
            dst_addr = memAlloc(buf_size); // but: reading is not done, so expand ret_str
    }
    return ret_str;
}

Str readFile(String const file_path) {
    FILE* const file = fopen(file_path, "rb");
    if (file == null)
        fail(str2(str("failed to open "), str(file_path)));
    Str const file_bytes = readUntilEof(file);
    fclose(file);
    return file_bytes;
}
