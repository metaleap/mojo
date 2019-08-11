
What facts should be extracted merely from def sources and abs/appl structurings?
Let's write out many examples with fullest results that should be preduced:

# itself -> \x -> x

```
    fn{ a:["0",used], r:[@0] }
```

# true -> \x -> \y -> x

```
    fn{
        a:["0",used],
        r: fn{
            a:["1"],
            r:[@0]
        }
    }
```

Might just drop the pretense and only have `first` and `second` instead
of `true`, `false`, `konst`... will see.

# false -> \x -> \y -> x

```
    fn{
        a:["0"],
        r: fn{
            a:["1",used],
            r:[@1]
        }
    }
```

# not -> \x -> (x false) true

After trivials above, it's starting to get interesting...

```
    fn{
        a:["0",used,callable{a:false,r:callable{a:true,r:_}}],
        r: [] // nothing known about it at all, could be "anything"
    }
```

W'd like (ie. all callers would like) to know that only the (structural) `true`
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
code (conditionals in the right places and using `undefined` for failure,
which amounts to bottom / infinity / crash / program rejection when encountered).
As they are analyzed same as hand-written code, they contribute to the set
of derived facts.

For the `not`, what annotation to write? It would not _have_ to be on `x`, eg.
`x:(false->true->(true|false))`, because 2/3rds of this is already known from
code. Interesting would be to support the scenario of annotations such as
`x:(_->_->(true|false))` to then be further filled in from the code. But also,
instead of def-arg annotation, a def-name annotation such as `not:(true|false)`
could be done and would then have to automagically further constrain `x` such
that the derived (not redundantly specified) fact of `x:(false->true->_)`
turns into `x:(false->true->(true|false))` from the def-name (ergo ret-val) spec.
We now still don't know the set of possible return values but we're now
undefined for any except structural `true` and `false`, so this becomes all
that callers now can expected from `not`, which robustifies all call-sites too.

Another annotation option for the same purposes: both ret-val (def-name) and
def-arg. Stating `not:(true|false) -> \x:(true|false) -> ...` (of course in
atmo syntax, not as here in IL pseudo-code) has the charm of explicit
self-documentation and friendlier signature docs / hover tips etc. Though,
since we brought that up: those will show normalized regardless of formulation.

An appendum for the ret-val (def-name) annotation that also `!=x` always
holds will of course be trivially addable but given the knowledge of
`not:(true|false) -> \x:(true|false) -> ...` plus full knowledge of `true`
and `false` &mdash; the same should really be derivable though).

> Generalizing further all of the above and below is up and running, ponder
> the possibility of `not` (or now-called `toggle` / `invert` / `flip` or such)
> handling any 2-value set &mdash; would have to cmp the now-possibly-non-callable
> set members though. This goes in generics territory, which will rarely be
> sensible or useful in our typing approach, but can anyway be done functionally
> (aka some sort of `gen` func taking any `_->_` plus sets `x` and `y`, and
> returns a `x->y` coercion-wrapper func around the original func &mdash; it's all
> supposed to be erased at code-gen time, way after static reductions, anyway).

This is how declared-truths = axioms are added to the obviously-observable
intrinsic truths derived from abs/appl structurings. It'd be nice if they
could be avoided but we don't want to infer facts for any given abs from the
totality of all its outside uses (currently loaded/known), neither generate
permutations of theorems about it to then verify against the (currently
loaded/known) call sites.

Axiom annotations aren't as undesirable as they first appear, in that they
desugar into as-if-handwritten wrapper code depending only on prim-`eq` and
the structural equivalent of `true`, aka. the K combinator or "konstant" func,
which luckily doesn't depend on itself. Onward...

One more thing, might be feasible to see how call-site preducing goes &mdash;
maybe just ignoring potential nonsensical uses of `not` is fine as the caller
accepts "any" ret-val as a consequence. Most calls will know that the arg
is `true|false` (such as when received from prim-`eq` or `and`/`or` whose
args are known to be `true|false`) &mdash; something to discover in practice.

# and -> \x -> \y -> (x y) false

```
    fn{
        a:["0",used,callable{a:@1, r:callable{a:false,r:_}}],
        r: fn{
            a:["1",used],
            r: []
        }
    }
```

