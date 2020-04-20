#include "metaleap.c"
#include "at_ast.c"
#include "std_io.c"


struct IrHLDef;
typedef struct IrHLDef IrHLDef;
typedef ·SliceOf(IrHLDef) IrHLDefs;

struct IrHLExpr;
typedef struct IrHLExpr IrHLExpr;
typedef ·SliceOf(IrHLExpr) IrHLExprs;

struct IrHLType;
typedef struct IrHLType IrHLType;
typedef ·SliceOf(IrHLType) IrHLTypes;




typedef struct IrHLProg {
    IrHLDefs defs;
    struct {
        Ast const* const origin_ast;
    } anns;
} IrHLProg;



typedef enum IrHLTypeKind {
    irhl_type_void,
    irhl_type_int,
    irhl_type_tag,
    irhl_type_struct,
    irhl_type_union,
    irhl_type_external,
    irhl_type_ptr,
    irhl_type_arr,
    irhl_type_func,
    irhl_type_type,
} IrHLTypeKind;

typedef struct IrHLTypeVoid {
} IrHLTypeVoid;

typedef struct IrHLTypeTag {
} IrHLTypeTag;

typedef struct IrHLTypeInt {
    U32 bit_width;
    struct {
        Bool word;
        Bool extc;
    } anns;
} IrHLTypeInt;

typedef struct IrHLTypeExt {
    Str name;
} IrHLTypeExt;

typedef struct IrHLTypePtr {
    IrHLType* ty;
} IrHLTypePtr;

typedef struct IrHLTypeArr {
    IrHLType* ty;
    Uint size;
} IrHLTypeArr;

typedef struct IrHLTypeFunc {
    IrHLType* ret;
    IrHLTypes params;
} IrHLTypeFunc;

typedef struct IrHLTypeType {
    IrHLType* ty;
} IrHLTypeType;

typedef struct {
    Str name;
    IrHLType* ty;
} IrHLBagField;
typedef ·SliceOf(IrHLBagField) IrHLBagFields;
typedef struct IrHLTypeBag {
    IrHLBagFields fields;
} IrHLTypeBag;

struct IrHLType {
    IrHLTypeKind kind;
    union {
        IrHLTypeVoid of_void;
        IrHLTypeInt of_int;
        IrHLTypeTag of_tag;
        IrHLTypeExt of_external;
        IrHLTypePtr of_ptr;
        IrHLTypeArr of_arr;
        IrHLTypeFunc of_func;
        IrHLTypeBag of_struct;
        IrHLTypeBag of_union;
        IrHLTypeType of_type;
    };
    struct {
    } anns;
};



typedef enum IrHLExprKind {
    irhl_expr_type,
    irhl_expr_void,
    irhl_expr_int,
    irhl_expr_func,
    irhl_expr_call,
    irhl_expr_list,
    irhl_expr_bag,
    irhl_expr_accessor,
    irhl_expr_pairing,
    irhl_expr_tagged,
    irhl_expr_ref,
    irhl_expr_prim,
} IrHLExprKind;

typedef struct IrHLExprType {
    IrHLType* ty_value;
} IrHLExprType;

typedef struct IrHLExprVoid {
} IrHLExprVoid;

typedef struct IrHLExprInt {
    I64 int_value;
} IrHLExprInt;

typedef struct IrHLFuncParam {
    struct {
        Str name;
    } anns;
} IrHLFuncParam;
typedef ·SliceOf(IrHLFuncParam) IrHLFuncParams;
typedef struct IrHLExprFunc {
    IrHLFuncParams params;
    IrHLExpr* body;
} IrHLExprFunc;

typedef struct IrHLExprCall {
    IrHLExpr* callee;
    IrHLExprs args;
} IrHLExprCall;

typedef struct IrHLExprList {
    IrHLExprs elems;
} IrHLExprList;

typedef struct IrHLExprBag {
    IrHLExprs elems;
} IrHLExprBag;

typedef struct IrHLExprAccessor {
    IrHLExpr* subj;
    IrHLExpr* selector;
} IrHLExprAccessor;

typedef struct IrHLExprPairing {
    IrHLExpr* key;
    IrHLExpr* val;
} IrHLExprPairing;
typedef ·SliceOf(IrHLExprPairing) IrHLExprPairings;

typedef struct IrHLExprTagged {
    Strs tags;
    IrHLExpr* subj;
} IrHLExprTagged;

