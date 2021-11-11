usingnamespace @import("./_usingnamespace.zig");

pub const LLModule = struct {
    target_datalayout: Str = "e-m:e-i64:64-f80:128-n8:16:32:64-S128",
    target_triple: Str = "x86_64-unknown-linux-gnu",
    globals: []LLGlobal,
    funcs: []LLFunc,

    pub fn emit(me: *LLModule) void {
        write("\ntarget datalayout = \"");
        write(me.target_datalayout);
        write("\"\ntarget triple = \"");
        write(me.target_triple);
        write("\"\n\n");
        for (me.globals) |*global| {
            global.emit();
            write("\n");
        }
        write("\n");
        for (me.funcs) |*func| {
            func.emit();
            write("\n");
        }
    }
};

pub const LLGlobal = struct {
    name: Str,
    constant: bool = false,
    external: bool = false,
    ty: LLType,
    initializer: ?LLExpr = null,

    fn emit(me: *LLGlobal) void {
        write("@");
        write(me.name);
        write(if (me.constant) " = constant " else if (me.external) " = external global " else " = global ");
        me.ty.emit();
        if (me.initializer) |*init| {
            write(" ");
            init.emit();
        }
    }
};

pub const LLFunc = struct {
    ty: LLType,
    name: Str,
    args: []Param,
    last_block: ?*Block = null,

    fn addBlock(me: *LLFunc, block: Block, prepend: bool) *Block {
        var ptr_block = keep(block);
        var maybe_block: ?*Block = me.last_block;
        if (maybe_block == null)
            me.last_block = ptr_block
        else if (!prepend) {
            ptr_block.prev_block = me.last_block;
            me.last_block = ptr_block;
        } else while (maybe_block) |this_block| {
            maybe_block = this_block.prev_block;
            if (maybe_block == null)
                this_block.prev_block = ptr_block;
        }
        return ptr_block;
    }

    fn emit(me: *LLFunc) void {
        write(if (me.last_block == null) "declare " else "\n\ndefine ");
        me.ty.emit();
        write(" @");
        write(me.name);
        write("(");
        for (me.args) |*param, i| {
            if (i > 0)
                write(", ");
            param.ty.emit();
            if (param.name) |name| {
                write(" %");
                write(name);
            }
        }
        write(")");
        if (me.last_block) |last_block| {
            write(" {\n");
            for (last_block.toSlice()) |block|
                block.emit();
            write("}\n");
        }
    }

    const FindLocalAliasLet = struct { stmt: *Stmt, block: *Block, name_let: Str, name_ref: Str };
    fn findLocalAliasLet(me: *const LLFunc) ?FindLocalAliasLet {
        var maybe_block = me.last_block;
        while (maybe_block) |block| : (maybe_block = block.prev_block) {
            var maybe_stmt = block.last_stmt;
            while (maybe_stmt) |stmt| : (maybe_stmt = stmt.prev_stmt) {
                switch (stmt.kind) {
                    else => {},
                    .let => |let| if (let.expr == .local) {
                        return FindLocalAliasLet{ .stmt = stmt, .block = block, .name_let = let.name, .name_ref = let.expr.local };
                    },
                }
            }
        }
        return null;
    }

    const FindLocalLitIntLet = struct { stmt: *Stmt, block: *Block, name_let: Str, lit_int: u128 };
    fn findLocalLitIntLet(me: *const LLFunc) ?FindLocalLitIntLet {
        var maybe_block = me.last_block;
        while (maybe_block) |block| : (maybe_block = block.prev_block) {
            var maybe_stmt = block.last_stmt;
            while (maybe_stmt) |stmt| : (maybe_stmt = stmt.prev_stmt) {
                switch (stmt.kind) {
                    else => {},
                    .let => |let| if (let.expr == .lit_int)
                        return FindLocalLitIntLet{ .stmt = stmt, .block = block, .name_let = let.name, .lit_int = let.expr.lit_int }
                    else if (let.expr == .bin_op and let.expr.bin_op.kind == .add and
                        let.expr.bin_op.op1 == .lit_int and let.expr.bin_op.op1.lit_int == 0 and
                        let.expr.bin_op.op2 == .lit_int)
                        return FindLocalLitIntLet{ .stmt = stmt, .block = block, .name_let = let.name, .lit_int = let.expr.bin_op.op2.lit_int },
                }
            }
        }
        return null;
    }

    fn rewriteLocalRefsAndDropAliasLet(me: *LLFunc, it: *FindLocalAliasLet) void {
        var maybe_block = me.last_block;
        while (maybe_block) |block| : (maybe_block = block.prev_block) {
            if (block.last_stmt != null and block.last_stmt.? == it.stmt)
                unreachable;
            var maybe_stmt = block.last_stmt;
            while (maybe_stmt) |stmt| : (maybe_stmt = stmt.prev_stmt) {
                if (stmt != it.stmt)
                    _ = stmt.rewriteLocalRefs(it.name_let, LLExpr{ .local = it.name_ref });
                if (stmt.prev_stmt != null and stmt.prev_stmt.? == it.stmt)
                    stmt.prev_stmt = it.stmt.prev_stmt;
            }
        }
    }

    fn rewriteLocalRefsAndDropLitIntLet(me: *LLFunc, it: *FindLocalLitIntLet) void {
        var maybe_block = me.last_block;
        while (maybe_block) |block| : (maybe_block = block.prev_block) {
            if (block.last_stmt != null and block.last_stmt.? == it.stmt)
                unreachable;
            var maybe_stmt = block.last_stmt;
            while (maybe_stmt) |stmt| : (maybe_stmt = stmt.prev_stmt) {
                if (stmt != it.stmt)
                    _ = stmt.rewriteLocalRefs(it.name_let, LLExpr{ .lit_int = it.lit_int });
                if (stmt.prev_stmt != null and stmt.prev_stmt.? == it.stmt)
                    stmt.prev_stmt = it.stmt.prev_stmt;
            }
        }
    }

    pub const Param = struct {
        name: ?Str = null,
        ty: LLType,
    };

    pub const Block = struct {
        name: Str,
        prev_block: ?*Block = null,
        last_stmt: ?*Stmt = null,
        fn emit(me: *Block) void {
            write(me.name);
            write(":\n");
            if (me.last_stmt) |last_stmt| for (last_stmt.toSlice()) |stmt|
                stmt.emit();
        }
        fn root(me: *Block) *Block {
            return if (me.prev_block) |prev_block| prev_block.root() else me;
        }
        fn addStmt(me: *Block, stmt: Stmt, prepend: bool) *Stmt {
            var ptr_stmt = keep(stmt);
            var maybe_stmt: ?*Stmt = me.last_stmt;
            if (maybe_stmt == null)
                me.last_stmt = ptr_stmt
            else if (!prepend) {
                ptr_stmt.prev_stmt = me.last_stmt;
                me.last_stmt = ptr_stmt;
            } else while (maybe_stmt) |this_stmt| {
                maybe_stmt = this_stmt.prev_stmt;
                if (maybe_stmt == null)
                    this_stmt.prev_stmt = ptr_stmt;
            }
            return ptr_stmt;
        }
        fn toSlice(me: *Block) []*Block {
            var num_blocks: usize = 0;
            var maybe_block: ?*Block = me;
            while (maybe_block) |block| {
                num_blocks += 1;
                maybe_block = block.prev_block;
            }
            var ret = alloc(*Block, num_blocks);
            maybe_block = me;
            while (maybe_block) |block| {
                num_blocks -= 1;
                ret[num_blocks] = block;
                maybe_block = block.prev_block;
            }
            return ret;
        }
    };

    pub const Stmt = struct {
        prev_stmt: ?*Stmt = null,
        kind: Kind,

        fn toSlice(me: *Stmt) []*Stmt {
            var num_stmts: usize = 0;
            var maybe_stmt: ?*Stmt = me;
            while (maybe_stmt) |stmt| {
                num_stmts += 1;
                maybe_stmt = stmt.prev_stmt;
            }
            var ret = alloc(*Stmt, num_stmts);
            maybe_stmt = me;
            while (maybe_stmt) |stmt| {
                num_stmts -= 1;
                ret[num_stmts] = stmt;
                maybe_stmt = stmt.prev_stmt;
            }
            return ret;
        }
        fn root(me: *Stmt) *Stmt {
            return if (me.prev_stmt) |prev_stmt| prev_stmt.root() else me;
        }
        fn emit(me: *Stmt) void {
            switch (me.kind) {
                .br => |br| {
                    write("  br label %");
                    write(br);
                    write("\n");
                },
                .let => |let| {
                    write("  %");
                    write(let.name);
                    write(" = ");
                    let.expr.emit();
                    write("\n");
                },
                .ret => |ret| {
                    write("  ret ");
                    ret.emit();
                    write("\n");
                },
                .sw => |sw| {
                    write("  switch ");
                    sw.scrut.emit();
                    write(", label %");
                    write(sw.label_default);
                    write(" [");
                    for (sw.cases) |*case, i| {
                        case.expr.emit();
                        write(", label %");
                        write(case.label);
                        if (i > 0)
                            write(",\n    ");
                    }
                    write("]\n");
                },
            }
        }
        fn rewriteLocalRefs(me: *Stmt, name_old: Str, expr_new: LLExpr) usize {
            var num_rewrites: usize = 0;
            switch (me.kind) {
                else => {},
                .let => |let| num_rewrites += let.expr.rewriteLocalRefs(name_old, expr_new),
                .ret => |ret| num_rewrites += ret.expr.rewriteLocalRefs(name_old, expr_new),
                .sw => |sw| {
                    num_rewrites += sw.scrut.expr.rewriteLocalRefs(name_old, expr_new);
                    for (sw.cases) |*case| {
                        num_rewrites += case.expr.expr.rewriteLocalRefs(name_old, expr_new);
                    }
                },
            }
            return num_rewrites;
        }

        const Kind = union(enum) {
            let: *Let,
            ret: *Ret,
            sw: *Switch,
            br: Str,

            pub const Ret = LLExpr.Typed;
            pub const Let = struct {
                name: Str,
                expr: LLExpr,
            };
            pub const Switch = struct {
                scrut: LLExpr.Typed,
                label_default: Str,
                cases: []Case,
                pub const Case = struct {
                    expr: LLExpr.Typed,
                    label: Str,
                };
            };
        };
    };
};

