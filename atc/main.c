#include "at_toks.h"
#include "at_ast.h"
#include "std.h"
#include "std_io.h"


/*

Fairly unidiomatic code! because we want to have a most compact C code base to
later transliterate smoothly into our initially very limited language iteration:

- no proper error handling, immediate `panic`s upon detecting a problem
- no 3rd-party imports / deps whatsoever
- no stdlib imports for *core* processing (just for basic program setup & I/O)
  (hence manual implementations like uintToStr, uintParse, strEql etc)
- use of macros limited to (eventual) WIP-lang meta-programming / generic facilities
- all would-be `malloc`s replaced by global fixed-size backing buffer allocation
- naming / casing conventions follow WIP-lang rather than C customs
- no zero-terminated C strings, except for `%s` in `fprintf` in `panic`

We want here to merely reach the "interpret-source-files-or-die" stage. No bells &
whistles, no *fancy* type stuff, no syntax sugars (not even operators, we endure
prim calls). No nifty optimizations, no proper byte code, will be slow. Once there,
get to the same stage in WIP-lang, interpreter-in-interpreter. At that point then,
worry about compilation next before advancing anything else. Thus the foundation
itself can move from a "host language" (like C) to self-hosted / rolling LLVM-IR.

*/


int main(int const argc, String const argv[]) {
    if (argc < 2)
        panic("expected usage: atc <src_file_path>");

    Str input_src_file_bytes = readFile(argv[1]);

    Tokens toks = tokenize(input_src_file_bytes, false);
    toksCheckBrackets(toks);
    printf("%zu\n", toks.len);
}
