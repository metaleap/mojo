package atmolang

func (me *AstTopLevel) EnsureDesugared() {
	if me.Def.Desugared != nil {
		return
	}
	if me.Def.Desugared = me.Def.Orig.desugar(); me.Def.Desugared == nil {
		me.Def.Desugared = me.Def.Orig
	}
}

func (me *AstDef) desugarExpr(expr IAstExpr) IAstExpr {
	switch x := expr.(type) {
	case *AstExprCase:

		return x.desugar()
	}
	return nil
}

func (me *AstDef) desugar() *AstDef {

	return nil
}

func (me *AstExprCase) desugar() *AstExprCase {
	if me.IsUnionSugar {

	}
	return nil
}
