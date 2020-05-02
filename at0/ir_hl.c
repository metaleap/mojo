#include "utils_and_libc_deps.c"
#include "fs_io.c"
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
        Asts origin_asts;
    } anns;
} IrHLProg;



typedef enum IrHLTypeKind {
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
    irhl_expr_tag,
    irhl_expr_field_name,
    irhl_expr_ref,
    irhl_expr_instr,
    irhl_expr_func,
    irhl_expr_call,
    irhl_expr_bag,
    irhl_expr_selector,
    irhl_expr_kvpair,
    irhl_expr_tagged,
    irhl_expr_let,
} IrHLExprKind;

typedef struct IrHLExprType {
    IrHLType* ty_value;
} IrHLExprType;

typedef struct IrHLExprNilish {
    enum {
        irhl_nilish_lack,
        irhl_nilish_unit,
        irhl_nilish_blank,
    } kind;
} IrHLExprNilish;

typedef struct IrHLExprInt {
    I64 int_value;
} IrHLExprInt;

typedef struct IrHLFuncParam {
    Str name;
    struct {
        AstNodeBase* origin_ast_node;
    } anns;
} IrHLFuncParam;
typedef ·SliceOf(IrHLFuncParam) IrHLFuncParams;
typedef struct IrHLExprFunc {
    IrHLFuncParams params;
    IrHLExpr* body;
    struct {
        Str qname;
        Strs free_vars;
    } anns;
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

typedef enum IrHLBagKind {
    irhl_bag_list,
    irhl_bag_tuple,
    irhl_bag_map,
    irhl_bag_struct,
} IrHLBagKind;

typedef struct IrHLExprBag {
    IrHLExprs items;
    IrHLBagKind kind;
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

typedef struct IrHLExprFieldName {
    Str field_name;
} IrHLExprFieldName;
typedef ·SliceOf(IrHLExprFieldName) IrHLExprFieldNames;

typedef struct IrHLExprTag {
    Str tag_ident;
} IrHLExprTag;
typedef ·SliceOf(IrHLExprTag) IrHLExprTags;

typedef struct IrHLExprTagged {
    IrHLExpr* subj;
    IrHLExprTags tags;
} IrHLExprTagged;

typedef struct IrHLRef {
    enum {
        irhl_ref_def,
        irhl_ref_let,
        irhl_ref_expr_let,
        irhl_ref_expr_func,
        irhl_ref_func_param,
    } kind;
    union {
        IrHLDef* of_def;
        IrHLLet* of_let;
        IrHLExpr* of_expr_let;
        IrHLExpr* of_expr_func;
        IrHLFuncParam* of_func_param;
    };
} IrHLRef;
typedef ·SliceOf(IrHLRef) IrHLRefs;
typedef struct IrHLExprRef {
    Str name_or_qname;
    IrHLRefs path;
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

typedef struct IrHLExprInstrIndexer {
    IrHLExpr* subj;
    IrHLExpr* idx;
} IrHLExprInstrIndexer;

typedef struct IrHLExprInstrReiniter {
    IrHLExpr* subj;
    IrHLExpr* bag;
} IrHLExprInstrReiniter;

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
        irhl_instr_indexer,
        irhl_instr_reiniter,
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
        IrHLExprInstrIndexer of_indexer;
        IrHLExprInstrReiniter of_reiniter;
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
        IrHLExprBag of_bag;
        IrHLExprFieldName of_field_name;
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
            AstExpr* ast_expr;
            AstDef* ast_def;
        } origin;
        IrHLType* ty;
    } anns;
};



struct IrHLDef {
    Str name;
    IrHLExpr* body;
    struct {
        Ast* origin_ast;
        AstDef* origin_ast_def;
        Bool is_auto_generated;
    } anns;
};




IrHLExpr* irHLExprKeep(IrHLExpr const expr) {
    IrHLExpr* new_expr = ·new(IrHLExpr);
    *new_expr = expr;
    return new_expr;
}

IrHLExpr* irHLExprCopy(IrHLExpr const* const expr) {
    IrHLExpr* new_expr = ·new(IrHLExpr);
    *new_expr = *expr;
    return new_expr;
}

IrHLExpr irHLExprInit(IrHLExprKind const kind, AstDef* const orig_ast_def, AstExpr* const orig_ast_expr) {
    return (IrHLExpr) {.anns = {.ty = NULL, .origin = {.ast_def = orig_ast_def, .ast_expr = orig_ast_expr}}, .kind = kind};
}

Bool irHLExprIsAtomic(IrHLExpr const* const expr) {
    switch (expr->kind) {
        case irhl_expr_type:
        case irhl_expr_nilish:
        case irhl_expr_int:
        case irhl_expr_tag:
        case irhl_expr_field_name:
        case irhl_expr_ref: return true;
        case irhl_expr_instr: return expr->of_instr.kind == irhl_instr_named;
        default: return false;
    }
}




#define idents_tracking_stack_capacity 40
#define func_free_vars_capacity 4
void irHLProcessIdentsPush(Strs* const names_stack, Str const name, IrHLRefs ref_stack, AstNodeBase const* const node, Ast const* const ast) {
    if (name.at[0] == '_') {
        Bool all_underscores = true;
        for (UInt i = 1; all_underscores && i < name.len; i += 1)
            if (name.at[i] != '_')
                all_underscores = false;
        if (all_underscores)
            ·fail(astNodeMsg(str("all-underscore identifiers are reserved"), node, ast));
    }
    for (UInt i = 0; i < names_stack->len; i += 1)
        if (strEql(names_stack->at[i], name))
            ·fail(astNodeMsg(str2(str("shadowing earlier definition of "), name), node, ast));
    if (names_stack->len == idents_tracking_stack_capacity)
        ·fail(str("irHLProcessIdentsPush: TODO increase idents_tracking_stack_capacity"));
    names_stack->at[names_stack->len] = name;
    names_stack->len += 1;
}

