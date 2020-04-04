package main

type Ast struct {
	src   Str
	toks  Tokens
	defs  []AstDef
	scope AstScopes
}

type AstNode struct {
	toks_idx int
	toks_len int
}

type AstDef struct {
	base       AstNode
	head       AstExpr
	body       AstExpr
	defs       []AstDef
	scope      AstScopes
	is_top_def bool
}

type AstExpr struct {
	base AstNode
	kind Any
}

type AstExprLitInt uint64

type AstExprLitStr Str

type AstExprIdent Str

type AstExprForm []AstExpr

type AstExprLitList []AstExpr

type AstExprLitCurl []struct {
	lhs AstExpr
	rhs AstExpr
}

type AstScopes struct {
	cur    []AstNameRef
	parent *AstScopes
}

type AstNameRef struct {
	name      Str
	refers_to Any
}

func astDefGatherAndRewriteLitStrs(def *AstDef, into [][2]Str, idx int) int {
	return idx
}
