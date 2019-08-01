package atmoil

import (
	"github.com/go-leap/str"
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/lang"
)

func DbgPrintToStderr(node INode) { atmolang.DbgPrintToStderr(node.Print()) }
func DbgPrintToString(node INode) string {
	var buf ustr.Buf
	atmolang.PrintTo(nil, node.Print(), &buf.BytesWriter, false, 1)
	return buf.String()
}

func (me *IrNonValue) Print() atmolang.IAstNode {
	if me.OneOf.Undefined {
		return atmolang.Build.Ident(atmo.KnownIdentUndef)
	} else if me.Orig != nil && len(me.Orig.Toks()) > 0 {
		return atmolang.Build.Ident(me.Orig.Toks().First1().Lexeme)
	}
	return atmolang.Build.Ident("SpecialBadlyInitialized")
}
func (me *IrLitFloat) Print() atmolang.IAstNode  { return atmolang.Build.LitFloat(me.Val) }
func (me *IrLitUint) Print() atmolang.IAstNode   { return atmolang.Build.LitUint(me.Val) }
func (me *IrLitStr) Print() atmolang.IAstNode    { return atmolang.Build.LitStr(me.Val) }
func (me *IrIdentBase) Print() atmolang.IAstNode { return atmolang.Build.Ident(me.Val) }
func (me *IrIdentName) Print() atmolang.IAstNode {
	return me.IrExprLetBase.print(me.IrIdentBase.Print().(atmolang.IAstExpr))
}
func (me *IrAppl) Print() atmolang.IAstNode {
	return me.IrExprLetBase.print(atmolang.Build.Appl(me.Callee.Print().(atmolang.IAstExpr), me.CallArg.Print().(atmolang.IAstExpr)))
}
func (me *IrExprLetBase) print(body atmolang.IAstExpr) atmolang.IAstNode {
	if len(me.Defs) == 0 {
		return body
	}
	let := atmolang.Build.Let(body)
	let.Defs = make([]atmolang.AstDef, len(me.Defs))
	for i := range me.Defs {
		let.Defs[i] = *me.Defs[i].Print().(*atmolang.AstDef)
	}
	return let
}

func (me *IrDef) Print() atmolang.IAstNode {
	var argnames []string
	if me.Arg != nil {
		argnames = []string{me.Arg.Val}
	}
	if me.Body == nil {
		return atmolang.Build.Def(me.Name.Val, atmolang.Build.Ident("?!?!?!"), argnames...)
	}
	return atmolang.Build.Def(me.Name.Val, me.Body.Print().(atmolang.IAstExpr), argnames...)
}

func (me *IrDefTop) Print() atmolang.IAstNode {
	def := me.IrDef.Print().(*atmolang.AstDef)
	def.IsTopLevel = true
	return def
}

func (me *IrDefArg) Print() atmolang.IAstNode {
	return atmolang.Build.Arg(me.IrIdentBase.Print().(atmolang.IAstExprAtomic), nil)
}