void irHLExprProcessIdents(IrHLExpr* const expr, Strs names_stack, IrHLRefs ref_stack, Strs qname_stack, IrHLDef* const cur_def,
                           IrHLProg const* const prog) {
    if (qname_stack.len == idents_tracking_stack_capacity)
        ·fail(str("irHLExprProcessIdents: TODO increase idents_tracking_stack_capacity"));
    const Str str_nil = ·len0(U8);
    switch (expr->kind) {
        case irhl_expr_int:
        case irhl_expr_type:
        case irhl_expr_nilish:
        case irhl_expr_field_name:
        case irhl_expr_tag: break;
        case irhl_expr_instr: ·assert(expr->of_instr.kind == irhl_instr_named); break;
        case irhl_expr_call: {
            ·push(qname_stack, str_nil);
            irHLExprProcessIdents(expr->of_call.callee, names_stack, ref_stack, qname_stack, cur_def, prog);
            ·forEach(IrHLExpr, arg, expr->of_call.args, {
                ·push(qname_stack, uIntToStr(iˇarg, 1, 16));
                irHLExprProcessIdents(arg, names_stack, ref_stack, qname_stack, cur_def, prog);
                qname_stack.len -= 1;
            });
        } break;
        case irhl_expr_bag: {
            ·push(qname_stack, str_nil);
            ·forEach(IrHLExpr, item, expr->of_bag.items, {
                if (item->kind == irhl_expr_kvpair && item->of_kvpair.key->kind == irhl_expr_field_name
                    && !strEql(strL("_", 1), item->of_kvpair.key->of_field_name.field_name))
                    ·push(qname_stack, item->of_kvpair.key->of_field_name.field_name);
                else
                    ·push(qname_stack, uIntToStr(iˇitem, 1, 16));
                irHLExprProcessIdents(item, names_stack, ref_stack, qname_stack, cur_def, prog);
                qname_stack.len -= 1;
            });
        } break;
        case irhl_expr_selector: {
            ·push(qname_stack, str_nil);
            irHLExprProcessIdents(expr->of_selector.subj, names_stack, ref_stack, qname_stack, cur_def, prog);
        } break;
        case irhl_expr_kvpair: {
            ·push(qname_stack, str_nil);
            irHLExprProcessIdents(expr->of_kvpair.key, names_stack, ref_stack, qname_stack, cur_def, prog);
            ·push(qname_stack, str_nil);
            irHLExprProcessIdents(expr->of_kvpair.val, names_stack, ref_stack, qname_stack, cur_def, prog);
        } break;
        case irhl_expr_tagged: {
            ·push(qname_stack, str_nil);
            irHLExprProcessIdents(expr->of_tagged.subj, names_stack, ref_stack, qname_stack, cur_def, prog);
        } break;
        case irhl_expr_func: {
            expr->of_func.anns.qname = strConcat(qname_stack, '-');

            ·forEach(IrHLFuncParam, param, expr->of_func.params,
                     { irHLProcessIdentsPush(&names_stack, param->name, ref_stack, param->anns.origin_ast_node, cur_def->anns.origin_ast); });
            ·push(ref_stack, ((IrHLRef) {.kind = irhl_ref_expr_func, .of_expr_func = expr}));
            ·push(qname_stack, str_nil);
            irHLExprProcessIdents(expr->of_func.body, names_stack, ref_stack, qname_stack, cur_def, prog);
        } break;
        case irhl_expr_let: {
            ·forEach(IrHLLet, let, expr->of_let.lets, {
                irHLProcessIdentsPush(&names_stack, let->name, ref_stack, &let->expr->anns.origin.ast_def->node_base,
                                      cur_def->anns.origin_ast);
            });
            ·push(ref_stack, ((IrHLRef) {.kind = irhl_ref_expr_let, .of_expr_let = expr}));
            ·push(qname_stack, str_nil);
            ·forEach(IrHLLet, let, expr->of_let.lets, {
                ·push(qname_stack, let->name);
                irHLExprProcessIdents(let->expr, names_stack, ref_stack, qname_stack, cur_def, prog);
                qname_stack.len -= 1;
            });
            irHLExprProcessIdents(expr->of_let.body, names_stack, ref_stack, qname_stack, cur_def, prog);
        } break;
        case irhl_expr_ref: {
            Str const ident = expr->of_ref.name_or_qname;
            if (expr->of_ref.path.at != NULL)
                break;

            if (expr->of_ref.path.at == NULL) {
                ·forEach(IrHLExpr, bag_field, cur_def->body->of_bag.items, {
                    ·assert(bag_field->kind == irhl_expr_kvpair);
                    ·assert(bag_field->of_kvpair.key->kind == irhl_expr_field_name);
                    if (strEql(ident, bag_field->of_kvpair.key->of_field_name.field_name)) {
                        expr->kind = irhl_expr_selector;
                        expr->of_selector = ((IrHLExprSelector) {
                            .subj = irHLExprKeep((IrHLExpr) {.kind = irhl_expr_ref,
                                                             .anns = expr->anns,
                                                             .of_ref = {.name_or_qname = cur_def->name, .path = ·sliceOf(IrHLRef, 1, 1)}}),
                            .member = irHLExprKeep((IrHLExpr) {.kind = irhl_expr_tag, .anns = expr->anns, .of_tag = {.tag_ident = ident}})});
                        expr->of_selector.subj->of_ref.path.at[0] = ((IrHLRef) {.kind = irhl_ref_def, .of_def = cur_def});
                        return;
                    }
                });
            }
            if (expr->of_ref.path.at == NULL)
                ·forEach(IrHLDef, def, prog->defs, {
                    if (strEql(def->name, ident)) {
                        expr->of_ref.path = ·sliceOf(IrHLRef, 1, 1);
                        expr->of_ref.path.at[0] = ((IrHLRef) {.kind = irhl_ref_def, .of_def = def});
                        break;
                    }
                });
            if (expr->of_ref.path.at == NULL) {
                for (UInt i = ref_stack.len - 1; i > 0 && expr->of_ref.path.at == NULL; i -= 1) { // dont need the 0th entry, its the cur_def
                    IrHLRef* ref = &ref_stack.at[i];
                    switch (ref->kind) {
                        case irhl_ref_expr_func: {
                            ·forEach(IrHLFuncParam, param, ref->of_expr_func->of_func.params, {
                                if (strEql(param->name, ident)) {
                                    expr->of_ref.path = ·sliceOf(IrHLRef, 2 + i, 0);
                                    for (UInt j = 0; j <= i; j += 1)
                                        expr->of_ref.path.at[j] = ref_stack.at[j];
                                    *·last(expr->of_ref.path) = ((IrHLRef) {.kind = irhl_ref_func_param, .of_func_param = param});
                                    break;
                                }
                            });
                        } break;
                        case irhl_ref_expr_let: {
                            ·forEach(IrHLLet, let, ref->of_expr_let->of_let.lets, {
                                if (strEql(let->name, ident)) {
                                    expr->of_ref.path = ·sliceOf(IrHLRef, 2 + i, 0);
                                    for (UInt j = 0; j <= i; j += 1)
                                        expr->of_ref.path.at[j] = ref_stack.at[j];
                                    *·last(expr->of_ref.path) = ((IrHLRef) {.kind = irhl_ref_let, .of_let = let});
                                    break;
                                }
                            });
                        } break;
                        default: ·fail(str("new BUG: should be unreachable here")); break;
                    }
                }
            }
            if (expr->of_ref.path.at == NULL)
                ·fail(astNodeMsg(str3(str("identifier '"), expr->of_ref.name_or_qname, str("' not in scope")),
                                 (&expr->anns.origin.ast_expr == NULL) ? NULL : &expr->anns.origin.ast_expr->node_base,
                                 cur_def->anns.origin_ast));

            IrHLExprFunc* parent_fn = NULL;
            for (UInt i = ref_stack.len - 1; (parent_fn == NULL) && (i > 0); i -= 1)
                if (ref_stack.at[i].kind == irhl_ref_expr_func)
                    parent_fn = &ref_stack.at[i].of_expr_func->of_func;
            if (parent_fn != NULL) {
                Bool is_free_in_parent_fn = !(expr->of_ref.path.len == 1 && expr->of_ref.path.at[0].kind == irhl_ref_def);
                if (is_free_in_parent_fn)
                    ·forEach(IrHLFuncParam, param, parent_fn->params, {
                        if (strEql(param->name, ident)) {
                            is_free_in_parent_fn = false;
                            break;
                        }
                    });
                if (is_free_in_parent_fn) {
                    Bool already_known = false;
                    for (UInt i = 0; (!already_known) && i < parent_fn->anns.free_vars.len; i += 1)
                        if (strEql(ident, parent_fn->anns.free_vars.at[i]))
                            already_known = true;
                    if (!already_known) {
                        if (parent_fn->anns.free_vars.at == NULL)
                            parent_fn->anns.free_vars = ·sliceOf(Str, 0, func_free_vars_capacity);
                        if (parent_fn->anns.free_vars.len == func_free_vars_capacity)
                            ·fail(str("TODO: irHLExprProcessIdents increase func_free_vars_capacity"));
                        ·push(parent_fn->anns.free_vars, ident);
                    }
                }
            }
        } break;
        default: {
            ·fail(astNodeMsg(str2(str("TODO: irHLExprProcessIdents for expr.kind of "), uIntToStr(expr->kind, 1, 10)),
                             &expr->anns.origin.ast_expr->node_base, cur_def->anns.origin_ast));
        } break;
    }
}

