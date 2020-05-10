#pragma once
#include "utils_and_libc_deps.c"


typedef enum MtpKindOfType {
    mtp_type_type,
    mtp_type_bottom,
    mtp_type_sym,
    mtp_type_int,
    mtp_type_lam,
    mtp_type_tup,
    mtp_type_arr,
    mtp_type_ptr,
} MtpKindOfType;

typedef enum MtpKindOfNode {
    mtp_node_lam,
    mtp_node_param,
    mtp_node_choice,
    mtp_node_jump,
    mtp_node_prim,
} MtpKindOfNode;

typedef enum MtpKindOfPrim {
    mtp_prim_val,
    mtp_prim_cmp_i,
    mtp_prim_bin_i,
    mtp_prim_cast,
    mtp_prim_item,
    mtp_prim_extcall,
} MtpKindOfPrim;

typedef enum MtpKindOfCmpI {
    mtp_cmp_i_eq,
    mtp_cmp_i_neq,
} MtpKindOfCmpI;

typedef enum MtpKindOfBinI {
    mtp_bin_i_add,
    mtp_bin_i_mul,
} MtpKindOfBinI;

typedef enum MtpKindOfCast {
    mtp_cast_ints,
    mtp_cast_bits,
} MtpKindOfCast;


struct MtpNode;
typedef struct MtpNode MtpNode;
typedef ·SliceOfPtrs(MtpNode) MtpPtrsOfNode;
typedef ·SliceOf(MtpNode) MtpNodes;


struct MtpType;
typedef struct MtpType MtpType;

typedef struct MtpTypeInt {
    U16 bit_width;
    Bool unsign;
} MtpTypeInt;

typedef struct MtpTypeBottom {
} MtpTypeBottom;

typedef struct MtpTypeSym {
} MtpTypeSym;

typedef struct MtpTypeTup {
    MtpPtrsOfNode types;
} MtpTypeTup;

typedef struct MtpTypePtr {
    MtpNode* type;
} MtpTypePtr;

typedef struct MtpTypeArr {
    MtpNode* type;
    UInt length;
} MtpTypeArr;

struct MtpType {
    union {
        MtpTypeSym sym;
        MtpTypeInt num_int;
        MtpTypeTup tup;
        MtpTypePtr ptr;
        MtpTypeArr arr;
    } of;
    MtpKindOfType kind;
};


typedef struct MtpNodeParam {
    UInt param_idx;
} MtpNodeParam;

typedef struct MtpNodeLam {
    MtpNodes params;
    MtpNode* body;
} MtpNodeLam;

typedef struct MtpNodeChoice {
    MtpNode* cond;
    MtpNode* if1;
    MtpNode* if0;
} MtpNodeChoice;

typedef struct MtpNodeJump {
    MtpNode* callee;
    MtpPtrsOfNode args;
} MtpNodeJump;

typedef struct MtpPrimVal {
    union {
        I64 int_val;
        U32 sym_val;
        MtpPtrsOfNode list_val;
        MtpType type;
        struct {
        } bottom;
    } of;
    MtpKindOfType kind;
} MtpPrimVal;

typedef struct MtpPrimCmpI {
    MtpNode* lhs;
    MtpNode* rhs;
    MtpKindOfCmpI kind;
} MtpPrimCmpI;

typedef struct MtpPrimBinI {
    MtpNode* lhs;
    MtpNode* rhs;
    MtpKindOfBinI kind;
} MtpPrimBinI;

typedef struct MtpPrimCast {
    MtpNode* subj;
    MtpNode* dst_type;
    MtpKindOfCast kind;
} MtpPrimCast;

typedef struct MtpPrimItem {
    MtpNode* subj;
    MtpNode* index;
    MtpNode* set_to; // if NULL, it's a getter, else a setter
} MtpPrimItem;

typedef struct MtpPrimExtCall {
    MtpNode* args_list_val;
    Str name;
} MtpPrimExtCall;

typedef struct MtpNodePrim {
    union {
        MtpPrimVal val;
        MtpPrimCmpI cmp_i;
        MtpPrimBinI bin_i;
        MtpPrimCast cast;
        MtpPrimItem item;
        MtpPrimExtCall ext_call;
    } of;
    MtpKindOfPrim kind;
} MtpNodePrim;

struct MtpNode {
    MtpType* tyype;
    union {
        MtpNodeLam lam;
        MtpNodeParam param;
        MtpNodeChoice choice;
        MtpNodeJump jump;
        MtpNodePrim prim;
    } of;
    struct {
        Bool preduced;
        MtpNode* type;
    } anns;
    MtpKindOfNode kind;
};


typedef struct MtpProg {
    struct {
        ·ListOfPtrs(MtpNode) prims;
        ·ListOfPtrs(MtpNode) choices;
        ·ListOfPtrs(MtpNode) jumps;
    } all;
    struct {
        U16 ptrs;
        U16 syms;
    } bit_widths;
} MtpProg;




MtpType* mtpNodeType(MtpNode const* const node) {
    if (node != NULL && node->anns.type != NULL)
        return &node->anns.type->of.prim.of.val.of.type;
    return NULL;
}

// a ° b  ==  b ° a
Bool mtpPrimIsCommutative(MtpKindOfPrim const prim_kind, int const op_kind) {
    return (prim_kind == mtp_prim_bin_i && (op_kind == mtp_bin_i_add || op_kind == mtp_bin_i_mul))
           || (prim_kind == mtp_prim_cmp_i && (op_kind == mtp_cmp_i_eq || op_kind == mtp_cmp_i_neq));
}

