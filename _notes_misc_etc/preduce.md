
What facts should be extracted merely from def sources and abs/appl structurings?
Let's write out many examples with fullest results that should be preduced:

# itself -> \x -> x

```
    fn{ a:["0",used], r:[eq@0] }
```

# true -> \x -> \y -> x

```
    fn{
        a:["0",used],
        r: fn{
            a:["1"],
            r:[eq@0]
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
            r:[eq@1]
        }
    }
```

# not -> \x -> (x false) true

After trivials above, it's starting to get interesting...

```
    fn{ // assuming no def-arg or def-name annotations
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
code (conditionals in the right places and using `abyss` for failure, the value
that amounts to undefined / bottom / infinity / crash / rejection / failure).
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
We now still don't know the set of possible return values but we're now forcing
`abyss` for all except structural `true` and `false`, so this set becomes all
that callers now can expect from `not`, which propagates upwards / outwards.

Another annotation option for the same purposes: both ret-val (def-name) and
def-arg. Stating `not:(true|false) -> \x:(true|false) -> ...` (of course in
atmo syntax, not as here in IL pseudo-code) has the charm of explicit
self-documentation and friendlier signature docs / hover tips etc. Though,
since we brought that up: those will show normalized regardless of formulation.

An appendum for the ret-val (def-name) annotation that also `!=x` always
holds will of course be trivially addable but given the knowledge of
`not:(true|false) -> \x:(true|false) -> ...` plus full knowledge of `true`
and `false` &mdash; the same should really be derivable though).

> Generalizing further, once all of the above and below is up and running, ponder
> the possibility of `not` (or now-called `toggle` / `invert` / `flip` or such)
> handling any 2-value set &mdash; would have to cmp the now-possibly-non-callable
> set members though. This goes in generics territory, which will rarely be
> sensible or useful in our typing approach, but can anyway be done functionally
> (aka some sort of `gen` func taking any `_->_` plus sets `x` and `y`, and
> returns a `x->y` coercion-wrapper func around the original func &mdash; it's all
> supposed to be erased at code-gen time, way after static reductions, anyway).

This is how declared-truths = axioms are added to the obviously-observable
intrinsic truths derived from abs/appl structurings. It'd be nice if they
could be avoided but we don't want to infer facts for any given lambda
abstraction from the totality of all its outside uses (that are currently
loaded/known), neither generate permutations of theorems about it to then verify
against the (currently loaded/known) call sites. (That being said, the
overwhelming majority of lambda abstractions being preduced would be local,
within the bounds of the top-level-def's preserved isolation from _its_ callers,
deducing facts about the local lambda from surrounding usage might be at least
permissible, likely beneficial, and possibly unavoidable at some point.)

Axiom annotations aren't as undesirable as they first appear, in that they
desugar into as-if-handwritten wrapper code depending only on prim-`eq` and
the structural equivalent of `true`, aka. the K combinator or "konstant" func,
which luckily doesn't depend on anything. Onward...

One more thing, might be feasible to see how call-site preducing goes &mdash;
maybe just ignoring potential nonsensical uses of `not` is fine as the caller
accepts "any" ret-val as a consequence. Most calls will know that the arg
is `true|false` (such as when received from prim-`eq` or `and`/`or` whose
args are known to be `true|false` from `eq`/`neq`/etc. calls and so on)
&mdash; something to explore and discover in practice.

# and -> \x -> \y -> (x y) false

```
    fn{ // assuming no def-arg or def-name annotations
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
    fn{ // assuming no def-arg or def-name annotations
        a:["0",used,callable{a:true,r:callable{a:@1,r:_}}],
        r: fn{
            a: ["1",used],
            r: []
        }
    }
```

Same principles apply as were just elaborated for `not` / `and`.

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

From the code of `flip` (flips func args, not bool-vals or bits or some such),
nothing is learned about its ret-val (only the arg `f` is being constrained
from the appl expr); of course at call sites further uses of said ret-val may
well constrain it further, possibly even just from the (unknown here) facts
already substantiated for `f` (and its ret-func) as well as for `x` and/or `y`.

# eq -> \x -> \y -> _

Hardcoded facts:

- `eq:(true|false)` &mdash; only returns the (structural) `true` func or `false` func
- `primTypeOf` both `x` and `y` identical
- `abyss` blow-up / crash bubble otherwise

In practice, will have to work out some added prim-type-compat checking that
is known to and utilizable by the static "preduce-stage" analysis. Any uses
of `eq` thus add equal prim-type membership to both args, or if different
ones were already substantiated, the `eq` call is rejected / bubbles up as
`abyss`.


# neq -> \x -> \y -> not ((eq x) y)

```
    fn{ // from `eq` hardcoded-facts
        a:["0",used,primTypeOf@1],
        r: fn{
            a:["1",used,primTypeOf@0],
            r:[(true|false)]
        }
    }
```

The above assumes that `not` is annotated as outlined above, but since `eq`
ret-val is passed to it such annotation shouldn't be necessary for the above
to be preduced. Given that, `neq` would not have to be annotated and still
know that the "inverse" of the ret-val of the `eq` appl is our ret-val.

# must -> \need -> \have -> (((eq need) have) have) abyss

A def with a possible blow-up. We want to establish that all calls that pass
a `have` equal to `need` return it, and all others crash / bubble-up `abyss` /
are rejected statically (which is the real purpose of `abyss`, not runtime crash).

Hence, the ret-val (if there is to be one at all) can only ever be `need`.

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
        a: ["0",used,eq@1],
        r: fn{
            a:["1", used,eq@0],
        }
    }
```

Now what's `eq@attr`. Does this refer to any old user-defined func incidentally
named `eq`? Not so. Well we saw `primTypeOf@attr` earlier but that was palpable
&mdash; prim stuff is tractable statically. Well `eq` however is also as prim as
it gets (even for unknown / semi-known values it firmly books certain ground
facts), so calls to it _must_ translate into new (or not) substantiated facts for
both its args as well as the result receiver (and as always bubbling up).

# dot -> \x -> \f -> f x

```
    fn {
        a: ["0",used],
        r: fn{
            a: ["1",used,callable{a:@0,r:_}],
            r: []
        }
    }
```

# ltr -> \f -> \g -> \x -> g (f x)

```
    fn {
        a: ["0",used,callable{a:@2,r:_}],
        r: fn{
            a: ["1",used,callable{a:_,r:_}],
            r: fn{
                a: ["2",used],
                r: []
            }
        }
    }
```

# rtl -> \g -> \f -> \x -> g (f x)

```
    fn {
        a: ["0",used,callable{a:_,r:_}],
        r: fn{
            a: ["1",used,callable{a:@2,r:_}],
            r: fn{
                a: ["2",used],
                r: []
            }
        }
    }
```
