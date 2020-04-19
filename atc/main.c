#include "metaleap.h"
#include "std_io.h"
#include "at_toks.h"
#include "at_ast.h"
#include "at_parse.h"
#include "ir_hl.h"


int main(int const argc, String const argv[]) {
    ·assert(argc > 1);

    Str full_src = (Str) {.at = null, .len = 0};
    for (int i = 1; i < argc; i += 1) {
        Str const this_file_src = readFile(argv[i]);
        full_src.len += this_file_src.len;
        if (i == 1)
            full_src.at = this_file_src.at;
    }

    Tokens const toks = tokenize(full_src, false);
    toksCheckBrackets(toks);

    Ast ast = parse(toks, full_src);
    astPrint(&ast);

    irHLFrom(&ast);
}