void irHLProcessIdents(IrHLProg const* const prog) {
    Strs names_stack = ·sliceOf(Str, 0, idents_tracking_stack_capacity);
    Strs qname_stack = ·sliceOf(Str, 1, idents_tracking_stack_capacity);
    IrHLRefs ref_stack = ·sliceOf(IrHLRef, 0, idents_tracking_stack_capacity);
    ·forEach(IrHLDef, def, prog->defs, {
        irHLProcessIdentsPush(&names_stack, def->name, ref_stack,
                              (def->anns.origin_ast_def == NULL) ? NULL : &def->anns.origin_ast_def->anns.head_node_base,
                              def->anns.origin_ast);
    });
    ref_stack.len = 1;
    ·forEach(IrHLDef, def, prog->defs, {
        if (def->anns.is_auto_generated)
            continue;
        ·assert(def->body->kind == irhl_expr_bag);
        ·assert(def->body->of_bag.kind == irhl_bag_struct);
        qname_stack.at[0] = def->name;
        ref_stack.at[0] = ((IrHLRef) {.kind = irhl_ref_def, .of_def = def});
        irHLExprProcessIdents(def->body, names_stack, ref_stack, qname_stack, def, prog);
    });
}




void irHLExprInlineRefsToNullaryAtomicDefs(IrHLExpr* const expr, IrHLDef* const cur_def, IrHLProg* const prog) {
    switch (expr->kind) {
        case irhl_expr_type:
        case irhl_expr_nilish:
        case irhl_expr_int:
        case irhl_expr_tag:
        case irhl_expr_field_name: break;
        case irhl_expr_instr: ·assert(expr->of_instr.kind == irhl_instr_named); break;
        case irhl_expr_call: {
            irHLExprInlineRefsToNullaryAtomicDefs(expr->of_call.callee, cur_def, prog);
            ·forEach(IrHLExpr, arg, expr->of_call.args, { irHLExprInlineRefsToNullaryAtomicDefs(arg, cur_def, prog); });
        } break;
        case irhl_expr_bag: {
            ·forEach(IrHLExpr, item, expr->of_bag.items, { irHLExprInlineRefsToNullaryAtomicDefs(item, cur_def, prog); });
        } break;
        case irhl_expr_selector: {
            irHLExprInlineRefsToNullaryAtomicDefs(expr->of_selector.subj, cur_def, prog);
            irHLExprInlineRefsToNullaryAtomicDefs(expr->of_selector.member, cur_def, prog);
        } break;
        case irhl_expr_kvpair: {
            irHLExprInlineRefsToNullaryAtomicDefs(expr->of_kvpair.key, cur_def, prog);
            irHLExprInlineRefsToNullaryAtomicDefs(expr->of_kvpair.val, cur_def, prog);
        } break;
        case irhl_expr_tagged: {
            irHLExprInlineRefsToNullaryAtomicDefs(expr->of_tagged.subj, cur_def, prog);
        } break;
        case irhl_expr_let: {
            ·forEach(IrHLLet, let, expr->of_let.lets, { irHLExprInlineRefsToNullaryAtomicDefs(let->expr, cur_def, prog); });
            irHLExprInlineRefsToNullaryAtomicDefs(expr->of_let.body, cur_def, prog);
        } break;
        case irhl_expr_func: {
            irHLExprInlineRefsToNullaryAtomicDefs(expr->of_func.body, cur_def, prog);
        } break;
        case irhl_expr_ref: {
            IrHLRef* ref = ·last(expr->of_ref.path);
            Str ref_name = ·len0(U8);
            Bool did_rewrite = false;
            switch (ref->kind) {
                case irhl_ref_def: {
                    if (irHLExprIsAtomic(ref->of_def->body)) {
                        ref_name = ref->of_def->name;
                        *expr = *ref->of_def->body;
                        did_rewrite = true;
                    }
                } break;
                case irhl_ref_let: {
                    if (irHLExprIsAtomic(ref->of_let->expr)) {
                        ref_name = ref->of_let->name;
                        *expr = *ref->of_let->expr;
                        did_rewrite = true;
                    }
                } break;
                default: break;
            }
            if (did_rewrite)
                irHLExprInlineRefsToNullaryAtomicDefs(expr, cur_def, prog);
        } break;
        default:
            ·fail(astNodeMsg(str2(str("TODO: irHLExprInlineRefsToNullaryAtomicDefs for expr.kind of "), uIntToStr(expr->kind, 1, 10)),
                             &expr->anns.origin.ast_expr->node_base, cur_def->anns.origin_ast));
    }
}

