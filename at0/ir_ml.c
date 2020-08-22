#pragma once
#include "utils_std_basics.c"
#include "utils_std_mem.c"
#include "ir_hl.c"

#define enum_invalid 4294967295

// milestone01 ir_ml:
// - funcs
// - calls
// - prims:
//   - @if
//   - val sym
//   - @iCmpEq
//   - @tmpRndSym()
//   - @never
// - cps: cont/jump

typedef enum IrMlNodeKind {
    irml_node_prim,
    irml_node_cont,
    irml_node_func,
    irml_node_call,
    irml_node_param,
} IrMlNodeKind;

typedef enum IrMlPrimKind {
    irml_prim_never,
    irml_prim_cond,
    irml_prim_val,
    irml_prim_cmpi,
    irml_prim_tmprndsym,
} IrMlPrimKind;

typedef enum IrMlValKind {
    irml_val_void,
    irml_val_sym,
} IrMlValKind;

typedef enum IrMlCmpIKind {
    irml_cmpi_eq,
} IrMlCmpIKind;



struct IrMlNode;
typedef struct IrMlNode IrMlNode;
typedef ·ListOfPtrs(IrMlNode) IrMlNodes;

typedef struct IrMlPrimCond {
    IrMlNode* pred;
    IrMlNode* if1;
    IrMlNode* if0;
} IrMlPrimCond;

typedef struct IrMlPrimVal {
    IrMlValKind kind;
    union {
        UInt sym;
    } of;
} IrMlPrimVal;

typedef struct IrMlPrimCmpI {
    IrMlCmpIKind kind;
    IrMlNode* lhs;
    IrMlNode* rhs;
} IrMlPrimCmpI;

typedef struct IrMlNodePrim {
    IrMlPrimKind kind;
    union {
        IrMlPrimCond cond;
        IrMlPrimVal val;
        IrMlPrimCmpI cmpi;
    } of;
} IrMlNodePrim;

typedef struct IrMlNodeFunc {
    UInt params_count;
    IrMlNode* body;
    struct {
        Str name;
    } anns;
} IrMlNodeFunc;

typedef struct IrMlNodeParam {
    IrMlNode* func_or_cont;
    UInt param_idx;
    struct {
        IrHlFuncParam* orig;
    } anns;
} IrMlNodeParam;

typedef struct IrMlNodeCall {
    IrMlNode* callee;
    IrMlNodes args;
} IrMlNodeCall;

typedef struct IrMlNodeCont {
    IrMlNodes params;
    struct {
        IrMlNode* to_cont;
        IrMlNodes args;
    } jump;
} IrMlNodeCont;

struct IrMlNode {
    IrMlNodeKind kind;
    union {
        IrMlNodeCont cont;
        IrMlNodeFunc func;
        IrMlNodeParam param;
        IrMlNodeCall call;
        IrMlNodePrim prim;
    } of;
    struct {
        struct {
            IrHlDef* def;
            IrHlExpr* expr;
        } orig;
    } anns;
};

typedef struct IrMlProg {
    IrMlNodes all_nodes;
    struct {
        IrHlProg* orig;
    } anns;
    struct {
        IrMlNode* bool_true;
        IrMlNode* bool_false;
    } cache;
} IrMlProg;




typedef struct IrMlCtxPrint {
    IrMlProg* prog;
    Bool fcimpl;
} IrMlCtxPrint;