// (a ° b) ° c  ==  a ° (b ° c)
Bool mtpPrimIsAssociative(MtpKindOfPrim const prim_kind, int const op_kind) {
    return prim_kind == mtp_prim_bin_i && (op_kind == mtp_bin_i_add || op_kind == mtp_bin_i_mul);
}

// a °¹ (b °² c)  ==  (a °¹ b)  °²  (a °¹ c)
Bool mtpPrimIsDistributive(MtpKindOfPrim const prim_kind, int const op_kind1, int const op_kind2) {
    return prim_kind == mtp_prim_bin_i && op_kind1 == mtp_bin_i_mul && op_kind2 == mtp_bin_i_add;
}

Bool mtpNodeIsPrimVal(MtpNode const* const node, MtpKindOfType const kind) {
    return node->kind == mtp_node_prim && node->of.prim.kind == mtp_prim_val && node->of.prim.of.val.kind == kind;
}

Bool mtpNodeIsBasicBlockishLam(MtpNode* const node) {
    MtpType* ty = mtpNodeType(node);
    return (ty != NULL) && (ty->kind == mtp_type_lam) && (ty->of.tup.types.len == 0);
}




Bool mtpTypesEql(MtpType const* const t1, MtpType const* const t2) {
    Bool mtpNodesEql(MtpNode const* const n1, MtpNode const* const n2);
    if (t1 == t2)
        return true;
    if (t1 != NULL & t2 != NULL && t1->kind == t2->kind)
        switch (t1->kind) {
            case mtp_type_sym:
            case mtp_type_bottom:
            case mtp_type_type: return true;
            case mtp_type_ptr: return mtpNodesEql(t1->of.ptr.type, t2->of.ptr.type);
            case mtp_type_arr: return (t1->of.arr.length == t2->of.arr.length) && mtpNodesEql(t1->of.arr.type, t2->of.arr.type);
            case mtp_type_int:
                return (t1->of.num_int.unsign == t2->of.num_int.unsign) && (t1->of.num_int.bit_width == t2->of.num_int.bit_width);
            case mtp_type_tup:
            case mtp_type_lam: {
                if (t1->of.tup.types.len == t2->of.tup.types.len) {
                    if (t1->of.tup.types.at != t2->of.tup.types.at)
                        for (UInt i = 0; i < t1->of.tup.types.len; i += 1)
                            if (!mtpNodesEql(t1->of.tup.types.at[i], t2->of.tup.types.at[i]))
                                return false;
                    return true;
                }
            } break;
            default: ·fail(uIntToStr(t1->kind, 1, 10));
        }
    return false;
}

UInt mtpTypeMinSizeInBits(MtpProg* const prog, MtpType* const type) {
    switch (type->kind) {
        case mtp_type_ptr: return prog->bit_widths.ptrs;
        case mtp_type_sym: return prog->bit_widths.syms;
        case mtp_type_int: return type->of.num_int.bit_width;
        case mtp_type_arr:
            if (mtpNodeIsPrimVal(type->of.arr.type, mtp_type_type))
                return type->of.arr.length * mtpTypeMinSizeInBits(prog, &type->of.arr.type->of.prim.of.val.of.type);
            else
                ·fail(str("arrays must be of of sized payload types"));
        case mtp_type_tup: {
            UInt size = 0;
            for (UInt i = 0; i < type->of.tup.types.len; i += 1)
                if (mtpNodeIsPrimVal(type->of.tup.types.at[i], mtp_type_type))
                    size += mtpTypeMinSizeInBits(prog, &type->of.tup.types.at[i]->of.prim.of.val.of.type);
                else
                    ·fail(str("tuple fields must be of sized types"));
            return size;
        } break;
        case mtp_type_type:
        case mtp_type_bottom:
        case mtp_type_lam: ·fail(str("expected a value of a sized type"));
        default: ·fail(uIntToStr(type->kind, 1, 10)); ;
    }
    return 0;
}

Bool mtpTypeIsIntCastable(MtpType* type) {
    return type->kind == mtp_type_int || type->kind == mtp_type_ptr || type->kind == mtp_type_sym;
}

