const std = @import("std");

pub usingnamespace @import("./lang_ast.zig");
pub usingnamespace @import("./tokenizing.zig");
pub usingnamespace @import("./parsing.zig");
pub usingnamespace @import("./llvm_ir.zig");
pub usingnamespace @import("./utils.zig");

pub var stdout: std.fs.File = undefined;

pub const Str = []const u8;
pub const assert = std.debug.assert;
pub const fail = std.debug.panic;
pub const print = std.debug.warn;

pub fn write(bytes: Str) void {
    stdout.writeAll(bytes) catch |err| fail("{}", .{err});
}
