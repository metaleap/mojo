package atmocorefn

import (
	"github.com/go-leap/str"
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/lang"
)

func (me *AstDef) nameInScope(name *atmolang.AstIdent) *atmolang.AstIdent {
	for i := range me.state.namesInScope {
		if me.state.namesInScope[i].Val == name.Val {
			return me.state.namesInScope[i]
		}
	}
	return nil
}

func (me *AstDef) namesInScopeAdd(errs *atmo.Errors, names ...*atmolang.AstIdent) (ok int) {
	ndone := make(map[string]bool, len(names))
	for i := range names {
		if ident := me.nameInScope(names[i]); ident != nil {
			errs.AddNaming(&names[i].Tokens[0], "name `"+names[i].Val+"` already taken by "+ident.Tokens[0].Meta.Position.String())
			names[i] = nil
		} else if ndone[names[i].Val] {
			names[i] = nil
		} else {
			ndone[names[i].Val] = true
		}
	}
	for i := range names {
		if names[i] != nil {
			ok, me.state.namesInScope = ok+1, append(me.state.namesInScope, names[i])
		}
	}
	return
}

func (me *AstDef) namesInScopeDrop(num int) {
	me.state.namesInScope = me.state.namesInScope[:len(me.state.namesInScope)-num]
}

func (me *AstDefBase) initFrom(ctx *AstDef, orig *atmolang.AstDef) (errs atmo.Errors) {
	me.Orig = orig
	var numnewnames int
	for i := range me.Orig.Args {
		if name, _ := me.Orig.Args[i].NameOrConstVal.(*atmolang.AstIdent); name != nil {
			numnewnames += ctx.namesInScopeAdd(&errs, name)
		}
	}
	errs.Add(me.initName(ctx))
	errs.Add(me.initBody(ctx))
	errs.Add(me.initArgs(ctx))
	errs.Add(me.initMetas(ctx))
	if !me.Orig.IsTopLevel {
		ctx.dynNameDrop(me.Orig.Name.Val)
	}
	ctx.namesInScopeDrop(numnewnames)
	return
}
func (me *AstDefBase) initName(ctx *AstDef) (errs atmo.Errors) {
	tok := me.Orig.Name.Tokens.First(nil)
	if me.Name, errs = ctx.newAstIdentFrom(&me.Orig.Name, true); me.Name != nil {
		switch name := me.Name.(type) {
		case *AstIdentName:
			if name.Val == "" || ustr.In(name.Val, langReservedOps...) {
				errs.AddNaming(tok, "reserved token not permissible as def name: `"+tok.Meta.Orig+"`")
			}
			name.Val = ctx.state.dynNamePrefs + name.Val
		case *AstIdentTag:
			errs.AddNaming(tok, "invalid def name: `"+name.Val+"` is upper-case, this is reserved for tags")
		case *AstIdentVar:
			errs.AddNaming(tok, "invalid def name: `"+tok.Meta.Orig+"` (begins with multiple underscores)")
		case *AstIdentUnderscores:
			errs.AddNaming(tok, "invalid def name: `"+tok.Meta.Orig+"`")
		default:
			panic(name)
		}
	}
	if !me.Orig.IsTopLevel {
		ctx.dynNameAdd(me.Orig.Name.Val)
	}
	if me.Orig.NameAffix != nil {
		me.coerceFunc = errs.AddVia(ctx.newAstExprFrom(me.Orig.NameAffix)).(IAstExpr)
	}

	return
}

func (me *AstDefBase) initBody(ctx *AstDef) (errs atmo.Errors) {
	me.Body, errs = ctx.newAstExprFrom(me.Orig.Body)
	return
}