void irHLProgInlineRefsToNullaryAtomicDefs(IrHLProg* const prog) {
    ·forEach(IrHLDef, def, prog->defs, { irHLExprInlineRefsToNullaryAtomicDefs(def->body, def, prog); });
}




typedef struct CtxIrHLLiftFuncs {
    IrHLProg* prog;
    IrHLDef* cur_def;
    Bool free_less_only;
} CtxIrHLLiftFuncs;

Bool irHLLiftFuncExprs(CtxIrHLLiftFuncs* const ctx, IrHLExpr* const expr) {
    Bool did_lift = false;
    switch (expr->kind) {
        case irhl_expr_type:
        case irhl_expr_nilish:
        case irhl_expr_int:
        case irhl_expr_tag:
        case irhl_expr_field_name:
        case irhl_expr_ref: break;
        case irhl_expr_instr: ·assert(expr->of_instr.kind == irhl_instr_named); break;
        case irhl_expr_call: {
            did_lift |= irHLLiftFuncExprs(ctx, expr->of_call.callee);
            ·forEach(IrHLExpr, arg, expr->of_call.args, { did_lift |= irHLLiftFuncExprs(ctx, arg); });
        } break;
        case irhl_expr_bag: {
            ·forEach(IrHLExpr, item, expr->of_bag.items, { did_lift |= irHLLiftFuncExprs(ctx, item); });
        } break;
        case irhl_expr_selector: {
            did_lift |= irHLLiftFuncExprs(ctx, expr->of_selector.subj);
            did_lift |= irHLLiftFuncExprs(ctx, expr->of_selector.member);
        } break;
        case irhl_expr_kvpair: {
            did_lift |= irHLLiftFuncExprs(ctx, expr->of_kvpair.key);
            did_lift |= irHLLiftFuncExprs(ctx, expr->of_kvpair.val);
        } break;
        case irhl_expr_tagged: {
            did_lift |= irHLLiftFuncExprs(ctx, expr->of_tagged.subj);
        } break;
        case irhl_expr_let: {
            ·forEach(IrHLLet, let, expr->of_let.lets, { did_lift |= irHLLiftFuncExprs(ctx, let->expr); });
            did_lift |= irHLLiftFuncExprs(ctx, expr->of_let.body);
        } break;
        case irhl_expr_func: {
            did_lift |= irHLLiftFuncExprs(ctx, expr->of_func.body);
            if (expr == ctx->cur_def->body)
                break;

            UInt n_free_vars = expr->of_func.anns.free_vars.len;
            if (n_free_vars == 0 || !ctx->free_less_only) {
                did_lift = true;

                IrHLExpr new_top_def_body = *expr;
                if (n_free_vars != 0) {
                    UInt n_params_new = n_free_vars + new_top_def_body.of_func.params.len;
                    IrHLFuncParams new_params = ·sliceOf(IrHLFuncParam, n_params_new, n_params_new);
                    for (UInt i = 0; i < n_free_vars; i += 1)
                        new_params.at[i] = (IrHLFuncParam) {.name = expr->of_func.anns.free_vars.at[i], .anns = {.origin_ast_node = NULL}};
                    ·forEach(IrHLFuncParam, param, new_top_def_body.of_func.params, { new_params.at[n_free_vars + iˇparam] = *param; });
                    new_top_def_body.of_func.params = new_params;
                    new_top_def_body.of_func.anns.free_vars.len = 0;
                }
                IrHLDef new_top_def =
                    (IrHLDef) {.anns = {.origin_ast_def = expr->anns.origin.ast_def, .origin_ast = NULL, .is_auto_generated = false},
                               .name = expr->of_func.anns.qname,
                               .body = irHLExprCopy(&new_top_def_body)};
                ·push(ctx->prog->defs, new_top_def);

                IrHLExpr new_expr = (IrHLExpr) {
                    .anns = expr->anns,
                    .kind = irhl_expr_ref,
                    .of_ref = (IrHLExprRef) {.name_or_qname = expr->of_func.anns.qname, .path = ·sliceOf(IrHLRef, 1, 1)},
                };
                new_expr.of_ref.path.at[0] = (IrHLRef) {.kind = irhl_ref_def, .of_def = ·last(ctx->prog->defs)};
                if (n_free_vars == 0)
                    *expr = new_expr;
                else {
                    IrHLExpr new_call = (IrHLExpr) {
                        .anns = expr->anns,
                        .kind = irhl_expr_call,
                        .of_call = {.callee = irHLExprKeep(new_expr), .args = ·sliceOf(IrHLExpr, n_free_vars, n_free_vars)},
                    };
                    for (UInt i = 0; i < n_free_vars; i += 1) {
                        IrHLRefs ref_path = ·sliceOf(IrHLRef, 3, 3);
                        ref_path.at[0] = (IrHLRef) {.kind = irhl_ref_def, .of_def = ·last(ctx->prog->defs)};
                        ref_path.at[1] = (IrHLRef) {.kind = irhl_ref_expr_func, .of_expr_func = ·last(ctx->prog->defs)->body};
                        ref_path.at[2] = (IrHLRef) {
                            .kind = irhl_ref_func_param,
                            .of_func_param = &·last(ctx->prog->defs)->body->of_func.params.at[i],
                        };
                        new_call.of_call.args.at[i] = irHLExprInit(irhl_expr_ref, ctx->cur_def->anns.origin_ast_def, NULL);
                        new_call.of_call.args.at[i].of_ref =
                            (IrHLExprRef) {.path = ref_path, .name_or_qname = expr->of_func.anns.free_vars.at[i]};
                    }
                    *expr = new_call;
                }
            }
        } break;
        default:
            ·fail(astNodeMsg(str2(str("TODO: irHLLiftFuncExprs for expr.kind of "), uIntToStr(expr->kind, 1, 10)),
                             &expr->anns.origin.ast_expr->node_base, ctx->cur_def->anns.origin_ast));
    }
    return did_lift;
}