typedef struct IrHLExprRef {
    Str name;
} IrHLExprRef;

typedef struct IrHLExprPrimArith {
    IrHLExpr* lhs;
    IrHLExpr* rhs;
    enum {
        irhl_arith_add,
        irhl_arith_sub,
        irhl_arith_mul,
        irhl_arith_div,
        irhl_arith_rem,
    } kind;
} IrHLExprPrimArith;

typedef struct IrHLExprPrimCmp {
    IrHLExpr* lhs;
    IrHLExpr* rhs;
    enum {
        irhl_cmp_eq,
        irhl_cmp_neq,
        irhl_cmp_lt,
        irhl_cmp_gt,
        irhl_cmp_leq,
        irhl_cmp_geq,
    } kind;
} IrHLExprPrimCmp;

typedef struct IrHLExprPrimExtCall {
    IrHLExpr* callee;
    IrHLExprs args;
} IrHLExprPrimExtCall;

typedef struct IrHLExprPrimCase {
    IrHLType* scrut;
    IrHLExprPairings cases;
    IrHLExpr* default_case;
} IrHLExprPrimCase;

typedef struct IrHLExprPrimLen {
    IrHLExpr* subj;
} IrHLExprPrimLen;

typedef struct IrHLExprPrimExtFn {
} IrHLExprPrimExtFn;

typedef struct IrHLExprPrimExtVar {
} IrHLExprPrimExtVar;

typedef struct IrHLExprPrim {
    Str kwd;
    enum {
        irhl_prim_unknown,
        irhl_prim_arith,
        irhl_prim_cmp,
        irhl_prim_ext_call,
        irhl_prim_case,
        irhl_prim_len,
        irhl_prim_ext_fn,
        irhl_prim_ext_var,
    } kind;
    union {
        Str of_unknown;
        IrHLExprPrimArith of_arith;
        IrHLExprPrimCmp of_cmp;
        IrHLExprPrimExtCall of_ext_call;
        IrHLExprPrimCase of_case;
        IrHLExprPrimLen of_len;
        IrHLExprPrimExtFn of_ext_fn;
        IrHLExprPrimExtVar of_ext_var;
    };
} IrHLExprPrim;

struct IrHLExpr {
    IrHLExprKind kind;
    union {
        IrHLExprType of_type;
        IrHLExprVoid of_void;
        IrHLExprInt of_int;
        IrHLExprFunc of_func;
        IrHLExprCall of_call;
        IrHLExprList of_list;
        IrHLExprBag of_bag;
        IrHLExprAccessor of_accessor;
        IrHLExprPairing of_pairing;
        IrHLExprTagged of_tagged;
        IrHLExprRef of_ref;
        IrHLExprPrim of_prim;
    };
    struct {
        struct {
            enum { irhl_expr_origin_def, irhl_expr_origin_expr } kind;
            union {
                AstExpr const* of_expr;
                AstDef const* of_def;
            };
        } origin;
    } anns;
};



struct IrHLDef {
    IrHLExpr body;
    struct {
        Str name;
        AstDef const* origin_def;
    } anns;
};







typedef struct CtxIrHLFromAst {
    IrHLProg ir;
} CtxIrHLFromAst;

static IrHLExpr* irHLExprCopy(IrHLExpr* const src) {
    IrHLExpr* new_expr = ·new(IrHLExpr);
    *new_expr = *src;
    return new_expr;
}

static IrHLExpr irHLExprFrom(CtxIrHLFromAst* const ctx, AstExpr* const ast_expr) {
    IrHLExpr ret_expr = (IrHLExpr) {.anns = {.origin = {.kind = irhl_expr_origin_expr, .of_expr = ast_expr}}};
    switch (ast_expr->kind) {
        default: {
            fail(str2(str("TODO: irHLExprFrom for .kind of "), uintToStr(ast_expr->kind, 1, 10)));
        } break;
    }
    return ret_expr;
}

