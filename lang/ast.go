package atmolang

import (
	"github.com/go-leap/dev/lex"
	"github.com/go-leap/str"
	"strconv"
)

type IAstNode interface {
	print(*CtxPrint)
	Toks() udevlex.Tokens
}

type IAstComments interface {
	Comments() *astBaseComments
}

type IAstExpr interface {
	IAstNode
	IAstComments
	IsAtomic() bool
}

type IAstExprAtomic interface {
	IAstExpr
	String() string
}

type AstBaseTokens struct {
	Tokens udevlex.Tokens
}

func (me *AstBaseTokens) Toks() udevlex.Tokens { return me.Tokens }

type astBaseComments = struct {
	Leading  AstComments
	Trailing AstComments
}

type AstBaseComments struct {
	comments astBaseComments
}

func (me *AstBaseComments) Comments() *astBaseComments {
	return &me.comments
}

type AstTopLevel struct {
	AstBaseTokens
	AstBaseComments
	Def struct {
		Orig         *AstDef
		IsUnexported bool
	}
}

type AstComments []AstComment

type AstComment struct {
	AstBaseTokens
	Val               string
	IsSelfTerminating bool
}

type AstDef struct {
	AstBaseTokens
	Name       AstIdent
	NameAffix  IAstExpr
	Args       []AstDefArg
	Meta       []IAstExpr
	Body       IAstExpr
	IsTopLevel bool
}

type AstDefArg struct {
	AstBaseTokens
	NameOrConstVal IAstExprAtomic
	Affix          IAstExpr
}

type AstBaseExpr struct {
	AstBaseTokens
	AstBaseComments
}

func (*AstBaseExpr) IsAtomic() bool { return false }

type AstBaseExprAtom struct {
	AstBaseExpr
}

func (*AstBaseExprAtom) IsAtomic() bool { return true }

type AstBaseExprAtomLit struct {
	AstBaseExprAtom
}

type AstExprLitUint struct {
	AstBaseExprAtomLit
	Val uint64
}

func (me *AstExprLitUint) String() string { return strconv.FormatUint(me.Val, 10) }

type AstExprLitFloat struct {
	AstBaseExprAtomLit
	Val float64
}

func (me *AstExprLitFloat) String() string { return strconv.FormatFloat(me.Val, 'g', -1, 64) }

type AstExprLitRune struct {
	AstBaseExprAtomLit
	Val rune
}

func (me *AstExprLitRune) String() string { return strconv.QuoteRune(me.Val) }

type AstExprLitStr struct {
	AstBaseExprAtomLit
	Val string
}

func (me *AstExprLitStr) String() string { return strconv.Quote(me.Val) }

type AstIdent struct {
	AstBaseExprAtom
	Val     string
	IsOpish bool
	IsTag   bool
}

func (me *AstIdent) String() string { return me.Val }

type AstExprAppl struct {
	AstBaseExpr
	Callee IAstExpr
	Args   []IAstExpr
}

type AstExprLet struct {
	AstBaseExpr
	Defs []AstDef
	Body IAstExpr
}

type AstExprCases struct {
	AstBaseExpr
	Scrutinee    IAstExpr
	Alts         []AstCase
	defaultIndex int
}

type AstCase struct {
	AstBaseTokens
	Conds []IAstExpr
	Body  IAstExpr
}

func (me *AstComments) initFrom(accumComments []udevlex.Tokens) {
	this := make(AstComments, len(accumComments))
	for i := range accumComments {
		this[i].initFrom(accumComments[i], 0)
	}
	*me = this
}

func (me *AstComment) initFrom(tokens udevlex.Tokens, at int) {
	me.Tokens = tokens[at : at+1]
	me.Val, me.IsSelfTerminating = me.Tokens[0].Str, me.Tokens[0].IsCommentSelfTerminating()
}

func (me *AstExprCases) Default() *AstCase {
	if me.defaultIndex < 0 {
		return nil
	}
	return &me.Alts[me.defaultIndex]
}

func (me *AstExprCases) seemsUnionSugar() bool {
	return len(me.Alts) == 1 && me.Alts[0].Body == nil
}

