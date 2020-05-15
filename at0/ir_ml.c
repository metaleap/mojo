#pragma once
#include "utils_and_libc_deps.c"


typedef enum IrMlKindOfType {
    irml_type_type,
    irml_type_void,
    irml_type_int,
    irml_type_fn,
    irml_type_tup,
    irml_type_arr,
    irml_type_ptr,
} IrMlKindOfType;

typedef enum IrMlKindOfNode {
    irml_node_fn,
    irml_node_param,
    irml_node_jump,
    irml_node_prim,
} IrMlKindOfNode;

typedef enum IrMlKindOfPrim {
    irml_prim_cond,
    irml_prim_cmpi,
    irml_prim_bini,
    irml_prim_cast,
    irml_prim_item,
    irml_prim_extcall,
    irml_prim_val,
} IrMlKindOfPrim;

typedef enum IrMlKindOfCmpI {
    irml_cmpi_eq,
    irml_cmpi_neq,
    irml_cmpi_leq,
    irml_cmpi_geq,
    irml_cmpi_lt,
    irml_cmpi_gt,
} IrMlKindOfCmpI;

typedef enum IrMlKindOfBinI {
    irml_bini_add,
    irml_bini_sub,
    irml_bini_mul,
    irml_bini_div,
    irml_bini_rem,
} IrMlKindOfBinI;

typedef enum IrMlKindOfCast {
    irml_cast_ints,
    irml_cast_bits,
} IrMlKindOfCast;


struct IrMlNode;
typedef struct IrMlNode IrMlNode;
typedef ·SliceOfPtrs(IrMlNode) IrMlPtrsOfNode;
typedef ·SliceOf(IrMlNode) IrMlNodes;


struct IrMlType;
typedef struct IrMlType IrMlType;

typedef struct IrMlTypeInt {
    U16 bit_width;
    Bool unsign;
} IrMlTypeInt;

typedef struct IrMlTypeVoid {
} IrMlTypeVoid;

typedef struct IrMlTypeTup {
    IrMlPtrsOfNode types;
} IrMlTypeTup;

typedef struct IrMlTypePtr {
    IrMlNode* type;
} IrMlTypePtr;

typedef struct IrMlTypeArr {
    IrMlNode* type;
    IrMlNode* length;
} IrMlTypeArr;

struct IrMlType {
    union {
        IrMlTypeInt num_int;
        IrMlTypeTup tup;
        IrMlTypePtr ptr;
        IrMlTypeArr arr;
    } of;
    IrMlKindOfType kind;
};


typedef struct IrMlNodeParam {
    IrMlNode* fn_node;
    UInt param_idx;
    struct {
        IrMlNode* cur_evald;
    } anns;
} IrMlNodeParam;

typedef struct IrMlNodeFn {
    IrMlNodes params;
    IrMlNode* jump;
} IrMlNodeFn;

typedef struct IrMlNodeJump {
    IrMlNode* target;
    IrMlPtrsOfNode args;
} IrMlNodeJump;

typedef struct IrMlPrimCond {
    IrMlNode* scrutinee;
    IrMlNode* default_result;
    IrMlPtrsOfNode comparee_ints;
    IrMlPtrsOfNode match_results;
} IrMlPrimCond;

typedef struct IrMlPrimVal {
    union {
        I64 int_val;
        IrMlPtrsOfNode list_val;
        IrMlType type;
    } of;
    IrMlKindOfType kind;
} IrMlPrimVal;

typedef struct IrMlPrimCmpI {
    IrMlNode* lhs;
    IrMlNode* rhs;
    IrMlKindOfCmpI kind;
} IrMlPrimCmpI;

typedef struct IrMlPrimBinI {
    IrMlNode* lhs;
    IrMlNode* rhs;
    IrMlKindOfBinI kind;
} IrMlPrimBinI;

typedef struct IrMlPrimCast {
    IrMlNode* subj;
    IrMlNode* dst_type;
    IrMlKindOfCast kind;
} IrMlPrimCast;

typedef struct IrMlPrimItem {
    IrMlNode* subj;
    IrMlNode* index;
    IrMlNode* set_to; // if NULL, it's a getter, else a setter
} IrMlPrimItem;

typedef struct IrMlPrimExtCall {
    IrMlNode* args_list_val;
    Str name;
} IrMlPrimExtCall;

typedef struct IrMlNodePrim {
    union {
        IrMlPrimCond cond;
        IrMlPrimVal val;
        IrMlPrimCmpI cmpi;
        IrMlPrimBinI bini;
        IrMlPrimCast cast;
        IrMlPrimItem item;
        IrMlPrimExtCall ext_call;
    } of;
    IrMlKindOfPrim kind;
} IrMlNodePrim;

struct IrMlNode {
    union {
        IrMlNodeFn fn;
        IrMlNodeParam param;
        IrMlNodeJump jump;
        IrMlNodePrim prim;
    } of;
    struct {
        IrMlNode* preduced;
        IrMlNode* type;
        Bool side_effects;
        Str name;
    } anns;
    IrMlKindOfNode kind;
};


typedef struct IrMlProg {
    struct {
        ·ListOfPtrs(IrMlNode) prims;
        ·ListOfPtrs(IrMlNode) jumps;
    } all;
    struct {
        U16 ptrs;
    } bit_widths;
} IrMlProg;




IrMlType* irmlNodeType(IrMlNode const* const node, Bool const must) {
    ·assert(node != NULL);
    if (must && node->anns.type == NULL)
        ·fail(str("encountered an untyped node post-preduce"));
    return (node->anns.type == NULL) ? NULL : &node->anns.type->of.prim.of.val.of.type;
}

// a ° b  ==  b ° a
Bool irmlPrimIsCommutative(IrMlKindOfPrim const prim_kind, int const op_kind) {
    return (prim_kind == irml_prim_bini && (op_kind == irml_bini_add || op_kind == irml_bini_mul))
           || (prim_kind == irml_prim_cmpi && (op_kind == irml_cmpi_eq || op_kind == irml_cmpi_neq));
}

// (a ° b) ° c  ==  a ° (b ° c)
Bool irmlPrimIsAssociative(IrMlKindOfPrim const prim_kind, int const op_kind) {
    return prim_kind == irml_prim_bini && (op_kind == irml_bini_add || op_kind == irml_bini_mul);
}

// a °¹ (b °² c)  ==  (a °¹ b)  °²  (a °¹ c)
Bool irmlPrimIsDistributive(IrMlKindOfPrim const prim_kind, int const op_kind1, int const op_kind2) {
    return prim_kind == irml_prim_bini && op_kind1 == irml_bini_mul && op_kind2 == irml_bini_add;
}

Bool irmlNodeIsPrimVal(IrMlNode const* const node, IrMlKindOfType const kind) {
    ·assert(node != NULL);
    return node->kind == irml_node_prim && node->of.prim.kind == irml_prim_val && node->of.prim.of.val.kind == kind;
}

Bool irmlNodeIsBasicBlockishFn(IrMlNode* const node, Bool const can_have_first_order_params) {
    IrMlType* ty = irmlNodeType(node, true);
    if (ty->kind == irml_type_fn) {
        if (!can_have_first_order_params)
            return (ty->of.tup.types.len == 0);
        for (UInt i = 0; i < ty->of.tup.types.len; i += 1)
            if (ty->of.tup.types.at[i]->of.prim.of.val.of.type.kind == irml_node_fn)
                return false;
        return true;
    }
    return false;
}

IrMlPtrsOfNode irmlNodes0() {
    IrMlPtrsOfNode ret_nodes = ·sliceOfPtrs(IrMlNode, 0, 0);
    return ret_nodes;
}
IrMlPtrsOfNode irmlNodes1(IrMlNode* const n0) {
    IrMlPtrsOfNode ret_nodes = ·sliceOfPtrs(IrMlNode, 1, 1);
    ret_nodes.at[0] = n0;
    return ret_nodes;
}
IrMlPtrsOfNode irmlNodes2(IrMlNode* const n0, IrMlNode* const n1) {
    IrMlPtrsOfNode ret_nodes = ·sliceOfPtrs(IrMlNode, 2, 2);
    ret_nodes.at[0] = n0;
    ret_nodes.at[1] = n1;
    return ret_nodes;
}




typedef struct IrMlCtxPrint {
    IrMlNode* cur_fn;
    ·ListOfPtrs(IrMlNode) fn_nodes_stack;
} IrMlCtxPrint;

