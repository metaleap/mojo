#pragma once
#include "utils_and_libc_deps.c"


typedef enum IrMlKindOfType {
    irml_type_type,
    irml_type_bottom,
    irml_type_int,
    irml_type_fn,
    irml_type_tup,
    irml_type_arr,
    irml_type_ptr,
} IrMlKindOfType;

typedef enum IrMlKindOfNode {
    irml_node_fn,
    irml_node_param,
    irml_node_choice,
    irml_node_jump,
    irml_node_prim,
} IrMlKindOfNode;

typedef enum IrMlKindOfPrim {
    irml_prim_cmp_i,
    irml_prim_bin_i,
    irml_prim_cast,
    irml_prim_item,
    irml_prim_extcall,
    irml_prim_val,
} IrMlKindOfPrim;

typedef enum IrMlKindOfCmpI {
    irml_cmp_i_eq,
    irml_cmp_i_neq,
} IrMlKindOfCmpI;

typedef enum IrMlKindOfBinI {
    irml_bin_i_add,
    irml_bin_i_mul,
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

typedef struct IrMlTypeBottom {
} IrMlTypeBottom;

typedef struct IrMlTypeTup {
    IrMlPtrsOfNode types;
} IrMlTypeTup;

typedef struct IrMlTypePtr {
    IrMlNode* type;
} IrMlTypePtr;

typedef struct IrMlTypeArr {
    IrMlNode* type;
    UInt length;
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
    UInt param_idx;
} IrMlNodeParam;

typedef struct IrMlNodeFn {
    IrMlNodes params;
    IrMlNode* body;
} IrMlNodeFn;

typedef struct IrMlNodeChoice {
    IrMlNode* cond;
    IrMlNode* if1;
    IrMlNode* if0;
} IrMlNodeChoice;

typedef struct IrMlNodeJump {
    IrMlNode* callee;
    IrMlPtrsOfNode args;
} IrMlNodeJump;

typedef struct IrMlPrimVal {
    union {
        I64 int_val;
        IrMlPtrsOfNode list_val;
        IrMlType type;
        struct {
        } bottom;
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
        IrMlPrimVal val;
        IrMlPrimCmpI cmp_i;
        IrMlPrimBinI bin_i;
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
        IrMlNodeChoice choice;
        IrMlNodeJump jump;
        IrMlNodePrim prim;
    } of;
    struct {
        IrMlNode* preduced;
        IrMlNode* type;
    } anns;
    IrMlKindOfNode kind;
};


typedef struct IrMlProg {
    struct {
        ·ListOfPtrs(IrMlNode) prims;
        ·ListOfPtrs(IrMlNode) choices;
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
    return (prim_kind == irml_prim_bin_i && (op_kind == irml_bin_i_add || op_kind == irml_bin_i_mul))
           || (prim_kind == irml_prim_cmp_i && (op_kind == irml_cmp_i_eq || op_kind == irml_cmp_i_neq));
}

// (a ° b) ° c  ==  a ° (b ° c)
Bool irmlPrimIsAssociative(IrMlKindOfPrim const prim_kind, int const op_kind) {
    return prim_kind == irml_prim_bin_i && (op_kind == irml_bin_i_add || op_kind == irml_bin_i_mul);
}

// a °¹ (b °² c)  ==  (a °¹ b)  °²  (a °¹ c)
Bool irmlPrimIsDistributive(IrMlKindOfPrim const prim_kind, int const op_kind1, int const op_kind2) {
    return prim_kind == irml_prim_bin_i && op_kind1 == irml_bin_i_mul && op_kind2 == irml_bin_i_add;
}

Bool irmlNodeIsPrimVal(IrMlNode const* const node, IrMlKindOfType const kind) {
    return node->kind == irml_node_prim && node->of.prim.kind == irml_prim_val && node->of.prim.of.val.kind == kind;
}

Bool irmlNodeIsBasicBlockishFn(IrMlNode* const node) {
    IrMlType* ty = irmlNodeType(node, true);
    return (ty->kind == irml_type_fn) && (ty->of.tup.types.len == 0);
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




Bool irmlTypesEql(IrMlType const* const t1, IrMlType const* const t2) {
    Bool irmlNodesEql(IrMlNode const* const n1, IrMlNode const* const n2);
    if (t1 == t2)
        return true;
    if (t1 != NULL & t2 != NULL && t1->kind == t2->kind)
        switch (t1->kind) {
            case irml_type_bottom:
            case irml_type_type: return true;
            case irml_type_ptr: return irmlNodesEql(t1->of.ptr.type, t2->of.ptr.type);
            case irml_type_arr: return (t1->of.arr.length == t2->of.arr.length) && irmlNodesEql(t1->of.arr.type, t2->of.arr.type);
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
            if (irmlNodeIsPrimVal(type->of.arr.type, irml_type_type))
                return type->of.arr.length * irmlTypeMinSizeInBits(prog, &type->of.arr.type->of.prim.of.val.of.type);
            else
                ·fail(str("arrays must be of of sized payload types"));
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
        case irml_type_bottom:
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
    if (kind != irml_type_bottom && kind != irml_type_type)
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
IrMlNode* irmlTypeBottom(IrMlProg* const prog) {
    return irmlType(prog, irml_type_bottom, NULL);
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
            case irml_node_choice: {
                return irmlNodesEql(n1->of.choice.if0, n2->of.choice.if0) && irmlNodesEql(n1->of.choice.if1, n2->of.choice.if1)
                       && irmlNodesEql(n1->of.choice.cond, n2->of.choice.cond);
            }
            case irml_node_jump: {
                if (n1->of.jump.args.len == n2->of.jump.args.len && irmlNodesEql(n1->of.jump.callee, n2->of.jump.callee)) {
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
                        case irml_prim_val: {
                            IrMlPrimVal const* const v1 = &n1->of.prim.of.val;
                            IrMlPrimVal const* const v2 = &n2->of.prim.of.val;
                            if (v1->kind != v2->kind)
                                break;

                            if ((v1->kind == irml_type_arr || v1->kind == irml_type_tup) && (v1->of.list_val.len == v2->of.list_val.len)) {
                                for (UInt i = 0; i < v1->of.list_val.len; i += 1)
                                    if (!irmlNodesEql(v1->of.list_val.at[i], v2->of.list_val.at[1]))
                                        return false;
                                return true;
                            }
                            return (v1->kind == irml_type_bottom) || (v1->kind == irml_type_int && v1->of.int_val == v2->of.int_val)
                                   || (v1->kind == irml_type_type && irmlTypesEql(&v1->of.type, &v2->of.type));
                        } break;
                        case irml_prim_item:
                            return n1->of.prim.of.item.index == n2->of.prim.of.item.index
                                   && irmlNodesEql(n1->of.prim.of.item.set_to, n2->of.prim.of.item.set_to)
                                   && irmlNodesEql(n1->of.prim.of.item.subj, n2->of.prim.of.item.subj);
                        case irml_prim_cast:
                            return n1->of.prim.of.cast.kind == n2->of.prim.of.cast.kind
                                   && irmlNodesEql(n1->of.prim.of.cast.dst_type, n2->of.prim.of.cast.dst_type)
                                   && irmlNodesEql(n1->of.prim.of.cast.subj, n2->of.prim.of.cast.subj);
                        case irml_prim_bin_i:
                            return n1->of.prim.of.bin_i.kind == n2->of.prim.of.bin_i.kind
                                   && (irmlNodesEql(n1->of.prim.of.bin_i.lhs, n2->of.prim.of.bin_i.lhs)
                                       && irmlNodesEql(n1->of.prim.of.bin_i.rhs, n2->of.prim.of.bin_i.rhs));
                        case irml_prim_cmp_i:
                            return n1->of.prim.of.cmp_i.kind == n2->of.prim.of.cmp_i.kind
                                   && (irmlNodesEql(n1->of.prim.of.cmp_i.lhs, n2->of.prim.of.cmp_i.lhs)
                                       && irmlNodesEql(n1->of.prim.of.cmp_i.rhs, n2->of.prim.of.cmp_i.rhs));
                        case irml_prim_extcall:
                            return irmlNodesEql(n1->of.prim.of.ext_call.args_list_val, n2->of.prim.of.ext_call.args_list_val)
                                   && strEql(n1->of.prim.of.ext_call.name, n2->of.prim.of.ext_call.name);
                        default: ·fail(uIntToStr(n1->of.prim.kind, 1, 10));
                    }
            } break;
            default: ·fail(uIntToStr(n1->kind, 1, 10));
        }
    return false;
}

IrMlNode* irmlNodeChoice(IrMlProg* const prog, IrMlNodeChoice const spec) {
    IrMlNode spec_node = (IrMlNode) {.kind = irml_node_choice, .of = {.choice = spec}, .anns = {.preduced = NULL, .type = NULL}};
    for (UInt i = 0; i < prog->all.choices.len; i += 1) {
        IrMlNode* node = prog->all.choices.at[i];
        if (irmlNodesEql(node, &spec_node))
            return node;
    }

    ·append(prog->all.choices, ·new(IrMlNode));
    IrMlNode* ret_node = prog->all.choices.at[prog->all.choices.len - 1];
    *ret_node = spec_node;
    return ret_node;
}

IrMlNode* irmlNodeJump(IrMlProg* const prog, IrMlNodeJump const spec) {
    IrMlNode spec_node = (IrMlNode) {
        .kind = irml_node_jump,
        .anns = {.preduced = NULL, .type = irmlTypeBottom(prog)},
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
    IrMlNode const spec_node = (IrMlNode) {.kind = irml_node_prim, .of = {.prim = spec}, .anns = {.preduced = NULL, .type = type}};
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
    return irmlNodePrim(prog, (IrMlNodePrim) {.kind = irml_prim_cmp_i, .of = {.cmp_i = spec}}, irmlTypeBool(prog));
}
IrMlNode* irmlNodePrimBinI(IrMlProg* const prog, IrMlPrimBinI spec) {
    return irmlNodePrim(prog, (IrMlNodePrim) {.kind = irml_prim_bin_i, .of = {.bin_i = spec}}, NULL);
}
IrMlNode* irmlNodePrimValArr(IrMlProg* const prog, IrMlPtrsOfNode const spec) {
    return irmlNodePrim(prog, (IrMlNodePrim) {.kind = irml_prim_val, .of = {.val = {.kind = irml_type_arr, .of = {.list_val = spec}}}},
                        irmlTypeArr(prog, (IrMlTypeArr) {.type = NULL, .length = spec.len}));
}
IrMlNode* irmlNodePrimValTup(IrMlProg* const prog, IrMlPtrsOfNode const spec) {
    return irmlNodePrim(prog, (IrMlNodePrim) {.kind = irml_prim_val, .of = {.val = {.kind = irml_type_tup, .of = {.list_val = spec}}}}, NULL);
}
IrMlNode* irmlNodePrimValType(IrMlProg* const prog, IrMlType spec) {
    return irmlNodePrim(prog, (IrMlNodePrim) {.kind = irml_prim_val, .of = {.val = {.kind = irml_type_type, .of = {.type = spec}}}},
                        prog->all.prims.at[0]);
}
IrMlNode* irmlNodePrimValInt(IrMlProg* const prog, I64 const spec) {
    return irmlNodePrim(prog, (IrMlNodePrim) {.kind = irml_prim_val, .of = {.val = {.kind = irml_type_int, .of = {.int_val = spec}}}},
                        irmlTypeIntStatic(prog));
}
IrMlNode* irmlNodePrimValBottom(IrMlProg* const prog) {
    return prog->all.prims.at[3];
}
IrMlNode* irmlNodePrimValBool(IrMlProg* const prog, Bool const spec) {
    return prog->all.prims.at[spec ? 2 : 1];
}

IrMlNode* irmlNodeFn(IrMlProg* const prog, IrMlNode* const fn_type_node) {
    if ((!irmlNodeIsPrimVal(fn_type_node, irml_type_type)) || fn_type_node->of.prim.of.val.of.type.kind != irml_type_fn)
        ·fail(str("irmlNodeFn must be called with a fn_type_node that was produced by irmlTypeFn, irmlTypeFn0, irmlTypeFn1, etc."));
    IrMlPtrsOfNode params_type_nodes = fn_type_node->of.prim.of.val.of.type.of.tup.types;

    IrMlNode* ret_node = ·new(IrMlNode);
    *ret_node = (IrMlNode) {
        .kind = irml_node_fn,
        .of = {.fn = (IrMlNodeFn) {.body = NULL, .params = ·sliceOf(IrMlNode, params_type_nodes.len, params_type_nodes.len)}},
        .anns = {.preduced = NULL, .type = fn_type_node},
    };
    for (UInt i = 0; i < params_type_nodes.len; i += 1)
        ret_node->of.fn.params.at[i] = (IrMlNode) {
            .kind = irml_node_param,
            .anns = {.preduced = NULL, .type = params_type_nodes.at[i]},
            .of = {.param = (IrMlNodeParam) {.param_idx = i}},
        };
    return ret_node;
}

IrMlProg irmlProg(UInt bit_width_ptrs, UInt const prims_capacity, UInt const choices_capacity, UInt const jumps_capacity) {
    IrMlProg ret_prog = (IrMlProg) {.all =
                                        {
                                            .prims = ·listOfPtrs(IrMlNode, 0, prims_capacity),
                                            .choices = ·listOfPtrs(IrMlNode, 0, choices_capacity),
                                            .jumps = ·listOfPtrs(IrMlNode, 0, jumps_capacity),
                                        },
                                    .bit_widths = {
                                        .ptrs = bit_width_ptrs,
                                    }};

    irmlNodePrimValType(&ret_prog, (IrMlType) {.kind = irml_type_type}); // this creates entry 0 in all.prims:
    ret_prog.all.prims.at[0]->anns.type = ret_prog.all.prims.at[0];
    irmlNodePrimValInt(&ret_prog, 0)->anns.type = irmlTypeBool(&ret_prog);
    irmlNodePrimValInt(&ret_prog, 1)->anns.type = irmlTypeBool(&ret_prog);
    irmlNodePrimValInt(&ret_prog, -1)->anns.type = irmlTypeBottom(&ret_prog);
    irmlNodePrimValInt(&ret_prog, 0);
    irmlNodePrimValInt(&ret_prog, 1);
    return ret_prog;
}




void irmlFnJump(IrMlProg* const prog, IrMlNode* const fn_node, IrMlNodeJump const jump) {
    fn_node->of.fn.body = irmlNodeJump(prog, jump);
}
void irmlFnChoice(IrMlProg* const prog, IrMlNode* const fn_node, IrMlNodeChoice const choice) {
    fn_node->of.fn.body = irmlNodeChoice(prog, choice);
}

IrMlNode* irmlUpdNodeChoice(IrMlProg* const prog, IrMlNode* const node, IrMlNodeChoice upd) {
    if (upd.cond == NULL)
        upd.cond = node->of.choice.cond;
    if (upd.if0 == NULL)
        upd.if0 = node->of.choice.if0;
    if (upd.if1 == NULL)
        upd.if1 = node->of.choice.if1;
    if (upd.if0 == node->of.choice.if0 && upd.if1 == node->of.choice.if1 && upd.cond == node->of.choice.cond)
        return node;
    return irmlNodeChoice(prog, upd);
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
    if (upd.callee == NULL)
        upd.callee = node->of.jump.callee;
    upd.args = irmlUpdPtrsOfNodeSlice(prog, node->of.jump.args, upd.args);
    if (upd.callee == node->of.jump.callee && upd.args.at == node->of.jump.args.at && upd.args.len == node->of.jump.args.len)
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
        upd.lhs = node->of.prim.of.bin_i.lhs;
    if (upd.rhs == NULL)
        upd.rhs = node->of.prim.of.bin_i.rhs;
    if (upd.lhs == node->of.prim.of.bin_i.lhs && upd.rhs == node->of.prim.of.bin_i.rhs)
        return node;
    return irmlNodePrimBinI(prog, upd);
}

IrMlNode* irmlUpdNodePrimCmpI(IrMlProg* const prog, IrMlNode* const node, IrMlPrimCmpI upd) {
    if (upd.lhs == NULL)
        upd.lhs = node->of.prim.of.cmp_i.lhs;
    if (upd.rhs == NULL)
        upd.rhs = node->of.prim.of.cmp_i.rhs;
    if (upd.lhs == node->of.prim.of.cmp_i.lhs && upd.rhs == node->of.prim.of.cmp_i.rhs)
        return node;
    return irmlNodePrimCmpI(prog, upd);
}



typedef struct IrMlCtxPreduce {
    IrMlProg* prog;
} IrMlCtxPreduce;

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

        case irml_node_fn: {
            node->anns.preduced = node; // unlike other node kinds, for irml_node_fn set this early
            if (node->of.fn.body == NULL) {
                // nothing to do, aka termination
            } else {
                IrMlNode* body = irmlPreduceNode(ctx, node->of.fn.body);
                if (body != NULL)
                    node->of.fn.body = body;
                // unlike other node kinds, for irml_node_fn our `ret_node` remains NULL
            }
        } break;

        case irml_node_choice: {
            IrMlNodeChoice new_choice = (IrMlNodeChoice) {.if0 = NULL, .if1 = NULL, .cond = irmlPreduceNode(ctx, node->of.choice.cond)};
            IrMlNode* cond = (new_choice.cond == NULL) ? node->of.choice.cond : new_choice.cond;
            if (!irmlNodesEql(cond->anns.type, irmlTypeBool(ctx->prog)))
                ·fail(str("choice condition isn't boolish"));
            Bool const is_cond_true = irmlNodesEql(cond, irmlNodePrimValBool(ctx->prog, true));
            Bool const is_cond_false = irmlNodesEql(cond, irmlNodePrimValBool(ctx->prog, false));
            Bool const is_cond_static = is_cond_true || is_cond_false;
            if (is_cond_false || !is_cond_static)
                new_choice.if0 = irmlPreduceNode(ctx, node->of.choice.if0);
            if (is_cond_true || !is_cond_static)
                new_choice.if1 = irmlPreduceNode(ctx, node->of.choice.if1);
            if (new_choice.cond != NULL || new_choice.if0 != NULL || new_choice.if1 != NULL)
                ret_node = irmlUpdNodeChoice(ctx->prog, node, new_choice);

            IrMlNode* chk_node = (ret_node == NULL) ? node : ret_node;
            if (chk_node->of.choice.if0->kind == irml_node_param || chk_node->of.choice.if1->kind == irml_node_param
                || !(irmlNodeIsBasicBlockishFn(chk_node->of.choice.if0) && irmlNodeIsBasicBlockishFn(chk_node->of.choice.if1)))
                ·fail(str("both non-param choices must preduce to basic blocks"));
            chk_node->anns.type = chk_node->of.choice.if0->anns.type;
            if (is_cond_true)
                ret_node = chk_node->of.choice.if1;
            else if (is_cond_false)
                ret_node = chk_node->of.choice.if0;
        } break;

        case irml_node_jump: {
            UInt const args_count = node->of.jump.args.len;
            Bool args_change = false;
            IrMlNodeJump new_jump = (IrMlNodeJump) {
                .callee = irmlPreduceNode(ctx, node->of.jump.callee),
                .args = ·sliceOfPtrs(IrMlNode, args_count, args_count),
            };
            for (UInt i = 0; i < new_jump.args.len; i += 1) {
                new_jump.args.at[i] = irmlPreduceNode(ctx, node->of.jump.args.at[i]);
                args_change |= (new_jump.args.at[i] != NULL);
            }
            if (new_jump.callee != NULL || args_change)
                ret_node = irmlUpdNodeJump(ctx->prog, node, new_jump);

            IrMlNode* chk_node = (ret_node == NULL) ? node : ret_node;
            IrMlType* fn_type = irmlNodeType(chk_node->of.jump.callee, true);
            if (fn_type->kind != irml_type_fn
                || !(chk_node->of.jump.callee->kind == irml_node_fn || chk_node->of.jump.callee->kind == irml_node_param))
                ·fail(str("not callable"));
            if (fn_type->of.tup.types.len != chk_node->of.jump.args.len)
                ·fail(str4(str("callee expected "), uIntToStr(fn_type->of.tup.types.len, 1, 10), str(" arg(s) but caller gave "),
                           uIntToStr(chk_node->of.jump.args.len, 1, 10)));
            for (UInt i = 0; i < chk_node->of.jump.args.len; i += 1) {
                IrMlNode* arg = chk_node->of.jump.args.at[i];
                if (!irmlNodesEql(arg->anns.type, fn_type->of.tup.types.at[i]))
                    ·fail(str2(str("type mismatch for arg "), uIntToStr(i, 1, 10)));
            }
        };

        case irml_node_prim: {
            switch (node->of.prim.kind) {
                case irml_prim_item: {
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
                    if (chk_node->of.prim.of.item.set_to != NULL
                        && !irmlTypesEql(item_type, irmlNodeType(chk_node->of.prim.of.item.set_to, true)))
                        ·fail(str("type mismatch for setting aggregate member"));
                    chk_node->anns.type = (chk_node->of.prim.of.item.set_to == NULL) ? chk_node->of.prim.of.item.subj->anns.type : node_type;
                } break;
                case irml_prim_extcall: {
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
                } break;
                case irml_prim_cast: {
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
                        && ((!irmlTypeIsIntCastable(&chk_node->of.prim.of.cast.dst_type->of.prim.of.val.of.type))
                            || (!irmlTypeIsIntCastable(subj_type))))
                        ·fail(str("intcast requires int-castable source and destination types"));
                    if (chk_node->of.prim.of.cast.kind == irml_cast_bits
                        && irmlTypeMinSizeInBits(ctx->prog, &chk_node->of.prim.of.cast.dst_type->of.prim.of.val.of.type)
                               != irmlTypeMinSizeInBits(ctx->prog, subj_type))
                        ·fail(str("bitcast requires same bit-width for source and destination type"));
                    chk_node->anns.type = chk_node->of.prim.of.cast.dst_type;
                } break;
                case irml_prim_cmp_i: {
                    IrMlPrimCmpI new_cmpi = (IrMlPrimCmpI) {.kind = node->of.prim.of.cmp_i.kind,
                                                            .lhs = irmlPreduceNode(ctx, node->of.prim.of.cmp_i.lhs),
                                                            .rhs = irmlPreduceNode(ctx, node->of.prim.of.cmp_i.rhs)};
                    if (new_cmpi.lhs != NULL || new_cmpi.rhs != NULL)
                        ret_node = irmlUpdNodePrimCmpI(ctx->prog, node, new_cmpi);

                    IrMlNode* chk_node = (ret_node == NULL) ? node : ret_node;
                    IrMlType* lhs_type = irmlNodeType(chk_node->of.prim.of.cmp_i.lhs, true);
                    IrMlType* rhs_type = irmlNodeType(chk_node->of.prim.of.cmp_i.lhs, true);
                    if ((!irmlTypesEql(lhs_type, rhs_type)) || lhs_type->kind != irml_type_int)
                        ·fail(str("invalid operand type(s) for int comparison operation"));
                    chk_node->anns.type = irmlTypeBool(ctx->prog);
                } break;
                case irml_prim_bin_i: {
                    IrMlPrimBinI new_bini = (IrMlPrimBinI) {.kind = node->of.prim.of.bin_i.kind,
                                                            .lhs = irmlPreduceNode(ctx, node->of.prim.of.bin_i.lhs),
                                                            .rhs = irmlPreduceNode(ctx, node->of.prim.of.bin_i.rhs)};
                    if (new_bini.lhs != NULL || new_bini.rhs != NULL)
                        ret_node = irmlUpdNodePrimBinI(ctx->prog, node, new_bini);

                    IrMlNode* chk_node = (ret_node == NULL) ? node : ret_node;
                    IrMlType* lhs_type = irmlNodeType(chk_node->of.prim.of.cmp_i.lhs, true);
                    IrMlType* rhs_type = irmlNodeType(chk_node->of.prim.of.cmp_i.lhs, true);
                    if ((!irmlTypesEql(lhs_type, rhs_type)) || lhs_type->kind != irml_type_int)
                        ·fail(str("invalid operand type(s) for int binary operation"));
                    chk_node->anns.type = chk_node->of.prim.of.cmp_i.lhs->anns.type;
                } break;
                case irml_prim_val:
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
                    }
                default: break;
            }
        } break;

        default: break;
    }

    IrMlNode* const the_non_null_node = (ret_node == NULL) ? node : ret_node;
    if (the_non_null_node->anns.type == NULL)
        ·fail(str("untyped node after preduce"));
    node->anns.type = the_non_null_node->anns.type;
    node->anns.preduced = the_non_null_node;
    the_non_null_node->anns.preduced = the_non_null_node;

    return ret_node;
}
