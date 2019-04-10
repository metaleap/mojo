package odlang

import (
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

func (me *AstFile) Print(pf IPrintFormatter) (formattedSrc []byte, err error) {
	ctx := CtxPrint{File: me, IPrintFormatter: pf, BytesWriter: ustd.BytesWriter{Data: make([]byte, 0, 1024)}}
	for i := range me.TopLevel {
		ctx.CurTopLevel = &me.TopLevel[i].Ast
		if err = ctx.CurTopLevel.print(&ctx); err != nil {
			return
		}
	}
	if err == nil {
		formattedSrc = ctx.BytesWriter.Data
	}
	return
}

func (me *AstTopLevel) print(p *CtxPrint) (err error) {
	p.WriteString("\n")
	for i := range me.Comments {
		if me.Comments[i].IsSelfTerminating {
			p.WriteString("/*")
			p.WriteString(me.Comments[i].ContentText)
			p.WriteString("*/")
		} else {
			p.WriteString("//")
			p.WriteString(me.Comments[i].ContentText)
			p.WriteString("\n")
		}
	}

	return
}

func (me *AstDefType) print(p *CtxPrint) (err error) {
	return
}

func (me *AstDefFunc) print(p *CtxPrint) (err error) {
	switch p.File.Options.ApplStyle {
	case APPLSTYLE_VSO:

	case APPLSTYLE_SVO:
	case APPLSTYLE_SOV:
	}
	return
}
