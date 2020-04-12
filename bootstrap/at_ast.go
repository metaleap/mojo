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
		return len(ident) == 0 || strEq(expr_ident, ident)
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
	for i := range exprs {
		if form, _ := exprs[i].kind.(AstExprForm); form != nil {
			lhs, rhs := astExprFormSplit(&exprs[i], ":", true, true, true, ast)
			if lhs != nil && strEq(astNodeSrcStr(&lhs.base, ast), key) {
				return rhs
			}
		}
	}
	return nil
}

func astExprTaggedIdent(expr *AstExpr) AstExprIdent {
	if form, _ := expr.kind.(AstExprForm); len(form) == 2 {
		if ident_op, _ := form[0].kind.(AstExprIdent); len(ident_op) == 1 && ident_op[0] == '#' {
			ident_ret, _ := form[1].kind.(AstExprIdent)
			return ident_ret
		}
	} else if ident, _ := expr.kind.(AstExprIdent); len(ident) == 1 && ident[0] == '#' {
		return AstExprIdent("")
	}
	return nil
}

func astExprGatherDefRefs(scope *AstScopes, ast_expr *AstExpr, gathered map[*AstDef][]*AstExpr) {
	switch expr := ast_expr.kind.(type) {
	case AstExprForm:
		for i := range expr {
			astExprGatherDefRefs(scope, &expr[i], gathered)
		}
	case AstExprLitClip:
		for i := range expr {
			astExprGatherDefRefs(scope, &expr[i], gathered)
		}
	case AstExprLitCurl:
		for i := range expr {
			astExprGatherDefRefs(scope, &expr[i], gathered)
		}
	case AstExprIdent:
		if ref := astScopesResolve(scope, expr, -1); ref != nil && ref.param_idx == -1 {
			gathered[ref.ref_def] = append(gathered[ref.ref_def], ast_expr)
		}
	}
}

func astDefCountAllSubDefs(def *AstDef) int {
	num_sub_defs := len(def.defs)
	for i := range def.defs {
		num_sub_defs += astDefCountAllSubDefs(&def.defs[i])
	}
	return num_sub_defs
}

func astHoistLocalDefsToTopDefs(ast *Ast, top_def_name Str) {
	astPopulateScopes(ast)
	top_defs_len := len(ast.defs)
	for i := range ast.defs {
		top_defs_len += astDefCountAllSubDefs(&ast.defs[i])
	}
	top_defs := allocˇAstDef(top_defs_len)
	top_defs_len = 0
	for i := range ast.defs {
		if strEql(ast.defs[i].anns.name, top_def_name) {
			top_defs_len = astDefHoistLocalDefsToTopDefs(ast, &ast.defs[i], top_defs, top_defs_len, nil, false)
			break
		}
	}
	ast.defs = top_defs[0:top_defs_len]
	astPopulateScopes(ast)
}

func astDefHoistLocalDefsToTopDefs(ast *Ast, cur_def *AstDef, top_defs []AstDef, top_defs_len int, name_pref Str, will_add bool) int {
	name_pref = strConcat([]Str{name_pref, cur_def.anns.name, Str(".")})
	gathered := make(map[*AstDef][]*AstExpr)
	astExprGatherDefRefs(&cur_def.scope, &cur_def.body, gathered)
	for refd_def, ast_exprs := range gathered {
		if refd_def.anns.is_top_def {
			top_defs_len = astDefHoistLocalDefsToTopDefs(ast, refd_def, top_defs, top_defs_len, nil, false)
		} else {
			new_name := strConcat([]Str{name_pref, refd_def.anns.name})
			for i := range ast_exprs {
				ast_exprs[i].kind = AstExprIdent(new_name)
			}
			already_hoisted := false
			for i := range top_defs[0:top_defs_len] {
				if strEql(top_defs[i].anns.name, new_name) {
					already_hoisted = true
					break
				}
			}
			if !already_hoisted {
				top_defs[top_defs_len] = *refd_def
				top_defs[top_defs_len].anns.name = new_name
				top_defs_len++
				top_defs_len = astDefHoistLocalDefsToTopDefs(ast, refd_def, top_defs, top_defs_len, name_pref, true)
			}
		}
	}
	if !will_add {
		top_defs[top_defs_len] = *cur_def
		top_defs_len++
	}
	return top_defs_len
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
