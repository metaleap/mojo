package atmoast

import (
	"io"
	"os"
	"strconv"

	"github.com/go-leap/std"
	. "github.com/metaleap/atmo"
)

func PrintToStderr(node IAstNode) { PrintTo(nil, node, os.Stderr, true, 1) }

func PrintTo(curTopLevel *AstDef, node IAstNode, out io.Writer, prominentForDebugPurposes bool, applStyle ApplStyle) {
	ctxp := &CtxPrint{NoComments: true, CurTopLevel: curTopLevel, OneIndentLevel: "    "}
	ctxp.Fmt = &PrintFmtPretty{}
	ctxp.ApplStyle = applStyle
	ctxp.Fmt.SetCtxPrint(ctxp)
	ctxp.Print(node)
	data := ctxp.BytesWriter.Data
	if prominentForDebugPurposes {
		data = append(append([]byte("\n\n\n▓▓▓"), data...), "▓▓▓\n\n\n"...)
	}
	if out == nil {
		out = os.Stderr
	}
	out.Write(data)
}

func (me *CtxPrint) Print(node IAstNode) *CtxPrint {
	var cmnts *astBaseComments
	if cmnt, _ := node.(IAstComments); (!me.NoComments) && cmnt != nil {
		cmnts = cmnt.Comments()
	}
	leadstrails := node
	if cmnts != nil {
		for i := range cmnts.Leading {
			c := &cmnts.Leading[i]
			me.Fmt.OnComment(leadstrails, nil, c)
		}
	}
	if node.print(me); cmnts != nil {
		for i := range cmnts.Trailing {
			c := &cmnts.Trailing[i]
			me.Fmt.OnComment(nil, leadstrails, c)
		}
	}
	return me
}

func (me *CtxPrint) WriteLineBreaksThenIndent(numLines int) {
	for i := 0; i < numLines; i++ {
		me.WriteByte('\n')
	}
	for i := 0; i < me.CurIndentLevel; i++ {
		me.WriteString(me.OneIndentLevel)
	}
}

func (me *AstFile) Print(fmt IPrintFmt) []byte {
	pctx := CtxPrint{Fmt: fmt,
		OneIndentLevel: "    ", ApplStyle: me.Options.ApplStyle, fmtCtxSet: true,
		BytesWriter: ustd.BytesWriter{Data: make([]byte, 0, 1024)},
	}
	fmt.SetCtxPrint(&pctx)
	for i := range me.TopLevel {
		me.TopLevel[i].Print(&pctx)
	}
	return pctx.BytesWriter.Data
}

func (me *AstFileChunk) Print(p *CtxPrint) {
	if p.CurIndentLevel, p.CurTopLevel = 0, me.Ast.Def.Orig; !p.fmtCtxSet {
		p.fmtCtxSet = true
		p.Fmt.SetCtxPrint(p)
	}
	p.Fmt.OnTopLevelChunk(me, &me.Ast)
	p.CurIndentLevel, p.CurTopLevel = 0, nil
}

func (me *AstTopLevel) print(p *CtxPrint) {
	if me.Def.Orig != nil {
		p.Fmt.OnDef(me, me.Def.Orig)
	}
}

func (me *AstComment) print(p *CtxPrint) {
	if me.IsLineComment {
		p.WriteString("//")
		p.WriteString(me.Val)
	} else {
		p.WriteString("/*")
		p.WriteString(me.Val)
		p.WriteString("*/")
	}
}

func (me *AstDef) Print(ctxp *CtxPrint) { me.print(ctxp) }

func (me *AstDef) print(p *CtxPrint) {
	switch p.ApplStyle {
	case APPLSTYLE_VSO:
		if p.Fmt.OnDefName(me, &me.Name); me.NameAffix != nil {
			p.WriteByte(':')
			p.Print(me.NameAffix)
		}
		for i := range me.Args {
			p.Fmt.OnDefArg(me, i, &me.Args[i])
		}
	case APPLSTYLE_SVO:
		if len(me.Args) > 0 {
			p.Fmt.OnDefArg(me, 0, &me.Args[0])
		}
		if p.Fmt.OnDefName(me, &me.Name); me.NameAffix != nil {
			p.WriteByte(':')
			p.Print(me.NameAffix)
		}
		for i := 1; i < len(me.Args); i++ {
			p.Fmt.OnDefArg(me, i, &me.Args[i])
		}
	case APPLSTYLE_SOV:
		for i := range me.Args {
			p.Fmt.OnDefArg(me, i, &me.Args[i])
		}
		if p.Fmt.OnDefName(me, &me.Name); me.NameAffix != nil {
			p.WriteByte(':')
			p.Print(me.NameAffix)
		}
	}
	for i := range me.Meta {
		p.WriteByte(',')
		p.Fmt.OnDefMeta(me, i, me.Meta[i])
	}
	p.WriteByte(' ')
	p.WriteString(KnownIdentDecl)
	p.Fmt.OnDefBody(me, me.Body)
}

