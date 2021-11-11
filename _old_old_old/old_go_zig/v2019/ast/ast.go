package atmoast

import (
	"strconv"

	"github.com/go-leap/dev/lex"
	"github.com/go-leap/str"
)

func (me *AstBaseTokens) at(self IAstNode, pos int) []IAstNode {
	if me.Tokens.AreEnclosing(pos) {
		return []IAstNode{self}
	}
	return nil
}
func (me *AstBaseTokens) Toks() udevlex.Tokens { return me.Tokens }

func (me *AstBaseComments) Comments() *astBaseComments {
	return &me.comments
}

func (me *AstTopLevel) at(_ IAstNode, pos int) (nodes []IAstNode) {
	if me.Tokens.AreEnclosing(pos) {
		if me.Def.Orig != nil {
			nodes = me.Def.Orig.at(me.Def.Orig, pos)
		}
		nodes = append(nodes, me)
	}
	return
}

func (me *AstDef) at(_ IAstNode, pos int) (nodes []IAstNode) {
	if me.Tokens.AreEnclosing(pos) {
		nodes = me.Name.at(&me.Name, pos)
		if len(nodes) == 0 && me.NameAffix != nil {
			nodes = me.NameAffix.at(me.NameAffix, pos)
		}
		if len(nodes) == 0 && me.Body != nil {
			nodes = me.Body.at(me.Body, pos)
		}
		if len(nodes) == 0 {
			for i := range me.Args {
				if nodes = me.Args[i].at(&me.Args[i], pos); len(nodes) != 0 {
					break
				}
			}
		}
		if len(nodes) == 0 {
			for i := range me.Meta {
				if nodes = me.Meta[i].at(me.Meta[i], pos); len(nodes) != 0 {
					break
				}
			}
		}
		nodes = append(nodes, me)
	}
	return
}

func (me *AstDefArg) at(_ IAstNode, pos int) (nodes []IAstNode) {
	if me.Tokens.AreEnclosing(pos) {
		if me.NameOrConstVal != nil {
			nodes = me.NameOrConstVal.at(me.NameOrConstVal, pos)
		}
		if len(nodes) == 0 && me.Affix != nil {
			nodes = me.Affix.at(me.Affix, pos)
		}
		nodes = append(nodes, me)
	}
	return
}

func (*AstBaseExpr) IsAtomic() bool { return false }

func (*AstBaseExprAtom) IsAtomic() bool { return true }

func (me *AstExprLitUint) FromRune() bool {
	return len(me.Tokens) == 1 && len(me.Tokens[0].Lexeme) != 0 && me.Tokens[0].Lexeme[0] == '\''
}

func (me *AstExprLitUint) String() string { return strconv.FormatUint(me.Val, 10) }

func (me *AstExprLitFloat) String() string { return strconv.FormatFloat(me.Val, 'f', -1, 64) }

func (me *AstExprLitStr) String() string { return strconv.Quote(me.Val) }

func (me *AstIdent) IsName(opishOk bool) bool {
	return ((!me.IsOpish) || opishOk) && (!me.IsTag) && me.Val[0] != '_'
}
func (me *AstIdent) IsVar() bool {
	return len(me.Val) > 1 && me.Val[0] == '_' && me.Val[1] != '_'
}
func (me *AstIdent) IsPlaceholder() bool { return ustr.IsRepeat(me.Val, '_') }
func (me *AstIdent) String() string      { return me.Val }

func (me *AstExprAppl) at(_ IAstNode, pos int) (nodes []IAstNode) {
	if me.Tokens.AreEnclosing(pos) {
		if me.Callee != nil {
			nodes = me.Callee.at(me.Callee, pos)
		}
		if len(nodes) == 0 {
			for _, arg := range me.Args {
				if nodes = arg.at(arg, pos); len(nodes) != 0 {
					break
				}
			}
		}
		nodes = append(nodes, me)
	}
	return
}

func (me *AstExprLet) at(_ IAstNode, pos int) (nodes []IAstNode) {
	if me.Tokens.AreEnclosing(pos) {
		if me.Body != nil {
			nodes = me.Body.at(me.Body, pos)
		}
		if len(nodes) == 0 {
			for i := range me.Defs {
				if nodes = me.Defs[i].at(&me.Defs[i], pos); len(nodes) != 0 {
					break
				}
			}
		}
		nodes = append(nodes, me)
	}
	return
}

func (me *AstExprCases) at(_ IAstNode, pos int) (nodes []IAstNode) {
	if me.Tokens.AreEnclosing(pos) {
		if me.Scrutinee != nil {
			nodes = me.Scrutinee.at(me.Scrutinee, pos)
		}
		if len(nodes) == 0 {
			for i := range me.Alts {
				if nodes = me.Alts[i].at(&me.Alts[i], pos); len(nodes) != 0 {
					break
				}
			}
		}
		nodes = append(nodes, me)
	}
	return
}

func (me *AstCase) at(_ IAstNode, pos int) (nodes []IAstNode) {
	if me.Tokens.AreEnclosing(pos) {
		if me.Body != nil {
			nodes = me.Body.at(me.Body, pos)
		}
		if len(nodes) == 0 {
			for _, c := range me.Conds {
				if nodes = c.at(c, pos); len(nodes) != 0 {
					break
				}
			}
		}
		nodes = append(nodes, me)
	}
	return
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
	me.Val, me.IsLineComment = me.Tokens[0].Val.(string), me.Tokens[0].IsLineComment()
}

func (me *AstExprCases) Default() *AstCase {
	if me.defaultIndex < 0 {
		return nil
	}
	return &me.Alts[me.defaultIndex]
}

func (me *AstExprCases) removeAltAt(idx int) {
	for i := idx; i < len(me.Alts)-1; i++ {
		me.Alts[i] = me.Alts[i+1]
	}
	me.Alts = me.Alts[:len(me.Alts)-1]
}

func (me *AstExprAppl) ClaspishByTokens() (claspish bool) {
	return len(me.Tokens) != 0 && (!me.Tokens.HasSpaces()) && !me.Tokens.HasKind(udevlex.TOKEN_COMMENT)
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

func (me *AstDef) ensureUnary(origName string) {
	subname, let := ustr.Int(len(me.Args)-1)+origName, AstExprLet{Defs: []AstDef{{Body: me.Body, AstBaseTokens: me.AstBaseTokens}}}
	subdef := &let.Defs[0]
	let.AstBaseTokens, let.comments, me.Body, let.Body, me.Args, subdef.Args, subdef.Name.Val =
		me.AstBaseTokens, *me.Body.Comments(), &let, &subdef.Name, me.Args[:1], me.Args[1:], subname
	if len(subdef.Args) > 1 {
		subdef.ensureUnary(origName)
	}
}

func (me *AstDef) ToUnary() *AstDef {
	if len(me.Args) <= 1 {
		return me
	}
	def := *me
	def.ensureUnary(me.Name.Val)
	return &def
}
