package atmoil

import (
	"github.com/go-leap/str"
	. "github.com/metaleap/atmo"
	. "github.com/metaleap/atmo/ast"
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
	} else if me.Orig != nil && len(me.Orig.Toks()) > 0 {
		return BuildAst.Ident(me.Orig.Toks().First1().Lexeme)
	}
	return BuildAst.Ident("!?SpecialBadlyInitialized")
}
func (me *IrLitFloat) Print() IAstNode  { return BuildAst.LitFloat(me.Val) }
func (me *IrLitUint) Print() IAstNode   { return BuildAst.LitUint(me.Val) }
func (me *IrLitTag) Print() IAstNode    { return BuildAst.Tag(me.Val) }
func (me *IrIdentBase) Print() IAstNode { return BuildAst.Ident(me.Val) }
func (me *IrIdentName) Print() IAstNode {
	return me.IrExprLetBase.print(me.IrIdentBase.Print().(IAstExpr))
}
func (me *IrAppl) Print() IAstNode {
	return me.IrExprLetBase.print(BuildAst.Appl(me.Callee.Print().(IAstExpr), me.CallArg.Print().(IAstExpr)))
}
func (me *IrExprLetBase) print(body IAstExpr) IAstNode {
	if len(me.Defs) == 0 {
		return body
	}
	let := BuildAst.Let(body)
	let.Defs = make([]AstDef, len(me.Defs))
	for i := range me.Defs {
		let.Defs[i] = *me.Defs[i].Print().(*AstDef)
	}
	return let
}

func (me *IrDef) Print() IAstNode {
	if me.Body == nil {
		return BuildAst.Def(me.Name.Val, BuildAst.Ident("?!?!?!"))
	}
	return BuildAst.Def(me.Name.Val, me.Body.Print().(IAstExpr))
}

func (me *IrDefTop) Print() IAstNode {
	def := me.IrDef.Print().(*AstDef)
	def.IsTopLevel = true
	return def
}

func (me *IrArg) Print() IAstNode {
	return BuildAst.Arg(me.IrIdentBase.Print().(IAstExprAtomic), nil)
}

func (me *IrLam) Print() IAstNode {
	return BuildAst.Let(BuildAst.Ident("λ"),
		*BuildAst.Def("λ", me.Body.Print().(IAstExpr), me.Arg.Val))
}