pub const LLExpr = union(enum) {
    local: Str,
    global: Str,
    lit_int: u128,
    lit_str: Str,
    lit_bool: bool,
    lit_null,
    phi: *Phi,
    typed: *Typed,
    alloca: *Alloca,
    load: *Load,
    call: *Call,
    gep: *Gep,
    icmp: *Icmp,
    bin_op: *BinOp,

    fn emit(me: *LLExpr) void {
        switch (me.*) {
            else => fail("TODO: emit for '{}'", .{@tagName(me.*)}),
            .phi => |phi| {
                phi.emit();
            },
            .local => |ident| {
                write("%");
                write(ident);
            },
            .global => |ident| {
                write("@");
                write(ident);
            },
            .lit_int => |lit_int| {
                write(uintToStr(lit_int, 10, 1, ""));
            },
            .lit_str => |lit_str| {
                write("c\"");
                for (lit_str) |byte| {
                    if (byte >= 32 and byte < 127)
                        write(&[_]u8{byte})
                    else {
                        write("\\");
                        write(uintToStr(byte, 16, 2, ""));
                    }
                }
                write("\"");
            },
            .bin_op => |bin_op| {
                write(@tagName(bin_op.kind));
                write(" ");
                bin_op.ty.emit();
                write(" ");
                bin_op.op1.emit();
                write(", ");
                bin_op.op2.emit();
            },
            .icmp => |icmp| {
                write("icmp ");
                write(@tagName(icmp.kind));
                write(" ");
                icmp.ty.emit();
                write(" ");
                icmp.op1.emit();
                write(", ");
                icmp.op2.emit();
            },
            .call => |call| {
                write("call ");
                call.func.emit();
                write("(");
                for (call.args) |*arg, i| {
                    if (i > 0)
                        write(",\n    ");
                    arg.emit();
                }
                write(")");
            },
            .alloca => |alloca| {
                write("alloca ");
                alloca.ty.emit();
                write(", ");
                alloca.num_elems.emit();
            },
            .load => |load| {
                write("load ");
                load.ty.emit();
                write(", ");
                load.ty.emit();
                write("* ");
                load.expr.emit();
            },
            .gep => |gep| {
                write("getelementptr ");
                gep.ty.emit();
                write(", ");
                gep.base_ptr.emit();
                for (gep.idxs) |*typed| {
                    write(", ");
                    typed.emit();
                }
            },
        }
    }

    fn rewriteLocalRefs(me: *LLExpr, name_old: Str, expr_new: LLExpr) usize {
        var num_rewrites: usize = 0;
        switch (me.*) {
            else => {},
            .local => |local| {
                if (eql(local, name_old)) {
                    me.* = expr_new;
                    num_rewrites += 1;
                }
            },
            .phi => |phi| for (phi.preds) |*pred| {
                num_rewrites += pred.expr.rewriteLocalRefs(name_old, expr_new);
            },
            .typed => |typed| num_rewrites += typed.expr.rewriteLocalRefs(name_old, expr_new),
            .load => |load| num_rewrites += load.expr.rewriteLocalRefs(name_old, expr_new),
            .alloca => |alloca| num_rewrites += alloca.num_elems.expr.rewriteLocalRefs(name_old, expr_new),
            .gep => |gep| {
                num_rewrites += gep.base_ptr.expr.rewriteLocalRefs(name_old, expr_new);
                for (gep.idxs) |*idx|
                    num_rewrites += idx.expr.rewriteLocalRefs(name_old, expr_new);
            },
            .call => |call| {
                num_rewrites += call.func.expr.rewriteLocalRefs(name_old, expr_new);
                for (call.args) |*arg|
                    num_rewrites += arg.expr.rewriteLocalRefs(name_old, expr_new);
            },
            .icmp => |icmp| {
                num_rewrites += icmp.op1.rewriteLocalRefs(name_old, expr_new);
                num_rewrites += icmp.op2.rewriteLocalRefs(name_old, expr_new);
            },
            .bin_op => |bin_op| {
                num_rewrites += bin_op.op1.rewriteLocalRefs(name_old, expr_new);
                num_rewrites += bin_op.op2.rewriteLocalRefs(name_old, expr_new);
            },
        }
        return num_rewrites;
    }

    pub const Phi = struct {
        ty: LLType,
        preds: []Pred,
        fn emit(me: *Phi) void {
            write("phi ");
            me.ty.emit();
            for (me.preds) |*pred, i| {
                if (i > 0)
                    write(",");
                write(" [");
                pred.expr.emit();
                write(", %");
                write(pred.block);
                write("]");
            }
        }
        const Pred = struct {
            expr: LLExpr,
            block: Str,
        };
    };
    pub const Typed = struct {
        ty: LLType,
        expr: LLExpr,
        fn emit(me: *Typed) void {
            me.ty.emit();
            write(" ");
            me.expr.emit();
        }
    };
    pub const Alloca = struct {
        ty: LLType,
        num_elems: LLExpr.Typed,
    };
    pub const Load = LLExpr.Typed;
    pub const Call = struct {
        func: LLExpr.Typed,
        args: []LLExpr.Typed,
    };
    pub const Gep = struct {
        ty: LLType,
        base_ptr: LLExpr.Typed,
        idxs: []LLExpr.Typed,
    };
    pub const BinOp = struct {
        kind: enum {
            add,
        },
        ty: LLType,
        op1: LLExpr,
        op2: LLExpr,
    };
    pub const Icmp = struct {
        kind: Kind,
        ty: LLType,
        op1: LLExpr,
        op2: LLExpr,
        const Kind = enum {
            eq,
            ne,
            ugt,
            uge,
            ult,
            ule,
            sgt,
            sge,
            slt,
            sle,
        };
    };
};

