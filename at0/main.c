#pragma clang diagnostic ignored "-Wunused-parameter"
#pragma clang diagnostic ignored "-Wunused-function"

#include "utils_and_libc_deps.c"
#include "fs_io.c"

#include "at_toks.c"
#include "at_ast.c"
#include "at_parse.c"
#include "ir_hl.c"
#include "ir_ll.c"

#include "mtp.c"


int main_MiniThorinProto(int const argc, CStr const argv[]);
int main_AstAndIrHL(int const argc, CStr const argv[]);

int main(int const argc, CStr const argv[]) {
    // return main_AstAndIrHL(argc,argv);
    return main_MiniThorinProto(argc, argv);
}

int main_MiniThorinProto(int const argc, CStr const argv[]) {
    MtpProg prog = mtpProg(64, 32, 32, 32);
    printf("%zu\n", prog.all.prims.len);
    return prog.all.prims.len;
}

int main_AstAndIrHL(int const argc, CStr const argv[]) {
    ·assert(argc == 2);

    CtxParseAsts ctx_parse = (CtxParseAsts) {.asts = ·listOf(Ast, 0, asts_capacity), .src_file_paths = ·sliceOf(Str, 0, asts_capacity)};
    ·push(ctx_parse.src_file_paths, str(argv[1]));
    loadAndParseRootSourceFileAndImports(&ctx_parse);
    // astPrint(&ctx_parse.asts.at[0]);

    IrHLProg ir_hl = irHLProgFrom(ctx_parse.asts);
    irHLProcessIdents(&ir_hl);
    irHLProgLiftFuncExprs(&ir_hl);
    // irHLProgRewriteLetsIntoLambdas(&ir_hl);
    // irHLProgLiftFuncExprs(&ir_hl);
    irHLPrintProg(&ir_hl);

    // IrHLDef* entry_def = irHLProgDef(&ir_hl, ctx_parse.asts.at[0].anns.path_based_ident_prefix, str("main"));
    // ·assert(entry_def != NULL);
    // IrLLProg ir_ll = irLLProgFrom(entry_def, &ir_hl);
    // irLLPrintProg(&ir_ll);

    // readLnLoop(&ir_hl);
    return 0;
}
