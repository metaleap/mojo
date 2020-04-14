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
	base  AstNode
	head  AstExpr
	body  AstExpr
	defs  []AstDef
	scope AstScopes
	anns  struct {
		is_top_def bool
		name       Str
	}
}

type AstExpr struct {
	base AstNode
	kind AstExprKind
}

type AstExprLitInt int64

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
	ref_def   *AstDef
	top_def   *AstDef
	param_idx int
}

func astNodeFrom(toks_idx int, toks_len int) AstNode {
	return AstNode{toks_idx: toks_idx, toks_len: toks_len}
}

func astNodeToks(node *AstNode, all_toks []Token) []Token {
	return all_toks[node.toks_idx : node.toks_idx+node.toks_len]
}

func astNodeSrcStr(node *AstNode, ast *Ast) Str {
	if node.toks_len == 0 {
		return nil
	}
	node_toks := astNodeToks(node, ast.toks)
	return toksSrcStr(node_toks, ast.src)
}

func astPopulateScopes(ast *Ast) {
	ast.scope.cur = allocˇAstNameRef(len(ast.defs))
	for i := range ast.defs {
		def := &ast.defs[i]
		ast.scope.cur[i] = AstNameRef{name: def.anns.name, param_idx: -1, top_def: def, ref_def: def}
	}
	for i := range ast.defs {
		astDefPopulateScopes(&ast.defs[i], &ast.defs[i], ast, &ast.scope)
	}
}

func astDefPopulateScopes(top_def *AstDef, cur_def *AstDef, ast *Ast, parent *AstScopes) {
	num_args := 0
	head_form, _ := cur_def.head.kind.(AstExprForm)
	if head_form != nil {
		num_args = len(head_form) - 1
	}

	cur_def.scope.parent = parent
	cur_def.scope.cur = allocˇAstNameRef(len(cur_def.defs) + num_args)
	for i := range cur_def.defs {
		sub_def := &cur_def.defs[i]
		if nil != astScopesResolve(&cur_def.scope, sub_def.anns.name, i) {
			fail("shadowing of identifier '", sub_def.anns.name, "' near:\n", astNodeSrcStr(&sub_def.base, ast))
		}
		cur_def.scope.cur[i] = AstNameRef{name: sub_def.anns.name, param_idx: -1, top_def: top_def, ref_def: sub_def}
	}
	if head_form != nil {
		for i := 1; i < len(head_form); i++ {
			param_idx := len(cur_def.defs) + (i - 1)
			param_name := head_form[i].kind.(AstExprIdent)
			if nil != astScopesResolve(&cur_def.scope, param_name, param_idx) {
				fail("shadowing of identifier '", param_name, "' near:\n", astNodeSrcStr(&cur_def.head.base, ast))
			}
			cur_def.scope.cur[param_idx] = AstNameRef{
				name:      param_name,
				param_idx: i - 1,
				top_def:   top_def,
				ref_def:   cur_def,
			}
		}
	}
	for i := range cur_def.defs {
		sub_def := &cur_def.defs[i]
		astDefPopulateScopes(top_def, sub_def, ast, &cur_def.scope)
	}
}

func astScopesResolve(scope *AstScopes, name Str, only_until_before_idx int) *AstNameRef {
	for i := range scope.cur {
		if i == only_until_before_idx {
			break
		} else if strEql(name, scope.cur[i].name) {
			return &scope.cur[i]
		}
	}
	if scope.parent != nil {
		return astScopesResolve(scope.parent, name, -1)
	}
	return nil
}

type AstExprKind interface{ implementsAstExprKind() }

func (AstExprForm) implementsAstExprKind()    {}
func (AstExprIdent) implementsAstExprKind()   {}
func (AstExprLitClip) implementsAstExprKind() {}
func (AstExprLitCurl) implementsAstExprKind() {}
func (AstExprLitInt) implementsAstExprKind()  {}
func (AstExprLitStr) implementsAstExprKind()  {}