void irHLProgLiftFuncExprs(IrHLProg* const prog) {
    irHLProgInlineRefsToNullaryAtomicDefs(prog);
    CtxIrHLLiftFuncs ctx = (CtxIrHLLiftFuncs) {.free_less_only = true, .prog = prog};
    Bool did_lift_some = false;
    while (true) {
        did_lift_some = false;
        ·forEach(IrHLDef, def, prog->defs, {
            ctx.cur_def = def;
            did_lift_some |= irHLLiftFuncExprs(&ctx, def->body);
        });
        if (!did_lift_some) {
            if (ctx.free_less_only)
                ctx.free_less_only = false;
            else
                break;
        }
    }
    irHLProgInlineRefsToNullaryAtomicDefs(prog);
}




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
    if (strEql(arg_tag->of_tag.tag_ident, str("integer"))) {
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
    if (instr->of_call.args.at[0].kind != irhl_expr_bag || instr->of_call.args.at[0].of_bag.kind != irhl_bag_list)
        ·fail(astNodeMsg(str("'@->' expects [list of #tags] as 1st arg"), err_node, ast));
    IrHLExprs const params = instr->of_call.args.at[0].of_bag.items;
    IrHLExpr ret_expr = (IrHLExpr) {.kind = irhl_expr_func,
                                    .anns = instr->anns,
                                    .of_func = (IrHLExprFunc) {
                                        .body = &instr->of_call.args.at[1],
                                        .params = ·sliceOf(IrHLFuncParam, params.len, params.len),
                                        .anns = {.free_vars = ·len0(Str), .qname = ·len0(U8)},
                                    }};
    ·forEach(IrHLExpr, param, params, {
        if (param->kind != irhl_expr_tag)
            ·fail(astNodeMsg(str("'@->' expects [list of #tags] as 1st arg"), err_node, ast));
        ret_expr.of_func.params.at[iˇparam] = ((IrHLFuncParam) {
            .name = param->of_tag.tag_ident,
            .anns = {.origin_ast_node = &param->anns.origin.ast_expr->node_base},
        });
        if (strEql(param->of_tag.tag_ident, strL("_", 1))) {
            counter += 1;
            ret_expr.of_func.params.at[iˇparam].name = str2(uIntToStr(counter, 1, 16), str("æ"));
        }
    });

    // detect if it's one of our `prependCommonFuncs`
    if (ret_expr.of_func.body->kind == irhl_expr_ref) {
        if (ret_expr.of_func.params.len == 1 && strEql(ret_expr.of_func.body->of_ref.name_or_qname, ret_expr.of_func.params.at[0].name)) {
            ret_expr = (IrHLExpr) {
                .anns = ret_expr.anns,
                .kind = irhl_expr_ref,
                .of_ref = (IrHLExprRef) {.path = ·len0(IrHLRef), .name_or_qname = str("-i-")},
            };
        } else if (ret_expr.of_func.params.len == 2) {
            Bool const ref0 = strEql(ret_expr.of_func.body->of_ref.name_or_qname, ret_expr.of_func.params.at[0].name);
            Bool const ref1 = (!ref0) && strEql(ret_expr.of_func.body->of_ref.name_or_qname, ret_expr.of_func.params.at[1].name);
            if (ref0 || ref1) {
                ret_expr = (IrHLExpr) {
                    .anns = ret_expr.anns,
                    .kind = irhl_expr_ref,
                    .of_ref = (IrHLExprRef) {.path = ·len0(IrHLRef), .name_or_qname = str(ref0 ? "-k-" : "-ki-")},
                };
            }
        }
    }
    return ret_expr;
}

