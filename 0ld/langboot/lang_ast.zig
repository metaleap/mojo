usingnamespace @import("./_usingnamespace.zig");

pub const Ast = struct {
    src: Str,
    toks: []Token,
    defs: []AstDef,
    scope: Scopes = Scopes{},

    pub fn ensureNameRefs(me: *Ast) void {
        me.scope.cur = alloc(NameRef, me.defs.len);
        for (me.defs) |*def, i|
            me.scope.cur[i] = .{ .name = def.name(), .refers_to = .{ .def = def } };
        for (me.defs) |*def|
            def.ensureNameRefs(me, &me.scope);
    }

    const Scopes = struct {
        cur: []NameRef = &[_]NameRef{},
        parent: ?*Scopes = null,

        fn resolve(me: *Scopes, name: Str, only_until_before_idx: ?usize) ?RefTo {
            for (me.cur) |*it, i| {
                if (only_until_before_idx != null and i == only_until_before_idx.?)
                    break
                else if (eql(it.name, name))
                    return it.refers_to;
            }
            if (me.parent) |parent|
                return parent.resolve(name, null);
            return null;
        }

        fn mustResolve(me: *Scopes, name: Str, only_until_before_idx: ?usize) RefTo {
            return me.resolve(name, only_until_before_idx) orelse fail("unresolvable identifier '{}'", .{name});
        }
    };

    const NameRef = struct {
        name: Str,
        refers_to: RefTo,
    };

    const RefTo = union(enum) {
        def: *AstDef,
    };
};

pub const AstNode = struct {
    toks_idx: usize,
    toks_len: usize,

    pub const zero = AstNode{ .toks_idx = 0, .toks_len = 0 };

    pub fn from(toks_idx: usize, toks_len: usize) AstNode {
        return .{ .toks_idx = toks_idx, .toks_len = toks_len };
    }

    pub fn fromExprs(exprs: []AstExpr) AstNode {
        const toks_idx = exprs[0].base.toks_idx;
        var toks_len: usize = 0;
        for (exprs) |*expr|
            toks_len += expr.base.toks_len;
        return from(toks_idx, toks_len);
    }

    pub fn toks(me: *AstNode, all_toks: []Token) []Token {
        return all_toks[me.toks_idx .. me.toks_idx + me.toks_len];
    }

    pub fn str(me: *AstNode, all_toks: []Token, full_src: Str) Str {
        return Token.str(me.toks(all_toks), full_src);
    }
};

pub const AstDef = struct {
    base: AstNode,
    head: AstExpr,
    body: AstExpr,
    defs: []AstDef,
    scope: Ast.Scopes = Ast.Scopes{},
    is_top_def: bool,

    pub fn name(me: *AstDef) Str {
        return switch (me.head.kind) {
            .ident => |ident| ident,
            else => fail("{}", .{me.head}),
        };
    }

    pub fn ty(me: *AstDef, ast: *Ast) ?AstExpr.Builtin.Ty {
        return me.body.maybeType(me, ast);
    }

    fn ensureNameRefs(me: *AstDef, ast: *Ast, scopes: *Ast.Scopes) void {
        me.scope.parent = scopes;
        me.scope.cur = alloc(Ast.NameRef, me.defs.len);
        for (me.defs) |*def, i| {
            const def_name = def.name();
            if (me.scope.resolve(def_name, i)) |_|
                fail("duplicate name '{}' near:\n{}", .{ def_name, def.base.str(ast.toks, ast.src) })
            else
                me.scope.cur[i] = .{ .name = def_name, .refers_to = .{ .def = def } };
        }
        for (me.defs) |*def|
            def.ensureNameRefs(ast, &me.scope);
    }

    fn rewriteStrLits(me: *AstDef, into: [][2]Str, idx_start: usize) usize {
        if (me.is_top_def and me.body.kind == .lit_str)
            return idx_start;
        var idx = me.body.rewriteStrLits(into, idx_start);
        for (me.defs) |*def|
            idx = def.rewriteStrLits(into, idx);
        return idx;
    }
};