Str irmlPrint(IrMlCtxPrint* const ctx, Str buf, IrMlNode* const node) {
    switch (node->kind) {
        case irml_node_func: {
            if (node->of.func.anns.name.len > 0 && !ctx->fcimpl)
                buf = strCopyTo(buf, node->of.func.anns.name);
            else {
                IrHlExprFunc* const fn_orig = (node->anns.orig.expr->kind != irhl_expr_func) ? NULL : &node->anns.orig.expr->of_func;
                ctx->fcimpl = false;
                ·push(buf, '(');
                for (UInt i = 0; i < node->of.func.params_count; i += 1) {
                    Str const name = (fn_orig == NULL) ? ·len0(U8) : fn_orig->params.at[i].name;
                    if (name.at != NULL && name.len > 0) {
                        buf = strCopyTo(buf, name);
                        ·push(buf, ' ');
                    } else {
                        buf = strCopyTo(buf, uIntToStr(NULL, i, 1, 10));
                        buf = strCopyTo(buf, str("a "));
                    }
                }
                buf = strCopyTo(buf, str("-> "));
                buf = irmlPrint(ctx, buf, node->of.func.body);
                ·push(buf, ')');
            }
        } break;
        case irml_node_param: {
            buf = strCopyTo(buf, node->of.param.anns.orig->name);
        } break;
        case irml_node_call: {
            ·push(buf, '(');
            buf = irmlPrint(ctx, buf, node->of.call.callee);
            for (UInt i = 0; i < node->of.call.args.len; i += 1) {
                ·push(buf, ' ');
                buf = irmlPrint(ctx, buf, node->of.call.args.at[i]);
            }
            ·push(buf, ')');
        } break;
        case irml_node_prim: {
            switch (node->of.prim.kind) {
                case irml_prim_val: {
                    switch (node->of.prim.of.val.kind) {
                        case irml_val_void: {
                            buf = strCopyTo(buf, str("()"));
                        } break;
                        case irml_val_sym: {
                            ·push(buf, '#');
                            buf = strCopyTo(buf, ctx->prog->anns.orig->anns.all_tags.at[node->of.prim.of.val.of.sym]);
                        } break;
                        default: ·fail(uIntToStr(NULL, node->of.prim.of.val.kind, 1, 10));
                    }
                } break;
                case irml_prim_cmpi: {
                    ·push(buf, '(');
                    buf = irmlPrint(ctx, buf, node->of.prim.of.cmpi.lhs);
                    switch (node->of.prim.of.cmpi.kind) {
                        case irml_cmpi_eq: buf = strCopyTo(buf, str(" == ")); break;
                        default: ·fail(uIntToStr(NULL, node->of.prim.of.cmpi.kind, 1, 10));
                    }
                    buf = irmlPrint(ctx, buf, node->of.prim.of.cmpi.rhs);
                    ·push(buf, ')');
                } break;
                case irml_prim_tmprndsym: {
                    buf = strCopyTo(buf, str("@tmpRndSym"));
                } break;
                case irml_prim_never: {
                    buf = strCopyTo(buf, str("@never"));
                } break;
                case irml_prim_cond: {
                    buf = strCopyTo(buf, str("(@if "));
                    buf = irmlPrint(ctx, buf, node->of.prim.of.cond.pred);
                    ·push(buf, ' ');
                    buf = irmlPrint(ctx, buf, node->of.prim.of.cond.if1);
                    ·push(buf, ' ');
                    buf = irmlPrint(ctx, buf, node->of.prim.of.cond.if0);
                    ·push(buf, ')');
                } break;
                default: ·fail(uIntToStr(NULL, node->of.prim.kind, 1, 10));
            }
            break;
        }
        default: ·fail(uIntToStr(NULL, node->kind, 1, 10));
    }
    return buf;
}




Bool irmlIsNonDet(IrMlNode* const node) {
    switch (node->kind) {
        case irml_node_param: return false;
        case irml_node_func: return irmlIsNonDet(node->of.func.body);
        case irml_node_cont: {
            Bool ret = irmlIsNonDet(node->of.cont.jump.to_cont);
            for (UInt i = 0; i < node->of.cont.jump.args.len && !ret; i += 1)
                ret = irmlIsNonDet(node->of.cont.jump.args.at[i]);
            return ret;
        }
        case irml_node_call: {
            Bool ret = irmlIsNonDet(node->of.call.callee);
            for (UInt i = 0; i < node->of.call.args.len && !ret; i += 1)
                ret = irmlIsNonDet(node->of.call.args.at[i]);
            return ret;
        }
        case irml_node_prim:
            switch (node->of.prim.kind) {
                case irml_prim_never: return false;
                case irml_prim_tmprndsym: return true;
                case irml_prim_cond:
                    return irmlIsNonDet(node->of.prim.of.cond.if0) || irmlIsNonDet(node->of.prim.of.cond.if1)
                           || irmlIsNonDet(node->of.prim.of.cond.pred);
                case irml_prim_cmpi: return irmlIsNonDet(node->of.prim.of.cmpi.lhs) || irmlIsNonDet(node->of.prim.of.cmpi.rhs);
                case irml_prim_val:
                    switch (node->of.prim.of.val.kind) {
                        default: ·fail(uIntToStr(NULL, node->of.prim.of.val.kind, 1, 10));
                    }
                default: ·fail(uIntToStr(NULL, node->of.prim.kind, 1, 10));
            }
        default: ·fail(uIntToStr(NULL, node->kind, 1, 10));
    }
    return false;
}

