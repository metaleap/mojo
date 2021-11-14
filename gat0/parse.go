package main

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

func parse(toks Tokens, origSrc string, srcFilePath string) (ret AstProg) {
	ret.toks, ret.srcFilePath = toks, srcFilePath
	toplevel := toks.indentLevelChunks(toks, 0)
	for _, tlc := range toplevel {
		print("\n>>>>>>>>>>")
		print(tlc.String(origSrc))
		println("<<<<<<<<<<\n")
	}
	return
}
