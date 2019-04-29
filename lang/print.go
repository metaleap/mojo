package atmolang

import (
	"strconv"

	"github.com/go-leap/std"
)

// IPrintFormatter is fully implemented by `PrintFormatterMinimal`, for custom
// formatters it'll be best to embed this and partially override specifics.
type IPrintFormatter interface {
	SetCtxPrint(*CtxPrint)
	OnTopLevelChunk(*AstFileTopLevelChunk, *AstTopLevel)
	OnDef(*AstTopLevel, *AstDef)
	OnDefName(*AstDef, *AstIdent)
	OnDefArg(*AstDef, int, *AstDefArg)
	OnDefMeta(*AstDef, int, IAstExpr)
	OnDefBody(*AstDef, IAstExpr)
	OnExprLetBody(bool, *AstExprLet, IAstExpr)
	OnExprLetDef(bool, *AstExprLet, int, *AstDef)
	OnExprApplName(*AstExprAppl, IAstExpr)
	OnExprApplArg(*AstExprAppl, int, IAstExpr)
}

type CtxPrint struct {
	Fmt            IPrintFormatter
	ApplStyle      ApplStyle
	CurTopLevel    *AstFileTopLevelChunk
	CurIndentLevel int
	OneIndentLevel string

	ustd.BytesWriter

	fmtCtxSet bool
}

func (me *CtxPrint) WriteLineBreaksThenIndent(numLines int) {
	for i := 0; i < numLines; i++ {
		me.WriteByte('\n')
	}
	for i := 0; i < me.CurIndentLevel; i++ {
		me.WriteString(me.OneIndentLevel)
	}
}