void irmlPrintNode(IrMlCtxPrint* ctx, IrMlNode* const node) {
    ·assert(node != NULL);
    switch (node->kind) {
        case irml_node_fn: {
            Bool have_already = false;
            for (UInt i = 0; (!have_already) && i < ctx->fn_nodes_stack.len; i += 1)
                have_already = (ctx->fn_nodes_stack.at[i] == node);
            if (!have_already)
                ·append(ctx->fn_nodes_stack, node);
            printStr(node->anns.name);
        } break;
        case irml_node_param: {
            if (node->of.param.fn_node != ctx->cur_fn)
                printStr(node->of.param.fn_node->anns.name);
            printChr('@');
            printStr(uIntToStr(node->of.param.param_idx, 1, 10));
        } break;
        case irml_node_jump: {
            irmlPrintNode(ctx, node->of.jump.target);
            printChr('(');
            for (UInt i = 0; i < node->of.jump.args.len; i += 1) {
                if (i != 0)
                    printStr(str(", "));
                irmlPrintNode(ctx, node->of.jump.args.at[i]);
            }
            printChr(')');
        } break;
        case irml_node_prim: {
            switch (node->of.prim.kind) {
                case irml_prim_cond: {
                    printChr('(');
                    irmlPrintNode(ctx, node->of.prim.of.cond.scrutinee);
                    for (UInt i = 0; i < node->of.prim.of.cond.comparee_ints.len; i += 1) {
                        if (i == 0)
                            printStr(str(" ?- "));
                        else
                            printStr(str(" |- "));
                        irmlPrintNode(ctx, node->of.prim.of.cond.comparee_ints.at[i]);
                        printStr(str(" => "));
                        irmlPrintNode(ctx, node->of.prim.of.cond.match_results.at[i]);
                    }
                    if (node->of.prim.of.cond.default_result != NULL) {
                        printStr(str(" |- _ => "));
                        irmlPrintNode(ctx, node->of.prim.of.cond.default_result);
                    }
                    printChr(')');
                } break;
                case irml_prim_bini: {
                    printChr('(');
                    irmlPrintNode(ctx, node->of.prim.of.bini.lhs);
                    switch (node->of.prim.of.bini.kind) {
                        case irml_bini_add: printStr(str(" + ")); break;
                        case irml_bini_mul: printStr(str(" * ")); break;
                        case irml_bini_sub: printStr(str(" - ")); break;
                        case irml_bini_div: printStr(str(" / ")); break;
                        case irml_bini_rem: printStr(str(" \x25 ")); break;
                        default: ·fail(uIntToStr(node->of.prim.of.bini.kind, 1, 10));
                    }
                    irmlPrintNode(ctx, node->of.prim.of.bini.rhs);
                    printChr(')');
                } break;
                case irml_prim_cmpi: {
                    printChr('(');
                    irmlPrintNode(ctx, node->of.prim.of.cmpi.lhs);
                    switch (node->of.prim.of.cmpi.kind) {
                        case irml_cmpi_eq: printStr(str(" == ")); break;
                        case irml_cmpi_neq: printStr(str(" != ")); break;
                        case irml_cmpi_leq: printStr(str(" <= ")); break;
                        case irml_cmpi_geq: printStr(str(" >= ")); break;
                        case irml_cmpi_gt: printStr(str(" > ")); break;
                        case irml_cmpi_lt: printStr(str(" < ")); break;
                        default: ·fail(uIntToStr(node->of.prim.of.cmpi.kind, 1, 10));
                    }
                    irmlPrintNode(ctx, node->of.prim.of.cmpi.rhs);
                    printChr(')');
                } break;
                case irml_prim_cast: {
                    irmlPrintNode(ctx, node->of.prim.of.cast.dst_type);
                    printChr('(');
                    irmlPrintNode(ctx, node->of.prim.of.cast.subj);
                    printChr(')');
                } break;
                case irml_prim_extcall: {
                    printStr(node->of.prim.of.ext_call.name);
                    ·assert(irmlNodeIsPrimVal(node->of.prim.of.ext_call.args_list_val, irml_type_tup));
                    IrMlPtrsOfNode args = node->of.prim.of.ext_call.args_list_val->of.prim.of.val.of.list_val;
                    printChr('(');
                    for (UInt i = 0; i < args.len; i += 1) {
                        if (i != 0)
                            printStr(str(", "));
                        irmlPrintNode(ctx, args.at[i]);
                    }
                    printChr(')');
                } break;
                case irml_prim_item: {
                    irmlPrintNode(ctx, node->of.prim.of.item.subj);
                    printChr('[');
                    irmlPrintNode(ctx, node->of.prim.of.item.index);
                    printChr(']');
                    if (node->of.prim.of.item.set_to != NULL) {
                        printChr('=');
                        irmlPrintNode(ctx, node->of.prim.of.item.set_to);
                    }
                } break;
                case irml_prim_val: {
                    switch (node->of.prim.of.val.kind) {
                        case irml_type_void: printChr('_'); break;
                        case irml_type_int: printStr(uIntToStr(node->of.prim.of.val.of.int_val, 1, 10)); break;
                        case irml_type_tup:
                        case irml_type_arr: {
                            Bool const is_tup = (node->of.prim.of.val.kind == irml_type_tup);
                            printChr(is_tup ? '{' : '[');
                            for (UInt i = 0; i < node->of.prim.of.val.of.list_val.len; i += 1) {
                                if (i != 0)
                                    printStr(str(", "));
                                irmlPrintNode(ctx, node->of.prim.of.val.of.list_val.at[i]);
                            }
                            printChr(is_tup ? '}' : ']');
                        } break;
                        case irml_type_type: {
                            IrMlType* ty = &node->of.prim.of.val.of.type;
                            switch (ty->kind) {
                                case irml_type_type: printStr(str("@Type")); break;
                                case irml_type_void: printStr(str("@Void")); break;
                                case irml_type_int: {
                                    printChr('@');
                                    printChr(ty->of.num_int.unsign ? 'U' : 'I');
                                    printStr(uIntToStr(ty->of.num_int.bit_width, 1, 10));
                                } break;
                                case irml_type_arr:
                                    printChr('[');
                                    irmlPrintNode(ctx, ty->of.arr.length);
                                    printChr(']');
                                    irmlPrintNode(ctx, ty->of.arr.type);
                                    break;
                                case irml_type_ptr:
                                    printChr('*');
                                    irmlPrintNode(ctx, ty->of.ptr.type);
                                    break;
                                case irml_type_tup:
                                    printChr('{');
                                    for (UInt i = 0; i < ty->of.tup.types.len; i += 1) {
                                        if (i != 0)
                                            printStr(str(", "));
                                        irmlPrintNode(ctx, ty->of.tup.types.at[i]);
                                    }
                                    printChr('}');
                                    break;
                                case irml_type_fn:
                                    printStr(str("fn("));
                                    for (UInt i = 0; i < ty->of.tup.types.len; i += 1) {
                                        if (i != 0)
                                            printStr(str(", "));
                                        irmlPrintNode(ctx, ty->of.tup.types.at[i]);
                                    }
                                    printChr(')');
                                    break;
                            }
                        } break;
                        default: ·fail(uIntToStr(node->of.prim.of.val.kind, 1, 10));
                    }
                } break;
                default: ·fail(uIntToStr(node->of.prim.kind, 1, 10));
            }
        } break;
        default: ·fail(uIntToStr(node->kind, 1, 10));
    }
}

void irmlPrint(IrMlNode* const root_fn_node) {
    ·assert(root_fn_node != NULL);
    IrMlCtxPrint ctx = (IrMlCtxPrint) {.cur_fn = root_fn_node, .fn_nodes_stack = ·listOfPtrs(IrMlNode, 1, 8)};
    ctx.fn_nodes_stack.at[0] = root_fn_node;
    UInt idx = 0;
    while (idx < ctx.fn_nodes_stack.len) {
        UInt const max = ctx.fn_nodes_stack.len;
        for (UInt i = idx; i < max; i += 1) {
            IrMlNode* fn_node = ctx.fn_nodes_stack.at[i];
            if (fn_node->kind != irml_node_fn)
                irmlPrintNode(&ctx, fn_node);
            else {
                IrMlNodeFn* fn = &fn_node->of.fn;
                printStr(fn_node->anns.name);
                printChr('(');
                for (UInt j = 0; j < fn->params.len; j += 1) {
                    if (j != 0)
                        printStr(str(", "));
                    irmlPrintNode(&ctx, fn->params.at[j].anns.type);
                }
                printStr(str(")\n\t"));
                ctx.cur_fn = fn_node;
                irmlPrintNode(&ctx, fn->jump);
                printStr(str("\n\n"));
            }
        }
        idx = max;
    }
}




Bool irmlTypesEql(IrMlType const* const t1, IrMlType const* const t2) {
    Bool irmlNodesEql(IrMlNode const* const n1, IrMlNode const* const n2);
    if (t1 == t2)
        return true;
    if (t1 != NULL & t2 != NULL && t1->kind == t2->kind)
        switch (t1->kind) {
            case irml_type_void:
            case irml_type_type: return true;
            case irml_type_ptr: return irmlNodesEql(t1->of.ptr.type, t2->of.ptr.type);
            case irml_type_arr: return irmlNodesEql(t1->of.arr.length, t2->of.arr.length) && irmlNodesEql(t1->of.arr.type, t2->of.arr.type);
            case irml_type_int:
                return (t1->of.num_int.unsign == t2->of.num_int.unsign) && (t1->of.num_int.bit_width == t2->of.num_int.bit_width);
            case irml_type_tup:
            case irml_type_fn: {
                if (t1->of.tup.types.len == t2->of.tup.types.len) {
                    if (t1->of.tup.types.at != t2->of.tup.types.at)
                        for (UInt i = 0; i < t1->of.tup.types.len; i += 1)
                            if (!irmlNodesEql(t1->of.tup.types.at[i], t2->of.tup.types.at[i]))
                                return false;
                    return true;
                }
            } break;
            default: ·fail(uIntToStr(t1->kind, 1, 10));
        }
    return false;
}

UInt irmlTypeMinSizeInBits(IrMlProg* const prog, IrMlType* const type) {
    switch (type->kind) {
        case irml_type_ptr: return prog->bit_widths.ptrs;
        case irml_type_int: return type->of.num_int.bit_width;
        case irml_type_arr:
            if (irmlNodeIsPrimVal(type->of.arr.type, irml_type_type) && irmlNodeIsPrimVal(type->of.arr.length, irml_type_int))
                return type->of.arr.length->of.prim.of.val.of.int_val
                       * irmlTypeMinSizeInBits(prog, &type->of.arr.type->of.prim.of.val.of.type);
            else
                ·fail(str("arrays must be of sized payload types"));
        case irml_type_tup: {
            UInt size = 0;
            for (UInt i = 0; i < type->of.tup.types.len; i += 1)
                if (irmlNodeIsPrimVal(type->of.tup.types.at[i], irml_type_type))
                    size += irmlTypeMinSizeInBits(prog, &type->of.tup.types.at[i]->of.prim.of.val.of.type);
                else
                    ·fail(str("tuple fields must be of sized types"));
            return size;
        } break;
        case irml_type_type:
        case irml_type_void:
        case irml_type_fn: ·fail(str("expected a value of a sized type"));
        default: ·fail(uIntToStr(type->kind, 1, 10)); ;
    }
    return 0;
}

