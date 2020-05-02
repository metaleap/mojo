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
    irll_expr_list,
    irll_expr_ref_arg,
    irll_expr_ref_func,
    irll_expr_instr,
    irll_expr_call,
} IrLLExprKind;

struct IrLLExprCall;
typedef struct IrLLExprCall IrLLExprCall;
struct IrLLExprList;
typedef struct IrLLExprList IrLLExprList;

typedef struct IrLLExpr {
    IrLLExprKind kind;
    union {
        I64 of_int;
        IrLLExprList* of_list;
        UInt of_ref_arg;
        UInt of_ref_func;
        IrLLInstrKind of_instr;
        IrLLExprCall* of_call;
    };
} IrLLExpr;
typedef ·SliceOf(IrLLExpr) IrLLExprs;

struct IrLLExprList {
    IrLLExprs items;
};

struct IrLLExprCall {
    IrLLExpr callee;
    IrLLExprs args;
    UInt is_closure;
};

typedef struct IrLLFunc {
    UInt num_args;
    IrLLExpr body;
} IrLLFunc;
typedef ·SliceOf(IrLLFunc) IrLLFuncs;

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

IrLLProg irLLProgFrom(IrHLProg* const ir_hl_prog, Str const module_name, Str const entry_func_name) {
    IrLLProg ret_prog = (IrLLProg) {.funcs = ·sliceOf(IrLLFunc, 0, ir_hl_prog->defs.len), .anns = {.origin = ir_hl_prog}};

    IrHLDef* module = NULL;
    ·forEach(IrHLDef, def, ir_hl_prog->defs, {
        if (strEql(module_name, def->name)) {
            module = def;
            break;
        }
    });
    if (module == NULL || module->body->kind != irhl_expr_bag || module->body->of_bag.kind != irhl_bag_struct)
        ·fail(str2(str("module not found: "), module_name));

    IrHLExprRef* entry_func_ref = NULL;
    ·forEach(IrHLExpr, bag_item, module->body->of_bag.items, {
        if (bag_item->kind == irhl_expr_kvpair && bag_item->of_kvpair.key->kind == irhl_expr_field_name
            && bag_item->of_kvpair.val->kind == irhl_expr_ref && strEql(entry_func_name, bag_item->of_kvpair.key->of_field_name.field_name)) {
            entry_func_ref = &bag_item->of_kvpair.val->of_ref;
            break;
        }
    });
    if (entry_func_ref == NULL)
        ·fail(str4(str("module "), module_name, str("has no func "), entry_func_name));

    ·assert(·last(entry_func_ref->path)->kind == irhl_ref_def);
    IrLLFunc entry_func = irLLFuncFrom(·last(entry_func_ref->path)->of_def, &ret_prog);
    ·push(ret_prog.funcs, entry_func);

    return ret_prog;
}
