package atmolang_irfun

import (
	"os"

	"github.com/go-leap/str"
	"github.com/metaleap/atmo/lang"
)

func DbgPrintToStderr(node IAstNode) { atmolang.PrintTo(nil, node.Print(), os.Stderr, true, 1) }
func DbgPrintToString(node IAstNode) string {
	var buf ustr.Buf
	atmolang.PrintTo(nil, node.Print(), &buf.BytesWriter, false, 1)
	return buf.String()
}

func (me *AstLitFloat) Print() atmolang.IAstNode  { return atmolang.B.LitFloat(me.Val) }
func (me *AstLitUint) Print() atmolang.IAstNode   { return atmolang.B.LitUint(me.Val) }
func (me *AstLitRune) Print() atmolang.IAstNode   { return atmolang.B.LitRune(me.Val) }
func (me *AstLitStr) Print() atmolang.IAstNode    { return atmolang.B.LitStr(me.Val) }
func (me *AstIdentBase) Print() atmolang.IAstNode { return atmolang.B.Ident(me.Val) }
func (me *AstIdentName) Print() atmolang.IAstNode {
	return me.AstExprLetBase.print(me.AstIdentBase.Print().(atmolang.IAstExpr))
}
func (me *AstLitUndef) Print() atmolang.IAstNode { return atmolang.B.Ident("()") }

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