func (me *AstDefArg) print(p *CtxPrint) {
	if p.Print(me.NameOrConstVal); me.Affix != nil {
		p.WriteByte(':')
		p.Print(me.Affix)
	}
}

func (me *AstIdent) print(p *CtxPrint) {
	p.WriteString(me.Val)
}

func (me *AstBaseExprAtomLit) print(p *CtxPrint) {
	p.WriteString(me.Tokens[0].Lexeme)
}

func (me *AstExprLitFloat) print(p *CtxPrint) {
	if len(me.Tokens) > 0 {
		p.Print(&me.AstBaseExprAtomLit)
	} else {
		p.WriteString(strconv.FormatFloat(me.Val, 'f', -1, 64))
	}
}

func (me *AstExprLitUint) print(p *CtxPrint) {
	if len(me.Tokens) > 0 {
		p.Print(&me.AstBaseExprAtomLit)
	} else {
		p.WriteString(strconv.FormatUint(me.Val, 10))
	}
}

func (me *AstExprLitStr) print(p *CtxPrint) {
	if len(me.Tokens) > 0 {
		p.Print(&me.AstBaseExprAtomLit)
	} else {
		p.WriteString(strconv.Quote(me.Val))
	}
}

func (me *AstExprAppl) print(p *CtxPrint) {
	istopleveldefsbody := (p.CurTopLevel == nil || me == p.CurTopLevel.Body)
	switch p.ApplStyle {
	case APPLSTYLE_VSO:
		p.Fmt.OnExprApplName(istopleveldefsbody, me, me.Callee)
		for i := range me.Args {
			p.Fmt.OnExprApplArg(istopleveldefsbody, me, i, me.Args[i])
		}
	case APPLSTYLE_SVO:
		if len(me.Args) > 0 {
			p.Fmt.OnExprApplArg(istopleveldefsbody, me, 0, me.Args[0])
		}
		p.Fmt.OnExprApplName(istopleveldefsbody, me, me.Callee)
		for i := 1; i < len(me.Args); i++ {
			p.Fmt.OnExprApplArg(istopleveldefsbody, me, i, me.Args[i])
		}
	case APPLSTYLE_SOV:
		for i := range me.Args {
			p.Fmt.OnExprApplArg(istopleveldefsbody, me, i, me.Args[i])
		}
		p.Fmt.OnExprApplName(istopleveldefsbody, me, me.Callee)
	}
}

func (me *AstExprLet) print(p *CtxPrint) {
	p.Fmt.OnExprLetBody(me, me.Body)
	for i := range me.Defs {
		p.Fmt.OnExprLetDef(me, i, &me.Defs[i])
	}
}

func (me *AstExprCases) print(p *CtxPrint) {
	istopleveldefsbody := (p.CurTopLevel == nil || me == p.CurTopLevel.Body)
	if me.Scrutinee != nil {
		p.Fmt.OnExprCasesScrutinee(istopleveldefsbody, me, me.Scrutinee)
	}
	p.WriteByte('?')
	for i := range me.Alts {
		if i > 0 {
			p.WriteByte('|')
		}
		me.Alts[i].print(p)
	}
}

func (me *AstCase) print(p *CtxPrint) {
	for i := range me.Conds {
		if i > 0 {
			p.WriteByte('|')
		}
		p.Fmt.OnExprCasesCond(me, i, me.Conds[i])
	}
	if me.Body != nil {
		p.WriteString("=>")
		p.Fmt.OnExprCasesBody(me, me.Body)
	}
}

