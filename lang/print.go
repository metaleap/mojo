package odlang

import (
	"io"
)

type IPrintFormatter interface {
	io.StringWriter
}

type PrintFormatterBase struct {
}

type PrintFormatterMinimal struct {
	PrintFormatterBase
	io.StringWriter
}

func (me *AstTopLevel) print(pf IPrintFormatter) (err error) {
	pf.WriteString("\n")
	for i := range me.Comments {
		if me.Comments[i].IsSelfTerminating {
			pf.WriteString("/*")
			pf.WriteString(me.Comments[i].ContentText)
			pf.WriteString("*/")
		} else {
			pf.WriteString("//")
			pf.WriteString(me.Comments[i].ContentText)
			pf.WriteString("\n")
		}
	}

	return
}

func (me *AstDefType) print(pf IPrintFormatter) (err error) {
	return
}

func (me *AstDefFunc) print(pf IPrintFormatter) (err error) {
	return
}