pub const LLType = union(enum) {
    i: u23,
    p: *LLType,
    a: Arr,
    s: []LLType,
    n: Str,
    f: []LLType,

    const Arr = struct {
        size: usize,
        of: *LLType,
    };

    fn emit(me: *LLType) void {
        switch (me.*) {
            else => fail("TODO: impl LLType.emit() for: {}", .{me.*}),
            .p => |ptr_of| {
                ptr_of.emit();
                write("*");
            },
            .i => |int| {
                write("i");
                write(uintToStr(int, 10, 1, ""));
            },
            .a => |arr| {
                write("[");
                write(uintToStr(arr.size, 10, 1, ""));
                write(" x ");
                arr.of.emit();
                write("]");
            },
        }
    }
};

fn shouldTopDefBeFuncDef(top_def: *AstDef) bool {
    return switch (top_def.body.kind) {
        else => false,
        .builtin => |builtin| return builtin.* == .fn_def,
    };
}

pub fn llModuleFrom(ast: *Ast) LLModule {
    var ret = LLModule{
        .globals = alloc(LLGlobal, ast.defs.len),
        .funcs = alloc(LLFunc, ast.defs.len),
    };
    var num_globals: usize = 0;
    var num_funcs: usize = 0;

    for (ast.defs) |*top_def| {
        const name = top_def.name();
        if (shouldTopDefBeFuncDef(top_def)) {
            const fn_def = &top_def.body.kind.builtin.fn_def;
            var llf = &ret.funcs[num_funcs];
            num_funcs += 1;
            llf.* = .{
                .name = name,
                .ty = llType(fn_def.ty, null),
                .args = alloc(LLFunc.Param, fn_def.args.len),
            };
            for (fn_def.args) |*fn_arg, i|
                llf.args[i] = .{ .name = fn_arg.name.kind.ident, .ty = llType(fn_arg.ty, null) };
            llGenFuncBody(llf, top_def, fn_def, ast);
        } else switch (top_def.body.kind) {
            .lit_str => |lit_str| {
                const llg = &ret.globals[num_globals];
                num_globals += 1;
                llg.* = .{
                    .constant = true,
                    .name = name,
                    .initializer = .{ .lit_str = lit_str },
                    .ty = LLType{ .a = .{ .size = lit_str.len, .of = keep(LLType{ .i = 8 }) } },
                };
            },
            .builtin => |builtin| switch (builtin.*) {
                else => unreachable,
                .ext_var => |*ext_var| {
                    const llg = &ret.globals[num_globals];
                    num_globals += 1;
                    llg.* = .{ .external = true, .name = ext_var.name, .ty = llType(ext_var.ty, null) };
                },
                .ext_fn => |*ext_fn| {
                    const llf = &ret.funcs[num_funcs];
                    num_funcs += 1;
                    llf.* = .{ .name = ext_fn.name, .ty = llType(ext_fn.ty, null), .args = alloc(LLFunc.Param, ext_fn.args.len) };
                    for (ext_fn.args) |ty, i|
                        llf.args[i] = .{ .ty = llType(ty, null) };
                },
            },
            else => fail("cannot make global from top-def with body expr of kind {}", .{top_def.body.kind}),
        }
    }
    print("\n\n~=~=~=~=~=~=~=~=~=~=~=~=~=~=~=~=~\n\n", .{});

    ret.globals = ret.globals[0..num_globals];
    ret.funcs = ret.funcs[0..num_funcs];
    return ret;
}

