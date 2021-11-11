Fairly unidiomatic code! Because we want to have a _most_-compact C code base to
later transliterate into our initial (extremely limited) language iteration, we
anticipate the various early-stage limitations and reflect them in this code base:

- no graceful error handling, immediate `panic`s upon detecting a problem
- no 3rd-party imports / deps whatsoever
- no stdlib imports for *core* processing (just for basic program setup & I/O)
  - hence crude minimal own implementations like uintToStr, uintParse, strEql etc.
- use of macros limited to (eventual) WIP-lang meta-programming / generic powers
- all would-be `malloc`s replaced by global fixed-size backing buffer allocation
- no zero-terminated "C strings", all array uses via (macro) `·SliceOf(T)` types

We want here to merely reach the "execute-input-sources-or-die" stage. No bells &
whistles, no *fancy* type stuff, not a lot of syntax sugars. No nifty optimizations,
no proper byte code, will be slow. Wasteful on RAM, no `free`ing. Not even
cross-source-file-imports, too bad. But must focus:

Once the "basic interpreter" destination is reached, build it out to where the
interpreter itself can be fully redone in WIP-lang, interpreter-in-interpreter.
At that point then, this C code base is frozen and done and served its purpose.
All further work proceeds in the very low level limited early initial WIP-lang.
Now worry about compilation to LLVM-IR next before advancing anything else. Thus
the ground itself moved: from a "host language" (like C) to "self-hosted" (LLVM).

With basic (theoretically sub-optimal but practically working) compilation in
place, now rewrite it all once again from scratch but this time "cleanly" &mdash;
focusing on optimal resource usage, real-time perf etc, error handling, REPL,
more optimal IR generation (at least rectifying the already obvious TODOs).

Still no bells or whistles, syntax sugars or type magic yet &mdash; until all the
above is in place properly and well-oiled, robust and sound. I'm impatient too!
But for once, want to reach for the sky on a bedrock-not-quicksand foundation.
