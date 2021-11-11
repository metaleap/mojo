#pragma once
#include "utils_std_mem.c"
#include "fs_io.c"
#include "at_ast.c"


UInt counter = 0;


struct IrHlDef;
typedef struct IrHlDef IrHlDef;
typedef ·ListOf(IrHlDef) IrHlDefs;

struct IrHlExpr;
typedef struct IrHlExpr IrHlExpr;
typedef ·SliceOf(IrHlExpr) IrHlExprs;




typedef struct IrHlProg {
    IrHlDefs defs;
    struct {
        Asts origin_asts;
        Strs all_tags;
    } anns;
} IrHlProg;




typedef enum IrHlExprKind {
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
} IrHlExprKind;

typedef struct IrHlExprNilish {
    enum {
        irhl_nilish_lack,
        irhl_nilish_unit,
        irhl_nilish_blank,
    } kind;
} IrHlExprNilish;

typedef struct IrHlExprInt {
    I64 int_value;
} IrHlExprInt;

typedef struct IrHlFuncParam {
    Str name;
    struct {
        AstNodeBase* origin_ast_node;
    } anns;
} IrHlFuncParam;
typedef ·SliceOf(IrHlFuncParam) IrHlFuncParams;
typedef struct IrHlExprFunc {
    IrHlFuncParams params;
    IrHlExpr* body;
    struct {
        Str qname;
        Strs free_vars;
    } anns;
} IrHlExprFunc;

typedef struct IrHlLet {
    Str name;
    IrHlExpr* expr;
    struct {
        UInt ref_count;
        Str qname;
    } anns;
} IrHlLet;
typedef ·SliceOf(IrHlLet) IrHlLets;

typedef struct IrHlExprLet {
    IrHlLets lets;
    IrHlExpr* body;
} IrHlExprLet;

typedef struct IrHlExprCall {
    IrHlExpr* callee;
    IrHlExprs args;
} IrHlExprCall;

typedef enum IrHlBagKind {
    irhl_bag_list,
    irhl_bag_tuple,
    irhl_bag_map,
    irhl_bag_struct,
} IrHlBagKind;

typedef struct IrHlExprBag {
    IrHlExprs items;
    IrHlBagKind kind;
} IrHlExprBag;

typedef struct IrHlExprSelector {
    IrHlExpr* subj;
    IrHlExpr* member;
} IrHlExprSelector;

typedef struct IrHlExprKVPair {
    IrHlExpr* key;
    IrHlExpr* val;
} IrHlExprKVPair;
typedef ·SliceOf(IrHlExprKVPair) IrHlExprKVPairs;

typedef struct IrHlExprFieldName {
    Str ident;
} IrHlExprFieldName;
typedef ·SliceOf(IrHlExprFieldName) IrHlExprFieldNames;

typedef struct IrHlExprTag {
    Str tag_ident;
    struct {
        UInt idx;
    } anns;
} IrHlExprTag;

typedef struct IrHlExprTagged {
    IrHlExpr* subj;
    IrHlExpr* tag;
} IrHlExprTagged;

typedef struct IrHlRef {
    enum {
        irhl_ref_def,
        irhl_ref_let,
        irhl_ref_expr_let,
        irhl_ref_expr_func,
        irhl_ref_func_param,
    } kind;
    union {
        IrHlDef* of_def;
        IrHlLet* of_let;
        IrHlExpr* of_expr_let;
        IrHlExpr* of_expr_func;
        IrHlFuncParam* of_func_param;
    };
} IrHlRef;
typedef ·SliceOf(IrHlRef) IrHlRefs;
typedef struct IrHlExprRef {
    Str name_or_qname;
    IrHlRefs path;
} IrHlExprRef;

typedef struct IrHlExprInstr {
    Str instr_name;
} IrHlExprInstr;

struct IrHlExpr {
    IrHlExprKind kind;
    union {
        IrHlExprNilish of_nilish;
        IrHlExprInt of_int;
        IrHlExprFunc of_func;
        IrHlExprCall of_call;
        IrHlExprBag of_bag;
        IrHlExprFieldName of_field_name;
        IrHlExprSelector of_selector;
        IrHlExprKVPair of_kvpair;
        IrHlExprTag of_tag;
        IrHlExprTagged of_tagged;
        IrHlExprRef of_ref;
        IrHlExprLet of_let;
        IrHlExprInstr of_instr;
    };
    struct {
        struct {
            AstExpr* ast_expr;
            AstDef* ast_def;
        } origin;
    } anns;
};
typedef ·Maybe(IrHlExpr) ºIrHlExpr;



struct IrHlDef {
    Str name;
    IrHlExpr* body;
    struct {
        Ast* origin_ast;
        AstDef* origin_ast_def;
        Bool is_pre_generated;
    } anns;
};




IrHlExpr* irhlExprKeep(IrHlExpr const expr) {
    IrHlExpr* new_expr = ·new(IrHlExpr, NULL);
    *new_expr = expr;
    return new_expr;
}

IrHlExpr* irhlExprCopy(IrHlExpr const* const expr) {
    IrHlExpr* new_expr = ·new(IrHlExpr, NULL);
    *new_expr = *expr;
    return new_expr;
}

IrHlExpr irhlExprInit(IrHlExprKind const kind, AstDef* const orig_ast_def, AstExpr* const orig_ast_expr) {
    return (IrHlExpr) {.anns = {.origin = {.ast_def = orig_ast_def, .ast_expr = orig_ast_expr}}, .kind = kind};
}