Bool irmlTypeIsIntCastable(IrMlType* type) {
    return type->kind == irml_type_int || type->kind == irml_type_ptr;
}

IrMlNode* irmlType(IrMlProg* const prog, IrMlKindOfType const kind, PtrAny const type_spec) {
    IrMlType specd_type = (IrMlType) {.kind = kind};
    if (kind != irml_type_void && kind != irml_type_type)
        switch (kind) {
            case irml_type_ptr: specd_type.of.ptr = *((IrMlTypePtr*)type_spec); break;
            case irml_type_arr: specd_type.of.arr = *((IrMlTypeArr*)type_spec); break;
            case irml_type_int: specd_type.of.num_int = *((IrMlTypeInt*)type_spec); break;
            case irml_type_fn:
            case irml_type_tup: specd_type.of.tup = *((IrMlTypeTup*)type_spec); break;
            default: ·fail(uIntToStr(kind, 1, 10));
        }
    IrMlNode* irmlNodePrimValType(IrMlProg* const prog, IrMlType spec);
    return irmlNodePrimValType(prog, specd_type);
}
IrMlNode* irmlTypePtr(IrMlProg* const prog, IrMlTypePtr type_spec) {
    return irmlType(prog, irml_type_ptr, &type_spec);
}
IrMlNode* irmlTypeArr(IrMlProg* const prog, IrMlTypeArr type_spec) {
    return irmlType(prog, irml_type_arr, &type_spec);
}
IrMlNode* irmlTypeTup(IrMlProg* const prog, IrMlTypeTup type_spec) {
    return irmlType(prog, irml_type_tup, &type_spec);
}
IrMlNode* irmlTypeFn(IrMlProg* const prog, IrMlTypeTup type_spec) {
    return irmlType(prog, irml_type_fn, &type_spec);
}
IrMlNode* irmlTypeFn0(IrMlProg* const prog) {
    IrMlPtrsOfNode params_type_nodes = ·sliceOfPtrs(IrMlNode, 0, 0);
    return irmlTypeFn(prog, (IrMlTypeTup) {.types = params_type_nodes});
}
IrMlNode* irmlTypeFn1(IrMlProg* const prog, IrMlNode* const param0_type) {
    IrMlPtrsOfNode params_type_nodes = ·sliceOfPtrs(IrMlNode, 1, 1);
    params_type_nodes.at[0] = param0_type;
    return irmlTypeFn(prog, (IrMlTypeTup) {.types = params_type_nodes});
}
IrMlNode* irmlTypeFn2(IrMlProg* const prog, IrMlNode* const param0_type, IrMlNode* const param1_type) {
    IrMlPtrsOfNode params_type_nodes = ·sliceOfPtrs(IrMlNode, 2, 2);
    params_type_nodes.at[0] = param0_type;
    params_type_nodes.at[1] = param1_type;
    return irmlTypeFn(prog, (IrMlTypeTup) {.types = params_type_nodes});
}
IrMlNode* irmlTypeInt(IrMlProg* const prog, IrMlTypeInt type_spec) {
    return irmlType(prog, irml_type_int, &type_spec);
}
IrMlNode* irmlTypeIntStatic(IrMlProg* const prog) {
    return irmlTypeInt(prog, (IrMlTypeInt) {.bit_width = 0, .unsign = false});
}
IrMlNode* irmlTypeVoid(IrMlProg* const prog) {
    return irmlType(prog, irml_type_void, NULL);
}
IrMlNode* irmlTypeBool(IrMlProg* const prog) {
    return irmlTypeInt(prog, (IrMlTypeInt) {.bit_width = 1, .unsign = true});
}
IrMlNode* irmlTypeLabel(IrMlProg* const prog) {
    return irmlTypeFn(prog, (IrMlTypeTup) {.types = ·sliceOfPtrs(IrMlNode, 0, 0)});
}
IrMlNode* irmlTypeType(IrMlProg* const prog) {
    return irmlType(prog, irml_type_type, NULL);
}



Bool irmlNodesEql(IrMlNode const* const n1, IrMlNode const* const n2) {
    if (n1 == n2)
        return true;
    if (n1 != NULL && n2 != NULL && n1->kind == n2->kind && irmlTypesEql(irmlNodeType(n1, false), irmlNodeType(n2, false)))
        switch (n1->kind) {
            case irml_node_jump: {
                if (n1->of.jump.args.len == n2->of.jump.args.len && irmlNodesEql(n1->of.jump.target, n2->of.jump.target)) {
                    if (n1->of.jump.args.at != n2->of.jump.args.at)
                        for (UInt i = 0; i < n1->of.jump.args.len; i += 1)
                            if (!irmlNodesEql(n1->of.jump.args.at[i], n2->of.jump.args.at[i]))
                                return false;
                    return true;
                }
                return false;
            }
            case irml_node_prim: {
                if (n1->of.prim.kind == n2->of.prim.kind)
                    switch (n1->of.prim.kind) {
                        case irml_prim_cond: {
                            if (irmlNodesEql(n1->of.prim.of.cond.scrutinee, n2->of.prim.of.cond.scrutinee)
                                && irmlNodesEql(n1->of.prim.of.cond.default_result, n2->of.prim.of.cond.default_result)
                                && n1->of.prim.of.cond.match_results.len == n2->of.prim.of.cond.match_results.len
                                && n1->of.prim.of.cond.comparee_ints.len == n2->of.prim.of.cond.comparee_ints.len) {
                                for (UInt i = 0; i < n1->of.prim.of.cond.comparee_ints.len; i += 1)
                                    if (!irmlNodesEql(n1->of.prim.of.cond.comparee_ints.at[i], n2->of.prim.of.cond.comparee_ints.at[i]))
                                        return false;
                                for (UInt i = 0; i < n1->of.prim.of.cond.match_results.len; i += 1)
                                    if (!irmlNodesEql(n1->of.prim.of.cond.match_results.at[i], n2->of.prim.of.cond.match_results.at[i]))
                                        return false;
                                return true;
                            }
                            return false;
                        } break;
                        case irml_prim_item:
                            return n1->of.prim.of.item.index == n2->of.prim.of.item.index
                                   && irmlNodesEql(n1->of.prim.of.item.set_to, n2->of.prim.of.item.set_to)
                                   && irmlNodesEql(n1->of.prim.of.item.subj, n2->of.prim.of.item.subj);
                        case irml_prim_cast:
                            return n1->of.prim.of.cast.kind == n2->of.prim.of.cast.kind
                                   && irmlNodesEql(n1->of.prim.of.cast.dst_type, n2->of.prim.of.cast.dst_type)
                                   && irmlNodesEql(n1->of.prim.of.cast.subj, n2->of.prim.of.cast.subj);
                        case irml_prim_bini:
                            return n1->of.prim.of.bini.kind == n2->of.prim.of.bini.kind
                                   && (irmlNodesEql(n1->of.prim.of.bini.lhs, n2->of.prim.of.bini.lhs)
                                       && irmlNodesEql(n1->of.prim.of.bini.rhs, n2->of.prim.of.bini.rhs));
                        case irml_prim_cmpi:
                            return n1->of.prim.of.cmpi.kind == n2->of.prim.of.cmpi.kind
                                   && (irmlNodesEql(n1->of.prim.of.cmpi.lhs, n2->of.prim.of.cmpi.lhs)
                                       && irmlNodesEql(n1->of.prim.of.cmpi.rhs, n2->of.prim.of.cmpi.rhs));
                        case irml_prim_extcall:
                            return irmlNodesEql(n1->of.prim.of.ext_call.args_list_val, n2->of.prim.of.ext_call.args_list_val)
                                   && strEql(n1->of.prim.of.ext_call.name, n2->of.prim.of.ext_call.name);
                        case irml_prim_val: {
                            IrMlPrimVal const* const v1 = &n1->of.prim.of.val;
                            IrMlPrimVal const* const v2 = &n2->of.prim.of.val;
                            if (v1->kind != v2->kind)
                                return false;
                            if ((v1->kind == irml_type_arr || v1->kind == irml_type_tup) && (v1->of.list_val.len == v2->of.list_val.len)) {
                                for (UInt i = 0; i < v1->of.list_val.len; i += 1)
                                    if (!irmlNodesEql(v1->of.list_val.at[i], v2->of.list_val.at[1]))
                                        return false;
                                return true;
                            }
                            return (v1->kind == irml_type_void) || (v1->kind == irml_type_int && v1->of.int_val == v2->of.int_val)
                                   || (v1->kind == irml_type_type && irmlTypesEql(&v1->of.type, &v2->of.type));
                        }
                        default: ·fail(uIntToStr(n1->of.prim.kind, 1, 10));
                    }
            } break;
            default: ·fail(uIntToStr(n1->kind, 1, 10));
        }
    return false;
}

IrMlNode* irmlNodeJump(IrMlProg* const prog, IrMlNodeJump const spec) {
    IrMlNode spec_node = (IrMlNode) {
        .kind = irml_node_jump,
        .anns = {.preduced = NULL, .type = NULL, .side_effects = false},
        .of = {.jump = spec},
    };
    for (UInt i = 0; i < prog->all.jumps.len; i += 1) {
        IrMlNode* node = prog->all.jumps.at[i];
        if (irmlNodesEql(node, &spec_node))
            return node;
    }

    ·append(prog->all.jumps, ·new(IrMlNode));
    IrMlNode* ret_node = prog->all.jumps.at[prog->all.jumps.len - 1];
    *ret_node = spec_node;
    return ret_node;
}

