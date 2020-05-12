#pragma once
#include "utils_and_libc_deps.c"
#include "ir_hl.c"

typedef enum IrLlInstrKind {
    irll_instr_invalid,
    irll_instr_extern,
    irll_instr_branch,
    irll_instr_branchcase,
    irll_instr_bag_item_idx,
    irll_instr_bag_item_name,
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
} IrLlInstrKind;

typedef enum IrLlExprKind {
    irll_expr_invalid,
    irll_expr_void,
    irll_expr_int,
    irll_expr_tag,
    irll_expr_bag,
    irll_expr_ref_param,
    irll_expr_ref_func,
    irll_expr_instr,
    irll_expr_call,
} IrLlExprKind;

struct IrLlExprCall;
typedef struct IrLlExprCall IrLlExprCall;
struct IrLlExprBag;
typedef struct IrLlExprBag IrLlExprBag;

typedef struct IrLlExpr {
    IrLlExprKind kind;
    union {
        I64 of_int;
        UInt of_tag;
        IrLlExprBag* of_bag;
        UInt of_ref_param;
        UInt of_ref_func;
        IrLlInstrKind of_instr;
        IrLlExprCall* of_call;
    };
} IrLlExpr;
typedef ·SliceOf(IrLlExpr) IrLlExprs;

struct IrLlExprBag {
    IrLlExprs items;
    Strs field_names;
};

struct IrLlExprCall {
    IrLlExpr callee;
    IrLlExprs args;
    UInt is_closure;
};

typedef struct IrLlFunc {
    UInt num_params;
    IrLlExpr body;
    struct {
        IrHlDef* origin;
    } anns;
} IrLlFunc;
typedef ·ListOf(IrLlFunc) IrLlFuncs;

typedef struct IrLlProg {
    IrLlFuncs funcs;
    struct {
        IrHlProg* origin;
    } anns;
} IrLlProg;




typedef ·ListOf(IrHlExprTag) IrHlExprTags;
typedef struct CtxIrLlFromHL {
    IrLlProg prog;
    IrHlExprTags tags;
} CtxIrLlFromHL;

UInt irllFuncFrom(CtxIrLlFromHL* const ctx, IrHlDef* const hl_def);

