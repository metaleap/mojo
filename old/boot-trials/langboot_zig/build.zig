const std = @import("std");

pub fn build(bld: *std.build.Builder) void {
    const mode = bld.standardReleaseOptions();

    const prog = bld.addExecutable("langboot", "main.zig");
    prog.setBuildMode(mode);
    if (std.os.getenv("PATH")) |env_path|
        if (std.mem.indexOf(u8, env_path, ":/home/_/.local/bin:")) |_| // only locally at my end:
            prog.setOutputDir("/home/_/.local/bin"); // place binary into in-PATH bin dir
    prog.install();
}
