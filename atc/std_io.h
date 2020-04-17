#pragma once
#include "at_toks.h"
#include "std.h"

Str readUntilEof(FILE *stream) {
    Str ret_str = {.len = 0, .at = memAlloc(4096)};
    for (Ptr addr = ret_str.at; true;) {
        ret_str.len += fread(addr, 1, 4096, stream);
        panicIf(ferror(stream));
        if (feof(stream))
            break;
        else if ((ret_str.len % 4096) == 0)
            addr = memAlloc(4096);
    }
    return ret_str;
}

Str readFile(String file_path) {
    FILE *file = fopen(file_path, "rb");
    if (file == NULL)
        panic("could not open %s", file_path);
    Str file_bytes = readUntilEof(file);
    fclose(file);
    return file_bytes;
}
