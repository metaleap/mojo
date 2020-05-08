#pragma once
#include "utils_and_libc_deps.c"

typedef enum MtpTypeKind {
    mtp_type_void,
    mtp_type_int,
    mtp_type_sym,
    mtp_type_lam,
} MtpTypeKind;

typedef enum MtpNodeKind {
    mtp_node_lam,
    mtp_node_param,
    mtp_node_choice,
    mtp_node_jump,
    mtp_node_val_int,
    mtp_node_val_sym,
    mtp_node_prim,
} MtpNodeKind;

typedef enum MtpNodePrimKind {
    mtp_prim_extcall,
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
} MtpTypeSym;

typedef struct MtpTypeLam {
    ·SliceOfPtrs(MtpType) params_types;
} MtpTypeLam;

struct MtpType {
    union {
        MtpTypeInt num_int;
        MtpTypeSym sym;
        MtpTypeLam lam;
    } of;
    MtpTypeKind kind;
};

struct MtpNode;
typedef struct MtpNode MtpNode;
typedef ·SliceOf(MtpNode) MtpNodes;

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
    MtpNode* dst_lam;
    ·SliceOfPtrs(MtpNode) args;
} MtpNodeJump;

typedef struct MtpNodeValInt {
    I64 int_val;
} MtpNodeValInt;

typedef struct MtpNodeValSym {
    UInt sym_val;
} MtpNodeValSym;

typedef struct MtpNodePrimExtCall {
    MtpType* params_types;
    Str name;
} MtpNodePrimExtCall;

typedef enum MtpPrimCmpIKind {
    mtp_cmp_i_eq,
    mtp_cmp_i_neq,
} MtpPrimCmpIKind;

typedef struct MtpNodePrimCmpI {
    MtpNode* lhs;
    MtpNode* rhs;
    MtpPrimCmpIKind kind;
} MtpNodePrimCmpI;

typedef enum MtpPrimBinIKind {
    mtp_bin_i_add,
    mtp_bin_i_mul,
} MtpPrimBinIKind;

typedef struct MtpNodePrimBinI {
    MtpNode* lhs;
    MtpNode* rhs;
    MtpPrimBinIKind kind;
} MtpNodePrimBinI;

typedef struct MtpNodePrim {
    union {
        MtpNodePrimExtCall ext_call;
        MtpNodePrimCmpI cmp;
        MtpNodePrimBinI bin;
    } of;
    MtpNodePrimKind kind;
} MtpNodePrim;

struct MtpNode {
    MtpType* type;
    union {
        MtpNodeLam lam;
        MtpNodeParam param;
        MtpNodeChoice choice;
        MtpNodeJump jump;
        MtpNodeValInt val_int;
        MtpNodeValSym val_sym;
        MtpNodePrim prim;
    } of;
    MtpNodeKind kind;
    Bool side_effects;
};

typedef struct MtpProg {
    struct {
        ·ListOfPtrs(MtpNode) lams;
        ·ListOfPtrs(MtpType) types;
        ·ListOfPtrs(MtpNode) int_vals;
        ·ListOfPtrs(MtpNode) sym_vals;
        ·ListOfPtrs(MtpNode) prims;
        ·ListOfPtrs(MtpNode) choices;
        ·ListOfPtrs(MtpNode) jumps;
    } all;
} MtpProg;

Str mtpStrType(MtpType* const type) {
    switch (type->kind) {
        case mtp_type_void: return str("void");
        case mtp_type_int: return str2(str(type->of.num_int.unsign ? "u" : "i"), uIntToStr(type->of.num_int.bit_width, 1, 10));
        case mtp_type_sym: return str("sym");
        case mtp_type_lam: return str2(str("lam"), uIntToStr(type->of.lam.params_types.len, 1, 10)); // TODO when pressing
        default: ·fail(uIntToStr(type->kind, 1, 10));
    }
    return ·len0(U8);
}

// a ° b  ===  b ° a
Bool mtpNodePrimIsCommutative(MtpNodePrimKind const prim_kind, int const op_kind) {
    return (prim_kind == mtp_prim_bin_i && (op_kind == mtp_bin_i_add || op_kind == mtp_bin_i_mul))
           || (prim_kind == mtp_prim_cmp_i && (op_kind == mtp_cmp_i_eq || op_kind == mtp_cmp_i_neq));
}

// (a ° b) ° c  ===  a ° (b ° c)
Bool mtpNodePrimIsAssociative(MtpNodePrimKind const prim_kind, int const op_kind) {
    return prim_kind == mtp_prim_bin_i && (op_kind == mtp_bin_i_add || op_kind == mtp_bin_i_mul);
}

