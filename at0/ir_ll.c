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
    irll_expr_tag,
    irll_expr_aggr,
    irll_expr_ref_param,
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
        UInt of_tag;
        IrLLExprAggr* of_aggr;
        UInt of_ref_param;
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
    UInt num_params;
    IrLLExpr body;
    struct {
        IrHLDef* origin;
    } anns;
} IrLLFunc;
typedef ·ListOf(IrLLFunc) IrLLFuncs;

typedef struct IrLLProg {
    IrLLFuncs funcs;
    struct {
        IrHLProg* origin;
    } anns;
} IrLLProg;




typedef struct CtxIrLLFromHL {
    IrLLProg prog;
    IrHLExprTags tags;
} CtxIrLLFromHL;

IrLLExpr irLLExprFrom(CtxIrLLFromHL* const ctx, IrHLExpr* const hl_expr) {
    IrLLExpr ret_expr = (IrLLExpr) {.kind = -1};
    switch (hl_expr->kind) {
        case irhl_expr_int: {
            ret_expr.kind = irll_expr_int;
            ret_expr.of_int = hl_expr->of_int.int_value;
        } break;
        case irhl_expr_tag: {
            UInt tag_idx = ctx->tags.len;
            if (tag_idx == ctx->tags.len)
                ·append(ctx->tags, hl_expr->of_tag);

            ret_expr.kind = irll_expr_tag;
            ret_expr.of_tag = tag_idx;
        } break;
        case irhl_expr_call: {
            ret_expr.kind = irll_expr_call;
            ret_expr.of_call = ·new(IrLLExprCall);
            ret_expr.of_call->callee = irLLExprFrom(ctx, hl_expr->of_call.callee);
            ret_expr.of_call->args = ·sliceOf(IrLLExpr, hl_expr->of_call.args.len, 0);
            ·forEach(IrHLExpr, arg, hl_expr->of_call.args, { ret_expr.of_call->args.at[iˇarg] = irLLExprFrom(ctx, arg); });
        } break;
        default: ·fail(str2(str("TODO: irLLExprFrom for .kind of "), uIntToStr(hl_expr->kind, 1, 10)));
    }
    return ret_expr;
}

UInt irLLFuncFrom(CtxIrLLFromHL* const ctx, IrHLDef* const hl_def) {
    ·forEach(IrLLFunc, func, ctx->prog.funcs, {
        if (func->anns.origin == hl_def)
            return iˇfunc;
    });

    UInt ret_idx = ctx->prog.funcs.len;
    ·append(ctx->prog.funcs, ((IrLLFunc) {.num_params = 0, .anns = {.origin = hl_def}}));
    IrLLFunc* func = ·last(ctx->prog.funcs);

    if (hl_def->body->kind != irhl_expr_func)
        func->body = irLLExprFrom(ctx, hl_def->body);
    else {
        func->num_params = hl_def->body->of_func.params.len;
        func->body = irLLExprFrom(ctx, hl_def->body->of_func.body);
    }

    return ret_idx;
}

IrLLProg irLLProgFrom(IrHLDef* hl_def, IrHLProg* const hl_prog) {
    CtxIrLLFromHL ctx = (CtxIrLLFromHL) {.tags = ·listOf(IrHLExprTag, 0, 4),
                                         .prog = (IrLLProg) {
                                             .funcs = ·listOf(IrLLFunc, 0, hl_prog->defs.len),
                                             .anns = {.origin = hl_prog},
                                         }};
    irLLFuncFrom(&ctx, hl_def);
    return ctx.prog;
}




void irLLPrintFunc(IrLLFunc const* const func) {
    printStr(func->anns.origin->name);
    printChr('\n');
}

void irLLPrintProg(IrLLProg const* const prog) {
    ·forEach(IrLLFunc, func, prog->funcs, { irLLPrintFunc(func); });
}