func (me *PrintFmtMinimal) SetCtxPrint(ctxPrint *CtxPrint) { me.CtxPrint = ctxPrint }
func (me *PrintFmtMinimal) OnTopLevelChunk(tlc *AstFileChunk, node *AstTopLevel) {
	if isntfirst := (tlc != &tlc.SrcFile.TopLevel[0]); isntfirst {
		me.WriteByte('\n')
	}
	me.Print(node)
	me.WriteByte('\n')
}
func (me *PrintFmtMinimal) OnDef(_ *AstTopLevel, node *AstDef)  { me.Print(node) }
func (me *PrintFmtMinimal) OnDefName(_ *AstDef, node *AstIdent) { me.Print(node) }
func (me *PrintFmtMinimal) OnDefArg(_ *AstDef, argIdx int, node *AstDefArg) {
	if me.ApplStyle == APPLSTYLE_VSO || (me.ApplStyle == APPLSTYLE_SVO && argIdx > 0) {
		me.WriteByte(' ')
	}
	me.Print(node)
	if me.ApplStyle == APPLSTYLE_SOV || (me.ApplStyle == APPLSTYLE_SVO && argIdx == 0) {
		me.WriteByte(' ')
	}
}
func (me *PrintFmtMinimal) OnDefMeta(_ *AstDef, _ int, node IAstExpr) {
	me.WriteByte(' ')
	me.Print(node)
}
func (me *PrintFmtMinimal) OnDefBody(def *AstDef, node IAstExpr) {
	me.WriteByte(' ')
	me.Print(node)
}
func (me *PrintFmtMinimal) OnExprLetBody(_ *AstExprLet, node IAstExpr) {
	me.PrintInParensIf(node, true, true)
}
func (me *PrintFmtMinimal) OnExprLetDef(_ *AstExprLet, _ int, node *AstDef) {
	me.WriteByte(',')
	me.Print(node)
}
func (me *PrintFmtMinimal) OnExprApplName(_ bool, _ *AstExprAppl, node IAstExpr) {
	me.PrintInParensIf(node, false, true)
}
func (me *PrintFmtMinimal) OnExprApplArg(_ bool, appl *AstExprAppl, argIdx int, node IAstExpr) {
	claspish, svo := appl.ClaspishByTokens(), (me.ApplStyle == APPLSTYLE_SVO)
	if (!claspish) && (me.ApplStyle == APPLSTYLE_VSO || (svo && argIdx > 0)) {
		me.WriteByte(' ')
	}
	me.PrintInParensIf(node, false, true)
	if (!claspish) && (me.ApplStyle == APPLSTYLE_SOV || (svo && argIdx == 0)) {
		me.WriteByte(' ')
	}
}
func (me *PrintFmtMinimal) OnExprCasesScrutinee(_ bool, _ *AstExprCases, node IAstExpr) {
	me.PrintInParensIf(node, true, false)
}
func (me *PrintFmtMinimal) OnExprCasesCond(_ *AstCase, _ int, node IAstExpr) {
	me.PrintInParensIf(node, true, false)
}
func (me *PrintFmtMinimal) OnExprCasesBody(_ *AstCase, node IAstExpr) {
	me.PrintInParensIf(node, true, false)
}
func (me *PrintFmtMinimal) OnComment(leads IAstNode, trails IAstNode, node *AstComment) {
	tl, istoplevelleadingcomment := leads.(*AstTopLevel)
	if me.Print(node); node.IsLineComment {
		if !istoplevelleadingcomment {
			me.CurIndentLevel++
		}
		needsnolinebreak :=
			(tl != nil && tl.Def.Orig == nil && node == &tl.comments.Leading[len(tl.comments.Leading)-1]) ||
				(leads == nil && trails != nil && me.CurTopLevel != nil && node.Tokens.Last1() == me.CurTopLevel.Tokens.Last1())
		if !needsnolinebreak {
			me.WriteLineBreaksThenIndent(1)
		}
	}
}
func (me *PrintFmtMinimal) PrintInParensIf(node IAstNode, ifCases bool, ifNotAtomicOrClaspish bool) {
	_, isatomic := node.(IAstExprAtomic)
	_, iscases := node.(*AstExprCases)
	if appl, ok := node.(*AstExprAppl); ok && ifNotAtomicOrClaspish {
		isatomic = appl.ClaspishByTokens()
	}
	parens := (ifCases && iscases) || (ifNotAtomicOrClaspish && !isatomic)
	if parens {
		me.WriteByte('(')
	}
	me.Print(node)
	if parens {
		me.WriteByte(')')
	}
}

func (me *PrintFmtPretty) OnTopLevelChunk(tlc *AstFileChunk, node *AstTopLevel) {
	me.WriteByte('\n')
	me.Print(node)
	me.WriteString("\n\n")
}
func (me *PrintFmtPretty) OnComment(leads IAstNode, trails IAstNode, node *AstComment) {
	if trails != nil {
		me.WriteByte(' ')
	}
	me.PrintFmtMinimal.OnComment(leads, trails, node)
	if tl, _ := leads.(*AstTopLevel); leads != nil && tl == nil {
		me.WriteByte(' ')
	}
}
func (me *PrintFmtPretty) OnDefBody(def *AstDef, node IAstExpr) {
	if def.IsTopLevel {
		me.CurIndentLevel++
		me.WriteLineBreaksThenIndent(1)
	} else {
		me.WriteByte(' ')
	}
	me.Print(node)
	if def.IsTopLevel {
		me.CurIndentLevel--
	}
}
func (me *PrintFmtPretty) OnExprLetBody(_ *AstExprLet, node IAstExpr) {
	me.Print(node)
	me.CurIndentLevel++
	me.WriteLineBreaksThenIndent(1)
}
func (me *PrintFmtPretty) OnExprLetDef(let *AstExprLet, idx int, node *AstDef) {
	me.WriteString(", ")
	me.Print(node)
	if idx == len(let.Defs)-1 {
		me.CurIndentLevel--
	}
	me.WriteLineBreaksThenIndent(1)
}
func (me *PrintFmtPretty) OnExprCasesScrutinee(_ bool, _ *AstExprCases, node IAstExpr) {
	me.PrintInParensIf(node, true, false)
	me.WriteByte(' ')
}
func (me *PrintFmtPretty) OnExprCasesCond(_ *AstCase, _ int, node IAstExpr) {
	me.WriteByte(' ')
	me.PrintInParensIf(node, true, false)
	me.WriteByte(' ')
}
func (me *PrintFmtPretty) OnExprCasesBody(_ *AstCase, node IAstExpr) {
	me.WriteByte(' ')
	me.PrintInParensIf(node, true, false)
	me.WriteByte(' ')
}
