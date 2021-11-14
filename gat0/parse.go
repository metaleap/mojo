package main

func parse(toks Tokens, origSrc string, srcFilePath string) (ret AstFile) {
	ret.toks, ret.srcFilePath, ret.origSrc = toks, srcFilePath, origSrc
	toplevel := toks.indentLevelChunks(0)
	for _, tlchunk := range toplevel {
		ret.topLevel = append(ret.topLevel, ret.parseNode(tlchunk))
	}
	return
}

func (me *AstFile) parseNode(toks Tokens) AstNode {
	nodes := me.parseNodes(toks)
	if len(nodes) > 1 {
		panic(nodes[1].base().toks.String(me.origSrc, "unexpected"))
	}
	return nodes[0]
}

func (me *AstFile) parseNodes(toks Tokens) (ret []AstNode) {
	for len(toks) > 0 {
		var node AstNode
		if toks[0].src == "[" || toks[0].src == "(" || toks[0].src == "{" {
			node, toks = me.parseNodeBraced(toks)
		}
		if node == nil {
			panic(toks.String(me.origSrc, "unexpected"))
		}
		ret = append(ret, node)
	}
	return
}

func (me *AstFile) parseNodeBraced(toks Tokens) (ret AstNodeBraced, tail Tokens) {
	ret.toks, ret.square, ret.curly = toks, (toks[0].src == "["), (toks[0].src == "{")
	idx := toks.idxOfClosingBrace()
	if idx <= 0 {
		panic(toks.String(me.origSrc, "unmatched brace"))
	}
	ret.list, tail = me.parseNodeList(toks[1:idx], ","), toks[idx+1:]
	return
}

func (me *AstFile) parseNodeList(toks Tokens, sep string) (ret AstNodeList) {
	ret.toks = toks
	for _, nodetoks := range toks.split(sep) {
		ret.nodes = append(ret.nodes, me.parseNode(nodetoks))
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