IrMlNode* irmlNodePrim(IrMlProg* const prog, IrMlNodePrim const spec, IrMlNode* const type) {
    IrMlNode const spec_node = (IrMlNode) {
        .kind = irml_node_prim,
        .of = {.prim = spec},
        .anns = {.preduced = NULL, .type = type, .side_effects = (spec.kind == irml_prim_extcall)},
    };
    for (UInt i = 0; i < prog->all.prims.len; i += 1) {
        IrMlNode* node = prog->all.prims.at[i];
        if (irmlNodesEql(node, &spec_node))
            return node;
    }

    ·append(prog->all.prims, ·new(IrMlNode));
    IrMlNode* ret_node = prog->all.prims.at[prog->all.prims.len - 1];
    *ret_node = spec_node;
    return ret_node;
}
IrMlNode* irmlNodePrimExtCall(IrMlProg* const prog, IrMlPrimExtCall const spec, IrMlNode* const ret_type) {
    return irmlNodePrim(prog, (IrMlNodePrim) {.kind = irml_prim_extcall, .of = {.ext_call = spec}}, ret_type);
}
IrMlNode* irmlNodePrimCast(IrMlProg* const prog, IrMlPrimCast spec) {
    return irmlNodePrim(prog, (IrMlNodePrim) {.kind = irml_prim_cast, .of = {.cast = spec}}, NULL);
}
IrMlNode* irmlNodePrimItem(IrMlProg* const prog, IrMlPrimItem spec) {
    return irmlNodePrim(prog, (IrMlNodePrim) {.kind = irml_prim_item, .of = {.item = spec}}, NULL);
}
IrMlNode* irmlNodePrimCmpI(IrMlProg* const prog, IrMlPrimCmpI spec) {
    return irmlNodePrim(prog, (IrMlNodePrim) {.kind = irml_prim_cmpi, .of = {.cmpi = spec}}, irmlTypeBool(prog));
}
IrMlNode* irmlNodePrimBinI(IrMlProg* const prog, IrMlPrimBinI spec) {
    return irmlNodePrim(prog, (IrMlNodePrim) {.kind = irml_prim_bini, .of = {.bini = spec}}, NULL);
}
IrMlNode* irmlNodePrimCond(IrMlProg* const prog, IrMlPrimCond spec) {
    return irmlNodePrim(prog, (IrMlNodePrim) {.kind = irml_prim_cond, .of = {.cond = spec}}, NULL);
}
IrMlNode* irmlNodePrimValArr(IrMlProg* const prog, IrMlPtrsOfNode const spec) {
    return irmlNodePrim(prog, (IrMlNodePrim) {.kind = irml_prim_val, .of = {.val = {.kind = irml_type_arr, .of = {.list_val = spec}}}}, NULL);
}
IrMlNode* irmlNodePrimValTup(IrMlProg* const prog, IrMlPtrsOfNode const spec) {
    return irmlNodePrim(prog, (IrMlNodePrim) {.kind = irml_prim_val, .of = {.val = {.kind = irml_type_tup, .of = {.list_val = spec}}}}, NULL);
}
IrMlNode* irmlNodePrimValType(IrMlProg* const prog, IrMlType spec) {
    IrMlNode* ret_node = irmlNodePrim(
        prog, (IrMlNodePrim) {.kind = irml_prim_val, .of = {.val = {.kind = irml_type_type, .of = {.type = spec}}}}, prog->all.prims.at[0]);
    ret_node->anns.preduced = ret_node;
    return ret_node;
}
IrMlNode* irmlNodePrimValInt(IrMlProg* const prog, I64 const spec) {
    IrMlNode* ret_node =
        irmlNodePrim(prog, (IrMlNodePrim) {.kind = irml_prim_val, .of = {.val = {.kind = irml_type_int, .of = {.int_val = spec}}}},
                     irmlTypeIntStatic(prog));
    ret_node->anns.preduced = ret_node;
    return ret_node;
}
IrMlNode* irmlNodePrimValVoid(IrMlProg* const prog) {
    IrMlNode* ret_node =
        irmlNodePrim(prog, (IrMlNodePrim) {.kind = irml_prim_val, .of = {.val = {.kind = irml_type_void}}}, irmlTypeVoid(prog));
    ret_node->anns.preduced = ret_node;
    return ret_node;
}
IrMlNode* irmlNodePrimValBool(IrMlProg* const prog, Bool const spec) {
    IrMlNode* ret_node = irmlNodePrim(
        prog, (IrMlNodePrim) {.kind = irml_prim_val, .of = {.val = {.kind = irml_type_int, .of = {.int_val = spec}}}}, irmlTypeBool(prog));
    ret_node->anns.preduced = ret_node;
    return ret_node;
}

IrMlNode* irmlNodeFn(IrMlProg* const prog, IrMlNode* const fn_type_node, CStr const maybe_name) {
    if ((!irmlNodeIsPrimVal(fn_type_node, irml_type_type)) || fn_type_node->of.prim.of.val.of.type.kind != irml_type_fn)
        ·fail(str("irmlNodeFn must be called with a fn_type_node that was produced by irmlTypeFn, irmlTypeFn0, irmlTypeFn1, etc."));
    IrMlPtrsOfNode params_type_nodes = fn_type_node->of.prim.of.val.of.type.of.tup.types;

    IrMlNode* ret_node = ·new(IrMlNode);
    *ret_node = (IrMlNode) {
        .kind = irml_node_fn,
        .of = {.fn = (IrMlNodeFn) {.jump = NULL, .params = ·sliceOf(IrMlNode, params_type_nodes.len, params_type_nodes.len)}},
        .anns = {.preduced = NULL, .type = fn_type_node, .side_effects = false, .name = (maybe_name == NULL) ? ·len0(U8) : str(maybe_name)},
    };
    for (UInt i = 0; i < params_type_nodes.len; i += 1)
        ret_node->of.fn.params.at[i] = (IrMlNode) {
            .kind = irml_node_param,
            .anns = {.preduced = NULL, .type = params_type_nodes.at[i], .side_effects = false},
            .of = {.param = (IrMlNodeParam) {.fn_node = ret_node, .param_idx = i, .anns = {.cur_evald = NULL}}},
        };
    return ret_node;
}

IrMlProg irmlProg(UInt bit_width_ptrs, UInt const prims_capacity, UInt const jumps_capacity) {
    IrMlProg ret_prog = (IrMlProg) {.all =
                                        {
                                            .prims = ·listOfPtrs(IrMlNode, 0, prims_capacity),
                                            .jumps = ·listOfPtrs(IrMlNode, 0, jumps_capacity),
                                        },
                                    .bit_widths = {
                                        .ptrs = bit_width_ptrs,
                                    }};

    irmlNodePrimValType(&ret_prog, (IrMlType) {.kind = irml_type_type}); // this creates entry 0 in all.prims:
    ret_prog.all.prims.at[0]->anns.type = ret_prog.all.prims.at[0];
    ret_prog.all.prims.at[0]->anns.preduced = ret_prog.all.prims.at[0];
    irmlNodePrimValInt(&ret_prog, 0)->anns.type = irmlTypeBool(&ret_prog);
    irmlNodePrimValInt(&ret_prog, 1)->anns.type = irmlTypeBool(&ret_prog);
    irmlNodePrimValInt(&ret_prog, -1)->anns.type = irmlTypeVoid(&ret_prog);
    irmlNodePrimValInt(&ret_prog, 0);
    irmlNodePrimValInt(&ret_prog, 1);
    return ret_prog;
}




void irmlFnJump(IrMlProg* const prog, IrMlNode* const fn_node, IrMlNodeJump const jump) {
    fn_node->of.fn.jump = irmlNodeJump(prog, jump);
}
IrMlPrimCond irmlCondBoolish(IrMlProg* const prog, IrMlNode* const scrutinee, IrMlNode* const if1, IrMlNode* const if0,
                             IrMlNode* const default_result) {
    IrMlPrimCond cond = (IrMlPrimCond) {.default_result = default_result,
                                        .scrutinee = scrutinee,
                                        .comparee_ints = (IrMlPtrsOfNode)·sliceOfPtrs(IrMlNode, 2, 2),
                                        .match_results = (IrMlPtrsOfNode)·sliceOfPtrs(IrMlNode, 2, 2)};
    cond.comparee_ints.at[0] = irmlNodePrimValBool(prog, true);
    cond.comparee_ints.at[1] = irmlNodePrimValBool(prog, false);
    cond.match_results.at[0] = if1;
    cond.match_results.at[1] = if0;
    return cond;
}




IrMlPtrsOfNode irmlUpdPtrsOfNodeSlice(IrMlProg* const prog, IrMlPtrsOfNode const nodes, IrMlPtrsOfNode upd) {
    if (upd.at != NULL && (upd.at != nodes.at || upd.len != nodes.len)) {
        Bool all_null = true;
        for (UInt i = 0; i < upd.len; i += 1)
            if (upd.at[i] != NULL)
                all_null = false;
            else if (i < nodes.len)
                upd.at[i] = nodes.at[i];
            else
                ·fail(str("BUG: tried to grow IrMlPtrsOfNode with NULLs"));
        if (all_null)
            upd.at = nodes.at;
    }
    return (upd.at == NULL || (upd.at == nodes.at && upd.len == nodes.len)) ? nodes : upd;
}

IrMlNode* irmlUpdNodeJump(IrMlProg* const prog, IrMlNode* const node, IrMlNodeJump upd) {
    if (upd.target == NULL)
        upd.target = node->of.jump.target;
    upd.args = irmlUpdPtrsOfNodeSlice(prog, node->of.jump.args, upd.args);
    if (upd.target == node->of.jump.target && upd.args.at == node->of.jump.args.at && upd.args.len == node->of.jump.args.len)
        return node;
    return irmlNodeJump(prog, upd);
}

IrMlNode* irmlUpdNodePrimItem(IrMlProg* const prog, IrMlNode* const node, IrMlPrimItem upd) {
    if (upd.index == NULL)
        upd.index = node->of.prim.of.item.index;
    if (upd.subj == NULL)
        upd.subj = node->of.prim.of.item.subj;
    if (upd.set_to == NULL)
        upd.set_to = node->of.prim.of.item.set_to;
    if (upd.index == node->of.prim.of.item.index && upd.subj == node->of.prim.of.item.subj && upd.set_to == node->of.prim.of.item.set_to)
        return node;
    return irmlNodePrimItem(prog, upd);
}

