const std = @import("std");
usingnamespace @import("./_usingnamespace.zig");

pub fn main() !void {
    stdout = std.io.getStdOut();
    var proc_args = try std.process.argsAlloc(std.heap.page_allocator);
    var input_src_bytes = try std.fs.cwd().readFileAlloc(std.heap.page_allocator, proc_args[1], 1024 * 1024 * 1024);

    const toks = tokenize(input_src_bytes, false);
    assert(toks.len != 0);

    var ast = parse(toks, input_src_bytes);
    assert(ast.defs.len != 0);
    ast.ensureNameRefs();

    var llmod = llModuleFrom(&ast);
    llmod.emit();
}