Bool irmlIsPrim(IrMlNode* const node, IrMlPrimKind const kind) {
    return node->kind == irml_node_prim && node->of.prim.kind == kind;
}

Bool irmlIsPrimVal(IrMlNode* const node, IrMlValKind const kind) {
    return node->kind == irml_node_prim && node->of.prim.kind == irml_prim_val && node->of.prim.of.val.kind == kind;
}

IrMlNode* irmlNodeByOrig(IrMlProg* const prog, Str const orig_name, IrHlDef* const orig_def, IrHlExpr* const orig_expr) {
    for (UInt i = 0; i < prog->all_nodes.len; i += 1) {
        IrMlNode* node = prog->all_nodes.at[i];
        if ((orig_def != NULL && orig_expr == NULL && node->anns.orig.def == orig_def)
            || (orig_expr != NULL && node->anns.orig.expr == orig_expr
                && (orig_def == NULL || node->anns.orig.def == NULL || node->anns.orig.def == orig_def))
            || (orig_name.at != NULL && orig_name.len > 0 && node->anns.orig.def != NULL && strEql(node->anns.orig.def->name, orig_name)))
            return node;
    }
    return NULL;
}

Bool irmlNodesEq(IrMlNode const* const n1, IrMlNode const* const n2) {
    if (n1 != n2 && n1 != NULL && n2 != NULL && n1->kind == n2->kind) {
        if (n1->anns.orig.expr != NULL && n1->anns.orig.expr == n2->anns.orig.expr)
            return true;
        switch (n1->kind) {

            case irml_node_cont: {
                if (n1->of.cont.params.len == n2->of.cont.params.len && n1->of.cont.jump.args.len == n2->of.cont.jump.args.len) {
                    if (irmlNodesEq(n1->of.cont.jump.to_cont, n2->of.cont.jump.to_cont)) {
                        if (n1->of.cont.params.at != n2->of.cont.params.at)
                            for (UInt i = 0; i < n1->of.cont.params.len; i += 1)
                                if (!irmlNodesEq(n1->of.cont.params.at[i], n2->of.cont.params.at[i]))
                                    return false;
                        if (n1->of.cont.jump.args.at != n2->of.cont.jump.args.at)
                            for (UInt i = 0; i < n1->of.cont.jump.args.len; i += 1)
                                if (!irmlNodesEq(n1->of.cont.jump.args.at[i], n2->of.cont.jump.args.at[i]))
                                    return false;
                        return true;
                    }
                }
            } break;

            case irml_node_func: {
                if (n1->of.func.params_count == n2->of.func.params_count)
                    return n1->of.func.body != NULL && n2->of.func.body != NULL && irmlNodesEq(n1->of.func.body, n2->of.func.body);
            } break;

            case irml_node_call: {
                if (n1->of.call.args.len == n2->of.call.args.len && irmlNodesEq(n1->of.call.callee, n2->of.call.callee)) {
                    if (n1->of.call.args.at != n2->of.call.args.at)
                        for (UInt i = 0; i < n1->of.call.args.len; i += 1)
                            if (!irmlNodesEq(n1->of.call.args.at[i], n2->of.call.args.at[i]))
                                return false;
                    return true;
                }
            } break;

            case irml_node_param: {
                return (n1->of.param.param_idx == n2->of.param.param_idx
                        && (n1->of.param.anns.orig == n2->of.param.anns.orig
                            || (n1->of.param.func_or_cont != NULL && n1->of.param.func_or_cont == n2->of.param.func_or_cont)));
            } break;

            case irml_node_prim: {
                if (n1->of.prim.kind == n2->of.prim.kind)
                    switch (n1->of.prim.kind) {
                        case irml_prim_never: return true;
                        case irml_prim_tmprndsym: return true;
                        case irml_prim_cmpi:
                            return (n1->of.prim.of.cmpi.kind == n2->of.prim.of.cmpi.kind)
                                   && irmlNodesEq(n1->of.prim.of.cmpi.lhs, n2->of.prim.of.cmpi.lhs)
                                   && irmlNodesEq(n1->of.prim.of.cmpi.rhs, n2->of.prim.of.cmpi.rhs);
                        case irml_prim_cond:
                            return irmlNodesEq(n1->of.prim.of.cond.pred, n2->of.prim.of.cond.pred)
                                   && irmlNodesEq(n1->of.prim.of.cond.if0, n2->of.prim.of.cond.if0)
                                   && irmlNodesEq(n1->of.prim.of.cond.if1, n2->of.prim.of.cond.if1);
                        case irml_prim_val: {
                            if (n1->of.prim.of.val.kind == n2->of.prim.of.val.kind)
                                switch (n1->of.prim.of.val.kind) {
                                    case irml_val_void: return true;
                                    case irml_val_sym: return (n1->of.prim.of.val.of.sym == n2->of.prim.of.val.of.sym);
                                    default: ·fail(uIntToStr(NULL, n1->of.prim.of.val.kind, 1, 10));
                                }
                        } break;
                        default: ·fail(uIntToStr(NULL, n1->of.prim.kind, 1, 10));
                    }
            } break;

            default: ·fail(uIntToStr(NULL, n1->kind, 1, 10));
        }
    }
    return n1 == n2;
}




