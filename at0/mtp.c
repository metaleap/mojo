#pragma once
#include "utils_and_libc_deps.c"

typedef enum MtpTypeKind {
    mt_type_int,
    mt_type_sym,
    mt_type_cont,
} MtpTypeKind;

typedef enum MtpNodeKind {
    mt_node_cont,
    mt_node_branch,
    mt_node_jump,
    mt_node_val_int,
    mt_node_val_sym,
    mt_node_prim,
} MtpNodeKind;

typedef enum MtpNodePrimKind {
    mt_prim_cmp_i,
    mt_prim_bin_i,
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

typedef ·SliceOfPtrs(MtpType) MtpTypeCont;

struct MtpType {
    MtpTypeKind kind;
    union {
        MtpTypeInt num_int;
        MtpTypeSym sym;
        MtpTypeCont cont;
    } of;
};

struct MtpNode;
typedef struct MtpNode MtpNode;

typedef struct MtpNodeCont {
    MtpType* type;
} MtpNodeCont;

typedef struct MtpNodeBranch {
} MtpNodeBranch;

typedef struct MtpNodeJump {
} MtpNodeJump;

typedef struct MtpNodeValInt {
    I64 int_val;
} MtpNodeValInt;

typedef struct MtpNodeValSym {
    UInt sym_val;
} MtpNodeValSym;

typedef struct MtpNodePrim {
} MtpNodePrim;

struct MtpNode {
    MtpNodeKind kind;
    union {
        MtpNodeCont cont;
        MtpNodeBranch branch;
        MtpNodeJump jump;
        MtpNodeValInt val_int;
        MtpNodeValSym val_sym;
        MtpNodePrim prim;
    } of;
};
typedef ·ListOf(MtpNode) MtpNodes;

typedef struct MtpProg {
    MtpNodes defs;
    ·ListOfPtrs(MtpType) types;
} MtpProg;

MtpProg mtpNewProg(UInt const defs_capacity, UInt const types_capacity) {
    MtpProg ret_prog = (MtpProg) {
        .defs = ·listOf(MtpNode, 0, defs_capacity),
        .types = ·listOfPtrs(MtpType, 0, types_capacity),
    };
    return ret_prog;
}

Bool mtpTypeEql(MtpType* const t1, MtpType* const t2) {
    if (t1->kind == t2->kind)
        switch (t1->kind) {
            case mt_type_int:
                return (t1->of.num_int.unsign == t2->of.num_int.unsign) && (t1->of.num_int.bit_width == t2->of.num_int.bit_width);
            case mt_type_sym: {
                if (t1->of.sym.ordered_set.len == t2->of.sym.ordered_set.len) {
                    if (t1->of.sym.ordered_set.at != t2->of.sym.ordered_set.at)
                        for (UInt i = 0; i < t1->of.sym.ordered_set.len; i += 1)
                            if (t1->of.sym.ordered_set.at[i] != t2->of.sym.ordered_set.at[i])
                                return false;
                    return true;
                }
            } break;
            case mt_type_cont: {
                if (t1->of.cont.len == t2->of.cont.len) {
                    if (t1->of.cont.at != t2->of.cont.at)
                        for (UInt i = 0; i < t1->of.cont.len; i += 1)
                            if (!mtpTypeEql(t1->of.cont.at[i], t2->of.cont.at[i]))
                                return false;
                    return true;
                }
            } break;
            default: ·fail(uIntToStr(t1->kind, 1, 10));
        }
    return false;
}

MtpType* mtpType(MtpProg* const prog, MtpTypeKind kind, PtrAny const type_spec) {
    MtpType specd_type = (MtpType) {.kind = kind};
    switch (kind) {
        case mt_type_int: specd_type.of.num_int = *((MtpTypeInt*)type_spec); break;
        case mt_type_sym: specd_type.of.sym = *((MtpTypeSym*)type_spec); break;
        case mt_type_cont: specd_type.of.cont = *((MtpTypeCont*)type_spec); break;
        default: ·fail(uIntToStr(kind, 1, 10));
    }
    // TODO: proper hash-map, or at least hash-cmp instead of deep-cmp
    for (UInt i = 0; i < prog->types.len; i += 1) {
        MtpType* const type = prog->types.at[i];
        if (mtpTypeEql(type, &specd_type))
            return type;
    }
    ·append(prog->types, ·new(MtpType));
    MtpType* type = prog->types.at[prog->types.len];
    *type = specd_type;
    return type;
}
MtpType* mtpTypeInt(MtpProg* const prog, MtpTypeInt type_spec) {
    return mtpType(prog, mt_type_int, &type_spec);
}
MtpType* mtpTypeSym(MtpProg* const prog, MtpTypeSym type_spec) {
    return mtpType(prog, mt_type_sym, &type_spec);
}
MtpType* mtpTypeCont(MtpProg* const prog, MtpTypeCont type_spec) {
    return mtpType(prog, mt_type_cont, &type_spec);
}

MtpNodeCont* mtpNewDef(MtpProg* const prog) {
    ·append(prog->defs, (MtpNode) {.kind = mt_node_cont});
    return ·as(MtpNodeCont, ·last(prog->defs));
}
