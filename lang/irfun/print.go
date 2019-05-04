package atmolang_irfun

import (
	"github.com/metaleap/atmo/lang"
)

func Print(ctxp *atmolang.CtxPrint) {
}

func (me *AstLitFloat) Print() atmolang.IAstNode  { return atmolang.B.LitFloat(me.Val) }
func (me *AstLitUint) Print() atmolang.IAstNode   { return atmolang.B.LitUint(me.Val) }
func (me *AstLitRune) Print() atmolang.IAstNode   { return atmolang.B.LitRune(me.Val) }
func (me *AstLitStr) Print() atmolang.IAstNode    { return atmolang.B.LitStr(me.Val) }
func (me *AstIdentBase) Print() atmolang.IAstNode { return atmolang.B.Ident(me.Val) }
func (me *AstIdentName) Print() atmolang.IAstNode {
	return me.AstExprLetBase.print(me.AstIdentBase.Print().(atmolang.IAstExpr))
}

func (me *AstAppl) Print() atmolang.IAstNode {
	return me.AstExprLetBase.print(atmolang.B.Appl(me.AtomicCallee.Print().(atmolang.IAstExpr), me.AtomicArg.Print().(atmolang.IAstExpr)))
}

func (me *AstCases) Print() atmolang.IAstNode {
	alts := make([]atmolang.AstCase, len(me.Ifs))
	for i := range alts {
		alts[i].Body = me.Thens[i].Print().(atmolang.IAstExpr)
		alts[i].Conds = []atmolang.IAstExpr{me.Ifs[i].Print().(atmolang.IAstExpr)}
	}
	return me.AstExprLetBase.print(atmolang.B.Cases(nil, alts...))
}

func (me *AstExprLetBase) print(body atmolang.IAstExpr) atmolang.IAstNode {
	if len(me.letDefs) == 0 {
		return body
	}
	let := atmolang.B.Let(body)
	let.Defs = make([]atmolang.AstDef, len(me.letDefs))
	for i := range me.letDefs {
		let.Defs[i] = *me.letDefs[i].Print().(*atmolang.AstDef)
	}
	return let
}

func (me *AstDef) Print() atmolang.IAstNode {
	var argname string
	if me.Arg != nil {
		argname = me.Arg.Val
	}
	return atmolang.B.Def(me.Name.Val, me.Body.Print().(atmolang.IAstExpr), argname)
}

func (me *AstDefTop) Print() atmolang.IAstNode {
	def := me.AstDef.Print().(*atmolang.AstDef)
	def.IsTopLevel = true
	return def
}

func (me *AstDefArg) Print() atmolang.IAstNode {
	return atmolang.B.Arg(me.AstIdentName.Print().(atmolang.IAstExprAtomic), nil)
}
