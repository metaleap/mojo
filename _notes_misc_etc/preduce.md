
What facts should be extracted merely from def sources and abs/appl structurings?
Let's write out many examples with fullest results that should be preduced:

## itself -> \x -> x

```
    fn{
        a: [used:1],
        r: [@x]
    }
```

## true -> \x -> \y -> x

```
    fn{
        a: [used:1],
        r: fn{
            a: [used:0],
            r: [@x]
        }
    }
```

## false -> \x -> \y -> y

```
    fn{
        a: [used:0],
        r: fn{
            a: [used:1],
            r: [@y]
        }
    }
```

## not -> \x -> (x false) true

After trivials above, it's starting to get interesting...

```
    fn{ // assuming no def-arg or def-name annotations
        a: [used:1, fn{a:false, r:fn{a:true, r:[]}}],
        r: [@x.fn.r.fn.r]
    }
```

We'd like (ie. all callers would like) to know that only the (structural) `true`
func or the (structural) `false` func will ever be returned, but this isn't
clear from the insides of `not` and that's all we'll go by (not what outside
callers claim or their uses of `not` could potentially indicate).
From the appl expr all that's known is that `x` must accept (structural, as
always) `false` and always return a callable that accepts (structural) `true`.
Equivalent "issue" / thinking point for `and` as well as `or` as far as
deducing from only their expr the set of possible rets (again for both should
be `true|false`, each structurally).

Now it's already decided that we can add further (non-contradictory with
facts derived from the rest of the code or specified) constraints via def-arg
and def-name affixes to any global or local def. They desugar into normal
code (conditionals in the right places and using `abyss` for failure, the value
that amounts to undefined / bottom / infinity / crash / rejection / failure).
As they are analyzed same as hand-written code, they contribute to the set
of derived facts.

For the `not`, what annotation to write? It would not _have_ to be on `x`, eg.
`x:(false->true->(true|false))`, because 2/3rds of this is already known from
code. Interesting would be to support the scenario of annotations such as
`x:(_->_->(true|false))` to then be further filled in from the code. But also,
instead of def-arg annotation, a def-name annotation such as `not:(true|false)`
could be done and would then _effectively appear to_ further constrain `x` such
that the derived (not redundantly specified) fact of `x:(false->true->_)`
turns into `x:(false->true->(true|false))` from the def-name (ergo ret-val) spec.
In truth, with such we now still don't know the set of possible return values
from `(x false) true` but we're now forcing `abyss` for all except structural
`true` and `false`, so this set becomes all that callers now can expect from
`not`, which propagates upwards / outwards. Since `x` is not a global, in
such situations it's fair to interpret such usage-site constraints / explicitly
forced `abyss` divergences as adding to the full "type-spec" of what `not` (and
other local uses) expects from `x`, _in addition to_ what the inner appl usage
indicates. All `not` callers must thusly scrutinize the `x` being passed, or
propagate the constraints to _their_ callers.

An appendum for the ret-val (def-name) annotation that also `!=x` always
holds will of course be trivially addable but given the knowledge of
`not:(true|false) -> \x:(false->true->true|false) -> ...` plus full knowledge
of `true` and `false` &mdash; the same should really be derivable though).

Such is how declared-truths / user-defined axioms are added to the obviously
observable "intrinsic" truths derived from abs/appl structurings. It'd be nice
if they could be avoided but we don't want to infer facts for any given abs
from the totality of all its outside uses (that are currently loaded/known and
naively assumed to be blunder-free), neither generate permutations of theorems
about it to then verify against the (currently loaded/known) call sites. (That
being said, the overwhelming majority of abs being preduced would be local,
within the bounds of the def's preserved isolation from _its_ callers, deducing
facts about the local abs from surrounding usage might be at least permissible,
likely beneficial, and possibly unavoidable at some point. On the other hand,
could be redundant work as appls in general own their local context and all
intel about it, further constraining the abs intel they gather in this local
context, whether the abs is local or not. Cleaner design to handle all abs
equally and keep local-related contextual intel in appl intel.)

Axiom annotations aren't as undesirable as they first appear, in that they
desugar into as-if-handwritten wrapper code depending _usually_ only on prim
cmps and the structural equivalent of `true`, ie. the K combinator aka. "konst",
which doesn't depend on anything. However: any pure func in scope is OK.
(Pure funcs reduce to prim-ops and similarly basal combinators necessarily,
only recursion must be considered as potential mayhem with the usual caution.)

