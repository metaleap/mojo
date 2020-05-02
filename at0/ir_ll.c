#include "utils_and_libc_deps.c"

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
} IrLLProg;