IrMlNode* irmlUpdNodePrimCond(IrMlProg* const prog, IrMlNode* const node, IrMlPrimCond upd) {
    if (upd.scrutinee == NULL)
        upd.scrutinee = node->of.prim.of.cond.scrutinee;
    if (upd.default_result == NULL)
        upd.default_result = node->of.prim.of.cond.default_result;
    upd.comparee_ints = irmlUpdPtrsOfNodeSlice(prog, node->of.prim.of.cond.comparee_ints, upd.comparee_ints);
    upd.match_results = irmlUpdPtrsOfNodeSlice(prog, node->of.prim.of.cond.match_results, upd.match_results);
    if (upd.comparee_ints.len == node->of.prim.of.cond.comparee_ints.len && upd.comparee_ints.at == node->of.prim.of.cond.comparee_ints.at
        && upd.match_results.len == node->of.prim.of.cond.match_results.len && upd.match_results.at == node->of.prim.of.cond.match_results.at
        && upd.default_result == node->of.prim.of.cond.default_result && upd.scrutinee == node->of.prim.of.cond.scrutinee)
        return node;
    return irmlNodePrimCond(prog, upd);
}

IrMlNode* irmlUpdNodePrimCast(IrMlProg* const prog, IrMlNode* const node, IrMlPrimCast upd) {
    if (upd.dst_type == NULL)
        upd.dst_type = node->of.prim.of.cast.dst_type;
    if (upd.subj == NULL)
        upd.subj = node->of.prim.of.cast.subj;
    if (upd.dst_type == node->of.prim.of.cast.dst_type && upd.subj == node->of.prim.of.cast.subj)
        return node;
    return irmlNodePrimCast(prog, upd);
}

IrMlNode* irmlUpdNodePrimValList(IrMlProg* const prog, IrMlNode* const node, IrMlPtrsOfNode upd) {
    IrMlPtrsOfNode const orig_list = node->of.prim.of.val.of.list_val;
    upd = irmlUpdPtrsOfNodeSlice(prog, orig_list, upd);
    if (upd.at == orig_list.at && upd.len == orig_list.len)
        return node;
    return irmlNodePrimValArr(prog, upd);
}

IrMlNode* irmlUpdNodePrimExtCall(IrMlProg* const prog, IrMlNode* const node, IrMlNode* upd_args, IrMlNode* upd_ret_type) {
    IrMlNode* const args_list = irmlUpdNodePrimValList(prog, node->of.prim.of.ext_call.args_list_val, upd_args->of.prim.of.val.of.list_val);
    if (upd_ret_type == NULL)
        upd_ret_type = node->anns.type;
    if (upd_ret_type == node->anns.type || args_list == node->of.prim.of.ext_call.args_list_val)
        return node;
    return irmlNodePrimExtCall(prog, (IrMlPrimExtCall) {.name = node->of.prim.of.ext_call.name, .args_list_val = args_list}, upd_ret_type);
}

IrMlNode* irmlUpdNodePrimBinI(IrMlProg* const prog, IrMlNode* const node, IrMlPrimBinI upd) {
    if (upd.lhs == NULL)
        upd.lhs = node->of.prim.of.bini.lhs;
    if (upd.rhs == NULL)
        upd.rhs = node->of.prim.of.bini.rhs;
    if (upd.lhs == node->of.prim.of.bini.lhs && upd.rhs == node->of.prim.of.bini.rhs)
        return node;
    return irmlNodePrimBinI(prog, upd);
}

IrMlNode* irmlUpdNodePrimCmpI(IrMlProg* const prog, IrMlNode* const node, IrMlPrimCmpI upd) {
    if (upd.lhs == NULL)
        upd.lhs = node->of.prim.of.cmpi.lhs;
    if (upd.rhs == NULL)
        upd.rhs = node->of.prim.of.cmpi.rhs;
    if (upd.lhs == node->of.prim.of.cmpi.lhs && upd.rhs == node->of.prim.of.cmpi.rhs)
        return node;
    return irmlNodePrimCmpI(prog, upd);
}




IrMlNode* irmlNodeWithParamsRewritten(IrMlProg* const prog, IrMlNode* const fn, IrMlNode* const node, IrMlPtrsOfNode const args) {
    if (node != NULL)
        switch (node->kind) {
            case irml_node_fn: break;

            case irml_node_param: {
                if (node->of.param.fn_node == fn)
                    return args.at[node->of.param.param_idx];
            } break;

            case irml_node_jump: {
                UInt const args_count = node->of.jump.args.len;
                Bool args_change = false;
                IrMlNodeJump new_jump = (IrMlNodeJump) {
                    .target = irmlNodeWithParamsRewritten(prog, fn, node->of.jump.target, args),
                    .args = ·sliceOfPtrs(IrMlNode, args_count, args_count),
                };
                for (UInt i = 0; i < new_jump.args.len; i += 1) {
                    new_jump.args.at[i] = irmlNodeWithParamsRewritten(prog, fn, node->of.jump.args.at[i], args);
                    args_change |= (new_jump.args.at[i] != NULL);
                }
                if (new_jump.target != NULL || args_change)
                    return irmlUpdNodeJump(prog, node, new_jump);
            } break;

            case irml_node_prim: {
                switch (node->of.prim.kind) {
                    case irml_prim_bini: {
                        IrMlPrimBinI const new_bini = (IrMlPrimBinI) {
                            .kind = node->of.prim.of.bini.kind,
                            .lhs = irmlNodeWithParamsRewritten(prog, fn, node->of.prim.of.bini.lhs, args),
                            .rhs = irmlNodeWithParamsRewritten(prog, fn, node->of.prim.of.bini.rhs, args),
                        };
                        if (new_bini.lhs != NULL || new_bini.rhs != NULL)
                            return irmlUpdNodePrimBinI(prog, node, new_bini);
                    } break;
                    case irml_prim_cmpi: {
                        IrMlPrimCmpI const new_cmpi = (IrMlPrimCmpI) {
                            .kind = node->of.prim.of.cmpi.kind,
                            .lhs = irmlNodeWithParamsRewritten(prog, fn, node->of.prim.of.cmpi.lhs, args),
                            .rhs = irmlNodeWithParamsRewritten(prog, fn, node->of.prim.of.cmpi.rhs, args),
                        };
                        if (new_cmpi.lhs != NULL || new_cmpi.rhs != NULL)
                            return irmlUpdNodePrimCmpI(prog, node, new_cmpi);
                    } break;
                    case irml_prim_cond: {
                        UInt const cases_count = node->of.prim.of.cond.comparee_ints.len;
                        Bool comparees_change = false;
                        Bool results_change = false;
                        IrMlPrimCond new_cond = (IrMlPrimCond) {
                            .comparee_ints = ·sliceOfPtrs(IrMlNode, cases_count, cases_count),
                            .match_results = ·sliceOfPtrs(IrMlNode, cases_count, cases_count),
                            .scrutinee = irmlNodeWithParamsRewritten(prog, fn, node->of.prim.of.cond.scrutinee, args),
                            .default_result = irmlNodeWithParamsRewritten(prog, fn, node->of.prim.of.cond.default_result, args),
                        };
                        for (UInt i = 0; i < cases_count; i += 1) {
                            new_cond.comparee_ints.at[i] =
                                irmlNodeWithParamsRewritten(prog, fn, node->of.prim.of.cond.comparee_ints.at[i], args);
                            new_cond.match_results.at[i] =
                                irmlNodeWithParamsRewritten(prog, fn, node->of.prim.of.cond.match_results.at[i], args);
                            comparees_change |= (new_cond.comparee_ints.at[i] != NULL);
                            results_change |= (new_cond.match_results.at[i] != NULL);
                        }
                        if (new_cond.scrutinee != NULL || new_cond.default_result != NULL || comparees_change || results_change)
                            return irmlUpdNodePrimCond(prog, node, new_cond);
                    } break;
                    case irml_prim_cast: {
                        IrMlPrimCast const new_cast = (IrMlPrimCast) {
                            .kind = node->of.prim.of.cast.kind,
                            .subj = irmlNodeWithParamsRewritten(prog, fn, node->of.prim.of.cast.subj, args),
                            .dst_type = irmlNodeWithParamsRewritten(prog, fn, node->of.prim.of.cast.dst_type, args),
                        };
                        if (new_cast.dst_type != NULL | new_cast.subj != NULL)
                            return irmlUpdNodePrimCast(prog, node, new_cast);
                    } break;
                    case irml_prim_item: {
                        IrMlPrimItem const new_item = (IrMlPrimItem) {
                            .subj = irmlNodeWithParamsRewritten(prog, fn, node->of.prim.of.item.subj, args),
                            .index = irmlNodeWithParamsRewritten(prog, fn, node->of.prim.of.item.index, args),
                            .set_to = irmlNodeWithParamsRewritten(prog, fn, node->of.prim.of.item.set_to, args),
                        };
                        if (new_item.index != NULL || new_item.subj != NULL || new_item.set_to != NULL)
                            return irmlUpdNodePrimItem(prog, node, new_item);
                    } break;
                    case irml_prim_extcall: {
                        IrMlPrimExtCall const new_call = (IrMlPrimExtCall) {
                            .name = node->of.prim.of.ext_call.name,
                            .args_list_val = irmlNodeWithParamsRewritten(prog, fn, node->of.prim.of.ext_call.args_list_val, args),
                        };
                        IrMlNode* const new_type = irmlNodeWithParamsRewritten(prog, fn, node->anns.type, args);
                        if (new_call.args_list_val != NULL || new_type != NULL)
                            return irmlUpdNodePrimExtCall(prog, node, new_call.args_list_val, new_type);
                    } break;
                    case irml_prim_val: {
                        if (node->of.prim.of.val.kind == irml_type_arr || node->of.prim.of.val.kind == irml_type_tup) {
                            UInt const len = node->of.prim.of.val.of.list_val.len;
                            IrMlPtrsOfNode new_list = ·sliceOfPtrs(IrMlNode, len, len);
                            Bool list_change = false;
                            for (UInt i = 0; i < len; i += 1) {
                                new_list.at[i] = irmlNodeWithParamsRewritten(prog, fn, node->of.prim.of.val.of.list_val.at[i], args);
                                list_change |= (new_list.at[i] != NULL);
                            }
                            if (list_change)
                                return irmlUpdNodePrimValList(prog, node, new_list);
                        }
                    } break;
                    default: ·fail(uIntToStr(node->of.prim.kind, 1, 10));
                }
            } break;

            default: ·fail(uIntToStr(node->kind, 1, 10));
        }
    return NULL;
}




