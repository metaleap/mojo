package atmolang

import (
	"strconv"

	"github.com/go-leap/std"
)

type CtxPrint struct {
	IPrintFormatter
	File           *AstFile
	CurTopLevel    *AstTopLevel
	CurIndentLevel int

	ustd.BytesWriter
}

type IPrintFormatter interface {
}

type PrintFormatterBase struct {
}

type PrintFormatterMinimal struct {
	PrintFormatterBase
}

func (me *CtxPrint) writeNewLines(times int) {
	for i := 0; i < times; i++ {
		me.WriteByte('\n')
	}
	for i := 0; i < me.CurIndentLevel; i++ {
		me.WriteString("    ")
	}
}

func (me *AstFile) Print(pf IPrintFormatter) []byte {
	ctx := CtxPrint{File: me, IPrintFormatter: pf, BytesWriter: ustd.BytesWriter{Data: make([]byte, 0, 1024)}}
	for i := range me.TopLevel {
		ctx.CurTopLevel = &me.TopLevel[i].Ast
		ctx.CurTopLevel.print(&ctx)
	}
	return ctx.BytesWriter.Data
}

func (me *AstTopLevel) print(p *CtxPrint) {
	p.CurIndentLevel = 0
	p.WriteByte('\n')
	for i := range me.Comments {
		me.Comments[i].print(p)
	}
	if me.Def != nil {
		me.Def.print(p)
	}
	p.WriteString("\n\n")
}

func (me *AstComment) print(p *CtxPrint) {
	if me.IsSelfTerminating {
		p.WriteString("/*")
		p.WriteString(me.ContentText)
		p.WriteString("*/")
	} else {
		p.WriteString("//")
		p.WriteString(me.ContentText)
		p.writeNewLines(1)
	}
}

func (me *AstDef) print(p *CtxPrint) {
	switch p.File.Options.ApplStyle {
	case APPLSTYLE_VSO:
		me.Name.print(p)
		for i := range me.Args {
			p.WriteByte(' ')
			me.Args[i].print(p)
		}
	case APPLSTYLE_SVO:
		if len(me.Args) > 0 {
			me.Args[0].print(p)
			p.WriteByte(' ')
		}
		me.Name.print(p)
		for i := 1; i < len(me.Args); i++ {
			p.WriteByte(' ')
			me.Args[i].print(p)
		}
	case APPLSTYLE_SOV:
		for i := range me.Args {
			me.Args[i].print(p)
			p.WriteByte(' ')
		}
		me.Name.print(p)
	}
	for i := range me.Meta {
		p.WriteString(", ")
		me.Meta[i].print(p)
	}
	p.WriteString(" :=")
	if me.IsTopLevel {
		p.CurIndentLevel++
		p.writeNewLines(1)
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
	for i := range me.Comments {
		p.WriteByte(' ')
		me.Comments[i].print(p)
	}
}

func (me *AstExprLitFloat) print(p *CtxPrint) {
	p.WriteString(strconv.FormatFloat(me.Val, 'g', -1, 64))
	for i := range me.Comments {
		p.WriteByte(' ')
		me.Comments[i].print(p)
	}
}

func (me *AstExprLitUint) print(p *CtxPrint) {
	p.WriteString(strconv.FormatUint(me.Val, 10))
	for i := range me.Comments {
		p.WriteByte(' ')
		me.Comments[i].print(p)
	}
}

func (me *AstExprLitRune) print(p *CtxPrint) {
	p.WriteString(strconv.QuoteRune(me.Val))
	for i := range me.Comments {
		p.WriteByte(' ')
		me.Comments[i].print(p)
	}
}

func (me *AstExprLitStr) print(p *CtxPrint) {
	p.WriteString(strconv.Quote(me.Val))
	for i := range me.Comments {
		p.WriteByte(' ')
		me.Comments[i].print(p)
	}
}

func (me *AstExprAppl) print(p *CtxPrint) {
	p.WriteByte('(')
	switch p.File.Options.ApplStyle {
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

func (me *AstExprCase) print(p *CtxPrint) {
	p.WriteByte('(')
	me.Scrutinee.print(p)
	for i := range me.Alts {
		p.WriteString(" | ")
		me.Alts[i].print(p)
	}
	p.WriteByte(')')
}

func (me *AstCaseAlt) print(p *CtxPrint) {
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