static IrHLExpr irHLDefExpr(CtxIrHLFromAst* const ctx, AstDef* const cur_ast_def) {
    // simplest case: param-less def with no sub-defs:
    IrHLExpr body_expr = irHLExprFrom(ctx, &cur_ast_def->body);

    // def has sub-defs? rewrite into `let` lambda form:
    // from `foo bar, bar := baz` into `(bar -> foo bar) baz`. will bite us soon, then we revisit.
    if (cur_ast_def->sub_defs.len > 0) {
        ·forEach(AstDef, sub_def, cur_ast_def->sub_defs, {
            IrHLExpr expr_func = ((IrHLExpr) {.anns = {.origin = {.kind = irhl_expr_origin_def, .of_def = sub_def}},
                                              .kind = irhl_expr_func,
                                              .of_func = (IrHLExprFunc) {
                                                  .body = irHLExprCopy(&body_expr),
                                                  .params = ·make(IrHLFuncParam, 1, 1),
                                              }});
            expr_func.of_func.params.at[0] = (IrHLFuncParam) {.anns = {.name = sub_def->anns.name}};
            body_expr = ((IrHLExpr) {.anns = {.origin = {.kind = irhl_expr_origin_def, .of_def = sub_def}},
                                     .kind = irhl_expr_call,
                                     .of_call = (IrHLExprCall) {
                                         .callee = irHLExprCopy(&expr_func),
                                         .args = ·make(IrHLExpr, 1, 1),
                                     }});
            body_expr.of_call.args.at[0] = irHLDefExpr(ctx, sub_def);
        });
    }

    // def has params? turn body_expr into irhl_expr_func
    if (cur_ast_def->head.kind == ast_expr_form) {
        Uint const num_args = cur_ast_def->head.of_form.len - 1;
        body_expr = (IrHLExpr) {.anns = {.origin = {.kind = irhl_expr_origin_def, .of_def = cur_ast_def}},
                                .kind = irhl_expr_func,
                                .of_func = (IrHLExprFunc) {
                                    .body = irHLExprCopy(&body_expr),
                                    .params = ·make(IrHLFuncParam, num_args, num_args),
                                }};
        ·forEach(AstExpr, param_expr, cur_ast_def->head.of_form, {
            if (iˇparam_expr > 0) {
                ·assert(param_expr->kind == ast_expr_ident);
                IrHLFuncParam p = (IrHLFuncParam) {.anns = {.name = param_expr->of_ident}};
                body_expr.of_func.params.at[iˇparam_expr - 1] = p;
            }
        });
    }
    return body_expr;
}

static void irHLDefFrom(CtxIrHLFromAst* const ctx, AstDef* const top_def) {
    IrHLDef this_def = (IrHLDef) {.anns = {.origin_def = top_def, .name = top_def->anns.name}};
    this_def.body = irHLDefExpr(ctx, top_def);
    ·append(ctx->ir.defs, this_def);
}

static IrHLProg irHLProgFrom(Ast* const ast) {
    CtxIrHLFromAst ctx = (CtxIrHLFromAst) {.ir = (IrHLProg) {
                                               .anns = {.origin_ast = ast},
                                               .defs = ·make(IrHLDef, 0, ast->top_defs.len),
                                           }};
    ·forEach(AstDef, the_def, ast->top_defs, { irHLDefFrom(&ctx, the_def); });
    return ctx.ir;
}



static void irHLTypePrint(IrHLType const* const the_type) {
    printStr(str("@T "));
    switch (the_type->kind) {
        case irhl_type_void: {
            printStr(str("#void"));
        } break;
        case irhl_type_tag: {
            printStr(str("#tag"));
        } break;
        default: {
            fail(str2(str("TODO: irHLTypePrint for .kind of "), uintToStr(the_type->kind, 1, 10)));
        } break;
    }
}

static void irHlExprPrint(IrHLExpr const* const the_expr, Bool const is_callee_or_arg, Uint const ind) {
    return;
    switch (the_expr->kind) {
        case irhl_expr_type: {
            irHLTypePrint(the_expr->of_type.ty_value);
        } break;
        default: {
            fail(str2(str("TODO: irHlExprPrint for .kind of "), uintToStr(the_expr->kind, 1, 10)));
        } break;
    }
}

static void irHLDefPrint(IrHLDef const* const the_def) {
    printStr(the_def->anns.name);
    printStr(str(" :=\n  "));
    irHlExprPrint(&the_def->body, false, 4);
    printChr('\n');
}

static void irHLProgPrint(IrHLProg const* const the_ir_hl) {
    ·forEach(IrHLDef, the_def, the_ir_hl->defs, {
        printChr('\n');
        irHLDefPrint(the_def);
        printChr('\n');
    });
}
