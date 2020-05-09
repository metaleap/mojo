#pragma once
#include "utils_and_libc_deps.c"


typedef enum MtpKindOfType {
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


struct MtpType;
typedef struct MtpType MtpType;
typedef ·SliceOfPtrs(MtpType) MtpPtrsOfType;

typedef struct MtpTypeInt {
    U16 bit_width;
    Bool unsign;
} MtpTypeInt;

typedef struct MtpTypeBottom {
} MtpTypeBottom;

typedef struct MtpTypeSym {
} MtpTypeSym;

typedef struct MtpTypeTup {
    MtpPtrsOfType types;
} MtpTypeTup;

typedef struct MtpTypePtr {
    MtpType* type;
} MtpTypePtr;

typedef struct MtpTypeArr {
    MtpType* type;
    UInt length;
} MtpTypeArr;

struct MtpType {
    union {
        MtpTypeBottom bottom;
        MtpTypeSym sym;
        MtpTypeInt num_int;
        MtpTypeTup tup;
        MtpTypePtr ptr;
        MtpTypeArr arr;
    } of;
    MtpKindOfType kind;
};


struct MtpNode;
typedef struct MtpNode MtpNode;
typedef ·SliceOfPtrs(MtpNode) MtpPtrsOfNode;

typedef struct MtpNodeParam {
    UInt param_idx;
} MtpNodeParam;

typedef struct MtpNodeLam {
    MtpPtrsOfNode params;
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

typedef union MtpPrimVal {
    I64 int_val;
    U32 sym_val;
    MtpType* type_val;
    struct {
    } bottom;
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
    MtpType* dst_type;
    MtpKindOfCast kind;
} MtpPrimCast;

typedef struct MtpPrimItem {
    MtpNode* subj;
    MtpNode* index;
    MtpNode* set_to; // if NULL, it's a getter, else a setter
} MtpPrimItem;

typedef struct MtpPrimExtCall {
    MtpType* params_types; // must be tup kind (MtpTypeTup)
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
    MtpType* type;
    union {
        MtpNodeLam lam;
        MtpNodeParam param;
        MtpNodeChoice choice;
        MtpNodeJump jump;
        MtpNodePrim prim;
    } of;
    struct {
        Bool preduced;
    } anns;
    MtpKindOfNode kind;
};


typedef struct MtpProg {
    struct {
        ·ListOfPtrs(MtpType) types;
        ·ListOfPtrs(MtpNode) lams;
        ·ListOfPtrs(MtpNode) prims;
        ·ListOfPtrs(MtpNode) choices;
        ·ListOfPtrs(MtpNode) jumps;
        ·ListOfPtrs(MtpNode) syms;
    } all;
    struct {
        U16 ptrs;
        U16 syms;
    } bit_widths;
} MtpProg;




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

Bool mtpNodeIsPrimVal(MtpNode const* const node) {
    return node->kind == mtp_node_prim && node->of.prim.kind == mtp_prim_val;
}

Bool mtpNodeIsPrimValInt(MtpNode* const node) {
    return node->type->kind == mtp_type_int && mtpNodeIsPrimVal(node);
}

Bool mtpNodeIsPrimValSym(MtpNode* const node) {
    return node->type->kind == mtp_type_sym && mtpNodeIsPrimVal(node);
}

Bool mtpNodeIsBasicBlockishLam(MtpNode* const node) {
    return node->type->kind == mtp_type_lam && (node->type->of.tup.types.len == 0);
}




Bool mtpTypesEql(MtpType const* const t1, MtpType const* const t2) {
    if (t1 == t2)
        return true;
    if (t1 != NULL & t2 != NULL && t1->kind == t2->kind)
        switch (t1->kind) {
            case mtp_type_sym:
            case mtp_type_bottom: return true;
            case mtp_type_ptr: return mtpTypesEql(t1->of.ptr.type, t2->of.ptr.type);
            case mtp_type_arr: return (t1->of.arr.length == t2->of.arr.length) && mtpTypesEql(t1->of.arr.type, t2->of.arr.type);
            case mtp_type_int:
                return (t1->of.num_int.unsign == t2->of.num_int.unsign) && (t1->of.num_int.bit_width == t2->of.num_int.bit_width);
            case mtp_type_tup:
            case mtp_type_lam: {
                if (t1->of.tup.types.len == t2->of.tup.types.len) {
                    if (t1->of.tup.types.at != t2->of.tup.types.at)
                        for (UInt i = 0; i < t1->of.tup.types.len; i += 1)
                            if (!mtpTypesEql(t1->of.tup.types.at[i], t2->of.tup.types.at[i]))
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
        case mtp_type_arr: return type->of.arr.length * mtpTypeMinSizeInBits(prog, type->of.arr.type);
        case mtp_type_tup: {
            UInt size = 0;
            for (UInt i = 0; i < type->of.tup.types.len; i += 1)
                size += mtpTypeMinSizeInBits(prog, type->of.tup.types.at[i]);
            return size;
        } break;
        default: ·fail(uIntToStr(type->kind, 1, 10)); return 0;
    }
}

Bool mtpTypeIsIntCastable(MtpType* type) {
    return type->kind == mtp_type_int || type->kind == mtp_type_ptr || type->kind == mtp_type_sym;
}

MtpType* mtpType(MtpProg* const prog, MtpKindOfType const kind, PtrAny const type_spec) {
    MtpType specd_type = (MtpType) {.kind = kind};
    switch (kind) {
        case mtp_type_sym:
        case mtp_type_bottom: break;
        case mtp_type_ptr: specd_type.of.ptr = *((MtpTypePtr*)type_spec); break;
        case mtp_type_arr: specd_type.of.arr = *((MtpTypeArr*)type_spec); break;
        case mtp_type_int: specd_type.of.num_int = *((MtpTypeInt*)type_spec); break;
        case mtp_type_lam:
        case mtp_type_tup: specd_type.of.tup = *((MtpTypeTup*)type_spec); break;
        default: ·fail(uIntToStr(kind, 1, 10));
    }
    // TODO: proper hash-map, or at least hash-cmp instead of deep-cmp
    for (UInt i = 0; i < prog->all.types.len; i += 1) {
        MtpType* const type = prog->all.types.at[i];
        if (mtpTypesEql(type, &specd_type))
            return type;
    }
    ·append(prog->all.types, ·new(MtpType));
    MtpType* ret_type = prog->all.types.at[prog->all.types.len];
    *ret_type = specd_type;
    return ret_type;
}
MtpType* mtpTypeInt(MtpProg* const prog, MtpTypeInt type_spec) {
    return mtpType(prog, mtp_type_int, &type_spec);
}
MtpType* mtpTypeIntStatic(MtpProg* const prog) {
    return prog->all.types.at[3];
}
MtpType* mtpTypeSym(MtpProg* const prog) {
    return prog->all.types.at[0];
}
MtpType* mtpTypeBottom(MtpProg* const prog) {
    return prog->all.types.at[1];
}
MtpType* mtpTypeBool(MtpProg* const prog) {
    return prog->all.types.at[2];
}
MtpType* mtpTypeLabel(MtpProg* const prog) {
    return prog->all.types.at[5];
}
MtpType* mtpTypePtr(MtpProg* const prog, MtpTypePtr type_spec) {
    return mtpType(prog, mtp_type_ptr, &type_spec);
}
MtpType* mtpTypeArr(MtpProg* const prog, MtpTypeArr type_spec) {
    return mtpType(prog, mtp_type_arr, &type_spec);
}
MtpType* mtpTypeTup(MtpProg* const prog, MtpTypeTup type_spec) {
    return mtpType(prog, mtp_type_tup, &type_spec);
}
MtpType* mtpTypeLam(MtpProg* const prog, MtpTypeTup type_spec) {
    return mtpType(prog, mtp_type_lam, &type_spec);
}



Bool mtpNodesEql(MtpNode const* const n1, MtpNode const* const n2) {
    // outside ctors, always true:
    if (n1 == n2)
        return true;
    // only when called from ctors:
    if (n1->kind == n2->kind && mtpTypesEql(n1->type, n2->type))
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
                        case mtp_prim_val:
                            return (n1->type->kind == mtp_type_bottom)
                                   || (n1->type->kind == mtp_type_int && n1->of.prim.of.val.int_val == n2->of.prim.of.val.int_val)
                                   || (n1->type->kind == mtp_type_sym && n1->of.prim.of.val.sym_val == n2->of.prim.of.val.sym_val);
                        case mtp_prim_item:
                            return n1->of.prim.of.item.index == n2->of.prim.of.item.index
                                   && mtpNodesEql(n1->of.prim.of.item.set_to, n2->of.prim.of.item.set_to)
                                   && mtpNodesEql(n1->of.prim.of.item.subj, n2->of.prim.of.item.subj);
                        case mtp_prim_cast:
                            return n1->of.prim.of.cast.kind == n2->of.prim.of.cast.kind
                                   && mtpTypesEql(n1->of.prim.of.cast.dst_type, n2->of.prim.of.cast.dst_type)
                                   && mtpNodesEql(n1->of.prim.of.cast.subj, n2->of.prim.of.cast.subj);
                        case mtp_prim_bin_i:
                            return n1->of.prim.of.bin_i.kind == n2->of.prim.of.bin_i.kind
                                   && (mtpNodesEql(n1->of.prim.of.bin_i.lhs, n2->of.prim.of.bin_i.lhs)
                                       && mtpNodesEql(n1->of.prim.of.bin_i.rhs, n2->of.prim.of.bin_i.rhs));
                        case mtp_prim_cmp_i: // NOTE same as above; 3rd time around, extract!
                            return n1->of.prim.of.cmp_i.kind == n2->of.prim.of.cmp_i.kind
                                   && (mtpNodesEql(n1->of.prim.of.cmp_i.lhs, n2->of.prim.of.cmp_i.lhs)
                                       && mtpNodesEql(n1->of.prim.of.cmp_i.rhs, n2->of.prim.of.cmp_i.rhs));
                        case mtp_prim_extcall:
                            return mtpTypesEql(n1->of.prim.of.ext_call.params_types, n2->of.prim.of.ext_call.params_types)
                                   && strEql(n1->of.prim.of.ext_call.name, n2->of.prim.of.ext_call.name);
                        default: ·fail(uIntToStr(n1->of.prim.kind, 1, 10));
                    }
            } break;
            default: ·fail(uIntToStr(n1->kind, 1, 10));
        }
    return false;
}

MtpNode* mtpNodeChoice(MtpProg* const prog, MtpNodeChoice const spec) {
    MtpNode spec_node = (MtpNode) {.kind = mtp_node_choice, .type = NULL, .of = {.choice = spec}, .anns = {.preduced = false}};
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
        .type = mtpTypeBottom(prog),
        .anns = {.preduced = false},
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

MtpNode* mtpNodePrim(MtpProg* const prog, MtpNodePrim const spec, MtpType* const type) {
    MtpNode const spec_node = (MtpNode) {.kind = mtp_node_prim, .type = type, .of = {.prim = spec}, .anns = {.preduced = false}};
    if (spec.kind == mtp_prim_val && type != NULL && type->kind == mtp_type_sym)
        return prog->all.syms.at[spec.of.val.sym_val];
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
MtpNode* mtpNodePrimExtCall(MtpProg* const prog, MtpPrimExtCall const spec, MtpType* const ret_type) {
    return mtpNodePrim(prog, (MtpNodePrim) {.kind = mtp_prim_extcall, .of = {.ext_call = spec}}, ret_type);
}
MtpNode* mtpNodePrimCast(MtpProg* const prog, MtpPrimCast spec) {
    return mtpNodePrim(prog, (MtpNodePrim) {.kind = mtp_prim_cast, .of = {.cast = spec}}, spec.dst_type);
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
MtpNode* mtpNodePrimValInt(MtpProg* const prog, I64 const int_val) {
    return mtpNodePrim(prog, (MtpNodePrim) {.kind = mtp_prim_val, .of = {.val = {.int_val = int_val}}}, mtpTypeIntStatic(prog));
}
MtpNode* mtpNodePrimValSym(MtpProg* const prog, U32 const sym_val) {
    return prog->all.syms.at[sym_val];
}
MtpNode* mtpNodePrimValBottom(MtpProg* const prog) {
    return prog->all.prims.at[2];
}
MtpNode* mtpNodePrimValBoolFalse(MtpProg* const prog) {
    return prog->all.prims.at[0];
}
MtpNode* mtpNodePrimValBoolTrue(MtpProg* const prog) {
    return prog->all.prims.at[1];
}

MtpNode* mtpNodeLam(MtpProg* const prog, MtpPtrsOfType const params) {
    ·append(prog->all.lams, ·new(MtpNode));
    MtpNode* ret_node = ·as(MtpNode, ·last(prog->all.lams));
    ret_node->kind = mtp_node_lam;
    ret_node->type = mtpTypeLam(prog, (MtpTypeTup) {.types = {.at = params.at, .len = params.len}});
    ret_node->of.lam = (MtpNodeLam) {.body = NULL, .params = ·sliceOfPtrs(MtpNode, params.len, params.len)};
    for (UInt i = 0; i < params.len; i += 1) {
        ret_node->of.lam.params.at[i] = ·new(MtpNode);
        *ret_node->of.lam.params.at[i] = (MtpNode) {
            .kind = mtp_node_param,
            .type = params.at[i],
            .anns = {.preduced = false},
            .of = {.param = (MtpNodeParam) {.param_idx = i}},
        };
    }
    return ret_node;
}

MtpProg mtpProg(UInt bit_width_ptrs, UInt const lams_capacity, UInt const types_capacity, UInt const int_vals_capacity,
                UInt const prims_capacity, UInt const choices_capacity, UInt const jumps_capacity, U32 const sym_vals_total_count) {
    MtpProg ret_prog = (MtpProg) {.all =
                                      {
                                          .types = ·listOfPtrs(MtpType, 6, types_capacity),
                                          .syms = ·listOfPtrs(MtpNode, sym_vals_total_count, sym_vals_total_count),
                                          .prims = ·listOfPtrs(MtpNode, 0, prims_capacity),
                                          .lams = ·listOfPtrs(MtpNode, 0, lams_capacity),
                                          .choices = ·listOfPtrs(MtpNode, 0, choices_capacity),
                                          .jumps = ·listOfPtrs(MtpNode, 0, jumps_capacity),
                                      },
                                  .bit_widths = {
                                      .ptrs = bit_width_ptrs,
                                      .syms = uIntMinSize(sym_vals_total_count - 1, 1),
                                  }};

    ret_prog.all.types.at[0] = ·new(MtpType); // returned by `mtpTypeSym(MtpProg*)`:
    ret_prog.all.types.at[0]->kind = mtp_type_sym;
    ret_prog.all.types.at[1] = ·new(MtpType); // returned by `mtpTypeBottom(MtpProg*)`:
    ret_prog.all.types.at[1]->kind = mtp_type_bottom;
    ret_prog.all.types.at[2] = ·new(MtpType); // returned by `mtpTypeBool(MtpProg*)`:
    *ret_prog.all.types.at[2] = (MtpType) {.kind = mtp_type_int, .of = {.num_int = {.bit_width = 1, .unsign = true}}};
    ret_prog.all.types.at[3] = ·new(MtpType); // statically known int values:
    *ret_prog.all.types.at[3] = (MtpType) {.kind = mtp_type_int, .of = {.num_int = {.bit_width = 0, .unsign = false}}};
    ret_prog.all.types.at[4] = ·new(MtpType); // raw / untyped / any pointer:
    *ret_prog.all.types.at[4] = (MtpType) {.kind = mtp_type_ptr, .of = {.ptr = {.type = NULL}}};
    ret_prog.all.types.at[5] = ·new(MtpType); // empty tuple = fn(void):
    *ret_prog.all.types.at[5] = (MtpType) {.kind = mtp_type_lam, .of = {.tup = {.types = ·sliceOfPtrs(MtpType, 0, 0)}}};

    mtpNodePrimValInt(&ret_prog, 0)->type = mtpTypeBool(&ret_prog);
    mtpNodePrimValInt(&ret_prog, 1)->type = mtpTypeBool(&ret_prog);
    mtpNodePrimValInt(&ret_prog, -1)->type = mtpTypeBottom(&ret_prog);
    mtpNodePrimValInt(&ret_prog, 0);
    mtpNodePrimValInt(&ret_prog, 1);

    for (UInt i = 0; i < sym_vals_total_count; i += 1) {
        ret_prog.all.syms.at[i] = ·new(MtpNode);
        *ret_prog.all.syms.at[i] = (MtpNode) {.kind = mtp_node_prim,
                                              .type = mtpTypeSym(&ret_prog),
                                              .anns = {.preduced = false},
                                              .of = {.prim = (MtpNodePrim) {
                                                         .kind = mtp_prim_val,
                                                         .of = {.val = (MtpPrimVal) {.sym_val = i}},
                                                     }}};
    }
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

MtpNode* mtpUpdNodeJump(MtpProg* const prog, MtpNode* const node, MtpNodeJump upd) {
    if (upd.callee == NULL)
        upd.callee = node->of.jump.callee;
    if (upd.args.at == NULL)
        upd.args = node->of.jump.args;
    else if (upd.args.at != node->of.jump.args.at) {
        Bool all_null = true;
        for (UInt i = 0; i < upd.args.len; i += 1)
            if (upd.args.at[i] == NULL)
                upd.args.at[i] = node->of.jump.args.at[i];
            else
                all_null = false;
        if (all_null) {
            UInt const len = upd.args.len;
            upd.args = node->of.jump.args;
            upd.args.len = len;
        }
    }
    if (upd.callee == node->of.jump.callee && upd.args.at == node->of.jump.args.at && upd.args.len == node->of.jump.args.len)
        return node;
    return mtpNodeJump(prog, upd);
}

MtpNode* mtpUpdNodePrimExtCall(MtpProg* const prog, MtpNode* const node, MtpPrimExtCall upd) {
    MtpTypeTup* tupd = &upd.params_types->of.tup;
    MtpTypeTup* tsrc = &node->of.prim.of.ext_call.params_types->of.tup;
    if (tupd->types.at == NULL)
        tupd->types = tsrc->types;
    else if (tupd->types.at != tsrc->types.at) {
        Bool all_null = true;
        for (UInt i = 0; i < tupd->types.len; i += 1)
            if (tupd->types.at[i] == NULL)
                tupd->types.at[i] = tsrc->types.at[i];
            else
                all_null = false;
        if (all_null) {
            UInt len = tupd->types.len;
            tupd->types = tsrc->types;
            tupd->types.len = len;
        }
    }
    if (tupd->types.at == tsrc->types.at && tupd->types.len == tsrc->types.len)
        return node;
    return mtpNodePrimExtCall(prog, upd, node->type);
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
            MtpNodeChoice new_choice = (MtpNodeChoice) {
                .cond = mtpPreduceNode(ctx, node->of.choice.cond),
                .if0 = mtpPreduceNode(ctx, node->of.choice.if0),
                .if1 = mtpPreduceNode(ctx, node->of.choice.if1),
            };
            if (new_choice.cond != NULL || new_choice.if0 != NULL || new_choice.if1 != 0)
                ret_node = mtpUpdNodeChoice(ctx->prog, node, new_choice);

            MtpNode* val_node = (ret_node == NULL) ? node : ret_node;
            if (!mtpTypesEql(val_node->of.choice.cond->type, mtpTypeBool(ctx->prog)))
                ·fail(str("choice condition isn't boolish"));
            if (!(mtpNodeIsBasicBlockishLam(val_node->of.choice.if0) && mtpNodeIsBasicBlockishLam(val_node->of.choice.if1)))
                ·fail(str("choice results must both preduce to basic blocks"));
            val_node->type = val_node->of.choice.if0->type;
            if (mtpNodesEql(val_node->of.choice.cond, mtpNodePrimValBoolTrue(ctx->prog)))
                ret_node = val_node->of.choice.if1;
            else if (mtpNodesEql(val_node->of.choice.cond, mtpNodePrimValBoolFalse(ctx->prog)))
                ret_node = val_node->of.choice.if0;
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

            MtpNode* val_node = (ret_node == NULL) ? node : ret_node;
            MtpType* const fn_type = val_node->of.jump.callee->type;
            if (fn_type->kind != mtp_type_lam
                || !(val_node->of.jump.callee->kind == mtp_node_lam || val_node->of.jump.callee->kind == mtp_node_param))
                ·fail(str("not callable"));
            if (val_node->of.jump.callee->of.lam.params.len != val_node->of.jump.args.len)
                ·fail(str4(str("callee expected "), uIntToStr(val_node->of.jump.callee->of.lam.params.len, 1, 10),
                           str(" arg(s) but caller gave "), uIntToStr(val_node->of.jump.args.len, 1, 10)));
            for (UInt i = 0; i < val_node->of.jump.args.len; i += 1) {
                MtpNode* arg = val_node->of.jump.args.at[i];
                MtpNode* param = val_node->of.jump.callee->of.lam.params.at[i];
                if (!mtpTypesEql(arg->type, param->type))
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

                    MtpNode* val_node = (ret_node == NULL) ? node : ret_node;
                    MtpType* item_type =
                        (val_node->of.prim.of.item.subj->type->kind == mtp_type_tup)
                            ? (val_node->of.prim.of.item.subj->type->of.tup.types.at[val_node->of.prim.of.item.index->of.prim.of.val.int_val])
                            : (val_node->of.prim.of.item.subj->type->kind == mtp_type_arr)
                                  ? (val_node->of.prim.of.item.subj->type->of.arr.type)
                                  : NULL;
                    if (val_node->of.prim.of.item.index->type->kind != mtp_type_int || !mtpNodeIsPrimVal(val_node->of.prim.of.item.index))
                        ·fail(str("expected statically-known index"));
                    if (item_type == NULL)
                        ·fail(str("cannot index into this node"));
                    if (val_node->of.prim.of.item.set_to != NULL && !mtpTypesEql(item_type, val_node->of.prim.of.item.set_to->type))
                        ·fail(str("type mismatch for setting aggregate member"));
                    val_node->type = (val_node->of.prim.of.item.set_to == NULL) ? item_type : val_node->of.prim.of.item.subj->type;
                } break;
                case mtp_prim_extcall: {
                    // TODO

                    MtpNode* val_node = (ret_node == NULL) ? node : ret_node;
                    if (val_node->of.prim.of.ext_call.params_types->kind != mtp_type_tup)
                        ·fail(str("specified illegal MtpPrimExtCall.params_types"));
                    if (val_node->of.prim.of.ext_call.name.at == NULL || val_node->of.prim.of.ext_call.name.len == 0)
                        ·fail(str("specified illegal MtpPrimExtCall.name"));
                } break;
                case mtp_prim_cast: {
                    MtpPrimCast new_cast = (MtpPrimCast) {
                        .kind = node->of.prim.of.cast.kind,
                        .dst_type = node->of.prim.of.cast.dst_type, // TODO
                        .subj = mtpPreduceNode(ctx, node->of.prim.of.cast.subj),
                    };
                    if (new_cast.subj != NULL /*|| new_cast.dst_type != NULL*/)
                        ret_node = mtpUpdNodePrimCast(ctx->prog, node, new_cast);

                    MtpNode* val_node = (ret_node == NULL) ? node : ret_node;
                    if (val_node->of.prim.of.cast.kind == mtp_cast_ints
                        && ((!mtpTypeIsIntCastable(val_node->of.prim.of.cast.dst_type))
                            || (!mtpTypeIsIntCastable(val_node->of.prim.of.cast.subj->type))))
                        ·fail(str("intcast requires int-castable source and destination types"));
                    if (val_node->of.prim.of.cast.kind == mtp_cast_bits
                        && mtpTypeMinSizeInBits(ctx->prog, val_node->of.prim.of.cast.dst_type)
                               != mtpTypeMinSizeInBits(ctx->prog, val_node->of.prim.of.cast.subj->type))
                        ·fail(str("bitcast requires same bit-width for source and destination type"));
                    val_node->type = val_node->of.prim.of.cast.dst_type;
                } break;
                case mtp_prim_cmp_i: {
                    MtpPrimCmpI new_cmpi = (MtpPrimCmpI) {.kind = node->of.prim.of.cmp_i.kind,
                                                          .lhs = mtpPreduceNode(ctx, node->of.prim.of.cmp_i.lhs),
                                                          .rhs = mtpPreduceNode(ctx, node->of.prim.of.cmp_i.rhs)};
                    if (new_cmpi.lhs != NULL || new_cmpi.rhs != NULL)
                        ret_node = mtpUpdNodePrimCmpI(ctx->prog, node, new_cmpi);

                    MtpNode* val_node = (ret_node == NULL) ? node : ret_node;
                    Bool ok =
                        mtpTypesEql(val_node->of.prim.of.cmp_i.lhs->type, val_node->of.prim.of.cmp_i.rhs->type)
                        && (val_node->of.prim.of.cmp_i.lhs->type->kind == mtp_type_int
                            || (val_node->of.prim.of.cmp_i.lhs->type->kind == mtp_type_sym
                                && (val_node->of.prim.of.cmp_i.kind == mtp_cmp_i_eq || val_node->of.prim.of.cmp_i.kind == mtp_cmp_i_neq)));
                    if (!ok)
                        ·fail(str("invalid operand type(s) for int comparison operation"));
                } break;
                case mtp_prim_bin_i: {
                    MtpPrimBinI new_bini = (MtpPrimBinI) {.kind = node->of.prim.of.bin_i.kind,
                                                          .lhs = mtpPreduceNode(ctx, node->of.prim.of.bin_i.lhs),
                                                          .rhs = mtpPreduceNode(ctx, node->of.prim.of.bin_i.rhs)};
                    if (new_bini.lhs != NULL || new_bini.rhs != NULL)
                        ret_node = mtpUpdNodePrimBinI(ctx->prog, node, new_bini);

                    MtpNode* val_node = (ret_node == NULL) ? node : ret_node;
                    if (val_node->of.prim.of.bin_i.lhs->type->kind != mtp_type_int
                        || val_node->of.prim.of.bin_i.rhs->type->kind != mtp_type_int
                        || !mtpTypesEql(val_node->of.prim.of.bin_i.lhs->type, val_node->of.prim.of.bin_i.rhs->type))
                        ·fail(str("invalid operand type(s) for int binary operation"));
                    val_node->type = val_node->of.prim.of.bin_i.lhs->type;
                } break;
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

void mtpPreduceProg(MtpProg* const prog) {
    MtpCtxPreduce ctx = (MtpCtxPreduce) {.prog = prog};
    for (UInt i = 0; i < prog->all.lams.len; i += 1)
        mtpPreduceNode(&ctx, prog->all.lams.at[i]);
}