IrLlExpr irllExprFrom(CtxIrLlFromHL* const ctx, IrHlExpr* const hl_expr) {
    IrLlExpr ret_expr = (IrLlExpr) {.kind = irll_expr_invalid};
    Str fail_msg = ·len0(U8);
    switch (hl_expr->kind) {

        case irhl_expr_int: {
            ret_expr.kind = irll_expr_int;
            ret_expr.of_int = hl_expr->of_int.int_value;
        } break;

        case irhl_expr_let: {
            // named lets inside ignored until they get pulled in on demand
            ret_expr = irllExprFrom(ctx, hl_expr->of_let.body);
        } break;

        case irhl_expr_nilish: {
            if (hl_expr->of_nilish.kind != irhl_nilish_unit)
                fail_msg = str("encountered a non-void nilish");
            else
                ret_expr.kind = irll_expr_void;
        } break;

        case irhl_expr_tag: {
            UInt tag_idx = ctx->tags.len;
            ·forEach(IrHlExprTag, tag, ctx->tags, {
                if (strEql(tag->tag_ident, hl_expr->of_tag.tag_ident)) {
                    tag_idx = iˇtag;
                    break;
                }
            });
            if (tag_idx == ctx->tags.len)
                ·append(ctx->tags, hl_expr->of_tag);

            ret_expr.kind = irll_expr_tag;
            ret_expr.of_tag = tag_idx;
        } break;

        case irhl_expr_bag: {
            ret_expr.kind = irll_expr_bag;
            ret_expr.of_bag = ·new(IrLlExprBag);
            Bool const is_struct = (hl_expr->of_bag.kind == irhl_bag_struct);
            ret_expr.of_bag->field_names = (!is_struct) ? ·len0(Str) : ·sliceOf(Str, hl_expr->of_bag.items.len, 0);
            ret_expr.of_bag->items = ·sliceOf(IrLlExpr, hl_expr->of_bag.items.len, 0);
            ·forEach(IrHlExpr, item, hl_expr->of_bag.items, {
                if (!is_struct)
                    ret_expr.of_bag->items.at[iˇitem] = irllExprFrom(ctx, item);
                else {
                    ·assert(item->kind == irhl_expr_kvpair);
                    ·assert(item->of_kvpair.key->kind == irhl_expr_field_name);
                    ret_expr.of_bag->field_names.at[iˇitem] = item->of_kvpair.key->of_field_name.field_name;
                    ret_expr.of_bag->items.at[iˇitem] = irllExprFrom(ctx, item->of_kvpair.val);
                }
            });
        } break;

        case irhl_expr_instr: {
            ret_expr.kind = irll_expr_instr;
            ret_expr.of_instr = irll_instr_invalid;

            Str const instr_name = hl_expr->of_instr.instr_name;
            if (strEq("extern", instr_name, 6))
                ret_expr.of_instr = irll_instr_extern;
            else if (strEq("?", instr_name, 1))
                ret_expr.of_instr = irll_instr_branch;
            else if (strEq("|", instr_name, 1))
                ret_expr.of_instr = irll_instr_branchcase;
            else if (strEq("==", instr_name, 2))
                ret_expr.of_instr = irll_instr_cmp_eq;
            else if (strEq("/=", instr_name, 2))
                ret_expr.of_instr = irll_instr_cmp_neq;
            else if (strEq("<=", instr_name, 2))
                ret_expr.of_instr = irll_instr_cmp_leq;
            else if (strEq(">=", instr_name, 2))
                ret_expr.of_instr = irll_instr_cmp_geq;
            else if (strEq("<", instr_name, 1))
                ret_expr.of_instr = irll_instr_cmp_lt;
            else if (strEq(">", instr_name, 1))
                ret_expr.of_instr = irll_instr_cmp_gt;
            else if (strEq("+", instr_name, 1))
                ret_expr.of_instr = irll_instr_arith_add;
            else if (strEq("-", instr_name, 1))
                ret_expr.of_instr = irll_instr_arith_sub;
            else if (strEq("*", instr_name, 1))
                ret_expr.of_instr = irll_instr_arith_mul;
            else if (strEq("/", instr_name, 1))
                ret_expr.of_instr = irll_instr_arith_div;
            else if (strEq("\x25", instr_name, 1))
                ret_expr.of_instr = irll_instr_arith_rem;

            if (ret_expr.of_instr == irll_instr_invalid)
                fail_msg = str2(str("TODO: irllExprFrom for .kind=instr of "), instr_name);
        } break;

        case irhl_expr_ref: {
            IrHlRef* ref = ·last(hl_expr->of_ref.path);
            switch (ref->kind) {
                case irhl_ref_def: {
                    ret_expr.kind = irll_expr_ref_func;
                    ret_expr.of_ref_func = irllFuncFrom(ctx, ref->of_def);
                } break;
                case irhl_ref_func_param: {
                    IrHlExpr* ref_func = NULL;
                    for (UInt i = hl_expr->of_ref.path.len - 2; i > 0; i -= 1) {
                        IrHlRef* this = &hl_expr->of_ref.path.at[i];
                        if (this->kind == irhl_ref_expr_func) {
                            ref_func = this->of_expr_func;
                            ·assert(ref_func->kind = irhl_expr_func);
                            break;
                        }
                    }
                    ·assert(ref_func != NULL);
                    ret_expr.kind = irll_expr_ref_param;
                    ret_expr.of_ref_param = ref_func->of_func.params.len;
                    ·forEach(IrHlFuncParam, param, ref_func->of_func.params, {
                        if (param->name.at == NULL) {
                            astPrintExpr(ref_func->anns.origin.ast_expr, false, 0);
                            ·fail(str("\n\n___NULL"));
                        }
                        Str const s1 = ref->of_func_param->name;
                        Str const s2 = param->name;
                        if (strEql(s1, s2)) {
                            ret_expr.of_ref_param = iˇparam;
                            break;
                        }
                    });
                    ·assert(ret_expr.of_ref_param < ref_func->of_func.params.len);
                } break;
                default: fail_msg = str2(str("TODO: handle ref kind of "), uIntToStr(ref->kind, 1, 10));
            }
        } break;

        case irhl_expr_selector: {
            ·assert(hl_expr->of_selector.member->kind = irhl_expr_tag);
            IrLlExpr member = irllExprFrom(ctx, hl_expr->of_selector.member);
            IrLlExpr subj = irllExprFrom(ctx, hl_expr->of_selector.subj);
            switch (subj.kind) {
                case irll_expr_call: // fall through to:
                case irll_expr_ref_func: {
                    ret_expr.kind = irll_expr_call;
                    ret_expr.of_call = ·new(IrLlExprCall);
                    ret_expr.of_call->is_closure = false;
                    ret_expr.of_call->callee = (IrLlExpr) {.kind = irll_expr_instr, .of_instr = irll_instr_bag_item_name};
                    ret_expr.of_call->args = ·sliceOf(IrLlExpr, 2, 2);
                    ret_expr.of_call->args.at[0] = subj;
                    ret_expr.of_call->args.at[1] = (IrLlExpr) {.kind = irll_expr_int, .of_int = member.of_tag};
                } break;
                default:
                    fail_msg = str3(str("TODO: handle selector subjects of .kind "), uIntToStr(subj.kind, 1, 10),
                                    hl_expr->of_selector.member->of_tag.tag_ident);
            }
        } break;

        case irhl_expr_call: {
            ret_expr.kind = irll_expr_call;
            ret_expr.of_call = ·new(IrLlExprCall);
            ret_expr.of_call->callee = irllExprFrom(ctx, hl_expr->of_call.callee);
            ret_expr.of_call->args = ·sliceOf(IrLlExpr, hl_expr->of_call.args.len, 0);
            ·forEach(IrHlExpr, arg, hl_expr->of_call.args, { ret_expr.of_call->args.at[iˇarg] = irllExprFrom(ctx, arg); });
        } break;

        default: break;
    }

    if (ret_expr.kind == irll_expr_invalid && fail_msg.at == NULL)
        fail_msg = str2(str("TODO: irllExprFrom for .kind of "), uIntToStr(hl_expr->kind, 1, 10));
    if (fail_msg.at != NULL) {
        printf("\n>>>>\n");
        irhlPrintExpr(hl_expr, false, 0);
        printf("\n<<<<\n");
        ·fail(fail_msg);
    }
    return ret_expr;
}

