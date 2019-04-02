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
	ctxTopLevelDef struct {
		file *AstFile
		def  IAstDef
		mtc  map[*udevlex.Token][]int
		mto  map[*udevlex.Token]int
	}
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
	toks := this.Tokens
	if this.AstTopLevel.Comments, toks = me.parseTopLevelLeadingComments(toks); len(toks) > 0 {
		this.AstTopLevel.Def, this.errs.parsing = me.parseTopLevelDef(toks)
	}
}

func (*AstFile) parseTopLevelLeadingComments(toks udevlex.Tokens) (ret []*AstComment, rest udevlex.Tokens) {
	for len(toks) > 0 && toks[0].Kind() == udevlex.TOKEN_COMMENT {
		toks, ret = toks[1:], append(ret, newAstComment(toks, 0))
	}
	rest = toks
	return
}

func (me *AstFile) parseTopLevelDef(tokens udevlex.Tokens) (def IAstDef, err *Error) {
	ctx := ctxTopLevelDef{file: me, mtc: make(map[*udevlex.Token][]int), mto: make(map[*udevlex.Token]int)}
	return ctx.parseDef(tokens, true)
}

func (me *ctxTopLevelDef) parseDef(tokens udevlex.Tokens, topLevel bool) (def IAstDef, err *Error) {
	toks := tokens
	if topLevel {
		toks = tokens.SansComments(me.mtc, me.mto)
	}
	tokshead, toksbody := toks.BreakOnOther(":=")
	if len(toksbody) == 0 {
		err = errAt(&tokens[0], "missing: definition body following `:=`")
	} else if len(tokshead) == 0 {
		err = errAt(&tokens[0], "missing: definition name preceding `:=`")
	} else if toksheadsig, _ := tokshead.BreakOnOther(","); len(toksheadsig) == 0 {
		err = errAt(&tokens[0], "missing: definition name preceding `,`")
	} else {
		var namepos int
		if len(toksheadsig) > 1 {
			if me.file.Options.ApplStyle == APPLSTYLE_SVO {
				namepos = 1
			} else if me.file.Options.ApplStyle == APPLSTYLE_SOV {
				namepos = len(toksheadsig) - 1
			}
		}

		var defbase *AstDefBase
		if isdeftype := ustr.BeginsUpper(toksheadsig[namepos].Str); isdeftype {
			var deftype AstDefType
			def, defbase, deftype.AstDefBase.IsDefType = &deftype, &deftype.AstDefBase, true
		} else {
			var deffunc AstDefFunc
			def, defbase = &deffunc, &deffunc.AstDefBase
		}
		if topLevel {
			me.def = def
		}
		defbase.Tokens = tokens
		if err = defbase.newIdent(me, -1, toksheadsig, namepos); err != nil {
			def = nil
		} else {
			defbase.ensureArgsLen(len(toksheadsig) - 1)
			for i, a := 0, 0; i < len(toksheadsig); i++ {
				if i != namepos {
					if err = defbase.newIdent(me, a, toksheadsig, i); err != nil {
						def = nil
						return
					}
					a++
				}
			}
			if err = def.parseDefBody(me, toksbody); err != nil {
				def = nil
			}
		}
	}
	return
}

func (me *AstDefFunc) parseDefBody(ctx *ctxTopLevelDef, toks udevlex.Tokens) (err *Error) {
	me.Body, err = ctx.parseExpr(toks)
	return
}

func (me *AstDefType) parseDefBody(ctx *ctxTopLevelDef, toks udevlex.Tokens) *Error {
	return nil
}

