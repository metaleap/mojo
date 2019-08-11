
What facts should be extracted merely from def sources and abs/appl structurings?
Let's write out many examples with fullest results that should be preduced:

# itself: \x -> x

```
    fn{ a:["0"], r:[@0] }
```

# true: \x -> \y -> x

```
    fn{
        a:["0",used],
        r: fn{
            a:["1"],
            r:[@0]
        }
    }
```

# false: \x -> \y -> x

```
    fn{
        a:["0"],
        r: fn{
            a:["1",used],
            r:[@1]
        }
    }
```

# and: \x -> \y -> (x y) false

```
    fn{
        a:["0",used,callable{a:@1, r:callable{a:false}}],
        r: fn{
            a:["1",used],
            r: [@0#callable.r(@1)#callable.r(false)] // TODO
        }
    }
```

# or: \x -> \y -> (x true) y

```
    fn{
        a:["0",used,callable{a:true,r:callable{a:@1}}],
        r: fn{
            a:["1,used],
            r: [@0#callable.r(true)#callable.r(@1)] // TODO
        }
    }
```

# not: \x -> (x false) true

```
    fn{
        a:["0",used,callable{a:false,r:callable{a:true}}],
        r: @0#callable.r(false)#callable.r(true)
    }
```

# flip: \f -> \x -> \y -> (f y) x

```
    ...
```

# neq: \x -> \y -> not ((== x) y)

```
    ...
```

# must: \need -> \have -> (((== need) have) have) undefined

```
    ...
```

# dot: \x -> \f -> f x

```
    ...
```

# ltr: \f -> \g -> \x -> g (f x)

```
    ...
```

# rtl: \g -> \f -> \x -> g (f x)

```
    ...
```