Similar situation to `not`, again outside call sites must at least observe
that `x` has just been substantiated to be a callable that can at least
accept `y` and will in that case return a callable accepting at least `false`.

Callers can and must learn from our facts for their local context / env;
the other way around is never occurring. Hence, all contradictions occur
in appl exprs (though those may be hidden generated from annotations).

Once again, a complete and all around helpful annotation would be
`and:(true|false) -> \x:(true|false) -> \y:(true|false) -> ...`. Again, such
is fair and benign for `Std`-lib kits to prepare the ground, so that user code
won't have to get as wordy all that frequently, ideally hardly ever. With such
complete constraint annotations in place, derivation of known ret-val from
known arg-vals at call sites should work in preduce, up to the point in the
call chain where both are known if ever, else the 2-sized ret-val set must
remain the ground-truth of course.

# or -> \x -> \y -> (x true) y

```
    fn{
        a:["0",used,callable{a:true,r:callable{a:@1,r:_}}],
        r: fn{
            a:["1,used],
            r: []
        }
    }
```

Same principles apply as were just elaborated for `not` &amp; `and`.

# flip -> \f -> \x -> \y -> (f y) x

```
    fn{
        a:["0",used,callable{a:@2,r:callable{a:@1,r:_}}],
        r: fn{
            a:["1",used],
            r: fn{
                a:["2",used],
                r:[]
            }
        }
    }
```

From the code of `flip`, nothing is learned about its ret-val (only the
arg `f` is being constrained from the appl expr); of course at call sites
further uses of said ret-val may well constrain it further, possibly even just
from the (unknown here) facts already substantiated for `f` and its ret-func.

# eq:(true|false) -> \x -> \y -> _

Hardcoded guarantee: only returns the (structural) `true` or `false` func.

In practice, will have to work out some added prim-type-compat checking that
is known to and utilizable by the static "preduce-stage" analysis. Any uses
of `eq` thus add equal prim-type membership to both args, or if different
ones were already substantiated, the `eq` call is rejected / bubbles up as
`undefined`.

Prim-types are for now: uint, real, tag, func-foo-to-bar. More to come plus
compound prims much later on. Compounds likely to be dually represented,
lambda-calculus style at static preduce time, natively-mapped during code-gen.

# neq -> \x -> \y -> not ((eq x) y)

```
    fn{
        a:["0",used,primTypeOf@1],
        r: fn{
            a:["1",used,primTypeOf@0],
            r:[(true|false)]
        }
    }
```

The above assumes that `not` is annotated as described. Given that (and
built-in hardcoded facts about `eq`), `neq` would not have to be annotated
and still know that the inverse of the ret-val of the `eq` appl is our ret-val.

Again, were no annotations in place, many-perhaps-most call sites might (or
rather _should_) still derive the ret-val set from the arg-vals in place.

# must -> \need -> \have -> (((eq need) have) have) undefined

A def with a possible blow-up. We want to establish that all calls that pass
a `have` equal to `need` return it, and all others crash / bubble-up `undefined`.

In other words, the ret-val if there is to be one can only ever be `need`.

This is the simplest scenario where we find we must statically capture
branchings / divergences. It is also the simplest "type-def" func, testing
for membership in a 1-sized set &mdash; `1234 must` defines the "type" that
contains only `1234`. "There are no types, only values, including sets."

```
    fn{
        a: ["0",used,eq@1],
        r: fn{
            a:["1", used,eq@0],
        }
    }
```

Now what's `eq@attr`. Does this refer to any old user-defined func named `eq`?
Not so. Well we saw `primTypeOf@attr` before above but that was palpable &mdash;
prim stuff is tractable statically. Well `eq` however is also as prim as it gets
(even for unknown / semi-known values it firmly books certain ground facts),
so calls to it _must_ translate into new (or not) substantiated facts for both
its args as well as the result receiver (and as always bubbling up to call sites).

# dot -> \x -> \f -> f x

```
    ...
```

# ltr -> \f -> \g -> \x -> g (f x)

```
    ...
```

# rtl -> \g -> \f -> \x -> g (f x)

```
    ...
```