func (me *ctxTopLevelDef) parseExpr(toks udevlex.Tokens) (ret IAstExpr, err *Error) {
	if len(toks) == 0 {
		panic("bug in parseExpr")
	}

	if chunks := toks.IndentBasedChunks(toks[0].Meta.Position.Column - 1); len(chunks) > 1 {
		ret, err = me.parseExprLetOuter(toks, chunks)

	} else {
		alltoks, accum := toks, make([]IAstExpr, 0, len(toks))
		for len(toks) > 0 {
			var exprcur IAstExpr
			switch k := toks[0].Kind(); k {
			case udevlex.TOKEN_FLOAT:
				exprcur = me.newExprLitFloat(toks)
				toks = toks[1:]
			case udevlex.TOKEN_UINT:
				exprcur = me.newExprLitUint(toks)
				toks = toks[1:]
			case udevlex.TOKEN_RUNE:
				exprcur = me.newExprLitRune(toks)
				toks = toks[1:]
			case udevlex.TOKEN_STR:
				exprcur = me.newExprLitStr(toks)
				toks = toks[1:]
			case udevlex.TOKEN_IDENT, udevlex.TOKEN_OTHER:
				switch toks[0].Str {
				case ",":
					exprcur, err = me.parseExprLetInner(toks, accum, alltoks)
					accum, toks = accum[0:0], nil
				// case "?":
				default:
					exprcur = me.newExprIdent(toks)
					toks = toks[1:]
				}
			case udevlex.TOKEN_SEP:
				if toks[0].Str == ")" {
					err = errAt(&toks[0], "closing parenthesis without matching opening")
				} else if sub, tail, numunclosed := toks.Sub("(", ")"); len(sub) == 0 {
					if numunclosed == 0 {
						err = errAt(&toks[0], "empty parentheses")
					} else {
						err = errAt(&toks[0], "unclosed parenthesis")
					}
				} else if exprcur, err = me.parseExpr(sub); err == nil {
					toks = tail
				}
			default:
				panic(k)
			}
			if err != nil {
				return
			}
			accum = append(accum, exprcur)
		}
		ret, err = me.parseExprFinalize(accum, alltoks)
	}
	return
}

func (me *ctxTopLevelDef) parseExprFinalize(accum []IAstExpr, allToks udevlex.Tokens) (ret IAstExpr, err *Error) {
	if len(accum) == 1 {
		ret = accum[0]
	} else {
		var call AstExprCall
		me.setTokensFor(&call.AstBase, allToks)
		l := len(accum) - 1
		switch me.file.Options.ApplStyle {
		case APPLSTYLE_SVO:
			call.Callee = accum[1]
			call.Args = append(accum[0:1], accum[2:]...)
		case APPLSTYLE_VSO:
			call.Callee = accum[0]
			call.Args = accum[1:]
		case APPLSTYLE_SOV:
			call.Callee = accum[l]
			call.Args = accum[:l]
		}
		ret = &call
	}
	return
}

func (me *ctxTopLevelDef) parseExprLetInner(toks udevlex.Tokens, accum []IAstExpr, allToks udevlex.Tokens) (ret IAstExpr, err *Error) {
	if ret, err = me.parseExprFinalize(accum, allToks); err == nil {
		if chunks := toks.Chunked(",", "(", ")"); len(chunks) > 0 {
			var let AstExprLet
			let.Body, let.Defs, ret = ret, make([]IAstDef, 0, len(chunks)), nil
			me.setTokensFor(&let.AstBase, allToks)
			var def IAstDef
			for i := range chunks {
				if def, err = me.parseDef(chunks[i], false); err != nil {
					return
				} else {
					let.Defs = append(let.Defs, def)
				}
			}
			ret = &let
		}
	}
	return
}

func (me *ctxTopLevelDef) parseExprLetOuter(toks udevlex.Tokens, toksChunked []udevlex.Tokens) (ret *AstExprLet, err *Error) {
	var let AstExprLet
	var def IAstDef
	me.setTokensFor(&let.AstBase, toks)
	for i := range toksChunked {
		if i == 0 {
			let.Body, err = me.parseExpr(toksChunked[i])
		} else if def, err = me.parseDef(toksChunked[i], false); err == nil {
			let.Defs = append(let.Defs, def)
		}
		if err != nil {
			return
		}
	}
	ret = &let
	return
}