fn llType(ty: AstExpr.Builtin.Ty, fallback_for_unsized_int: ?LLType) LLType {
    return switch (ty) {
        .int => |bit_width| if (bit_width) |bw|
            LLType{ .i = bw }
        else
            (if (fallback_for_unsized_int) |t| t else unreachable),
        .arr => |arr| LLType{ .a = .{ .size = arr.size, .of = keep(llType(arr.of.*, null)) } },
        .ptr => |maybe_of| if (maybe_of) |ptr_of|
            LLType{ .p = keep(llType(ptr_of.*, null)) }
        else
            LLType{ .p = keep(LLType{ .i = 8 }) },
    };
}

fn llGenFuncBody(llf: *LLFunc, top_def: *AstDef, fn_def: *AstExpr.Builtin.Fn, ast: *Ast) void {
    var ctx = LLGenFuncBody{
        .ast = ast,
        .def_top = top_def,
        .def_cur = top_def,
        .llf = llf,
        .fn_def = fn_def,
    };
    const ret_type = llType(fn_def.ty, null);
    llf.last_block = keep(LLFunc.Block{ .name = ".return" });
    ctx.genLocalFromExpr(&fn_def.expr, &LLGenFuncBody.Dst{
        .name = "ret",
        .ty = ret_type,
        .block = llf.last_block.?,
    });
    _ = llf.last_block.?.addStmt(LLFunc.Stmt{
        .kind = .{
            .ret = keep(LLExpr.Typed{
                .ty = ret_type,
                .expr = LLExpr{ .local = "ret" },
            }),
        },
    }, false);

    while (llf.findLocalAliasLet()) |*found|
        llf.rewriteLocalRefsAndDropAliasLet(found);
    while (llf.findLocalLitIntLet()) |*found|
        llf.rewriteLocalRefsAndDropLitIntLet(found);
}

