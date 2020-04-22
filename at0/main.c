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

    // read and concat together all input source files specified via args
    Str full_src = (Str) {.at = NULL, .len = 0};
    for (int i = 1; i < argc; i += 1) {
        // hacky: all allocs in this loop (strCopy and readFile) are contiguous in memory,
        // so our `full_src` bytes-slice just gets the starting addr and its `len` increased
        Str const comment_part_1 = strCopy("//AT_TOKS_SRC_FILE:");
        Str const comment_part_2 = strCopy(argv[i]);
        Str const comment_part_3 = strCopy("\n");
        Str const this_file_src = readFile(argv[i]);
        full_src.len += comment_part_1.len + comment_part_2.len + comment_part_3.len + this_file_src.len;
        if (full_src.at == NULL)
            full_src.at = comment_part_1.at;
    }


    // tokenize
    Tokens const toks = tokenize(full_src, false, str(""));
    toksVerifyBrackets(toks);


    // parse into a rudimentary raw context-free generic AST first
    Ast ast = parse(toks, full_src);
    astRewriteGlyphsIntoInstrs(&ast);
    astDefsVerifyNoShadowings(ast.top_defs, ·make(Str, 0, 64), 64, &ast);
    // astHoistFuncsExprsToNewTopDefs(&ast);
    // astReorderSubDefs(&ast);
    // astPrint(&ast);


    // interpret raw-and-dumb *syntax* tree into actual language *semantics*:
    // IrHLProg ir_hl_prog = irHLProgFrom(&ast);
    // irHLProgPrint(&ir_hl_prog);
}
