#pragma clang diagnostic ignored "-Wunused-parameter"
#pragma clang diagnostic ignored "-Wunused-function"

#include "utils_and_libc_deps.c"
#include "std_io.c"

#include "at_toks.c"
#include "at_ast.c"
#include "at_parse.c"
#include "ir_hl.c"
#include "ir_ll.c"



void readLnLoop(IrHLProg const* const);

int main(int const argc, CStr const argv[]) {
    ·assert(argc > 1);

    // read and concat together all input source files specified via args
    Str full_src = ·len0(U8);
    for (int i = 1; i < argc; i += 1) {
        // hacky: all allocs in this loop (strCopy and readFile) are contiguous in `mem`,
        // so our `full_src` bytes-slice just gets the starting addr and its `len` increased.
        // this means to never introduce any calls in here that also alloc from `mem`!
        Str const comment_part_1 = strCopy("//AT_TOKS_SRC_FILE:");
        Str const comment_part_2 = strCopy(argv[i]);
        Str const comment_part_3 = strCopy("\n");
        Str const this_file_src = readFile(argv[i]);
        full_src.len += comment_part_1.len + comment_part_2.len + comment_part_3.len + this_file_src.len;
        if (full_src.at == NULL)
            full_src.at = comment_part_1.at;
    }

    // tokenize
    Tokens const toks = tokenize(full_src, false, str(""));
    toksVerifyBrackets(toks);

    // parse into a rudimentary raw context-free generic AST first
    Ast ast = parse(toks, full_src);
    astRewriteGlyphsIntoInstrs(&ast);
    // astPrint(&ast);

    // interpret raw-and-dumb *syntax* tree into actual language *semantics*:
    IrHLProg ir_hl = irHLProgFrom(&ast);
    irHLProcessIdents(&ir_hl); // resolve references: throw on shadowings or unresolvables
    irHLProgLiftFuncExprs(&ir_hl);
    // irHLPrintProg(&ir_hl);

    readLnLoop(&ir_hl);
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