// a °¹ (b °² c)  ===  (a °¹ b)  °²  (a °¹ c)
Bool mtpNodePrimIsDistributive(MtpNodePrimKind const prim_kind, int const op_kind1, int const op_kind2) {
    return prim_kind == mtp_prim_bin_i && op_kind1 == mtp_bin_i_mul && op_kind2 == mtp_bin_i_add;
}

Bool mtpNodeIsVal(MtpNode* const node) {
    return node->kind == mtp_node_val_int || node->kind == mtp_node_val_sym;
}

Bool mtpTypesEql(MtpType* const t1, MtpType* const t2) {
    if (t1 == t2)
        return true;
    if (t1->kind == t2->kind)
        switch (t1->kind) {
            case mtp_type_void:
            case mtp_type_sym: return true;
            case mtp_type_int:
                return (t1->of.num_int.unsign == t2->of.num_int.unsign && t1->of.num_int.bit_width == t2->of.num_int.bit_width)
                       || (t1->of.num_int.bit_width == 0 && !t1->of.num_int.unsign)
                       || (t2->of.num_int.bit_width == 0 && !t2->of.num_int.unsign);
            case mtp_type_lam: {
                if (t1->of.lam.params_types.len == t2->of.lam.params_types.len) {
                    if (t1->of.lam.params_types.at != t2->of.lam.params_types.at)
                        for (UInt i = 0; i < t1->of.lam.params_types.len; i += 1)
                            if (!mtpTypesEql(t1->of.lam.params_types.at[i], t2->of.lam.params_types.at[i]))
                                return false;
                    return true;
                }
            } break;
            default: ·fail(uIntToStr(t1->kind, 1, 10));
        }
    return false;
}

