// For a rough idea, imagine a console screenshot instead of this paste
// ...
//        ~/c/atmo> atmo repl
//
//        ┌─────────────────────────────────────────
//        │:intro
//        └─────────────────────────────────────────
//
//        This is a read-eval-print loop (repl).
//
//        — repl commands start with `:`, any other
//        inputs are eval'd as atmo expressions
//
//        — in case of the latter, a line ending in ,,,
//        introduces or concludes a multi-line input
//
//        — to see --flags, quit and run `atmo help`
//
//
//        ┌─────────────────────────────────────────
//        │:
//        └─────────────────────────────────────────
//
//        Unknown command `:` — try:
//
//            :list ‹kit›
//            :info ‹kit› [‹def›]
//            :srcs ‹kit› ‹def›
//            :quit
//            :intro
//
//        (For usage details on an arg-ful
//        command, invoke it without args.)
//
//
//        ┌─────────────────────────────────────────
//        │:l
//        └─────────────────────────────────────────
//
//        Input `` insufficient for command `:list`.
//
//        Usage:
//
//        :list ‹kit/import/path› ── list defs in the specified kit
//        :list _                 ── list all currently known kits
//
//
//        ┌─────────────────────────────────────────
//        │:l _
//        └─────────────────────────────────────────
//
//        LIST of kits from current search paths:
//        ─── /home/_/c/atmo/kits
//
//        Found 3 kits:
//        ├── [×] omni
//        ├── [×] ·home·_·c·atmo
//        ├── [_] omni/tmp
//
//        Legend: [_] = unloaded, [×] = loaded or load attempted
//        (To see kit details, use `:info ‹kit›`.)
//
//
//        ┌─────────────────────────────────────────
//        │:l ·
//        └─────────────────────────────────────────
//
//        LIST of defs in kit:    `·home·_·c·atmo`
//                found in:    /home/_/c/atmo
//
//        red.at: 10 top-level defs
//        ├── it ─── (line 2)
//        ├── ever ─── (line 4)
//        ├── usefloat ─── (line 6)
//        ├── dafl ─── (line 8)
//        ├── daFl ─── (line 10)
//        ├── leFl ─── (line 12)
//        ├── fn ─── (line 14)
//        ├── fVal ─── (line 16)
//        ├── fval ─── (line 18)
//        ├── test ─── (line 20)
//
//        Total: 10 defs in 1 `*.at` source file
//
//        (To see more details, try also:
//        `:info ·` or `:info · ‹def›`.)
//
//
//        ┌─────────────────────────────────────────
//        │:i ·
//        └─────────────────────────────────────────
//
//        INFO summary on kit:    `·home·_·c·atmo`
//                found in:    /home/_/c/atmo
//
//        1 source file in kit `·`:
//        ├── red.at
//            20 lines (10 sloc), 10 top-level defs, 10 exported
//        Total:
//            20 lines (10 sloc), 10 top-level defs, 10 exported
//            (Counts exclude failed-to-parse defs, if any.)
//
//
//        (To see kit defs, use `:list ·`.)
//
//
//        ┌─────────────────────────────────────────
//        │:s · it
//        └─────────────────────────────────────────
//
//        1 def named `it` found in kit `·home·_·c·atmo`:
//
//
//        ├── /home/_/c/atmo/red.at
//
//        x it :=
//            x
//
//        ├── internal representation:
//
//        it x :=
//            x
//
//
//        ┌─────────────────────────────────────────
//        │:i · it
//        └─────────────────────────────────────────
//
//        1 def named `it` found in kit `·home·_·c·atmo`:
//
//
//        ‹x: ()› » x
//
//
//        ┌─────────────────────────────────────────
//        │ :quit
//        └─────────────────────────────────────────
//
//        ~/c/atmo>
//
package atmo