Just accepting potential nonsensical uses of `not` wouldn't break things per se,
as the caller evidently accepts "any" ret-val as a consequence. Most appls
would know if the arg is `true|false` (such as when received from prim-`eq`
or `and`/`or` whose args are known to be `true|false` from `eq`/`neq`/etc.
calls and so on) &mdash; to be exhaustively verified before putting such
annotations in place in `Std`! In general, the API / lib designer must choose
how much to restrict (or not) "nonsensical" (from their PoV) usages.

## and -> \x -> \y -> (x y) false

```
    fn{ // assuming no def-arg or def-name annotations
        a: [used:1, fn{a:@y, r:fn{a:false, r:[]}}],
        r: fn{
            a: [used:1, link:@x.fn.a],
            r: [@x.fn.r.fn.r]
        }
    }
```

Similar situation to `not`, again outside call sites must at least observe
that `x` has just been substantiated to be a callable that can at least
accept `y` and will in that case return a callable accepting at least `false`.

Appls can and must gather and integrate abs facts for their local context /
env, the other way around is not happening (conceptually, with the caveats from
`not` thoughts above). Hence, all contradictions occur in appl exprs (though
those may be hidden generated from annotations).

Once again, a complete and all around helpful annotation would be
`and:(true|false) -> \x:(true|false) -> \y:(true|false) -> ...`. Again, such
is fair and benign for `Std`-lib kits to prepare the ground, so that user code
won't have to get as wordy all that frequently, ideally hardly ever. With such
complete constraint annotations in place, derivation of known ret-val from
known arg-vals at call sites should work in preduce, up to the point in the
call chain where both are known if ever, else the 2-sized ret-val set must
remain the ground-truth of course.

## or -> \x -> \y -> (x true) y

```
    fn{ // assuming no def-arg or def-name annotations
        a: [used:1, fn{a:true, r:fn{a:@y, r:[]}}],
        r: fn{
            a: [used:1, link:@x.fn.r.fn.a],
            r: [@x.fn.r.fn.r]
        }
    }
```

Same principles apply as were just elaborated for `not` / `and`.

## flip -> \f -> \x -> \y -> (f y) x

```
    fn{
        a: [used:1, fn{a:@y, r:fn{a:@x, r:[]}}],
        r: fn{
            a: [used:1, link:@f.fn.r.fn.a],
            r: fn{
                a: [used:1, link:@f.fn.a],
                r: [@f.fn.r.fn.r]
            }
        }
    }
```

