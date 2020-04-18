#pragma once
#include "std.h"
#include "at_toks.h"

typedef struct AstNode {
    Uint toks_idx;
    Uint toks_len;
} AstNode;

typedef struct AstExpr {
    AstNode base;
    struct {
        Uint parensed;
        Bool toks_throng;
    } anns;
} AstExpr;

struct AstDef;
typedef struct AstDef AstDef;
typedef SliceOf(AstDef) AstDefs;
struct AstDef {
    AstNode base;
    AstExpr head;
    AstExpr body;
    AstDefs sub_defs;
    struct {
        Bool is_top_def;
        Str name;
    } anns;
};

typedef struct Ast {
    Str src;
    Tokens toks;
    AstDefs top_defs;
} Ast;

typedef struct AstNameRef {
    Str name;
    AstDef *top_def;
    Uints sub_def_path;
    ˇUint param_idx;

} AstNameRef;
typedef SliceOf(AstNameRef) AstNameRefs;

struct AstScopes;
typedef struct AstScopes AstScopes;
struct AstScopes {
    AstNameRefs names;
    AstScopes *parent;
};