MtpNode* mtpType(MtpProg* const prog, MtpKindOfType const kind, PtrAny const type_spec) {
    MtpType specd_type = (MtpType) {.kind = kind};
    if (kind != mtp_type_sym && kind != mtp_type_bottom && kind != mtp_type_type)
        switch (kind) {
            case mtp_type_ptr: specd_type.of.ptr = *((MtpTypePtr*)type_spec); break;
            case mtp_type_arr: specd_type.of.arr = *((MtpTypeArr*)type_spec); break;
            case mtp_type_int: specd_type.of.num_int = *((MtpTypeInt*)type_spec); break;
            case mtp_type_lam:
            case mtp_type_tup: specd_type.of.tup = *((MtpTypeTup*)type_spec); break;
            default: ·fail(uIntToStr(kind, 1, 10));
        }
    MtpNode* mtpNodePrimValType(MtpProg* const prog, MtpType spec);
    return mtpNodePrimValType(prog, specd_type);
}
MtpNode* mtpTypePtr(MtpProg* const prog, MtpTypePtr type_spec) {
    return mtpType(prog, mtp_type_ptr, &type_spec);
}
MtpNode* mtpTypeArr(MtpProg* const prog, MtpTypeArr type_spec) {
    return mtpType(prog, mtp_type_arr, &type_spec);
}
MtpNode* mtpTypeTup(MtpProg* const prog, MtpTypeTup type_spec) {
    return mtpType(prog, mtp_type_tup, &type_spec);
}
MtpNode* mtpTypeLam(MtpProg* const prog, MtpTypeTup type_spec) {
    return mtpType(prog, mtp_type_lam, &type_spec);
}
MtpNode* mtpTypeInt(MtpProg* const prog, MtpTypeInt type_spec) {
    return mtpType(prog, mtp_type_int, &type_spec);
}
MtpNode* mtpTypeIntStatic(MtpProg* const prog) {
    return mtpTypeInt(prog, (MtpTypeInt) {.bit_width = 0, .unsign = false});
}
MtpNode* mtpTypeSym(MtpProg* const prog) {
    return mtpType(prog, mtp_type_sym, NULL);
}
MtpNode* mtpTypeBottom(MtpProg* const prog) {
    return mtpType(prog, mtp_type_bottom, NULL);
}
MtpNode* mtpTypeBool(MtpProg* const prog) {
    return mtpTypeInt(prog, (MtpTypeInt) {.bit_width = 1, .unsign = true});
}
MtpNode* mtpTypeLabel(MtpProg* const prog) {
    return mtpTypeLam(prog, (MtpTypeTup) {.types = ·sliceOfPtrs(MtpNode, 0, 0)});
}
MtpNode* mtpTypeType(MtpProg* const prog) {
    return mtpType(prog, mtp_type_type, NULL);
}



Bool mtpNodesEql(MtpNode const* const n1, MtpNode const* const n2) {
    if (n1 == n2)
        return true;
    if (n1 != NULL && n2 != NULL && n1->kind == n2->kind && mtpTypesEql(mtpNodeType(n1), mtpNodeType(n2)))
        switch (n1->kind) {
            case mtp_node_choice: {
                return mtpNodesEql(n1->of.choice.if0, n2->of.choice.if0) && mtpNodesEql(n1->of.choice.if1, n2->of.choice.if1)
                       && mtpNodesEql(n1->of.choice.cond, n2->of.choice.cond);
            }
            case mtp_node_jump: {
                if (n1->of.jump.args.len == n2->of.jump.args.len && mtpNodesEql(n1->of.jump.callee, n2->of.jump.callee)) {
                    if (n1->of.jump.args.at != n2->of.jump.args.at)
                        for (UInt i = 0; i < n1->of.jump.args.len; i += 1)
                            if (!mtpNodesEql(n1->of.jump.args.at[i], n2->of.jump.args.at[i]))
                                return false;
                    return true;
                }
                return false;
            }
            case mtp_node_prim: {
                if (n1->of.prim.kind == n2->of.prim.kind)
                    switch (n1->of.prim.kind) {
                        case mtp_prim_val: {
                            MtpPrimVal const* const v1 = &n1->of.prim.of.val;
                            MtpPrimVal const* const v2 = &n2->of.prim.of.val;
                            if (v1->kind != v2->kind)
                                break;

                            if ((v1->kind == mtp_type_arr || v1->kind == mtp_type_tup) && (v1->of.list_val.len == v2->of.list_val.len)) {
                                for (UInt i = 0; i < v1->of.list_val.len; i += 1)
                                    if (!mtpNodesEql(v1->of.list_val.at[i], v2->of.list_val.at[1]))
                                        return false;
                                return true;
                            }
                            return (v1->kind == mtp_type_type && mtpTypesEql(&v1->of.type, &v2->of.type))
                                   || (v1->kind == mtp_type_int && v1->of.int_val == v2->of.int_val)
                                   || (v1->kind == mtp_type_sym && v1->of.sym_val == v2->of.sym_val) || (v1->kind == mtp_type_bottom);
                        } break;
                        case mtp_prim_item:
                            return n1->of.prim.of.item.index == n2->of.prim.of.item.index
                                   && mtpNodesEql(n1->of.prim.of.item.set_to, n2->of.prim.of.item.set_to)
                                   && mtpNodesEql(n1->of.prim.of.item.subj, n2->of.prim.of.item.subj);
                        case mtp_prim_cast:
                            return n1->of.prim.of.cast.kind == n2->of.prim.of.cast.kind
                                   && mtpNodesEql(n1->of.prim.of.cast.dst_type, n2->of.prim.of.cast.dst_type)
                                   && mtpNodesEql(n1->of.prim.of.cast.subj, n2->of.prim.of.cast.subj);
                        case mtp_prim_bin_i:
                            return n1->of.prim.of.bin_i.kind == n2->of.prim.of.bin_i.kind
                                   && (mtpNodesEql(n1->of.prim.of.bin_i.lhs, n2->of.prim.of.bin_i.lhs)
                                       && mtpNodesEql(n1->of.prim.of.bin_i.rhs, n2->of.prim.of.bin_i.rhs));
                        case mtp_prim_cmp_i:
                            return n1->of.prim.of.cmp_i.kind == n2->of.prim.of.cmp_i.kind
                                   && (mtpNodesEql(n1->of.prim.of.cmp_i.lhs, n2->of.prim.of.cmp_i.lhs)
                                       && mtpNodesEql(n1->of.prim.of.cmp_i.rhs, n2->of.prim.of.cmp_i.rhs));
                        case mtp_prim_extcall:
                            return mtpNodesEql(n1->of.prim.of.ext_call.args_list_val, n2->of.prim.of.ext_call.args_list_val)
                                   && strEql(n1->of.prim.of.ext_call.name, n2->of.prim.of.ext_call.name);
                        default: ·fail(uIntToStr(n1->of.prim.kind, 1, 10));
                    }
            } break;
            default: ·fail(uIntToStr(n1->kind, 1, 10));
        }
    return false;
}

