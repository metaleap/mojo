package main

type Ast struct {
	src   Str
	toks  []Token
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

func astNodeToks(node *AstNode, all_toks []Token) []Token {
	return all_toks[node.toks_idx : node.toks_idx+node.toks_len]
}

func astNodeSrcStr(node *AstNode, full_src Str, all_toks []Token) Str {
	return toksSrcStr(astNodeToks(node, all_toks), full_src)
}

func astDefName(def *AstDef) Str {
	return def.head.kind.(AstExprIdent)
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

func astResolveIdents(ast *Ast) {
	ast.scope.cur = allocˇAstNameRef(len(ast.defs))
	for i := range ast.defs {
		def := &ast.defs[i]
		ast.scope.cur[i] = AstNameRef{name: astDefName(def), refers_to: def}
	}
	for i := range ast.defs {
		astDefResolveIdents(&ast.defs[i], ast, &ast.scope)
	}
}

func astDefResolveIdents(def *AstDef, ast *Ast, parent *AstScopes) {
	def.scope.parent = parent
	def.scope.cur = allocˇAstNameRef(len(def.defs))
	for i := range def.defs {
		sub_def := &def.defs[i]
		def_name := astDefName(sub_def)
		if nil != astScopesResolve(&def.scope, def_name, i) {
			fail("duplicate name '", def_name, "' near:\n", astNodeSrcStr(&def.base, ast.src, ast.toks))
		}
		def.scope.cur[i] = AstNameRef{name: def_name, refers_to: sub_def}
	}
	for i := range def.defs {
		sub_def := &def.defs[i]
		astDefResolveIdents(sub_def, ast, &def.scope)
	}
}

func astScopesResolve(scope *AstScopes, name Str, only_until_before_idx int) Any {
	for i := range scope.cur {
		if i == only_until_before_idx {
			break
		} else if strEql(name, scope.cur[i].name) {
			return scope.cur[i].refers_to
		}
	}
	if scope.parent != nil {
		return astScopesResolve(scope.parent, name, -1)
	}
	return nil
}
