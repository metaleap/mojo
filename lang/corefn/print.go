package atmocorefn

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

func (me *AstAppl) Print() atmolang.IAstNode {
	return atmolang.B.Appl(me.Callee.Print().(atmolang.IAstExpr), me.Arg.Print().(atmolang.IAstExpr))
}

func (me *AstCases) Print() atmolang.IAstNode {
	alts := make([]atmolang.AstCase, len(me.Ifs))
	for i := range alts {
		alts[i].Body, alts[i].Conds = me.Thens[i].Print().(atmolang.IAstExpr), make([]atmolang.IAstExpr, len(me.Ifs[i]))
		for j := range alts[i].Conds {
			alts[i].Conds[j] = me.Ifs[i][j].Print().(atmolang.IAstExpr)
		}
	}
	return atmolang.B.Cases(atmolang.B.IdentTrue(), alts...)
}

func (me *AstLet) Print() atmolang.IAstNode {
	let := atmolang.B.Let(me.Body.Print().(atmolang.IAstExpr))
	let.Defs = make([]atmolang.AstDef, len(me.Defs))
	for i := range me.Defs {
		let.Defs[i] = *me.Defs[i].Print().(*atmolang.AstDef)
	}
	return let
}

func (me *AstDefBase) Print() atmolang.IAstNode {
	argnames := make([]string, len(me.Args))
	for i := range argnames {
		argnames[i] = me.Args[i].AstIdentName.Val
	}
	return atmolang.B.Def(me.Name.Val, me.Body.Print().(atmolang.IAstExpr), argnames...)
}

func (me *AstDef) Print() atmolang.IAstNode {
	def := me.AstDefBase.Print().(*atmolang.AstDef)
	if len(me.Locals) > 0 {
		defs := make([]atmolang.AstDef, len(me.Locals))
		for i := range me.Locals {
			defs[i] = *(me.Locals[i].Print().(*atmolang.AstDef))
		}
		def.Body = atmolang.B.Let(def.Body, defs...)
	}
	return def
}

func (me *AstDefArg) Print() atmolang.IAstNode {
	return atmolang.B.Arg(me.AstIdentName.Print().(atmolang.IAstExprAtomic), nil)
}
