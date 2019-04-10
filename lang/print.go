package odlang

import (
	"strconv"

	"github.com/go-leap/std"
)

type CtxPrint struct {
	IPrintFormatter
	File        *AstFile
	CurTopLevel *AstTopLevel

	ustd.BytesWriter
}

type IPrintFormatter interface {
}

type PrintFormatterBase struct {
}

type PrintFormatterMinimal struct {
	PrintFormatterBase
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
		p.WriteString("\n")
	}
}

func (me *AstDefType) print(p *CtxPrint) {
}

func (me *AstDefFunc) print(p *CtxPrint) {
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
	p.WriteString(" :=\n    ")
	me.Body.print(p)
}

func (me *astIdent) print(p *CtxPrint)        { p.WriteString(me.Val) }
func (me *AstExprLitFloat) print(p *CtxPrint) { p.WriteString(strconv.FormatFloat(me.Val, 'g', -1, 64)) }
func (me *AstExprLitUint) print(p *CtxPrint)  { p.WriteString(strconv.FormatUint(me.Val, 10)) }
func (me *AstExprLitRune) print(p *CtxPrint)  { p.WriteString(strconv.QuoteRune(me.Val)) }
func (me *AstExprLitStr) print(p *CtxPrint)   { p.WriteString(strconv.Quote(me.Val)) }

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
	me.Scrutinee.print(p)
	for i := range me.Alts {
		if i == 0 {
			p.WriteString(" ? ")
		} else {
			p.WriteString(" | ")
		}
		me.Alts[i].print(p)
	}
}

func (me *AstCaseAlt) print(p *CtxPrint) {
	for i := range me.Conds {
		if i > 0 {
			p.WriteString(" | ")
		}
		me.Conds[i].print(p)
	}
	p.WriteString(" : ")
	me.Body.print(p)
}

func (me *AstTypeExprAppl) print(p *CtxPrint) {}