IrHLExpr irHLExprFrom(AstExpr* const ast_expr, AstDef* const ast_def, Ast const* const ast) {
    IrHLExpr ret_expr = (IrHLExpr) {.anns = {.ty = NULL, .origin = {.ast_def = ast_def, .ast_expr = ast_expr}}};
    switch (ast_expr->kind) {

        case ast_expr_lit_int: {
            ret_expr.kind = irhl_expr_int;
            ret_expr.of_int = (IrHLExprInt) {.int_value = ast_expr->of_lit_int};
        } break;

        case ast_expr_lit_str: {
            ret_expr.kind = irhl_expr_bag;
            UInt const str_len = ast_expr->of_lit_str.len;
            ret_expr.of_bag = (IrHLExprBag) {.kind = irhl_bag_list, .items = ·sliceOf(IrHLExpr, str_len, str_len)};
            for (UInt i = 0; i < str_len; i += 1)
                ret_expr.of_bag.items.at[i] = (IrHLExpr) {
                    .anns = {.ty = NULL, .origin = {.ast_def = ast_def, .ast_expr = ast_expr}},
                    .kind = irhl_expr_int,
                    .of_int = (IrHLExprInt) {.int_value = ast_expr->of_lit_str.at[i]},
                };
        } break;

        case ast_expr_ident: {
            if (strEql(strL("()", 2), ast_expr->of_ident)) {
                ret_expr.kind = irhl_expr_nilish;
                ret_expr.of_nilish = (IrHLExprNilish) {.kind = irhl_nilish_unit};
            } else if (strEql(strL("_", 2), ast_expr->of_ident)) {
                ret_expr.kind = irhl_expr_nilish;
                ret_expr.of_nilish = (IrHLExprNilish) {.kind = irhl_nilish_blank};
            } else if (ast_expr->of_ident.at[0] == '@' && ast_expr->of_ident.len > 1) {
                ret_expr.kind = irhl_expr_instr;
                ret_expr.of_instr =
                    (IrHLExprInstr) {.kind = irhl_instr_named, .of_named = ·slice(U8, ast_expr->of_ident, 1, ast_expr->of_ident.len)};
            } else if (ast_expr->of_ident.at[0] == '#' && ast_expr->of_ident.len > 1) {
                ret_expr.kind = irhl_expr_tag;
                ret_expr.of_tag = (IrHLExprTag) {.tag_ident = ·slice(U8, ast_expr->of_ident, 1, ast_expr->of_ident.len)};
            } else {
                ret_expr.kind = irhl_expr_ref;
                ret_expr.of_ref = (IrHLExprRef) {.path = ·len0(IrHLRef), .name_or_qname = ast_expr->of_ident};
            }
        } break;

        case ast_expr_lit_bracket: // fall through to:
        case ast_expr_lit_braces: {
            ret_expr.kind = irhl_expr_bag;
            UInt const bag_len = ast_expr->of_exprs.len;
            ret_expr.of_bag = (IrHLExprBag) {.kind = (ast_expr->kind == ast_expr_lit_bracket) ? irhl_bag_list : irhl_bag_map,
                                             .items = ·sliceOf(IrHLExpr, bag_len, bag_len)};
            ·forEach(AstExpr, item_expr, ast_expr->of_exprs, {
                IrHLExpr bag_item_expr = irHLExprFrom(item_expr, ast_def, ast);
                ret_expr.of_bag.items.at[iˇitem_expr] = bag_item_expr;
            });
            if (ret_expr.of_bag.kind != irhl_bag_list) {
                UInt kvp_count = 0;
                UInt field_names_count = 0;
                ·forEach(IrHLExpr, bag_item, ret_expr.of_bag.items, {
                    if (bag_item->kind == irhl_expr_kvpair) {
                        kvp_count += 1;
                        if (bag_item->of_kvpair.key->kind == irhl_expr_ref) {
                            Bool const was_underscore = strEql(strL("_", 1), bag_item->of_kvpair.key->of_ref.name_or_qname);
                            if (was_underscore) {
                                if (bag_item->of_kvpair.val->kind == irhl_expr_ref)
                                    bag_item->of_kvpair.key->of_ref.name_or_qname = bag_item->of_kvpair.val->of_ref.name_or_qname;
                                else
                                    bag_item->of_kvpair.key->of_ref.name_or_qname = uIntToStr(iˇbag_item, 1, 10);
                            }
                            if (bag_item->of_kvpair.key->anns.origin.ast_expr->anns.parensed == 0 || was_underscore) {
                                field_names_count += 1;
                                bag_item->of_kvpair.key->kind = irhl_expr_field_name;
                                bag_item->of_kvpair.key->of_field_name =
                                    (IrHLExprFieldName) {.field_name = bag_item->of_kvpair.key->of_ref.name_or_qname};
                            }
                        }
                    }
                });
                if (kvp_count == 0)
                    ret_expr.of_bag.kind = irhl_bag_tuple;
                else if (kvp_count != bag_len)
                    ·fail(astNodeMsg(str("clarify whether this is a tuple or a struct"), &ast_expr->node_base, ast));
                if (field_names_count == bag_len)
                    ret_expr.of_bag.kind = irhl_bag_struct;
                else if (field_names_count > 0 && field_names_count != bag_len)
                    ·fail(astNodeMsg(str("mix of field members and non-field members"), &ast_expr->node_base, ast));
            }
        } break;

        case ast_expr_form: {
            if (ast_expr->of_exprs.len == 0) {
                ret_expr.kind = irhl_expr_nilish;
                ret_expr.of_nilish = (IrHLExprNilish) {.kind = irhl_nilish_lack};
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
                    default: ·fail(astNodeMsg(str("unsupported tag payload"), &ast_expr->node_base, ast)); break;
                }
                break;
            }
            Str const maybe_incl = astExprIsIncl(ast_expr);
            if (maybe_incl.at != 0) {
                ret_expr.kind = irhl_expr_ref;
                ret_expr.of_ref =
                    (IrHLExprRef) {.path = ·len0(IrHLRef), .name_or_qname = ident(relPathFromRelPath(ast->anns.src_file_path, maybe_incl))};
                break;
            }

            ret_expr.kind = irhl_expr_call;
            UInt const num_args = ast_expr->of_exprs.len - 1;
            IrHLExpr const callee = irHLExprFrom(&ast_expr->of_exprs.at[0], ast_def, ast);
            ret_expr.of_call = (IrHLExprCall) {
                .callee = irHLExprCopy(&callee),
                .args = ·sliceOf(IrHLExpr, num_args, num_args),
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
            ·fail(str2(str("TODO: irHLExprFrom for expr.kind of "), uIntToStr(ast_expr->kind, 1, 10)));
        } break;
    }
    return ret_expr;
}

