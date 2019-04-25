package atmocorefn

import (
	"github.com/go-leap/str"
)

type AstBuilder struct{}

func (*AstBuilder) Appl(callee IAstIdent, arg IAstExprAtomic) *AstAppl {
	return &AstAppl{Callee: callee, Arg: arg}
}

func (*AstBuilder) Cases() *AstCases {
	return &AstCases{}
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
