# atem
--
A simple executable form of the [atem reference interpreter](../../readme.md).
The first (and required) command arg is the `.json` source file for the
`atem.Prog` to first `atem.LoadFromJson()` and then run. All further process
args are passed on to the loaded source program's main `FuncDef`.

Since there are no identifiers in `atem` programs, by (hereby decreed)
convention the very last `FuncDef` in the `Prog` is expected to be the one to
run (atem code emitters must ensure this if their outputs are to be run in
here), and is expected to have a `FuncDef.Args` of `len` 2. The first one will
be populated by this interpreter executable with the current process args (sans
the current process executable name and the input `.json` source file path) as a
linked-list of text strings, the second will be the current process environment
variables as a linked-list of `NAME=Value` text strings.

## stdout, stderr, stdin

The main `FuncDef` is by default expected to return a linked list of
`atem.ExprNumInt`s in the range of 0 .. 255, if it does that is considered the
text output to be written to `stdout` and so will it be done. Other returned
`Expr`s will have their `.JsonSrc()` written to `stderr` instead. For source
programs to force extra writes to `stderr` during their run, the `atem.OpPrt`
op-code is to be used. For access to `stdin`, the main `FuncDef` must return a
specific predefined linked-list meeting the following conditions:

- it has 4 elements, in order:

    1. an `ExprFuncRef` (the "handler"),
    2. an `ExprNumInt` in the range of 0 .. 255 (the "separator char"),
    3. any `Expr` (the "initial state"),
    4. and a text string linked list (the "initial output")

- the "handler" must take 2 args, the "previous state" (for the first call this
will be the "initial state" mentioned above) and the "input" (a text string
linked list of any length). It must always return a linked-list of length 2,
with the first element being "next state" (will be passed as is into "handler"
in the next upcoming call) and "output", a text string linked list of any length
incl. zero to be immediately written to `stdout`. If "next state" is returned as
an `ExprFuncRef` pointing to `StdFuncId` (aka. `0`), this indicates to cease
further `stdin` reading and "handler" calling, essentially terminating the
program.

- the "separator char" is 0 to indicate to read in all `stdin` data at once
until EOF before passing it all at once to "handler" in a single and final call.
If it isn't 0, "handler" is called with fresh input whenever in incoming `stdin`
data the "separator char" is next encountered, it's never included in the
"handler" input. So to achieve a typical read-line functionality, one would use
a "separator char" of `10` aka. `'\n'`.

- the "initial state" is what gets passed in the first call to "handler".
Subsequent "handler" calls will instead receive the previous call's returned
"next state" as described above.

- the "initial output", a text string linked list of any length incl. zero, will
be written to `stdout` before the first read from `stdin` and the first call to
"handler".
