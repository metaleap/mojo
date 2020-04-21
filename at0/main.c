#pragma clang diagnostic ignored "-Wunused-parameter"
#pragma clang diagnostic ignored "-Wunused-function"

#include "metaleap.c"
#include "std_io.c"

#include "at_toks.c"
#include "at_ast.c"
#include "at_parse.c"
#include "ir_hl.c"



int main(int const argc, String const argv[]) {
    ·assert(argc > 1);

    Str full_src = (Str) {.at = NULL, .len = 0};
    for (int i = 1; i < argc; i += 1) {
        Str const comment1 = strCopy("//AT_TOKS_SRC_FILE:");
        Str const comment2 = strCopy(argv[i]);
        Str const comment3 = strCopy("\n");
        Str const this_file_src = readFile(argv[i]);
        full_src.len += comment1.len + comment2.len + comment3.len + this_file_src.len;
        if (i == 1)
            full_src.at = comment1.at;
    }

    Tokens const toks = tokenize(full_src, false, str(""));
    toksCheckBrackets(toks);

    Ast ast = parse(toks, full_src);
    astDesugarGlyphsIntoInstrs(&ast);
    astDefsVerifyNoShadowings(ast.top_defs, ·make(Str, 0, 64), 64, &ast);
    astHoistFuncsExprsToNewTopDefs(&ast);
    // astReorderSubDefs(&ast);
    // astPrint(&ast);

    IrHLProg ir_hl_prog = irHLProgFrom(&ast);
    irHLProgPrint(&ir_hl_prog);
}
