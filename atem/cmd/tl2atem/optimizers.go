package main

import (
	tl "github.com/metaleap/go-machines/toylam"
)

func optIsCallIdentity(it *tl.ExprCall) bool {
	if name, _ := it.Callee.(*tl.ExprName); name != nil {
	again:
		if topdef, _ := inProg.TopDefs[name.NameVal]; topdef != nil {
			if fn, _ := topdef.(*tl.ExprFunc); fn != nil && fn.IsIdentity() {
				return true
			} else if name, _ = topdef.(*tl.ExprName); name != nil {
				goto again
			}
		}
	}
	return false
}

func optIsNameIdentity(it tl.Expr) bool {
	for name, _ := it.(*tl.ExprName); name != nil; {
		if td := inProg.TopDefs[name.NameVal]; td == nil {
			name = nil
		} else if fn, _ := td.(*tl.ExprFunc); fn == nil {
			name = nil
		} else if fn.IsIdentity() {
			return true
		} else {
			name, _ = fn.Body.(*tl.ExprName)
		}
	}
	return false
}
