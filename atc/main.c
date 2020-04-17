#include "at_toks.h"
#include "std.h"
#include "std_io.h"


int main(int argc, String argv[]) {
    if (argc < 2)
        panic("expected usage: atc <src_file_path>\n");

    Str input_src_file_bytes = readFile(argv[1]);

    Tokens toks = tokenize(input_src_file_bytes, false);
    printf("Len: %zu\n", toks.len); // want 403

    printf("Uint parsed: ___%zu___ thx!\n", uintParse(str(argv[2])));
}