func (me *AstDefBase) initArgs(ctx *AstDef) (errs atmo.Errors) {
	if len(me.Orig.Args) > 0 {
		opeq, args := ctx.b.IdentName("=="), make([]AstDefArg, len(me.Orig.Args))
		for i := range me.Orig.Args {
			if errs.Add(args[i].initFrom(ctx, &me.Orig.Args[i], i)); args[i].coerceValue != nil {
				me.Body = ctx.b.Case(ctx.b.Appls(ctx, opeq, args[i].coerceValue, &args[i].AstIdentName), me.Body)
			}
			// if args[i].coerceFunc!=nil {
			// 	me.Body=
			// }
		}
		me.Args = args
	}
	return
}

func (me *AstDefBase) initMetas(ctx *AstDef) (errs atmo.Errors) {
	if len(me.Orig.Meta) > 0 {
		errs.AddTodo(&me.Orig.Meta[0].Toks()[0], "def metas")
	}
	return
}

func (me *AstDefArg) initFrom(ctx *AstDef, orig *atmolang.AstDefArg, argIdx int) (errs atmo.Errors) {
	me.Orig = orig

	var constexpr IAstExprAtomic
	switch v := orig.NameOrConstVal.(type) {
	case *atmolang.AstIdent:
		if constexpr, errs = ctx.newAstIdentFrom(v, true); constexpr != nil {
			if cxn, ok1 := constexpr.(*AstIdentName); ok1 {
				constexpr, me.AstIdentName = nil, *cxn
			} else if cxu, ok2 := constexpr.(*AstIdentUnderscores); ok2 {
				if constexpr, me.AstIdentName.Val, me.AstIdentName.Orig = nil, "ª"+ustr.Int(argIdx), v; cxu.Num() > 1 {
					errs.AddNaming(&v.Tokens[0], "invalid def arg name")
				}
			}
		}
	case *atmolang.AstExprLitFloat, *atmolang.AstExprLitUint, *atmolang.AstExprLitRune, *atmolang.AstExprLitStr:
		constexpr, errs = ctx.newAstExprAtomicFrom(v)
	default:
		panic(v)
	}
	if constexpr != nil {
		me.AstIdentName.Val = "ª" + ustr.Int(argIdx)
		me.coerceValue = constexpr
	}

	if orig.Affix != nil {
		me.coerceFunc = errs.AddVia(ctx.newAstExprFrom(orig.Affix)).(IAstExpr)
	}
	return
}

func (me *AstAppl) initFrom(ctx *AstDef, orig *atmolang.AstExprAppl) (errs atmo.Errors) {
	if len(orig.Args) > 1 {
		errs = me.initFrom(ctx, orig.ToUnary())
	} else {
		me.Arg, errs = ctx.ensureAstAtomFrom(orig.Args[0], false)
		me.Callee = errs.AddVia(ctx.ensureAstAtomFrom(orig.Callee, true)).(IAstIdent)
	}
	me.Orig = orig
	return
}

func (me *AstCases) initFrom(ctx *AstDef, orig *atmolang.AstExprCases) (errs atmo.Errors) {
	me.Orig = orig

	var scrut IAstExpr
	ctx.dynNameAdd("S")
	if orig.Scrutinee != nil {
		scrut = errs.AddVia(ctx.newAstExprFrom(orig.Scrutinee)).(IAstExpr)
	} else {
		scrut = ctx.b.IdentTagTrue()
	}
	scrut = ctx.b.Appl(ctx.b.IdentName("=="), ctx.ensureAstAtomFor(scrut, false))
	scrutid := ctx.ensureAstAtomFor(scrut, true).(IAstIdent)

	me.Ifs, me.Thens = make([][]IAstExpr, len(orig.Alts)), make([]IAstExpr, len(orig.Alts))
	for i := range orig.Alts {
		dna := ustr.Int(i)
		ctx.dynNameAdd(dna)
		alt := &orig.Alts[i]
		if alt.Body == nil {
			panic("bug in atmo/lang: received an `AstCase` with a `nil` `Body`")
		} else {
			me.Thens[i] = errs.AddVia(ctx.newAstExprFrom(alt.Body)).(IAstExpr)
		}
		me.Ifs[i] = make([]IAstExpr, len(alt.Conds))
		for c, cond := range alt.Conds {
			dnc := ustr.Int(c)
			ctx.dynNameAdd(dnc)
			me.Ifs[i][c] = errs.AddVia(ctx.newAstExprFrom(cond)).(IAstExpr)
			me.Ifs[i][c] = ctx.b.Appl(scrutid, ctx.ensureAstAtomFor(me.Ifs[i][c], false))
			ctx.dynNameDrop(dnc)
		}
		ctx.dynNameDrop(dna)
	}
	return
}

