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
        Bool sign;
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
    UInt size;
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
    irhl_expr_nilish,
    irhl_expr_int,
    irhl_expr_func,
    irhl_expr_call,
    irhl_expr_list,
    irhl_expr_bag,
    irhl_expr_selector,
    irhl_expr_kvpair,
    irhl_expr_tag,
    irhl_expr_tagged,
    irhl_expr_ref,
    irhl_expr_instr,
    irhl_expr_let,
} IrHLExprKind;

typedef struct IrHLExprType {
    IrHLType* ty_value;
} IrHLExprType;

typedef struct IrHLExprNilish {
    enum {
        lack,
        unit,
        blank,
    } kind;
} IrHLExprNilish;

typedef struct IrHLExprInt {
    I64 int_value;
} IrHLExprInt;

typedef struct IrHLFuncParam {
    Str name;
} IrHLFuncParam;
typedef ·SliceOf(IrHLFuncParam) IrHLFuncParams;
typedef struct IrHLExprFunc {
    IrHLFuncParams params;
    IrHLExpr* body;
} IrHLExprFunc;

typedef struct IrHLLet {
    Str name;
    IrHLExpr* expr;
} IrHLLet;
typedef ·SliceOf(IrHLLet) IrHLLets;

typedef struct IrHLExprLet {
    IrHLLets lets;
    IrHLExpr* body;
} IrHLExprLet;

typedef struct IrHLExprCall {
    IrHLExpr* callee;
    IrHLExprs args;
} IrHLExprCall;

typedef struct IrHLExprList {
    IrHLExprs items;
} IrHLExprList;

typedef struct IrHLExprBag {
    IrHLExprs items;
} IrHLExprBag;

typedef struct IrHLExprSelector {
    IrHLExpr* subj;
    IrHLExpr* member;
} IrHLExprSelector;

typedef struct IrHLExprKVPair {
    IrHLExpr* key;
    IrHLExpr* val;
} IrHLExprKVPair;
typedef ·SliceOf(IrHLExprKVPair) IrHLExprKVPairs;

typedef struct IrHLExprTag {
    Str tag_ident;
} IrHLExprTag;
typedef ·SliceOf(IrHLExprTag) IrHLExprTags;

typedef struct IrHLExprTagged {
    IrHLExpr* subj;
    IrHLExprTags tags;
} IrHLExprTagged;

typedef struct IrHLExprRef {
    Str name;
} IrHLExprRef;

typedef struct IrHLExprInstrArith {
    IrHLExpr* lhs;
    IrHLExpr* rhs;
    enum {
        irhl_arith_add,
        irhl_arith_sub,
        irhl_arith_mul,
        irhl_arith_div,
        irhl_arith_rem,
    } kind;
} IrHLExprInstrArith;

typedef struct IrHLExprInstrCmp {
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
} IrHLExprInstrCmp;

typedef struct IrHLExprInstrExtCall {
    IrHLExpr* callee;
    IrHLExprs args;
} IrHLExprInstrExtCall;

typedef struct IrHLExprInstrCase {
    IrHLType* scrut;
    IrHLExprKVPairs cases;
    IrHLExpr* default_case;
} IrHLExprInstrCase;

typedef struct IrHLExprInstrLen {
    IrHLExpr* subj;
} IrHLExprInstrLen;

typedef struct IrHLExprInstrExtFn {
} IrHLExprInstrExtFn;

typedef struct IrHLExprInstrExtVar {
} IrHLExprInstrExtVar;

typedef struct IrHLExprInstr {
    enum {
        irhl_instr_named,
        irhl_instr_arith,
        irhl_instr_cmp,
        irhl_instr_ext_call,
        irhl_instr_case,
        irhl_instr_len,
        irhl_instr_ext_fn,
        irhl_instr_ext_var,
    } kind;
    union {
        Str of_named;
        IrHLExprInstrArith of_arith;
        IrHLExprInstrCmp of_cmp;
        IrHLExprInstrExtCall of_ext_call;
        IrHLExprInstrCase of_case;
        IrHLExprInstrLen of_len;
        IrHLExprInstrExtFn of_ext_fn;
        IrHLExprInstrExtVar of_ext_var;
    };
} IrHLExprInstr;

