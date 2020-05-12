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
    IrMlProg p = irmlProg(64, 32, 32, 32);

#define _ IrMlNode*
    // main x := (x == 123) ?- 22 |- 44
    _ fn_main = irmlNodeFn(&p, irmlTypeFn2(&p, irmlTypeIntStatic(&p), irmlTypeFn1(&p, irmlTypeIntStatic(&p))));
    _ fn_if_then = irmlNodeFn(&p, irmlTypeFn0(&p));
    _ fn_if_else = irmlNodeFn(&p, irmlTypeFn0(&p));
    _ fn_next = irmlNodeFn(&p, irmlTypeFn1(&p, irmlTypeIntStatic(&p)));
    _ cmp_p0_eq_123 = irmlNodePrimCmpI(&p, (IrMlPrimCmpI) {
                                               .kind = irml_cmp_i_eq,
                                               .lhs = &fn_main->of.fn.params.at[0],
                                               .rhs = irmlNodePrimValInt(&p, 123),
                                           });
    irmlFnChoice(&p, fn_main, (IrMlNodeChoice) {.cond = cmp_p0_eq_123, .if1 = fn_if_then, .if0 = fn_if_else});
    irmlFnJump(&p, fn_if_then, (IrMlNodeJump) {.callee = fn_next, .args = irmlNodes1(irmlNodePrimValInt(&p, 22))});
    irmlFnJump(&p, fn_if_else, (IrMlNodeJump) {.callee = fn_next, .args = irmlNodes1(irmlNodePrimValInt(&p, 44))});
    irmlFnJump(&p, fn_next,
               (IrMlNodeJump) {
                   .callee = &fn_main->of.fn.params.at[1],
                   .args = irmlNodes1(&fn_next->of.fn.params.at[0]),
               });

    IrMlCtxPreduce ctx = (IrMlCtxPreduce) {.prog = &p};
    irmlPreduceNode(&ctx, fn_main);

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
