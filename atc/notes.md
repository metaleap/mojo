### Side-effecting nullary defs:

could auto-desugar into a unary func taking continuation lambda.

ie. what is expressed as `printIntroMsg : -> Int` really specializes / instantiates
into `printIntroMsg : (Int -> T) -> T` where call-site indicates the nature of `T`.

That is for all cases of named nullary bindings to known side-effecting calls, say:
`printAboutNote := @call libc.fwrite "(C) 2020 Haxor Inc" 1 18 libc.stdout`
which could desugar into:
`printAboutNote := cont -> cont (libc.fwrite "(C) 2020 Haxor Inc" 1 18 libc.stdout)`

What about explicit anon nullary lambdas of `( -> someCall pure or not )` form?

They are sugar to express a "future", similarly transformed (just like above) into:

`(cont -> cont (someCall pure or not))` and this becomes their type.

(Whenever undesirable, user can prevent this by lifting any `-> X`  into
`_ -> X` and passing `_` as call arg, it'll be legal, mere sugar over void.)

Result: `printStr (readLn reverseStr)` to mean `printStr(reverseStr(readLn()))`.

Reads a bit odd though. But 1-vs-3 parens levels has charme, an infix no-op operator
for the above call form could further improve how it reads. Just mustn't go wild
with funky-operator inflation all over the place! Keep the set of them tight & simple.

Alternative, instead of typing as `(X -> Y) -> Y`, type as a struct with `then`
(or so) field, same-function-typed. Thus `printStr (readLn.then reverseStr)`.

But usually small anon lambdas are passed, giving a neat-enough "capture variable" flair.

### Detect de-facto boolish logic:

Consider

```
eq a b :=
    @case (@cmp #eq a b) { 1: true, 0: false }
not b :=
    @case b { true: false, false: true }
neq a b :=
    not (eq a b)
```

- Detect any funcs (incl. anon lambdas etc) not-ing the well-known boolish tags.
  - Those are fixed to 0 and 1 always. So instead of `@case`ing, i1 xor: not1=1-1, not0=1-0, notX=1~X
- Also detect (logical) and-ers / or-ers among user funcs for the well-known boolish tags
  - since call args are already fully evaluated, can go for `and` / `or` LL instrs, no short-circuiting
  - likewise for any `@and`s and `@or`s with evaluated operands, no need to gen `@case`s here
- With these boolean-logic detections in place, a not-ing of an (direct or indirect) `cmp`
  can be replaced by the opposite `cmp` with flipped operands
