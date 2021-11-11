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


int main_IrMl(int const argc, CStr const argv[]);
int main_AstAndIrHl(int const argc, CStr const argv[]);

int main(int const argc, CStr const argv[]) {
    // return main_AstAndIrHl(argc,argv);
    return main_IrMl(argc, argv);
}

int main_IrMl(int const argc, CStr const argv[]) {
    IrMlProg p = irmlProg(64, 32, 32);

#define _ IrMlNode*
    // fn (int, fn(int) -> _) -> _
    _ fn_main_type = irmlTypeCont2(&p, irmlTypeIntStatic(&p), irmlTypeCont1(&p, irmlTypeIntStatic(&p)));
    // main x := (x == 123) ?- 22 |- 44
    _ fn_main = irmlNodeCont(&p, fn_main_type, "main");
    _ fn_if_then = irmlNodeCont(&p, irmlTypeCont0(&p), "main_if_then");
    _ fn_if_else = irmlNodeCont(&p, irmlTypeCont0(&p), "main_if_else");
    _ fn_next = irmlNodeCont(&p, irmlTypeCont1(&p, irmlTypeIntStatic(&p)), "main_next");
    _ cmp_p0_eq_123 = irmlNodePrimCmpI(&p, (IrMlPrimCmpI) {
                                               .kind = irml_cmpi_eq,
                                               //    .lhs = irmlNodePrimValInt(&p, 123),
                                               .lhs = &fn_main->of.cont.params.at[0],
                                               .rhs = irmlNodePrimValInt(&p, 123),
                                           });
    irmlContJump(&p, fn_main,
                 (IrMlNodeJump) {
                     .args = irmlNodes0(),
                     .target = irmlNodePrimCond(&p, irmlCondBoolish(&p, cmp_p0_eq_123, fn_if_then, fn_if_else, NULL)),
                 });
    irmlContJump(&p, fn_if_then, (IrMlNodeJump) {.target = fn_next, .args = irmlNodes1(irmlNodePrimValInt(&p, 22))});
    irmlContJump(&p, fn_if_else, (IrMlNodeJump) {.target = fn_next, .args = irmlNodes1(irmlNodePrimValInt(&p, 44))});
    irmlContJump(&p, fn_next,
                 (IrMlNodeJump) {
                     .target = &fn_main->of.cont.params.at[1],
                     .args = irmlNodes1(&fn_next->of.cont.params.at[0]),
                 });

    irmlPrint(fn_main);
    printf("\n\n———————————\n\n");

    IrMlCtxPreduce ctx = (IrMlCtxPreduce) {.prog = &p, .cur_cont = NULL, .reduce = true};
    irmlPreduceNode(&ctx, fn_main);

    printf("\n\n———————————\n\n");
    irmlPrint(fn_main);

    printf("\n\n———————————\n\n");
    IrMlNode* result = irmlRun(&p, fn_main,
                               irmlNodes2(irmlNodePrimBinI(&p,
                                                           (IrMlPrimBinI) {
                                                               .kind = irml_bini_add,
                                                               .lhs = irmlNodePrimValInt(&p, 100),
                                                               .rhs = irmlNodePrimValInt(&p, 23),
                                                           }),
                                          NULL));
    irmlPrint(result);
    printf("\n\n———————————\n\n");

    return 0;
}

int main_AstAndIrHl(int const argc, CStr const argv[]) {
    ·assert(argc == 2);

    CtxParseAsts ctx_parse = (CtxParseAsts) {.asts = ·listOf(Ast, 0, asts_capacity), .src_file_paths = ·sliceOf(Str, 0, asts_capacity)};
    ·push(ctx_parse.src_file_paths, str(argv[1]));
    loadAndParseRootSourceFileAndImports(&ctx_parse);
    // astPrint(&ctx_parse.asts.at[0]);

    IrHlProg ir_hl = irhlProgFrom(ctx_parse.asts);
    irhlProcessIdents(&ir_hl);
    irhlProgLiftFuncExprs(&ir_hl);
    // irhlProgRewriteLetsIntoLambdas(&ir_hl);
    // irhlProgLiftFuncExprs(&ir_hl);
    irhlPrintProg(&ir_hl);

    // IrHlDef* entry_def = irhlProgDef(&ir_hl, ctx_parse.asts.at[0].anns.path_based_ident_prefix, str("main"));
    // ·assert(entry_def != NULL);
    // IrLLProg ir_ll = irLLProgFrom(entry_def, &ir_hl);
    // irLLPrintProg(&ir_ll);

    // readLnLoop(&ir_hl);
    return 0;
}
