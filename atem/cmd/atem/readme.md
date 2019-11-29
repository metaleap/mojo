# atem
--
A simple executable form of the [atem reference interpreter](../../readme.md).
The first (and required) command arg is the `.json` source file for the
`atem.Prog` to `atem.LoadFromJson()`. All further command args are passed on to
the loaded program's `main`.

Since there are no identifiers in `atem` programs, by (hereby decreed)
convention the very last `FuncDef` in the `Prog` is assumed to be the `main` to
run (atem code emitters must ensure this if their outputs are to be run in
here), and is expected to have a `FuncDef.Args` of `len` 2. The first one will
be populated by this interpreter executable with the current process args (sans
the current process executable name and the input `.json` source file path) as a
linked-list of text strings, the second will be the current process environment
variables as a linked-list of `NAME=Value` text strings.
