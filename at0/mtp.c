#pragma once
#include "utils_and_libc_deps.c"

typedef enum MtpTypeKind {
    mtp_type_int,
    mtp_type_sym,
    mtp_type_cont,
} MtpTypeKind;

typedef enum MtpNodeKind {
    mtp_node_cont,
    mtp_node_param,
    mtp_node_branch,
    mtp_node_jump,
    mtp_node_val_int,
    mtp_node_val_sym,
    mtp_node_prim,
} MtpNodeKind;

typedef enum MtpNodePrimKind {
    mtp_prim_cmp_i,
    mtp_prim_bin_i,
} MtpNodePrimKind;

struct MtpType;
typedef struct MtpType MtpType;

typedef struct MtpTypeInt {
    U32 bit_width;
    Bool unsign;
} MtpTypeInt;

typedef struct MtpTypeSym {
    UInts ordered_set;
} MtpTypeSym;

typedef struct MtpTypeCont {
    ·SliceOfPtrs(MtpType) params;
} MtpTypeCont;

struct MtpType {
    union {
        MtpTypeInt num_int;
        MtpTypeSym sym;
        MtpTypeCont cont;
    } of;
    MtpTypeKind kind;
};

struct MtpNode;
typedef struct MtpNode MtpNode;
typedef ·SliceOf(MtpNode) MtpNodes;

typedef struct MtpNodeParam {
    MtpType* param_type;
} MtpNodeParam;

typedef struct MtpNodeCont {
    MtpType* cont_type;
    MtpNodes params;
} MtpNodeCont;

typedef struct MtpNodeBranch {
    MtpNode* cond;
    MtpNode* if1;
    MtpNode* if0;
} MtpNodeBranch;

typedef struct MtpNodeJump {
    MtpNode* dst_cont;
    ·SliceOfPtrs(MtpNode) args;
} MtpNodeJump;

typedef struct MtpNodeValInt {
    I64 int_val;
} MtpNodeValInt;

typedef struct MtpNodeValSym {
    UInt sym_val;
} MtpNodeValSym;

typedef struct MtpNodePrimCmpI {
    MtpNode* lhs;
    MtpNode* rhs;
    enum {
        mtp_cmp_i_eq,
        mtp_cmp_i_neq,
    } kind;
} MtpNodePrimCmpI;

typedef struct MtpNodePrimBinI {
    MtpNode* lhs;
    MtpNode* rhs;
    enum {
        mtp_bin_i_add,
        mtp_bin_i_mul,
    } kind;
} MtpNodePrimBinI;

typedef struct MtpNodePrim {
    union {
        MtpNodePrimCmpI cmp;
        MtpNodePrimBinI bin;
    } of;
    MtpNodePrimKind kind;
} MtpNodePrim;

struct MtpNode {
    union {
        MtpNodeCont cont;
        MtpNodeParam param;
        MtpNodeBranch branch;
        MtpNodeJump jump;
        MtpNodeValInt val_int;
        MtpNodeValSym val_sym;
        MtpNodePrim prim;
    } of;
    MtpNodeKind kind;
};

typedef struct MtpProg {
    ·ListOfPtrs(MtpNode) defs;
    struct {
        ·ListOfPtrs(MtpType) types;
        ·ListOfPtrs(MtpNode) int_vals;
        ·ListOfPtrs(MtpNode) sym_vals;
        ·ListOfPtrs(MtpNode) prims;
    } all;
} MtpProg;

Bool mtpTypesEql(MtpType* const t1, MtpType* const t2) {
    if (t1 == t2)
        return true;
    if (t1->kind == t2->kind)
        switch (t1->kind) {
            case mtp_type_int:
                return (t1->of.num_int.unsign == t2->of.num_int.unsign) && (t1->of.num_int.bit_width == t2->of.num_int.bit_width);
            case mtp_type_sym: {
                if (t1->of.sym.ordered_set.len == t2->of.sym.ordered_set.len) {
                    if (t1->of.sym.ordered_set.at != t2->of.sym.ordered_set.at)
                        for (UInt i = 0; i < t1->of.sym.ordered_set.len; i += 1)
                            if (t1->of.sym.ordered_set.at[i] != t2->of.sym.ordered_set.at[i])
                                return false;
                    return true;
                }
            } break;
            case mtp_type_cont: {
                if (t1->of.cont.params.len == t2->of.cont.params.len) {
                    if (t1->of.cont.params.at != t2->of.cont.params.at)
                        for (UInt i = 0; i < t1->of.cont.params.len; i += 1)
                            if (!mtpTypesEql(t1->of.cont.params.at[i], t2->of.cont.params.at[i]))
                                return false;
                    return true;
                }
            } break;
            default: ·fail(uIntToStr(t1->kind, 1, 10));
        }
    return false;
}

Bool mtpNodesEql(MtpNode* const n1, MtpNode* const n2) {
    if (n1 == n2)
        return true;
    if (n1->kind == n2->kind)
        switch (n1->kind) {
            case mtp_node_val_int: return n1->of.val_int.int_val == n2->of.val_int.int_val;
            case mtp_node_val_sym: return n1->of.val_sym.sym_val == n2->of.val_sym.sym_val;
            case mtp_node_prim: {
                if (n1->of.prim.kind == n2->of.prim.kind)
                    switch (n1->of.prim.kind) {
                        case mtp_prim_bin_i:
                            return n1->of.prim.of.bin.kind == n2->of.prim.of.bin.kind
                                   && mtpNodesEql(n1->of.prim.of.bin.lhs, n2->of.prim.of.bin.lhs)
                                   && mtpNodesEql(n1->of.prim.of.bin.rhs, n2->of.prim.of.bin.rhs);
                        case mtp_prim_cmp_i:
                            return n1->of.prim.of.cmp.kind == n2->of.prim.of.cmp.kind
                                   && mtpNodesEql(n1->of.prim.of.cmp.lhs, n2->of.prim.of.cmp.lhs)
                                   && mtpNodesEql(n1->of.prim.of.cmp.rhs, n2->of.prim.of.cmp.rhs);
                        default: ·fail(uIntToStr(n1->of.prim.kind, 1, 10));
                    }
            } break;
            default: ·fail(uIntToStr(n1->kind, 1, 10));
        }
    return false;
}

