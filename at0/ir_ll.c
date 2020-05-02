#pragma once
#include "utils_and_libc_deps.c"
#include "ir_hl.c"

typedef enum IrLLInstrKind {
    irll_instr_arith_add,
    irll_instr_arith_sub,
    irll_instr_arith_mul,
    irll_instr_arith_div,
    irll_instr_arith_rem,
    irll_instr_cmp_eq,
    irll_instr_cmp_neq,
    irll_instr_cmp_leq,
    irll_instr_cmp_geq,
    irll_instr_cmp_gt,
    irll_instr_cmp_lt,
} IrLLInstrKind;

typedef enum IrLLExprKind {
    irll_expr_int,
    irll_expr_aggr,
    irll_expr_ref_arg,
    irll_expr_ref_func,
    irll_expr_instr,
    irll_expr_call,
} IrLLExprKind;

struct IrLLExprCall;
typedef struct IrLLExprCall IrLLExprCall;
struct IrLLExprAggr;
typedef struct IrLLExprAggr IrLLExprAggr;

typedef struct IrLLExpr {
    IrLLExprKind kind;
    union {
        I64 of_int;
        IrLLExprAggr* of_aggr;
        UInt of_ref_arg;
        UInt of_ref_func;
        IrLLInstrKind of_instr;
        IrLLExprCall* of_call;
    };
} IrLLExpr;
typedef ·SliceOf(IrLLExpr) IrLLExprs;

struct IrLLExprAggr {
    IrLLExprs items;
    Strs field_names;
};

struct IrLLExprCall {
    IrLLExpr callee;
    IrLLExprs args;
    UInt is_closure;
};

typedef struct IrLLFunc {
    UInt num_args;
    IrLLExpr body;
    struct {
        Str name;
    } anns;
} IrLLFunc;
typedef ·ListOf(IrLLFunc) IrLLFuncs;

typedef struct IrLLProg {
    IrLLFuncs funcs;
    struct {
        IrHLProg* origin;
    } anns;
} IrLLProg;



IrLLFunc irLLFuncFrom(IrHLDef const* const func_def, IrLLProg* const ll_prog) {
    ·assert(func_def->body->kind == irhl_expr_func);
    IrHLExprFunc* func = &func_def->body->of_func;
    return (IrLLFunc) {.num_args = 0, .body = (IrLLExpr) {.kind = irll_expr_int, .of_int = func->params.len}};
}

IrLLProg irLLProgFrom(IrHLDef* entry_def, IrHLProg* const ir_hl_prog) {
    IrLLProg ret_prog = (IrLLProg) {.funcs = ·listOf(IrLLFunc, 0, ir_hl_prog->defs.len), .anns = {.origin = ir_hl_prog}};

    return ret_prog;
}




void irLLPrintFunc(IrLLFunc const* const func) {
    printStr(func->anns.name);
    printChr('\n');
}

void irLLPrintProg(IrLLProg const* const prog) {
    ·forEach(IrLLFunc, func, prog->funcs, { irLLPrintFunc(func); });
}
