package odlang

import (
	"text/scanner"

	"github.com/go-leap/dev/lex"
	"github.com/go-leap/str"
)

type ApplStyle int

const (
	APPLSTYLE_SVO ApplStyle = iota
	APPLSTYLE_VSO
	APPLSTYLE_SOV
)

var Config struct {
	ApplStyle ApplStyle
}

func init() {
	udevlex.RestrictedWhitespace, udevlex.StandaloneSeps = true, []string{"(", ")"}
}

type Error struct {
	Msg string
	Pos scanner.Position
	Len int
}

func errAt(tok *udevlex.Token, msg string) *Error {
	return &Error{Msg: msg, Pos: tok.Meta.Position, Len: len(tok.Meta.Orig)}
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
		comment := AstComment{ContentText: toks[0].Str, SelfTerminates: toks[0].IsCommentSelfTerminating()}
		comment.toks = toks[0:1]
		toks, ret = toks[1:], append(ret, &comment)
	}
	rest = toks
	return
}

func parseTopLevelDefinition(tokens udevlex.Tokens) (def IAstNode, err *Error) {
	var comments map[*udevlex.Token][]int
	toks := tokens
	if toks.HasKind(udevlex.TOKEN_COMMENT) {
		comments = make(map[*udevlex.Token][]int)
		toks = toks.SansComments(comments)
	}
	head, body := toks.BreakOnOther(":=")
	if len(body) == 0 {
		err = errAt(&tokens[0], "missing: definition body following `:=`")
	} else if len(head) == 0 {
		err = errAt(&tokens[0], "missing: definition name preceding `:=`")
	} else if headmain, _ := head.BreakOnOther(","); len(headmain) == 0 {
		err = errAt(&tokens[0], "missing: definition name preceding `,`")
	} else {
		var namepos int
		if len(headmain) > 1 {
			if Config.ApplStyle == APPLSTYLE_SVO {
				namepos = 1
			} else if Config.ApplStyle == APPLSTYLE_SOV {
				namepos = len(headmain) - 1
			}
		}
		defbase := astDefBase{Name: headmain[namepos].Str}
		for i := range headmain {
			if k, isarg := headmain[i].Kind(), i != namepos; k != udevlex.TOKEN_IDENT && (k != udevlex.TOKEN_OTHER || isarg) {
				err = errAt(&headmain[i], "not a valid "+ustr.If(isarg, "argument", "definition")+" name")
				return
			} else if isarg {
				defbase.Args = append(defbase.Args, headmain[i].Str)
			}
		}
		if ustr.BeginsUpper(defbase.Name) {
			def = &AstDefType{astDefBase: defbase, astNode: astNode{toks: tokens}}
		} else {
			def = &AstDefFunc{astDefBase: defbase, astNode: astNode{toks: tokens}}
		}
	}
	return
}