MtpType* mtpType(MtpProg* const prog, MtpTypeKind const kind, PtrAny const type_spec) {
    MtpType specd_type = (MtpType) {.kind = kind};
    switch (kind) {
        case mtp_type_int: specd_type.of.num_int = *((MtpTypeInt*)type_spec); break;
        case mtp_type_sym: specd_type.of.sym = *((MtpTypeSym*)type_spec); break;
        case mtp_type_cont: specd_type.of.cont = *((MtpTypeCont*)type_spec); break;
        default: ·fail(uIntToStr(kind, 1, 10));
    }
    // TODO: proper hash-map, or at least hash-cmp instead of deep-cmp
    for (UInt i = 0; i < prog->all.types.len; i += 1) {
        MtpType* const type = prog->all.types.at[i];
        if (mtpTypesEql(type, &specd_type))
            return type;
    }
    ·append(prog->all.types, ·new(MtpType));
    MtpType* type = prog->all.types.at[prog->all.types.len];
    *type = specd_type;
    return type;
}
MtpType* mtpTypeInt(MtpProg* const prog, MtpTypeInt type_spec) {
    return mtpType(prog, mtp_type_int, &type_spec);
}
MtpType* mtpTypeSym(MtpProg* const prog, MtpTypeSym type_spec) {
    return mtpType(prog, mtp_type_sym, &type_spec);
}
MtpType* mtpTypeCont(MtpProg* const prog, MtpTypeCont type_spec) {
    return mtpType(prog, mtp_type_cont, &type_spec);
}

MtpNode* mtpValInt(MtpProg* const prog, I64 const int_val) {
    for (UInt i = 0; i < prog->all.int_vals.len; i += 1) {
        MtpNode* node = prog->all.int_vals.at[i];
        if (node->of.val_int.int_val == int_val)
            return node;
    }
    ·append(prog->all.int_vals, ·new(MtpNode));
    MtpNode* ret_node = prog->all.int_vals.at[prog->all.int_vals.len - 1];
    ret_node->kind = mtp_node_val_int;
    ret_node->of.val_int = (MtpNodeValInt) {.int_val = int_val};
    return ret_node;
}

MtpNode* mtpValSym(MtpProg* const prog, UInt const sym_val) {
    for (UInt i = 0; i < prog->all.sym_vals.len; i += 1) {
        MtpNode* node = prog->all.sym_vals.at[i];
        if (node->of.val_sym.sym_val == sym_val)
            return node;
    }
    ·append(prog->all.sym_vals, ·new(MtpNode));
    MtpNode* ret_node = prog->all.sym_vals.at[prog->all.sym_vals.len - 1];
    ret_node->kind = mtp_node_val_sym;
    ret_node->of.val_sym = (MtpNodeValSym) {.sym_val = sym_val};
    return ret_node;
}

MtpNode* mtpPrim(MtpProg* const prog, MtpNodePrim spec) {
    MtpNode spec_node = (MtpNode) {.kind = mtp_node_prim, .of = {.prim = spec}};
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
MtpNode* mtpPrimCmpI(MtpProg* const prog, MtpNodePrimCmpI spec) {
    return mtpPrim(prog, (MtpNodePrim) {.kind = mtp_prim_cmp_i, .of = {.cmp = spec}});
}
MtpNode* mtpPrimBinI(MtpProg* const prog, MtpNodePrimBinI spec) {
    return mtpPrim(prog, (MtpNodePrim) {.kind = mtp_prim_bin_i, .of = {.bin = spec}});
}

MtpNode* mtpDefParam(MtpNode* const def_node, UInt const param_index) {
    return &def_node->of.cont.params.at[param_index];
}

MtpNode* mtpNewDef(MtpProg* const prog, ·SliceOfPtrs(MtpType) params) {
    ·append(prog->defs, ·new(MtpNode));
    MtpNode* ret_node = ·as(MtpNode, ·last(prog->defs));
    ret_node->kind = mtp_node_cont;
    ret_node->of.cont = (MtpNodeCont) {
        .cont_type = mtpTypeCont(prog, (MtpTypeCont) {.params = {.at = params.at, .len = params.len}}),
        .params = ·sliceOf(MtpNode, params.len, params.len),
    };
    for (UInt i = 0; i < params.len; i += 1)
        ret_node->of.cont.params.at[i] = (MtpNode) {
            .kind = mtp_node_param,
            .of = {.param = (MtpNodeParam) {.param_type = params.at[i]}},
        };
    return ret_node;
}

MtpProg mtpNewProg(UInt const defs_capacity, UInt const types_capacity, UInt const int_vals_capacity, UInt const sym_vals_capacity,
                   UInt const prims_capacity) {
    MtpProg ret_prog = (MtpProg) {.defs = ·listOfPtrs(MtpNode, 0, defs_capacity),
                                  .all = {
                                      .types = ·listOfPtrs(MtpType, 0, types_capacity),
                                      .int_vals = ·listOfPtrs(MtpNode, 0, int_vals_capacity),
                                      .sym_vals = ·listOfPtrs(MtpNode, 0, sym_vals_capacity),
                                      .prims = ·listOfPtrs(MtpNode, 0, prims_capacity),
                                  }};
    return ret_prog;
}
