package atmolang

import (
	"strconv"

	"github.com/go-leap/std"
)

// IPrintFmt is fully implemented by `PrintFormatterMinimal`, for custom
// formatters it'll be best to embed this and then override specifics.
type IPrintFmt interface {
	SetCtxPrint(*CtxPrint)
	OnTopLevelChunk(*AstFileTopLevelChunk, *AstTopLevel)
	OnDef(*AstTopLevel, *AstDef)
	OnDefName(*AstDef, *AstIdent)
	OnDefArg(*AstDef, int, *AstDefArg)
	OnDefMeta(*AstDef, int, IAstExpr)
	OnDefBody(*AstDef, IAstExpr)
	OnExprLetBody(bool, *AstExprLet, IAstExpr)
	OnExprLetDef(bool, *AstExprLet, int, *AstDef)
	OnExprApplName(bool, *AstExprAppl, IAstExpr)
	OnExprApplArg(bool, *AstExprAppl, int, IAstExpr)
	OnExprCasesScrutinee(bool, *AstExprCases, IAstExpr)
	OnExprCasesCond(*AstCase, int, IAstExpr)
	OnExprCasesBody(*AstCase, IAstExpr)
	OnComment(IAstNode, IAstNode, *AstComment)
}

type CtxPrint struct {
	Fmt            IPrintFmt
	ApplStyle      ApplStyle
	CurTopLevel    *AstFileTopLevelChunk
	CurIndentLevel int
	OneIndentLevel string

	ustd.BytesWriter

	fmtCtxSet bool
}

func (me *CtxPrint) Print(node IAstNode) {
	var cmnts *astBaseComments
	if cmnt, _ := node.(IAstComments); cmnt != nil {
		cmnts = cmnt.Comments()
	}
	if cmnts != nil {
		for i := range cmnts.Leading {
			me.Print(&cmnts.Leading[i])
		}
	}
	if node.print(me); cmnts != nil {
		for i := range cmnts.Trailing {
			me.Print(&cmnts.Trailing[i])
		}
	}
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
	ctx := CtxPrint{Fmt: fmt,
		OneIndentLevel: "    ", ApplStyle: me.Options.ApplStyle, fmtCtxSet: true,
		BytesWriter: ustd.BytesWriter{Data: make([]byte, 0, 1024)},
	}
	fmt.SetCtxPrint(&ctx)
	for i := range me.TopLevel {
		ctx.CurTopLevel = &me.TopLevel[i]
		ctx.CurTopLevel.Print(&ctx)
	}
	return ctx.BytesWriter.Data
}

func (me *AstFileTopLevelChunk) Print(p *CtxPrint) {
	if !p.fmtCtxSet {
		p.fmtCtxSet = true
		p.Fmt.SetCtxPrint(p)
	}
	p.CurIndentLevel = 0
	p.Fmt.OnTopLevelChunk(me, &me.Ast)
}

func (me *AstTopLevel) print(p *CtxPrint) {
	if me.Def.Orig != nil {
		p.Fmt.OnDef(me, me.Def.Orig)
	}
}

func (me *AstComment) print(p *CtxPrint) {
	if me.IsSelfTerminating {
		p.WriteString("/*")
		p.WriteString(me.ContentText)
		p.WriteString("*/")
	} else {
		p.WriteString("//")
		p.WriteString(me.ContentText)
		p.WriteByte('\n')
	}
}

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
	p.WriteString(me.Tokens[0].Meta.Orig)
}

func (me *AstExprLitFloat) print(p *CtxPrint) {
	if len(me.Tokens) > 0 {
		p.Print(&me.AstBaseExprAtomLit)
	} else {
		p.WriteString(strconv.FormatFloat(me.Val, 'g', -1, 64))
	}
}

func (me *AstExprLitUint) print(p *CtxPrint) {
	if len(me.Tokens) > 0 {
		p.Print(&me.AstBaseExprAtomLit)
	} else {
		p.WriteString(strconv.FormatUint(me.Val, 10))
	}
}

func (me *AstExprLitRune) print(p *CtxPrint) {
	if len(me.Tokens) > 0 {
		p.Print(&me.AstBaseExprAtomLit)
	} else {
		p.WriteString(strconv.QuoteRune(me.Val))
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
	istopleveldefsbody := (me == p.CurTopLevel.Ast.Def.Orig.Body)
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
	istopleveldefsbody := (me == p.CurTopLevel.Ast.Def.Orig.Body)
	p.Fmt.OnExprLetBody(istopleveldefsbody, me, me.Body)
	for i := range me.Defs {
		p.WriteByte(',')
		p.Fmt.OnExprLetDef(istopleveldefsbody, me, i, &me.Defs[i])
	}
}

func (me *AstExprCases) print(p *CtxPrint) {
	istopleveldefsbody := (me == p.CurTopLevel.Ast.Def.Orig.Body)
	if me.Scrutinee != nil {
		p.Fmt.OnExprCasesScrutinee(istopleveldefsbody, me, me.Scrutinee)
		p.WriteByte('|')
	}
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
		p.WriteByte('?')
		p.Fmt.OnExprCasesBody(me, me.Body)
	}
}

// PrintFmtMinimal implements `IPrintFmt`.
type PrintFmtMinimal struct {
	*CtxPrint
}

func (me *PrintFmtMinimal) SetCtxPrint(ctxPrint *CtxPrint) { me.CtxPrint = ctxPrint }
func (me *PrintFmtMinimal) OnTopLevelChunk(tlc *AstFileTopLevelChunk, node *AstTopLevel) {
	me.WriteByte('\n')
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
	me.WriteString(" := ")
	me.Print(node)
}
func (me *PrintFmtMinimal) OnExprLetBody(_ bool, _ *AstExprLet, node IAstExpr) {
	me.Print(node)
}
func (me *PrintFmtMinimal) OnExprLetDef(_ bool, _ *AstExprLet, _ int, node *AstDef) {
	me.Print(node)
}
func (me *PrintFmtMinimal) OnExprApplName(_ bool, _ *AstExprAppl, node IAstExpr) {
	me.Print(node)
}
func (me *PrintFmtMinimal) OnExprApplArg(_ bool, appl *AstExprAppl, argIdx int, node IAstExpr) {
	claspish, svo := appl.Claspish(), (me.ApplStyle == APPLSTYLE_SVO)
	if (!claspish) && (me.ApplStyle == APPLSTYLE_VSO || (svo && argIdx > 0)) {
		me.WriteByte(' ')
	}
	me.Print(node)
	if (!claspish) && (me.ApplStyle == APPLSTYLE_SOV || (svo && argIdx == 0)) {
		me.WriteByte(' ')
	}
}
func (me *PrintFmtMinimal) OnExprCasesScrutinee(_ bool, _ *AstExprCases, node IAstExpr) {
	me.Print(node)
}
func (me *PrintFmtMinimal) OnExprCasesCond(_ *AstCase, _ int, node IAstExpr)   { me.Print(node) }
func (me *PrintFmtMinimal) OnExprCasesBody(_ *AstCase, node IAstExpr)          { me.Print(node) }
func (me *PrintFmtMinimal) OnComment(_ IAstNode, _ IAstNode, node *AstComment) { me.Print(node) }