IrMlNode* irmlNode(IrMlProg* const prog, IrMlNode const spec) {
    for (UInt i = 0; i < prog->all_nodes.len; i += 1) {
        IrMlNode* this_node = prog->all_nodes.at[i];
        if (irmlNodesEq(this_node, &spec))
            return this_node;
    }
    IrMlNode* keep = ·new(IrMlNode, NULL);
    *keep = spec;
    ·append(prog->all_nodes, keep);
    return keep;
}
IrMlNode* irmlNodePrim(IrMlProg* const prog, IrMlNodePrim const spec) {
    return irmlNode(prog, (IrMlNode) {.kind = irml_node_prim, .anns = {.orig = {.def = NULL, .expr = NULL}}, .of = {.prim = spec}});
}
IrMlNode* irmlNodePrimNever(IrMlProg* const prog) {
    return irmlNodePrim(prog, (IrMlNodePrim) {.kind = irml_prim_never});
}
IrMlNode* irmlNodePrimTmpRndSym(IrMlProg* const prog) {
    return irmlNodePrim(prog, (IrMlNodePrim) {.kind = irml_prim_tmprndsym});
}
IrMlNode* irmlNodePrimVal(IrMlProg* const prog, IrMlPrimVal const spec) {
    return irmlNodePrim(prog, (IrMlNodePrim) {.kind = irml_prim_val, .of = {.val = spec}});
}
IrMlNode* irmlNodePrimValVoid(IrMlProg* const prog) {
    return irmlNodePrimVal(prog, (IrMlPrimVal) {.kind = irml_val_void});
}
IrMlNode* irmlNodePrimValSym(IrMlProg* const prog, UInt const spec) {
    return irmlNodePrimVal(prog, (IrMlPrimVal) {.kind = irml_val_sym, .of = {.sym = spec}});
}
IrMlNode* irmlNodePrimValSymBool(IrMlProg* const prog, Bool const spec) {
    return irmlNodePrimValSym(prog, spec);
}
IrMlNode* irmlNodePrimCond(IrMlProg* const prog, IrMlPrimCond spec) {
    if (spec.pred == prog->cache.bool_false)
        return spec.if0;
    if (spec.pred == prog->cache.bool_true)
        return spec.if1;
    if (irmlIsPrim(spec.pred, irml_prim_cmpi)) {
        Bool const is_lhs_false = (spec.pred->of.prim.of.cmpi.lhs == prog->cache.bool_false);
        Bool const is_rhs_false = (spec.pred->of.prim.of.cmpi.rhs == prog->cache.bool_false);
        if (is_lhs_false || is_rhs_false) {
            if (is_lhs_false)
                spec.pred = spec.pred->of.prim.of.cmpi.rhs;
            else if (is_rhs_false)
                spec.pred = spec.pred->of.prim.of.cmpi.lhs;

            IrMlNode* if0 = spec.if0;
            spec.if0 = spec.if1;
            spec.if1 = if0;
        }
    }
    return irmlNodePrim(prog, (IrMlNodePrim) {.kind = irml_prim_cond, .of = {.cond = spec}});
}
IrMlNode* irmlNodePrimCmpI(IrMlProg* const prog, IrMlPrimCmpI const spec) {
    if (irmlIsPrimVal(spec.lhs, irml_val_sym) && irmlIsPrimVal(spec.rhs, irml_val_sym))
        return irmlNodePrimValSymBool(prog, spec.lhs->of.prim.of.val.of.sym == spec.rhs->of.prim.of.val.of.sym);
    if (irmlNodesEq(spec.lhs, prog->cache.bool_true))
        return spec.rhs;
    if (irmlNodesEq(spec.rhs, prog->cache.bool_true))
        return spec.lhs;
    return irmlNodePrim(prog, (IrMlNodePrim) {.kind = irml_prim_cmpi, .of = {.cmpi = spec}});
}
IrMlNode* irmlNodeCall(IrMlProg* const prog, IrMlNodeCall const spec) {
    return irmlNode(prog, (IrMlNode) {.kind = irml_node_call, .anns = {.orig = {.def = NULL, .expr = NULL}}, .of = {.call = spec}});
}
IrMlNode* irmlNodeCont(IrMlProg* const prog, IrMlNodeCont const spec) {
    return irmlNode(prog, (IrMlNode) {.kind = irml_node_cont, .anns = {.orig = {.def = NULL, .expr = NULL}}, .of = {.cont = spec}});
}
IrMlNode* irmlNodeFunc(IrMlProg* const prog, IrMlNodeFunc const spec) {
    return irmlNode(prog, (IrMlNode) {.kind = irml_node_func, .anns = {.orig = {.def = NULL, .expr = NULL}}, .of = {.func = spec}});
}
IrMlNode* irmlNodeParam(IrMlProg* const prog, IrMlNodeParam const spec) {
    return irmlNode(prog, (IrMlNode) {.kind = irml_node_param, .anns = {.orig = {.def = NULL, .expr = NULL}}, .of = {.param = spec}});
}




