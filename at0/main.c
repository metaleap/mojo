#pragma clang diagnostic ignored "-Wunused-parameter"
#pragma clang diagnostic ignored "-Wunused-function"

#include "utils_and_libc_deps.c"
#include "fs_io.c"

#include "at_toks.c"
#include "at_ast.c"
#include "at_parse.c"
#include "ir_hl.c"
#include "ir_ll.c"

#include "ir_ml.c"


int main_MiniThorinProto(int const argc, CStr const argv[]);
int main_AstAndIrHL(int const argc, CStr const argv[]);

int main(int const argc, CStr const argv[]) {
    // return main_AstAndIrHL(argc,argv);
    return main_MiniThorinProto(argc, argv);
}

int main_MiniThorinProto(int const argc, CStr const argv[]) {
    MtpProg p = mtpProg(64, 32, 32, 32);

#define _ MtpNode*
    // main x := (x == 123) ?- 22 |- 44
    _ fn_main = mtpNodeFn(&p, mtpTypeFn2(&p, mtpTypeIntStatic(&p), mtpTypeFn1(&p, mtpTypeIntStatic(&p))));
    _ fn_if_then = mtpNodeFn(&p, mtpTypeFn0(&p));
    _ fn_if_else = mtpNodeFn(&p, mtpTypeFn0(&p));
    _ fn_next = mtpNodeFn(&p, mtpTypeFn1(&p, mtpTypeIntStatic(&p)));
    _ cmp_p0_eq_123 = mtpNodePrimCmpI(&p, (MtpPrimCmpI) {
                                              .kind = mtp_cmp_i_eq,
                                              .lhs = &fn_main->of.fn.params.at[0],
                                              .rhs = mtpNodePrimValInt(&p, 123),
                                          });
    mtpFnChoice(&p, fn_main, (MtpNodeChoice) {.cond = cmp_p0_eq_123, .if1 = fn_if_then, .if0 = fn_if_else});
    mtpFnJump(&p, fn_if_then, (MtpNodeJump) {.callee = fn_next, .args = mtpNodes1(mtpNodePrimValInt(&p, 22))});
    mtpFnJump(&p, fn_if_else, (MtpNodeJump) {.callee = fn_next, .args = mtpNodes1(mtpNodePrimValInt(&p, 44))});
    mtpFnJump(&p, fn_next,
              (MtpNodeJump) {
                  .callee = &fn_main->of.fn.params.at[1],
                  .args = mtpNodes1(&fn_next->of.fn.params.at[0]),
              });

    MtpCtxPreduce ctx = (MtpCtxPreduce) {.prog = &p};
    mtpPreduceNode(&ctx, fn_main);

    return 0;
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