UInt irllFuncFrom(CtxIrLlFromHL* const ctx, IrHlDef* const hl_def) {
    ·forEach(IrLlFunc, func, ctx->prog.funcs, {
        if (func->anns.origin == hl_def)
            return iˇfunc;
    });

    UInt ret_idx = ctx->prog.funcs.len;
    ·append(ctx->prog.funcs, ((IrLlFunc) {.num_params = 0, .body = (IrLlExpr) {.kind = 0}, .anns = {.origin = hl_def}}));
    IrLlFunc* func = ·last(ctx->prog.funcs);

    if (hl_def->body->kind != irhl_expr_func)
        func->body = irllExprFrom(ctx, hl_def->body);
    else {
        func->num_params = hl_def->body->of_func.params.len;
        func->body = irllExprFrom(ctx, hl_def->body->of_func.body);
    }

    return ret_idx;
}

IrLlProg irllProgFrom(IrHlDef* hl_def, IrHlProg* const hl_prog) {
    CtxIrLlFromHL ctx = (CtxIrLlFromHL) {.tags = ·listOf(IrHlExprTag, 0, 32),
                                         .prog = (IrLlProg) {
                                             .funcs = ·listOf(IrLlFunc, 0, hl_prog->defs.len),
                                             .anns = {.origin = hl_prog},
                                         }};
    irllFuncFrom(&ctx, hl_def);
    return ctx.prog;
}




void irllPrintFunc(IrLlFunc const* const func) {
    printStr(func->anns.origin->name);
    printChr('\n');
}

void irllPrintProg(IrLlProg const* const prog) {
    ·forEach(IrLlFunc, func, prog->funcs, { irllPrintFunc(func); });
}