func (me *AstLitBase) initFrom(ctx *AstDef, orig atmolang.IAstExprAtomic) {
	me.Orig = orig
}

func (me *AstLitFloat) initFrom(ctx *AstDef, orig atmolang.IAstExprAtomic) {
	me.AstLitBase.initFrom(ctx, orig)
	me.Val = orig.Toks()[0].Float
}

func (me *AstLitUint) initFrom(ctx *AstDef, orig atmolang.IAstExprAtomic) {
	me.AstLitBase.initFrom(ctx, orig)
	me.Val = orig.Toks()[0].Uint
}

func (me *AstLitRune) initFrom(ctx *AstDef, orig atmolang.IAstExprAtomic) {
	me.AstLitBase.initFrom(ctx, orig)
	me.Val = orig.Toks()[0].Rune()
}

func (me *AstLitStr) initFrom(ctx *AstDef, orig atmolang.IAstExprAtomic) {
	me.AstLitBase.initFrom(ctx, orig)
	me.Val = orig.Toks()[0].Str
}

func (me *AstDef) ensureAstAtomFor(expr IAstExpr, retMustBeIAstIdent bool) IAstExprAtomic {
	if !retMustBeIAstIdent {
		if xat, _ := expr.(IAstExprAtomic); xat != nil {
			return xat
		}
	}
	if xid, _ := expr.(IAstIdent); xid != nil {
		return xid
	}
	defname := "¨" + me.dynName(expr)
	idx := me.Locals.index(defname)
	if idx < 0 {
		idx, me.Locals = len(me.Locals), append(me.Locals,
			AstDefBase{Body: expr, Name: me.b.IdentName(defname)})
	}
	return me.Locals[idx].Name
}

func (me *AstDef) ensureAstAtomFrom(orig atmolang.IAstExpr, retMustBeIAstIdent bool) (ret IAstExprAtomic, errs atmo.Errors) {
	if (!retMustBeIAstIdent) && orig.IsAtomic() {
		return me.newAstExprAtomicFrom(orig.(atmolang.IAstExprAtomic))
	}
	if oid, _ := orig.(*atmolang.AstIdent); oid != nil {
		ret, errs = me.newAstIdentFrom(oid, false)
	} else {
		var def AstDefBase
		def.Body, errs = me.newAstExprFrom(orig)
		def.Name = me.b.IdentName("¨" + me.dynName(def.Body))
		me.Locals = append(me.Locals, def)
		ret = me.Locals[len(me.Locals)-1].Name
	}
	return
}

