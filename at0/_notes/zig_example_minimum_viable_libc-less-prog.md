Snippet from github.com/ziglang/zig/issues/5260

```zig
const msg = "Hello, World!\n";

export fn _start() callconv(.Naked) {
    _ = syscall3(SYS_write, 1, msg, msg.len);
    exit();
}

fn exit() noreturn {
    syscall1(SYS_exit_group, 0);
}

fn syscall1(number: SYS, arg1: usize) usize {
    return asm volatile ("syscall"
        : [ret] "={rax}" (-> usize)
        : [number] "{rax}" (@enumToInt(number)),
          [arg1] "{rdi}" (arg1)
        : "rcx", "r11", "memory"
    );
}

fn syscall3(number: SYS, arg1: usize, arg2: usize, arg3: usize) usize {
    return asm volatile ("syscall"
        : [ret] "={rax}" (-> usize)
        : [number] "{rax}" (@enumToInt(number)),
          [arg1] "{rdi}" (arg1),
          [arg2] "{rsi}" (arg2),
          [arg3] "{rdx}" (arg3)
        : "rcx", "r11", "memory"
    );
}
```