pub const AstExpr = struct {
    base: AstNode,
    kind: Kind,
    ty: ?Builtin.Ty = null,

    pub fn dynNamePart(me: *AstExpr, dyn_name_prefixes: []Str) Str {
        return switch (me.kind) {
            else => unreachable,
            .lit_int => |lit_int| uintToStr(lit_int, 10, 1, ""),
            .ident => |ident| ident,
            .builtin => |builtin| switch (builtin.*) {
                else => fail("dynNamePart for 'builtin' expr of kind '{}'", .{@tagName(builtin.*)}),
                .atom => |atom| @tagName(atom),
                .ext_var => |ext_var| ext_var.name,
            },
        };
    }

    fn mustType(me: *AstExpr, def: *AstDef, ast: *Ast) Builtin.Ty {
        return me.maybeType(def, ast) orelse fail("failed to determine type of '{}' expression '{}'", .{ @tagName(me.kind), me.base.str(ast.toks, ast.src) });
    }

    fn maybeType(me: *AstExpr, def: *AstDef, ast: *Ast) ?Builtin.Ty {
        if (me.ty == null) {
            switch (me.kind) {
                else => unreachable,
                .lit_str => |lit_str| me.ty = Builtin.Ty{ .arr = .{ .size = lit_str.len, .of = keep(Builtin.Ty{ .int = 8 }) } },
                .lit_int => me.ty = Builtin.Ty{ .int = null },
                .ident => |ident| if (def.scope.resolve(ident, null)) |refers_to| {
                    switch (refers_to) {
                        .def => |ref_def| me.ty = ref_def.ty(ast),
                    }
                },
                .builtin => |builtin| switch (builtin.*) {
                    else => fail("TODO type of builtin exprs of kind '{}'", .{@tagName(builtin.*)}),
                    .len => me.ty = Builtin.Ty{ .int = 64 },
                    .atom => |atom| switch (atom) {
                        else => unreachable,
                        .@"true", .@"false" => {
                            me.ty = Builtin.Ty{ .int = 1 };
                        },
                    },
                    .alloca => |alloca| me.ty = alloca.ty,
                    .bin_op => |bin_op| switch (bin_op.kind) {
                        .sge, .sgt, .eq, .ne => me.ty = Builtin.Ty{ .int = 1 },
                    },
                    .case => |case| {
                        for (case.prongs) |*pair, i| {
                            me.ty = pair.rhs.maybeType(def, ast);
                            if (me.ty != null and me.ty.? == .int and me.ty.?.int == null)
                                me.ty = null;
                        }
                    },
                    .ext_call => |*call| {
                        switch (call.callee.kind) {
                            else => fail("{}", .{@tagName(call.callee.kind)}),
                            .ident => |ident| switch (def.scope.mustResolve(ident, null)) {
                                .def => |ref_def| me.ty = ref_def.ty(ast),
                            },
                        }
                    },
                    .ext_fn => |*ext_fn| me.ty = ext_fn.ty,
                },
            }
        }
        return me.ty;
    }

    fn from(exprs: []AstExpr, all_toks: []Token, full_src: Str) AstExpr {
        var ret: AstExpr = if (exprs.len == 1) exprs[0] else AstExpr{
            .base = AstNode.fromExprs(exprs),
            .kind = .{ .form = exprs },
        };
        if (ret.toBuiltin(all_toks, full_src)) |builtin|
            ret.kind = .{ .builtin = builtin };
        return ret;
    }

    pub fn toBuiltin(me: *AstExpr, all_toks: []Token, full_src: Str) ?*Builtin {
        var ret: ?Builtin = null;
        var name: Str = "";
        switch (me.kind) {
            else => {},
            .builtin => |builtin| return builtin,
            .ident => |ident| if (ident[0] == '/') {
                name = ident;
                if (name.len == 2 and name[1] == '*')
                    ret = Builtin{ .ty = .{ .ptr = null } }
                else if (name.len > 2 and name[1] == '*' and name[2] == '/') {
                    var tmp = AstExpr{ .base = me.base, .kind = .{ .ident = name[2..] } };
                    if (tmp.toBuiltin(all_toks, full_src)) |builtin|
                        ret = Builtin{ .ty = .{ .ptr = keep(builtin.ty) } };
                } else if (name.len > 2 and name[1] == 'I' and name[2] > '0' and name[2] <= '9')
                    ret = Builtin{ .ty = .{ .int = @intCast(u23, parseExprLitInt(name[2..])) } }
                else if (eql(name, "/true"))
                    ret = Builtin{ .atom = .@"true" }
                else if (eql(name, "/false"))
                    ret = Builtin{ .atom = .@"false" }
                else if (eql(name, "/null"))
                    ret = Builtin{ .atom = .@"null" }
                else if (eql(name, "/else"))
                    ret = Builtin{ .atom = .@"else" };
            },
            .form => |exprs| if (exprs[0].kind == .ident and exprs[0].kind.ident[0] == '/') {
                name = exprs[0].kind.ident;
                if (eql(name, "/call") and exprs.len >= 3) {
                    ret = Builtin{ .ext_call = .{ .opts = exprs[1], .callee = exprs[2], .args = exprs[3..] } };
                } else if (eql(name, "/alloca") and exprs.len == 3) {
                    if (exprs[1].toBuiltin(all_toks, full_src)) |builtin| {
                        if (builtin.* == .ty)
                            ret = Builtin{ .alloca = .{ .num_elems = exprs[2], .ty = builtin.ty } };
                    }
                } else if (eql(name, "/len") and exprs.len == 2) {
                    ret = Builtin{ .len = exprs[1] };
                } else if (eql(name, "/case") and exprs.len == 3) {
                    const prongs = exprs[2].kind.lit_obj.elems;
                    var case = Builtin.Case{ .scrut = exprs[1], .prongs = exprs[2].kind.lit_obj.elems, .default_idx = null };
                    for (prongs) |*prong, i| {
                        const is_default_case = (prong.lhs.kind == .builtin and prong.lhs.kind.builtin.* == .atom and prong.lhs.kind.builtin.atom == .@"else");
                        if (is_default_case) {
                            if (case.default_idx == null)
                                case.default_idx = i
                            else {
                                case.default_idx = null;
                                break;
                            }
                        }
                    }
                    if (case.default_idx != null)
                        ret = Builtin{ .case = case };
                } else if (eql(name, "/extVar") and exprs.len == 3) {
                    if (exprs[2].toBuiltin(all_toks, full_src)) |builtin| {
                        if (builtin.* == .ty and exprs[1].kind == .lit_str)
                            ret = Builtin{ .ext_var = .{ .name = exprs[1].kind.lit_str, .ty = builtin.ty } };
                    }
                } else if (eql(name, "/extFn") and exprs.len == 4) {
                    if (exprs[2].toBuiltin(all_toks, full_src)) |builtin| {
                        if (builtin.* == .ty and exprs[1].kind == .lit_str and exprs[3].kind == .lit_obj) {
                            ret = Builtin{ .ext_fn = .{ .name = exprs[1].kind.lit_str, .ty = builtin.ty, .args = alloc(Builtin.Ty, exprs[3].kind.lit_obj.elems.len) } };
                            for (exprs[3].kind.lit_obj.elems) |*elem, i| {
                                if (ret == null)
                                    break
                                else if (elem.rhs.toBuiltin(all_toks, full_src)) |bi| {
                                    if (bi.* == .ty)
                                        ret.?.ext_fn.args[i] = bi.ty
                                    else
                                        ret = null;
                                } else
                                    ret = null;
                            }
                        }
                    }
                } else if (eql(name, "/fn") and exprs.len >= 4) {
                    if (exprs[1].toBuiltin(all_toks, full_src)) |builtin| {
                        if (builtin.* == .ty and exprs[2].kind == .lit_obj) {
                            ret = Builtin{
                                .fn_def = .{
                                    .ty = builtin.ty,
                                    .expr = AstExpr.from(exprs[3..], all_toks, full_src),
                                    .args = alloc(Builtin.Fn.Arg, exprs[2].kind.lit_obj.elems.len),
                                },
                            };
                            for (exprs[2].kind.lit_obj.elems) |*elem, i| {
                                if (ret == null)
                                    break
                                else if (elem.rhs.toBuiltin(all_toks, full_src)) |bi| {
                                    if (bi.* == .ty)
                                        ret.?.fn_def.args[i] = .{ .name = elem.lhs, .ty = bi.ty }
                                    else
                                        ret = null;
                                } else
                                    ret = null;
                            }
                        }
                    }
                } else if (exprs.len == 3 and name.len > 3 and eql(name[0..3], "/op")) {
                    ret = Builtin{ .bin_op = Builtin.BinOp{ .kind = undefined, .op1 = exprs[1], .op2 = exprs[2] } };
                    const op_name = name[3..];
                    if (eql(op_name, "Sgt"))
                        ret.?.bin_op.kind = .sgt
                    else if (eql(op_name, "Sge"))
                        ret.?.bin_op.kind = .sge
                    else if (eql(op_name, "Eq"))
                        ret.?.bin_op.kind = .eq
                    else if (eql(op_name, "Ne"))
                        ret.?.bin_op.kind = .ne
                    else
                        ret = null;
                }
            },
        }
        return if (ret) |r| keep(r) else (if (name.len == 0)
            null
        else
            fail("malformed builtin '{}' near:\n{}", .{ name, me.base.str(all_toks, full_src) }));
    }

    fn rewriteStrLits(me: *AstExpr, into: [][2]Str, idx_start: usize) usize {
        var idx = idx_start;
        switch (me.kind) {
            else => {},
            .lit_str => |lit_str| {
                counter += 1;
                const new_name = uintToStr(counter, 10, 1, "str.s");
                me.kind = .{ .ident = new_name };
                into[idx] = [2]Str{ lit_str, new_name };
                idx += 1;
            },
            .lit_obj => |lit_obj| for (lit_obj.elems) |*elem| {
                idx = elem.rhs.rewriteStrLits(into, idx);
            },
            .lit_arr => |lit_arr| for (lit_arr.elems) |*expr| {
                idx = expr.rewriteStrLits(into, idx);
            },
            .form => |exprs| for (exprs) |*expr| {
                idx = expr.rewriteStrLits(into, idx);
            },
            .builtin => |builtin| switch (builtin.*) {
                else => {},
                .ext_call => |*call| {
                    idx = call.callee.rewriteStrLits(into, idx);
                    for (call.args) |*expr|
                        idx = expr.rewriteStrLits(into, idx);
                },
                .case => |*case| {
                    idx = case.scrut.rewriteStrLits(into, idx);
                    for (case.prongs) |*elem| {
                        idx = elem.lhs.rewriteStrLits(into, idx);
                        idx = elem.rhs.rewriteStrLits(into, idx);
                    }
                },
                .fn_def => |*fn_def| {
                    idx = fn_def.expr.rewriteStrLits(into, idx);
                },
            },
        }
        return idx;
    }

    pub const Kind = union(enum) {
        lit_int: u128,
        lit_str: Str,
        lit_obj: LitObj,
        lit_arr: LitArr,
        ident: Str,
        form: []AstExpr,
        builtin: *Builtin,
    };

    pub const LitArr = struct {
        elems: []AstExpr,
    };

    pub const LitObj = struct {
        elems: []Pair,

        pub const Pair = struct {
            lhs: AstExpr,
            rhs: AstExpr,
        };
    };

    pub const Builtin = union(enum) {
        ext_var: ExtVar,
        ext_fn: ExtFn,
        fn_def: Fn,
        case: Case,
        alloca: Alloca,
        ext_call: ExtCall,
        atom: enum {
            @"true",
            @"false",
            @"else",
            @"null",
        },
        bin_op: BinOp,
        ty: Ty,
        len: AstExpr,

        pub const BinOp = struct {
            op1: AstExpr,
            op2: AstExpr,
            kind: enum {
                eq,
                ne,
                sgt,
                sge,
            },
        };
        pub const ExtVar = struct {
            name: Str,
            ty: Ty,
        };
        pub const Fn = struct {
            args: []Arg,
            ty: Ty,
            expr: AstExpr,
            pub const Arg = struct {
                name: AstExpr,
                ty: Ty,
            };
        };
        pub const ExtFn = struct {
            name: Str,
            ty: Ty,
            args: []Ty,
        };
        pub const Case = struct {
            scrut: AstExpr,
            prongs: []LitObj.Pair,
            default_idx: ?usize,
        };
        pub const Alloca = struct {
            ty: Ty,
            num_elems: AstExpr,
        };
        pub const ExtCall = struct {
            opts: AstExpr,
            callee: AstExpr,
            args: []AstExpr,
        };
        pub const Ty = union(enum) {
            ptr: ?*Ty,
            int: ?u23,
            arr: Arr,

            pub const Arr = struct {
                of: *Ty,
                size: usize,
            };
        };
    };
};
