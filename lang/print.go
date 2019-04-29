package atmolang

import (
	"strconv"

	"github.com/go-leap/std"
)

type IPrintFormatter interface {
	SetCtxPrint(*CtxPrint)
	BeforeTopLevelChunk(*AstFileTopLevelChunk)
	AfterTopLevelChunk(*AstFileTopLevelChunk)
	OnDefName(*AstDef, IAstNode)
	OnDefArg(*AstDef, int, IAstNode)
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

type PrintFormatterMinimal struct {
	*CtxPrint
}

func (me *PrintFormatterMinimal) SetCtxPrint(ctxPrint *CtxPrint)            { me.CtxPrint = ctxPrint }
func (me *PrintFormatterMinimal) BeforeTopLevelChunk(*AstFileTopLevelChunk) { me.WriteByte('\n') }
func (me *PrintFormatterMinimal) AfterTopLevelChunk(*AstFileTopLevelChunk)  { me.WriteString("\n\n") }
func (me *PrintFormatterMinimal) OnDefName(_ *AstDef, node IAstNode)        { node.print(me.CtxPrint) }
func (me *PrintFormatterMinimal) OnDefArg(_ *AstDef, _ int, node IAstNode)  { node.print(me.CtxPrint) }

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
		OneIndentLevel: "    ", ApplStyle: me.Options.ApplStyle,
		BytesWriter: ustd.BytesWriter{Data: make([]byte, 0, 1024)},
	}
	fmt.SetCtxPrint(&ctx)
	for i := range me.TopLevel {
		ctx.CurTopLevel = &me.TopLevel[i]
		ctx.CurTopLevel.print(&ctx)
	}
	return ctx.BytesWriter.Data
}

func (me *AstFileTopLevelChunk) Print(p *CtxPrint) {
	if !p.fmtCtxSet {
		p.fmtCtxSet = true
		p.Fmt.SetCtxPrint(p)
	}
	me.print(p)
}

func (me *AstFileTopLevelChunk) print(p *CtxPrint) {
	p.CurIndentLevel = 0
	p.Fmt.BeforeTopLevelChunk(me)
	if me.Ast.Def.Orig != nil {
		me.Ast.Def.Orig.print(p)
	}
	p.Fmt.AfterTopLevelChunk(me)
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
			// p.WriteByte(' ')
			p.Fmt.OnDefArg(me, i, &me.Args[i])
		}
	case APPLSTYLE_SVO:
		if len(me.Args) > 0 {
			p.Fmt.OnDefArg(me, 0, &me.Args[0])
			// p.WriteByte(' ')
		}
		p.Fmt.OnDefName(me, &me.Name)
		for i := 1; i < len(me.Args); i++ {
			// p.WriteByte(' ')
			p.Fmt.OnDefArg(me, i, &me.Args[i])
		}
	case APPLSTYLE_SOV:
		for i := range me.Args {
			me.Args[i].print(p)
			p.WriteByte(' ')
		}
		p.Fmt.OnDefName(me, &me.Name)
	}
	for i := range me.Meta {
		p.WriteString(", ")
		me.Meta[i].print(p)
	}
	p.WriteString(" :=")
	if me.IsTopLevel {
		p.CurIndentLevel++
		p.WriteLineBreaksThenIndent(1)
	} else {
		p.WriteByte(' ')
	}
	me.Body.print(p)
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
