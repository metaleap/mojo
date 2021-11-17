package main

type IrModule struct {
    ast *AstFile
}

func (me *IrModule) buildLLvmIr() (ret LlTopLevel) {
    ret.source_filename = me.ast.srcFilePath
    return
}