MtpNode* mtpNodeChoice(MtpProg* const prog, MtpNodeChoice const spec) {
    MtpNode spec_node = (MtpNode) {.kind = mtp_node_choice, .of = {.choice = spec}, .anns = {.preduced = false, .type = NULL}};
    for (UInt i = 0; i < prog->all.choices.len; i += 1) {
        MtpNode* node = prog->all.choices.at[i];
        if (mtpNodesEql(node, &spec_node))
            return node;
    }

    ·append(prog->all.choices, ·new(MtpNode));
    MtpNode* ret_node = prog->all.choices.at[prog->all.choices.len - 1];
    *ret_node = spec_node;
    return ret_node;
}

MtpNode* mtpNodeJump(MtpProg* const prog, MtpNodeJump const spec) {
    MtpNode spec_node = (MtpNode) {
        .kind = mtp_node_jump,
        .anns = {.preduced = false, .type = mtpTypeBottom(prog)},
        .of = {.jump = spec},
    };
    for (UInt i = 0; i < prog->all.jumps.len; i += 1) {
        MtpNode* node = prog->all.jumps.at[i];
        if (mtpNodesEql(node, &spec_node))
            return node;
    }

    ·append(prog->all.jumps, ·new(MtpNode));
    MtpNode* ret_node = prog->all.jumps.at[prog->all.jumps.len - 1];
    *ret_node = spec_node;
    return ret_node;
}

MtpNode* mtpNodePrim(MtpProg* const prog, MtpNodePrim const spec, MtpNode* const type) {
    MtpNode const spec_node = (MtpNode) {.kind = mtp_node_prim, .of = {.prim = spec}, .anns = {.preduced = false, .type = type}};
    for (UInt i = 0; i < prog->all.prims.len; i += 1) {
        MtpNode* node = prog->all.prims.at[i];
        if (mtpNodesEql(node, &spec_node))
            return node;
    }

    ·append(prog->all.prims, ·new(MtpNode));
    MtpNode* ret_node = prog->all.prims.at[prog->all.prims.len - 1];
    *ret_node = spec_node;
    return ret_node;
}
MtpNode* mtpNodePrimExtCall(MtpProg* const prog, MtpPrimExtCall const spec, MtpNode* const ret_type) {
    return mtpNodePrim(prog, (MtpNodePrim) {.kind = mtp_prim_extcall, .of = {.ext_call = spec}}, ret_type);
}
MtpNode* mtpNodePrimCast(MtpProg* const prog, MtpPrimCast spec) {
    return mtpNodePrim(prog, (MtpNodePrim) {.kind = mtp_prim_cast, .of = {.cast = spec}}, NULL);
}
MtpNode* mtpNodePrimItem(MtpProg* const prog, MtpPrimItem spec) {
    return mtpNodePrim(prog, (MtpNodePrim) {.kind = mtp_prim_item, .of = {.item = spec}}, NULL);
}
MtpNode* mtpNodePrimCmpI(MtpProg* const prog, MtpPrimCmpI spec) {
    return mtpNodePrim(prog, (MtpNodePrim) {.kind = mtp_prim_cmp_i, .of = {.cmp_i = spec}}, mtpTypeBool(prog));
}
MtpNode* mtpNodePrimBinI(MtpProg* const prog, MtpPrimBinI spec) {
    return mtpNodePrim(prog, (MtpNodePrim) {.kind = mtp_prim_bin_i, .of = {.bin_i = spec}}, NULL);
}
MtpNode* mtpNodePrimValArr(MtpProg* const prog, MtpPtrsOfNode const spec) {
    return mtpNodePrim(prog, (MtpNodePrim) {.kind = mtp_prim_val, .of = {.val = {.kind = mtp_type_arr, .of = {.list_val = spec}}}},
                       mtpTypeArr(prog, (MtpTypeArr) {.type = NULL, .length = spec.len}));
}
MtpNode* mtpNodePrimValTup(MtpProg* const prog, MtpPtrsOfNode const spec) {
    return mtpNodePrim(prog, (MtpNodePrim) {.kind = mtp_prim_val, .of = {.val = {.kind = mtp_type_tup, .of = {.list_val = spec}}}}, NULL);
}
MtpNode* mtpNodePrimValType(MtpProg* const prog, MtpType spec) {
    return mtpNodePrim(prog, (MtpNodePrim) {.kind = mtp_prim_val, .of = {.val = {.kind = mtp_type_type, .of = {.type = spec}}}},
                       prog->all.prims.at[0]);
}
MtpNode* mtpNodePrimValInt(MtpProg* const prog, I64 const spec) {
    return mtpNodePrim(prog, (MtpNodePrim) {.kind = mtp_prim_val, .of = {.val = {.kind = mtp_type_int, .of = {.int_val = spec}}}},
                       mtpTypeIntStatic(prog));
}
MtpNode* mtpNodePrimValSym(MtpProg* const prog, U32 const spec) {
    return mtpNodePrim(prog, (MtpNodePrim) {.kind = mtp_prim_val, .of = {.val = {.kind = mtp_type_sym, .of = {.sym_val = spec}}}},
                       mtpTypeSym(prog));
}
MtpNode* mtpNodePrimValBottom(MtpProg* const prog) {
    return prog->all.prims.at[4];
}
MtpNode* mtpNodePrimValBool(MtpProg* const prog, Bool const spec) {
    return prog->all.prims.at[spec ? 3 : 2];
}