IrMlNode* irmlNodeFrom(IrMlProg* const prog, IrHlDef* const def, IrHlExpr* const expr);

IrMlNode* irmlNodePrimCmpIFrom(IrMlProg* const prog, IrHlDef* const def, IrHlExpr* const expr) {
    ·assert(expr->kind = irhl_expr_call);
    ·assert(expr->of_call.callee->kind = irhl_expr_instr);
    ·assert(expr->of_call.args.len == 2);

    IrMlCmpIKind kind = enum_invalid;
    if (strEq("==", expr->of_call.callee->of_instr.instr_name, 2))
        kind = irml_cmpi_eq;
    ·assert(kind != enum_invalid);

    return irmlNodePrimCmpI(prog, (IrMlPrimCmpI) {
                                      .kind = kind,
                                      .lhs = irmlNodeFrom(prog, def, &expr->of_call.args.at[0]),
                                      .rhs = irmlNodeFrom(prog, def, &expr->of_call.args.at[1]),
                                  });
}

IrMlNode* irmlNodePrimCondFrom(IrMlProg* const prog, IrHlDef* const def, IrHlExpr* const expr) {
    IrMlNode* ret_node = NULL;

    ·assert(expr->kind = irhl_expr_call);
    ·assert(expr->of_call.callee->kind = irhl_expr_instr);
    ·assert(strEq("?", expr->of_call.callee->of_instr.instr_name, 1));
    ·assert(expr->of_call.args.len == 2);
    ·assert(expr->of_call.args.at[1].kind == irhl_expr_bag);
    ·assert(expr->of_call.args.at[1].of_bag.kind == irhl_bag_list);
    IrHlExprs const cases_bag = expr->of_call.args.at[1].of_bag.items;
    ºUInt default_case = ·none(UInt);
    { // if any default case, reposition it at end of cases (modifies orig irhl list but seems ok to do)
        for (UInt i = 0; i < cases_bag.len && !default_case.got; i += 1) {
            ·assert(cases_bag.at[i].kind == irhl_expr_call);
            ·assert(cases_bag.at[i].of_call.callee->kind == irhl_expr_instr);
            ·assert(strEq("|", cases_bag.at[i].of_call.callee->of_instr.instr_name, 1));
            ·assert(cases_bag.at[i].of_call.args.len == 2);
            if (cases_bag.at[i].of_call.args.at[0].kind == irhl_expr_nilish
                && cases_bag.at[i].of_call.args.at[0].of_nilish.kind == irhl_nilish_blank)
                default_case = ·got(UInt, i);
        }
        if (default_case.got && default_case.it != cases_bag.len - 1) {
            IrHlExpr const def_case = cases_bag.at[default_case.it];
            cases_bag.at[default_case.it] = cases_bag.at[cases_bag.len - 1];
            cases_bag.at[cases_bag.len - 1] = def_case;
            default_case.it = cases_bag.len - 1;
        }
    }
    IrMlNode* const scrut = irmlNodeFrom(prog, def, &expr->of_call.args.at[0]);
    IrMlNodes preds = ·listOfPtrs(IrMlNode, NULL, 0, 1 + cases_bag.len);
    IrMlNodes thens = ·listOfPtrs(IrMlNode, NULL, 0, 1 + cases_bag.len);
    for (UInt i = 0; i < cases_bag.len; i += 1) {
        ·append(preds, (default_case.got && i == default_case.it)
                           ? NULL
                           : irmlNodePrimCmpI(prog, (IrMlPrimCmpI) {.kind = irml_cmpi_eq,
                                                                    .lhs = scrut,
                                                                    .rhs = irmlNodeFrom(prog, def, &cases_bag.at[i].of_call.args.at[0])}));
        ·append(thens, irmlNodeFrom(prog, def, &cases_bag.at[i].of_call.args.at[1]));
    }
    if (!default_case.got) {
        ·append(preds, NULL);
        ·append(thens, irmlNodePrimNever(prog));
    }
    ·assert(preds.len == thens.len);
    ·assert(preds.len >= 2);
    //      @?      day          [@| #sun #true,        @| #sat #true,      @| _ #false]
    ret_node = irmlNodePrimCond(prog, (IrMlPrimCond) {
                                          .pred = preds.at[preds.len - 2],
                                          .if1 = thens.at[thens.len - 2],
                                          .if0 = thens.at[thens.len - 1],
                                      });
    for (UInt i = preds.len - 2; i > 0;) {
        i -= 1;
        ret_node = irmlNodePrimCond(prog, (IrMlPrimCond) {.pred = preds.at[i], .if1 = thens.at[i], .if0 = ret_node});
    }
    return ret_node;
}

