### Hoisting:

after glyph-desugaring:
- for each func-expr that isnt the value of a top-def-body
  - first process its body with this sequence of steps, then:
  - determine new-top-name
  - gather all refs-to-locals
  - make new top-def with these prepended as params
  - rewrite the encountered func-expr with:
    - a partial call to the new top def if any new params were prepended, with those names as args
    - an ident to the new top def otherwise

result: all sub-defs are param-less locals. so they can be turned into "lambda
let"s and for this purpose, re-ordered correctly. any remaining mutual-recursion
found in them is a reportable error (circular defs ie. `a := c, b := a, c := b`).

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
