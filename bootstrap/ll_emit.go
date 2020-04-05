package main

func llEmit(ll_something Any) {
	switch ll := ll_something.(type) {
	case *LLModule:
		llEmitModule(ll)
	default:
		fail(ll)
	}
}

func llEmitModule(ll *LLModule) {
	write(Str("\ntarget datalayout = \""))
	write(ll.target_datalayout)
	write(Str("\"\ntarget triple = \""))
	write(ll.target_triple)
	write(Str("\"\n"))
}