typedef struct IrMlCtxPreduce {
    IrMlProg* prog;
    IrMlNode* cur_fn;
} IrMlCtxPreduce;

IrMlNode* irmlPreduceNode(IrMlCtxPreduce* const ctx, IrMlNode* const node);

IrMlNode* irmlPreduceNodeOfJump(IrMlCtxPreduce* const ctx, IrMlNode* const node) {
    IrMlNode* ret_node = NULL;

    UInt const args_count = node->of.jump.args.len;
    Bool args_change = false;
    IrMlNodeJump new_jump = (IrMlNodeJump) {
        .target = irmlPreduceNode(ctx, node->of.jump.target),
        .args = ·sliceOfPtrs(IrMlNode, args_count, args_count),
    };
    for (UInt i = 0; i < new_jump.args.len; i += 1) {
        new_jump.args.at[i] = irmlPreduceNode(ctx, node->of.jump.args.at[i]);
        args_change |= (new_jump.args.at[i] != NULL);
    }
    if (new_jump.target != NULL || args_change)
        ret_node = irmlUpdNodeJump(ctx->prog, node, new_jump);

    IrMlNode* chk_node = (ret_node == NULL) ? node : ret_node;
    IrMlType* const fn_type = irmlNodeType(chk_node->of.jump.target, true);
    if (fn_type->kind != irml_type_fn)
        ·fail(str("not a jump target"));
    if (fn_type->of.tup.types.len != chk_node->of.jump.args.len)
        ·fail(str4(str("jump target expected "), uIntToStr(fn_type->of.tup.types.len, 1, 10), str(" arg(s) but jump gave "),
                   uIntToStr(chk_node->of.jump.args.len, 1, 10)));
    for (UInt i = 0; i < chk_node->of.jump.args.len; i += 1) {
        IrMlNode* arg = chk_node->of.jump.args.at[i];
        if (arg->anns.type != fn_type->of.tup.types.at[i])
            ·fail(str2(str("type mismatch for jump arg "), uIntToStr(i, 1, 10)));
    }
    chk_node->anns.side_effects = chk_node->of.jump.target->anns.side_effects;
    for (UInt i = 0; (!chk_node->anns.side_effects) && i < chk_node->of.jump.args.len; i += 1)
        chk_node->anns.side_effects = (chk_node->of.jump.args.at[i]->anns.side_effects);

    while (true) {
        Bool can_inline = (chk_node->kind == irml_node_jump) && (chk_node->of.jump.target->kind == irml_node_fn)
                          && (chk_node->of.jump.target != ctx->cur_fn);
        for (UInt i = 0; can_inline && i < chk_node->of.jump.args.len; i += 1)
            if (chk_node->of.jump.args.at[i]->kind == irml_node_prim
                && !(irmlNodeIsPrimVal(chk_node->of.jump.args.at[i], irml_type_int)
                     || irmlNodeIsPrimVal(chk_node->of.jump.args.at[i], irml_type_type)
                     || irmlNodeIsPrimVal(chk_node->of.jump.args.at[i], irml_type_void)))
                can_inline = false;
        if (!can_inline)
            break;

        IrMlNode* inl_node =
            irmlNodeWithParamsRewritten(ctx->prog, chk_node->of.jump.target, chk_node->of.jump.target->of.fn.jump, chk_node->of.jump.args);
        if (inl_node == NULL)
            inl_node = chk_node->of.jump.target->of.fn.jump;
        IrMlNode* const pred_node = irmlPreduceNode(ctx, inl_node);
        ret_node = (pred_node == NULL) ? inl_node : pred_node;
        chk_node = ret_node;
    }

    return ret_node;
}

IrMlNode* irmlPreduceNodeOfFn(IrMlCtxPreduce* const ctx, IrMlNode* const node) {
    node->anns.preduced = node; // unlike all other node kinds, for irml_node_fn set this early
    ·assert(node->of.fn.jump != NULL);

    IrMlNode* const cur_fn = ctx->cur_fn;
    ctx->cur_fn = node;
    IrMlNode* jump = irmlPreduceNode(ctx, node->of.fn.jump);
    ctx->cur_fn = cur_fn;
    if (jump != NULL)
        node->of.fn.jump = jump;

    if (node->of.fn.jump->kind != irml_node_jump)
        ·fail(uIntToStr(node->of.fn.jump->kind, 1, 10));

    return NULL; // unlike all other node kinds, always NULL for irml_node_fn
}

IrMlNode* irmlPreduceNodeOfPrimVal(IrMlCtxPreduce* const ctx, IrMlNode* const node) {
    IrMlNode* ret_node = NULL;
    if (node->of.prim.of.val.kind == irml_type_arr || node->of.prim.of.val.kind == irml_type_tup) {
        IrMlPtrsOfNode new_list = ·sliceOfPtrs(IrMlNode, node->of.prim.of.val.of.list_val.len, 0);
        Bool all_null = true;
        for (UInt i = 0; i < new_list.len; i += 1) {
            new_list.at[i] = irmlPreduceNode(ctx, node->of.prim.of.val.of.list_val.at[i]);
            if (new_list.at[i] != NULL)
                all_null = false;
        }
        if (!all_null)
            ret_node = irmlUpdNodePrimValList(ctx->prog, node, new_list);

        IrMlNode* chk_node = (ret_node == NULL) ? node : ret_node;
        for (UInt i = 0; (!chk_node->anns.side_effects) && i < chk_node->of.prim.of.val.of.list_val.len; i += 1)
            chk_node->anns.side_effects = chk_node->of.prim.of.val.of.list_val.at[i]->anns.side_effects;
    }
    return ret_node;
}

IrMlNode* irmlPreduceNodeOfPrimCond(IrMlCtxPreduce* const ctx, IrMlNode* const node) {
    IrMlNode* ret_node = NULL;

    if (node->of.prim.of.cond.match_results.len != node->of.prim.of.cond.comparee_ints.len)
        ·fail(str("code-gen BUG: cond with unequal comparee_ints/match_results lengths"));
    if (node->of.prim.of.cond.comparee_ints.len == 0 && node->of.prim.of.cond.default_result == NULL)
        ·fail(str("code-gen BUG: cond with no match_results"));
    IrMlNode* ty_node = (node->of.prim.of.cond.default_result == NULL) ? NULL : node->of.prim.of.cond.default_result->anns.type;

    UInt const cases_count = node->of.prim.of.cond.match_results.len;
    IrMlPrimCond new_cond = (IrMlPrimCond) {.scrutinee = irmlPreduceNode(ctx, node->of.prim.of.cond.scrutinee),
                                            .default_result = NULL,
                                            .match_results = ·sliceOfPtrs(IrMlNode, cases_count, cases_count),
                                            .comparee_ints = ·sliceOfPtrs(IrMlNode, cases_count, cases_count)};
    IrMlNode* scrutinee = (new_cond.scrutinee == NULL) ? node->of.prim.of.cond.scrutinee : new_cond.scrutinee;
    if (irmlNodeType(scrutinee, true)->kind != irml_type_int)
        ·fail(str("cond scrutinee isn't integer"));
    Bool const is_scrut_static = irmlNodeIsPrimVal(scrutinee, irml_type_int);

    Bool comparees_change = false;
    Bool results_change = false;
    ºUInt found_case_if_static = ·none(UInt);
    for (UInt i = 0; i < new_cond.comparee_ints.len; i += 1) {
        new_cond.comparee_ints.at[i] = irmlPreduceNode(ctx, node->of.prim.of.cond.comparee_ints.at[i]);
        comparees_change |= (new_cond.comparee_ints.at[i] != NULL);
        IrMlNode* chk_node =
            (new_cond.comparee_ints.at[i] != NULL) ? new_cond.comparee_ints.at[i] : node->of.prim.of.cond.comparee_ints.at[i];
        if (!irmlNodeIsPrimVal(chk_node, irml_type_int))
            ·fail(str("cond case comparee did not preduce to statically-known int"));
        for (UInt j = 0; j < i; j += 1) {
            IrMlNode* const cmp_node =
                (new_cond.comparee_ints.at[j] != NULL) ? new_cond.comparee_ints.at[j] : node->of.prim.of.cond.comparee_ints.at[j];
            if (chk_node->of.prim.of.val.of.int_val == cmp_node->of.prim.of.val.of.int_val)
                ·fail(str("code-gen BUG: duplicate cases in cond"));
        }

        if ((!is_scrut_static) || chk_node->of.prim.of.val.of.int_val == scrutinee->of.prim.of.val.of.int_val) {
            new_cond.match_results.at[i] = irmlPreduceNode(ctx, node->of.prim.of.cond.match_results.at[i]);
            results_change |= (new_cond.match_results.at[i] != NULL);
            if (is_scrut_static) {
                ·assert(!found_case_if_static.ok);
                found_case_if_static = ·ok(UInt, i);
            }
            chk_node = (new_cond.match_results.at[i] != NULL) ? new_cond.match_results.at[i] : node->of.prim.of.cond.match_results.at[i];
            if (ty_node == NULL)
                ty_node = chk_node->anns.type;
            else if (ty_node != chk_node->anns.type)
                ·fail(str("code-gen BUG: type mismatches among cond match results"));
        } else
            new_cond.match_results.at[i] = NULL;
    }
    if (node->of.prim.of.cond.default_result != NULL && ((!is_scrut_static) || !found_case_if_static.ok))
        new_cond.default_result = irmlPreduceNode(ctx, node->of.prim.of.cond.default_result);

    if (new_cond.scrutinee != NULL || new_cond.default_result != NULL || results_change || comparees_change)
        ret_node = irmlUpdNodePrimCond(ctx->prog, node, new_cond);

    IrMlNode* chk_node = (ret_node == NULL) ? node : ret_node;
    chk_node->anns.type = ty_node;
    chk_node->anns.side_effects =
        (chk_node->of.prim.of.cond.default_result != NULL && chk_node->of.prim.of.cond.default_result->anns.side_effects)
        || chk_node->of.prim.of.cond.scrutinee->anns.side_effects;
    for (UInt i = 0; (!chk_node->anns.side_effects) && i < chk_node->of.prim.of.cond.match_results.len; i += 1)
        chk_node->anns.side_effects = chk_node->of.prim.of.cond.match_results.at[i]->anns.side_effects;

    if (is_scrut_static) {
        if (found_case_if_static.ok)
            ret_node = chk_node->of.prim.of.cond.match_results.at[found_case_if_static.it];
        else if (chk_node->of.prim.of.cond.default_result == NULL)
            ·fail(str("code-gen BUG: statically preducable cond with no match"));
        else
            ret_node = chk_node->of.prim.of.cond.default_result;
    }

    return ret_node;
}