MtpNode* mtpNodeLam(MtpProg* const prog, MtpPtrsOfNode const params_type_nodes) {
    MtpNode* ret_node = ·new(MtpNode);
    *ret_node =
        (MtpNode) {.kind = mtp_node_lam,
                   .of = {.lam = (MtpNodeLam) {.body = NULL, .params = ·sliceOf(MtpNode, params_type_nodes.len, params_type_nodes.len)}},
                   .anns = {.preduced = false, .type = mtpTypeLam(prog, (MtpTypeTup) {.types = params_type_nodes})}};
    for (UInt i = 0; i < params_type_nodes.len; i += 1)
        ret_node->of.lam.params.at[i] = (MtpNode) {
            .kind = mtp_node_param,
            .anns = {.preduced = false, .type = params_type_nodes.at[i]},
            .of = {.param = (MtpNodeParam) {.param_idx = i}},
        };
    return ret_node;
}

MtpProg mtpProg(UInt bit_width_ptrs, UInt const prims_capacity, UInt const choices_capacity, UInt const jumps_capacity,
                U32 const sym_vals_total_count) {
    MtpProg ret_prog = (MtpProg) {.all =
                                      {
                                          .prims = ·listOfPtrs(MtpNode, 0, prims_capacity + sym_vals_total_count),
                                          .choices = ·listOfPtrs(MtpNode, 0, choices_capacity),
                                          .jumps = ·listOfPtrs(MtpNode, 0, jumps_capacity),
                                      },
                                  .bit_widths = {
                                      .ptrs = bit_width_ptrs,
                                      .syms = uIntMinSize(sym_vals_total_count - 1, 1),
                                  }};

    mtpNodePrimValType(&ret_prog, (MtpType) {.kind = mtp_type_type})->anns.type = ret_prog.all.prims.at[0];
    mtpTypeSym(&ret_prog)->anns.type = ret_prog.all.prims.at[0];
    mtpNodePrimValInt(&ret_prog, 0)->anns.type = mtpTypeBool(&ret_prog);
    mtpNodePrimValInt(&ret_prog, 1)->anns.type = mtpTypeBool(&ret_prog);
    mtpNodePrimValInt(&ret_prog, -1)->anns.type = mtpTypeBottom(&ret_prog);
    mtpNodePrimValInt(&ret_prog, 0);
    mtpNodePrimValInt(&ret_prog, 1);
    return ret_prog;
}




void mtpLamJump(MtpProg* const prog, MtpNodeLam* const lam, MtpNodeJump const jump) {
    lam->body = mtpNodeJump(prog, jump);
}
void mtpLamChoice(MtpProg* const prog, MtpNodeLam* const lam, MtpNodeChoice const choice) {
    lam->body = mtpNodeChoice(prog, choice);
}

MtpNode* mtpUpdNodeChoice(MtpProg* const prog, MtpNode* const node, MtpNodeChoice upd) {
    if (upd.cond == NULL)
        upd.cond = node->of.choice.cond;
    if (upd.if0 == NULL)
        upd.if0 = node->of.choice.if0;
    if (upd.if1 == NULL)
        upd.if1 = node->of.choice.if1;
    if (upd.if0 == node->of.choice.if0 && upd.if1 == node->of.choice.if1 && upd.cond == node->of.choice.cond)
        return node;
    return mtpNodeChoice(prog, upd);
}

MtpPtrsOfNode mtpUpdPtrsOfNodeSlice(MtpProg* const prog, MtpPtrsOfNode const nodes, MtpPtrsOfNode upd) {
    if (upd.at != NULL && (upd.at != nodes.at || upd.len != nodes.len)) {
        Bool all_null = true;
        for (UInt i = 0; i < upd.len; i += 1)
            if (upd.at[i] != NULL)
                all_null = false;
            else if (i < nodes.len)
                upd.at[i] = nodes.at[i];
            else
                ·fail(str("BUG: tried to grow MtpPtrsOfNode with NULLs"));
        if (all_null)
            upd.at = nodes.at;
    }
    return (upd.at == NULL || (upd.at == nodes.at && upd.len == nodes.len)) ? nodes : upd;
}

MtpNode* mtpUpdNodeJump(MtpProg* const prog, MtpNode* const node, MtpNodeJump upd) {
    if (upd.callee == NULL)
        upd.callee = node->of.jump.callee;
    upd.args = mtpUpdPtrsOfNodeSlice(prog, node->of.jump.args, upd.args);
    if (upd.callee == node->of.jump.callee && upd.args.at == node->of.jump.args.at && upd.args.len == node->of.jump.args.len)
        return node;
    return mtpNodeJump(prog, upd);
}