Bool mtpNodesEql(MtpNode* const n1, MtpNode* const n2) {
    // outside ctors, always true:
    if (n1 == n2)
        return true;
    // only when called from ctors:
    if (n1->kind == n2->kind && n1->side_effects == n2->side_effects && mtpTypesEql(n1->type, n2->type))
        switch (n1->kind) {
            case mtp_node_val_int: {
                return n1->of.val_int.int_val == n2->of.val_int.int_val;
            }
            case mtp_node_val_sym: {
                return n1->of.val_sym.sym_val == n2->of.val_sym.sym_val;
            }
            case mtp_node_choice: {
                return mtpNodesEql(n1->of.choice.if0, n2->of.choice.if0) && mtpNodesEql(n1->of.choice.if1, n2->of.choice.if1)
                       && mtpNodesEql(n1->of.choice.cond, n2->of.choice.cond);
            }
            case mtp_node_jump: {
                if (n1->of.jump.args.len == n2->of.jump.args.len && mtpNodesEql(n1->of.jump.dst_lam, n2->of.jump.dst_lam)) {
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
                        case mtp_prim_bin_i:
                            return n1->of.prim.of.bin.kind == n2->of.prim.of.bin.kind
                                   && ((mtpNodesEql(n1->of.prim.of.bin.lhs, n2->of.prim.of.bin.lhs)
                                        && mtpNodesEql(n1->of.prim.of.bin.rhs, n2->of.prim.of.bin.rhs))
                                       || (mtpNodePrimIsCommutative(mtp_prim_bin_i, n1->of.prim.of.bin.kind)
                                           && (!(n1->side_effects || n2->side_effects))
                                           && (mtpNodesEql(n1->of.prim.of.bin.rhs, n2->of.prim.of.bin.lhs)
                                               && mtpNodesEql(n1->of.prim.of.bin.lhs, n2->of.prim.of.bin.rhs))));
                        case mtp_prim_cmp_i: // NOTE same as above; 3rd time around, extract!
                            return n1->of.prim.of.cmp.kind == n2->of.prim.of.cmp.kind
                                   && ((mtpNodesEql(n1->of.prim.of.cmp.lhs, n2->of.prim.of.cmp.lhs)
                                        && mtpNodesEql(n1->of.prim.of.cmp.rhs, n2->of.prim.of.cmp.rhs))
                                       || (mtpNodePrimIsCommutative(mtp_prim_cmp_i, n1->of.prim.of.cmp.kind)
                                           && (!(n1->side_effects || n2->side_effects))
                                           && (mtpNodesEql(n1->of.prim.of.cmp.rhs, n2->of.prim.of.cmp.lhs)
                                               && mtpNodesEql(n1->of.prim.of.cmp.lhs, n2->of.prim.of.cmp.rhs))));
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

MtpType* mtpType(MtpProg* const prog, MtpTypeKind const kind, PtrAny const type_spec) {
    MtpType specd_type = (MtpType) {.kind = kind};
    switch (kind) {
        case mtp_type_void:
        case mtp_type_sym: break;
        case mtp_type_int: specd_type.of.num_int = *((MtpTypeInt*)type_spec); break;
        case mtp_type_lam: specd_type.of.lam = *((MtpTypeLam*)type_spec); break;
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
MtpType* mtpTypeVoid(MtpProg* const prog) {
    return prog->all.types.at[0];
}
MtpType* mtpTypeBool(MtpProg* const prog) {
    return prog->all.types.at[1];
}
MtpType* mtpTypeLam(MtpProg* const prog, MtpTypeLam type_spec) {
    return mtpType(prog, mtp_type_lam, &type_spec);
}

MtpNode* mtpNodeValInt(MtpProg* const prog, I64 const int_val) {
    for (UInt i = 0; i < prog->all.int_vals.len; i += 1) {
        MtpNode* node = prog->all.int_vals.at[i];
        if (node->of.val_int.int_val == int_val) // TODO: switch to mtpNodesEql
            return node;
    }
    ·append(prog->all.int_vals, ·new(MtpNode));
    MtpNode* ret_node = prog->all.int_vals.at[prog->all.int_vals.len - 1];
    ret_node->kind = mtp_node_val_int;
    ret_node->side_effects = false;
    ret_node->type = prog->all.types.at[2];
    ret_node->of.val_int = (MtpNodeValInt) {.int_val = int_val};
    return ret_node;
}

MtpNode* mtpNodeValSym(MtpProg* const prog, UInt const sym_val) {
    return prog->all.sym_vals.at[sym_val];
}

MtpNode* mtpNodePrim(MtpProg* const prog, MtpNodePrim spec, MtpType* type, Bool const can_side_effect) {
    MtpNode spec_node = (MtpNode) {.kind = mtp_node_prim, .type = type, .side_effects = can_side_effect, .of = {.prim = spec}};
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
MtpNode* mtpNodePrimExtCall(MtpProg* const prog, MtpNodePrimExtCall const spec, MtpType* const ret_type) {
    return mtpNodePrim(prog, (MtpNodePrim) {.kind = mtp_prim_extcall, .of = {.ext_call = spec}}, ret_type, true);
}
MtpNode* mtpNodePrimCmpI(MtpProg* const prog, MtpNodePrimCmpI spec) {
    Bool fail = !mtpTypesEql(spec.lhs->type, spec.rhs->type);
    switch (spec.lhs->type->kind) {
        case mtp_type_int: break;
        case mtp_type_sym: fail |= (spec.kind != mtp_cmp_i_eq && spec.kind != mtp_cmp_i_neq); break;
        default: fail = true;
    }
    if (fail)
        ·fail(str("incompatible operand types for int comparison operation"));
    if ((!(spec.lhs->side_effects || spec.rhs->side_effects)) && mtpNodeIsVal(spec.rhs) && !mtpNodeIsVal(spec.lhs))
        ·swap(MtpNode, spec.lhs, spec.rhs);
    return mtpNodePrim(prog, (MtpNodePrim) {.kind = mtp_prim_cmp_i, .of = {.cmp = spec}}, mtpTypeBool(prog),
                       spec.lhs->side_effects || spec.rhs->side_effects);
}
MtpNode* mtpNodePrimBinI(MtpProg* const prog, MtpNodePrimBinI const spec) {
    if (spec.lhs->type->kind != mtp_type_int || spec.rhs->type->kind != mtp_type_int || !mtpTypesEql(spec.lhs->type, spec.rhs->type))
        ·fail(str("incompatible operand types for int binary operation"));
    return mtpNodePrim(prog, (MtpNodePrim) {.kind = mtp_prim_bin_i, .of = {.bin = spec}}, spec.lhs->type,
                       spec.lhs->side_effects || spec.rhs->side_effects);
}

MtpNode* mtpNodeChoice(MtpProg* const prog, MtpNodeChoice const spec) {
    if (!mtpTypesEql(spec.if0->type, spec.if1->type))
        ·fail(str("incompatible choice result types"));
    if (!mtpTypesEql(spec.cond->type, mtpTypeBool(prog)))
        ·fail(str("choice condition isn't boolish"));
    MtpNode spec_node = (MtpNode) {.kind = mtp_node_choice,
                                   .type = spec.if0->type,
                                   .side_effects = (spec.cond->side_effects || spec.if0->side_effects || spec.if1->side_effects),
                                   .of = {.choice = spec}};
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
        .type = mtpTypeVoid(prog),
        .side_effects = spec.dst_lam->of.lam.body->side_effects,
        .of = {.jump = spec},
    };
    if (!spec_node.side_effects) // "jump" means CPS call, so track "side-effecting" as per usual FP slang
        for (UInt i = 0; i < spec.args.len; i += 1)
            if (spec.args.at[i]->side_effects) {
                spec_node.side_effects = true;
                break;
            }
    for (UInt i = 0; i < prog->all.jumps.len; i += 1) {
        MtpNode* node = prog->all.jumps.at[i];
        if (mtpNodesEql(node, &spec_node))
            return node;
    }

    if (spec.dst_lam->of.lam.params.len != spec.args.len)
        ·fail(str4(str("callee expected "), uIntToStr(spec.dst_lam->of.lam.params.len, 1, 10), str(" arg(s) but caller gave "),
                   uIntToStr(spec.args.len, 1, 10)));
    for (UInt i = 0; i < spec.args.len; i += 1) {
        MtpNode* arg = spec.args.at[i];
        MtpNode* param = &spec.dst_lam->of.lam.params.at[i];
        if (!mtpTypesEql(arg->type, param->type))
            ·fail(str6(str("type mismatch for arg "), uIntToStr(i, 1, 10), str(": callee expected "), mtpStrType(param->type),
                       str(" but caller gave "), mtpStrType(arg->type)));
    }

    ·append(prog->all.jumps, ·new(MtpNode));
    MtpNode* ret_node = prog->all.jumps.at[prog->all.jumps.len - 1];
    *ret_node = spec_node;
    return ret_node;
}

MtpNode* mtpLamParam(MtpNode* const lam_node, UInt const param_index) {
    return &lam_node->of.lam.params.at[param_index];
}

MtpNode* mtpLam(MtpProg* const prog, ·SliceOfPtrs(MtpType) const params) {
    ·append(prog->all.lams, ·new(MtpNode));
    MtpNode* ret_node = ·as(MtpNode, ·last(prog->all.lams));
    ret_node->kind = mtp_node_lam;
    ret_node->side_effects = false;
    ret_node->type = mtpTypeLam(prog, (MtpTypeLam) {.params_types = {.at = params.at, .len = params.len}});
    ret_node->of.lam = (MtpNodeLam) {.params = ·sliceOf(MtpNode, params.len, params.len)};
    for (UInt i = 0; i < params.len; i += 1)
        ret_node->of.lam.params.at[i] = (MtpNode) {
            .kind = mtp_node_param,
            .type = params.at[i],
            .side_effects = false,
            .of = {.param = (MtpNodeParam) {.param_idx = i}},
        };
    return ret_node;
}

MtpProg mtpProg(UInt const lams_capacity, UInt const types_capacity, UInt const int_vals_capacity, UInt const prims_capacity,
                UInt const choices_capacity, UInt const jumps_capacity, UInt const sym_vals_total_count) {
    MtpProg ret_prog = (MtpProg) {.all = {
                                      .types = ·listOfPtrs(MtpType, 2, types_capacity),
                                      .int_vals = ·listOfPtrs(MtpNode, 0, int_vals_capacity),
                                      .sym_vals = ·listOfPtrs(MtpNode, sym_vals_total_count, sym_vals_total_count),
                                      .lams = ·listOfPtrs(MtpNode, 0, lams_capacity),
                                      .prims = ·listOfPtrs(MtpNode, 0, prims_capacity),
                                      .choices = ·listOfPtrs(MtpNode, 0, choices_capacity),
                                      .jumps = ·listOfPtrs(MtpNode, 0, jumps_capacity),
                                  }};
    ret_prog.all.types.at[0]->kind = mtp_type_void;
    ret_prog.all.types.at[1]->kind = mtp_type_sym;
    mtpTypeInt(&ret_prog, (MtpTypeInt) {.bit_width = 1, .unsign = true});  // returned by `mtpTypeBool(MtpProg*)`
    mtpTypeInt(&ret_prog, (MtpTypeInt) {.bit_width = 0, .unsign = false}); // int literals
    mtpNodeValInt(&ret_prog, 0)->type = mtpTypeBool(&ret_prog);
    mtpNodeValInt(&ret_prog, 1)->type = mtpTypeBool(&ret_prog);
    mtpNodeValInt(&ret_prog, 0);
    mtpNodeValInt(&ret_prog, 1);
    for (UInt i = 0; i < sym_vals_total_count; i += 1)
        *ret_prog.all.sym_vals.at[i] = (MtpNode) {.kind = mtp_node_val_sym,
                                                  .type = ret_prog.all.types.at[1],
                                                  .side_effects = false,
                                                  .of = {.val_sym = (MtpNodeValSym) {.sym_val = i}}};
    return ret_prog;
}
