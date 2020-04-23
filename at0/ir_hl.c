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
    struct {
    } anns;
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
    irhl_expr_pairing,
    irhl_expr_tag,
    irhl_expr_tagged,
    irhl_expr_ref,
    irhl_expr_instr,
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
    IrHLExprs items;
} IrHLExprList;

typedef struct IrHLExprBag {
    IrHLExprs items;
} IrHLExprBag;

typedef struct IrHLExprSelector {
    IrHLExpr* subj;
    IrHLExpr* member;
} IrHLExprSelector;

typedef struct IrHLExprPairing {
    IrHLExpr* key;
    IrHLExpr* val;
} IrHLExprPairing;
typedef ·SliceOf(IrHLExprPairing) IrHLExprPairings;

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
    IrHLExprPairings cases;
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
        IrHLExprPairing of_pairing;
        IrHLExprTag of_tag;
        IrHLExprTagged of_tagged;
        IrHLExprRef of_ref;
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




IrHLExpr* irHLExprCopy(IrHLExpr const* const src) {
    IrHLExpr* new_expr = ·new(IrHLExpr);
    *new_expr = *src;
    return new_expr;
}

IrHLExpr irHLExprFrom(AstExpr* const ast_expr, AstDef* const ast_def) {
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
                IrHLExpr list_item_expr = irHLExprFrom(item_expr, ast_def);
                ret_expr.of_list.items.at[iˇitem_expr] = list_item_expr;
            });
        } break;
        case ast_expr_lit_braces: {
            ret_expr.kind = irhl_expr_bag;
            UInt const bag_len = ast_expr->of_exprs.len;
            ret_expr.of_bag = (IrHLExprBag) {.items = ·make(IrHLExpr, bag_len, bag_len)};
            ·forEach(AstExpr, item_expr, ast_expr->of_exprs, {
                IrHLExpr bag_item_expr = irHLExprFrom(item_expr, ast_def);
                ret_expr.of_bag.items.at[iˇitem_expr] = bag_item_expr;
            });
        } break;
        case ast_expr_form: {
            if (ast_expr->of_exprs.len == 0) {
                ret_expr.kind = irhl_expr_nilish;
                ret_expr.of_nilish = (IrHLExprNilish) {.kind = lack};
                break;
            }
            if (astExprIsInstrOrTag(ast_expr, false, true, true)) {
                ret_expr.kind = irhl_expr_tag;
                ret_expr.of_tag = (IrHLExprTag) {.tag_ident = ast_expr->of_exprs.at[1].of_ident};
                break;
            } else if (astExprIsInstrOrTag(ast_expr, true, false, false)) {
                Str const instr_name = ast_expr->of_exprs.at[1].of_ident;
                ret_expr.kind = irhl_expr_instr;
                ret_expr.of_instr = (IrHLExprInstr) {.kind = irhl_instr_named, .of_named = instr_name};
                break;
            }

            ret_expr.kind = irhl_expr_call;
            UInt const num_args = ast_expr->of_exprs.len - 1;
            IrHLExpr const callee = irHLExprFrom(&ast_expr->of_exprs.at[0], ast_def);
            ret_expr.of_call = (IrHLExprCall) {
                .callee = irHLExprCopy(&callee),
                .args = ·make(IrHLExpr, num_args, num_args),
            };
            ·forEach(AstExpr, arg_expr, ast_expr->of_exprs, {
                if (iˇarg_expr > 0)
                    ret_expr.of_call.args.at[iˇarg_expr - 1] = irHLExprFrom(arg_expr, ast_def);
            });
        } break;
        default: {
            ·fail(str2(str("TODO: irHLExprFrom for .kind of "), uIntToStr(ast_expr->kind, 1, 10)));
        } break;
    }
    return ret_expr;
}

IrHLExpr irHLDefExpr(AstDef* const cur_ast_def) {
    IrHLExpr body_expr = irHLExprFrom(&cur_ast_def->body, cur_ast_def);

    // def has sub-defs? rewrite into `let` lambda form:
    // from `foo bar, bar := baz` into `(bar -> foo bar) baz`.  naive approach
    // of doing so will bite us soon, then we revisit (if the bite really hurts).
    if (cur_ast_def->sub_defs.len > 0) {
        ·forEach(AstDef, sub_def, cur_ast_def->sub_defs, {
            IrHLExpr expr_func = ((IrHLExpr) {.anns = {.origin = {.ast_expr = NULL, .ast_def = sub_def}},
                                              .kind = irhl_expr_func,
                                              .of_func = (IrHLExprFunc) {
                                                  .body = irHLExprCopy(&body_expr),
                                                  .params = ·make(IrHLFuncParam, 1, 1),
                                              }});
            expr_func.of_func.params.at[0] = (IrHLFuncParam) {.anns = {.name = sub_def->name}};
            body_expr = ((IrHLExpr) {.anns = {.origin = {.ast_expr = &cur_ast_def->body, .ast_def = sub_def}},
                                     .kind = irhl_expr_call,
                                     .of_call = (IrHLExprCall) {
                                         .callee = irHLExprCopy(&expr_func),
                                         .args = ·make(IrHLExpr, 1, 1),
                                     }});
            body_expr.of_call.args.at[0] = irHLDefExpr(sub_def);
        });
    }
    return body_expr;
}

IrHLDef irHLDefFrom(AstDef* const top_def) {
    IrHLDef this_def = (IrHLDef) {
        .anns = {.origin_def = top_def, .name = top_def->name},
        .body = irHLDefExpr(top_def),
    };
    return this_def;
}

IrHLProg irHLProgFrom(Ast* const ast) {
    IrHLProg ret_prog = (IrHLProg) {
        .anns = {.origin_ast = ast},
        .defs = ·make(IrHLDef, 0, ast->top_defs.len),
    };
    ·forEach(AstDef, ast_top_def, ast->top_defs, {
        IrHLDef ir_hl_top_def = irHLDefFrom(ast_top_def);
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
                default: printStr(str("…?!?!?!…")); break;
            }
        } break;
        case irhl_expr_tag: {
            printChr('#');
            printStr(the_expr->of_tag.tag_ident);
        } break;
        case irhl_expr_ref: {
            printStr(the_expr->of_ref.name);
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
                printStr(param->anns.name);
            });
            printStr(str("->\n"));
            for (UInt i = 0; i < ind; i += 1)
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
    printStr(str(" :=\n  "));
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
