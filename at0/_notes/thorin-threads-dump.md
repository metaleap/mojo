10:04 AM
Hello, here is Arik from the Group in Mainz.
We recently encountered a problem with the Spawn-function. Using the function spawn with some runtime-generated value crashes the programm. I included a small programm to reproduce this, where I fetch a random int from c's rand-function, and then try to make spawn print that value. If I do this with the parallel-loop everything works just fine.
(@roland: the line that crashes the programm was commented out. It is now active again)
10:05 AM
example.tar.gz
Spawn-Error example
Owner
1:43 PM
arik: fixed: https://github.com/AnyDSL/thorin/commit/d77a6168316d9ffe9bd0eec1d75af365a8aa5d83

    github.com
    Fix for tuple problem in spawn(). · AnyDSL/thorin@d77a616
    The Higher-Order Intermediate Representation. Contribute to AnyDSL/thorin development by creating an account on GitHub.

Owner
5:03 PM
Does someone know what's the current state with alignment and vectorization?
So how do I tell RV that sth is properly aligned?
Owner
6:02 PM
As far as I know that was broken and not used by RV.
9:59 AM
Has joined the channel.
10:01 AM
Hi, ich hatte mir in den Kalender eingetragen, dass das nächste Treffen am 14. Januar in Mainz stattfindet. Hab ich da was verpeilt oder ist das so richtig? Bin nur verwundert weil diesbezüglich keine E-Mail mehr kam.
Owner
3:08 PM
Hi Jonas, ja das stimmt, unser nächste Treffen ist am 14. Januar in Mainz.
Start wieder gegen 10:30 Uhr.
3:09 PM
Alles klar^^
Owner
3:14 PM
sebastian.hack schreibst du noch eine E-Mail?
Owner
3:31 PM
wegen Montag?
Owner
3:32 PM
jap
Owner
3:36 PM
ja, mach' ich
Owner
9:57 AM
schaut so aus als ob llvm kurz davor sind ihre svn->git migration abzuschliessen
https://llvm.org/docs/Proposals/GitHubMove.html

    llvm.org
    Moving LLVM Projects to GitHub — LLVM 8 documentation

haben auch die zwei optionen diskutiert: mono-repo vs. multi-repo mit umbrella-submodule-project (read-only)
der aktuelle prototyp verwendet das mono-repo: https://github.com/llvm-git-prototype/llvm

    github.com
    llvm-git-prototype/llvm
    Contribute to llvm-git-prototype/llvm development by creating an account on GitHub.

10:37 AM
Message removed
Owner
8:28 PM
Can we switch to English again? This is supposed to be a public AnyDSL support channel 🙂
Owner
12:00 PM
Sorry to answer that late, but for the alignment of a variable, you can use the rv_align intrinsic
See https://github.com/madmann91/rodent/blob/master/src/traversal/mapping_cpu.impala#L90

    github.com
    madmann91/rodent
    Contribute to madmann91/rodent development by creating an account on GitHub.

Owner
7:23 PM
thanks 🙂
1:40 PM
Mit ist gerade aufgefallen, dass ich leider immer noch nicht auf das Antragsrepo pushen kann.
Fehlermeldung:

remote: GitLab: You are not allowed to push code to protected branches on this project.
To https://public.cdl.uni-saarland.de/metacca/antrag.git
! [remote rejected] master -> master (pre-receive hook declined)
error: failed to push some refs to 'https://public.cdl.uni-saarland.de/metacca/antrag.git'

SSH und HTTPS funktioniert beides nicht.

    public.cdl.uni-saarland.de
    Sign in
    Compiler Design Lab Gitlab

    public.cdl.uni-saarland.de
    Sign in
    Compiler Design Lab Gitlab

Owner
12:35 PM
ich ab dich mal gerade zum master gemacht. hilft das?
4:21 PM
Has joined the channel.
11:21 AM
Why does an array require a literal value instead of a constant expression?
let buffer = [0, .. 10]; vs. let size = 10; let buffer = [0, .. size]; /* do something with buffer */
Owner
11:50 AM
For technical reasons, this is not implemented. This will be fixed in the next version of Thorin that we are working on.
1:40 PM
At least we can simulate it on the library level I guess. https://gist.github.com/DasNaCl/eafaf4db6c4b4b5f6d023e6dc9148d41
But I still don't like that I have to rewrite this whole thing under a different name if I want to use another
DataVector
with a different value type. 😥

    gist.github.com
    Simulating constant sized arrays with a library.
    Simulating constant sized arrays with a library. GitHub Gist: instantly share code, notes, and snippets.

4:45 PM
Has joined the channel.
4:54 PM
Hello guys
I'm having a problem when running my AnyDSL code in our university cluster. I keep getting the following illegal instruction messages when submitting the jobs:

[i10hpc11:23312:0:23312] Caught signal 4 (Illegal instruction: illegal operand)
==== backtrace ====
0 /usr/lib/libucs.so.0(+0x1d280) [0x7f1c64304280]
1 /usr/lib/libucs.so.0(+0x1d303) [0x7f1c64304303]
2 ./md(md_initialize_grid+0x11) [0x56543514e2b1]
3 ./md(_Z21init_rectangular_gridj4AABBPdddi+0x8ce) [0x565435141a1e]
4 ./md(main+0x389) [0x56543513c4c9]
5 /lib/x86_64-linux-gnu/libc.so.6(__libc_start_main+0xe7) [0x7f1c777c9b97]
6 ./md(_start+0x2a) [0x56543513e27a]

The code is available here: https://github.com/AnyDSL/molecular-dynamics/tree/mpi/neighborlists
Cluster information can be found here: https://www.cs10.tf.fau.de/research/hpc-cluster/

    github.com
    AnyDSL/molecular-dynamics
    Contribute to AnyDSL/molecular-dynamics development by creating an account on GitHub.

    www.cs10.tf.fau.de
    HPC cluster › www.cs10.tf.fau.de
    Since 2013, the LSS maintains a 36-CPU High-Performance-Cluster. In the following, general information and hints about the usage of the ...

We suspected that could be something related to the vectorization code, but I'm compiling for CPU backend and by looking at the message it looks like the problem is happening at the md_initialize_grid call
Do you have any idea what could be causing this? If you need more info just ask me here
Owner
8:00 PM
can you try to replace all
vectorize loops with appropriate range
loops?
8:38 PM
I already tried but the same problem happens
11:28 PM
I noticed that the address md_initialize_grid+0x11 contains the following instruction (objdump):

