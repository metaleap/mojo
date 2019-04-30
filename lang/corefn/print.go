package atmocorefn

import (
	"github.com/metaleap/atmo/lang"
)

func Print(ctxp *atmolang.CtxPrint) {
}

func (me *AstLitFloat) Print() atmolang.IAstNode  { return atmolang.Builder.LitFloat(me.Val) }
func (me *AstLitUint) Print() atmolang.IAstNode   { return atmolang.Builder.LitUint(me.Val) }
func (me *AstLitRune) Print() atmolang.IAstNode   { return atmolang.Builder.LitRune(me.Val) }
func (me *AstLitStr) Print() atmolang.IAstNode    { return atmolang.Builder.LitStr(me.Val) }
func (me *AstIdentBase) Print() atmolang.IAstNode { return atmolang.Builder.Ident(me.Val) }

func (me *AstAppl) Print() atmolang.IAstNode {
	return atmolang.Builder.Appl(me.Callee.Print().(atmolang.IAstExpr), me.Arg.Print().(atmolang.IAstExpr))
}

func (me *AstCases) Print() atmolang.IAstNode {
	alts := make([]atmolang.AstCase, len(me.Ifs))
	for i := range alts {
		alts[i].Body, alts[i].Conds = me.Thens[i].Print().(atmolang.IAstExpr), make([]atmolang.IAstExpr, len(me.Ifs[i]))
		for j := range alts[i].Conds {
			alts[i].Conds[j] = me.Ifs[i][j].Print().(atmolang.IAstExpr)
		}
	}
	return atmolang.Builder.Cases(atmolang.Builder.IdentTrue(), alts...)
}

func (me *AstDefBase) Print() atmolang.IAstNode {
	argnames := make([]string, len(me.Args))
	for i := range argnames {
		argnames[i] = me.Args[i].AstIdentName.String()
	}
	return atmolang.Builder.Def(me.Name.String(), me.Body.Print().(atmolang.IAstExpr), argnames...)
}

func (me *AstDef) Print() atmolang.IAstNode {
	def := me.AstDefBase.Print().(*atmolang.AstDef)
	if len(me.Locals) > 0 {
		defs := make([]atmolang.AstDef, len(me.Locals))
		for i := range me.Locals {
			defs[i] = *(me.Locals[i].Print().(*atmolang.AstDef))
		}
		def.Body = atmolang.Builder.Let(def.Body, defs...)
	}
	return def
}

func (me *AstDefArg) Print() atmolang.IAstNode {
	return atmolang.Builder.Arg(me.AstIdentName.Print().(atmolang.IAstExprAtomic), nil)
}
