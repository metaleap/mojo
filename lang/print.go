package atmolang

import (
	"strconv"

	"github.com/go-leap/std"
)

type IPrintFormatter interface {
	SetCtxPrint(*CtxPrint)
	DoesNewlineBeforeTopLevelDefBody() bool
	OnTopLevelChunk(*AstFileTopLevelChunk, *AstTopLevel)
	OnDef(*AstTopLevel, *AstDef)
	OnDefName(*AstDef, *AstIdent)
	OnDefArg(*AstDef, int, *AstDefArg)
	OnDefMeta(*AstDef, int, IAstExpr)
	OnDefBody(*AstDef, IAstExpr)
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
		p.WriteLineBreaksThenIndent(1)
	}
}

func (me *AstDef) print(p *CtxPrint) {
	switch p.ApplStyle {
	case APPLSTYLE_VSO:
		p.Fmt.OnDefName(me, &me.Name)
		for i := range me.Args {
			p.Fmt.OnDefArg(me, i, &me.Args[i])
		}
	case APPLSTYLE_SVO:
		if len(me.Args) > 0 {
			p.Fmt.OnDefArg(me, 0, &me.Args[0])
		}
		p.Fmt.OnDefName(me, &me.Name)
		for i := 1; i < len(me.Args); i++ {
			p.Fmt.OnDefArg(me, i, &me.Args[i])
		}
	case APPLSTYLE_SOV:
		for i := range me.Args {
			p.Fmt.OnDefArg(me, i, &me.Args[i])
		}
		p.Fmt.OnDefName(me, &me.Name)
	}
	for i := range me.Meta {
		p.WriteByte(',')
		p.Fmt.OnDefMeta(me, i, me.Meta[i])
	}
	if p.WriteString(" :="); !p.Fmt.DoesNewlineBeforeTopLevelDefBody() {
		p.WriteByte(' ')
	}
	p.Fmt.OnDefBody(me, me.Body)
}

func (me *AstDefArg) print(p *CtxPrint) {
	if me.NameOrConstVal.print(p); me.Affix != nil {
		p.WriteByte(':')
		me.Affix.print(p)
	}
}

func (me *AstIdent) print(p *CtxPrint) {
	p.WriteString(me.Val)
}

func (me *AstExprLitFloat) print(p *CtxPrint) {
	p.WriteString(strconv.FormatFloat(me.Val, 'g', -1, 64))
}

func (me *AstExprLitUint) print(p *CtxPrint) {
	p.WriteString(strconv.FormatUint(me.Val, 10))
}

func (me *AstExprLitRune) print(p *CtxPrint) {
	p.WriteString(strconv.QuoteRune(me.Val))
}

func (me *AstExprLitStr) print(p *CtxPrint) {
	p.WriteString(strconv.Quote(me.Val))
}

func (me *AstExprAppl) print(p *CtxPrint) {
	p.WriteByte('(')
	switch p.ApplStyle {
	case APPLSTYLE_VSO:
		me.Callee.print(p)
		for i := range me.Args {
			p.WriteByte(' ')
			me.Args[i].print(p)
		}
	case APPLSTYLE_SVO:
		if len(me.Args) > 0 {
			me.Args[0].print(p)
			p.WriteByte(' ')
		}
		me.Callee.print(p)
		for i := 1; i < len(me.Args); i++ {
			p.WriteByte(' ')
			me.Args[i].print(p)
		}
	case APPLSTYLE_SOV:
		for i := range me.Args {
			me.Args[i].print(p)
			p.WriteByte(' ')
		}
		me.Callee.print(p)
	}
	p.WriteByte(')')
}

func (me *AstExprLet) print(p *CtxPrint) {
	me.Body.print(p)
	for i := range me.Defs {
		p.WriteString(", ")
		me.Defs[i].print(p)
	}
}

func (me *AstExprCases) print(p *CtxPrint) {
	p.WriteByte('(')
	me.Scrutinee.print(p)
	for i := range me.Alts {
		p.WriteString(" | ")
		me.Alts[i].print(p)
	}
	p.WriteByte(')')
}

func (me *AstCase) print(p *CtxPrint) {
	for i := range me.Conds {
		if i > 0 {
			p.WriteString(" | ")
		}
		me.Conds[i].print(p)
	}
	if me.Body != nil {
		p.WriteString(" ? ")
		me.Body.print(p)
	}
}

type PrintFormatterMinimal struct {
	*CtxPrint
}

func (me *PrintFormatterMinimal) SetCtxPrint(ctxPrint *CtxPrint)         { me.CtxPrint = ctxPrint }
func (me *PrintFormatterMinimal) DoesNewlineBeforeTopLevelDefBody() bool { return false }
func (me *PrintFormatterMinimal) OnTopLevelChunk(tlc *AstFileTopLevelChunk, node *AstTopLevel) {
	me.WriteByte('\n')
	node.print(me.CtxPrint)
}
func (me *PrintFormatterMinimal) OnDef(_ *AstTopLevel, node *AstDef)  { node.print(me.CtxPrint) }
func (me *PrintFormatterMinimal) OnDefName(_ *AstDef, node *AstIdent) { node.print(me.CtxPrint) }
func (me *PrintFormatterMinimal) OnDefArg(_ *AstDef, argIdx int, node *AstDefArg) {
	if me.ApplStyle == APPLSTYLE_VSO || (me.ApplStyle == APPLSTYLE_SVO && argIdx > 0) {
		me.WriteByte(' ')
	}
	node.print(me.CtxPrint)
	if me.ApplStyle == APPLSTYLE_SOV || (me.ApplStyle == APPLSTYLE_SVO && argIdx == 0) {
		me.WriteByte(' ')
	}
}
func (me *PrintFormatterMinimal) OnDefMeta(_ *AstDef, _ int, node IAstExpr) {
	me.WriteByte(' ')
	node.print(me.CtxPrint)
}
func (me *PrintFormatterMinimal) OnDefBody(def *AstDef, node IAstExpr) {
	// if def.IsTopLevel {
	// 	me.CurIndentLevel++
	// 	me.WriteLineBreaksThenIndent(1)
	// }
	node.print(me.CtxPrint)
}
