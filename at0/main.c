#pragma clang diagnostic ignored "-Wunused-parameter"
#pragma clang diagnostic ignored "-Wunused-function"

#include "utils_and_libc_deps.c"
#include "fs_io.c"

#include "at_toks.c"
#include "at_ast.c"
#include "at_parse.c"
#include "ir_hl.c"
#include "ir_ll.c"



void readLnLoop(IrHLProg const* const);

int main(int const argc, CStr const argv[]) {
    ·assert(argc == 2);

    CtxParseAsts ctx_parse = (CtxParseAsts) {.asts = ·make(Ast, 0, asts_capacity), .src_file_paths = ·make(Str, 0, asts_capacity)};
    ·append(ctx_parse.src_file_paths, str(argv[1]));
    loadAndParseRootSourceFileAndImports(&ctx_parse);

    IrHLProg ir_hl = irHLProgFrom(ctx_parse.asts);
    // irHLProcessIdents(&ir_hl); // resolve references: throw on shadowings or unresolvables
    // irHLProgLiftFuncExprs(&ir_hl);
    irHLPrintProg(&ir_hl);

    // readLnLoop(&ir_hl);
    return 0;
}

void readLnOnInput(IrHLProg const* const prog, Str const input) {
    writeStr(str("————————————————————————————————————————————————————————————\n"));
    if (strEql(str("?"), input))
        ·forEach(IrHLDef, def, prog->defs, {
            writeStr(def->name);
            writeStr(str("\n"));
        });
    else {
        IrHLDef const* found = NULL;
        ·forEach(IrHLDef, def, prog->defs, {
            if (strEql(def->name, input))
                found = def;
        });
        if (found == NULL)
            writeStr(str("‹unknown def name›\n"));
        else
            irHLPrintDef(found);
    }
    writeStr(str("————————————————————————————————————————————————————————————\n"));
}

void readLnLoop(IrHLProg const* const prog) {
#define buf_size 8
    Str buf = newStr(0, buf_size);
    while (true) {
        int chr = fgetc(stdin);
        if (chr < 1 || chr > 255)
            exit((feof(stdin) != 0) ? 0 : 1);
        if (chr == '\n') {
            if (buf.len != 0)
                readLnOnInput(prog, buf);
            buf.len = 0;
        } else {
            buf.at[buf.len] = (U8)chr;
            buf.len += 1;
            if (buf.len == buf_size)
                ·fail(str("TODO: moar buf_size"));
        }
    }
}

//  gdb at0 -ex run  asm volatile("int3");
