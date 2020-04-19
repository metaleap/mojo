#pragma once
#include "metaleap.h"
#include "at_ast.h"


struct IrHLDef;
typedef struct IrHLDef IrHLDef;
typedef ·SliceOf(IrHLDef) IrHLDefs;

struct IrHLExpr;
typedef struct IrHLExpr IrHLExpr;
typedef ·SliceOf(IrHLExpr) IrHLExprs;

struct IrHLType;
typedef struct IrHLType IrHLType;
typedef ·SliceOf(IrHLType) IrHLTypes;




typedef struct IrHL {
    IrHLDefs defs;
    struct {
        Ast const* const origin_ast;
    } anns;
} IrHL;



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
};



typedef enum IrHLExprKind {
    irhl_expr_type,
    irhl_expr_void,
    irhl_expr_tag,
    irhl_expr_int,
    irhl_expr_func,
    irhl_expr_call,
    irhl_expr_list,
    irhl_expr_bag,
    irhl_expr_accessor,
    irhl_expr_pairing,
    irhl_expr_ref,

    irhl_expr_prim,
    irhl_expr_prim_arith,
    irhl_expr_prim_cmp,
    irhl_expr_prim_extcall,
    irhl_expr_prim_case,
    irhl_expr_prim_len,
    irhl_expr_prim_extfn,
    irhl_expr_prim_extvar,
} IrHLExprKind;

struct IrHLExpr {
    IrHLExprKind kind;
    struct {
        AstDef const* const origin_def;
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
    IrHL ir;
} CtxIrHLFromAst;

IrHL irHLFrom(Ast const* const ast) {
    CtxIrHLFromAst ctx = (CtxIrHLFromAst) {
        .ir =
            (IrHL) {
                .anns = {.origin_ast = ast},
            },
    };

    return ctx.ir;
}






void irHLTypePrint(IrHLType const* const the_type) {
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
