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
	for i := range this.toks {
		tok := &this.toks[i]
		if tok.Kind() == udevlex.TOKEN_COMMENT {
			node := &AstComment{Text: tok.Str, SelfTerminates: tok.IsCommentSelfTerminating()}
			node.astNode.toks = this.toks[i : i+1]
			this.topLevel.nodes = append(this.topLevel.nodes, node)
		}
	}
}
