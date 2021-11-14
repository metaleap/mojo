package main

type AstProg struct {
	srcFilePath string
	toks        Tokens
}

func (me *AstProg) buildLLvmIr() (ret LlTopLevel) {
	ret.source_filename = me.srcFilePath
	return
}
