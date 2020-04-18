#pragma once
#include "at_toks.h"
#include "std.h"

Str readUntilEof(FILE *const stream) {
    Uint const buf_size = 4096;
    Str ret_str = {.len = 0, .at = memAlloc(buf_size)};
    for (Ptr addr = ret_str.at; true;) {
        Uint n_read = fread(addr, 1, buf_size, stream);
        ret_str.len += n_read;

        if (n_read != buf_size) {
            panicIf(ferror(stream));
            if (feof(stream)) // reading is done
                break;
        }

        if ((ret_str.len % buf_size) == 0) // ret_str is "full"?
            addr = memAlloc(buf_size);     // but reading is not done, so expand ret_str
    }
    return ret_str;
}

Str readFile(String const file_path) {
    FILE *file = fopen(file_path, "rb");
    if (file == NULL)
        panic("could not open %s", file_path);
    Str file_bytes = readUntilEof(file);
    fclose(file);
    return file_bytes;
}
