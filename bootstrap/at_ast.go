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

func astNodeToks(node *AstNode, all_toks Tokens) Tokens {
	return all_toks[node.toks_idx : node.toks_idx+node.toks_len]
}

func astNodeSrcStr(node *AstNode, full_src Str, all_toks Tokens) Str {
	return toksSrcStr(astNodeToks(node, all_toks), full_src)
}
