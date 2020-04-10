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
	kind AstExprKind
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

func astNodeSrcStr(node *AstNode, ast *Ast) Str {
	if node.toks_len == 0 {
		return nil
	}
	node_toks := astNodeToks(node, ast.toks)
	return toksSrcStr(node_toks, ast.src)
}

func astDefName(def *AstDef) Str {
	return def.head.kind.(AstExprIdent)
}

func astExprFormExtract(expr_form AstExprForm, idx_start int, idx_end int) AstExpr {
	sub_form := expr_form[idx_start:idx_end]
	ret_expr := AstExpr{kind: sub_form}
	ret_expr.base.toks_idx = expr_form[idx_start].base.toks_idx
	for i := idx_start; i < idx_end; i++ {
		ret_expr.base.toks_len += expr_form[i].base.toks_len
	}
	if form, is := ret_expr.kind.(AstExprForm); is && 1 == len(form) {
		ret_expr = form[0]
	}
	if form, is := ret_expr.kind.(AstExprForm); is {
		assert(len(form) > 1)
	}
	return ret_expr
}

func astExprFormSplit(expr *AstExpr, ident string, must bool, must_lhs bool, must_rhs bool, ast *Ast) (lhs *AstExpr, rhs *AstExpr) {
	ident_needle := Str(ident)
	if !must {
		assert(!(must_lhs || must_rhs))
	}
	idx := -1
	if form, _ := expr.kind.(AstExprForm); len(form) != 0 {
		for i := range form {
			switch maybe_ident := form[i].kind.(type) {
			case AstExprIdent:
				if strEql(ident_needle, maybe_ident) {
					idx = i
					break
				}
			}
		}
		if idx >= 0 {
			both := allocˇAstExpr(2)
			if idx > 0 {
				both[0] = astExprFormExtract(form, 0, idx)
				lhs = &both[0]
			}
			if idx < len(form)-1 {
				both[1] = astExprFormExtract(form, idx+1, len(form))
				rhs = &both[1]
			}
		}
	}
	if idx < 0 && must {
		fail("expected '", ident_needle, "' in:\n", astNodeSrcStr(&expr.base, ast))
	}
	if lhs == nil && must_lhs {
		fail("expected expression before '", ident_needle, "' in:\n", astNodeSrcStr(&expr.base, ast))
	}
	if rhs == nil && must_rhs {
		fail("expected expression after '", ident_needle, "' in:\n", astNodeSrcStr(&expr.base, ast))
	}
	return
}

func astExprIsIdent(expr *AstExpr, ident string) bool {
	if expr_ident, is_ident := expr.kind.(AstExprIdent); is_ident {
		return len(ident) == 0 || strEql(expr_ident, Str(ident))
	}
	return false
}

func astExprSlashed(expr *AstExpr) (ret_parts []*AstExpr) {
	if form, _ := expr.kind.(AstExprForm); form != nil {
		num_parts := 0
		for i := 0; i < len(form); i += 2 {
			if astExprIsIdent(&form[i], "/") && i != len(form)-1 {
				num_parts++
			} else {
				return
			}
		}
		if num_parts != 0 {
			ret_parts = allocˇAstExprPtr(num_parts)
			ret_idx := 0
			for i := 1; i < len(form); i += 2 {
				ret_parts[ret_idx] = &form[i]
				ret_idx++
			}
			assert(ret_idx == num_parts)
		}
	}
	return
}

func astExprsFindKeyedValue(exprs []AstExpr, key string, ast *Ast) *AstExpr {
	str_key := Str(key)
	for i := range exprs {
		if form, _ := exprs[i].kind.(AstExprForm); form != nil {
			lhs, rhs := astExprFormSplit(&exprs[i], ":", true, true, true, ast)
			if lhs != nil && strEql(astNodeSrcStr(&lhs.base, ast), str_key) {
				return rhs
			}
		}
	}
	return nil
}

func astExprTaggedIdent(expr *AstExpr) AstExprIdent {
	if form, is_form := expr.kind.(AstExprForm); len(form) == 2 {
		if ident_op, _ := form[0].kind.(AstExprIdent); len(ident_op) == 1 && ident_op[0] == '#' {
			ident_ret, _ := form[1].kind.(AstExprIdent)
			return ident_ret
		}
	} else if !is_form {
		assert(expr.kind.(AstExprIdent)[0] == '#')
	}
	return nil
}

func astPopulateScopes(ast *Ast) {
	ast.scope.cur = allocˇAstNameRef(len(ast.defs))
	for i := range ast.defs {
		def := &ast.defs[i]
		ast.scope.cur[i] = AstNameRef{name: astDefName(def), refers_to: def}
	}
	for i := range ast.defs {
		astDefPopulateScopes(&ast.defs[i], ast, &ast.scope)
	}
}

func astDefPopulateScopes(def *AstDef, ast *Ast, parent *AstScopes) {
	def.scope.parent = parent
	def.scope.cur = allocˇAstNameRef(len(def.defs))
	for i := range def.defs {
		sub_def := &def.defs[i]
		def_name := astDefName(sub_def)
		if nil != astScopesResolve(&def.scope, def_name, i) {
			fail("duplicate name '", def_name, "' near:\n", astNodeSrcStr(&def.base, ast))
		}
		def.scope.cur[i] = AstNameRef{name: def_name, refers_to: sub_def}
	}
	for i := range def.defs {
		sub_def := &def.defs[i]
		astDefPopulateScopes(sub_def, ast, &def.scope)
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

type AstExprKind interface{ implementsAstExprKind() }

func (AstExprForm) implementsAstExprKind()    {}
func (AstExprIdent) implementsAstExprKind()   {}
func (AstExprLitClip) implementsAstExprKind() {}
func (AstExprLitCurl) implementsAstExprKind() {}
func (AstExprLitInt) implementsAstExprKind()  {}
func (AstExprLitStr) implementsAstExprKind()  {}
