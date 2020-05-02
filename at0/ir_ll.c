#pragma once
#include "utils_and_libc_deps.c"
#include "ir_hl.c"

typedef enum IrLLInstrKind {
    irll_instr_invalid,
    irll_instr_extern,
    irll_instr_bag_item,
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
    irll_expr_invalid = 0,
    irll_expr_int = 1,
    irll_expr_tag = 2,
    irll_expr_bag = 3,
    irll_expr_ref_param = 4,
    irll_expr_ref_func = 5,
    irll_expr_instr = 6,
    irll_expr_call = 7,
} IrLLExprKind;

struct IrLLExprCall;
typedef struct IrLLExprCall IrLLExprCall;
struct IrLLExprBag;
typedef struct IrLLExprBag IrLLExprBag;

typedef struct IrLLExpr {
    IrLLExprKind kind;
    union {
        I64 of_int;
        UInt of_tag;
        IrLLExprBag* of_bag;
        UInt of_ref_param;
        UInt of_ref_func;
        IrLLInstrKind of_instr;
        IrLLExprCall* of_call;
    };
} IrLLExpr;
typedef ·SliceOf(IrLLExpr) IrLLExprs;

struct IrLLExprBag {
    IrLLExprs items;
    Strs field_names;
};

struct IrLLExprCall {
    IrLLExpr callee;
    IrLLExprs args;
    UInt is_closure;
};

typedef struct IrLLFunc {
    UInt num_params;
    IrLLExpr body;
    struct {
        IrHLDef* origin;
    } anns;
} IrLLFunc;
typedef ·ListOf(IrLLFunc) IrLLFuncs;

typedef struct IrLLProg {
    IrLLFuncs funcs;
    struct {
        IrHLProg* origin;
    } anns;
} IrLLProg;




typedef struct CtxIrLLFromHL {
    IrLLProg prog;
    IrHLExprTags tags;
} CtxIrLLFromHL;

UInt irLLFuncFrom(CtxIrLLFromHL* const ctx, IrHLDef* const hl_def);

