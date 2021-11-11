#define mem_max (1 * 1024 * 1024)
#pragma clang diagnostic ignored "-Wunused-parameter"
#pragma clang diagnostic ignored "-Wunused-function"

#include <stdio.h>

#include "utils_std_basics.c"
#include "utils_std_mem.c"
#include "utils_json.c"
#include "utils_toks.c"
#include "fs_io.c"
#include "lsp.c"
#include "at_parse.c"
#include "at_ast.c"
#include "ir_hl.c"
#include "ir_ml.c"


int main(int const argc, CStr const argv[]) {
    if (argc == 1)
        lspMainLoop();
    else {
        CtxSrcFileSess ctx = loadFromSrcFile(str(argv[1]));

        // if (ctx.prog_hl != NULL)
        //     irhlPrintProg(ctx.prog_hl);

        if (ctx.parsed->issues)
            ·forEach(Ast, ast, ctx.parsed->asts, {
                ·forEach(SrcFileIssue, issue, ast->issues, {
                    fprintf(stderr, "%s:%zu:%zu: %s\n", strZ(ast->anns.src_file_path), 1 + issue->src_pos.line_nr,
                            1 + (issue->src_pos.char_pos - issue->src_pos.char_pos_line_start), strZ(issue->msg));
                });
            });
    }
    return 0;
}
