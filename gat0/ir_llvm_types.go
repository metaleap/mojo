package main

const (
    llBuiltinLitTrue     = "true"
    llBuiltinLitFalse    = "false"
    llBuiltinLitNull     = "null"
    llBuiltinLitZeroInit = "zeroinitializer"
    llBuiltinLitUndef    = "undef"
    llBuiltinLitPoison   = "poison"
)

type LlIntrinsic int

const (
    _ LlIntrinsic = iota
    abs
    smax
    smin
    umax
    umin
    memcpy
    memcpy_inline
    memmove
    memset
    sqrt
    powi
    sin
    cos
    pow
    exp
    exp2
    log
    log10
    log2
    fma
    fabs
    minnum
    maxnum
    minimum
    maximum
    copysign
    floor
    ceil
    trunc
    rint
    nearbyint
    round
    roundeven
    lround
    llround
    lrint
    bitreverse
    bswap
    ctpop
    ctlz
    cttz
    fshl
    fshr
    overflow_sadd
    overflow_uadd
    overflow_ssub
    overflow_usub
    overflow_smul
    overflow_umul
    sat_sadd
    sat_uadd
    sat_ssub
    sat_usub
    sat_sshl
    sat_ushl
    fix_smul
    fix_umul
    fix_sat_smul
    fix_sat_umul
    fix_sdiv
    fix_udiv
    fix_sat_sdiv
    fix_sat_udiv
    canonicalize
    fmuladd
    converttofp16
    convertfromfp16
    sat_fptoui
    sat_fptosi
    trap
    trap_debug
    trap_ubsan
    donothing
)

type LlNamed struct {
    name string
}

type LlCommented struct {
    comment string
}

type LlTopLevel struct {
    source_filename string

    ExtDecls   []LlTopLevelExtDecl
    FuncDefs   []LlTopLevelFuncDef
    GlobalVars []LlTopLevelGlobalVar
}

type LlTopLevelGlobalVar struct {
    LlNamed
    LlCommented
    constant bool
    init     LlExpr
    ty       LlType
}

type LlTopLevelExtDecl struct {
    LlNamed
    LlCommented
    intrinsic LlIntrinsic
    ty        LlType // unless intrinsic, must be *LlTypeFunc
}

type LlTopLevelFuncDef struct {
    LlNamed
    LlCommented
    ty     LlTypeFunc
    blocks []LlBlock
}

type LlFuncParam struct {
    LlNamed
    ty LlType
}

type LlType interface{}

type LlTypeVoid struct{}

type LlTypeFunc struct {
    ret    LlFuncParam
    params []LlFuncParam
}

type LlTypeInt struct {
    bitWidth int
}

type LlTypeFloat struct {
    bitWidth int
}

type LlTypePtr struct {
    elemTy LlType
}

type LlTypeArr struct {
    numElems int
    elemTy   LlType
}

type LlTypeStruct struct {
    fields []LlType
}

type LlBlock struct {
    LlNamed
    LlCommented
    instrs []LlInstr
}

type LlInstr interface{}

type LlInstrSsa struct {
    name  string
    value LlInstr
}

type LlInstrRet struct {
    expr LlExpr
}

type LlInstrBr struct {
    cond    LlExpr
    ifTrue  *LlBlock
    ifFalse *LlBlock
}

type LlInstrSwitch struct {
    scrut LlExpr
    def   *LlBlock
    cases []struct {
        intConst LlExpr
        dst      *LlBlock
    }
}

type LlInstrUnreachable struct {
}

type LlInstrOp1Fneg struct {
    op1 LlExpr
}

type LlInstrOp2 struct {
    op1 LlExpr
    op2 LlExpr
}

type LlInstrOp2Wrappable struct {
    LlInstrOp2
    noUnsignedWrap bool
    noSignedWrap   bool
}

type LlInstrOp2Add LlInstrOp2Wrappable
type LlInstrOp2Sub LlInstrOp2Wrappable
type LlInstrOp2Mul LlInstrOp2Wrappable
type LlInstrOp2Fadd LlInstrOp2
type LlInstrOp2Fsub LlInstrOp2
type LlInstrOp2Fmul LlInstrOp2
type LlInstrOp2Fdiv LlInstrOp2
type LlInstrOp2Urem LlInstrOp2
type LlInstrOp2Srem LlInstrOp2
type LlInstrOp2Frem LlInstrOp2