func (me *AstFile) Print(fmt IPrintFormatter) []byte {
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

func (me *AstTopLevel) Print(p *CtxPrint) {
	if me.Def.Orig != nil {
		p.Fmt.OnDef(me, me.Def.Orig)
	}
}

func (me *AstComment) Print(p *CtxPrint) {
	if me.IsSelfTerminating {
		p.WriteString("/*")
		p.WriteString(me.ContentText)
		p.WriteString("*/")
	} else {
		p.WriteString("//")
		p.WriteString(me.ContentText)
		p.WriteLineBreaksThenIndent(1)
	}
}

func (me *AstDef) Print(p *CtxPrint) {
	switch p.ApplStyle {
	case APPLSTYLE_VSO:
		if p.Fmt.OnDefName(me, &me.Name); me.NameAffix != nil {
			p.WriteByte(':')
			me.NameAffix.Print(p)
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
			me.NameAffix.Print(p)
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
			me.NameAffix.Print(p)
		}
	}
	for i := range me.Meta {
		p.WriteByte(',')
		p.Fmt.OnDefMeta(me, i, me.Meta[i])
	}
	p.Fmt.OnDefBody(me, me.Body)
}

func (me *AstDefArg) Print(p *CtxPrint) {
	if me.NameOrConstVal.Print(p); me.Affix != nil {
		p.WriteByte(':')
		me.Affix.Print(p)
	}
}

func (me *AstIdent) Print(p *CtxPrint) {
	p.WriteString(me.Val)
}

func (me *AstExprLitBase) Print(p *CtxPrint) {
	p.WriteString(me.Tokens[0].Meta.Orig)
}

func (me *AstExprLitFloat) Print(p *CtxPrint) {
	if len(me.Tokens) > 0 {
		me.AstExprLitBase.Print(p)
	} else {
		p.WriteString(strconv.FormatFloat(me.Val, 'g', -1, 64))
	}
}

func (me *AstExprLitUint) Print(p *CtxPrint) {
	if len(me.Tokens) > 0 {
		me.AstExprLitBase.Print(p)
	} else {
		p.WriteString(strconv.FormatUint(me.Val, 10))
	}
}

func (me *AstExprLitRune) Print(p *CtxPrint) {
	if len(me.Tokens) > 0 {
		me.AstExprLitBase.Print(p)
	} else {
		p.WriteString(strconv.QuoteRune(me.Val))
	}
}

func (me *AstExprLitStr) Print(p *CtxPrint) {
	if len(me.Tokens) > 0 {
		me.AstExprLitBase.Print(p)
	} else {
		p.WriteString(strconv.Quote(me.Val))
	}
}

func (me *AstExprAppl) Print(p *CtxPrint) {
	p.WriteByte('(')
	switch p.ApplStyle {
	case APPLSTYLE_VSO:
		me.Callee.Print(p)
		for i := range me.Args {
			p.WriteByte(' ')
			me.Args[i].Print(p)
		}
	case APPLSTYLE_SVO:
		if len(me.Args) > 0 {
			me.Args[0].Print(p)
			p.WriteByte(' ')
		}
		me.Callee.Print(p)
		for i := 1; i < len(me.Args); i++ {
			p.WriteByte(' ')
			me.Args[i].Print(p)
		}
	case APPLSTYLE_SOV:
		for i := range me.Args {
			me.Args[i].Print(p)
			p.WriteByte(' ')
		}
		me.Callee.Print(p)
	}
	p.WriteByte(')')
}

func (me *AstExprLet) Print(p *CtxPrint) {
	istopleveldeflocal := me == p.CurTopLevel.Ast.Def.Orig.Body
	p.Fmt.OnExprLetBody(istopleveldeflocal, me, me.Body)
	for i := range me.Defs {
		p.WriteByte(',')
		p.Fmt.OnExprLetDef(istopleveldeflocal, me, i, &me.Defs[i])
	}
}

func (me *AstExprCases) Print(p *CtxPrint) {
	p.WriteByte('(')
	me.Scrutinee.Print(p)
	for i := range me.Alts {
		p.WriteString(" | ")
		me.Alts[i].Print(p)
	}
	p.WriteByte(')')
}

func (me *AstCase) Print(p *CtxPrint) {
	for i := range me.Conds {
		if i > 0 {
			p.WriteString(" | ")
		}
		me.Conds[i].Print(p)
	}
	if me.Body != nil {
		p.WriteString(" ? ")
		me.Body.Print(p)
	}
}

type PrintFormatterMinimal struct {
	*CtxPrint
}

func (me *PrintFormatterMinimal) SetCtxPrint(ctxPrint *CtxPrint) { me.CtxPrint = ctxPrint }
func (me *PrintFormatterMinimal) OnTopLevelChunk(tlc *AstFileTopLevelChunk, node *AstTopLevel) {
	me.WriteByte('\n')
	node.Print(me.CtxPrint)
}
func (me *PrintFormatterMinimal) OnDef(_ *AstTopLevel, node *AstDef)  { node.Print(me.CtxPrint) }
func (me *PrintFormatterMinimal) OnDefName(_ *AstDef, node *AstIdent) { node.Print(me.CtxPrint) }
func (me *PrintFormatterMinimal) OnDefArg(_ *AstDef, argIdx int, node *AstDefArg) {
	if me.ApplStyle == APPLSTYLE_VSO || (me.ApplStyle == APPLSTYLE_SVO && argIdx > 0) {
		me.WriteByte(' ')
	}
	node.Print(me.CtxPrint)
	if me.ApplStyle == APPLSTYLE_SOV || (me.ApplStyle == APPLSTYLE_SVO && argIdx == 0) {
		me.WriteByte(' ')
	}
}
func (me *PrintFormatterMinimal) OnDefMeta(_ *AstDef, _ int, node IAstExpr) {
	me.WriteByte(' ')
	node.Print(me.CtxPrint)
}
func (me *PrintFormatterMinimal) OnDefBody(def *AstDef, node IAstExpr) {
	me.WriteString(" := ")
	node.Print(me.CtxPrint)
}
func (me *PrintFormatterMinimal) OnExprLetBody(_ bool, _ *AstExprLet, node IAstExpr) {
	node.Print(me.CtxPrint)
}
func (me *PrintFormatterMinimal) OnExprLetDef(_ bool, _ *AstExprLet, _ int, node *AstDef) {
	node.Print(me.CtxPrint)
}
func (me *PrintFormatterMinimal) OnExprApplName(*AstExprAppl, IAstExpr) {
}
func (me *PrintFormatterMinimal) OnExprApplArg(*AstExprAppl, int, IAstExpr) {

}
