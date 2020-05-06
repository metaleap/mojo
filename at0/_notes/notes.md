### Rearrangement of locals vs. short-circuiting logical operator sugars

User-land expression:

```dart
  (n == str.len || ferr == 0) ?- #ok n |- #err ferr
  // where
  n := libc.fwrite str 1 str.len file
  ferr := libc.ferror file
```

Originally desugared `||` naively into a nested branch:

```dart
  (n == str.len ?- true |- ferr == 0) ?- #ok n |- #err ferr
  // where
  ferr := libc.ferror file
  n := libc.fwrite str 1 str.len file
```

As it stands with that, two sub-optimal choices remain: either generate the `ferror` call
"beforehand"/"outside" even if `ferr` will turn out to not be consumed, or generate it twice.
Both would violate our intended semantics and must be prevented. (Also want to stick to
non-lazy evaluation, the lazy-eval world doesn't have this conundrum.)

The "essence-ial" objective is to reach the equivalent of the below imperative statement
sequence, with `ferror` only being called / `ferr` only being written when it clearly _will_
be consumed:

```c
  int n = fwrite(str, 1, str.len, file);
  if (n == str.len)
    return (ok){n};
  int ferr = ferror(file);
  if (ferr == 0)
    return (ok){n};
  return (err){ferr};
```

So at first glance want to reach a `||`-sugarless expression like:

```dart
  (n ->
    (n == str.len) ?- (#ok n) |-
      (ferr -> ferr == 0 ?- #ok n |- #err ferr)
        (libc.ferror file)
  ) (libc.fwrite str 1 str.len file)
```

Or more sugary notation:

```dart
  (libc.fwrite str 1 str.len file) \n\
    (n == str.len) ?- (#ok n) |-
      (libc.ferror file) \ferr\
        ferr == 0 ?- #ok n |- #err ferr
```

But this transformation could realistically only be done for specially-detected
short-circuiting  logical-operators **visibly placed directly inside branch conds**.
The scheme would fail for those stand-alone ones just bound to a name. And they
mustn't imply distinct different operational semantics just based on placement!

This means one of two avenues, to be investigated:
- either need to translate into an SSA / CPS IR instead of a SAPL-like "functional
  (de-lambda'd) IR" right from the start, even during the "only an interpreter" stage 0.
- or stick to the "functional interpreted-IR" but the short-circuiting logical
  operators don't get desugared into branches, instead becoming prim-ops in there.

The former wins at first sight: first, want to treat the short-circuiting sugars
equivalently to hand-written nested-branches (even if they'd hardly ever be) and
this suggests keeping the early desugaring instead of introducing specialized
`@or` / `@and` prim-ops, deferring semantic complications to run time. Secondly,
would need to get there eventually anyway when leaving "interpreter-only" stage0
and embarking upon compilations / transpilations much later on.

Assuming the switch from a functional-ish to an imperative-ish / SSA-ish / CPS-ish
"byte-code" / interpreted-IR, this then suggests a sort of "statically determined
lazy-ish / latest-necessary" placement of SSA assignments at the right locations such
that they won't be computed unless consumed and won't ever "run more than once" by accident,
because our order-independent "locals" are still to behave like mainstream named-vars.

This could turn out tricky indeed!

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