IrLLExpr irLLExprFrom(CtxIrLLFromHL* const ctx, IrHLExpr* const hl_expr) {
    IrLLExpr ret_expr = (IrLLExpr) {.kind = irll_expr_invalid};
    Str fail_msg = ·len0(U8);
    switch (hl_expr->kind) {

        case irhl_expr_int: {
            ret_expr.kind = irll_expr_int;
            ret_expr.of_int = hl_expr->of_int.int_value;
        } break;

        case irhl_expr_let: {
            // named lets inside get pulled in on demand
            ret_expr = irLLExprFrom(ctx, hl_expr->of_let.body);
        } break;

        case irhl_expr_tag: {
            UInt tag_idx = ctx->tags.len;
            ·forEach(IrHLExprTag, tag, ctx->tags, {
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
            ret_expr.of_bag = ·new(IrLLExprBag);
            Bool const is_struct = (hl_expr->of_bag.kind == irhl_bag_struct);
            ret_expr.of_bag->field_names = (!is_struct) ? ·len0(Str) : ·sliceOf(Str, hl_expr->of_bag.items.len, 0);
            ret_expr.of_bag->items = ·sliceOf(IrLLExpr, hl_expr->of_bag.items.len, 0);
            ·forEach(IrHLExpr, item, hl_expr->of_bag.items, {
                if (!is_struct)
                    ret_expr.of_bag->items.at[iˇitem] = irLLExprFrom(ctx, item);
                else {
                    ·assert(item->kind == irhl_expr_kvpair);
                    ·assert(item->of_kvpair.key->kind == irhl_expr_field_name);
                    ret_expr.of_bag->field_names.at[iˇitem] = item->of_kvpair.key->of_field_name.field_name;
                    ret_expr.of_bag->items.at[iˇitem] = irLLExprFrom(ctx, item->of_kvpair.val);
                }
            });
        } break;

        case irhl_expr_instr: {
            ret_expr.kind = irll_expr_instr;
            ret_expr.of_instr = irll_instr_invalid;

            Str const instr_name = hl_expr->of_instr.instr_name;
            if (strEq("extern", instr_name, 6))
                ret_expr.of_instr = irll_instr_extern;

            if (ret_expr.of_instr == irll_instr_invalid)
                fail_msg = str2(str("TODO: irLLExprFrom for .kind=instr of "), instr_name);
        } break;

        case irhl_expr_ref: {
            IrHLRef* ref = ·last(hl_expr->of_ref.path);
            switch (ref->kind) {
                case irhl_ref_def: {
                    ret_expr.kind = irll_expr_ref_func;
                    ret_expr.of_ref_func = irLLFuncFrom(ctx, ref->of_def);
                } break;
                default: fail_msg = str2(str("TODO: handle ref kind of "), uIntToStr(ref->kind, 1, 10));
            }
        } break;

        case irhl_expr_selector: {
            ·assert(hl_expr->of_selector.member->kind == irhl_expr_tag);
            IrLLExpr subj = irLLExprFrom(ctx, hl_expr->of_selector.subj);
            switch (subj.kind) {
                case irll_expr_ref_func: {
                    ºUInt idx_field = ·none(UInt);
                    printf("ref-func %zu (%s) body-kind %d (%s)\n", subj.of_ref_func,
                           strZ(ctx->prog.funcs.at[subj.of_ref_func].anns.origin->name), ctx->prog.funcs.at[subj.of_ref_func].body.kind,
                           strZ(hl_expr->of_selector.member->of_tag.tag_ident));
                    if (ctx->prog.funcs.at[subj.of_ref_func].body.kind == irll_expr_bag)
                        ·forEach(Str, fld, ctx->prog.funcs.at[subj.of_ref_func].body.of_bag->field_names, {
                            if (strEql(*fld, hl_expr->of_selector.member->of_tag.tag_ident)) {
                                idx_field = ·ok(UInt, iˇfld);
                                break;
                            }
                        });
                    if (idx_field.ok) {
                        ret_expr.kind = irll_expr_call;
                        ret_expr.of_call = ·new(IrLLExprCall);
                        ret_expr.of_call->is_closure = false;
                        ret_expr.of_call->callee = (IrLLExpr) {.kind = irll_expr_instr, .of_instr = irll_instr_bag_item};
                        ret_expr.of_call->args = ·sliceOf(IrLLExpr, 2, 2);
                        ret_expr.of_call->args.at[0] = subj;
                        ret_expr.of_call->args.at[1] = (IrLLExpr) {.kind = irll_expr_int, .of_int = idx_field.it};
                    }
                } break;
                default: fail_msg = str2(str("TODO: handle selector subjects of .kind "), uIntToStr(subj.kind, 1, 10));
            }
        } break;

        case irhl_expr_call: {
            ret_expr.kind = irll_expr_call;
            ret_expr.of_call = ·new(IrLLExprCall);
            ret_expr.of_call->callee = irLLExprFrom(ctx, hl_expr->of_call.callee);
            ret_expr.of_call->args = ·sliceOf(IrLLExpr, hl_expr->of_call.args.len, 0);
            ·forEach(IrHLExpr, arg, hl_expr->of_call.args, { ret_expr.of_call->args.at[iˇarg] = irLLExprFrom(ctx, arg); });
        } break;

        default: break;
    }

    if (ret_expr.kind == irll_expr_invalid && fail_msg.at == NULL)
        fail_msg = str2(str("TODO: irLLExprFrom for .kind of "), uIntToStr(hl_expr->kind, 1, 10));
    if (fail_msg.at != NULL) {
        printf("\n>>>>\n");
        irHLPrintExpr(hl_expr, false, 0);
        printf("\n<<<<\n");
        ·fail(fail_msg);
    }
    return ret_expr;
}

UInt irLLFuncFrom(CtxIrLLFromHL* const ctx, IrHLDef* const hl_def) {
    ·forEach(IrLLFunc, func, ctx->prog.funcs, {
        if (func->anns.origin == hl_def)
            return iˇfunc;
    });

    UInt ret_idx = ctx->prog.funcs.len;
    printf("PUT %s in %zu\n", strZ(hl_def->name), ret_idx);
    ·append(ctx->prog.funcs, ((IrLLFunc) {.num_params = 0, .anns = {.origin = hl_def}}));
    IrLLFunc* func = ·last(ctx->prog.funcs);

    if (hl_def->body->kind != irhl_expr_func)
        func->body = irLLExprFrom(ctx, hl_def->body);
    else {
        func->num_params = hl_def->body->of_func.params.len;
        func->body = irLLExprFrom(ctx, hl_def->body->of_func.body);
    }

    return ret_idx;
}

IrLLProg irLLProgFrom(IrHLDef* hl_def, IrHLProg* const hl_prog) {
    CtxIrLLFromHL ctx = (CtxIrLLFromHL) {.tags = ·listOf(IrHLExprTag, 0, 16),
                                         .prog = (IrLLProg) {
                                             .funcs = ·listOf(IrLLFunc, 0, hl_prog->defs.len),
                                             .anns = {.origin = hl_prog},
                                         }};
    irLLFuncFrom(&ctx, hl_def);
    return ctx.prog;
}




void irLLPrintFunc(IrLLFunc const* const func) {
    printStr(func->anns.origin->name);
    printChr('\n');
}

void irLLPrintProg(IrLLProg const* const prog) {
    ·forEach(IrLLFunc, func, prog->funcs, { irLLPrintFunc(func); });
}
