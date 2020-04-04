usingnamespace @import("./_usingnamespace.zig");

pub var mem_buf: [2 * 1024 * 1024]u8 = undefined;
pub var mem_fba = @import("std").heap.FixedBufferAllocator.init(mem_buf[0..]);
pub var counter: usize = 0;

pub fn alloc(comptime T: type, n: usize) []T {
    // TODO: replace stdlib allocator scheme with homegrown global-fixed-buf
    // before transliterating to first minimal iteration of WIP lang
    return mem_fba.allocator.alloc(T, n) catch |err| fail("{}", .{err});
}

pub fn keep(it: var) *@TypeOf(it) {
    var ptr = &alloc(@TypeOf(it), 1)[0];
    ptr.* = it;
    return ptr;
}

pub fn join(comptime T: type, slices: []const []const T, sep: ?T) []T {
    var len: usize = 0;
    for (slices) |slice, i| {
        if (sep != null and i > 0)
            len += 1;
        len += slice.len;
    }
    var ret = alloc(T, len);
    var idx: usize = 0;
    for (slices) |slice, i_slice| {
        if (sep != null and i_slice > 0) {
            ret[idx] = sep.?;
            idx += 1;
        }
        for (slice) |_, i_char|
            ret[idx + i_char] = slice[i_char];
        idx += slice.len;
    }
    return ret;
}

pub fn concat(comptime T: type, lhs: []const T, rhs: []const T) []T {
    var ret = alloc(T, lhs.len + rhs.len);
    for (lhs) |byte, i|
        ret[i] = byte;
    for (rhs) |byte, i|
        ret[i + lhs.len] = byte;
    return ret;
}

pub fn eql(one: var, two: var) bool {
    if (one.len == two.len) {
        var i: usize = one.len;
        while (i > 0) {
            i -= 1;
            if (one[i] != two[i])
                return false;
        }
        return true;
    }
    return false;
}

pub fn uintToStr(integer: var, base: @TypeOf(integer), min_len: usize, prefix: Str) Str {
    assert(integer >= 0);
    var int = integer;
    var num_digits: usize = 1;
    while (int >= base) : (int /= base)
        num_digits += 1;
    const pad = (num_digits < min_len);
    if (pad)
        num_digits = min_len;
    var ret = alloc(u8, prefix.len + num_digits);
    for (ret) |_, i|
        ret[i] = '0';
    for (prefix) |byte, i|
        ret[i] = byte;
    // 123 / 10     123 % 10            12 / 10     12 % 10
    // =12          =3                  =1          =2
    var idx: usize = ret.len - 1;
    int = integer;
    while (true) : (idx -= 1) {
        defer {
            if (base > 10 and ret[idx] > '9')
                ret[idx] += 7;
        }
        if (int < base) {
            ret[idx] = @intCast(u8, 48 + int);
            break;
        } else {
            ret[idx] = @intCast(u8, 48 + (int % base));
            int /= base;
        }
    }
    return if (pad or prefix.len > 0) ret else ret[idx..];
}
