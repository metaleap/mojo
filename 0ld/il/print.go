package atmoil

import (
	"github.com/go-leap/str"
	. "github.com/metaleap/atmo/0ld"
	. "github.com/metaleap/atmo/0ld/ast"
)

func DbgPrintToStderr(node IIrNode) { PrintToStderr(node.Print()) }
func DbgPrintToString(node IIrNode) string {
	var buf ustr.Buf
	PrintTo(nil, node.Print(), &buf.BytesWriter, false, 1)
	return buf.String()
}

func (me *IrNonValue) Print() IAstNode {
	if me.OneOf.Undefined {
		return BuildAst.Ident(KnownIdentUndef)
	} else if me.Orig != nil && len(me.Orig.Toks()) != 0 {
		return BuildAst.Ident(me.Orig.Toks().First1().Lexeme)
	}
	return BuildAst.Ident("!?SpecialBadlyInitialized")
}
func (me *IrLitFloat) Print() IAstNode  { return BuildAst.LitFloat(me.Val) }
func (me *IrLitUint) Print() IAstNode   { return BuildAst.LitUint(me.Val) }
func (me *IrLitTag) Print() IAstNode    { return BuildAst.Tag(me.Val) }
func (me *IrIdentBase) Print() IAstNode { return BuildAst.Ident(me.Name) }
func (me *IrAppl) Print() IAstNode {
	return BuildAst.Appl(me.Callee.Print().(IAstExpr), me.CallArg.Print().(IAstExpr))
}

func (me *IrDef) print() IAstNode {
	if me.Body == nil {
		return BuildAst.Def(me.Ident.Name, BuildAst.Ident("?!?!?!"))
	}
	return BuildAst.Def(me.Ident.Name, me.Body.Print().(IAstExpr))
}

func (me *IrDef) Print() IAstNode {
	def := me.print().(*AstDef)
	def.IsTopLevel = true
	return def
}

func (me *IrArg) Print() IAstNode {
	return BuildAst.Arg(me.IrIdentBase.Print().(IAstExprAtomic), nil)
}

func (me *IrAbs) Print() IAstNode {
	name := "λ" + StrRand(false)
	return BuildAst.Let(BuildAst.Ident(name),
		*BuildAst.Def(name, me.Body.Print().(IAstExpr), me.Arg.Name))
}
