package atmocorefn

import (
	"github.com/go-leap/str"
)

type AstBuilder struct{}

func (*AstBuilder) Appl(callee IAstIdent, arg IAstExprAtomic) *AstAppl {
	return &AstAppl{Callee: callee, Arg: arg}
}

func (*AstBuilder) Appls(ctx *AstDef, callee IAstIdent, args ...IAstExprAtomic) *AstAppl {
	var appl AstAppl
	for i := range args {
		if i == 0 {
			appl.Callee, appl.Arg = callee, args[i]
		} else {
			appl = AstAppl{Callee: ctx.ensureAstAtomFor(&appl, true, "__appl_arg_"+ustr.Int(i)+"__").(IAstIdent), Arg: args[i]}
		}
	}
	return &appl
}

func (*AstBuilder) Case(ifThis IAstExpr, thenThat IAstExpr) *AstCases {
	return &AstCases{Ifs: [][]IAstExpr{{ifThis}}, Thens: []IAstExpr{thenThat}}
}

func (*AstBuilder) IdName(name string) *AstIdentName {
	return &AstIdentName{AstIdentBase{Val: name}}
}

func (*AstBuilder) IdEmptyParens() *AstIdentEmptyParens {
	return &AstIdentEmptyParens{AstIdentBase{Val: "()"}}
}

func (*AstBuilder) IdOp(name string) *AstIdentOp {
	return &AstIdentOp{AstIdentBase{Val: name}}
}

func (*AstBuilder) IdVar(name string) *AstIdentVar {
	return &AstIdentVar{AstIdentBase{Val: name}}
}

func (*AstBuilder) IdUnderscores(num int) *AstIdentUnderscores {
	return &AstIdentUnderscores{AstIdentBase{Val: ustr.Times("_", num)}}
}

func (*AstBuilder) IdTag(name string) *AstIdentTag {
	return &AstIdentTag{AstIdentBase{Val: name}}
}

func (*AstBuilder) DefArgs(names ...string) (r []AstDefArg) {
	r = make([]AstDefArg, len(names))
	for i := range names {
		r[i].AstIdentName.Val = names[i]
	}
	return
}