func (me *AstExprCases) ToLetIfUnionSugar(prefix int) (let *AstExprLet) {
	if me.Scrutinee == nil && me.seemsUnionSugar() {
		let = &AstExprLet{Defs: make([]AstDef, 1)}
		letdef, letcase := &let.Defs[0], &AstExprCases{defaultIndex: -1, Scrutinee: &AstIdent{Val: "æ"}, Alts: make([]AstCase, 1)}
		let.Body, letcase.Alts[0].Body, letcase.Alts[0].Conds, letdef.Name.Val, letdef.Body, letdef.Args, let.Tokens, letdef.Tokens, letcase.Tokens =
			&letdef.Name, letcase.Scrutinee, me.Alts[0].Conds, ustr.Int(prefix)+"sumt", letcase, []AstDefArg{{NameOrConstVal: letcase.Scrutinee.(IAstExprAtomic)}}, me.Tokens, me.Tokens, me.Tokens
	}
	return
}

func (me *AstExprCases) removeAltAt(idx int) {
	for i := idx; i < len(me.Alts)-1; i++ {
		me.Alts[i] = me.Alts[i+1]
	}
	me.Alts = me.Alts[:len(me.Alts)-1]
}

func (me *AstExprAppl) ClaspishByTokens() (claspish bool) {
	return len(me.Tokens) > 0 && (!me.Tokens.HasSpaces()) && !me.Tokens.HasKind(udevlex.TOKEN_COMMENT)
}

func (me *AstExprAppl) CalleeAndArgsOrdered(applStyle ApplStyle) (ret []IAstExpr) {
	ret = make([]IAstExpr, 1+len(me.Args))
	switch applStyle {
	case APPLSTYLE_VSO:
		for i := range me.Args {
			ret[i+1] = me.Args[i]
		}
		ret[0] = me.Callee
	case APPLSTYLE_SOV:
		for i := range me.Args {
			ret[i] = me.Args[i]
		}
		ret[len(ret)-1] = me.Callee
	case APPLSTYLE_SVO:
		for i := range me.Args {
			ret[i+1] = me.Args[i]
		}
		ret[0], ret[1] = me.Args[0], me.Callee
	}
	return
}

func (me *AstExprAppl) ToUnary() (unary *AstExprAppl) {
	if unary = me; len(me.Args) > 1 {
		appl := *me
		for len(appl.Args) > 1 {
			appl.Callee = &AstExprAppl{Callee: appl.Callee, Args: appl.Args[:1]}
			appl.Args = appl.Args[1:]
		}
		unary = &appl
	}
	return
}

func (me *AstExprAppl) ToLetExprIfUnderscores(prefix int) (let *AstExprLet) {
	lamargs := make(map[IAstExpr]string, 0)
	if ident, _ := me.Callee.(*AstIdent); ident != nil && ustr.IsRepeat(ident.Val, '_') {
		lamargs[me.Callee] = ustr.Int(len(ident.Val)-1) + "æ"
	}
	for i := range me.Args {
		if ident, _ := me.Args[i].(*AstIdent); ident != nil && ustr.IsRepeat(ident.Val, '_') {
			lamargs[me.Args[i]] = ustr.Int(len(ident.Val)-1) + "æ"
		}
	}
	if len(lamargs) > 0 {
		def := AstDef{Name: AstIdent{Val: ustr.Int(prefix) + "lamb"}, Args: make([]AstDefArg, len(lamargs))}
		for i := range def.Args {
			def.Args[i].NameOrConstVal = B.Ident(ustr.Int(i) + "æ")
		}
		var appl AstExprAppl
		appl.Callee, appl.Args = me.Callee, make([]IAstExpr, len(me.Args))
		if la := lamargs[me.Callee]; la != "" {
			appl.Callee = B.Ident(la)
		}
		for i := range appl.Args {
			if la := lamargs[me.Args[i]]; la != "" {
				appl.Args[i] = B.Ident(la)
			} else {
				appl.Args[i] = me.Args[i]
			}
		}

		def.Body = &appl
		let = B.Let(&def.Name, def)
	}
	return
}
