package main

var (
	parsingOpsNumeric         = []string{"+", "-", "*", "/", "%"}
	parsingOpsShortcircuiting = []string{"&&", "||"}
	parsingOpsBitwise         = []string{"!", "&", "|", "~", "^", "<<", ">>"}
	parsingOpsCmp             = []string{"<", ">", "==", "!=", ">=", "<="}
)

func parse(toks Tokens, origSrc string, srcFilePath string) (ret AstFile) {
	ret.toks, ret.srcFilePath, ret.origSrc = toks, srcFilePath, origSrc
	toplevelchunks := toks.indentLevelChunks(0)
	for _, tlc := range toplevelchunks {
		if node := ret.parseNode(tlc); node != nil {
			ret.topLevel = append(ret.topLevel, node)
		}
	}
	return
}

func (me *AstFile) parseNode(toks Tokens) AstNode {
	nodes := me.parseNodes(toks)
	if len(nodes) == 0 {
		return nil
	} else if len(nodes) > 1 {
		return AstNodeList{AstNodeBase: AstNodeBase{toks: toks}, nodes: nodes}
	}
	return nodes[0]
}

func (me *AstFile) parseNodes(toks Tokens) (ret []AstNode) {
	for len(toks) > 0 {
		var node AstNode
		if t := &toks[0]; t.kind == tokKindComment {
			toks = toks[1:]
		} else if t.src == "[" || t.src == "(" || t.src == "{" {
			node, toks = me.parseNodeBraced(toks)
		} else if toks.idxAtLevel0(",") >= 0 {
			node, toks = me.parseNodeList(toks, ","), nil
		} else if idx := toks.idxAtLevel0("="); idx > 0 {
			node, toks = me.parseNodePair(toks, idx), nil
		} else if idx = toks.idxAtLevel0(":"); idx > 0 {
			node, toks = me.parseNodePair(toks, idx), nil
		} else if toks.anyAtLevel0(parsingOpsShortcircuiting...) {
			node, toks = me.parseNodeList(toks, parsingOpsShortcircuiting...), nil
		} else if toks.anyAtLevel0(parsingOpsCmp...) {
			node, toks = me.parseNodeList(toks, parsingOpsCmp...), nil
		} else if toks.anyAtLevel0(parsingOpsNumeric...) {
			node, toks = me.parseNodeList(toks, parsingOpsNumeric...), nil
		} else if toks.anyAtLevel0(parsingOpsBitwise...) {
			node, toks = me.parseNodeList(toks, parsingOpsBitwise...), nil
		} else {
			node, toks = AstNodeAtom{AstNodeBase: AstNodeBase{toks: toks}}, toks[1:]
		}
		if node != nil {
			ret = append(ret, node)
		}
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

func (me *AstFile) parseNodeList(toks Tokens, seps ...string) (ret AstNodeList) {
	tokss, sep := toks.split(seps...)
	ret.sep, ret.toks = sep, toks
	for _, nodetoks := range tokss {
		ret.nodes = append(ret.nodes, me.parseNode(nodetoks))
	}
	return
}

func (me *AstFile) parseNodePair(toks Tokens, idx int) (ret AstNodePair) {
	ret.toks, ret.sep = toks, toks[idx].src
	ret.lhs = me.parseNode(toks[:idx])
	ret.rhs = me.parseNode(toks[idx+1:])
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