type LlInstrOp2Exactable struct {
    LlInstrOp2
    exact bool
}

type LlInstrOp2Udiv LlInstrOp2Exactable
type LlInstrOp2Sdiv LlInstrOp2Exactable

type LlInstrOp2Shl LlInstrOp2Wrappable
type LlInstrOp2Lshr LlInstrOp2Exactable
type LlInstrOp2Ashr LlInstrOp2Exactable

type LlInstrOp2And LlInstrOp2
type LlInstrOp2Or LlInstrOp2
type LlInstrOp2Xor LlInstrOp2

type LlInstrExtractValue struct {
    aggr LlExpr
    idxs []LlExpr
}

type LlInstrInsertValue struct {
    aggr LlExpr
    idxs []LlExpr
    elem LlExpr
}

type LlInstrLoad struct {
    ty     LlType
    ptrSrc LlExpr
}

type LlInstrStore struct {
    LlExpr
    ptrDst LlExpr
}

type LlInstrGetElementPtr struct {
    ty       LlTypeFunc
    ptrSrc   LlExpr
    inbounds bool
    idxs     []LlExpr
    inrange  *int
}

type LlInstrConv struct {
    src   LlExpr
    dstTy LlType
}

type LlInstrConvTrunc LlInstrConv
type LlInstrConvZext LlInstrConv
type LlInstrConvSext LlInstrConv
type LlInstrConvFptrunc LlInstrConv
type LlInstrConvFpext LlInstrConv
type LlInstrConvFptoui LlInstrConv
type LlInstrConvFptosi LlInstrConv
type LlInstrConvUitofp LlInstrConv
type LlInstrConvSitofp LlInstrConv
type LlInstrConvPtrtoint LlInstrConv
type LlInstrConvInttoptr LlInstrConv
type LlInstrConvBitcast LlInstrConv

type LlInstrIcmpEq LlInstrOp2
type LlInstrIcmpNe LlInstrOp2
type LlInstrIcmpUgt LlInstrOp2
type LlInstrIcmpUge LlInstrOp2
type LlInstrIcmpUlt LlInstrOp2
type LlInstrIcmpUle LlInstrOp2
type LlInstrIcmpSgt LlInstrOp2
type LlInstrIcmpSge LlInstrOp2
type LlInstrIcmpSlt LlInstrOp2
type LlInstrIcmpSle LlInstrOp2
type LlInstrFcmpFalse LlInstrOp2
type LlInstrFcmpOeq LlInstrOp2
type LlInstrFcmpOgt LlInstrOp2
type LlInstrFcmpOge LlInstrOp2
type LlInstrFcmpOlt LlInstrOp2
type LlInstrFcmpOle LlInstrOp2
type LlInstrFcmpOne LlInstrOp2
type LlInstrFcmpOrd LlInstrOp2
type LlInstrFcmpUeq LlInstrOp2
type LlInstrFcmpUgt LlInstrOp2
type LlInstrFcmpUge LlInstrOp2
type LlInstrFcmpUlt LlInstrOp2
type LlInstrFcmpUle LlInstrOp2
type LlInstrFcmpUne LlInstrOp2
type LlInstrFcmpUno LlInstrOp2
type LlInstrFcmpTrue LlInstrOp2

type LlInstrPhi struct {
    ty    LlType
    pairs []struct {
        val LlExpr
        dst *LlBlock
    }
}

type LlInstrSelect struct {
    cond    LlExpr
    ifTrue  LlExpr
    ifFalse LlExpr
}

type LlInstrCall struct {
    ty       LlType
    fnTy     LlTypeFunc
    fnPtrVal LlExpr
    args     []LlExpr
}

type LlExpr interface{ expr() (LlType, interface{}) }

type LlExprLitInt struct {
    ty    LlType
    value int
}

type LlExprLitFloat struct {
    ty    LlType
    value float64
}

type LlExprLitCStr struct {
    ty    LlType
    value string
}

type LlExprLitBuiltin struct { // true false null undef poison etc..
    ty   LlType
    name string
}

type LlExprRefGlobal struct {
    ty   LlType
    name string
}

type LlExprRefLocal struct {
    ty   LlType
    name string
}