md_initialize_grid+0x11
9ed1: c5 fb 11 84 24 b8 00 vmovsd %xmm0,0xb8(%rsp)
Does this instruction could be generated even without the use of the vectorize function?
Owner
11:36 PM
I'm really puzzled why this happens. I have the impression that somehwere in the pipeline LLVM is seeing the wrong flags and generates code for the wrong CPU
what cmake magic do you use to link the
.ll
file AnyDSL generates with the rest of your project?
11:53 PM
I use anydsl_runtime_wrap to compile the Anydsl code and then add_executable with both the obtained Anydsl compiled code and the cpp files (don't know if this answers your question)
I think the problem is related to the fact that the code is compiled in the visualization node (so LLVM generates these vectorization instructions since it works there) and then runs in the compute nodes
Owner
11:55 PM
yes. that explains a thing or two... 😃
long story short, we use the host architecture to configure the compilation pipeline in LLVM
I have to talk with richard.membarth tomorrow. this is a known problem and we've already talked about workarounds
12:01 AM
Alright, is there a way I can make some workaround at the moment or it shouldn't be too easy?
Owner
12:03 AM
rafaelravedutti do you specify march=native to CLANG_FLAGS?
https://github.com/AnyDSL/molecular-dynamics/blob/mpi/neighborlists/CMakeLists.txt#L28

    github.com
    AnyDSL/molecular-dynamics
    Contribute to AnyDSL/molecular-dynamics development by creating an account on GitHub.

here you set CLANG_FLAGS to -march=native
this optimizes the code for the architecture of the node where you're compiling, not for the target node
12:06 AM
Oh, alright
Owner
12:06 AM
but I still think we have to change the target triple in llvm, no?
Owner
12:07 AM
I think we're fine with specifying CLANG_FLAGS
12:07 AM
I'll change it and make some tests
Thank you for the support guys
Owner
12:08 AM
unless there happen low-level optimizations in the vectorizer 🙂
Owner
12:08 AM
yes, the target triple should be compatible enough in this setting
Owner
12:08 AM
typically those error come from the march=native or missing of it 🙂
if this is not enough, set the target_tripe & target_cpu when running impala
https://anydsl.github.io/Device-Code-Generation-and-Execution.html

    anydsl.github.io
    Device Code Generation and Execution - AnyDSL
    Describes how device code generation and execution works in Impala.

https://anydsl.github.io/Device-Code-Generation-and-Execution.html#cross-compilation

    anydsl.github.io
    Device Code Generation and Execution - AnyDSL
    Describes how device code generation and execution works in Impala.

Owner
12:10 AM
ah. this is already implemented. nice. this is the stuff I was talking about
Owner
12:11 AM
and this is even propagated to RV 😛
Owner
12:11 AM
kewl 😎
Owner
12:12 AM
rafaelravedutti note that we only generate the .ll for the target architecture, for generating the proper executable, you need a compiler for the target machine
this shouldn't be a problem in you case, but for generating ARM binaries on a x86 machine, you need a ARM cross compiler
12:14 AM
Alright, got it, thanks!
Owner
12:15 AM
you're welcome 🙂
12:15 AM
😁
11:10 AM
Has joined the channel.
12:41 PM
I'm having trouble getting rodent to work on my home computer, I've tried Ubuntu 19.04 and couldn't get that to work, since it ships with llvm8 I decided to try Debian buster ( since I have got it working on my thinkpad using Debian too ), but no dice
the anydsl meta-repo builds fine ( I enabled JIT support, disabled aarch64, arm and and nvvm since I don't need those, to get faster builds ), and so does rodent (by pointing it to the runtime lib via -DCMAKE_PREFIX_PATH )
but I get a fail when I launch the app

gobrosse@xolioware-ryzen:~/git/anydsl/rodent/build$ bin/rodent
rodent: /home/gobrosse/git/anydsl/llvm/lib/Support/CommandLine.cpp:282: void {anonymous}::CommandLineParser::registerCategory(llvm::cl::OptionCategory*): Assertion `count_if(RegisteredOptionCategories, [cat](const OptionCategory *Category) { return cat->getName() == Category->getName(); }) == 0 && "Duplicate option categories"' failed.

I think I even had a similar issue on my laptop but I don't remember what I did 😕
here's my
config.sh
https://privatebin.net/?4945604b29aa3ca5#r45aRjbOC5MvljZw8krhbGNgY8qrEuYS3gqmgr3niUI=

    privatebin.net
    PrivateBin

5:02 PM
looks like the most opencl drivers except intel's have their own version of LLVM and clash with anydsl's
Owner
5:50 PM
The error comes from different shared libraries linking to LLVM.
We are aware of this and working towards a solution (not exposing LLVM).
You might even get the same issue when different OpenCL implementations link to (different?) LLVM (versions?).
9:07 PM
What's the impala way of doing pointer arithmetic ? ie I want to return a "subarray" starting at a given offset within an existing array, I have came up with this

let subarray = bitcast[&mut[(i32,i32)]]((bitcast[u64](temp_array) + ((sizeof[(i32, i32)]() * bin_offset) as u64)));


assuming 64-bit addresses, but that's rather messy and might not even work (I'm working on efficiently sorting my splits list)
with
temp_array: &mut[(i32, i32)]
Owner
9:25 PM
sure, you can do some wild bitcasts and hope for the best but impala doesn't do this C-style madness of more or less equating pointers and arrays
the question is more: what do you really want to do?
maybe pass the array by reference and the starting index
9:27 PM
Sorting a section of a an array - yeah that's what I'm probably wind up doing, I'm also asking for the sake of knowing if it's in good taste or not ( so, no )
10:22 PM
Any specific gotchas on using
parallel()
? I doesn't look like it's spawning threads, my app just sits eating 100% of the CPU, the parallelized loop shares a mutable array maybe the compiler is inserting some sort of automatic barriers ?
reference: https://github.com/AnyDSL/bvh/commit/ad4198243ec74dd1aeff8138e39bc259691bbfa6
11:18 PM
I figured it out, my bins weren't at all equally loaded, and so only one thread had actual work to do
Owner
12:09 PM
great 🙂
1:02 AM
Hello again, I have some more questions 🙂 does
anydsl_link work as advertised in the docs (as in, imports structs & function definitions) in the current version ? I'm trying to modify rodent so it has a dynamically compiled main ( for linking w/ the BVH builder i'm working on, I have an issue open for discussion about that over on GH ), but it doesn't seem to do anything:

std::string jit_code = "struct Vec3 {\n"
                           "    x: f32,\n"
                           "    y: f32,\n"
                           "    z: f32\n"
                           "}"
                           ""
                           "struct Settings {\n"
                           "    eye: Vec3,\n"
                           "    dir: Vec3,\n"
                           "    up: Vec3,\n"
                           "    right: Vec3,\n"
                           "    width: f32,\n"
                           "    height: f32\n"
                           "}\n"
                           ""
                           "extern fn render_via_jit(settings: &Settings, iter: i32) -> () {"
                           "    render(settings, iter);"
                           "}";

    anydsl_link("lib/librodent_lib.so");
    //anydsl_link("/home/hugo/git/anydsl2/rodent/build/lib/librodent_lib.so");
    //anydsl_link("librodent_lib.so");
    auto cpld = anydsl_compile(jit_code.c_str(), jit_code.size(), 3);
    typedef void(* render_fn)(struct Settings const* settings, int iter);
    render_fn obtained = reinterpret_cast<render_fn >(anydsl_lookup_function(cpld, "render_via_jit"));



crashes, with a complaint about 'render' not being found in the current scope, even though it's marked "extern" and I can see it when doing readelf -Ws librodent_lib.so
I obtained that .so with the following cmake line:

add_library(rodent_lib SHARED rodent.o))



for further ref, I have uploaded my hacked-up copy of rodent there: https://github.com/Hugobros3/rodent/commit/baae22116f005ffe236ff4c2d2954e4f57aa7bc2#diff-264e4c834e0c0adf5707f58779a59b4bR349

maybe there is some magic Impala flag to enable on compilation (JIT support is of course enabled in config.sh) ?
maybe I need to "forward-declare" the definitions I want to link with ?
Is there is any AnyDSL application/test that uses that anydsl_link
function I have trouble with ?

    github.com
    trying to get rodent to jit itself up · Hugobros3/rodent@baae221
    Custom version of AnyDSL/rodent for use in the Real-time rendering seminar - Hugobros3/rodent

Owner
8:58 AM
are the structs and render declared as
extern "C"
in your library?
impala allows you to export the interface
impala -emit-c-interface:

/* bla.h : Impala interface file generated by impala */
#ifndef BLA_H
#define BLA_H

#ifdef __cplusplus
extern "C" {
#endif

struct Vec3 {
    float x;
    float y;
    float z;
};

struct Settings {
    struct Vec3 eye;
    struct Vec3 dir;
    struct Vec3 up;
    struct Vec3 right;
    float width;
    float height;
};

void render_via_jit(struct Settings const* settings, int iter);

#ifdef __cplusplus
}
#endif

#endif /* BLA_H */

That's the other way round, but shows you what you need to expose from C++.
10:40 AM
Unfortunately going through the C interface generator stops me from having functions in the structs, which I need in order to implement the design Arsène advised me:
https://github.com/AnyDSL/bvh/issues/1
The interface generator just gives up as soon as it encounters a field with a lambda type
I'd want my BVH library entry point to take something like:

struct Input {
    bboxes: fn (i32) -> BBox,
    centers: fn (i32) -> Vec3,
    num_prims: i32
}

So be clear i'm trying to get Impala code to link with other Impala code that isn't directly in the same source set, without going through the C interface (or alternatively having a way to pass lambdas around via that C interface)
Owner
2:56 PM
I see. That doesn't work. Impala needs to specialize all functions in a struct. That is, we cannot generate code if this is not the case.
You need to compile all impala code in one go.
3:16 PM
That's what Arsène told me too, I got confused when reading the doc and thought the JIT could somehow merge additional impala code, but in fact you can only link to C-style interfaces
anydsl_link does indeed work for linking to Impala functions that export a C signature
5:41 PM
Has joined the channel.
5:48 PM
Hi! Lately I've been playing around with anydsl and I have a question. Is there any way to check out the specialized/partially evaluated version of the program, e.g. impala -show-specialized?
Owner
6:59 PM
No, not really.
You can only look at the non-specialized IR with
impala -emit-thorin -nocleanup
impala -emit-thorin
will show the IR after specialization
You can watch the progress of PE by using
pe_info(string, var)
, these calls will be printed during PE.
https://github.com/AnyDSL/stincilla/blob/master/matmul.impala#L70

    github.com
    AnyDSL/stincilla
    A DSL for Stencil Codes. Contribute to AnyDSL/stincilla development by creating an account on GitHub.

Owner
7:17 PM
This will print during compilation:

I:anydsl/stincilla/matmul.impala:70 col 13 - 38: pe_info: step size: qs32 256
I:anydsl/stincilla/matmul.impala:70 col 13 - 38: pe_info: step size: qs32 64
I:anydsl/stincilla/matmul.impala:70 col 13 - 38: pe_info: step size: qs32 1

5:17 PM
I've tried to specialize a bit and got a question
whether it is possible to specialize on statically defined array passed as input. E.g. suppose I wanna find all substrings from some set of templates in one string and I want to specialize string-matcher on those templates.(or maybe there is another convenient approach to implement such things ) Since this kind of stuff seems not to work: https://pastebin.com/gzxvYNYL

This pe outputs I:/home/gerwant/cmake-build/anydsl/stincilla/my_test/fun.impala:27 col 5 - 28: pe_info not constant: array index: _6424

Also I have impala -emit-thorin -nocleanup and impala -emit-thorin flags output the same result both in the above case and in the case of pow
function (with the pow they both show specialized version).

    pastebin.com
    [Rust] extern "C" { fn println(a : int) -> (); } static a = [1,3,5,7,9,11,13,1 - Pastebin.com

Owner
8:05 PM
you're manipulating mutual state there, something that is not properly supported via PE on master
we're working on a new version where this will be allowed
you also take the address of a, which might prevent PE from treating it as static
for simple string manipulation, you might want to have a look at https://github.com/AnyDSL/regex

    github.com
    AnyDSL/regex
    Regular Expression Matcher generator using Just-In-Time Compilation - AnyDSL/regex

4:51 PM
Message removed
4:57 PM
Message removed
5:07 PM
So far I have done the thing with JIT to allow specialization of static array data and it indeed did specialization, however I got confused a bit.

It appears that PE is able to decide whcih part of the input is static and thus specialize the call cite with respect to some of the arguments even w/o any annotations, e.g. in my case all the variants of annotations (i.e. their presence or absence) produced the same result, except for explicit
fn @(false)range_ function declaration. And if it is the case, then what is the purpose of annotations? Just prevent the specialization? So, am I right, or I marked some annotations in a wrong way?

The second thing is that PE seems to not specialize cuda code (compiler produced same results for both fn range_(...) and fn @(false)range_(...), while with cpu code the results are completely different. I passed array wrapped in a structure to keep it static and let it to being passed to constant cuda memory without explicit allocations)

I have summarized all the cases of specializations and outputs here: https://nbviewer.jupyter.org/gist/Tiltedprogrammer/53510850af2e03c5b050fad13355c72b

The question is whether it is possible to specialize cuda code as in cpu case, i.e. with respect to static arrays of data

    nbviewer.jupyter.org
    Notebook on nbviewer
    Check out this Jupyter notebook!

Owner
5:27 PM
Can you give the input source used for compilation?
Your first examples doesn't type.
For the others, the generated code (from C++) is missing.
6:01 PM
https://github.com/Tiltedprogrammer/spec

I built it from build dir running
cmake .. -DAnyDSL_runtime_DIR="PATHtoRUNTIME"
then make cd src ./fun 1 2 3 4 5 6 7 8

    github.com
    Tiltedprogrammer/spec
    Contribute to Tiltedprogrammer/spec development by creating an account on GitHub.

7:09 PM
Finally I got along with the part of this partial evaluation stuff which was about cuda code. The PE indeed specializes cuda code and the problem was that I somehow had two llvm outputs for cpu code with different llvm optimisation levels (llvm seems to optimise some thing really well ), thus I had different outputs for cpu but not cuda. Another trick was to use unroll instead of range, so cycle in cuda code eventually got folded to return value. https://nbviewer.jupyter.org/github/Tiltedprogrammer/spec/blob/master/Feedback.ipynb
Also I updated the sources.

The thing that is still unclear is that PE produces same results for
fn get42 function no matter whether it has any annotations or not. I.e. fn @(false)get42(keys : Keys, values: &[u8], key: u8) -> u8 {, fn get42(keys : Keys, values: &[u8], key: u8) -> u8 , fn get42(@keys : Keys, values: &[u8], @key: u8) -> u8
produce the same result. So what is the purpose of annotating once again? And maybe there is a way to not specialize at all, like to be able to see the difference between specialized and unspecialized versions?

    nbviewer.jupyter.org
    Notebook on nbviewer
    Check out this Jupyter notebook!

7:09 PM
Has joined the channel.
Owner
7:12 PM
How many uses of
get42
do you have in your code? I guess our inliner inlines this function. This happens regardless of any annotations that are relevant to the partial evaluator
Owner
4:15 PM
gerwant So everything works then as expected?
4:33 PM
Up to my current level of progress, yes. And in case of multiple calls for
get42
annotations indeed begin to matter. Thanks 🙃
4:10 PM
Hello once again!
Are there any implicit limitations on cuda
block_size and grid_size inside impala? As far as I understand there is a requirement that grid_size should be multiple of block_size. However I have run into some unexpected behavior: impala seems to strip the number of specified threads to launch or simply doesn't launch them.

Considering the following code for substring matching:



             let block = (2, 1, 1);

             let grid = (6, 1, 1);

             //single pattern for now // patterns will be continious arrays of chars
             with cuda(0, grid, block) {
                 // t_id is the position in the buffer
                //  let t_id = cuda_threadId_x() + cuda_blockDim_x() * cuda_blockId_x();
                let t_id = cuda_threadIdx_x() + cuda_blockDim_x() * cuda_blockIdx_x(); //dynamic :(
                if t_id < ibuffer_size {
                    let mut matched : i8 = 1i8;
                    result_buf(t_id) = -1;

                    for i in unroll(0,template.size){
                        if ibuffer(t_id + i) != template.array(i) {
                            matched = -1i8;
                        }
                    }
                    if matched == 1i8 {
                        result_buf(t_id) = 1;
                    }


                }

             }
             synchronize_cuda(0);


the grid has 12 threads that are sufficient for input strings up to 12 chars. However when the function is called on string of size 7, the 7th thread is not laucned which results in result_buf(7) = 0 instead of -1 or 1. The same code written in cuda produces the expected output.

Complete code is available here: https://github.com/Tiltedprogrammer/spec/tree/matching/src

to build impala version: go to spec dir and run mkdir build cmake .. -DAnyDSL_runtime_DIR="PATHtoRUNTIME" cd src ./fun aaaa aaaaaaa which should've produced 1111-1-1-1, i.e. positions where substring aaaa has occurrences, however produces 1111-1-10, while ./fun aaaa aaaaaa gives a correct result

to run cuda code: go to spec/src nvcc -o fun test.cu ./fun aaaa aaaaaaa
will give correct result with the same block&grid sizes.

P.S. increasing block_size or grid size solves the problem though.

    github.com
    Tiltedprogrammer/spec
    Contribute to Tiltedprogrammer/spec development by creating an account on GitHub.

Owner
4:17 PM
I am confused. Your grid size in the provided example specifies 6 work_items, processed by groups of 2, not 12 elements. Maybe this is why you observe the behavior you described?
(We follow the convention that, as in e.g. CUDA, the block size and the grid size are expressed in terms of threads)
4:30 PM
So in cuda, e.g.
match<<<6,2>>>(...) results in 12 threads spawned (6 blocks of 2 threads) and is not equivalent to

        let block = (2, 1, 1);

             let grid = (6, 1, 1);

             //single pattern for now // patterns will be continious arrays of chars
             with cuda(0, grid, block) ....

?
Owner
4:33 PM
You are right, those things are not equivalent. I think my comparison is incorrect 😉 The convention we use is not the one of CUDA, but the one used by clEnqueueNDRangeKernel, where the grid size is given in number of work-items (threads for CUDA).
4:58 PM
Ah, the
match<<<6,2>>>(...) should be equivalent to

        let block = (2, 1, 1);

             let grid = (12, 1, 1);

             //single pattern for now // patterns will be continious arrays of chars
             with cuda(0, grid, block) {


i.e. 12 threads with 2 threads per block(6 blocks), right?
Owner
4:59 PM
Yes.
4:59 PM
Thanks! 🙃
Owner
5:03 PM
Welcome~
7:12 PM
I've finally finished some of specialization workaround for strings patterns matching and done some time measurements.
The results are summarized in Notebook there: https://nbviewer.jupyter.org/gist/Tiltedprogrammer/b155474903d604b9cedb06cf67e8c878

Tests were run on random subjects strings from 1MB to 100MB partially evaluating with respect to random pattern strings of length 18. The thing I noticed that the very first run of the program is reasonably slower than consequent runs even on different patterns/subject strings. Surely I don't include JIT compilation time. E.g. on subject string of size 100MB and newly generated pattern string pattern matching takes 1275ms, while the run after this on a different subject string takes only 59ms. I know that in that case JIT part is loaded from cache, but I take measurement only around of the call site. So, I wonder why this happens and why programs where JIT part is loaded from cache are faster compared to that where JIT is actually performed with respect to call site (i.e. not including compilation time)

Another issue I have is that PE aborts on patterns which length is more than 18 bytes, the screenshot of what happens is provided at the end of the Notebook.

    nbviewer.jupyter.org
    Notebook on nbviewer
    Check out this Jupyter notebook!

Owner
7:33 PM
It's a bit hard to understand what's going on there just looking at the output and those graphs.
However, as a general guideline you can try to increase stack size for large PE problems
With e.g.
ulimit -s 65536
(to show the current stack size, use
ulimit -s
)
This is because PE uses recursion in the current version of Thorin, and thus may reach a stack overflow for very large problems.
You also have to understand how the runtime works.
Right now, you probably get a
.o and .nvvm.bc file from Impala. The runtime has to transform that into a .ptx
file, and that happens at run-time, the first time the kernel is called.
The resulting
.ptx
file is then cached so that subsequent runs are faster.
Also note that right now you are using the
Debug version of the runtime, which will make things slower. Compiling runtime in Release mode will make it a bit faster, but I'm afraid that you'll need to recompile LLVM entirely in Release mode to be able to compile from .nvvm.bc to .ptx
faster.
(that's because we use the
NVPTX backend in LLVM to do the .nvvm.bc to .ptx
translation)
From that, you see that taking the measurements from the call site means that you
will
include the compilation times in your timings
What you need to do is to ignore the first iteration, or do an empty run (e.g. if you have a size parameter, set it to zero so that the kernel exits immediately upon launch).
Then you are sure that the
.ptx
module has been compiled, is in memory and is available to the driver.
Even if the
.ptx kernel is taken from the cache, there is still an overhead on the first iteration since we have to open the file from disk and give it to the driver which will translate the .ptx
to the internal instruction set of your GPU.
TLDR: Just do a dry run before benchmarking with anydsl.
8:25 PM
Thanks a lot, stacksize indeed solves these exceptions so far
9:28 PM
Clipboard - 25 сентября 2019 г., 22:27
9:28 PM
Actually it seems not to solve )
9:29 PM
photo_2019-09-25_21-43-05.jpg
Owner
10:53 AM
the
poor hash function
issue may be a bug or just coincidence. the second one is defintely a bug. how can I reproduce?
3:13 PM
HOW TO REPRODUCE:
1)
git clone https://github.com/Tiltedprogrammer/spec cd spec && git checkout matching
2) mkdir build && cd build && cmake .. -DAnyDSL_runtime_DIR:PATH={PATH_TO_RUNTIMECMAKE}
3) make
4) cd .. && mv data/ build/src
5) cd build/src && ./fun hkfhsdkhfksdhkfhdk (the length of pattern is 18) gives correct output
6) ./fun leerqakjvtjuzayvvzaby
(the length of pattern is 19 and for any with length > 19 it hold and eat all the RAM or crashes)

    github.com
    Tiltedprogrammer/spec
    Contribute to Tiltedprogrammer/spec development by creating an account on GitHub.

Owner
3:13 PM
thanks. I will look into that
Owner
3:46 PM
okay, the
poor hash function
issue is just bad luck. all the hashes are different so it's just bad luck that we end up with that many collisions. so you can ignore this error. the size of the hash set in this example is only 12. so not a big deal.
I will look into your other issue later
5:14 PM
What happened to the hls_top branch in the runtime?
Owner
7:22 PM
it's still there, what's the problem?
7:24 PM
Nevermind. 🐸
9:45 PM
hello!
I have some troubles partially evaluating KMP algorithm, which follows already matched string(ams) logic and prefix_function to move through the subject string. AMS is always between 0 and the size of the pattern, so there must be finite function calls that could be specialized. However my variant eats all the ram while compiling, so what can I look at? Can I manually terminate the partial evaluation or give some hints to pe? (I want the template string to be flattened through the code, i.e. embedded)

https://pastebin.com/rGD0UNUF
I call the function that way :
kmp(template,t_size,left_bound,left_bound,right_bound,ibuffer,result_buf,chunk,0)

    pastebin.com
    [Rust] fn kmp(@template : &[u8], @template_size: i32, text_index : i32, left_bound : i3 - Pastebin.com

Also I have another problem:
I have similar function to search a pattern in a text that specializes well. However compiler produces
2MB .cu file for it with template string of length 3 (bytes), while with all other sizes of input .cu files
are in bounds of several kb. Can share the sources if it helps.
9:58 PM
and as for this trouble, I have rebuilt anydsl in release build and now it just consumes all the ram on inputs with length more than 18 🙃
So there might be some kind of combinatorial explosion of states or smth, however all the parameters are bound with values about 20 and it seems strange that everything is fine with input of size 18 and is terrible with one of 19.
gerwant September 25, 2019 9:29 PM
photo_2019-09-25_21-43-05.jpg
Owner
10:25 AM
Can you give a full-fledged minimum example?
e.g. a main that includes a call to
kmp(template,t_size,left_bound,left_bound,right_bound,ibuffer,result_buf,chunk,0)
which arguments are known, which are not, etc.?
8:22 PM
yeah, the whole template (and its size consequently) is known and I call it utilizing jit compiling feature:

 std::string match_KMP_chunk;
    match_KMP_chunk += "extern fn dummy(text : &[u8], text_size : i32, result_buf : &mut[i32]) -> (){\n";

    match_KMP_chunk += "  *match_kmp*(\"" + pattern + "\","
              + std::to_string(pattern_size) + ",32i8 ,text, text_size,result_buf,256,256,0)}"; //;


    std::string program = std::string((char*)fun_impala) + program_;
    auto key = anydsl_compile(program.c_str(),program.size(),0);
    typedef void (*function) (const char*, int, const int *);
    auto call = reinterpret_cast<function>(anydsl_lookup_function(key,"dummy"));
    // cuda allocations
    call(dtext,text_size,dresult_buf); //dtext is device global memory containing the text we want to search a pattern in, dresult_buf --- global memory where store result to
   // move results to host
 // deallocations



where match_kmp does the next thing:

fn match_kmp(@template : &[u8], @t_size: i32 ,@maximalPatternSize: i8, ibuffer : &[u8], @ibuffer_size : i32, result_buf : &mut[i32], @block_size : i32, @chunk_size: i32, @nochunk : i32) -> (){
    if nochunk == 0 {
        let block = (1024,1,1); //1024
        let chunk = chunk_size; //experiment with it
        let grid_size : i32 = (((ibuffer_size + chunk - 1) / chunk) + block(0) - 1) / block(0);
        let grid = (grid_size * block(0),1,1);

        with cuda(0,grid,block) {

            let t_id = cuda_threadIdx_x() + cuda_blockDim_x() * cuda_blockIdx_x();

            let left_bound = t_id * chunk;
            let mut right_bound = left_bound + chunk + t_size-1; //right not included

            if right_bound >= ibuffer_size {
                right_bound = ibuffer_size;
            }

            if left_bound < ibuffer_size {

                kmp(template,t_size,left_bound,left_bound,right_bound,ibuffer,result_buf,chunk,0)
            }
        }

3:36 PM
Clipboard - 24 октября 2019 г., 16:36
3:36 PM
Hi there!
I've got to some experiments about pattern matching and I am currently doing some measurements of different implementations of pattern matching algorithms. Everything worked fine on my machine, but I have access to a more powerful one. However compilation from generated cu files to ptx fails with the following message. Maybe someone has encountered?
Owner
3:43 PM
The GPU in your other machine doesn't support one of the instructions / intrinsics / assembly you use.
https://docs.nvidia.com/cuda/parallel-thread-execution/index.html#changes-in-ptx-isa-version-6-2

    docs.nvidia.com
    PTX ISA :: CUDA Toolkit Documentation
    The programming guide to using PTX (Parallel Thread Execution) and ISA (Instruction Set Architecture).

3:51 PM
so it is somewhere in low-level generated code and I can't really do anything with it?
Owner
3:58 PM
most likely it's one intrinsic you use
we generate a .cu or .nvvm file which shouldn't contain such intrinsics
when you execute your program, the .cu / .nvvm files are compiled to .ptx for the current GPU you're executing on and stored to cache. The driver compiles then the .ptx to the actual assembly. The next time you execute the program, the .ptx gets loaded from cache.
If you copied also the cache to the other host, it loads the generated .ptx for a different architecture.
can you delete the cache directory and execute again?
2:35 PM
Hi!
The last problem most likely seems to appear due to troubles between cuda tookit and nvidia driver, so nevermind 🙃

I have a question regarding partial evaluation for the following piece of code:

fn match_internal(@t : &[u8],@start : i32,@end : i32,id :u8, t_id : i64, i:i32,ibuffer : &[u8],result_buf:  &mut[u8]) -> () {

            if start == end {
                result_buf(t_id) = id;
                return()
            }else{
                if t(start) != ibuffer(t_id + i as i64){
                    return()
                }else{
                    match_internal(t,start + 1,end,id,t_id, i + 1,ibuffer,result_buf)
                }
            }
}



fn string_match_multiple(@template : &[u8],@t_sizes : &[i32], @t_num : u8, ibuffer : &[u8], ibuffer_size : i64, result_buf : &mut[u8], @block_size : i32) -> (){
   ...
  with cuda(0,grid,block){

        let t_id : i64 = threadId();

        if t_id < ibuffer_size {

            result_buf(t_id) = 0u8;
            let mut start = 0;

            for i in unroll(0,t_num as i32){
                match_internal(template,start,start + t_sizes(i),i as u8 + 1u8,t_id,0,ibuffer,result_buf);
                start = start + t_sizes(i);
            }
        }
    }
}



and everything is called in the following way:

 dummy_fun += "extern fn dummy(text : &[u8], text_size : i64, result_buf : &mut[u8]) -> (){\n";
    dummy_fun += "  string_match_multiple( \"" + patterns + "\","
              + "["+ sizes + "]" + "," + std::to_string(vpatterns.size()) + "u8,text, text_size,result_buf,"+std::to_string(block_size)+")}";
    ....
    auto key = anydsl_compile(program.c_str(),program.size(),0);
    typedef void (*function) (const char*, long, const char *);
    auto call = reinterpret_cast<function>(anydsl_lookup_function(key,"dummy"));
    call(dtextptr,text_size,dresult_buf); //remaining dynamic arguments


The code just does multiple patterns search with all the patterns stored as linear array/single string accessed through offsets.
So I have inplaced static arguments for pe to be able to look into. However it does not completely unroll the access to t(start) in the first piece in if t(start) != ibuffer(t_id + i as i64){ resulting in a following .cu code:

unsigned char _87147;
        _87147 = *_87116;
        array_42738 _87145_10;
        {
        array_42738 _87145_10_tmp = { { 97, 97, 97, 98, 98, 0, } };
         _87145_10 = _87145_10_tmp;
        }


Are there any hints to tell the PE to unroll/inline all the calls completely?
The idea is to get rid of these tmp arrays
and end up with multiple lines of equality tests. For the case with one pattern (where there are no offset accesses and only single unroll is needed ) everything works fine.
Owner
5:10 PM
Can you provide a minimal working example for this?
sorry for the late reply
3:41 PM
Has joined the channel.
3:41 PM
Hi all! Thanks for AnyDSL! I'm working on my own compiler-related project and already learnt alot from your system!
It looks like that the Thorin stage lacks some error checking. For example:

static mut buf: [int*4] = [0, 0, 0, main(0)];
fn main(i: int) -> int {
    buf(i+1) = 5;
    buf(i)
}
Segmentation fault (core dumped)


static mut buf: [int] = [0, 0, 0];
fn main(i: int) -> int {
    buf(i+1) = 5;
    buf(i)
}
void thorin::Def::replace(thorin::Tracker) const: Assertion `type() == with->type()' failed.
Aborted (core dumped)

Owner
3:42 PM
yes, this is a known bug
we are working on that 🙂
3:44 PM
Ok, great! 🙂 Have you considered the adding a simple constant folding pass right inside Impala? For computing the size in array type etc.
Owner
3:44 PM
we don't want to add any optimizations in impala itself
instead we want to integrate thorin more tightly within impala
3:47 PM
Sure, for now the architecture looks very clean. Do you plan to add compile-time execution of functions to generate global data values?
Owner
3:47 PM
yes
3:47 PM
Then it all makes sense, of course 🙂
Owner
3:48 PM
right now, there are just a couple of technical reasons why we can't do that
we're currently working on a huge rewrite. have a look at the
t2
branch. also we're working on a completely rewritten frontend
3:51 PM
Really? Great, I will look into it, thank you!
I have some experience with libfirm, which has very clean code. And while I'm not a big fan of C++ I should say that Impala/Thorin projects are very readable and compact.
Owner
3:52 PM
yes, we know the libfirm guys. thorin is - so to say - the spirtual successor 🙂
the
t2
branch is even more progressive in the sense that we got rid of a couple of more concepts and treat more stuff like the same - if that makes sense
e.g., we got rid of
class Type. Now everything is a Def
3:56 PM
One of my favorite things in libfirm is Sync node. I'm working on code generation/codesign for specialized processors/ASIPs and things like Sync help to extract ILP.
But I guess that Thorin takes place of MIR, not LIR.
Owner
3:58 PM
split/sync is sth that we don't have right now - because we didn't need it. but it's not hard to add.
t2
tries to be sth that fits the bill for many domains. you can model high-level IRs like Lift or XLA, typical middle-end stuff like LLVM or even back-ends like x86. That being said, the stuff that is there is more like a typical middle-end IR. but it's easy to extend
4:03 PM
By the way, in libfirm one can easily parse text-based representation of Firm. I would say that such simple graph description format is much more heplful than LLVM-style language for which you need to write a whole parser.
Owner
4:04 PM
yeah. sth we are still lacking. no one had the patience to do that 😃
4:06 PM
I mean I'm completely ok with that Thorin is MIR. I had a project in which I translated C code with help of libfirm to text-based Firm representation (with all libfirm optimizations inside) and then I used my own backend tools to generate the code for an exotic processor.
Owner
4:06 PM
sounds cool 🙂
4:11 PM
I think such scenario should be really interesting: when I have my own high performance DSL, a ready to use module (Thorin?) for middle-stage ("architecture independent") transformations and my own code generator.
Owner
4:23 PM
I'm also working on a complete new optimizer. this might be interesting for your project. I'll let you know as soon as I'm getting somewhere
5:03 PM
That's great! I'll try to find a chat client to be in sync! 🙂
Owner
5:03 PM
I'm fine with a pinned tab in firefox
5:16 PM
Ok, then I'll do the same in Chrome. I'm looking at the code of Impala t2 branch now. Looks like the architecture is basically the same. I found some new interesting Thorin nodes, like lea_unsafe.
Owner
5:17 PM
the interesting stuff is all in thorin
we will kill the old impala front end sooner or later anyway
5:20 PM
One of things which I'm interested in is the new ways to implement parts of a compiler in a higher-level declarative forms. What do you think about using some DSL to describe, for example, algebraic simplification rules for Thorin? Maybe someday you'll implement them directly in Impala? 🙂
Owner
5:21 PM
this is sth we want to do
the
t2
branch is already fairly extensible. in fact most nodes, +, -, load, store, ptr, and so on are all implemented as extensions
we would "just" need some sugar to do this in impala
but having a complete high-level declarative DSL to specify such rules is a bit too much for my taste. I mean, such a system must be fairly complex to deal with everything from constant folding to several instances of similar but not exactly the same identities
in
t2 you just specify an Axiom. That is a thing that does have a type - typically a function type. So you can App
these things like a function. And you can specify a function pointer - called normalizer - that implements constant folding and all desired identities
have a look at
normalize.cpp
to see how these things look like. and I don't see how some DSL could drastically reduce the boilerplate there. I think it's already concise enough.
5:34 PM
Yes, it's more readable than in libfirm. My idea is to give a user the way to add their own domain-specific transformation rules. Because you just can't provide the whole set of rules for all possible targets and domains.
Owner
5:36 PM
this is what we want to do as well
and the basics are already working
5:36 PM
On a very basic level it may look like this: https://github.com/golang/go/blob/master/src/cmd/compile/internal/ssa/gen/generic.rules
But I'm more interested in a Stratego-like approach. Have you seen Stratego language?

    github.com
    golang/go
    The Go programming language. Contribute to golang/go development by creating an account on GitHub.

Owner
5:37 PM
not yet
5:39 PM
It's a system for strategic term rewriting. Looks somewhat like a Prolog.
Owner
5:40 PM
but reading over the rules, I mean whether I write

(Mul(8|16|32|64)  (Const(8|16|32|64)  [1]) x) -> x


or

if (la == world.lit_int(*w, 1)) {
    switch (op) {
        //...
        case WOp::mul: return b;    // 1  * b -> b
    }
}

things are factored a bit differently, but it's not that one impl is significantly more complex/larger than the other
if you want to do more with your rules, than things are different.
Right now, we assume that you specify one normal form via your rules
5:42 PM
Ok, I think I need to read Thorin t2 more carefully.
5:47 PM
A reason to have rule sets in a highly modular form is to help to automate process of compiler retargeting. It's something like auto-tuning -- you try to turn on and off some rules and see if it helps to generate effective machine code for a new processor. It's not enough just to have a generic set of "architecture independent" rules, because some of these rules may affect badly on a code generation for exotic architectures.
Owner
5:51 PM
yes. that's true. but this is sth we want to look into when everything else works 😃
1:28 PM
HI! Don't you think that the rule in Impala about that shift operators << and >> must have the same type is too restrictive?
Owner
3:40 PM
I don't like implicit casts at all. Most programmers have no clue what's going on and do neither know the rules - when exactly conversions happen - and the semantics - how it is done. For this reason, I'm fine with having to use explicit casts. That being said, why are you using different types to begin with?
6:52 PM
I totally agree with you about explicit typing rules. But I see bit shift operation as a special case where type of the right operand (shift amount) doesn't need to have the same sign as the left operand. In fact, the shift amount should be unsigned.
9:43 AM
Has joined the channel.
2:12 PM
Hi! your project looks super interesting and promising, especially because it is so compact and close to the math 🙂
I am trying to use thorin as a backend for a pseudo-dynamic language which is a Julia variant ... the documentation is a bit lacking or outdated
I managed to get the example in AnyDSL website ,in the thorin programming guide section, to compile by replacing
lambda with continuation
. I do a thorin output but I dont get any llvm output , here is the code:

// create the four functions:
    auto main = world.continuation(world.fn_type(
                                                 {world.type_qu32(), world.fn_type({world.type_qu32()})}),{"main"});
    main->make_external();
    auto if_then = world.continuation(); // defaults to 'fn()'
    auto if_else = world.continuation(); // defaults to 'fn()'
    auto next = world.continuation(world.fn_type({world.type_qu32()}),{"next"});

    // create comparison:
    auto cmp = world.cmp_eq(main->param(0), world.literal_qu32(0,{}));

    // wire lambdas
    main->branch(cmp, if_then, if_else);
    if_then->jump(next, {world.literal_qu32(23,{})});
    if_else->jump(next, {world.literal_qu32(43,{})});
    next->jump(main->param(1), {next->param(0)});

    world.dump();

    auto cg = thorin::Backends(world).cpu_cg.get();
    cg->emit(std::cout,true,true);

I still don't understand what is the difference between a lambda and a continuation. I looked on the t2 branch, where again there is no continuations only
Lam
-bdas?
Owner
3:12 PM
in
master there are only Continuations. In t2 we renamed it to Lam
bdas because those guys can be used in both continuation-passing as well as in direct style. Besides it's less key strokes 😉
Owner
5:35 PM
What is more, in
master the final call of a Continuation is baked in to a Continuation itself. In t2, the body of a Lam that acts as a continuation must be an App that invokes another Lam of type T -> Bot
.
In order to emit llvm you need to do sth like this:

 emit_to_file(backends.cpu_cg.get(),    ".ll");

6:51 PM
So what’s an ‘App’? If I understand correctly , A lambda that has only one other lambda of type ‘’’T -> Bot’’’ in its parameters can be considered a direct style function and can be emitted to llvm and be called externally. All other lambdas are inlined in the process of creating the lambdas of the first kind.
So what does “App” stands for?
Owner
9:58 PM
App
is an application AKA as a function call
Owner
1:04 PM
in
t2 we use direct style functions to model typical instructions (things that used to be PrimOps in master) or types (what used to be a Type). Arbitrary direct-style programs are not supported yet in t2
if you have an arbitrary direct-style function in your frontend, you will cps-convert it into a thorin program
does that make sense?
1:07 PM
Yes it makes sense. Arbitrary in the sense that it accepts function pointers? “Normal” functions on parameters with a typed return can be lowered to llvm right?
As long as it does not have function pointers in its parameters? Or are we not on the same page completely?
Owner
1:11 PM
eh, I find it quite hard to follow. So I guess, we are not on the same page?
1:12 PM
An App is a direct style function?
Or basicblock in llvm?
Owner
1:13 PM
an
App
is just a function application - a jump, a call
a
Lam of type [int, float, bool] -> Bot
is a basicblock
a
Lam of type [int, float, bool, int -> Bot] -> Bot the cps-converted version of a direct-style function of type [int, float, bool] -> int
1:14 PM
I see, and Lam with one other Lam of the type you just wrote, can be called externally
Owner
1:15 PM
the
int -> Bot
above is the "return"
yes, you can call an
external Lam of type [int, float, bool, int -> Bot] -> Bot
from C, for instance
1:16 PM
Ok and if a LAM has more than one “return” Lambdas ... it must be inlined when going to llvm right?
Owner
1:16 PM
yes
1:17 PM
Got it. A LAM can either branch or call an APP?
Owner
1:17 PM
short answer: yes
1:18 PM
I am trying to model a dynamic language where there could be more than one return path
Owner
1:19 PM
with "more than one return path" you mean several
return
s as in C, or different return points as checked exceptions (as in Java)?
1:20 PM
The latter let’s say, a function that either returns a value of type T or an error of type ET
As long as the different number of types is small, cps can be used to avoid dynamic dispatch
Owner
1:22 PM
the trick we are doing on
master
is basically aggressive inlining. but there are problems with this. I'd like to change that in our current rewriting and also support proper "exceptions". LLVM at least has support for this
1:25 PM
I am thinking on the lines of letting the backend choose where to draw the line with a concrete function call and where to inline. Generated code does not need to follow the written code.
The backend may choose to introduce aux functions to reduce code size.
I did a small test on Julia code , where I converted a “poorly written” piece of code that has divergence in return types ... into cps form, and the resultant code was much faster.
In a sense inlining the CPS way works backward .. you inline the big computation into the small function. That let’s you do dynamic dispatch at the source of the “dynamicism” instead of handling return values that might be of different types
Owner
1:37 PM
instead of returning an optional or variant or sth like that?
1:46 PM
Yes, and if a function is needed externally expose a wrapper that returns a tagged union or a variant
Owner
1:46 PM
makes sense
11:45 PM
Has joined the channel.
11:45 PM
Message removed
11:49 PM
Hello everyone, I am just started to work with any DSL. I want to know if t is possible to pass a reference to an object in c++ in impala and then call methods of the object in impala function body. I do not have any sense yet. Just investigating. Thank you
Owner
11:51 PM
No, that's not possible.
We only support the C ABI, no C++.
This is the same for almost any language interfacing C/C++.
11:54 PM
Is there any solution?
RUST either. But there are some work around.
Owner
11:58 PM
yes, you can re-implement them in AnyDSL
what do need them for?
12:02 AM
We are working on system in which we expose some capabilities. We want to use a DSL by which we could pass capabilities references by taking rights of using those capabilities for computation in realtime.
Owner
12:06 AM
I see, basically we only support C ABI and the C++ ABI is not cross platform since its handled in the frontend, which is platform and compiler specific.
you can encode your class in an id and pass a void pointer and try to reinterpret this, but this is not portable ...
12:17 AM
Thank you. I was already thinking about this solution. I will give it a try. And now I am thinking about a bridge/adapter pattern which is not bound to ABI problem. I will let others to know about it if feasible. Thanks.
2:48 PM
https://gist.github.com/DasNaCl/6aebdcc3c36456f6df8836d8d05c6fad

Something along the lines of this, I guess. However, you don't need to use a
&[u8] as handle, it could very well be just an i32
. If your C++ code has a memory arena where you basically just "look the index up" I don't see any issues. Moreover, this can eliminate the explicit free in Impala. But, to get a "fully automagic" look and feel of the memory management, I presume the entry point needs to be on the C++ side.

    gist.github.com
    Calling methods of a C++ struct "in Impala"
    Calling methods of a C++ struct "in Impala". GitHub Gist: instantly share code, notes, and snippets.

February 11, 2020 9:12 PM
9:11 PM
Message removed
9:21 PM
How can we pass impala function arguments through C/C++ before running? I mean before a function get called, its arguments could be passed through C/C++ code? Consider a decoding stage in a RISC architecture pipelining in which before executing instruction, it determines what data to use by execution stage. This is somehow identical to the question I already asked.
7:10 AM
Not sure if I understand the question correctly. Anyhow, you'll need to expose an Impala or a C function. Impala functions are exposed to C with
extern, e.g. extern fn pass_arguments_to_impala_from_c(/*parameters*/);
. For the other way around, see my gist above.
February 12, 2020 9:50 AM
9:14 AM
Is there any good document for Impala/thorin?
Owner
8:13 PM
omid.shahraki did you read the AnyDSL papers or https://anydsl.github.io ?

    anydsl.github.io
    AnyDSL - A Partial Evaluation Framework for Programming High-Performance Libraries - AnyDSL

February 12, 2020 8:26 PM
7:49 PM
Hi! I have run into some troubles when running impala programs that utilize JIT compilation after an update in Thorin repo. More specifically I have created a fresh Ubuntu vm and installed AnyDSL from https://github.com/AnyDSL/anydsl. After trying to build some AnyDSL application that needs AnyDSL_runtime via
cmake .. -DAnyDSL_runtime_DIR cmake shows an error

Make Error at {PATH_TO_ANYDSL_FOLDER}/runtime/build/share/anydsl/cmake/anydsl_runtime-config.cmake:135 (list):
  list sub-command REMOVE_ITEM requires list to be present.
Call Stack (most recent call first):
  CMakeLists.txt:4 (find_package)



So I have to comment out the 135th line in {PATH_TO_ANYDSL_FOLDER}/runtime/build/share/anydsl/cmake/anydsl_runtime-config.cmake, i.e. list(REMOVE_ITEM _release_configs ${_debug_configs}), so the error is eliminated and everything is built fine.

However, after running an application I got error when entering anydsl_compile() function

filter_spec.out: /home/bolkonskiy322_gmail_com/anydsl/anydsl-release-last/anydsl/thorin/src/thorin/util/cast.h:26: L* thorin::scast(R*) [with L = thorin::Continuation; R = thorin::Def]: Assertion `(!r || dynamic_cast<L*>(r)) && "cast not possible"' failed.
Aborted (core dumped)


(here AnyDSL is built in Debug mode, if the mode is set to Release it just aborts with a Segfault)

Since before some commits in your projects everything worked fine I played around a bit with checkouting in ./setup.sh script and found out that a checkout to d93d0defe7b4e0d4ccaec051ca974e02a0b1dc2a commit in Thorin solves the problem (i.e. the commit before the last one)

However, that behavior was observed on my AnyDSL applications and on AnyDSL regex project (it also utilizes JIT) everything worked fine even with the latest commits in Thorin.

My AnyDSL applications are available here: https://github.com/Tiltedprogrammer/spec with build instructions at the very bottom.
I hope the error could be reproduces with running cd build/convolution/impala/ && ./filter_spec.out --fsize=3 --isize=1024
after making the project

Don't really know whether this is a problem or not, just a very strange behavior (since Regex project works fine, however my programs utilize cuda).

    github.com
    Tiltedprogrammer/spec
    Benchmarks for partial evaluation of different GPU application scenarios - Tiltedprogrammer/spec

Owner
9:07 AM
gerwant do you also have the latest runtime commits?
The error should have been fixed: https://github.com/AnyDSL/runtime/commit/5941939efe13443a7ca939b5a24729665f19268c

    github.com
    Fix: CMAKE_CONFIGURATION_TYPES not always defined. · AnyDSL/runtime@5941939
    AnyDSL Runtime Library. Contribute to AnyDSL/runtime development by creating an account on GitHub.

Owner
2:23 PM
The JIT error is now also fixed in the runtime: https://github.com/AnyDSL/runtime/commit/8748946762f797582741781ab8ba9ca03a466524

    github.com
    JIT bug fix: AnyDSL/thorin@1378c499 · AnyDSL/runtime@8748946
    AnyDSL Runtime Library. Contribute to AnyDSL/runtime development by creating an account on GitHub.

Does this solve all your problems?
10:00 PM
The last commit indeed solved the abort error, thanks!
Cmake still shows the error though (Yes, I have all the latest runtime commits, however the error is still solved through commenting out a line, so no big deal)
Owner
8:32 AM
Did you set the Build type for your application?
specifying
cmake ../ -DCMAKE_BUILD_TYPE=Release -DAnyDSL_runtime_DIR=...
?
February 28, 2020 5:06 PM
Owner
5:15 PM
gerwant we have updated the runtime & cmake logic to handle this properly
you will need at least CMake 3.11 now - our anydsl setup script can be configured to also download and build CMake if required
5:07 PM
With the last commits everything now works fine just like it used to before with only specifying
-DAnyDSL_runtime_DIR=...
. Thanks!
4:17 PM
Hi! Is there any intrinsic to fetch cuda texture memory from inside the impala? Or any workaround to include cuda function wrapper into impala code? E.g. something like that: wrapper.cu file containing:

texture<int,1,cudaReadModeElementType> texRef;
extern "C" {
    int texFetchWrapper(int index){
         return tex1Dfetch(texRef,index);
    }
}



And corresponding impala file with something like:

extern "C"{
   fn texFetchWrapper(i : i32) -> i32
}
...
...
with cuda(0,block,grid){
       ....
       let fetched = texFetchWrapper(...);
}



I tried a simple example:
main.cu:

#include <stdio.h>
extern "C" {
    void println(char * str){
        printf("%s\n",str);
    }
}

extern "C" void hello();


int main(int argc, char** argv) {

    hello();

    return 0;
}


and test.impala

extern "C" {
    fn println(&[u8]) -> ();
}

extern
fn hello() -> () {
    let block = (1,0,0);
    let grid = (1,0,0);
    with cuda(0,block,grid){
        println("Hello World!");
    }
}


Invocation of compiled program fails with build/src/./test.cu(18): error: identifier "println" is undefined cause println has not been placed inside generated test.cu file. Also would be great to get this scenario to work with JIT feature, i.e. invocation of impala code above through anydsl_compile
.
Owner
5:16 PM
What do you need textures for - do you want to make use of the texture cache?
If you only want to use the texture cache, there is no need to use textures.
Instead, you can use the __ldg() intrinsic in CUDA / NVVM:
https://docs.nvidia.com/cuda/cuda-c-programming-guide/index.html#ldg-function

    docs.nvidia.com
    Programming Guide :: CUDA Toolkit Documentation
    The programming guide to the CUDA model and interface.

Those are available in our runtime:
https://github.com/AnyDSL/runtime/blob/master/platforms/intrinsics_cuda.impala#L58
https://github.com/AnyDSL/runtime/blob/master/platforms/intrinsics_nvvm.impala#L384

    github.com
    AnyDSL/runtime
    AnyDSL Runtime Library. Contribute to AnyDSL/runtime development by creating an account on GitHub.

    github.com
    AnyDSL/runtime
    AnyDSL Runtime Library. Contribute to AnyDSL/runtime development by creating an account on GitHub.

12:43 AM
I've been checking out some library for string matching that utilizes texture memory in order to compare with itself being rewritten in impala. However it really seems that using texture memory for the sake of texture cache has become some kind outdated since the version with global memory access just outperformed the one with textures. So, nevermind and thanks.

And for the second part of the question there is no possibility of using external functions inside impala
cuda(...){}
block?
Owner
10:02 AM
https://arxiv.org/pdf/2003.06324.pdf
Looks as we could easily do thin is anydsl as well
Owner
11:27 AM
sebastian.hack I had a quick look: we should be able to do the same in AnyDSL.
The only problem I see is that PE takes way longer than what they do, in particular for evaluating thousands of variants.
1:01 PM
Hi guys, just a very quick question. How to import an Impala file as a module into another Impala?
Owner
1:18 PM
omid.shahraki We don't have support for modules in Impala. You need to compile everything in one unit.
March 16, 2020 4:25 PM
What you can do is to compile individual files and mark interface functions as extern. However, this will prevent from specialization across functions from different compilation units.
Owner
2:36 PM
gerwant You can use printf in CUDA, no need to implement it yourself:

extern "C" {
    fn "printf" cuda_printf(&[u8]) -> ();
}
cuda(..., || { cuda_printf("bla"); });

Owner
4:27 PM
yes, we are aware of this and having a module system is definitely a mid-term goal
March 16, 2020 4:34 PM
Owner
8:30 PM
I think for urgent feature requests you should use the issue tracker on github
March 16, 2020 9:15 PM
1:34 PM
Does AnyDsl Impala supports RUST fully? I was thinking to use a RUST library in order to embed C++ into RUST and use it in Impala.
Owner
6:19 PM
Not at the moment. This is something we are discussing ...
9:38 PM
What extra features Impala 2 supports? Any documents for it? Thanks.
Owner
9:39 PM
https://anydsl.github.io/Impala.html

    anydsl.github.io
    Impala - AnyDSL

March 29, 2020 9:40 PM
impala2 is just an experiment. not meant for a broader audience
9:41 PM
For example, does it support Self in traits?
We can define a trait in Impala but what's its usage?
Owner
9:43 PM
impala1: not implemented
impala2 is just an experiment. not meant for a broader audience
9:44 PM
So, you are going to do code clean up in Impala1 or extend to Impala2?
Owner
10:53 PM
yes. sth along these lines. currently, we are working on a new rewritten frontend
March 30, 2020 11:50 AM
12:07 PM
I was thinking of a bit more loose implementation which could be useful in distributed systems. Consider a capability mechanism by which variety of services could be implemented based on requirements. Suppose AnyDsl runtime has been used as one of the capabilities and the kernel of choice is cpu platform. Such distributed capability based security has been developed fully in C++. Now consider a node in network which do marshalling to send an Impala script as a view to another node for execution. In addition, the Impala script could query data from other capabilities which is the main concern. Such query through Impala script should be loose enough for extensible, lower costs, lower time to market and implementation purposes. This should be in a manner which is ABI independent either. And this is what I was thinking of.
4:59 PM
Has joined the channel.
5:21 PM
hi guys! I have a question regarding the building of anydsl (following https://anydsl.github.io/Build-Instructions.html): is the build of llvm 8.0 needed (a) to be independent of any system packages/operating systems or (b) is llvm patched for the anydsl build?
Owner
5:23 PM
depends 🙂
at least for vectorization we need a custom build of llvm
5:24 PM
lol, that was the answer I feared
Owner
5:24 PM
you might also need RTTI to be enabled
5:24 PM
RT-what?
Owner
5:24 PM
for amdgpu we use their llvm branch
Run-time type information (RTTI)
5:25 PM
thx
Owner
5:25 PM
if your system-llvm is build with RTTI enabled, then you're fine
most likely it's not
5:25 PM
how can I check it?
Owner
5:27 PM
llvm-config --has-rtti
5:27 PM
YES - thx again
9.0.1 on archlinux
Owner
5:28 PM
so, then you should be able to use system clang/llvm, but without vectorization support
5:29 PM
is there something in the config.sh file to achieve this?
I just want to get familiar with anydsl, so vectorization is not strictly needed
rome wasn't built in a day ...
Owner
5:30 PM
compiling llvm takes about 30 minutes on a recent machine
it will give you a working system
you will need a mixture of master, llvm_90, and llvm100 branches of impala, runtime, and thorin ... with a modified anydsl setup script
we're on LLVM 8.0 right now and planning to move to LLVM 10.0, skipping LLVM 9.0
5:54 PM
no problem - I was interested in the dependencies rather than a speedup. stability first ...

-- Build files have been written to: /home/ram/src/tools/anydsl/runtime/build
make[1]: Entering directory '/home/ram/src/tools/anydsl/runtime/build'
make[2]: Entering directory '/home/ram/src/tools/anydsl/runtime/build'
Scanning dependencies of target runtime
make[2]: Leaving directory '/home/ram/src/tools/anydsl/runtime/build'
make[2]: Entering directory '/home/ram/src/tools/anydsl/runtime/build'
[ 25%] Building CXX object src/CMakeFiles/runtime.dir/runtime.cpp.o
[ 50%] Building CXX object src/CMakeFiles/runtime.dir/opencl_platform.cpp.o
[ 75%] Building CXX object src/CMakeFiles/runtime.dir/cpu_platform.cpp.o
In file included from /home/ram/src/tools/anydsl/runtime/src/log.h:5,
                 from /home/ram/src/tools/anydsl/runtime/src/platform.h:5,
                 from /home/ram/src/tools/anydsl/runtime/src/opencl_platform.h:4,
                 from /home/ram/src/tools/anydsl/runtime/src/opencl_platform.cpp:1:
/usr/include/c++/9.3.0/cstdlib:75:15: fatal error: stdlib.h: No such file or directory
   75 | #include_next <stdlib.h>
      |               ^~~~~~~~~~
compilation terminated.
make[2]: *** [src/CMakeFiles/runtime.dir/build.make:106: src/CMakeFiles/runtime.dir/opencl_platform.cpp.o] Error 1
make[2]: *** Waiting for unfinished jobs....
In file included from /usr/include/c++/9.3.0/bits/stl_algo.h:59,
                 from /usr/include/c++/9.3.0/algorithm:62,
                 from /home/ram/src/tools/anydsl/runtime/src/runtime.cpp:1:
/usr/include/c++/9.3.0/cstdlib:75:15: fatal error: stdlib.h: No such file or directory
   75 | #include_next <stdlib.h>
      |               ^~~~~~~~~~
compilation terminated.
make[2]: *** [src/CMakeFiles/runtime.dir/build.make:80: src/CMakeFiles/runtime.dir/runtime.cpp.o] Error 1
make[2]: Leaving directory '/home/ram/src/tools/anydsl/runtime/build'
make[1]: *** [CMakeFiles/Makefile2:139: src/CMakeFiles/runtime.dir/all] Error 2
make[1]: Leaving directory '/home/ram/src/tools/anydsl/runtime/build'

any idea?

make[2]: Entering directory '/home/ram/src/tools/anydsl/runtime/build'
[ 25%] Building CXX object src/CMakeFiles/runtime.dir/runtime.cpp.o
cd /home/ram/src/tools/anydsl/runtime/build/src && /usr/bin/c++  -DAnyDSL_runtime_EXPORTS -isystem /usr/include  -I/home/ram/src/tools/anydsl/runtime/build/include  -g -fPIC   -Wall -Wextra -fvisibility=hidden -std=gnu++14 -o CMakeFiles/runtime.dir/runtime.cpp.o -c /home/ram/src/tools/anydsl/runtime/src/runtime.cpp
In file included from /usr/include/c++/9.3.0/bits/stl_algo.h:59,
                 from /usr/include/c++/9.3.0/algorithm:62,
                 from /home/ram/src/tools/anydsl/runtime/src/runtime.cpp:1:
/usr/include/c++/9.3.0/cstdlib:75:15: fatal error: stdlib.h: No such file or directory
   75 | #include_next <stdlib.h>
      |               ^~~~~~~~~~
compilation terminated.

works when I use
-I instead of -isystem
though
and of course when I leave out this option completly
Owner
6:12 PM
try to remove the SYSTEM from line 87 in src/CMakeLists.txt
6:16 PM
thx I managed to hack it
hello world works
Owner
6:19 PM
great!
6:23 PM
is there a
spack
package for anydsl (https://github.com/spack/spack)
it's a port-system designed for HPC
Owner
7:17 PM
nope
Owner
3:32 PM
question regarding the runtime: Is there a spedific reason why the size of an
alloc is given by an i32. The buffer size is i64
. And 4GB of data (i32) can be too small in some settings
Owner
3:48 PM
I guess no particular reason apart from saving casts.
The actual implementation uses
int64_t for the size.
