#include "metaleap.c"
#include "at_ast.c"


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

typedef struct IrHLExprFunc {
    Strs params;
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
        AstExpr const* const origin_expr;
    } anns;
};



struct IrHLDef {
    IrHLExpr body;
    struct {
        Str name;
        AstDef const* const origin_def;
    } anns;
};







typedef struct CtxIrHLFromAst {
    IrHLProg ir;
} CtxIrHLFromAst;

static IrHLProg irHLProgFrom(Ast const* const ast) {
    CtxIrHLFromAst ctx = (CtxIrHLFromAst) {
        .ir =
            (IrHLProg) {
                .anns = {.origin_ast = ast},
            },
    };

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

static void irHlExprPrint(IrHLExpr const* const the_expr) {
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
    printStr(str(" := "));

    irHlExprPrint(&the_def->body);
}

static void irHLProgPrint(IrHLProg const* const the_ir_hl) {
    ·forEach(IrHLDef, the_def, the_ir_hl->defs, { irHLDefPrint(the_def); });
}
