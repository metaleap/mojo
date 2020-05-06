### Rearrangement of locals vs. short-circuiting logical operator sugars

User-land expression:

```dart
  (n == str.len || ferr == 0) ?- #ok n |- #err ferr
  // where
  n := libc.fwrite str 1 str.len file
  ferr := libc.ferror file
```

We originally desugared `||` naively into a nested branch:

```dart
  (n == str.len ?- true |- ferr == 0) ?- #ok n |- #err ferr
  // where
  ferr := libc.ferror file
  n := libc.fwrite str 1 str.len file
```

As it stands with that, two sub-optimal choices remain: either generate the `ferror` call
"beforehand"/"outside" even if `ferr` will turn out to not be consumed, or generate it twice. Both would violate our intended semantics and must be prevented. We'll also stick to non-lazy evaluation, the lazy-eval world doesn't have this conundrum.

The "plain and simple" objective is to reach the equivalent of the below imperative statement sequence, with `ferror` only being called / `ferr` only being written when it clearly _will_ be consumed:

```c
  int n = fwrite(str, 1, str.len, file);
  if (n == str.len)
    return (ok){n};
  int ferr = ferror(file);
  if (ferr == 0)
    return (ok){n};
  else
    return (err){ferr};
```

So want to reach a `||`-less expression more like:

```dart
  (n == str.len) ?- (#ok n) |-
    (ferr -> ferr == 0 ?- #ok n |- #err ferr)
      (libc.ferror file)
  // where
  n := libc.fwrite str 1 str.len file
```

But this could only be done for short-circuiting logical-operator exprs directly placed
inside branches. The scheme would fail for those stand-alone ones bound to a name. They
mustn't imply distinct different operational semantics just based on placement.

This means we need to translate into an SSA / CPS IR instead of a SAPL-like "functional IR"
right from the start, even during the "only an interpreter" stage 0.

Time to take another, deeper look into the Thorin approach..

### Detect de-facto boolish logic:

Consider

```dart
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