const LLGenFuncBody = struct {
    ast: *Ast,
    def_top: *AstDef,
    def_cur: *AstDef,
    llf: *LLFunc,
    fn_def: *AstExpr.Builtin.Fn,
    locals_done: [][2]Str = &[_][2]Str{},
    dyn_name_pref: Str = "",

    const Dst = struct {
        name: Str,
        ty: LLType,
        block: *LLFunc.Block,
    };

    fn isLocalDone(me: *LLGenFuncBody, name: Str) ?Str {
        for (me.locals_done) |pair|
            if (eql(name, pair[0]))
                return pair[1];
        return null;
    }

    fn genLocalFromExpr(me: *LLGenFuncBody, expr: *AstExpr, dst: *const Dst) void {
        assert(dst.name.len != 0);
        var dst_expr: LLExpr = switch (expr.kind) {
            else => fail("TODO: generate IR for exprs of '{}'", .{@tagName(expr.kind)}),
            .builtin => me.genLocalFromExprBuiltin(expr, dst),
            .lit_int => me.genLocalFromExprLitInt(expr, dst),
            .ident => me.genLocalFromExprIdent(expr, dst),
        };
        _ = dst.block.addStmt(.{
            .kind = .{
                .let = keep(LLFunc.Stmt.Kind.Let{
                    .name = dst.name,
                    .expr = dst_expr,
                }),
            },
        }, false);
    }

    fn genLocalFromExprIdent(me: *LLGenFuncBody, expr: *AstExpr, dst: *const Dst) LLExpr {
        var ident = expr.kind.ident;
        const ref_def = me.def_cur.scope.mustResolve(ident, null).def;
        if (ref_def.is_top_def) {
            switch (ref_def.body.kind) {
                else => fail("genLocalFromExprIdent with is_top_def of {}", .{ref_def.body.kind}),
                .builtin => |builtin| switch (builtin.*) {
                    else => fail("genLocalFromExprIdent with is_top_def of builtin {}", .{@tagName(builtin.*)}),
                    .ext_fn => |ext_fn| ident = ext_fn.name,
                    .ext_var => |ext_var| {
                        return LLExpr{
                            .load = keep(LLExpr.Load{
                                .ty = llType(ext_var.ty, null),
                                .expr = LLExpr{ .global = ext_var.name },
                            }),
                        };
                    },
                },
                .lit_str => |lit_str| {
                    const idxs = alloc(LLExpr.Typed, 2);
                    idxs[0] = .{ .ty = .{ .i = 64 }, .expr = LLExpr{ .lit_int = 0 } };
                    idxs[1] = idxs[0];
                    const arr_type = LLType{ .a = LLType.Arr{ .size = lit_str.len, .of = dst.ty.p } };
                    return LLExpr{
                        .gep = keep(LLExpr.Gep{
                            .ty = arr_type,
                            .base_ptr = .{
                                .ty = .{ .p = keep(arr_type) },
                                .expr = LLExpr{ .global = ident },
                            },
                            .idxs = idxs,
                        }),
                    };
                },
            }
            return LLExpr{ .global = ident };
        } else {
            var name_local: Str = join(u8, &[_]Str{ me.dyn_name_pref, ident }, '_');
            if (me.isLocalDone(ident)) |name_local_used|
                name_local = name_local_used
            else {
                me.locals_done = concat([2]Str, me.locals_done, &[_][2]Str{[2]Str{ ident, name_local }});

                const dyn_name_pref = me.dyn_name_pref;
                defer me.dyn_name_pref = dyn_name_pref;
                me.dyn_name_pref = name_local;

                const def_cur = me.def_cur;
                defer me.def_cur = def_cur;
                me.def_cur = ref_def;

                me.genLocalFromExpr(&ref_def.body, &Dst{
                    .name = name_local,
                    .ty = if (ref_def.ty(me.ast)) |ty|
                        llType(ty, dst.ty)
                    else
                        dst.ty,
                    .block = dst.block,
                });
            }
            return LLExpr{ .local = name_local };
        }
    }

    fn genLocalFromExprLitInt(me: *LLGenFuncBody, expr: *AstExpr, dst: *const Dst) LLExpr {
        const lit_int = expr.kind.lit_int;
        return LLExpr{
            .bin_op = keep(LLExpr.BinOp{
                .kind = .add,
                .ty = dst.ty,
                .op1 = LLExpr{ .lit_int = 0 },
                .op2 = LLExpr{ .lit_int = lit_int },
            }),
        };
    }

    fn genLocalFromExprBuiltin(me: *LLGenFuncBody, expr: *AstExpr, dst: *const Dst) LLExpr {
        const builtin = expr.kind.builtin;
        return switch (builtin.*) {
            else => fail("TODO: generate IR for 'builtin'-exprs of '{}'", .{@tagName(builtin.*)}),
            .atom => me.genLocalFromExprBuiltinAtom(expr, dst),
            .case => me.genLocalFromExprBuiltinCase(expr, dst),
            .bin_op => me.genLocalFromExprBuiltinBinOp(expr, dst),
            .ext_call => me.genLocalFromExprBuiltinExtCall(expr, dst),
            .alloca => me.genLocalFromExprBuiltinAlloca(expr, dst),
            .len => me.genLocalFromExprBuiltinLen(expr, dst),
        };
    }

    fn genLocalFromExprBuiltinAlloca(me: *LLGenFuncBody, expr: *AstExpr, dst: *const Dst) LLExpr {
        const alloca = &expr.kind.builtin.alloca;
        const name_num_elems = concat(u8, dst.name, ".n");
        const type_num_elems = if (alloca.num_elems.maybeType(me.def_cur, me.ast)) |ty|
            llType(ty, LLType{ .i = 32 })
        else
            LLType{ .i = 32 };
        me.genLocalFromExpr(&alloca.num_elems, &Dst{
            .name = name_num_elems,
            .block = dst.block,
            .ty = type_num_elems,
        });
        return LLExpr{
            .alloca = keep(LLExpr.Alloca{
                .ty = llType(alloca.ty, null),
                .num_elems = LLExpr.Typed{
                    .ty = type_num_elems,
                    .expr = LLExpr{ .local = name_num_elems },
                },
            }),
        };
    }

    fn genLocalFromExprBuiltinAtom(me: *LLGenFuncBody, expr: *AstExpr, dst: *const Dst) LLExpr {
        return LLExpr{
            .bin_op = keep(LLExpr.BinOp{
                .kind = .add,
                .ty = LLType{ .i = 1 },
                .op1 = LLExpr{ .lit_int = 0 },
                .op2 = switch (expr.kind.builtin.atom) {
                    else => unreachable,
                    .@"true" => LLExpr{ .lit_int = 1 },
                    .@"false" => LLExpr{ .lit_int = 0 },
                },
            }),
        };
    }

    fn genLocalFromExprBuiltinLen(me: *LLGenFuncBody, expr: *AstExpr, dst: *const Dst) LLExpr {
        const name = concat(u8, dst.name, ".len");
        const len = @intCast(u128, switch (expr.kind.builtin.len.kind) {
            else => fail("cannot /len of {}", .{@tagName(expr.kind.builtin.len.kind)}),
            .lit_str => |lit_str| lit_str.len,
            .ident => |ident| resolve: {
                const ref_def = me.def_cur.scope.resolve(ident, null).?.def;
                switch (ref_def.body.kind) {
                    else => fail("cannot /len of {}", .{@tagName(ref_def.body.kind)}),
                    .lit_str => |lit_str| break :resolve lit_str.len,
                }
            },
        });
        me.genLocalFromExpr(
            &AstExpr{ .base = AstNode.zero, .kind = .{ .lit_int = len }, .ty = AstExpr.Builtin.Ty{ .int = 64 } },
            &Dst{ .name = name, .block = dst.block, .ty = LLType{ .i = 64 } },
        );
        return LLExpr{ .local = name };
    }

    fn genLocalFromExprBuiltinExtCall(me: *LLGenFuncBody, expr: *AstExpr, dst: *const Dst) LLExpr {
        const ext_call = &expr.kind.builtin.ext_call;
        const ident_callee = ext_call.callee.kind.ident;
        const ref_def = me.def_cur.scope.resolve(ident_callee, null).?.def;
        const ext_fn = &ref_def.body.kind.builtin.ext_fn;
        const call = keep(LLExpr.Call{
            .func = LLExpr.Typed{
                .expr = LLExpr{ .global = ext_fn.name },
                .ty = if (ref_def.ty(me.ast)) |ty| llType(ty, null) else dst.ty,
            },
            .args = alloc(LLExpr.Typed, ext_call.args.len),
        });
        assert(ext_call.args.len == ext_fn.args.len);
        const suff_name_arg = concat(u8, dst.name, ".a");
        for (ext_call.args) |*arg, i| {
            const arg_type = llType(ext_fn.args[i], null);
            const name_arg = uintToStr(i, 10, 1, suff_name_arg);
            me.genLocalFromExpr(arg, &Dst{ .name = name_arg, .ty = arg_type, .block = dst.block });
            call.args[i] = LLExpr.Typed{
                .ty = arg_type,
                .expr = LLExpr{ .local = name_arg },
            };
        }
        return LLExpr{ .call = call };
    }

    fn genLocalFromExprBuiltinBinOp(me: *LLGenFuncBody, expr: *AstExpr, dst: *const Dst) LLExpr {
        const bin_op = &expr.kind.builtin.bin_op;
        const op_kind = switch (bin_op.kind) {
            .eq => LLExpr.Icmp.Kind.eq,
            .ne => LLExpr.Icmp.Kind.ne,
            .sgt => LLExpr.Icmp.Kind.sgt,
            .sge => LLExpr.Icmp.Kind.sge,
        };
        var ty_args = llType(bin_op.op1.maybeType(me.def_cur, me.ast) orelse
            bin_op.op2.mustType(me.def_cur, me.ast), null);

        counter += 1;
        const name_op1: Str = join(u8, &[_]Str{ me.dyn_name_pref, uintToStr(counter, 10, 1, @tagName(op_kind)), "l" }, '.');
        const name_op2: Str = join(u8, &[_]Str{ me.dyn_name_pref, uintToStr(counter, 10, 1, @tagName(op_kind)), "r" }, '.');
        me.genLocalFromExpr(&bin_op.op1, &Dst{ .name = name_op1, .block = dst.block, .ty = ty_args });
        me.genLocalFromExpr(&bin_op.op2, &Dst{ .name = name_op2, .block = dst.block, .ty = ty_args });
        return LLExpr{
            .icmp = keep(LLExpr.Icmp{
                .ty = ty_args,
                .op1 = LLExpr{ .local = name_op1 },
                .op2 = LLExpr{ .local = name_op2 },
                .kind = op_kind,
            }),
        };
    }

    fn genLocalFromExprBuiltinCase(me: *LLGenFuncBody, expr: *AstExpr, dst: *const Dst) LLExpr {
        const case = &expr.kind.builtin.case;

        var blocks = alloc(*LLFunc.Block, 1 + case.prongs.len);
        var ret = LLExpr.Phi{ .ty = dst.ty, .preds = alloc(LLExpr.Phi.Pred, case.prongs.len) };

        const dyn_name_pref = me.dyn_name_pref;
        defer me.dyn_name_pref = dyn_name_pref;
        me.dyn_name_pref = join(u8, &[_]Str{ me.dyn_name_pref, "c" }, '.');

        // generate empty blocks for all prongs
        for (case.prongs) |*case_prong, i|
            blocks[i] = me.llf.addBlock(
                .{ .name = uintToStr(i, 10, 1, concat(u8, me.dyn_name_pref, ".")) },
                true,
            );

        const scrut_type = llType(case.scrut.mustType(me.def_cur, me.ast), null);
        const switch_cases = alloc(LLFunc.Stmt.Kind.Switch.Case, case.prongs.len - 1);
        { // generate empty block for scrutinee
            const blk_scrut = me.llf.addBlock(.{ .name = me.dyn_name_pref }, true);
            blocks[blocks.len - 1] = blk_scrut;
            {
                var idx: usize = 0;
                for (case.prongs) |*case_prong, i| {
                    if (i != case.default_idx.?) {
                        const name_case = concat(u8, blocks[i].name, ".alt");
                        switch_cases[idx] = .{
                            .label = blocks[i].name,
                            .expr = LLExpr.Typed{
                                .ty = scrut_type,
                                .expr = LLExpr{ .local = name_case },
                            },
                        };
                        me.genLocalFromExpr(&case_prong.lhs, &Dst{
                            .name = name_case,
                            .block = blk_scrut,
                            .ty = scrut_type,
                        });
                    }
                    idx += 1;
                }
            }
        }

        { // populate empty block for scrutinee
            const blk_scrut = blocks[blocks.len - 1];
            const name_scrut = concat(u8, me.dyn_name_pref, ".scrut");
            me.genLocalFromExpr(&case.scrut, &Dst{
                .name = name_scrut,
                .block = blk_scrut,
                .ty = scrut_type,
            });
            _ = blk_scrut.addStmt(.{
                .kind = .{
                    .sw = keep(LLFunc.Stmt.Kind.Switch{
                        .label_default = blocks[case.default_idx.?].name,
                        .cases = switch_cases,
                        .scrut = LLExpr.Typed{
                            .ty = scrut_type,
                            .expr = LLExpr{ .local = name_scrut },
                        },
                    }),
                },
            }, false);
        }

        // populate empty blocks for all prongs
        for (case.prongs) |*case_prong, i| {
            const name_case_prong_result = concat(u8, blocks[i].name, ".result");
            ret.preds[i] = .{ .expr = .{ .local = name_case_prong_result }, .block = blocks[i].name };
            me.genLocalFromExpr(&case_prong.rhs, &Dst{
                .name = name_case_prong_result,
                .ty = dst.ty,
                .block = blocks[i],
            });
            _ = blocks[i].addStmt(.{ .kind = .{ .br = dst.block.name } }, false);
        }

        return LLExpr{ .phi = keep(ret) };
    }
};
