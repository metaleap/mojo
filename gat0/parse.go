package main

func parse(toks Tokens, origSrc string, srcFilePath string) (ret AstFile) {
	ret.toks, ret.srcFilePath, ret.origSrc = toks, srcFilePath, origSrc
	toplevel := toks.indentLevelChunks(0)
	for _, tlchunk := range toplevel {
		nodes := ret.parseNodes(tlchunk)
		if len(nodes) != 0 {
			panic(tlchunk.String(origSrc, "unexpected"))
		}
		ret.topLevel = append(ret.topLevel, nodes[0])
	}
	return
}

func (me *AstFile) parseNodes(toks Tokens) (ret []AstNode) {
	for len(toks) > 0 {
		var node AstNode
		if toks[0].src == "[" || toks[0].src == "(" || toks[0].src == "{" {
			node, toks = me.parseNodeGrouped(toks)
		}
		if node == nil {
			panic(toks.String(me.origSrc, "unexpected"))
		}
	}
	return
}

func (me *AstFile) parseNodeGrouped(toks Tokens) (ret AstNodeBraces, tail Tokens) {
	ret.square = (toks[0].src == "[") && (toks[len(toks)-1].src == "]")
	ret.curly = (toks[0].src == "{") && (toks[len(toks)-1].src == "}")
	if (toks[0].src == "(") && (toks[len(toks)-1].src != ")") && !ret.square && !ret.curly {
		panic(toks.String(me.origSrc, "unmatched brace"))
	}
	return
}

/*
L1 C1 'IdentName'>>>>str<<<<
L1 C4 'Sep'>>>>:<<<<
L1 C6 'IdentOp'>>>>@<<<<
L1 C8 'IdentOp'>>>>=<<<<
L1 C10 'StrLit'>>>>"hello\nworld"<<<<
L3 C1 'IdentName'>>>>c_puts<<<<
L3 C7 'Sep'>>>>:<<<<
L3 C9 'Sep'>>>>(<<<<
L3 C10 'IdentOp'>>>>@<<<<
L3 C11 'Sep'>>>>)<<<<
L3 C12 'IdentName'>>>>·I32<<<<
L3 C18 'IdentOp'>>>>=<<<<
L3 C20 'IdentName'>>>>·extern<<<<
L3 C29 'StrLit'>>>>"puts"<<<<
L5 C1 'IdentName'>>>>main<<<<
L5 C5 'Sep'>>>>:<<<<
L5 C7 'Sep'>>>>(<<<<
L5 C8 'Sep'>>>>)<<<<
L5 C9 'IdentName'>>>>·I32<<<<
L5 C15 'IdentOp'>>>>=<<<<
L5 C17 'Sep'>>>>(<<<<
L5 C18 'Sep'>>>>)<<<<
L6 C5 'IdentName'>>>>c_puts<<<<
L6 C11 'Sep'>>>>(<<<<
L6 C12 'IdentName'>>>>str<<<<
L6 C15 'Sep'>>>>)<<<<
L7 C5 'IdentName'>>>>·ret<<<<
L7 C11 'NumLit'>>>>0<<<<
*/
