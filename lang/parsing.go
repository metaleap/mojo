package odlang

import (
	"text/scanner"

	"github.com/go-leap/dev/lex"
)

func init() {
	udevlex.RestrictedWhitespace, udevlex.StandaloneSeps = true, []string{"(", ")"}
}

type Error struct {
	Msg string
	Pos scanner.Position
}

func errPos(pos *scanner.Position, msg string) *Error {
	return &Error{Pos: *pos, Msg: msg}
}

func errTok(tok *udevlex.Token, msg string) *Error {
	return errPos(&tok.Meta.Position, msg)
}

func (this *Error) Error() string {
	return this.Pos.String() + ": " + this.Msg
}

func (me *AstFile) parse(tlcIdx int) {
	this := &me.topLevelChunks[tlcIdx]
	toks := this.toks
	this.topLevel.toks = toks
	if this.topLevel.comments, toks = parseTopLevelLeadingComments(toks); len(toks) > 0 {
		if def, err := parseTopLevelDefinition(toks); err != nil {
			this.errs.parsing = err
		} else if def != nil {
			switch d := def.(type) {
			case *AstDefFunc:
				this.topLevel.defFunc = d
			case *AstDefType:
				this.topLevel.defType = d
			}
		}
	}
}

func parseTopLevelLeadingComments(toks udevlex.Tokens) (ret []*AstComment, rest udevlex.Tokens) {
	for len(toks) > 0 && toks[0].Kind() == udevlex.TOKEN_COMMENT {
		comment := AstComment{Text: toks[0].Str, SelfTerminates: toks[0].IsCommentSelfTerminating()}
		comment.toks = toks[0:1]
		toks, ret = toks[1:], append(ret, &comment)
	}
	rest = toks
	return
}

func parseTopLevelDefinition(toks udevlex.Tokens) (def IAstNode, err *Error) {
	head, body := toks.BreakOnOther(":=")
	if len(body) == 0 {
		err = errTok(&toks[0], "missing: definition body following `:=`")
	} else if len(head) == 0 {
		err = errTok(&toks[0], "missing: definition name preceding `:=`")
	}
	return
}
