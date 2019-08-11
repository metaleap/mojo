
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
        a:["0",used,callable{a:false,r:callable{a:true}}],
        r: @0#callable.r(false)#callable.r(true)
    }
```

W'd like (ie. all callers would like) to know that only the (structural) `true`
func or the (structural) `false` func will ever be returned, but this isn't
clear from the insides of `not` and could only be with some sub-par confidence
_theorized_ after scrutinizing all its (currently-known / -loaded, _and_
assuming they're all blunder-free!) outside uses.
From the appl expr all that's known is that `x` must accept (structural, as
always) `false` and always return a callable that accepts (structural) `true`.

Equivalent "issue" / thinking point for `and` as well as `or` as far as
deducing from only their expr the set of possible rets (again for both should
be `true|false`, again for both structurally).

Now it's already decided that we can add further (non-contradictory with
intrinsic truths derived from the rest of the code) constraints via def-arg
and def-name affixes to any global or local def. They desugar into normal
code (conditionals in the right places and using `undefined` which equals
to botttom / infinity / crash / program rejection in the failure paths).

If such an annotation, it should not have to be on `x` instead of `not`, ie.
`x:(false->true->_)`, because all this is already known from code. So instead
def-nam annotation such as `not:(true|false)` would then further constrain `x`
such that the implicit (not redundantly explicated) knowledge of
`x:(false->true->_)` would then be forced into `x:(false->true->(true|false))`.

On the other hand, stating `not:(true|false) = \x:(true|false) -> ...`, ideally
even appendum for the ret that `!=x` always holds, has the charme of explicit
self-documentation.

This is where axioms are added to the naked-eye-obviously-observable intrinsic
truths extracted from abs/appl structurings. It'd be nice if they could be avoided
but we don't want to infer facts for any given abs from the totality of all
its outside uses (currently loaded/known), neither generate permutations of
theorems about it to then have proven against the (currently loaded/known)
outside uses.

Axiom annotations aren't as undesirable as they first appear, in that they
desugar into as-if-handwritten wrapper code. Would this cause an infinite loop
for prime building blocks such as `and`, `or`, `not` &mdash; let's find out
in the active doing (and feel the real pain hard) rather than mere pondering!

Onward.

# and -> \x -> \y -> (x y) false

```
    fn{
        a:["0",used,callable{a:@1, r:callable{a:false}}],
        r: fn{
            a:["1",used],
            r: [@0#callable.r(@1)#callable.r(false)] // TODO
        }
    }
```

# or -> \x -> \y -> (x true) y

```
    fn{
        a:["0",used,callable{a:true,r:callable{a:@1}}],
        r: fn{
            a:["1,used],
            r: [@0#callable.r(true)#callable.r(@1)] // TODO
        }
    }
```

# flip -> \f -> \x -> \y -> (f y) x

```
    fn{
        a:["0",used,callable{a:@2,r:callable{a:@1}}],
        r: fn{
            a:["1",used],
            r: fn{
                a:["2",used],
                r:@0#callable.r(@2)#callable.r(@1)
            }
        }
    }
```

# eq:(true|false) -> \x -> \y -> _

hardcoded guarantee: only returns the true func or the false func (structurally).

in practice, will have to work out some added prim-type-compat checking that
is known to and utilizable by the static "preduce-stage" analysis.

# neq -> \x -> \y -> not ((eq x) y)

```
    fn{
        a:["0",used],
        r: fn{
            a:["1",used],
            r:[@not.r(eq.r(@0)#callable.r(@1)) aka _]
        }
    }
```

# must -> \need -> \have -> (((eq need) have) have) undefined

```
    fn {
        a:["0",used],
        r: fn{
            a:["1",used],

        }
    }
```

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
