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
    irll_expr_ref_arg,
    irll_expr_ref_def,
    irll_expr_ref_instr,
    irll_expr_call,
} IrLLExprKind;

struct IrLLExprCall;
typedef struct IrLLExprCall IrLLExprCall;

typedef struct IrLLExpr {
    IrLLExprKind kind;
    union {
        I64 of_int;
        UInt of_ref_arg;
        UInt of_ref_def;
        IrLLInstrKind of_ref_instr;
        IrLLExprCall* of_call;
    };
} IrLLExpr;
typedef ·SliceOf(IrLLExpr) IrLLExprs;

struct IrLLExprCall {
    IrLLExpr callee;
    IrLLExprs args;
    UInt is_closure;
};

typedef struct IrLLDef {
    UInt num_args;
    IrLLExpr body;

} IrLLDef;
typedef ·SliceOf(IrLLDef) IrLLDefs;

typedef struct IrLLProg {
    IrLLDefs defs;
} IrLLProg;

void foo() {
}