IrMlNode* irmlPreduceNodeOfPrimCast(IrMlCtxPreduce* const ctx, IrMlNode* const node) {
    IrMlNode* ret_node = NULL;

    IrMlPrimCast new_cast = (IrMlPrimCast) {
        .kind = node->of.prim.of.cast.kind,
        .dst_type = irmlPreduceNode(ctx, node->of.prim.of.cast.dst_type),
        .subj = irmlPreduceNode(ctx, node->of.prim.of.cast.subj),
    };
    if (new_cast.subj != NULL || new_cast.dst_type != NULL)
        ret_node = irmlUpdNodePrimCast(ctx->prog, node, new_cast);

    IrMlNode* chk_node = (ret_node == NULL) ? node : ret_node;
    if (!irmlNodeIsPrimVal(chk_node->of.prim.of.cast.dst_type, irml_type_type))
        ·fail(str("cast requires type-typed destination"));
    IrMlType* const subj_type = irmlNodeType(chk_node->of.prim.of.cast.subj, true);
    if (chk_node->of.prim.of.cast.kind == irml_cast_ints
        && ((!irmlTypeIsIntCastable(&chk_node->of.prim.of.cast.dst_type->of.prim.of.val.of.type)) || (!irmlTypeIsIntCastable(subj_type))))
        ·fail(str("intcast requires int-castable source and destination types"));
    if (chk_node->of.prim.of.cast.kind == irml_cast_bits
        && irmlTypeMinSizeInBits(ctx->prog, &chk_node->of.prim.of.cast.dst_type->of.prim.of.val.of.type)
               != irmlTypeMinSizeInBits(ctx->prog, subj_type))
        ·fail(str("bitcast requires same bit-width for source and destination type"));
    chk_node->anns.type = chk_node->of.prim.of.cast.dst_type;
    chk_node->anns.side_effects = chk_node->of.prim.of.cast.subj->anns.side_effects;

    return ret_node;
}

IrMlNode* irmlPreduceNodeOfPrimExtCall(IrMlCtxPreduce* const ctx, IrMlNode* const node) {
    IrMlNode* ret_node = NULL;

    IrMlPrimExtCall new_call = (IrMlPrimExtCall) {
        .name = node->of.prim.of.ext_call.name,
        .args_list_val = irmlPreduceNode(ctx, node->of.prim.of.ext_call.args_list_val),
    };
    IrMlNode* new_ret_type = irmlPreduceNode(ctx, node->anns.type);
    if (new_ret_type != NULL || new_call.args_list_val != NULL) {
        if (!irmlNodeIsPrimVal(new_call.args_list_val, irml_type_tup))
            ·fail(str("specified illegal IrMlPrimExtCall.params_types"));
        ret_node = irmlUpdNodePrimExtCall(ctx->prog, node, new_call.args_list_val, new_ret_type);
    }

    IrMlNode* chk_node = (ret_node == NULL) ? node : ret_node;
    if (!irmlNodeIsPrimVal(chk_node->of.prim.of.ext_call.args_list_val, irml_type_tup))
        ·fail(str("specified illegal IrMlPrimExtCall.params_types"));
    if (chk_node->of.prim.of.ext_call.name.at == NULL || chk_node->of.prim.of.ext_call.name.len == 0)
        ·fail(str("specified illegal IrMlPrimExtCall.name"));
    chk_node->anns.side_effects = true;

    return ret_node;
}

IrMlNode* irmlPreduceNodeOfPrimItem(IrMlCtxPreduce* const ctx, IrMlNode* const node) {
    IrMlNode* ret_node = NULL;

    IrMlPrimItem new_item = (IrMlPrimItem) {
        .subj = irmlPreduceNode(ctx, node->of.prim.of.item.subj),
        .index = irmlPreduceNode(ctx, node->of.prim.of.item.index),
        .set_to = irmlPreduceNode(ctx, node->of.prim.of.item.set_to),
    };
    if (new_item.subj != NULL || new_item.index != NULL || new_item.set_to != NULL)
        ret_node = irmlUpdNodePrimItem(ctx->prog, node, new_item);

    IrMlNode* chk_node = (ret_node == NULL) ? node : ret_node;
    if (!irmlNodeIsPrimVal(chk_node->of.prim.of.item.index, irml_type_int))
        ·fail(str("expected statically-known index"));
    IrMlType* subj_type = irmlNodeType(chk_node->of.prim.of.item.subj, true);
    IrMlNode* node_type = (subj_type->kind == irml_type_tup)
                              ? subj_type->of.tup.types.at[chk_node->of.prim.of.item.index->of.prim.of.val.of.int_val]
                              : (subj_type->kind == irml_type_arr) ? subj_type->of.arr.type : NULL;
    if (!irmlNodeIsPrimVal(node_type, irml_type_type))
        ·fail(str("cannot index into this expression"));
    IrMlType* item_type = &node_type->of.prim.of.val.of.type;
    if (chk_node->of.prim.of.item.set_to != NULL && item_type != irmlNodeType(chk_node->of.prim.of.item.set_to, true))
        ·fail(str("type mismatch for setting aggregate member"));
    chk_node->anns.type = (chk_node->of.prim.of.item.set_to == NULL) ? chk_node->of.prim.of.item.subj->anns.type : node_type;
    chk_node->anns.side_effects = chk_node->of.prim.of.item.subj->anns.side_effects || chk_node->of.prim.of.item.index->anns.side_effects
                                  || (chk_node->of.prim.of.item.set_to != NULL && chk_node->of.prim.of.item.set_to->anns.side_effects);

    return ret_node;
}

IrMlNode* irmlPreduceNodeOfPrimCmpI(IrMlCtxPreduce* const ctx, IrMlNode* const node) {
    IrMlNode* ret_node = NULL;

    IrMlPrimCmpI new_cmpi = (IrMlPrimCmpI) {.kind = node->of.prim.of.cmpi.kind,
                                            .lhs = irmlPreduceNode(ctx, node->of.prim.of.cmpi.lhs),
                                            .rhs = irmlPreduceNode(ctx, node->of.prim.of.cmpi.rhs)};
    if (new_cmpi.lhs != NULL || new_cmpi.rhs != NULL)
        ret_node = irmlUpdNodePrimCmpI(ctx->prog, node, new_cmpi);

    IrMlNode* chk_node = (ret_node == NULL) ? node : ret_node;
    IrMlType* lhs_type = irmlNodeType(chk_node->of.prim.of.cmpi.lhs, true);
    IrMlType* rhs_type = irmlNodeType(chk_node->of.prim.of.cmpi.lhs, true);
    if (lhs_type != rhs_type || lhs_type->kind != irml_type_int)
        ·fail(str("invalid operand type(s) for int comparison operation"));
    chk_node->anns.type = irmlTypeBool(ctx->prog);
    chk_node->anns.side_effects = chk_node->of.prim.of.cmpi.lhs->anns.side_effects || chk_node->of.prim.of.cmpi.rhs->anns.side_effects;

    if (!chk_node->anns.side_effects) {
        if (chk_node->of.prim.of.cmpi.kind == irml_cmpi_eq && chk_node->of.prim.of.cmpi.lhs == chk_node->of.prim.of.cmpi.rhs)
            ret_node = irmlNodePrimValBool(ctx->prog, true);
        else if (irmlNodeIsPrimVal(chk_node->of.prim.of.cmpi.lhs, irml_type_int)
                 && irmlNodeIsPrimVal(chk_node->of.prim.of.cmpi.rhs, irml_type_int)) {
            I64 const lhs = chk_node->of.prim.of.cmpi.lhs->of.prim.of.val.of.int_val;
            I64 const rhs = chk_node->of.prim.of.cmpi.rhs->of.prim.of.val.of.int_val;
            Bool result;
            switch (chk_node->of.prim.of.cmpi.kind) {
                case irml_cmpi_eq: result = (lhs == rhs); break;
                case irml_cmpi_neq: result = (lhs != rhs); break;
                case irml_cmpi_geq: result = (lhs >= rhs); break;
                case irml_cmpi_leq: result = (lhs <= rhs); break;
                case irml_cmpi_gt: result = (lhs > rhs); break;
                case irml_cmpi_lt: result = (lhs < rhs); break;
                default: ·fail(str("TODO: how come we got a new int-cmp op?!"));
            }
            ret_node = irmlNodePrimValBool(ctx->prog, result);
        }
    }

    return ret_node;
}

