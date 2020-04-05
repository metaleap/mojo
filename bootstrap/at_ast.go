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

type AstExprLitClip []AstExpr

type AstExprLitCurl []AstExpr

type AstScopes struct {
	cur    []AstNameRef
	parent *AstScopes
}

type AstNameRef struct {
	name      Str
	refers_to Any
}

func astNodeFrom(toks_idx int, toks_len int) AstNode {
	return AstNode{toks_idx: toks_idx, toks_len: toks_len}
}

func astNodeToks(node *AstNode, all_toks Tokens) Tokens {
	return all_toks[node.toks_idx : node.toks_idx+node.toks_len]
}

func astNodeSrcStr(node *AstNode, full_src Str, all_toks Tokens) Str {
	return toksSrcStr(astNodeToks(node, all_toks), full_src)
}

func astExprIsLitStr(expr *AstExpr) bool {
	_, ok := expr.kind.(AstExprLitStr)
	return ok
}

func astExprIsBuiltin(expr *AstExpr) bool {
	switch expr_kind := expr.kind.(type) {
	case AstExprIdent:
		return expr_kind[0] == '/'
	case AstExprForm:
		ident, ok := expr_kind[0].kind.(AstExprIdent)
		return ok && ident[0] == '/'
	}
	return false
}

func astDefGatherAndRewriteLitStrs(def *AstDef, into []StrNamed, idx int) int {
	if def.is_top_def && astExprIsLitStr(&def.body) {
		return idx
	}
	idx = astExprGatherAndRewriteLitStrs(&def.body, into, idx)
	for i := range def.defs {
		idx = astDefGatherAndRewriteLitStrs(&def.defs[i], into, idx)
	}
	return idx
}

func astExprGatherAndRewriteLitStrs(expr *AstExpr, into []StrNamed, idx int) int {
	switch expr_kind := expr.kind.(type) {
	case AstExprForm:
		for i := range expr_kind {
			idx = astExprGatherAndRewriteLitStrs(&expr_kind[i], into, idx)
		}
	case AstExprLitCurl:
		for i := range expr_kind {
			idx = astExprGatherAndRewriteLitStrs(&expr_kind[i], into, idx)
		}
	case AstExprLitClip:
		for i := range expr_kind {
			idx = astExprGatherAndRewriteLitStrs(&expr_kind[i], into, idx)
		}
	case AstExprLitStr:
		counter++
		new_name := uintToStr(counter, 10, 1, Str(".str_"))
		expr.kind = AstExprIdent(new_name)
		into[idx] = StrNamed{name: new_name, value: expr_kind}
		idx++
	}
	return idx
}
