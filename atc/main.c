#include "metaleap.h"
#include "std_io.h"
#include "at_toks.h"
#include "at_ast.h"
#include "at_parse.h"



int main(int const argc, String const argv[]) {
    Str const input_src_file_bytes = (argc < 2) ? readUntilEof(stdin) : readFile(argv[1]);

    Tokens const toks = tokenize(input_src_file_bytes, false);
    toksCheckBrackets(toks);

    Ast ast = parse(toks, input_src_file_bytes);
    astPrint(&ast);
}