IrHLExpr irHLDefExpr(AstDef* const cur_ast_def, Ast const* const ast) {
    IrHLExpr body_expr = irHLExprFrom(&cur_ast_def->body, cur_ast_def, ast);

    UInt const def_count = cur_ast_def->sub_defs.len;
    if (def_count == 0)
        return body_expr;

    IrHLExpr lets = irHLExprInit(irhl_expr_let, cur_ast_def, NULL);
    lets.of_let = (IrHLExprLet) {.body = NULL, .lets = ·sliceOf(IrHLLet, 0, def_count)};
    ·forEach(AstDef, sub_def, cur_ast_def->sub_defs, {
        ·push(lets.of_let.lets, ((IrHLLet) {
                                    .name = sub_def->name,
                                    .expr = irHLExprKeep(irHLDefExpr(sub_def, ast)),
                                }));
    });
    if (cur_ast_def->anns.param_names.at == NULL) {
        lets.of_let.body = irHLExprKeep(body_expr);
        return lets;
    } else {
        ·assert(astExprIsFunc(&cur_ast_def->body));
        ·assert(body_expr.kind == irhl_expr_func);
        lets.of_let.body = body_expr.of_func.body;
        body_expr.of_func.body = irHLExprCopy(&lets);
        return body_expr;
    }
}

void prependCommonFuncs(IrHLProg* const prog) {
#define ·prependCommonFunc(def_name, num_params, arg_ref_idx)                                                                                \
    do {                                                                                                                                     \
        Str const fname = str(def_name);                                                                                                     \
        IrHLExprFunc fn = (IrHLExprFunc) {.anns = {.qname = fname, .free_vars = ·len0(Str)},                                                 \
                                          .params = ·sliceOf(IrHLFuncParam, num_params, num_params),                                         \
                                          .body = irHLExprKeep(irHLExprInit(irhl_expr_ref, NULL, NULL))};                                    \
        for (UInt i = 0; i < num_params; i += 1)                                                                                             \
            fn.params.at[i] = (IrHLFuncParam) {.anns = {.origin_ast_node = NULL}, .name = uIntToStr(i, 1, 10)};                              \
        fn.body->of_ref = (IrHLExprRef) {.name_or_qname = ·len0(U8), .path = ·sliceOf(IrHLRef, 3, 3)};                                       \
                                                                                                                                             \
        IrHLExpr def_body = irHLExprInit(irhl_expr_func, NULL, NULL);                                                                        \
        def_body.of_func = fn;                                                                                                               \
        ·push(prog->defs, ((IrHLDef) {                                                                                                       \
                              .name = fname,                                                                                                 \
                              .anns = {.origin_ast = NULL, .origin_ast_def = NULL, .is_auto_generated = true},                               \
                              .body = irHLExprKeep(def_body),                                                                                \
                          }));                                                                                                               \
        IrHLExpr* ªfn = ·last(prog->defs)->body;                                                                                             \
        ªfn->of_func.body->of_ref.path.at[0] = (IrHLRef) {.kind = irhl_ref_def, .of_def = ·last(prog->defs)};                                \
        ªfn->of_func.body->of_ref.path.at[1] = (IrHLRef) {.kind = irhl_ref_expr_func, .of_expr_func = ªfn};                                  \
        ªfn->of_func.body->of_ref.path.at[2] =                                                                                               \
            (IrHLRef) {.kind = irhl_ref_func_param, .of_func_param = &ªfn->of_func.params.at[arg_ref_idx]};                                  \
    } while (0)


    ·prependCommonFunc("-i-", 1, 0);  // i := x -> x
    ·prependCommonFunc("-k-", 2, 0);  // k := x y -> x
    ·prependCommonFunc("-ki-", 2, 1); // k i := x y -> y
}