func (me *AstDef) newAstIdentFrom(orig *atmolang.AstIdent, isDecl bool) (ret IAstIdent, errs atmo.Errors) {
	if t1, t2 := orig.IsTag, ustr.BeginsUpper(orig.Val); t1 && t2 {
		var ident AstIdentTag
		ret, ident.Val, ident.Orig = &ident, orig.Val, orig
	} else if t1 != t2 {
		panic("bug in `atmo/lang`: an `atmolang.AstIdent` had wrong `IsTag` value for its `Val` casing")

	} else if orig.IsOpish && orig.Val == "()" {
		var ident AstIdentEmptyParens
		ret, ident.Val, ident.Orig = &ident, orig.Val, orig

	} else if ustr.IsRepeat(orig.Val, '_') {
		var ident AstIdentUnderscores
		ret, ident.Val, ident.Orig = &ident, orig.Val, orig

	} else if orig.Val[0] == '_' {
		var ident AstIdentVar
		ret, ident.Val, ident.Orig = &ident, orig.Val, orig

	} else {
		var ident AstIdentName
		ret, ident.Val, ident.Orig = &ident, orig.Val, orig

	}
	if !isDecl {
		switch ret.(type) {
		case *AstIdentUnderscores:
			errs.AddSyn(&orig.Tokens[0], "lone placeholder illegal here: only permissible in def args or calls")
		case *AstIdentName:
			if me.nameInScope(orig) == nil {
				errs.AddNaming(&orig.Tokens[0], "name unknown: `"+orig.Val+"`")
			}
		}
	}
	return
}

func (me *AstDef) newAstExprAtomicFrom(orig atmolang.IAstExprAtomic) (expr IAstExprAtomic, errs atmo.Errors) {
	expr = errs.AddVia(me.newAstExprFrom(orig)).(IAstExprAtomic)
	return
}

func (me *AstDef) newAstExprFrom(orig atmolang.IAstExpr) (expr IAstExpr, errs atmo.Errors) {
	if orig != nil {
		switch o := orig.(type) {
		case *atmolang.AstIdent:
			expr, errs = me.newAstIdentFrom(o, false)
		case *atmolang.AstExprLitFloat:
			var lit AstLitFloat
			lit.initFrom(me, o)
			expr = &lit
		case *atmolang.AstExprLitUint:
			var lit AstLitUint
			lit.initFrom(me, o)
			expr = &lit
		case *atmolang.AstExprLitRune:
			var lit AstLitRune
			lit.initFrom(me, o)
			expr = &lit
		case *atmolang.AstExprLitStr:
			var lit AstLitStr
			lit.initFrom(me, o)
			expr = &lit
		case *atmolang.AstExprLet:
			renames := make(map[string]string, len(o.Defs))
			locals := make(astDefs, len(o.Defs))
			nunames := make([]*atmolang.AstIdent, len(o.Defs))
			for i := range o.Defs {
				nunames[i] = &o.Defs[i].Name
			}
			numnames := me.namesInScopeAdd(&errs, nunames...)

			for i := range o.Defs {
				errs.Add(locals[i].initFrom(me, &o.Defs[i]))
				renames[o.Defs[i].Name.Val] = locals[i].Name.String()
			}

			expr = errs.AddVia(me.newAstExprFrom(o.Body)).(IAstExpr)
			expr.renameIdents(renames)
			for i := range locals {
				locals[i].Body.renameIdents(renames)
			}
			me.Locals = append(me.Locals, locals...)
			me.namesInScopeDrop(numnames)
		case *atmolang.AstExprAppl:
			if let := o.ToLetExprIfUnderscores(); let == nil {
				var appl AstAppl
				errs = appl.initFrom(me, o)
				expr = &appl
			} else {
				expr, errs = me.newAstExprFrom(let)
			}
		case *atmolang.AstExprCases:
			if let := o.ToLetIfUnionSugar(); let == nil {
				var cases AstCases
				errs = cases.initFrom(me, o)
				expr = &cases
			} else {
				expr, errs = me.newAstExprFrom(let)
			}
		default:
			panic(o)
		}
	}
	return
}

func (me *AstDef) dynNameAdd(s string) string {
	old := me.state.dynNamePrefs
	me.state.dynNamePrefs += s
	return old
}

func (me *AstDef) dynNameDrop(s string) string {
	me.state.dynNamePrefs = me.state.dynNamePrefs[:len(me.state.dynNamePrefs)-len(s)]
	return me.state.dynNamePrefs
}

func (me *AstDef) dynName(expr IAstExpr) string {
	return me.state.dynNamePrefs + expr.DynName()
}