IrMlNode* irmlNodeFrom(IrMlProg* const prog, IrHlDef* const def, IrHlExpr* const expr) {
    IrMlNode* ret_node = NULL;
    switch (expr->kind) {

        case irhl_expr_instr: {
            ·fail(expr->of_instr.instr_name);
        } break;

        case irhl_expr_nilish: {
            ·fail(uIntToStr(NULL, expr->of_nilish.kind, 1, 10));
        } break;

        case irhl_expr_tag: {
            ºUInt idx = ·none(UInt);
            for (UInt i = 0; i < prog->anns.orig->anns.all_tags.len && !idx.got; i += 1)
                if (strEql(expr->of_tag.tag_ident, prog->anns.orig->anns.all_tags.at[i]))
                    idx = ·got(UInt, i);
            ret_node = irmlNodePrimValSym(prog, idx.it);
        } break;

        case irhl_expr_call: {
            if (expr->of_call.callee->kind == irhl_expr_instr) {
                Str const instr_name = expr->of_call.callee->of_instr.instr_name;
                if (strEq("?", instr_name, 1))
                    ret_node = irmlNodePrimCondFrom(prog, def, expr);
                else if (strEq("tmpRndSym", instr_name, 9))
                    ret_node = irmlNodePrimTmpRndSym(prog);
                else if (strEq("==", instr_name, 2))
                    ret_node = irmlNodePrimCmpIFrom(prog, def, expr);
                else
                    ·fail(instr_name);
            }

            if (ret_node == NULL) {
                IrMlNodeCall const spec = (IrMlNodeCall) {.callee = irmlNodeFrom(prog, def, expr->of_call.callee),
                                                          .args = ·listOfPtrs(IrMlNode, NULL, expr->of_call.args.len, 0)};
                for (UInt i = 0; i < spec.args.len; i += 1)
                    spec.args.at[i] = irmlNodeFrom(prog, def, &expr->of_call.args.at[i]);
                ret_node = irmlNodeCall(prog, spec);
            }
        } break;

        case irhl_expr_func: {
            ret_node = irmlNodeByOrig(prog, ·len0(U8), def, expr);
            if (ret_node == NULL) {
                IrMlNodeFunc const spec =
                    (IrMlNodeFunc) {.body = NULL, .params_count = expr->of_func.params.len, .anns = {.name = expr->of_func.anns.qname}};
                ret_node = irmlNodeFunc(prog, spec);
                ret_node->anns.orig.def = def;
                ret_node->anns.orig.expr = expr;
                ret_node->of.func.body = irmlNodeFrom(prog, def, expr->of_func.body);
            }
        } break;

        case irhl_expr_ref: {
            ·assert(expr->of_ref.path.len > 0);
            IrHlRef* const ref = ·last(expr->of_ref.path);
            switch (ref->kind) {
                case irhl_ref_func_param: {
                    ·assert(expr->of_ref.path.len > 1);
                    IrHlRef* const fn = &expr->of_ref.path.at[expr->of_ref.path.len - 2];
                    ·assert(fn->kind == irhl_ref_expr_func);
                    Bool const is_param_of_cur_def = (fn->of_expr_func == def->body);
                    IrHlRef* maybe_def = NULL;
                    if (!is_param_of_cur_def)
                        for (UInt i = expr->of_ref.path.len - 2; i > 0 && maybe_def == NULL;) {
                            i -= 1;
                            if (expr->of_ref.path.at[i].kind == irhl_ref_def)
                                maybe_def = &expr->of_ref.path.at[i];
                        }
                    ºUInt idx = ·none(UInt);
                    for (UInt i = 0; i < fn->of_expr_func->of_func.params.len && !idx.got; i += 1)
                        if (strEql(ref->of_func_param->name, fn->of_expr_func->of_func.params.at[i].name))
                            idx = ·got(UInt, i);
                    ·assert(idx.got);
                    ret_node = irmlNodeParam(
                        prog, (IrMlNodeParam) {.func_or_cont =
                                                   is_param_of_cur_def
                                                       ? NULL
                                                       : irmlNodeFrom(prog, (maybe_def == NULL) ? def : maybe_def->of_def, fn->of_expr_func),
                                               .param_idx = idx.it,
                                               .anns = {.orig = ref->of_func_param}});
                } break;
                case irhl_ref_let: {
                    ret_node = irmlNodeFrom(prog, def, ref->of_let->expr);
                } break;
                default: ·fail(uIntToStr(NULL, ref->kind, 1, 10));
            }
        } break;

        case irhl_expr_selector: {
            ·assert(expr->of_selector.subj->kind == irhl_expr_ref);
            ·assert(expr->of_selector.member->kind == irhl_expr_tag);
            IrHlDef* ref_to_def = NULL;
            IrHlExpr* ref_to_expr = NULL;
            irhlRefTarget(prog->anns.orig, def, expr, &ref_to_def, &ref_to_expr, NULL, NULL);
            if (ref_to_expr != NULL)
                ret_node = irmlNodeFrom(prog, (ref_to_def == NULL) ? def : ref_to_def, ref_to_expr);
            else if (ref_to_def != NULL)
                ret_node = irmlNodeFrom(prog, ref_to_def, ref_to_def->body);
            else
                ·fail(expr->of_selector.member->of_tag.tag_ident);
        } break;

        case irhl_expr_let: {
            ret_node = irmlNodeFrom(prog, def, expr->of_let.body);
        } break;

        default: ·fail(uIntToStr(NULL, expr->kind, 1, 10));
    }
    ·assert(ret_node != NULL);
    if (ret_node->anns.orig.def == NULL)
        ret_node->anns.orig.def = def;
    if (ret_node->anns.orig.expr == NULL)
        ret_node->anns.orig.expr = expr;
    if (ret_node->kind == irml_node_func)
        for (UInt i = 0; i < prog->all_nodes.len; i += 1) {
            IrMlNode* const node = prog->all_nodes.at[i];
            if (node->kind == irml_node_param && node->of.param.func_or_cont == NULL) {
                IrHlExpr* const fn = node->anns.orig.expr->of_ref.path.at[node->anns.orig.expr->of_ref.path.len - 2].of_expr_func;
                for (UInt j = 0; j < prog->all_nodes.len && (node->of.param.func_or_cont == NULL); j += 1)
                    if (prog->all_nodes.at[j]->anns.orig.expr == fn)
                        node->of.param.func_or_cont = prog->all_nodes.at[j];
            }
        }
    return ret_node;
}

