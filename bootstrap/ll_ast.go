package main

const llmodule_default_target_datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
const llmodule_default_target_triple = "x86_64-unknown-linux-gnu"

type LLModule struct {
	target_datalayout Str
	target_triple     Str
	globals           []LLGlobal
	funcs             []LLFunc
}

type LLGlobal struct {
}

type LLFunc struct {
}