MtpNode* mtpUpdNodePrimItem(MtpProg* const prog, MtpNode* const node, MtpPrimItem upd) {
    if (upd.index == NULL)
        upd.index = node->of.prim.of.item.index;
    if (upd.subj == NULL)
        upd.subj = node->of.prim.of.item.subj;
    if (upd.set_to == NULL)
        upd.set_to = node->of.prim.of.item.set_to;
    if (upd.index == node->of.prim.of.item.index && upd.subj == node->of.prim.of.item.subj && upd.set_to == node->of.prim.of.item.set_to)
        return node;
    return mtpNodePrimItem(prog, upd);
}

MtpNode* mtpUpdNodePrimCast(MtpProg* const prog, MtpNode* const node, MtpPrimCast upd) {
    if (upd.dst_type == NULL)
        upd.dst_type = node->of.prim.of.cast.dst_type;
    if (upd.subj == NULL)
        upd.subj = node->of.prim.of.cast.subj;
    if (upd.dst_type == node->of.prim.of.cast.dst_type && upd.subj == node->of.prim.of.cast.subj)
        return node;
    return mtpNodePrimCast(prog, upd);
}

MtpNode* mtpUpdNodePrimValList(MtpProg* const prog, MtpNode* const node, MtpPtrsOfNode upd) {
    MtpPtrsOfNode const orig_list = node->of.prim.of.val.of.list_val;
    upd = mtpUpdPtrsOfNodeSlice(prog, orig_list, upd);
    if (upd.at == orig_list.at && upd.len == orig_list.len)
        return node;
    return mtpNodePrimValArr(prog, upd);
}

MtpNode* mtpUpdNodePrimExtCall(MtpProg* const prog, MtpNode* const node, MtpNode* upd_args, MtpNode* upd_ret_type) {
    MtpNode* const args_list = mtpUpdNodePrimValList(prog, node->of.prim.of.ext_call.args_list_val, upd_args->of.prim.of.val.of.list_val);
    if (upd_ret_type == NULL)
        upd_ret_type = node->anns.type;
    if (upd_ret_type == node->anns.type || args_list == node->of.prim.of.ext_call.args_list_val)
        return node;
    return mtpNodePrimExtCall(prog, (MtpPrimExtCall) {.name = node->of.prim.of.ext_call.name, .args_list_val = args_list}, upd_ret_type);
}

MtpNode* mtpUpdNodePrimBinI(MtpProg* const prog, MtpNode* const node, MtpPrimBinI upd) {
    if (upd.lhs == NULL)
        upd.lhs = node->of.prim.of.bin_i.lhs;
    if (upd.rhs == NULL)
        upd.rhs = node->of.prim.of.bin_i.rhs;
    if (upd.lhs == node->of.prim.of.bin_i.lhs && upd.rhs == node->of.prim.of.bin_i.rhs)
        return node;
    return mtpNodePrimBinI(prog, upd);
}

MtpNode* mtpUpdNodePrimCmpI(MtpProg* const prog, MtpNode* const node, MtpPrimCmpI upd) {
    if (upd.lhs == NULL)
        upd.lhs = node->of.prim.of.cmp_i.lhs;
    if (upd.rhs == NULL)
        upd.rhs = node->of.prim.of.cmp_i.rhs;
    if (upd.lhs == node->of.prim.of.cmp_i.lhs && upd.rhs == node->of.prim.of.cmp_i.rhs)
        return node;
    return mtpNodePrimCmpI(prog, upd);
}



typedef struct MtpCtxPreduce {
    MtpProg* prog;
} MtpCtxPreduce;