IrHLProg irHLProgFrom(Asts const asts) {
    UInt total_defs_capacity = 0;
    ·forEach(Ast, ast, asts, { total_defs_capacity += ast->anns.total_nr_of_def_toks; });

    IrHLProg ret_prog = (IrHLProg) {
        .anns = {.origin_asts = asts},
        .defs = ·sliceOf(IrHLDef, 0, 3 + total_defs_capacity),
    };
    prependCommonFuncs(&ret_prog);
    ·forEach(Ast, ast, asts, {
        IrHLExpr module_struct = irHLExprInit(irhl_expr_bag, NULL, NULL);
        module_struct.of_bag = ((IrHLExprBag) {.kind = irhl_bag_struct, .items = ·sliceOf(IrHLExpr, 0, ast->top_defs.len)});
        ·forEach(AstDef, ast_top_def, ast->top_defs, {
            IrHLExpr key = irHLExprInit(irhl_expr_field_name, ast_top_def, NULL);
            key.of_field_name = ((IrHLExprFieldName) {.field_name = ast_top_def->name});
            IrHLExpr kvp = irHLExprInit(irhl_expr_kvpair, ast_top_def, NULL);
            kvp.of_kvpair = ((IrHLExprKVPair) {.key = irHLExprKeep(key), .val = irHLExprKeep(irHLDefExpr(ast_top_def, ast))});
            ·push(module_struct.of_bag.items, kvp);
        });
        ·push(ret_prog.defs, ((IrHLDef) {.name = ast->anns.path_based_ident_prefix,
                                         .anns = {.is_auto_generated = false, .origin_ast = ast, .origin_ast_def = NULL},
                                         .body = irHLExprKeep(module_struct)}));
    });
    return ret_prog;
}




void irHLPrintType(IrHLType const* const the_type) {
    printStr(str("@T "));
    switch (the_type->kind) {
        case irhl_type_tag: {
            printStr(str("#tag"));
        } break;
        default: {
            ·fail(str2(str("TODO: irHLPrintType for ty.kind of "), uIntToStr(the_type->kind, 1, 10)));
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
                case irhl_nilish_blank: printChr('_'); break;
                case irhl_nilish_unit: printStr(str("()")); break;
                case irhl_nilish_lack: printStr(str("!GONE!")); break;
                default: ·fail(str2(str("TODO: irHLPrintExprNilish for .kind of "), uIntToStr(the_expr->of_nilish.kind, 1, 10))); break;
            }
        } break;
        case irhl_expr_selector: {
            irHLPrintExpr(the_expr->of_selector.subj, false, ind);
            printChr('.');
            irHLPrintExpr(the_expr->of_selector.member, false, ind);
        } break;
        case irhl_expr_tag: {
            printStr(strL("#", 1));
            printStr(the_expr->of_tag.tag_ident);
        } break;
        case irhl_expr_ref: {
            if (the_expr->of_ref.path.at == NULL) {
                printChr('`');
                printStr(the_expr->of_ref.name_or_qname);
                printChr('`');
            } else {
                IrHLRef* const ref = ·last(the_expr->of_ref.path);
                switch (ref->kind) {
                    case irhl_ref_def: printStr(ref->of_def->name); break;
                    case irhl_ref_let: printStr(ref->of_let->name); break;
                    case irhl_ref_func_param: printStr(ref->of_func_param->name); break;
                    default: ·fail(str2(str("TODO: irHLPrintExprRef for .kind of "), uIntToStr(ref->kind, 1, 10))); break;
                }
            }
        } break;
        case irhl_expr_kvpair: {
            irHLPrintExpr(the_expr->of_kvpair.key, false, ind);
            printStr(str(": "));
            irHLPrintExpr(the_expr->of_kvpair.val, false, ind);
        } break;
        case irhl_expr_bag: {
            if (the_expr->of_bag.kind == irhl_bag_list) {
                printChr('[');
                ·forEach(IrHLExpr, sub_expr, the_expr->of_bag.items, {
                    if (iˇsub_expr != 0)
                        printStr(str(", "));
                    irHLPrintExpr(sub_expr, false, ind);
                });
                printChr(']');
            } else {
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
            }
        } break;
        case irhl_expr_field_name: {
            printChr('.');
            printStr(the_expr->of_field_name.field_name);
        } break;
        case irhl_expr_let: {
            IrHLExpr bag = (IrHLExpr) {.kind = irhl_expr_bag,
                                       .anns = the_expr->anns,
                                       .of_bag = {.kind = irhl_bag_struct, .items = ·sliceOf(IrHLExpr, 0, the_expr->of_let.lets.len)}};
            ·forEach(IrHLLet, let, the_expr->of_let.lets, {
                ·push(bag.of_bag.items, ((IrHLExpr) {.anns = let->expr->anns,
                                                     .kind = irhl_expr_kvpair,
                                                     .of_kvpair = {.key = irHLExprKeep((IrHLExpr) {
                                                                       .kind = irhl_expr_field_name,
                                                                       .anns = let->expr->anns,
                                                                       .of_field_name = (IrHLExprFieldName) {.field_name = let->name},
                                                                   }),
                                                                   .val = let->expr}}));
            });
            IrHLExpr faux = (IrHLExpr) {.kind = irhl_expr_call,
                                        .anns = the_expr->anns,
                                        .of_call = {
                                            .callee = &bag,
                                            .args = ·sliceOf(IrHLExpr, 1, 1),
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
            ·fail(str2(str("TODO: irHLPrintExpr for expr.kind of "), uIntToStr(the_expr->kind, 1, 10)));
        } break;
    }

    if (orig_ast_expr != NULL)
        for (UInt i = 0; i < orig_ast_expr->anns.parensed; i += 1)
            printChr(')');
}

void irHLPrintDef(IrHLDef const* const the_def) {
    printStr(the_def->name);
    printStr(str(" :=\n    "));
    irHLPrintExpr(the_def->body, false, 4);
    printChr('\n');
}

void irHLPrintProg(IrHLProg const* const the_ir_hl) {
    ·forEach(IrHLDef, the_def, the_ir_hl->defs, {
        printChr('\n');
        irHLPrintDef(the_def);
        printChr('\n');
    });
}
