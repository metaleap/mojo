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

type (
	mapTokCmnts   = map[*udevlex.Token][]int
	mapTokOldIdxs = map[*udevlex.Token]int
)

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

func (me *AstFile) parse(this *AstFileTopLevelChunk) {
	toks := this.toks
	this.AstTopLevel.Tokens = toks
	if this.AstTopLevel.Comments, toks = me.parseTopLevelLeadingComments(toks); len(toks) > 0 {
		if def, err := me.parseTopLevelDefinition(toks); err != nil {
			this.errs.parsing = err
		} else if def != nil {
			switch d := def.(type) {
			case nil:
			case *AstDefFunc:
				this.AstTopLevel.DefFunc = d
			case *AstDefType:
				this.AstTopLevel.DefType = d
			default:
				panic("new bug introduced into parseTopLevelDefinition")
			}
		}
	}
}

func (*AstFile) parseTopLevelLeadingComments(toks udevlex.Tokens) (ret []*AstComment, rest udevlex.Tokens) {
	for len(toks) > 0 && toks[0].Kind() == udevlex.TOKEN_COMMENT {
		toks, ret = toks[1:], append(ret, newAstComment(toks, 0))
	}
	rest = toks
	return
}

func (me *AstFile) parseTopLevelDefinition(tokens udevlex.Tokens) (def interface{}, err *Error) {
	var mtokcmnts mapTokCmnts
	var mtokoldidxs mapTokOldIdxs
	toks, mtoks := tokens, tokens.HasKind(udevlex.TOKEN_COMMENT)
	if mtoks {
		mtokcmnts, mtokoldidxs = make(mapTokCmnts), make(mapTokOldIdxs)
		toks = toks.SansComments(mtokcmnts, mtokoldidxs)
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
			if me.Options.ApplStyle == APPLSTYLE_SVO {
				namepos = 1
			} else if me.Options.ApplStyle == APPLSTYLE_SOV {
				namepos = len(headmain) - 1
			}
		}

		var defbase *AstDefBase
		if ustr.BeginsUpper(headmain[namepos].Str) {
			var deftype AstDefType
			def, defbase = &deftype, &deftype.AstDefBase
		} else {
			var deffunc AstDefFunc
			def, defbase = &deffunc, &deffunc.AstDefBase
		}
		defbase.Tokens = tokens
		defbase.newIdent(tokens, headmain, namepos, mtokcmnts, mtokoldidxs)

		for i := range headmain {
			if k, isarg := headmain[i].Kind(), i != namepos; k != udevlex.TOKEN_IDENT && (k != udevlex.TOKEN_OTHER || isarg) {
				def, err = nil, errAt(&headmain[i], "not a valid "+ustr.If(isarg, "argument", "definition")+" name")
				return
			} else if isarg {
				defbase.Args = append(defbase.Args, headmain[i].Str)
			}
		}
	}
	return
}
