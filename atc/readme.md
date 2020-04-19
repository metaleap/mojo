Fairly unidiomatic code! Because we want to have a _most_-compact C code base to
later transliterate into our initial (extremely limited) language iteration, we
anticipate the various early-stage limitations and reflect them in these sources:

- no proper error handling, immediate `panic`s upon detecting a problem
- no 3rd-party imports / deps whatsoever
- no stdlib imports for *core* processing (just for basic program setup & I/O)
  (hence crude minimal own implementations like uintToStr, uintParse, strEql etc)
- use of macros limited to (eventual) WIP-lang meta-programming / generic powers
- all would-be `malloc`s replaced by global fixed-size backing buffer allocation
- naming / casing conventions follow WIP-lang rather than C customs
- no zero-terminated "C strings", all array uses via (macro) `·SliceOf(T)` types

We want here to merely reach the "execute-input-sources-or-die" stage. No bells &
whistles, no *fancy* type stuff, no syntax sugars (not even operators, we endure
prim calls). No nifty optimizations, no proper byte code, will be slow. Wasteful
on RAM too, no freeing. Once the destination is reached, build it out to where it
can be fully redone in WIP-lang, interpreter-in-interpreter. At that point then,
worry about compilation next before advancing anything else. Thus the foundation
itself can move from a "host language" (like C) to "self-hosted" (aka. LLVM-IR).

At that point then, months from now, rewrite once again from scratch but this
time "cleanly" &mdash; focusing on optimal resource usage, real-time perf etc.