struct IrHLExpr {
    IrHLExprKind kind;
    union {
        IrHLExprType of_type;
        IrHLExprNilish of_nilish;
        IrHLExprInt of_int;
        IrHLExprFunc of_func;
        IrHLExprCall of_call;
        IrHLExprList of_list;
        IrHLExprBag of_bag;
        IrHLExprSelector of_selector;
        IrHLExprKVPair of_kvpair;
        IrHLExprTag of_tag;
        IrHLExprTagged of_tagged;
        IrHLExprRef of_ref;
        IrHLExprLet of_let;
        IrHLExprInstr of_instr;
    };
    struct {
        struct {
            AstExpr const* ast_expr;
            AstDef const* ast_def;
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



ºBool irHLExprTagBool(IrHLExpr const* const expr) {
    if (expr->kind == irhl_expr_tag) {
        if (strEql(strL("true", 4), expr->of_tag.tag_ident))
            return ·ok(Bool, true);
        else if (strEql(strL("false", 5), expr->of_tag.tag_ident))
            return ·ok(Bool, false);
    }
    return ·none(Bool);
}

IrHLExpr* irHLBagFieldOrFail(IrHLExpr const* const expr_bag, Str const field_name, Ast const* const ast) {
    ·forEach(IrHLExpr, item, expr_bag->of_bag.items, {
        if (item->kind == irhl_expr_kvpair && item->of_kvpair.key->kind == irhl_expr_tag
            && strEql(field_name, item->of_kvpair.key->of_tag.tag_ident))
            return item->of_kvpair.val;
    });
    ·fail(astNodeMsg(str3(str("expected field named '"), field_name, str("'")), &expr_bag->anns.origin.ast_expr->node_base, ast));
    return NULL;
}

Bool irHLBagFieldBoolOrFail(IrHLExpr const* const expr_bag, Str const field_name, Ast const* const ast) {
    IrHLExpr const* const kvp_val = irHLBagFieldOrFail(expr_bag, field_name, ast);
    if (kvp_val->kind == irhl_expr_tag) {
        ºBool const maybe = irHLExprTagBool(kvp_val);
        if (maybe.ok)
            return maybe.it;
    }
    ·fail(astNodeMsg(str3(str("expected #true or #false for '"), field_name, str("'")), &kvp_val->anns.origin.ast_expr->node_base, ast));
    return false;
}

IrHLExpr* irHLExprKeep(IrHLExpr const src) {
    IrHLExpr* new_expr = ·new(IrHLExpr);
    *new_expr = src;
    return new_expr;
}

IrHLExpr* irHLExprCopy(IrHLExpr const* const src) {
    IrHLExpr* new_expr = ·new(IrHLExpr);
    *new_expr = *src;
    return new_expr;
}

IrHLExpr irHLExprTypeFrom(IrHLExpr const* const instr, Ast const* const ast) {
    AstNodeBase const* const err_node = &instr->anns.origin.ast_expr->node_base;
    if (instr->of_call.args.len < 1)
        ·fail(astNodeMsg(str("'@T' requires type #tag"), err_node, ast));
    if (instr->of_call.args.len > 2)
        ·fail(astNodeMsg(str("'@T' expects 1 or 2 args"), err_node, ast));
    IrHLExpr const* const arg_tag = &instr->of_call.args.at[0];
    IrHLExpr const* const arg_opt = (instr->of_call.args.len > 1) ? &instr->of_call.args.at[1] : NULL;
    if (arg_tag->kind != irhl_expr_tag)
        ·fail(astNodeMsg(str("'@T' requires type #tag"), err_node, ast));

    IrHLType* ty = ·new(IrHLType);
    if (strEql(arg_tag->of_tag.tag_ident, str("void"))) {
        ty->kind = irhl_type_void;

    } else if (strEql(arg_tag->of_tag.tag_ident, str("integer"))) {
        if (arg_opt == NULL || arg_opt->kind != irhl_expr_bag)
            ·fail(astNodeMsg(str("'@T #integer' requires 2nd arg to be {}"), err_node, ast));
        ty->kind = irhl_type_int;
        ty->of_int = (IrHLTypeInt) {.bit_width = 0,
                                    .anns = {.word = false, .extc = false, .sign = irHLBagFieldBoolOrFail(arg_opt, strL("signed", 6), ast)}};
        IrHLExpr const* const bit_width = irHLBagFieldOrFail(arg_opt, str("bit_width"), ast);
        if (bit_width->kind == irhl_expr_int) {
            if (bit_width->of_int.int_value < 1 || bit_width->of_int.int_value > 64)
                ·fail(astNodeMsg(str("'@T #integer' requires a mainstream #bit_width for now"), err_node, ast));
            ty->of_int.bit_width = bit_width->of_int.int_value;
        } else if (bit_width->kind == irhl_expr_tag) {
            ty->of_int.anns.word = strEql(str("word"), bit_width->of_tag.tag_ident);
            ty->of_int.anns.extc = strEql(str("extc"), bit_width->of_tag.tag_ident);
        }
        if (ty->of_int.bit_width == 0 && !(ty->of_int.anns.word || ty->of_int.anns.extc))
            ·fail(astNodeMsg(str("'@T #integer' requires a mainstream #bit_width for now"), err_node, ast));

    } else
        ·fail(astNodeMsg(str3(str("'@T' with unrecognized type #tag '"), arg_tag->of_tag.tag_ident, str("'")), err_node, ast));
    return (IrHLExpr) {.anns = instr->anns, .kind = irhl_expr_type, .of_type = (IrHLExprType) {.ty_value = ty}};
}

IrHLExpr irHLExprSelectorFromInstr(IrHLExpr const* const instr, Ast const* const ast) {
    AstNodeBase const* const err_node = &instr->anns.origin.ast_expr->node_base;
    if (instr->of_call.args.len != 2)
        ·fail(astNodeMsg(str("'@.' expects 2 args"), err_node, ast));
    return (IrHLExpr) {.kind = irhl_expr_selector,
                       .anns = instr->anns,
                       .of_selector = (IrHLExprSelector) {
                           .subj = &instr->of_call.args.at[0],
                           .member = &instr->of_call.args.at[1],
                       }};
}

IrHLExpr irHLExprKvPairFromInstr(IrHLExpr const* const instr, Ast const* const ast) {
    AstNodeBase const* const err_node = &instr->anns.origin.ast_expr->node_base;
    if (instr->of_call.args.len != 2)
        ·fail(astNodeMsg(str("'@:' expects 2 args"), err_node, ast));
    return (IrHLExpr) {.kind = irhl_expr_kvpair,
                       .anns = instr->anns,
                       .of_kvpair = (IrHLExprKVPair) {
                           .key = &instr->of_call.args.at[0],
                           .val = &instr->of_call.args.at[1],
                       }};
}

IrHLExpr irHLExprFuncFromInstr(IrHLExpr const* const instr, Ast const* const ast) {
    AstNodeBase const* const err_node = &instr->anns.origin.ast_expr->node_base;
    if (instr->of_call.args.len != 2)
        ·fail(astNodeMsg(str("'@->' expects 2 args"), err_node, ast));
    if (instr->of_call.args.at[0].kind != irhl_expr_list)
        ·fail(astNodeMsg(str("'@->' expects [list of #tags] as 1st arg"), err_node, ast));
    IrHLExprs const params = instr->of_call.args.at[0].of_list.items;
    IrHLExpr ret_expr = (IrHLExpr) {.kind = irhl_expr_func,
                                    .anns = instr->anns,
                                    .of_func = (IrHLExprFunc) {
                                        .body = &instr->of_call.args.at[1],
                                        .params = ·make(IrHLFuncParam, params.len, 0),
                                    }};
    ·forEach(IrHLExpr, param, params, {
        if (param->kind != irhl_expr_tag)
            ·fail(astNodeMsg(str("'@->' expects [list of #tags] as 1st arg"), err_node, ast));
        ret_expr.of_func.params.at[iˇparam].name = param->of_tag.tag_ident;
    });
    return ret_expr;
}

IrHLExpr irHLExprFrom(AstExpr* const ast_expr, AstDef* const ast_def, Ast const* const ast) {
    IrHLExpr ret_expr = (IrHLExpr) {.anns = {.origin = {.ast_def = ast_def, .ast_expr = ast_expr}}};
    switch (ast_expr->kind) {
        case ast_expr_lit_int: {
            ret_expr.kind = irhl_expr_int;
            ret_expr.of_int = (IrHLExprInt) {.int_value = ast_expr->of_lit_int};
        } break;
        case ast_expr_lit_str: {
            ret_expr.kind = irhl_expr_list;
            UInt const str_len = ast_expr->of_lit_str.len;
            ret_expr.of_list = (IrHLExprList) {.items = ·make(IrHLExpr, str_len, str_len)};
            for (UInt i = 0; i < str_len; i += 1)
                ret_expr.of_list.items.at[i] = (IrHLExpr) {
                    .anns = {.origin = {.ast_def = ast_def, .ast_expr = ast_expr}},
                    .kind = irhl_expr_int,
                    .of_int = (IrHLExprInt) {.int_value = ast_expr->of_lit_str.at[i]},
                };
        } break;
        case ast_expr_ident: {
            if (strEql(strL("()", 2), ast_expr->of_ident)) {
                ret_expr.kind = irhl_expr_nilish;
                ret_expr.of_nilish = (IrHLExprNilish) {.kind = unit};
            } else if (strEql(strL("_", 2), ast_expr->of_ident)) {
                ret_expr.kind = irhl_expr_nilish;
                ret_expr.of_nilish = (IrHLExprNilish) {.kind = blank};
            } else if (ast_expr->of_ident.at[0] == '@' && ast_expr->of_ident.len > 1) {
                ret_expr.kind = irhl_expr_instr;
                ret_expr.of_instr =
                    (IrHLExprInstr) {.kind = irhl_instr_named, .of_named = ·slice(U8, ast_expr->of_ident, 1, ast_expr->of_ident.len)};
            } else if (ast_expr->of_ident.at[0] == '#' && ast_expr->of_ident.len > 1) {
                ret_expr.kind = irhl_expr_tag;
                ret_expr.of_tag = (IrHLExprTag) {.tag_ident = ·slice(U8, ast_expr->of_ident, 1, ast_expr->of_ident.len)};
            } else {
                ret_expr.kind = irhl_expr_ref;
                ret_expr.of_ref = (IrHLExprRef) {.name = ast_expr->of_ident};
            }
        } break;
        case ast_expr_lit_bracket: {
            ret_expr.kind = irhl_expr_list;
            UInt const list_len = ast_expr->of_exprs.len;
            ret_expr.of_list = (IrHLExprList) {.items = ·make(IrHLExpr, list_len, list_len)};
            ·forEach(AstExpr, item_expr, ast_expr->of_exprs, {
                IrHLExpr list_item_expr = irHLExprFrom(item_expr, ast_def, ast);
                ret_expr.of_list.items.at[iˇitem_expr] = list_item_expr;
            });
        } break;
        case ast_expr_lit_braces: {
            ret_expr.kind = irhl_expr_bag;
            UInt const bag_len = ast_expr->of_exprs.len;
            ret_expr.of_bag = (IrHLExprBag) {.items = ·make(IrHLExpr, bag_len, bag_len)};
            ·forEach(AstExpr, item_expr, ast_expr->of_exprs, {
                IrHLExpr bag_item_expr = irHLExprFrom(item_expr, ast_def, ast);
                ret_expr.of_bag.items.at[iˇitem_expr] = bag_item_expr;
            });
        } break;
        case ast_expr_form: {
            if (ast_expr->of_exprs.len == 0) {
                ret_expr.kind = irhl_expr_nilish;
                ret_expr.of_nilish = (IrHLExprNilish) {.kind = lack};
                break;
            }
            if (astExprIsInstrOrTag(ast_expr, true, false, false)) {
                Str const instr_name = ast_expr->of_exprs.at[1].of_ident;
                ret_expr.kind = irhl_expr_instr;
                ret_expr.of_instr = (IrHLExprInstr) {.kind = irhl_instr_named, .of_named = instr_name};
                break;
            }
            if (astExprIsInstrOrTag(ast_expr, false, true, false)) {
                AstExpr const* const tag_along = &ast_expr->of_exprs.at[1];
                switch (tag_along->kind) {
                    case ast_expr_ident:
                        ret_expr.kind = irhl_expr_tag;
                        ret_expr.of_tag = (IrHLExprTag) {.tag_ident = tag_along->of_ident};
                        break;
                    case ast_expr_lit_str:
                        ret_expr.kind = irhl_expr_tag;
                        ret_expr.of_tag = (IrHLExprTag) {.tag_ident = tag_along->of_lit_str};
                        break;
                    case ast_expr_lit_int:
                        ret_expr.kind = irhl_expr_tag;
                        ret_expr.of_tag = (IrHLExprTag) {.tag_ident = uIntToStr(tag_along->of_lit_int, 1, 10)};
                        break;
                    default: break;
                }
                break;
            }

            ret_expr.kind = irhl_expr_call;
            UInt const num_args = ast_expr->of_exprs.len - 1;
            IrHLExpr const callee = irHLExprFrom(&ast_expr->of_exprs.at[0], ast_def, ast);
            ret_expr.of_call = (IrHLExprCall) {
                .callee = irHLExprCopy(&callee),
                .args = ·make(IrHLExpr, num_args, num_args),
            };
            ·forEach(AstExpr, arg_expr, ast_expr->of_exprs, {
                if (iˇarg_expr > 0)
                    ret_expr.of_call.args.at[iˇarg_expr - 1] = irHLExprFrom(arg_expr, ast_def, ast);
            });

            if (ret_expr.of_call.callee->kind == irhl_expr_instr && ret_expr.of_call.callee->of_instr.kind == irhl_instr_named) {
                Str const instr_name = ret_expr.of_call.callee->of_instr.of_named;
                Bool matched = false;
                if ((!matched) && strEql(strL(".", 1), instr_name)) {
                    matched = true;
                    ret_expr = irHLExprSelectorFromInstr(&ret_expr, ast);
                }
                if ((!matched) && strEql(strL(":", 1), instr_name)) {
                    matched = true;
                    ret_expr = irHLExprKvPairFromInstr(&ret_expr, ast);
                }
                if ((!matched) && strEql(strL("->", 2), instr_name)) {
                    matched = true;
                    ret_expr = irHLExprFuncFromInstr(&ret_expr, ast);
                }
            }
        } break;
        default: {
            ·fail(str2(str("TODO: irHLExprFrom for .kind of "), uIntToStr(ast_expr->kind, 1, 10)));
        } break;
    }
    return ret_expr;
}

IrHLExpr irHLDefExpr(AstDef* const cur_ast_def, Ast const* const ast) {
    IrHLExpr const body_expr = irHLExprFrom(&cur_ast_def->body, cur_ast_def, ast);
    UInt const def_count = cur_ast_def->sub_defs.len;
    if (def_count == 0)
        return body_expr;

    IrHLExpr lets = (IrHLExpr) {.kind = irhl_expr_let, .anns = {.origin = {.ast_expr = NULL, .ast_def = cur_ast_def}}};
    lets.of_let = (IrHLExprLet) {.body = irHLExprKeep(body_expr), .lets = ·make(IrHLLet, 0, def_count)};
    ·forEach(AstDef, sub_def, cur_ast_def->sub_defs, {
        ·append(lets.of_let.lets, ((IrHLLet) {
                                      .name = sub_def->name,
                                      .expr = irHLExprKeep(irHLDefExpr(sub_def, ast)),
                                  }));
    });
    return lets;
}

IrHLDef irHLDefFrom(AstDef* const top_def, Ast const* const ast) {
    IrHLDef this_def = (IrHLDef) {
        .anns = {.origin_def = top_def, .name = top_def->name},
        .body = irHLDefExpr(top_def, ast),
    };
    return this_def;
}

IrHLProg irHLProgFrom(Ast* const ast) {
    IrHLProg ret_prog = (IrHLProg) {
        .anns = {.origin_ast = ast},
        .defs = ·make(IrHLDef, 0, ast->top_defs.len),
    };
    ·forEach(AstDef, ast_top_def, ast->top_defs, {
        IrHLDef ir_hl_top_def = irHLDefFrom(ast_top_def, ast);
        ·append(ret_prog.defs, ir_hl_top_def);
    });
    return ret_prog;
}



void irHLPrintType(IrHLType const* const the_type) {
    printStr(str("@T "));
    switch (the_type->kind) {
        case irhl_type_void: {
            printStr(str("#void"));
        } break;
        case irhl_type_tag: {
            printStr(str("#tag"));
        } break;
        default: {
            ·fail(str2(str("TODO: irHLPrintType for .kind of "), uIntToStr(the_type->kind, 1, 10)));
        } break;
    }
}

void irHLPrintExpr(IrHLExpr const* const the_expr, Bool const is_callee_or_arg, UInt const ind) {
    AstExpr const* const orig_ast_expr = the_expr->anns.origin.ast_expr;
    if (orig_ast_expr != NULL)
        for (UInt i = 0; i < orig_ast_expr->anns.parensed; i += 1)
            printChr('(');

    switch (the_expr->kind) {
        case irhl_expr_type: {
            irHLPrintType(the_expr->of_type.ty_value);
        } break;
        case irhl_expr_int: {
            printStr(uIntToStr(the_expr->of_int.int_value, 1, 10));
        } break;
        case irhl_expr_nilish: {
            switch (the_expr->of_nilish.kind) {
                case blank: printChr('_'); break;
                case unit: printStr(str("()")); break;
                case lack: printStr(str("…")); break;
                default: ·fail(str2(str("TODO: irHLPrintExprNilish for .kind of "), uIntToStr(the_expr->of_nilish.kind, 1, 10))); break;
            }
        } break;
        case irhl_expr_selector: {
            irHLPrintExpr(the_expr->of_selector.subj, false, ind);
            printChr('.');
            irHLPrintExpr(the_expr->of_selector.member, false, ind);
        } break;
        case irhl_expr_tag: {
            printStr(str("#`"));
            printStr(the_expr->of_tag.tag_ident);
            printStr(str("`"));
        } break;
        case irhl_expr_ref: {
            printStr(the_expr->of_ref.name);
        } break;
        case irhl_expr_kvpair: {
            irHLPrintExpr(the_expr->of_kvpair.key, false, ind);
            printStr(str(": "));
            irHLPrintExpr(the_expr->of_kvpair.val, false, ind);
        } break;
        case irhl_expr_list: {
            printChr('[');
            ·forEach(IrHLExpr, sub_expr, the_expr->of_list.items, {
                if (iˇsub_expr != 0)
                    printStr(str(", "));
                irHLPrintExpr(sub_expr, false, ind);
            });
            printChr(']');
        } break;
        case irhl_expr_bag: {
            printStr(str("{\n"));
            UInt const ind_next = 2 + ind;
            ·forEach(IrHLExpr, sub_expr, the_expr->of_bag.items, {
                for (UInt i = 0; i < ind_next; i += 1)
                    printChr(' ');
                irHLPrintExpr(sub_expr, false, ind_next);
                printStr(str(",\n"));
            });
            for (UInt i = 0; i < ind; i += 1)
                printChr(' ');
            printChr('}');
        } break;
        case irhl_expr_let: {
            IrHLExpr bag = (IrHLExpr) {.kind = irhl_expr_bag, .of_bag = {.items = ·make(IrHLExpr, 0, the_expr->of_let.lets.len)}};
            ·forEach(IrHLLet, let, the_expr->of_let.lets, {
                ·append(bag.of_bag.items, ((IrHLExpr) {.kind = irhl_expr_kvpair,
                                                       .of_kvpair = {.key = irHLExprKeep((IrHLExpr) {
                                                                         .kind = irhl_expr_tag,
                                                                         .of_tag = (IrHLExprTag) {.tag_ident = let->name},
                                                                     }),
                                                                     .val = let->expr}}));
            });
            IrHLExpr faux = (IrHLExpr) {.kind = irhl_expr_call,
                                        .of_call = {
                                            .callee = &bag,
                                            .args = ·make(IrHLExpr, 1, 1),
                                        }};
            faux.of_call.args.at[0] = *the_expr->of_let.body;
            irHLPrintExpr(&faux, false, ind);
        } break;
        case irhl_expr_call: {
            Bool const clasp = orig_ast_expr != NULL && orig_ast_expr->anns.toks_throng;
            Bool const parens = is_callee_or_arg && (orig_ast_expr == NULL || (orig_ast_expr->anns.parensed == 0 && !clasp));
            if (parens)
                printChr('(');
            irHLPrintExpr(the_expr->of_call.callee, true, ind);
            ·forEach(IrHLExpr, sub_expr, the_expr->of_call.args, {
                if (!clasp)
                    printChr(' ');
                irHLPrintExpr(sub_expr, true, ind);
            });
            if (parens)
                printChr(')');
        } break;
        case irhl_expr_func: {
            printStr(str(" \\_"));
            ·forEach(IrHLFuncParam, param, the_expr->of_func.params, {
                if (iˇparam > 0)
                    printChr(' ');
                printStr(param->name);
            });
            printStr(str("->\n"));
            for (UInt i = 0; i < 4 + ind; i += 1)
                printChr(' ');
            irHLPrintExpr(the_expr->of_func.body, false, 4 + ind);
            printStr(str("_/ "));
        } break;
        case irhl_expr_instr: {
            if (the_expr->of_instr.kind == irhl_instr_named) {
                printChr('@');
                printStr(the_expr->of_instr.of_named);
            } else
                ·fail(str2(str("TODO: irHLPrintInstr for .kind of "), uIntToStr(the_expr->of_instr.kind, 1, 10)));
        } break;
        default: {
            ·fail(str2(str("TODO: irHLPrintExpr for .kind of "), uIntToStr(the_expr->kind, 1, 10)));
        } break;
    }

    if (orig_ast_expr != NULL)
        for (UInt i = 0; i < orig_ast_expr->anns.parensed; i += 1)
            printChr(')');
}

void irHLPrintDef(IrHLDef const* const the_def) {
    printStr(the_def->anns.name);
    printStr(str(" :=\n    "));
    irHLPrintExpr(&the_def->body, false, 4);
    printChr('\n');
}

void irHLPrintProg(IrHLProg const* const the_ir_hl) {
    ·forEach(IrHLDef, the_def, the_ir_hl->defs, {
        printChr('\n');
        irHLPrintDef(the_def);
        printChr('\n');
    });
}