Bool irhlExprIsAtomic(IrHlExpr const* const expr) {
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




void irhlPrintExpr(IrHlExpr const* const expr, Bool const is_callee_or_arg, UInt const ind) {
    AstExpr const* const orig_ast_expr = expr->anns.origin.ast_expr;
    switch (expr->kind) {
        case irhl_expr_int: {
            printStr(uIntToStr(NULL, expr->of_int.int_value, 1, 10));
        } break;
        case irhl_expr_nilish: {
            switch (expr->of_nilish.kind) {
                case irhl_nilish_blank: printChr('_'); break;
                case irhl_nilish_unit: printStr(str("()")); break;
                case irhl_nilish_lack: printStr(str("!GONE!")); break;
                default:
                    ·fail(str2(NULL, str("TODO: irhlPrintExprNilish for .kind of "), uIntToStr(NULL, expr->of_nilish.kind, 1, 10)));
                    break;
            }
        } break;
        case irhl_expr_selector: {
            irhlPrintExpr(expr->of_selector.subj, false, ind);
            printChr('.');
            irhlPrintExpr(expr->of_selector.member, false, ind);
        } break;
        case irhl_expr_tag: {
            printStr(strL("#", 1));
            printStr(expr->of_tag.tag_ident);
        } break;
        case irhl_expr_ref: {
            if (expr->of_ref.path.at == NULL) {
                printChr('`');
                printStr(expr->of_ref.name_or_qname);
                printChr('`');
            } else {
                IrHlRef* const ref = ·last(expr->of_ref.path);
                switch (ref->kind) {
                    case irhl_ref_def: printStr(ref->of_def->name); break;
                    case irhl_ref_let: printStr(ref->of_let->name); break;
                    case irhl_ref_func_param: printStr(ref->of_func_param->name); break;
                    default: ·fail(str2(NULL, str("TODO: irhlPrintExprRef for .kind of "), uIntToStr(NULL, ref->kind, 1, 10))); break;
                }
            }
        } break;
        case irhl_expr_kvpair: {
            irhlPrintExpr(expr->of_kvpair.key, false, ind);
            printStr(str(": "));
            irhlPrintExpr(expr->of_kvpair.val, false, ind);
        } break;
        case irhl_expr_bag: {
            if (expr->of_bag.kind == irhl_bag_list) {
                printChr('[');
                ·forEach(IrHlExpr, sub_expr, expr->of_bag.items, {
                    if (iˇsub_expr != 0)
                        printStr(str(", "));
                    irhlPrintExpr(sub_expr, false, ind);
                });
                printChr(']');
            } else {
                printStr(str("{\n"));
                UInt const ind_next = 2 + ind;
                ·forEach(IrHlExpr, sub_expr, expr->of_bag.items, {
                    for (UInt i = 0; i < ind_next; i += 1)
                        printChr(' ');
                    irhlPrintExpr(sub_expr, false, ind_next);
                    printStr(str(",\n"));
                });
                for (UInt i = 0; i < ind; i += 1)
                    printChr(' ');
                printChr('}');
            }
        } break;
        case irhl_expr_field_name: {
            printChr('.');
            printStr(expr->of_field_name.ident);
        } break;
        case irhl_expr_let: {
            IrHlExpr bag = (IrHlExpr) {.kind = irhl_expr_bag,
                                       .anns = expr->anns,
                                       .of_bag = {.kind = irhl_bag_struct, .items = ·sliceOf(IrHlExpr, NULL, 0, expr->of_let.lets.len)}};
            ·forEach(IrHlLet, let, expr->of_let.lets, {
                ·push(bag.of_bag.items, ((IrHlExpr) {.anns = let->expr->anns,
                                                     .kind = irhl_expr_kvpair,
                                                     .of_kvpair = {.key = irhlExprKeep((IrHlExpr) {
                                                                       .kind = irhl_expr_field_name,
                                                                       .anns = let->expr->anns,
                                                                       .of_field_name = (IrHlExprFieldName) {.ident = let->name},
                                                                   }),
                                                                   .val = let->expr}}));
            });
            IrHlExpr faux = (IrHlExpr) {.kind = irhl_expr_call,
                                        .anns = expr->anns,
                                        .of_call = {
                                            .callee = &bag,
                                            .args = ·sliceOf(IrHlExpr, NULL, 1, 1),
                                        }};
            faux.of_call.args.at[0] = *expr->of_let.body;
            irhlPrintExpr(&faux, false, ind);
        } break;
        case irhl_expr_call: {
            Bool const clasp = orig_ast_expr != NULL && orig_ast_expr->anns.toks_throng;
            Bool const parens = is_callee_or_arg;
            if (parens)
                printChr('(');
            irhlPrintExpr(expr->of_call.callee, true, ind);
            ·forEach(IrHlExpr, sub_expr, expr->of_call.args, {
                if (!clasp)
                    printChr(' ');
                irhlPrintExpr(sub_expr, true, ind);
            });
            if (parens)
                printChr(')');
        } break;
        case irhl_expr_func: {
            printStr(str(" \\_"));
            ·forEach(IrHlFuncParam, param, expr->of_func.params, {
                if (iˇparam > 0)
                    printChr(' ');
                printStr(param->name);
            });
            printStr(str("->\n"));
            for (UInt i = 0; i < 4 + ind; i += 1)
                printChr(' ');
            irhlPrintExpr(expr->of_func.body, false, 4 + ind);
            printStr(str(" _"));
            for (UInt i = 0; i < expr->of_func.anns.free_vars.len; i += 1) {
                printChr(',');
                printStr(expr->of_func.anns.free_vars.at[i]);
            }
            printChr('/');
        } break;
        case irhl_expr_instr: {
            printChr('@');
            printStr(expr->of_instr.instr_name);
        } break;
        case irhl_expr_tagged: {
            printChr('(');
            irhlPrintExpr(expr->of_tagged.tag, false, ind);
            printChr(' ');
            irhlPrintExpr(expr->of_tagged.subj, false, ind);
            printChr(')');
        } break;
        default: {
            ·fail(str2(NULL, str("TODO: irhlPrintExpr for expr.kind of "), uIntToStr(NULL, expr->kind, 1, 10)));
        } break;
    }
}

void irhlPrintDef(IrHlDef const* const the_def) {
    printStr(the_def->name);
    printStr(str(" :=\n    "));
    irhlPrintExpr(the_def->body, false, 4);
    printChr('\n');
}

void irhlPrintProg(IrHlProg const* const the_ir_hl) {
    ·forEach(IrHlDef, the_def, the_ir_hl->defs, {
        printChr('\n');
        irhlPrintDef(the_def);
        printChr('\n');
    });
}




void irhlRefTarget(IrHlProg* const prog, IrHlDef* const def, IrHlExpr* ref_expr, IrHlDef** ref_to_def, IrHlExpr** ref_to_expr,
                   Ast** ref_to_ast, AstNodeBase** ref_to_node) {
    *ref_to_def = NULL;
    *ref_to_expr = NULL;
    if (ref_to_node != NULL)
        *ref_to_node = NULL;
    if (def != NULL && ref_to_ast != NULL)
        *ref_to_ast = def->anns.origin_ast;

    IrHlExpr* field = NULL;
    if (ref_expr->kind == irhl_expr_selector && ref_expr->of_selector.subj->kind == irhl_expr_ref
        && ref_expr->of_selector.member->kind == irhl_expr_tag) {
        // naive for now as we dont have selectors in user syntax in milestone 01, except the
        // structures generated after-the-fact / on-the-fly from imports
        field = ref_expr->of_selector.member;
        ref_expr = ref_expr->of_selector.subj;
    }

    if (ref_expr->kind == irhl_expr_ref) {
        IrHlRef* ref = ·last(ref_expr->of_ref.path);
    chk:
        switch (ref->kind) {
            case irhl_ref_def: {
                *ref_to_def = ref->of_def;
                if (ref->of_def->anns.origin_ast_def != NULL && ref_to_node != NULL)
                    *ref_to_node = &ref->of_def->anns.origin_ast_def->node_base;
                if (field != NULL && ref->of_def->body->kind == irhl_expr_bag)
                    ·forEach(IrHlExpr, item, ref->of_def->body->of_bag.items, {
                        if (item->kind == irhl_expr_kvpair && item->of_kvpair.key->kind == irhl_expr_field_name
                            && strEql(item->of_kvpair.key->of_field_name.ident, field->of_tag.tag_ident)
                            && item->of_kvpair.val->kind == irhl_expr_ref) {
                            ref = ·last(item->of_kvpair.val->of_ref.path);
                            field = NULL;
                            goto chk;
                        }
                    });
            } break;
            case irhl_ref_func_param: {
                if (ref_to_node != NULL)
                    *ref_to_node = ref->of_func_param->anns.origin_ast_node;
            } break;
            case irhl_ref_let: {
                *ref_to_expr = ref->of_let->expr;
                if (ref_to_node != NULL) {
                    if (ref->of_let->expr->anns.origin.ast_def != NULL)
                        *ref_to_node = &ref->of_let->expr->anns.origin.ast_def->node_base;
                    else if (ref->of_let->expr->anns.origin.ast_expr != NULL)
                        *ref_to_node = &ref->of_let->expr->anns.origin.ast_expr->node_base;
                }
            } break;
            default: ·fail(uIntToStr(NULL, ref->kind, 1, 10));
        }
    }
    if ((*ref_to_def) != NULL && ref_to_ast != NULL)
        *ref_to_ast = (*ref_to_def)->anns.origin_ast;
}

IrHlExpr* irhlExprFind(IrHlProg* const prog, IrHlDef* def, IrHlExpr* expr, AstExpr* const ast_expr) {
    IrHlExpr* ret_expr = NULL;

    switch (expr->kind) {
        case irhl_expr_selector: {
            ret_expr = irhlExprFind(prog, def, expr->of_selector.member, ast_expr);
            if (ret_expr == NULL)
                ret_expr = irhlExprFind(prog, def, expr->of_selector.subj, ast_expr);
            else
                ret_expr = expr;
        } break;
        case irhl_expr_kvpair: {
            ret_expr = irhlExprFind(prog, def, expr->of_kvpair.key, ast_expr);
            if (ret_expr == NULL)
                ret_expr = irhlExprFind(prog, def, expr->of_kvpair.val, ast_expr);
            else
                ret_expr = expr;
        } break;
        case irhl_expr_tagged: {
            ret_expr = irhlExprFind(prog, def, expr->of_tagged.tag, ast_expr);
            if (ret_expr == NULL)
                ret_expr = irhlExprFind(prog, def, expr->of_tagged.subj, ast_expr);
            else
                ret_expr = expr;
        } break;
        case irhl_expr_func: {
            ret_expr = irhlExprFind(prog, def, expr->of_func.body, ast_expr);
        } break;
        case irhl_expr_call: {
            for (UInt i = 0; (ret_expr == NULL) && i < expr->of_call.args.len; i += 1)
                ret_expr = irhlExprFind(prog, def, &expr->of_call.args.at[i], ast_expr);
            if (ret_expr == NULL)
                ret_expr = irhlExprFind(prog, def, expr->of_call.callee, ast_expr);
        } break;
        case irhl_expr_bag: {
            for (UInt i = 0; (ret_expr == NULL) && i < expr->of_bag.items.len; i += 1)
                ret_expr = irhlExprFind(prog, def, &expr->of_bag.items.at[i], ast_expr);
        } break;
        case irhl_expr_let: {
            ret_expr = irhlExprFind(prog, def, expr->of_let.body, ast_expr);
            for (UInt i = 0; (ret_expr == NULL) && i < expr->of_let.lets.len; i += 1)
                ret_expr = irhlExprFind(prog, def, expr->of_let.lets.at[i].expr, ast_expr);
        } break;
        default: break;
    }

    if ((ret_expr == NULL)
        && (expr->anns.origin.ast_expr == ast_expr
            || (expr->anns.origin.ast_expr != NULL && expr->anns.origin.ast_expr->node_base.toks_idx == ast_expr->node_base.toks_idx
                && expr->anns.origin.ast_expr->node_base.toks_len == ast_expr->node_base.toks_len)))
        ret_expr = expr;
    return ret_expr;
}

void irhlDefFind(IrHlProg* const prog, IrHlDef* def, Ast* const ast, AstDef* ast_def, AstExpr* const ast_expr, IrHlDef** found_def,
                 IrHlExpr** found_expr, Bool const force) {
    AstDef const* the_def = ast_def;
    while (the_def != NULL && the_def != def->anns.origin_ast_def)
        the_def = the_def->anns.parent_def;
    if (the_def != NULL || force) {
        *found_def = def;
        if (ast_expr != NULL)
            *found_expr = irhlExprFind(prog, def, def->body, ast_expr);
    }
}

void irhlFind(IrHlProg* const prog, Ast* const ast, AstDef* ast_def, AstExpr* const ast_expr, IrHlDef** found_def, IrHlExpr** found_expr) {
    *found_def = NULL;
    *found_expr = NULL;
    ·forEach(IrHlDef, def, prog->defs, {
        if (def->anns.origin_ast == ast)
            irhlDefFind(prog, def, ast, ast_def, ast_expr, found_def, found_expr, false);
        if ((*found_def) != NULL)
            return;
    });
}




typedef struct CtxIrHlProcessIdents {
    IrHlDef* cur_def;
    IrHlProg* prog;
} CtxIrHlProcessIdents;

#define irhl_idents_tracking_stack_capacity 48
#define irhl_func_free_vars_capacity 4
ºSrcFileIssue irhlProcessIdentsPush(CtxIrHlProcessIdents* const ctx, Strs* const names_stack, Str const name, IrHlRefs ref_stack,
                                    AstNodeBase const* const node) {
    if (name.at[0] == '_') {
        Bool all_underscores = true;
        for (UInt i = 1; all_underscores && i < name.len; i += 1)
            if (name.at[i] != '_')
                all_underscores = false;
        if (all_underscores)
            return srcFileErr(issue_irhl, str("all-underscore identifiers are reserved"),
                              ctx->cur_def->anns.origin_ast->toks.at[node->toks_idx]);
    }
    for (UInt i = 0; i < names_stack->len; i += 1)
        if (strEql(names_stack->at[i], name))
            return srcFileErr(issue_irhl, str2(NULL, str("shadowing earlier definition of "), name),
                              ctx->cur_def->anns.origin_ast->toks.at[node->toks_idx]);
    if (ctx->cur_def->anns.origin_ast != NULL)
        for (UInt i = 0; i < ctx->cur_def->anns.origin_ast->anns.all_top_def_names.len; i += 1)
            if (strEql(name, ctx->cur_def->anns.origin_ast->anns.all_top_def_names.at[i]))
                return srcFileErr(issue_irhl, str2(NULL, str("shadowing earlier definition of "), name),
                                  ctx->cur_def->anns.origin_ast->toks.at[node->toks_idx]);
    if (names_stack->len == irhl_idents_tracking_stack_capacity)
        ·fail(str("irhlProcessIdentsPush: TODO increase irhl_idents_tracking_stack_capacity"));
    names_stack->at[names_stack->len] = name;
    names_stack->len += 1;
    return ·none(SrcFileIssue);
}

ºSrcFileIssue irhlExprProcessIdents(CtxIrHlProcessIdents* const ctx, IrHlExpr* const expr, Strs names_stack, IrHlRefs ref_stack,
                                    Strs qname_stack) {
    ºSrcFileIssue issue = ·none(SrcFileIssue);
    if (qname_stack.len == irhl_idents_tracking_stack_capacity)
        ·fail(str("irhlExprProcessIdents: TODO increase irhl_idents_tracking_stack_capacity"));
    const Str str_nil = ·len0(U8);
    switch (expr->kind) {
        case irhl_expr_int:
        case irhl_expr_nilish:
        case irhl_expr_field_name:
        case irhl_expr_tag:
        case irhl_expr_instr: break;
        case irhl_expr_call: {
            ·push(qname_stack, str_nil);
            issue = irhlExprProcessIdents(ctx, expr->of_call.callee, names_stack, ref_stack, qname_stack);
            if (issue.got)
                return issue;
            ·forEach(IrHlExpr, arg, expr->of_call.args, {
                ·push(qname_stack, uIntToStr(NULL, iˇarg, 1, 16));
                issue = irhlExprProcessIdents(ctx, arg, names_stack, ref_stack, qname_stack);
                if (issue.got)
                    return issue;
                qname_stack.len -= 1;
            });
        } break;
        case irhl_expr_bag: {
            ·push(qname_stack, str_nil);
            ·forEach(IrHlExpr, item, expr->of_bag.items, {
                if (item->kind == irhl_expr_kvpair && item->of_kvpair.key->kind == irhl_expr_field_name
                    && !strEql(strL("_", 1), item->of_kvpair.key->of_field_name.ident))
                    ·push(qname_stack, item->of_kvpair.key->of_field_name.ident);
                else
                    ·push(qname_stack, uIntToStr(NULL, iˇitem, 1, 16));
                issue = irhlExprProcessIdents(ctx, item, names_stack, ref_stack, qname_stack);
                if (issue.got)
                    return issue;
                qname_stack.len -= 1;
            });
        } break;
        case irhl_expr_selector: {
            ·push(qname_stack, str_nil);
            issue = irhlExprProcessIdents(ctx, expr->of_selector.subj, names_stack, ref_stack, qname_stack);
        } break;
        case irhl_expr_kvpair: {
            ·push(qname_stack, str_nil);
            issue = irhlExprProcessIdents(ctx, expr->of_kvpair.key, names_stack, ref_stack, qname_stack);
            if (!issue.got) {
                ·push(qname_stack, str_nil);
                issue = irhlExprProcessIdents(ctx, expr->of_kvpair.val, names_stack, ref_stack, qname_stack);
            }
        } break;
        case irhl_expr_tagged: {
            ·push(qname_stack, str_nil);
            issue = irhlExprProcessIdents(ctx, expr->of_tagged.subj, names_stack, ref_stack, qname_stack);
        } break;
        case irhl_expr_func: {
            expr->of_func.anns.qname = strConcat(NULL, qname_stack, '-');
            ·forEach(IrHlFuncParam, param, expr->of_func.params, {
                issue = irhlProcessIdentsPush(ctx, &names_stack, param->name, ref_stack, param->anns.origin_ast_node);
                if (issue.got)
                    return issue;
            });
            ·push(ref_stack, ((IrHlRef) {.kind = irhl_ref_expr_func, .of_expr_func = expr}));
            ·push(qname_stack, str_nil);
            issue = irhlExprProcessIdents(ctx, expr->of_func.body, names_stack, ref_stack, qname_stack);
        } break;
        case irhl_expr_let: {
            ·forEach(IrHlLet, let, expr->of_let.lets, {
                let->anns.ref_count = 0;
                issue = irhlProcessIdentsPush(ctx, &names_stack, let->name, ref_stack, &let->expr->anns.origin.ast_def->node_base);
                if (issue.got)
                    return issue;
            });
            ·push(ref_stack, ((IrHlRef) {.kind = irhl_ref_expr_let, .of_expr_let = expr}));
            ·push(qname_stack, str_nil);
            ·forEach(IrHlLet, let, expr->of_let.lets, {
                ·push(qname_stack, let->name);
                let->anns.qname = strConcat(NULL, qname_stack, '-');
                issue = irhlExprProcessIdents(ctx, let->expr, names_stack, ref_stack, qname_stack);
                if (issue.got)
                    return issue;
                qname_stack.len -= 1;
            });
            issue = irhlExprProcessIdents(ctx, expr->of_let.body, names_stack, ref_stack, qname_stack);
        } break;
        case irhl_expr_ref: {
            Str const ident = expr->of_ref.name_or_qname;
            expr->of_ref.path.at = NULL;

            UInt ref_stack_idx = 0;
            if (expr->of_ref.path.at == NULL) // refers to top-level def?
                ·forEach(IrHlDef, def, ctx->prog->defs, {
                    if (strEql(def->name, ident)) {
                        expr->of_ref.path = ·sliceOf(IrHlRef, NULL, 1, 1);
                        expr->of_ref.path.at[0] = ((IrHlRef) {.kind = irhl_ref_def, .of_def = def});
                        break;
                    }
                });
            if (expr->of_ref.path.at == NULL) { // refers to some parent func param or parent let?
                for (UInt i = ref_stack.len - 1; i > 0 && expr->of_ref.path.at == NULL; i -= 1) { // dont need the 0th entry, its the cur_def
                    IrHlRef* ref = &ref_stack.at[i];
                    switch (ref->kind) {
                        case irhl_ref_expr_func: {
                            ·forEach(IrHlFuncParam, param, ref->of_expr_func->of_func.params, {
                                if (strEql(param->name, ident)) {
                                    expr->of_ref.path = ·sliceOf(IrHlRef, NULL, 2 + i, 0);
                                    for (UInt j = 0; j <= i; j += 1)
                                        expr->of_ref.path.at[j] = ref_stack.at[j];
                                    *·last(expr->of_ref.path) = ((IrHlRef) {.kind = irhl_ref_func_param, .of_func_param = param});
                                    ref_stack_idx = i;
                                    break;
                                }
                            });
                        } break;
                        case irhl_ref_expr_let: {
                            ·forEach(IrHlLet, let, ref->of_expr_let->of_let.lets, {
                                if (strEql(let->name, ident)) {
                                    let->anns.ref_count += 1;
                                    expr->of_ref.path = ·sliceOf(IrHlRef, NULL, 2 + i, 0);
                                    for (UInt j = 0; j <= i; j += 1)
                                        expr->of_ref.path.at[j] = ref_stack.at[j];
                                    *·last(expr->of_ref.path) = ((IrHlRef) {.kind = irhl_ref_let, .of_let = let});
                                    ref_stack_idx = i;
                                    break;
                                }
                            });
                        } break;
                        default: ·fail(str("new BUG: should be unreachable here")); break;
                    }
                }
            }
            if (expr->of_ref.path.at == NULL)
                return srcFileErr(issue_irhl, str3(NULL, str("identifier '"), expr->of_ref.name_or_qname, str("' not in scope")),
                                  ctx->cur_def->anns.origin_ast->toks.at[expr->anns.origin.ast_expr->node_base.toks_idx]);

            IrHlExprFunc* parent_fn = NULL;
            UInt parent_fn_idx = 0;
            for (UInt i = ref_stack.len - 1; (parent_fn == NULL) && (i > 0); i -= 1)
                if (ref_stack.at[i].kind == irhl_ref_expr_func) {
                    parent_fn = &ref_stack.at[i].of_expr_func->of_func;
                    parent_fn_idx = i;
                }
            if (parent_fn != NULL) {
                Bool is_free_in_parent_fn =
                    (ref_stack_idx < parent_fn_idx) && !(expr->of_ref.path.len == 1 && expr->of_ref.path.at[0].kind == irhl_ref_def);
                if (is_free_in_parent_fn)
                    ·forEach(IrHlFuncParam, param, parent_fn->params, {
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
                            parent_fn->anns.free_vars = ·sliceOf(Str, NULL, 0, irhl_func_free_vars_capacity);
                        if (parent_fn->anns.free_vars.len == irhl_func_free_vars_capacity)
                            ·fail(str("TODO: irhlExprProcessIdents increase irhl_func_free_vars_capacity"));
                        ·push(parent_fn->anns.free_vars, ident);
                    }
                }
            }
        } break;
        default: {
            ·fail(astNodeMsg(str2(NULL, str("TODO: irhlExprProcessIdents for expr.kind of "), uIntToStr(NULL, expr->kind, 1, 10)),
                             &expr->anns.origin.ast_expr->node_base, ctx->cur_def->anns.origin_ast));
        } break;
    }
    return issue;
}

void irhlProcessIdents(IrHlProg* const prog) {
    CtxIrHlProcessIdents ctx = (CtxIrHlProcessIdents) {
        .cur_def = NULL,
        .prog = prog,
    };
    Strs names_stack = ·sliceOf(Str, NULL, 0, irhl_idents_tracking_stack_capacity);
    Strs qname_stack = ·sliceOf(Str, NULL, 1, irhl_idents_tracking_stack_capacity);
    IrHlRefs ref_stack = ·sliceOf(IrHlRef, NULL, 0, irhl_idents_tracking_stack_capacity);
    ·forEach(IrHlDef, def, prog->defs, {
        ctx.cur_def = def;
        ºSrcFileIssue issue =
            irhlProcessIdentsPush(&ctx, &names_stack, def->name, ref_stack,
                                  (def->anns.origin_ast_def == NULL) ? NULL : &def->anns.origin_ast_def->anns.head_node_base);
        if (issue.got)
            ·push(def->anns.origin_ast->issues, issue.it);
    });
    ref_stack.len = 1;
    for (UInt i = 0; i < prog->defs.len; i += 1) {
        ctx.cur_def = &prog->defs.at[i];
        if (ctx.cur_def->anns.is_pre_generated)
            continue;
        qname_stack.at[0] = ctx.cur_def->name;
        ref_stack.at[0] = ((IrHlRef) {.kind = irhl_ref_def, .of_def = ctx.cur_def});
        ºSrcFileIssue issue = irhlExprProcessIdents(&ctx, ctx.cur_def->body, names_stack, ref_stack, qname_stack);
        if (issue.got)
            ·push(ctx.cur_def->anns.origin_ast->issues, issue.it);
    }
}




void prependCommonFuncs(IrHlProg* const prog) {
#define ·prependCommonFunc(def_name, num_params, arg_ref_idx)                                                                                \
    do {                                                                                                                                     \
        Str const fname = str(def_name);                                                                                                     \
        IrHlExprFunc fn = (IrHlExprFunc) {.anns = {.qname = fname, .free_vars = ·len0(Str)},                                                 \
                                          .params = ·sliceOf(IrHlFuncParam, NULL, num_params, num_params),                                   \
                                          .body = irhlExprKeep(irhlExprInit(irhl_expr_ref, NULL, NULL))};                                    \
        for (UInt i = 0; i < num_params; i += 1)                                                                                             \
            fn.params.at[i] =                                                                                                                \
                (IrHlFuncParam) {.anns = {.origin_ast_node = NULL}, .name = str2(NULL, uIntToStr(NULL, i, 1, 10), strL("a", 1))};            \
        fn.body->of_ref = (IrHlExprRef) {.name_or_qname = ·len0(U8), .path = ·sliceOf(IrHlRef, NULL, 3, 3)};                                 \
                                                                                                                                             \
        IrHlExpr def_body = irhlExprInit(irhl_expr_func, NULL, NULL);                                                                        \
        def_body.of_func = fn;                                                                                                               \
        ·append(prog->defs, ((IrHlDef) {                                                                                                     \
                                .name = fname,                                                                                               \
                                .anns = {.origin_ast = NULL, .origin_ast_def = NULL, .is_pre_generated = true},                              \
                                .body = irhlExprKeep(def_body),                                                                              \
                            }));                                                                                                             \
        IrHlExpr* ªfn = ·last(prog->defs)->body;                                                                                             \
        ªfn->of_func.body->of_ref.path.at[0] = (IrHlRef) {.kind = irhl_ref_def, .of_def = ·last(prog->defs)};                                \
        ªfn->of_func.body->of_ref.path.at[1] = (IrHlRef) {.kind = irhl_ref_expr_func, .of_expr_func = ªfn};                                  \
        ªfn->of_func.body->of_ref.path.at[2] =                                                                                               \
            (IrHlRef) {.kind = irhl_ref_func_param, .of_func_param = &ªfn->of_func.params.at[arg_ref_idx]};                                  \
    } while (0)


    ·prependCommonFunc("-i-", 1, 0);  // i := x -> x
    ·prependCommonFunc("-k-", 2, 0);  // k := x y -> x
    ·prependCommonFunc("-ki-", 2, 1); // k i := x y -> y
}




typedef struct CtxIrHlFromAsts {
    Ast* ast;
    AstDef* cur_ast_def;
    IrHlProg* prog;
} CtxIrHlFromAsts;

IrHlDef* irhlProgDef(IrHlProg* const prog, Str const module_name, Str const module_member_name) {
    IrHlDef* module = NULL;
    ·forEach(IrHlDef, def, prog->defs, {
        if (strEql(module_name, def->name)) {
            module = def;
            break;
        }
    });

    if (module != NULL && module->body->kind == irhl_expr_bag && module->body->of_bag.kind == irhl_bag_struct)
        ·forEach(IrHlExpr, bag_item, module->body->of_bag.items, {
            if (bag_item->kind == irhl_expr_kvpair && bag_item->of_kvpair.key->kind == irhl_expr_field_name
                && strEql(module_member_name, bag_item->of_kvpair.key->of_field_name.ident)
                && bag_item->of_kvpair.val->kind == irhl_expr_ref) {
                IrHlRef* func_ref = ·last(bag_item->of_kvpair.val->of_ref.path);
                if (func_ref->kind == irhl_ref_def)
                    return func_ref->of_def;
            }
        });

    return NULL;
}

#define irhl_prog_tags_capacity 128
void irhlExprTagFrom(CtxIrHlFromAsts const* const ctx, Str const tag_ident, IrHlExpr* ret_expr) {
    ret_expr->kind = irhl_expr_tag;
    ret_expr->of_tag = (IrHlExprTag) {.tag_ident = tag_ident, .anns = {.idx = ctx->prog->anns.all_tags.len}};
    for (UInt i = 0; i < ctx->prog->anns.all_tags.len && ret_expr->of_tag.anns.idx == ctx->prog->anns.all_tags.len; i += 1)
        if (strEql(ctx->prog->anns.all_tags.at[i], ret_expr->of_tag.tag_ident))
            ret_expr->of_tag.anns.idx = i;
    if (ret_expr->of_tag.anns.idx == irhl_prog_tags_capacity)
        ·fail(str("TODO: increase irhl_prog_tags_capacity"));
    else if (ret_expr->of_tag.anns.idx == ctx->prog->anns.all_tags.len)
        ·push(ctx->prog->anns.all_tags, ret_expr->of_tag.tag_ident);
}

IrHlExpr irhlExprSelectorFromInstr(CtxIrHlFromAsts const* const ctx, IrHlExpr const* const instr) {
    return (IrHlExpr) {.kind = irhl_expr_selector,
                       .anns = instr->anns,
                       .of_selector = (IrHlExprSelector) {
                           .subj = &instr->of_call.args.at[0],
                           .member = &instr->of_call.args.at[1],
                       }};
}

IrHlExpr irhlExprKvPairFromInstr(CtxIrHlFromAsts const* const ctx, IrHlExpr const* const instr) {
    return (IrHlExpr) {.kind = irhl_expr_kvpair,
                       .anns = instr->anns,
                       .of_kvpair = (IrHlExprKVPair) {
                           .key = &instr->of_call.args.at[0],
                           .val = &instr->of_call.args.at[1],
                       }};
}

ºIrHlExpr irhlExprFuncFromInstr(CtxIrHlFromAsts const* const ctx, IrHlExpr const* const instr) {
    AstNodeBase const* const err_node = &instr->anns.origin.ast_expr->node_base;
    if (instr->of_call.args.len != 2) {
        ·append(ctx->ast->issues, srcFileErr(issue_irhl, str("'@->' expects 2 args"), ctx->ast->toks.at[err_node->toks_idx]).it);
        return ·none(IrHlExpr);
    } else if (instr->of_call.args.at[0].kind != irhl_expr_bag || instr->of_call.args.at[0].of_bag.kind != irhl_bag_list) {
        ·append(ctx->ast->issues,
                srcFileErr(issue_irhl, str("'@->' expects [list of #tags] as 1st arg"), ctx->ast->toks.at[err_node->toks_idx]).it);
        return ·none(IrHlExpr);
    }
    IrHlExprs const params = instr->of_call.args.at[0].of_bag.items;
    IrHlExpr ret_expr = (IrHlExpr) {.kind = irhl_expr_func,
                                    .anns = instr->anns,
                                    .of_func = (IrHlExprFunc) {
                                        .body = &instr->of_call.args.at[1],
                                        .params = ·sliceOf(IrHlFuncParam, NULL, params.len, params.len),
                                        .anns = {.free_vars = ·len0(Str), .qname = ·len0(U8)},
                                    }};
    ·forEach(IrHlExpr, param, params, {
        if (param->kind != irhl_expr_tag) {
            ·append(ctx->ast->issues,
                    srcFileErr(issue_irhl, str("'@->' expects [list of #tags] as 1st arg"), ctx->ast->toks.at[err_node->toks_idx]).it);
            return ·none(IrHlExpr);
        }
        ret_expr.of_func.params.at[iˇparam] = ((IrHlFuncParam) {
            .name = param->of_tag.tag_ident,
            .anns = {.origin_ast_node = &param->anns.origin.ast_expr->node_base},
        });
        if (strEql(param->of_tag.tag_ident, strL("_", 1))) {
            counter += 1;
            ret_expr.of_func.params.at[iˇparam].name = str2(NULL, uIntToStr(NULL, counter, 1, 16), str("æ"));
        }
    });

    // detect if it's one of our `prependCommonFuncs`
    if (ret_expr.of_func.body->kind == irhl_expr_ref) {
        if (ret_expr.of_func.params.len == 1 && strEql(ret_expr.of_func.body->of_ref.name_or_qname, ret_expr.of_func.params.at[0].name)) {
            ret_expr = (IrHlExpr) {
                .anns = ret_expr.anns,
                .kind = irhl_expr_ref,
                .of_ref = (IrHlExprRef) {.path = ·len0(IrHlRef), .name_or_qname = str("-i-")},
            };
        } else if (ret_expr.of_func.params.len == 2) {
            Bool const ref0 = strEql(ret_expr.of_func.body->of_ref.name_or_qname, ret_expr.of_func.params.at[0].name);
            Bool const ref1 = (!ref0) && strEql(ret_expr.of_func.body->of_ref.name_or_qname, ret_expr.of_func.params.at[1].name);
            if (ref0 || ref1) {
                ret_expr = (IrHlExpr) {
                    .anns = ret_expr.anns,
                    .kind = irhl_expr_ref,
                    .of_ref = (IrHlExprRef) {.path = ·len0(IrHlRef), .name_or_qname = str(ref0 ? "-k-" : "-ki-")},
                };
            }
        }
    }
    return ·got(IrHlExpr, ret_expr);
}

IrHlExpr irhlExprFrom(CtxIrHlFromAsts* ctx, AstExpr* const ast_expr) {
    ·assert(ast_expr != NULL);
    IrHlExpr ret_expr = (IrHlExpr) {.anns = {.origin = {.ast_def = ctx->cur_ast_def, .ast_expr = ast_expr}}};
    switch (ast_expr->kind) {

        case ast_expr_lit_int: {
            ret_expr.kind = irhl_expr_int;
            ret_expr.of_int = (IrHlExprInt) {.int_value = ast_expr->of_lit_int};
        } break;

        case ast_expr_lit_str: {
            ret_expr.kind = irhl_expr_bag;
            UInt const str_len = ast_expr->of_lit_str.len;
            ret_expr.of_bag = (IrHlExprBag) {.kind = irhl_bag_list, .items = ·sliceOf(IrHlExpr, NULL, str_len, str_len)};
            for (UInt i = 0; i < str_len; i += 1)
                ret_expr.of_bag.items.at[i] = (IrHlExpr) {
                    .anns = {.origin = {.ast_def = ctx->cur_ast_def, .ast_expr = ast_expr}},
                    .kind = irhl_expr_int,
                    .of_int = (IrHlExprInt) {.int_value = ast_expr->of_lit_str.at[i]},
                };
        } break;

        case ast_expr_ident: {
            if (strEql(strL("()", 2), ast_expr->of_ident)) {
                ret_expr.kind = irhl_expr_nilish;
                ret_expr.of_nilish = (IrHlExprNilish) {.kind = irhl_nilish_unit};
            } else if (strEql(strL("_", 1), ast_expr->of_ident)) {
                ret_expr.kind = irhl_expr_nilish;
                ret_expr.of_nilish = (IrHlExprNilish) {.kind = irhl_nilish_blank};
            } else if (ast_expr->of_ident.at[0] == '@' && ast_expr->of_ident.len > 1) {
                ret_expr.kind = irhl_expr_instr;
                ret_expr.of_instr = (IrHlExprInstr) {.instr_name = ·slice(U8, ast_expr->of_ident, 1, ast_expr->of_ident.len)};
            } else if (ast_expr->of_ident.at[0] == '#' && ast_expr->of_ident.len > 1) {
                irhlExprTagFrom(ctx, ·slice(U8, ast_expr->of_ident, 1, ast_expr->of_ident.len), &ret_expr);
            } else {
                ret_expr.kind = irhl_expr_ref;
                ret_expr.of_ref = (IrHlExprRef) {.path = ·len0(IrHlRef), .name_or_qname = ast_expr->of_ident};
                for (UInt i = 0; i < ctx->ast->anns.all_top_def_names.len; i += 1) {
                    Str const top_def_name = ctx->ast->anns.all_top_def_names.at[i];
                    if (strEql(ast_expr->of_ident, top_def_name)) {
                        IrHlExpr subj = ret_expr;
                        subj.of_ref.name_or_qname = ctx->ast->anns.path_based_ident_prefix;
                        IrHlExpr member = ret_expr;
                        member.kind = irhl_expr_tag;
                        member.of_tag.tag_ident = ast_expr->of_ident;
                        ret_expr.kind = irhl_expr_selector;
                        ret_expr.of_selector.subj = irhlExprCopy(&subj);
                        ret_expr.of_selector.member = irhlExprCopy(&member);
                        break;
                    }
                }
            }
        } break;

        case ast_expr_lit_bracket: // fall through to:
        case ast_expr_lit_braces: {
            ret_expr.kind = irhl_expr_bag;
            UInt const bag_len = ast_expr->of_exprs.len;
            ret_expr.of_bag = (IrHlExprBag) {.kind = (ast_expr->kind == ast_expr_lit_bracket) ? irhl_bag_list : irhl_bag_map,
                                             .items = ·sliceOf(IrHlExpr, NULL, bag_len, bag_len)};
            ·forEach(AstExpr, item_expr, ast_expr->of_exprs, {
                IrHlExpr bag_item_expr = irhlExprFrom(ctx, item_expr);
                ret_expr.of_bag.items.at[iˇitem_expr] = bag_item_expr;
            });
            if (ret_expr.of_bag.kind != irhl_bag_list) {
                UInt kvp_count = 0;
                UInt field_names_count = 0;
                ·forEach(IrHlExpr, bag_item, ret_expr.of_bag.items, {
                    if (bag_item->kind == irhl_expr_kvpair) {
                        kvp_count += 1;
                        if (bag_item->of_kvpair.key->kind == irhl_expr_ref) {
                            Bool const was_underscore = strEql(strL("_", 1), bag_item->of_kvpair.key->of_ref.name_or_qname);
                            if (was_underscore) {
                                if (bag_item->of_kvpair.val->kind == irhl_expr_ref)
                                    bag_item->of_kvpair.key->of_ref.name_or_qname = bag_item->of_kvpair.val->of_ref.name_or_qname;
                                else
                                    bag_item->of_kvpair.key->of_ref.name_or_qname = uIntToStr(NULL, iˇbag_item, 1, 10);
                            }
                            if (bag_item->of_kvpair.key->anns.origin.ast_expr->anns.parensed == 0 || was_underscore) {
                                field_names_count += 1;
                                bag_item->of_kvpair.key->kind = irhl_expr_field_name;
                                bag_item->of_kvpair.key->of_field_name =
                                    (IrHlExprFieldName) {.ident = bag_item->of_kvpair.key->of_ref.name_or_qname};
                            }
                        }
                    }
                });
                if (kvp_count == 0)
                    ret_expr.of_bag.kind = irhl_bag_tuple;
                else if (kvp_count != bag_len)
                    ·append(ctx->ast->issues, srcFileErr(issue_irhl, str("clarify whether this is a tuple or a struct"),
                                                         ctx->ast->toks.at[ast_expr->node_base.toks_idx])
                                                  .it);
                if (field_names_count == bag_len)
                    ret_expr.of_bag.kind = irhl_bag_struct;
                else if (field_names_count > 0 && field_names_count != bag_len)
                    ·append(ctx->ast->issues, srcFileErr(issue_irhl, str("mix of field members and non-field members"),
                                                         ctx->ast->toks.at[ast_expr->node_base.toks_idx])
                                                  .it);
            }
        } break;

        case ast_expr_form: {
            if (ast_expr->of_exprs.len == 0) {
                ret_expr.kind = irhl_expr_nilish;
                ret_expr.of_nilish = (IrHlExprNilish) {.kind = irhl_nilish_lack};
                break;
            }
            ·assert(ast_expr->of_exprs.len > 1);
            if (astExprIsInstrOrTag(ast_expr, true, false, false)) {
                Str const instr_name = ast_expr->of_exprs.at[1].of_ident;
                ret_expr.kind = irhl_expr_instr;
                ret_expr.of_instr = (IrHlExprInstr) {.instr_name = instr_name};
                break;
            }
            if (astExprIsInstrOrTag(ast_expr, false, true, false)) {
                AstExpr const* const tag_along = &ast_expr->of_exprs.at[1];
                switch (tag_along->kind) {
                    case ast_expr_ident: irhlExprTagFrom(ctx, tag_along->of_ident, &ret_expr); break;
                    case ast_expr_lit_str: irhlExprTagFrom(ctx, tag_along->of_lit_str, &ret_expr); break;
                    case ast_expr_lit_int: irhlExprTagFrom(ctx, uIntToStr(NULL, tag_along->of_lit_int, 1, 10), &ret_expr); break;
                    default:
                        ·append(ctx->ast->issues,
                                srcFileErr(issue_irhl, str("unsupported tag payload"), ctx->ast->toks.at[ast_expr->node_base.toks_idx]).it);
                        break;
                }
                break;
            }
            Str const maybe_incl = astExprIsIncl(ast_expr);
            if (maybe_incl.at != 0) {
                ret_expr.kind = irhl_expr_ref;
                ret_expr.of_ref = (IrHlExprRef) {.path = ·len0(IrHlRef),
                                                 .name_or_qname = ident(NULL, relPathFromRelPath(ctx->ast->anns.src_file_path, maybe_incl))};
                break;
            }

            if (astExprIsInstrOrTag(&ast_expr->of_exprs.at[0], false, true, false)) {
                AstExpr tail_form = astExprFormSub(ast_expr, 1, ast_expr->of_exprs.len);
                ret_expr.kind = irhl_expr_tagged;
                ret_expr.of_tagged.subj =
                    irhlExprKeep(irhlExprFrom(ctx, (tail_form.of_exprs.len > 1) ? &tail_form : &tail_form.of_exprs.at[0]));
                ret_expr.of_tagged.tag = irhlExprKeep(irhlExprFrom(ctx, &ast_expr->of_exprs.at[0]));
                break;
            }

            ret_expr.kind = irhl_expr_call;
            UInt const num_args = ast_expr->of_exprs.len - 1;
            IrHlExpr const callee = irhlExprFrom(ctx, &ast_expr->of_exprs.at[0]);
            ret_expr.of_call = (IrHlExprCall) {
                .callee = irhlExprCopy(&callee),
                .args = ·sliceOf(IrHlExpr, NULL, num_args, num_args),
            };
            ·forEach(AstExpr, arg_expr, ast_expr->of_exprs, {
                if (iˇarg_expr > 0)
                    ret_expr.of_call.args.at[iˇarg_expr - 1] = irhlExprFrom(ctx, arg_expr);
            });

            if (ret_expr.of_call.callee->kind == irhl_expr_instr) {
                Str const instr_name = ret_expr.of_call.callee->of_instr.instr_name;
                Bool matched = false;
                if ((!matched) && strEql(strL(".", 1), instr_name)) {
                    matched = true;
                    if (ret_expr.of_call.args.len != 2)
                        ·append(ctx->ast->issues, srcFileErr(issue_irhl, str("'@.' expects 2 args"),
                                                             ctx->ast->toks.at[ret_expr.anns.origin.ast_expr->node_base.toks_idx])
                                                      .it);
                    else
                        ret_expr = irhlExprSelectorFromInstr(ctx, &ret_expr);
                }
                if ((!matched) && strEql(strL(":", 1), instr_name)) {
                    matched = true;
                    if (ret_expr.of_call.args.len != 2)
                        ·append(ctx->ast->issues, srcFileErr(issue_irhl, str("'@:' expects 2 args"),
                                                             ctx->ast->toks.at[ret_expr.anns.origin.ast_expr->node_base.toks_idx])
                                                      .it);
                    else
                        ret_expr = irhlExprKvPairFromInstr(ctx, &ret_expr);
                }
                if ((!matched) && strEql(strL("->", 2), instr_name)) {
                    matched = true;
                    ºIrHlExpr maybe = irhlExprFuncFromInstr(ctx, &ret_expr);
                    if (maybe.got)
                        ret_expr = maybe.it;
                }
            }
        } break;

        default: {
            ·fail(str2(NULL, str("TODO: irhlExprFrom for expr.kind of "), uIntToStr(NULL, ast_expr->kind, 1, 10)));
        } break;
    }
    return ret_expr;
}

IrHlExpr irhlDefExpr(CtxIrHlFromAsts* ctx) {
    IrHlExpr body_expr = irhlExprFrom(ctx, &ctx->cur_ast_def->body);

    UInt const def_count = ctx->cur_ast_def->sub_defs.len;
    if (def_count == 0)
        return body_expr;

    IrHlExpr lets = irhlExprInit(irhl_expr_let, ctx->cur_ast_def, NULL);
    lets.of_let = (IrHlExprLet) {.body = NULL, .lets = ·sliceOf(IrHlLet, NULL, 0, def_count)};
    AstDef* old_cur_ast_def = ctx->cur_ast_def;
    ·forEach(AstDef, sub_def, old_cur_ast_def->sub_defs, {
        ctx->cur_ast_def = sub_def;

        ·push(lets.of_let.lets, ((IrHlLet) {
                                    .name = sub_def->name,
                                    .expr = irhlExprKeep(irhlDefExpr(ctx)),
                                    .anns = {.ref_count = 0},
                                }));
    });
    ctx->cur_ast_def = old_cur_ast_def;
    if (ctx->cur_ast_def->anns.param_names.at == NULL) {
        lets.of_let.body = irhlExprKeep(body_expr);
        return lets;
    } else {
        ·assert(astExprIsFunc(&ctx->cur_ast_def->body));
        ·assert(body_expr.kind == irhl_expr_func);
        lets.of_let.body = body_expr.of_func.body;
        body_expr.of_func.body = irhlExprCopy(&lets);
        return body_expr;
    }
}

IrHlProg irhlProgFrom(Asts const asts) {
    UInt total_defs_capacity = 0;

    ·forEach(Ast, ast, asts, { total_defs_capacity += ast->anns.total_nr_of_def_toks; });

    IrHlProg ret_prog = (IrHlProg) {.anns = {.origin_asts = asts, .all_tags = ·sliceOf(Str, NULL, 0, irhl_prog_tags_capacity)},
                                    .defs = ·listOf(IrHlDef, NULL, 0, 4 + total_defs_capacity)};
    prependCommonFuncs(&ret_prog);
    ·push(ret_prog.anns.all_tags, str("false"));
    ·push(ret_prog.anns.all_tags, str("true"));
    CtxIrHlFromAsts ctx = (CtxIrHlFromAsts) {.prog = &ret_prog, .cur_ast_def = NULL, .ast = NULL};
    ·forEach(Ast, ast, asts, {
        ctx.ast = ast;
        IrHlExpr module_struct = irhlExprInit(irhl_expr_bag, NULL, NULL);
        module_struct.of_bag = ((IrHlExprBag) {.kind = irhl_bag_struct, .items = ·sliceOf(IrHlExpr, NULL, 0, ast->top_defs.len)});
        ·forEach(AstDef, ast_top_def, ast->top_defs, {
            ctx.cur_ast_def = ast_top_def;

            Str const top_def_name = str3(NULL, ast->anns.path_based_ident_prefix, strL("$", 1), ast_top_def->name);
            ·append(ret_prog.defs, ((IrHlDef) {.name = top_def_name,
                                               .anns = {.is_pre_generated = false, .origin_ast = ast, .origin_ast_def = ast_top_def},
                                               .body = irhlExprKeep(irhlDefExpr(&ctx))}));

            IrHlExpr kvp_key = irhlExprInit(irhl_expr_field_name, ast_top_def, NULL);
            kvp_key.of_field_name = ((IrHlExprFieldName) {.ident = ast_top_def->name});
            IrHlExpr kvp_val = irhlExprInit(irhl_expr_ref, ast_top_def, NULL);
            kvp_val.of_ref.path = ·sliceOf(IrHlRef, NULL, 1, 1);
            kvp_val.of_ref.path.at[0] = ((IrHlRef) {.kind = irhl_ref_def, .of_def = ·last(ret_prog.defs)});
            kvp_val.of_ref.name_or_qname = top_def_name;
            IrHlExpr kvp = irhlExprInit(irhl_expr_kvpair, ast_top_def, NULL);
            kvp.of_kvpair = ((IrHlExprKVPair) {.key = irhlExprCopy(&kvp_key), .val = irhlExprCopy(&kvp_val)});
            ·push(module_struct.of_bag.items, kvp);
        });
        ·append(ret_prog.defs, ((IrHlDef) {.name = ast->anns.path_based_ident_prefix,
                                           .anns = {.is_pre_generated = true, .origin_ast = ast, .origin_ast_def = NULL},
                                           .body = irhlExprKeep(module_struct)}));
    });
    irhlProcessIdents(&ret_prog);
    return ret_prog;
}