IrMlNode* irmlPreduceNodeOfPrimBinI(IrMlCtxPreduce* const ctx, IrMlNode* const node) {
    IrMlNode* ret_node = NULL;

    IrMlPrimBinI new_bini = (IrMlPrimBinI) {.kind = node->of.prim.of.bini.kind,
                                            .lhs = irmlPreduceNode(ctx, node->of.prim.of.bini.lhs),
                                            .rhs = irmlPreduceNode(ctx, node->of.prim.of.bini.rhs)};
    if (new_bini.lhs != NULL || new_bini.rhs != NULL)
        ret_node = irmlUpdNodePrimBinI(ctx->prog, node, new_bini);

    IrMlNode* chk_node = (ret_node == NULL) ? node : ret_node;
    IrMlType* lhs_type = irmlNodeType(chk_node->of.prim.of.bini.lhs, true);
    IrMlType* rhs_type = irmlNodeType(chk_node->of.prim.of.bini.lhs, true);
    if (lhs_type != rhs_type || lhs_type->kind != irml_type_int)
        ·fail(str("invalid operand type(s) for int binary operation"));
    chk_node->anns.type = chk_node->of.prim.of.bini.lhs->anns.type;
    chk_node->anns.side_effects = chk_node->of.prim.of.bini.lhs->anns.side_effects || chk_node->of.prim.of.bini.rhs->anns.side_effects;

    if (chk_node->of.prim.of.bini.kind == irml_bini_rem && (chk_node->of.prim.of.bini.lhs == chk_node->of.prim.of.bini.rhs)
        && !chk_node->anns.side_effects)
        ret_node = irmlNodePrimValInt(ctx->prog, 0); // x%x=0
    else {
        IrMlNode* const lhs_static = irmlNodeIsPrimVal(chk_node->of.prim.of.bini.lhs, irml_type_int) ? chk_node->of.prim.of.bini.lhs : NULL;
        IrMlNode* const rhs_static = irmlNodeIsPrimVal(chk_node->of.prim.of.bini.rhs, irml_type_int) ? chk_node->of.prim.of.bini.rhs : NULL;
        Bool const l = (lhs_static != NULL);
        Bool const r = (rhs_static != NULL);
        if (l || r) {
            Bool const both = l && r;
            I64 const lhs = l ? lhs_static->of.prim.of.val.of.int_val : 0;
            I64 const rhs = r ? rhs_static->of.prim.of.val.of.int_val : 0;
            switch (chk_node->of.prim.of.bini.kind) {
                case irml_bini_add: {
                    if (r && rhs == 0) // x+0=x
                        ret_node = chk_node->of.prim.of.bini.lhs;
                    else if (l && lhs == 0) // 0+x=x
                        ret_node = chk_node->of.prim.of.bini.rhs;
                    else if (both)
                        ret_node = irmlNodePrimValInt(ctx->prog, lhs + rhs);
                } break;
                case irml_bini_sub: {
                    if (r && rhs == 0) // x-0=x
                        ret_node = chk_node->of.prim.of.bini.lhs;
                    else if (both)
                        ret_node = irmlNodePrimValInt(ctx->prog, lhs - rhs);
                } break;
                case irml_bini_mul: {
                    if (r && rhs == 1) // x*1=x
                        ret_node = chk_node->of.prim.of.bini.lhs;
                    else if (l && lhs == 1) // 1*x=x
                        ret_node = chk_node->of.prim.of.bini.rhs;
                    else if (both)
                        ret_node = irmlNodePrimValInt(ctx->prog, lhs * rhs);
                } break;
                case irml_bini_div: {
                    if (r && rhs == 0) // x/0=!
                        ·fail(str("div by zero"));
                    else if (r && rhs == 1) // x/1=x
                        ret_node = chk_node->of.prim.of.bini.lhs;
                    else if (l && lhs == 0 && !chk_node->anns.side_effects) // 0/x=0
                        ret_node = irmlNodePrimValInt(ctx->prog, 0);
                    else if (both)
                        ret_node = irmlNodePrimValInt(ctx->prog, lhs / rhs);
                } break;
                case irml_bini_rem: {
                    if (r && rhs == 0) // x%0=!
                        ·fail(str("rem by zero"));
                    else if (r && rhs == 1 && !chk_node->anns.side_effects) // x%1=0
                        ret_node = irmlNodePrimValInt(ctx->prog, 0);
                    else if (both)
                        ret_node = irmlNodePrimValInt(ctx->prog, lhs % rhs);
                } break;
            }
        }
    }
    return ret_node;
}

IrMlNode* irmlPreduceNode(IrMlCtxPreduce* const ctx, IrMlNode* const node) {
    IrMlNode* ret_node = NULL;
    if (node == NULL)
        ·fail(str("BUG: irmlPreduceNode called with NULL IrMlNode"));
    if (node->anns.preduced != NULL) {
        if (node->anns.preduced == node)
            return NULL; // dependant can keep their reference to `node`,
        else             // dependant picks up the already-previously-preduced instance of `node`
            return node->anns.preduced;
    }

    switch (node->kind) {
        case irml_node_param: break;
        case irml_node_fn: ret_node = irmlPreduceNodeOfFn(ctx, node); break;
        case irml_node_jump: ret_node = irmlPreduceNodeOfJump(ctx, node); break;
        case irml_node_prim: {
            switch (node->of.prim.kind) {
                case irml_prim_cond: ret_node = irmlPreduceNodeOfPrimCond(ctx, node); break;
                case irml_prim_item: ret_node = irmlPreduceNodeOfPrimItem(ctx, node); break;
                case irml_prim_extcall: ret_node = irmlPreduceNodeOfPrimExtCall(ctx, node); break;
                case irml_prim_cast: ret_node = irmlPreduceNodeOfPrimCast(ctx, node); break;
                case irml_prim_cmpi: ret_node = irmlPreduceNodeOfPrimCmpI(ctx, node); break;
                case irml_prim_bini: ret_node = irmlPreduceNodeOfPrimBinI(ctx, node); break;
                case irml_prim_val: ret_node = irmlPreduceNodeOfPrimVal(ctx, node); break;
                default: break;
            }
        } break;
        default: break;
    }

    IrMlNode* const one_node = (ret_node == NULL) ? node : ret_node;
    if (one_node->anns.type == NULL && one_node->kind != irml_node_jump)
        ·fail(str("untyped expr-node after preduce"));
    node->anns.type = one_node->anns.type;
    node->anns.side_effects = one_node->anns.side_effects;
    node->anns.preduced = one_node;
    one_node->anns.preduced = one_node;

    return ret_node;
}




IrMlNode* irmlEvalNode(IrMlProg* const prog, IrMlNode* const node) {
    if (node != NULL)
        switch (node->kind) {
            case irml_node_fn: break;
            case irml_node_param: return node->of.param.anns.cur_evald;
            case irml_node_prim: {
                switch (node->of.prim.kind) {
                    case irml_prim_val: {
                        if (node->of.prim.of.val.kind == irml_type_tup || node->of.prim.of.val.kind == irml_type_arr) {
                            // TODO..
                        }
                    } break;
                    case irml_prim_cond: {
                        IrMlNode* const scrut = irmlEvalNode(prog, node->of.prim.of.cond.scrutinee);
                        for (UInt i = 0; i < node->of.prim.of.cond.comparee_ints.len; i += 1)
                            if (scrut->of.prim.of.val.of.int_val == node->of.prim.of.cond.comparee_ints.at[i]->of.prim.of.val.of.int_val)
                                return irmlEvalNode(prog, node->of.prim.of.cond.match_results.at[i]);
                        return irmlEvalNode(prog, node->of.prim.of.cond.default_result);
                    } break;
                    case irml_prim_cmpi: {
                        IrMlNode* const lhs = irmlEvalNode(prog, node->of.prim.of.cmpi.lhs);
                        IrMlNode* const rhs = irmlEvalNode(prog, node->of.prim.of.cmpi.rhs);
                        switch (node->of.prim.of.cmpi.kind) {
                            case irml_cmpi_eq:
                                return irmlNodePrimValBool(prog, lhs->of.prim.of.val.of.int_val == rhs->of.prim.of.val.of.int_val);
                            case irml_cmpi_neq:
                                return irmlNodePrimValBool(prog, lhs->of.prim.of.val.of.int_val != rhs->of.prim.of.val.of.int_val);
                            case irml_cmpi_geq:
                                return irmlNodePrimValBool(prog, lhs->of.prim.of.val.of.int_val >= rhs->of.prim.of.val.of.int_val);
                            case irml_cmpi_leq:
                                return irmlNodePrimValBool(prog, lhs->of.prim.of.val.of.int_val <= rhs->of.prim.of.val.of.int_val);
                            case irml_cmpi_gt:
                                return irmlNodePrimValBool(prog, lhs->of.prim.of.val.of.int_val > rhs->of.prim.of.val.of.int_val);
                            case irml_cmpi_lt:
                                return irmlNodePrimValBool(prog, lhs->of.prim.of.val.of.int_val < rhs->of.prim.of.val.of.int_val);
                            default: ·fail(uIntToStr(node->of.prim.of.cmpi.kind, 1, 10));
                        }
                    } break;
                    default: ·fail(uIntToStr(node->of.prim.kind, 1, 10));
                }
            } break;
            default: ·fail(uIntToStr(node->kind, 1, 10));
        }
    return node;
}

IrMlNode* irmlEval(IrMlProg* const prog, IrMlNode* target, IrMlPtrsOfNode args) {
    while (target != NULL) {
        IrMlNodeFn* const fn = &target->of.fn;
        for (UInt i = 0; i < fn->params.len; i += 1)
            fn->params.at[i].of.param.anns.cur_evald = irmlEvalNode(prog, args.at[i]);
        target = irmlEvalNode(prog, fn->jump->of.jump.target);
        args = fn->jump->of.jump.args;
    }

    return irmlEvalNode(prog, args.at[0]);
}