From the code of `flip` (flips `f`'s args, not bool-vals or bits or some such),
nothing is learned about its ret-val (only the arg `f` is being constrained
from the appl expr); of course at call sites further uses of said ret-val may
well constrain it further, possibly even just from the (unknown here) facts
already substantiated for `f` (and its ret-func) as well as for `x` and/or `y`.

Once again it's seen that the meat of intel is gathered from inside appls,
whereas abs scope mostly merely transcribes existing structures mechanistically
(though in turn itself utilizing richer inner-appl-scoped intel of course).

## eq

Hardcoded facts:

```
    fn {
        a: [used:1, primTypeEqToPrimTypeOf:@1],
        r: fn {
            a: [used:1, primTypeEqToPrimTypeOf:@0],
            r: [(@0==@1->true|_->false)]
        }
    }
```

- `eq:(true|false)` &mdash; only returns the (structural) `true` func or `false` func, ever, for valid inputs defined as:
- `primTypeEqToPrimTypeOf` `y` for `x` and vice-versa
- `abyss` blow-up / crash bubble otherwise (implicit, `eq` only defined for
  above pre-condition)

In practice, will have to work out some added prim-type-compat checking that
is known to and utilizable by the static "preduce-stage" analysis. Any uses
of `eq` thus add equal prim-type membership to both args, linking each other
on that facet. Also, all intrinsic facts (eg. commutativity) of `eq` (and thus
bool ops in general) as well as all `eq`'s prim-type impls must be fully known
and available to preduce.

## neq -> \x -> \y -> not ((eq x) y)

```
    fn{
        a: [used:1, link:eq.fn.a],
        r: fn{
            a: [used:1, link:eq.fn.r.fn.a],
            r: [(@x==@y->false|_->true)] // the not.eq appl must derive, not just link:.fn.r
        }
    }
```

The above regardless of whether `not` is annotated as outlined earlier or not:
since `eq` ret-val (known as most-generally `true|false`, or more specific) is
passed to `not`, such annotation shouldn't be necessary for the above to be
preduced. Given that, `neq` would not have to be annotated itself either and
still know that the "flip-side" of the ret of the `eq` appl is _its_ ret.

## must -> \need -> \have -> (((eq need) have) have) abyss

A def with a possible blow-up. We want to establish that all calls that pass
a `have` equal to `need` return it, and all others crash / bubble-up `abyss` /
are rejected statically (which is the real purpose of `abyss`, not runtime crash).

Hence, the ret-val (if there is to be one at all) can only ever be `have` aka. `need`.

At call site: if `have` is known statically as is `need`, the equality can be
verified statically and the fact recorded for both, or rather, the code
rejected on contradiction, because the fact of their mutual equality now holds
always due to the call to `must` and because `abyss` cannot be explicitly
"caught" (also not explicitly "thrown") in pure funcs, only outside of them.

This is the simplest scenario where we find we must statically capture truth
values and consequences &mdash; also imagine a similar `mustnt` using `neq`!
It is also the simplest "type-def" func, testing for membership in a 1-sized
set &mdash; `1234 must` defines the "type" that contains only `1234`.
("There are no types, only values, including sets.")

```
    fn{
        a: [used:1, link:@eq.fn.a],
        r: fn{
            a: [used:1, link:@eq.fn.r.fn.a],
            r: [@have] // we know eq, true, false --- and abyss is undefined.
        }
    }
```

## dot -> \x -> \f -> f x

```
    fn {
        a: [used:1, link:@f.fn.a],
        r: fn{
            a: [used:1, fn{a:@x, r: []}],
            r: [@f.fn.r]
        }
    }
```

## ltr -> \f -> \g -> \x -> g (f x)

```
    fn {
        a: [used:1, fn{a:@x, r:[link:@g.fn.a]}],
        r: fn{
            a: [used:1, fn{a:[@f.fn.r], r:[]}],
            r: fn{
                a: [used:1, link:@f.fn.a],
                r: [@g.fn.r]
            }
        }
    }
```

## rtl -> \g -> \f -> \x -> g (f x)

```
    fn {
        a: [used:1, fn{a: [@f.fn.r], r: []}],
        r: fn{
            a: [used:1, fn{a:@x, r:[@g.fn.a]}],
            r: fn{
                a: [used:1, ref:@f.fn.a],
                r: [@g.fn.r]
            }
        }
    }
```

# Appls

contribute in building up ancestor abs details (their `a` and `r`) from
inner abs-refs / arg-refs. Mostly merging "any-and-all findings" for
later scrutinizing.

- _lhs_: preduce to `fn`
  - if resolves to hole (ie. local), ensure its `fn`, it can be freely annotated
  - if resolves to non-hole `fn` (ie. global), added annotations in local-ctx copy
- _rhs_: preduce,
  - if resolves to hole, link up with _lhs'_ `a`
  - else if _lhs_ is "ours" (own / local / hole), add findings to its `a`
  - else to local-ctx copy's `a`
- whole:
  - if _lhs_ not `fn` then `abyss`
  - else _lhs_.r after any refs to _rhs_ filled in

Most is naive propagation with the meat coming from prim-ops and prim-types.

The above distinctions about "own-or-not" holes / contexts is conceptual.
In practice, there is always an "own" / current / local context linked up
with a parent and possibly others, and the whole web of links is
flattened-copy-merged when it is complete:

# Procedure

Each abs and each appl when encountered constructs its local context, linked
to "current" (now "parent") context. Entering a def ie. global, the current
context is stored away temporarily prior to preducing its expr and restored
afterwards. Before such restoring of the old ctx, the full tree of the def-expr's
own just-gathered ctx-nodes should also be flattened into the complete finished
facts-sheet for the def, with no obvious (structural / "lexical") redundancies /
duplications. However, no implication deductions / derivations / contradiction
detections just yet: to be done once all defs being (pre)preduced are up to date.

# Pseudo-code &amp; walk-throughs

```

preduce(lit):
    ret hole{const: lit.val, primType: lit.primType}

preduce(var):
    v = preduce(var.dst)
    if v.isHole
        v.used++

preduce(appl):
    lhs = preduce(appl.lhs)
    if !lhs.isFn
        if lhs.isHole lhs.addFn() else fail(nonFn)
    ctx = prepCtxAppropriateForFnOf(lhs)
    rhs = preduce(appl.rhs)
    ctx.fn.a.linkUpWith(rhs)
    ret ctx.fn.r

preduce(abs):
    ret fn{
        a: hole{},
        r: preduce(abs.expr)
    }

```

## itself -> \x -> x

```
    fn{
        a: [used:1],
        r: [@x]
    }
```

### Pseudo-steps:

```
    argx = hole{"x"}
    ret.addFactsFrom(preduce(body))
        argx.used++
```

## true -> \x -> \y -> x

```
    fn{
        a: [used:1],
        r: fn{
            a: [used:0],
            r: [@x]
        }
    }
```

### Pseudo-steps:

```
    argx = hole{"x"}
    ret.addFactsFrom(preduce(body))
        argy = hole{"y"}
        ret.addFactsFrom(preduce(body))
            argx.used++
```

## false -> \x -> \y -> y

```
    fn{
        a: [used:0],
        r: fn{
            a: [used:1],
            r: [@y]
        }
    }
```

### Pseudo-steps:

```
    argx = hole{"x"}
    ret.addFactsFrom(preduce(body))
        argy = hole{"y"}
        ret.addFactsFrom(preduce(body))
            argy.used++
```

## not -> \x -> (x false) true

```
    fn{ // assuming no def-arg or def-name annotations
        a: [used:1, fn{a:false, r:fn{a:true, r:[]}}],
        r: [@x.fn.r.fn.r]
    }
```

### Pseudo-steps:

```
    argx = hole{"x"}
    ret.addFactsFrom(preduce(body)) // (x false) true
        lhs = preduce(lhs).ensureFn() // x false
            lhs = preduce(lhs).ensureFn() // x
                argx.used++
            rhs = preduce(rhs).linkWith(lhs.fn.a) // false <-> x.fn.a
        rhs = preduce(rhs).linkWith(lhs.fn.a) // true
```

## and -> \x -> \y -> (x y) false

```
    fn{ // assuming no def-arg or def-name annotations
        a: [used:1, fn{a:@y, r:fn{a:false, r:[]}}],
        r: fn{
            a: [used:1, link:@x.fn.a],
            r: [@x.fn.r.fn.r]
        }
    }
```

### Pseudo-steps:

```
    argx = hole{"x"}
    ret.addFactsFrom(preduce(body))
        argy = hole{"y"}
        ret.addFactsFrom(preduce(body)) // (x y) false
            lhs = preduce(lhs).ensureFn() // x y
                lhs = preduce(lhs).ensureFn()
                    argx.used++
                rhs = preduce(rhs).linkWith(lhs.fn.a)
                    argy.used++
            rhs = preduce(rhs).linkWith(lhs.fn.a)
```

## or -> \x -> \y -> (x true) y

```
    fn{ // assuming no def-arg or def-name annotations
        a: [used:1, fn{a:true, r:fn{a:@y, r:[]}}],
        r: fn{
            a: [used:1, link:@x.fn.r.fn.a],
            r: [@x.fn.r.fn.r]
        }
    }
```

### Pseudo-steps:

```
    argx = hole{"x"}
    ret.addFactsFrom(preduce(body))
        argy = hole{"y"}
        ret.addFactsFrom(preduce(body)) // (x true) y
            lhs = preduce(lhs).ensureFn() // x true
                lhs = preduce(lhs).ensureFn()
                    argx.used++
                rhs = preduce(rhs).linkWith(lhs.fn.a)
            rhs = preduce(rhs).linkWith(lhs.fn.a)
                argy.used++
```

## flip -> \f -> \x -> \y -> (f y) x

```
    fn{
        a: [used:1, fn{a:@y, r:fn{a:@x, r:[]}}],
        r: fn{
            a: [used:1, link:@f.fn.r.fn.a],
            r: fn{
                a: [used:1, link:@f.fn.a],
                r: [@f.fn.r.fn.r]
            }
        }
    }
```

### Pseudo-steps:

```
    argf = hole{"f"}
    ret.addFactsFrom(preduce(body))
        argx = hole{"x"}
        ret.addFactsFrom(preduce(body))
            argy = hole{"y"}
            ret.addFactsFrom(preduce(body)) // (f y) x
                lhs = preduce(lhs).ensureFn() // f y
                    lhs = preduce(lhs).ensureFn()
                        argf.used++
                    rhs = preduce(rhs).linkWith(lhs.fn.a)
                        argy.used++
                rhs = preduce(rhs).linkWith(lhs.fn.a)
                    argx.used++
```

## eq

```
    fn {
        a: [used:1, primTypeEqToPrimTypeOf:@1],
        r: fn {
            a: [used:1, primTypeEqToPrimTypeOf:@0],
            r: [(@0==@1->true|_->false)]
        }
    }
```

## neq -> \x -> \y -> not ((eq x) y)

```
    fn{
        a: [used:1, link:eq.fn.a],
        r: fn{
            a: [used:1, link:eq.fn.r.fn.a],
            r: [(@x==@y->false|_->true)] // the not.eq appl must derive, not just link:.fn.r
        }
    }
```

### Pseudo-steps:

```
    argx = hole{"x"}
    ret.addFactsFrom(preduce(body))
        argy = hole{"y"}
        ret.addFactsFrom(preduce(body))
            lhs = preduce(lhs).ensureFn() // not
            rhs = preduce(rhs).linkWith(lhs.fn.a) // ((eq x) y)
                lhs = preduce(lhs).ensureFn() // eq x
                    lhs = preduce(lhs).ensureFn() // eq
                    rhs = preduce(rhs).linkWith(lhs.fn.a) // x
                        argx.used++
                rhs = preduce(rhs).linkWith(lhs.fn.a) // y
                    argy.used++
```

## must -> \need -> \have -> (((eq need) have) have) abyss

```
    fn{
        a: [used:1, link:@eq.fn.a],
        r: fn{
            a: [used:1, link:@eq.fn.r.fn.a],
            r: [@have] // we know eq, true, false --- and abyss is undefined.
        }
    }
```

### Pseudo-steps:

```
    argn = hole{"need"}
    ret.addFactsFrom(preduce(body))
        argh = hole{"have"}
        ret.addFactsFrom(preduce(body)) // (((eq need) have) have) abyss
            lhs = preduce(lhs).ensureFn() // (((eq need) have) have)
                lhs = preduce(lhs).ensureFn() // (eq need) have)
                    lhs = preduce(lhs).ensureFn() // eq need
                        lhs = preduce(lhs).ensureFn() // eq
                        rhs = preduce(rhs).linkWith(lhs.fn.a)  // need
                            argn.used++
                    rhs = preduce(rhs).linkWith(lhs.fn.a)  // have
                        argh.used++
                rhs = preduce(rhs).linkWith(lhs.fn.a)  // have
                    argh.used++
            rhs = preduce(rhs).linkWith(lhs.fn.a) // abyss
```

## dot -> \x -> \f -> f x

```
    fn {
        a: [used:1, link:@f.fn.a],
        r: fn{
            a: [used:1, fn{a:@x, r: []}],
            r: [@f.fn.r]
        }
    }
```

### Pseudo-steps:

```
    argx = hole{"x"}
    ret.addFactsFrom(preduce(body))
        argf = hole{"f"}
        ret.addFactsFrom(preduce(body)) // f x
            lhs = preduce(lhs).ensureFn() // f
                argf.used++
            rhs = preduce(rhs).linkWith(lhs.fn.a) // x
                argx.used++
```

## ltr -> \f -> \g -> \x -> g (f x)

```
    fn {
        a: [used:1, fn{a:@x, r:[link:@g.fn.a]}],
        r: fn{
            a: [used:1, fn{a:[@f.fn.r], r:[]}],
            r: fn{
                a: [used:1, link:@f.fn.a],
                r: [@g.fn.r]
            }
        }
    }
```

### Pseudo-steps:

```
    argf = hole{"f"}
    ret.addFactsFrom(preduce(body))
        argg = hole{"g"}
        ret.addFactsFrom(preduce(body))
            argx = hole{"x"}
            ret.addFactsFrom(preduce(body)) // g (f x)
                lhs = preduce(lhs).ensureFn() // g
                    argg.used++
                rhs = preduce(rhs).linkWith(lhs.fn.a) // f x
                    lhs = preduce(lhs).ensureFn() // f
                        argf.used++
                    rhs = preduce(rhs).linkWith(lhs.fn.a) // x
                        argx.used++
```

## rtl -> \g -> \f -> \x -> g (f x)

```
    fn {
        a: [used:1, fn{a: [@f.fn.r], r: []}],
        r: fn{
            a: [used:1, fn{a:@x, r:[@g.fn.a]}],
            r: fn{
                a: [used:1, ref:@f.fn.a],
                r: [@g.fn.r]
            }
        }
    }
```

### Pseudo-steps:

```
    argg = hole{"g"}
    ret.addFactsFrom(preduce(body))
        argf = hole{"f"}
        ret.addFactsFrom(preduce(body))
            argx = hole{"x"}
            ret.addFactsFrom(preduce(body)) // g (f x)
                lhs = preduce(lhs).ensureFn() // g
                    argg.used++
                rhs = preduce(rhs).linkWith(lhs.fn.a) // f x
                    lhs = preduce(lhs).ensureFn() // f
                        argf.used++
                    rhs = preduce(rhs).linkWith(lhs.fn.a) // x
                        argx.used++
```