IrMlProg irmlProgFrom(IrHlProg* const hl_prog, Str const path_based_ident_prefix) {
    ·assert(hl_prog->anns.all_tags.len >= 2 && strEql(hl_prog->anns.all_tags.at[0], str("false"))
            && strEql(hl_prog->anns.all_tags.at[1], str("true")));
    IrMlProg ret_prog = (IrMlProg) {
        .anns = {.orig = hl_prog},
        .all_nodes = ·listOfPtrs(IrMlNode, NULL, 0, 256),
    };
    ret_prog.cache.bool_true = irmlNodePrimValSymBool(&ret_prog, true);
    ret_prog.cache.bool_false = irmlNodePrimValSymBool(&ret_prog, false);
    ·ListOfPtrs(IrHlDef) hl_defs = ·listOfPtrs(IrHlDef, NULL, 0, 16);

    { // temp milestone01 logic: ir_hl is already much farther ahead of what we want for now from ir_ml
        IrHlDef* namespace = NULL;
        ·forEach(IrHlDef, top_def, hl_prog->defs, {
            if (strEql(path_based_ident_prefix, top_def->name)) {
                namespace = top_def;
                break;
            }
        });
        ·assert(namespace != NULL);
        ·assert(namespace->body->kind == irhl_expr_bag);
        ·assert(namespace->body->of_bag.kind == irhl_bag_struct);
        ·forEach(IrHlExpr, struct_item, namespace->body->of_bag.items, {
            ·assert(struct_item->kind == irhl_expr_kvpair);
            ·assert(struct_item->of_kvpair.key->kind == irhl_expr_field_name);
            ·assert(struct_item->of_kvpair.val->kind = irhl_expr_ref);
            ·assert(·last(struct_item->of_kvpair.val->of_ref.path)->kind == irhl_ref_def);
            ·append(hl_defs, ·last(struct_item->of_kvpair.val->of_ref.path)->of_def);
        });
    }

    for (UInt i = 0; i < hl_defs.len; i += 1) {
        IrHlDef* def = hl_defs.at[i];
        IrMlNode* node = irmlNodeFrom(&ret_prog, def, def->body);
        if (node->anns.orig.def == NULL)
            node->anns.orig.def = def;
    }

    {
        IrMlCtxPrint ctx = (IrMlCtxPrint) {.prog = &ret_prog, .fcimpl = true};
        U8 bbuf[1024];
        for (UInt i = 0; i < hl_defs.len; i += 1) {
            IrHlDef* def = hl_defs.at[i];
            IrMlNode* node = irmlNodeByOrig(&ret_prog, ·len0(U8), def, NULL);
            ·assert(node != NULL);
            ctx.fcimpl = true;
            Str buf = irmlPrint(&ctx, (Str) {.at = &bbuf[0], .len = 0}, node);
            fprintf(stderr, "\n%s:\n   ", strZ(def->name));
            fwrite(buf.at, 1, buf.len, stderr);
            fprintf(stderr, "\n");
            fflush(stderr);
        }
    }

    fprintf(stderr, "NN: %zu\n", ret_prog.all_nodes.len);
    return ret_prog;
}
