package atmoil

import (
	"github.com/go-leap/str"
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/lang"
)

func DbgPrintToStderr(node IAstNode) { atmolang.DbgPrintToStderr(node.Print()) }
func DbgPrintToString(node IAstNode) string {
	var buf ustr.Buf
	atmolang.PrintTo(nil, node.Print(), &buf.BytesWriter, false, 1)
	return buf.String()
}

func (me *AstSpecial) Print() atmolang.IAstNode {
	if me.OneOf.Undefined {
		return atmolang.B.Ident(atmo.KnownIdentUndef)
	} else if me.Orig != nil && len(me.Orig.Toks()) > 0 {
		return atmolang.B.Ident(me.Orig.Toks().First(nil).Meta.Orig)
	}
	return atmolang.B.Ident("SpecialBadlyInitialized")
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

func (me *AstExprLetBase) print(body atmolang.IAstExpr) atmolang.IAstNode {
	if len(me.Defs) == 0 {
		return body
	}
	let := atmolang.B.Let(body)
	let.Defs = make([]atmolang.AstDef, len(me.Defs))
	for i := range me.Defs {
		let.Defs[i] = *me.Defs[i].Print().(*atmolang.AstDef)
	}
	return let
}

func (me *AstDef) Print() atmolang.IAstNode {
	var argnames []string
	if me.Arg != nil {
		argnames = []string{me.Arg.Val}
	}
	if me.Body == nil {
		return atmolang.B.Def(me.Name.Val, atmolang.B.Ident("?!?!?!"), argnames...)
	}
	return atmolang.B.Def(me.Name.Val, me.Body.Print().(atmolang.IAstExpr), argnames...)
}

func (me *AstDefTop) Print() atmolang.IAstNode {
	def := me.AstDef.Print().(*atmolang.AstDef)
	def.IsTopLevel = true
	return def
}

func (me *AstDefArg) Print() atmolang.IAstNode {
	return atmolang.B.Arg(me.AstIdentBase.Print().(atmolang.IAstExprAtomic), nil)
}
