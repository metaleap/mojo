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

func (me *AstTopLevel) print(ctx *CtxPrint) (err error) {
	ctx.WriteString("\n")
	for i := range me.Comments {
		if me.Comments[i].IsSelfTerminating {
			ctx.WriteString("/*")
			ctx.WriteString(me.Comments[i].ContentText)
			ctx.WriteString("*/")
		} else {
			ctx.WriteString("//")
			ctx.WriteString(me.Comments[i].ContentText)
			ctx.WriteString("\n")
		}
	}

	return
}

func (me *AstDefType) print(ctx *CtxPrint) (err error) {
	return
}

func (me *AstDefFunc) print(ctx *CtxPrint) (err error) {
	return
}