MtpNode* mtpPreduceNode(MtpCtxPreduce* const ctx, MtpNode* const node) {
    MtpNode* ret_node = NULL;
    if (node == NULL || node->anns.preduced)
        return NULL;
    switch (node->kind) {

        case mtp_node_lam: {
            node->anns.preduced = true;
            MtpNode* body = mtpPreduceNode(ctx, node->of.lam.body);
            if (body != NULL)
                node->of.lam.body = body;
            // ret_node stays NULL for lams: they're the only "mutable" node kind
        } break;

        case mtp_node_choice: {
            MtpNodeChoice new_choice = (MtpNodeChoice) {.if0 = NULL, .if1 = NULL, .cond = mtpPreduceNode(ctx, node->of.choice.cond)};
            if (!mtpNodesEql(new_choice.cond->anns.type, mtpTypeBool(ctx->prog)))
                ·fail(str("choice condition isn't boolish"));
            Bool const is_cond_true =
                mtpNodesEql((new_choice.cond != NULL) ? new_choice.cond : node->of.choice.cond, mtpNodePrimValBool(ctx->prog, true));
            Bool const is_cond_false =
                mtpNodesEql((new_choice.cond != NULL) ? new_choice.cond : node->of.choice.cond, mtpNodePrimValBool(ctx->prog, false));
            Bool const is_cond_static = is_cond_true || is_cond_false;
            if (is_cond_false || !is_cond_static)
                new_choice.if0 = mtpPreduceNode(ctx, node->of.choice.if0);
            if (is_cond_true || !is_cond_static)
                new_choice.if1 = mtpPreduceNode(ctx, node->of.choice.if1);
            if (new_choice.cond != NULL || new_choice.if0 != NULL || new_choice.if1 != 0)
                ret_node = mtpUpdNodeChoice(ctx->prog, node, new_choice);

            MtpNode* chk_node = (ret_node == NULL) ? node : ret_node;
            if (!(mtpNodeIsBasicBlockishLam(chk_node->of.choice.if0) && mtpNodeIsBasicBlockishLam(chk_node->of.choice.if1)))
                ·fail(str("choice results must both preduce to basic blocks"));
            chk_node->anns.type = chk_node->of.choice.if0->anns.type;
            if (is_cond_true)
                ret_node = chk_node->of.choice.if1;
            else if (is_cond_false)
                ret_node = chk_node->of.choice.if0;
        } break;

        case mtp_node_jump: {
            UInt const args_count = node->of.jump.args.len;
            Bool args_change = false;
            MtpNodeJump new_jump = (MtpNodeJump) {
                .callee = mtpPreduceNode(ctx, node->of.jump.callee),
                .args = ·sliceOfPtrs(MtpNode, args_count, args_count),
            };
            for (UInt i = 0; i < new_jump.args.len; i += 1) {
                new_jump.args.at[i] = mtpPreduceNode(ctx, node->of.jump.args.at[i]);
                args_change |= (new_jump.args.at[i] != NULL);
            }
            if (new_jump.callee != NULL || args_change)
                ret_node = mtpUpdNodeJump(ctx->prog, node, new_jump);

            MtpNode* chk_node = (ret_node == NULL) ? node : ret_node;
            MtpType* fn_type = mtpNodeType(chk_node->of.jump.callee);
            if (fn_type->kind != mtp_type_lam
                || !(chk_node->of.jump.callee->kind == mtp_node_lam || chk_node->of.jump.callee->kind == mtp_node_param))
                ·fail(str("not callable"));
            if (fn_type->of.tup.types.len != chk_node->of.jump.args.len)
                ·fail(str4(str("callee expected "), uIntToStr(chk_node->of.jump.callee->of.lam.params.len, 1, 10),
                           str(" arg(s) but caller gave "), uIntToStr(chk_node->of.jump.args.len, 1, 10)));
            for (UInt i = 0; i < chk_node->of.jump.args.len; i += 1) {
                MtpNode* arg = chk_node->of.jump.args.at[i];
                MtpNode* param = &chk_node->of.jump.callee->of.lam.params.at[i];
                if (!mtpNodesEql(arg->anns.type, param->anns.type))
                    ·fail(str2(str("type mismatch for arg "), uIntToStr(i, 1, 10)));
            }
        };

        case mtp_node_prim: {
            switch (node->of.prim.kind) {
                case mtp_prim_item: {
                    MtpPrimItem new_item = (MtpPrimItem) {
                        .subj = mtpPreduceNode(ctx, node->of.prim.of.item.subj),
                        .index = mtpPreduceNode(ctx, node->of.prim.of.item.index),
                        .set_to = mtpPreduceNode(ctx, node->of.prim.of.item.set_to),
                    };
                    if (new_item.subj != NULL || new_item.index != NULL || new_item.set_to != NULL)
                        ret_node = mtpUpdNodePrimItem(ctx->prog, node, new_item);

                    MtpNode* chk_node = (ret_node == NULL) ? node : ret_node;
                    if (!mtpNodeIsPrimVal(chk_node->of.prim.of.item.index, mtp_type_int))
                        ·fail(str("expected statically-known index"));
                    MtpType* subj_type = mtpNodeType(chk_node->of.prim.of.item.subj);
                    MtpNode* node_type = (subj_type->kind == mtp_type_tup)
                                             ? subj_type->of.tup.types.at[chk_node->of.prim.of.item.index->of.prim.of.val.of.int_val]
                                             : (subj_type->kind == mtp_type_arr) ? subj_type->of.arr.type : NULL;
                    if (!mtpNodeIsPrimVal(node_type, mtp_type_type))
                        ·fail(str("cannot index into this expression"));
                    MtpType* item_type = &node_type->of.prim.of.val.of.type;
                    if (chk_node->of.prim.of.item.set_to != NULL && !mtpTypesEql(item_type, mtpNodeType(chk_node->of.prim.of.item.set_to)))
                        ·fail(str("type mismatch for setting aggregate member"));
                    chk_node->anns.type = (chk_node->of.prim.of.item.set_to == NULL) ? chk_node->of.prim.of.item.subj->anns.type : node_type;
                } break;
                case mtp_prim_extcall: {
                    MtpPrimExtCall new_call = (MtpPrimExtCall) {
                        .name = node->of.prim.of.ext_call.name,
                        .args_list_val = mtpPreduceNode(ctx, node->of.prim.of.ext_call.args_list_val),
                    };
                    MtpNode* new_ret_type = mtpPreduceNode(ctx, node->anns.type);
                    if (new_ret_type != NULL || new_call.args_list_val != NULL) {
                        if (!mtpNodeIsPrimVal(new_call.args_list_val, mtp_type_tup))
                            ·fail(str("specified illegal MtpPrimExtCall.params_types"));
                        ret_node = mtpUpdNodePrimExtCall(ctx->prog, node, new_call.args_list_val, new_ret_type);
                    }

                    MtpNode* chk_node = (ret_node == NULL) ? node : ret_node;
                    if (!mtpNodeIsPrimVal(chk_node->of.prim.of.ext_call.args_list_val, mtp_type_tup))
                        ·fail(str("specified illegal MtpPrimExtCall.params_types"));
                    if (chk_node->of.prim.of.ext_call.name.at == NULL || chk_node->of.prim.of.ext_call.name.len == 0)
                        ·fail(str("specified illegal MtpPrimExtCall.name"));
                } break;
                case mtp_prim_cast: {
                    MtpPrimCast new_cast = (MtpPrimCast) {
                        .kind = node->of.prim.of.cast.kind,
                        .dst_type = mtpPreduceNode(ctx, node->of.prim.of.cast.dst_type),
                        .subj = mtpPreduceNode(ctx, node->of.prim.of.cast.subj),
                    };
                    if (new_cast.subj != NULL || new_cast.dst_type != NULL)
                        ret_node = mtpUpdNodePrimCast(ctx->prog, node, new_cast);

                    MtpNode* chk_node = (ret_node == NULL) ? node : ret_node;
                    if (!mtpNodeIsPrimVal(chk_node->of.prim.of.cast.dst_type, mtp_type_type))
                        ·fail(str("cast requires type-typed destination"));
                    if (chk_node->of.prim.of.cast.kind == mtp_cast_ints
                        && ((!mtpTypeIsIntCastable(&chk_node->of.prim.of.cast.dst_type->of.prim.of.val.of.type))
                            || (!mtpTypeIsIntCastable(chk_node->of.prim.of.cast.subj->tyype))))
                        ·fail(str("intcast requires int-castable source and destination types"));
                    if (chk_node->of.prim.of.cast.kind == mtp_cast_bits
                        && mtpTypeMinSizeInBits(ctx->prog, &chk_node->of.prim.of.cast.dst_type->of.prim.of.val.of.type)
                               != mtpTypeMinSizeInBits(ctx->prog, chk_node->of.prim.of.cast.subj->tyype))
                        ·fail(str("bitcast requires same bit-width for source and destination type"));
                    chk_node->tyype = &chk_node->of.prim.of.cast.dst_type->of.prim.of.val.of.type;
                } break;
                case mtp_prim_cmp_i: {
                    MtpPrimCmpI new_cmpi = (MtpPrimCmpI) {.kind = node->of.prim.of.cmp_i.kind,
                                                          .lhs = mtpPreduceNode(ctx, node->of.prim.of.cmp_i.lhs),
                                                          .rhs = mtpPreduceNode(ctx, node->of.prim.of.cmp_i.rhs)};
                    if (new_cmpi.lhs != NULL || new_cmpi.rhs != NULL)
                        ret_node = mtpUpdNodePrimCmpI(ctx->prog, node, new_cmpi);

                    MtpNode* chk_node = (ret_node == NULL) ? node : ret_node;
                    Bool ok =
                        mtpTypesEql(chk_node->of.prim.of.cmp_i.lhs->tyype, chk_node->of.prim.of.cmp_i.rhs->tyype)
                        && (chk_node->of.prim.of.cmp_i.lhs->tyype->kind == mtp_type_int
                            || (chk_node->of.prim.of.cmp_i.lhs->tyype->kind == mtp_type_sym
                                && (chk_node->of.prim.of.cmp_i.kind == mtp_cmp_i_eq || chk_node->of.prim.of.cmp_i.kind == mtp_cmp_i_neq)));
                    if (!ok)
                        ·fail(str("invalid operand type(s) for int comparison operation"));
                } break;
                case mtp_prim_bin_i: {
                    MtpPrimBinI new_bini = (MtpPrimBinI) {.kind = node->of.prim.of.bin_i.kind,
                                                          .lhs = mtpPreduceNode(ctx, node->of.prim.of.bin_i.lhs),
                                                          .rhs = mtpPreduceNode(ctx, node->of.prim.of.bin_i.rhs)};
                    if (new_bini.lhs != NULL || new_bini.rhs != NULL)
                        ret_node = mtpUpdNodePrimBinI(ctx->prog, node, new_bini);

                    MtpNode* chk_node = (ret_node == NULL) ? node : ret_node;
                    if (chk_node->of.prim.of.bin_i.lhs->tyype->kind != mtp_type_int
                        || chk_node->of.prim.of.bin_i.rhs->tyype->kind != mtp_type_int
                        || !mtpTypesEql(chk_node->of.prim.of.bin_i.lhs->tyype, chk_node->of.prim.of.bin_i.rhs->tyype))
                        ·fail(str("invalid operand type(s) for int binary operation"));
                    chk_node->tyype = chk_node->of.prim.of.bin_i.lhs->tyype;
                } break;
                case mtp_prim_val:
                    if (node->tyype->kind == mtp_type_arr) {
                        MtpPtrsOfNode new_args = ·sliceOfPtrs(MtpNode, node->of.prim.of.val.of.list_val.len, 0);
                        Bool all_null = true;
                        for (UInt i = 0; i < new_args.len; i += 1) {
                            new_args.at[i] = mtpPreduceNode(ctx, node->of.prim.of.val.of.list_val.at[i]);
                            if (new_args.at[i] != NULL)
                                all_null = false;
                        }
                        // TODO: type-infer / type-check here (arr vs tup etc)
                        if (!all_null)
                            ret_node = mtpUpdNodePrimValList(ctx->prog, node, new_args);
                    }
                default: break;
            }
        } break;

        default: break;
    }
    node->anns.preduced = true;
    if (ret_node != NULL)
        ret_node->anns.preduced = true;

    return ret_node;
}
