#pragma once
#include "utils_and_libc_deps.c"
#include "fs_io.c"
#include "at_ast.c"


struct IrHLDef;
typedef struct IrHLDef IrHLDef;
typedef ·ListOf(IrHLDef) IrHLDefs;

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




typedef enum IrHLExprKind {
    irhl_expr_nilish = 1,
    irhl_expr_int = 2,
    irhl_expr_tag = 3,
    irhl_expr_field_name = 4,
    irhl_expr_ref = 5,
    irhl_expr_instr = 6,
    irhl_expr_func = 7,
    irhl_expr_call = 8,
    irhl_expr_bag = 9,
    irhl_expr_selector = 10,
    irhl_expr_kvpair = 11,
    irhl_expr_tagged = 12,
    irhl_expr_let = 13,
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
typedef ·ListOf(IrHLExprTag) IrHLExprTags;

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

typedef struct IrHLExprInstr {
    Str instr_name;
} IrHLExprInstr;

struct IrHLExpr {
    IrHLExprKind kind;
    union {
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
        case irhl_expr_nilish:
        case irhl_expr_int:
        case irhl_expr_tag:
        case irhl_expr_field_name:
        case irhl_expr_ref:
        case irhl_expr_instr: return true;
        default: return false;
    }
}




void irHLExprInlineRefsToNullaryAtomicDefs(IrHLExpr* const expr, IrHLDef* const cur_def, IrHLProg* const prog) {
    switch (expr->kind) {
        case irhl_expr_nilish:
        case irhl_expr_int:
        case irhl_expr_tag:
        case irhl_expr_field_name:
        case irhl_expr_instr: break;
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




typedef struct CtxIrHLProcessIdents {
    IrHLDef* cur_def;
    IrHLProg* prog;
    Bool force_re_resolve;
    Bool mark_free_vars;
} CtxIrHLProcessIdents;

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

void irHLExprProcessIdents(CtxIrHLProcessIdents* const ctx, IrHLExpr* const expr, Strs names_stack, IrHLRefs ref_stack, Strs qname_stack) {
    if (qname_stack.len == idents_tracking_stack_capacity)
        ·fail(str("irHLExprProcessIdents: TODO increase idents_tracking_stack_capacity"));
    const Str str_nil = ·len0(U8);
    switch (expr->kind) {
        case irhl_expr_int:
        case irhl_expr_nilish:
        case irhl_expr_field_name:
        case irhl_expr_tag:
        case irhl_expr_instr: break;
        case irhl_expr_call: {
            ·push(qname_stack, str_nil);
            irHLExprProcessIdents(ctx, expr->of_call.callee, names_stack, ref_stack, qname_stack);
            ·forEach(IrHLExpr, arg, expr->of_call.args, {
                ·push(qname_stack, uIntToStr(iˇarg, 1, 16));
                irHLExprProcessIdents(ctx, arg, names_stack, ref_stack, qname_stack);
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
                irHLExprProcessIdents(ctx, item, names_stack, ref_stack, qname_stack);
                qname_stack.len -= 1;
            });
        } break;
        case irhl_expr_selector: {
            ·push(qname_stack, str_nil);
            irHLExprProcessIdents(ctx, expr->of_selector.subj, names_stack, ref_stack, qname_stack);
        } break;
        case irhl_expr_kvpair: {
            ·push(qname_stack, str_nil);
            irHLExprProcessIdents(ctx, expr->of_kvpair.key, names_stack, ref_stack, qname_stack);
            ·push(qname_stack, str_nil);
            irHLExprProcessIdents(ctx, expr->of_kvpair.val, names_stack, ref_stack, qname_stack);
        } break;
        case irhl_expr_tagged: {
            ·push(qname_stack, str_nil);
            irHLExprProcessIdents(ctx, expr->of_tagged.subj, names_stack, ref_stack, qname_stack);
        } break;
        case irhl_expr_func: {
            expr->of_func.anns.qname = strConcat(qname_stack, '-');

            ·forEach(IrHLFuncParam, param, expr->of_func.params, {
                irHLProcessIdentsPush(&names_stack, param->name, ref_stack, param->anns.origin_ast_node, ctx->cur_def->anns.origin_ast);
            });
            ·push(ref_stack, ((IrHLRef) {.kind = irhl_ref_expr_func, .of_expr_func = expr}));
            ·push(qname_stack, str_nil);
            irHLExprProcessIdents(ctx, expr->of_func.body, names_stack, ref_stack, qname_stack);
        } break;
        case irhl_expr_let: {
            ·forEach(IrHLLet, let, expr->of_let.lets, {
                irHLProcessIdentsPush(&names_stack, let->name, ref_stack, &let->expr->anns.origin.ast_def->node_base,
                                      ctx->cur_def->anns.origin_ast);
            });
            ·push(ref_stack, ((IrHLRef) {.kind = irhl_ref_expr_let, .of_expr_let = expr}));
            ·push(qname_stack, str_nil);
            ·forEach(IrHLLet, let, expr->of_let.lets, {
                ·push(qname_stack, let->name);
                irHLExprProcessIdents(ctx, let->expr, names_stack, ref_stack, qname_stack);
                qname_stack.len -= 1;
            });
            irHLExprProcessIdents(ctx, expr->of_let.body, names_stack, ref_stack, qname_stack);
        } break;
        case irhl_expr_ref: {
            Str const ident = expr->of_ref.name_or_qname;
            if (ctx->force_re_resolve)
                expr->of_ref.path.at = NULL;
            if (expr->of_ref.path.at != NULL)
                break;

            if (expr->of_ref.path.at == NULL && ctx->cur_def->body->kind == irhl_expr_bag) {
                ·forEach(IrHLExpr, bag_field, ctx->cur_def->body->of_bag.items, {
                    ·assert(bag_field->kind == irhl_expr_kvpair);
                    ·assert(bag_field->of_kvpair.key->kind == irhl_expr_field_name);
                    if (strEql(ident, bag_field->of_kvpair.key->of_field_name.field_name)) {
                        expr->kind = irhl_expr_selector;
                        expr->of_selector = ((IrHLExprSelector) {
                            .subj =
                                irHLExprKeep((IrHLExpr) {.kind = irhl_expr_ref,
                                                         .anns = expr->anns,
                                                         .of_ref = {.name_or_qname = ctx->cur_def->name, .path = ·sliceOf(IrHLRef, 1, 1)}}),
                            .member = irHLExprKeep((IrHLExpr) {.kind = irhl_expr_tag, .anns = expr->anns, .of_tag = {.tag_ident = ident}})});
                        expr->of_selector.subj->of_ref.path.at[0] = ((IrHLRef) {.kind = irhl_ref_def, .of_def = ctx->cur_def});
                        return;
                    }
                });
            }
            if (expr->of_ref.path.at == NULL)
                ·forEach(IrHLDef, def, ctx->prog->defs, {
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
                                 ctx->cur_def->anns.origin_ast));

            if (!ctx->mark_free_vars)
                break;
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
                             &expr->anns.origin.ast_expr->node_base, ctx->cur_def->anns.origin_ast));
        } break;
    }
}

void irHLProcessIdents(IrHLProg* const prog, UInt const idx, Bool const force_re_resolve, Bool const mark_free_vars) {
    Strs names_stack = ·sliceOf(Str, 0, idents_tracking_stack_capacity);
    Strs qname_stack = ·sliceOf(Str, 1, idents_tracking_stack_capacity);
    IrHLRefs ref_stack = ·sliceOf(IrHLRef, 0, idents_tracking_stack_capacity);
    ·forEach(IrHLDef, def, prog->defs, {
        irHLProcessIdentsPush(&names_stack, def->name, ref_stack,
                              (def->anns.origin_ast_def == NULL) ? NULL : &def->anns.origin_ast_def->anns.head_node_base,
                              def->anns.origin_ast);
    });
    CtxIrHLProcessIdents ctx = (CtxIrHLProcessIdents) {
        .force_re_resolve = force_re_resolve,
        .mark_free_vars = mark_free_vars,
        .cur_def = NULL,
        .prog = prog,
    };
    ref_stack.len = 1;
    for (UInt i = idx; i < prog->defs.len; i += 1) {
        ctx.cur_def = &prog->defs.at[i];
        if (ctx.cur_def->anns.is_auto_generated)
            continue;
        if (idx == 0) {
            ·assert(ctx.cur_def->body->kind == irhl_expr_bag);
            ·assert(ctx.cur_def->body->of_bag.kind == irhl_bag_struct);
        }
        qname_stack.at[0] = ctx.cur_def->name;
        ref_stack.at[0] = ((IrHLRef) {.kind = irhl_ref_def, .of_def = ctx.cur_def});
        irHLExprProcessIdents(&ctx, ctx.cur_def->body, names_stack, ref_stack, qname_stack);
    }
}




typedef struct CtxIrHLLiftFuncs {
    IrHLProg* prog;
    IrHLDef* cur_def;
    Bool free_less_only;
} CtxIrHLLiftFuncs;

Bool irHLLiftFuncExprs(CtxIrHLLiftFuncs* const ctx, IrHLExpr* const expr) {
    Bool did_lift = false;
    switch (expr->kind) {
        case irhl_expr_nilish:
        case irhl_expr_int:
        case irhl_expr_tag:
        case irhl_expr_field_name:
        case irhl_expr_ref:
        case irhl_expr_instr: break;
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
                ·append(ctx->prog->defs, new_top_def);
                IrHLDef* ptr_new_top_def = ·last(ctx->prog->defs);

                IrHLExpr new_expr = (IrHLExpr) {
                    .anns = expr->anns,
                    .kind = irhl_expr_ref,
                    .of_ref = (IrHLExprRef) {.name_or_qname = expr->of_func.anns.qname, .path = ·sliceOf(IrHLRef, 1, 1)},
                };
                new_expr.of_ref.path.at[0] = (IrHLRef) {.kind = irhl_ref_def, .of_def = ptr_new_top_def};
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
                        ref_path.at[0] = (IrHLRef) {.kind = irhl_ref_def, .of_def = ptr_new_top_def};
                        ref_path.at[1] = (IrHLRef) {.kind = irhl_ref_expr_func, .of_expr_func = ptr_new_top_def->body};
                        ref_path.at[2] = (IrHLRef) {
                            .kind = irhl_ref_func_param,
                            .of_func_param = &ptr_new_top_def->body->of_func.params.at[i],
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
    UInt idx = prog->defs.len;
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
    irHLProcessIdents(prog, idx, true, false);
    irHLProgInlineRefsToNullaryAtomicDefs(prog);
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
        ·append(prog->defs, ((IrHLDef) {                                                                                                     \
                                .name = fname,                                                                                               \
                                .anns = {.origin_ast = NULL, .origin_ast_def = NULL, .is_auto_generated = true},                             \
                                .body = irHLExprKeep(def_body),                                                                              \
                            }));                                                                                                             \
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




IrHLDef* irHLProgDef(IrHLProg* const prog, Str const module_name, Str const module_member_name) {
    IrHLDef* module = NULL;
    ·forEach(IrHLDef, def, prog->defs, {
        if (strEql(module_name, def->name)) {
            module = def;
            break;
        }
    });

    if (module != NULL && module->body->kind == irhl_expr_bag && module->body->of_bag.kind == irhl_bag_struct)
        ·forEach(IrHLExpr, bag_item, module->body->of_bag.items, {
            if (bag_item->kind == irhl_expr_kvpair && bag_item->of_kvpair.key->kind == irhl_expr_field_name
                && strEql(module_member_name, bag_item->of_kvpair.key->of_field_name.field_name)
                && bag_item->of_kvpair.val->kind == irhl_expr_ref) {
                IrHLRef* func_ref = ·last(bag_item->of_kvpair.val->of_ref.path);
                if (func_ref->kind == irhl_ref_def)
                    return func_ref->of_def;
            }
        });

    return NULL;
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
                ret_expr.of_instr = (IrHLExprInstr) {.instr_name = ·slice(U8, ast_expr->of_ident, 1, ast_expr->of_ident.len)};
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
                ret_expr.of_instr = (IrHLExprInstr) {.instr_name = instr_name};
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

            if (ret_expr.of_call.callee->kind == irhl_expr_instr) {
                Str const instr_name = ret_expr.of_call.callee->of_instr.instr_name;
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

IrHLProg irHLProgFrom(Asts const asts) {
    UInt total_defs_capacity = 0;
    ·forEach(Ast, ast, asts, { total_defs_capacity += ast->anns.total_nr_of_def_toks; });

    IrHLProg ret_prog = (IrHLProg) {
        .anns = {.origin_asts = asts},
        .defs = ·listOf(IrHLDef, 0, 4 + total_defs_capacity),
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
        ·append(ret_prog.defs, ((IrHLDef) {.name = ast->anns.path_based_ident_prefix,
                                           .anns = {.is_auto_generated = false, .origin_ast = ast, .origin_ast_def = NULL},
                                           .body = irHLExprKeep(module_struct)}));
    });
    return ret_prog;
}




void irHLPrintExpr(IrHLExpr const* const the_expr, Bool const is_callee_or_arg, UInt const ind) {
    AstExpr const* const orig_ast_expr = the_expr->anns.origin.ast_expr;
    if (orig_ast_expr != NULL)
        for (UInt i = 0; i < orig_ast_expr->anns.parensed; i += 1)
            printChr('(');

    switch (the_expr->kind) {
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
            printChr('@');
            printStr(the_expr->of_instr.instr_name);
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
