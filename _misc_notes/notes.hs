
/*  --  atoms: float int tags
    --  composite:
        --  tagged foo
        --  sequence of foo
            --  n-ary tuple: fixed-length any-typed sequence
            --  string: tagged list of int32
        --  sets:
            --  unary: set of foo
            --  binary: relations or maybe map or maybe record
            --  n-ary: set of n-ary-tuple relations
*/

/*
foo any         :=  True
min num max     :=  _ >= min && _ <= max
min float max   :=  min+0.0 num max
min int max     :=  min+0 num max
uint max        :=  0 num max
true            :=  True
false           :=  False
t list          :=  _   | Link                  ? True
                        | Link: (_foo & _rest)  ? foo t && rest (t list)
*/



true    := True
false   := False
bool    := _    | True  ? True
                | False ? True


check must cmp arg val  :=
    val check cmp arg   | True  ? val
                        | False ? Err msg="must on $T$val not satisfied: $check $cmp $arg"


||  := _    | True  ? True
            | False ? (__   | True  ? True
                            | False ? False)

not := _    | True  ? False
            | False ? True

&&  := _    | False ? False
            | True  ? (__   | False ? False
                            | True  ? True)

cond if True ifTrue False ifFalse :=
    cond    | True  ? ifTrue
            | False ? ifFalse


// -- compose rtl
f2 <. f1 := _ f1 f2

// -- compose ltr
f1 .> f2 := _ f1 f2

// -- force VSO call style
callee: arg := arg callee

// -- force SVO call style
arg.callee := arg callee

// -- id , the well-known: func id(foo) {return foo}
self := _

// -- const
v only _ := v


list first , list must /= Empty :=
    list | Link: (_f & _) ? f

list rest :=
    list    | Link: (_ & _r)    ? r
            | Empty             ? msg="rest: list must not be Empty" Err
    x foo   := (x trim len == 0) if True "(none)" False x


x pow y :=
    y < 0 if True 1 / (x pow y.neg) False x* accum 1 y
    // -- y < 0   | True  ? 1 / (x pow y.neg)
    // --         | False ? x* accum 1 y , tmp := x * _   // -- * accumL 1 (y × x)


f accum initial n , n must >= 0 , x  :=
    True    | n==0  ? f accum x y , x := initial f , _ unused := 123
            |       ? initial

    y := n - 1


a × b /* huh1 */, /* huh2 */  a must >= 0   /* huh3 */ :=
    a == 0  | True  ? Empty
            | False
            | True  ? b ret  // -- should catch such

    foo ret := Link: (foo & ab)
    ab      := a-1 × 3*-b


someRec :=
    Name: (First:"Phil" & Last:"Shoeman") & Age:37
    // -- { Name: { First: "Phil", Last: "Schumann" }, Age: 37 }


f accumR initial list :=
    list    | Empty                     ? initial
            | Link: (_first & _rest)    ? first f (f accumR initial rest)


f accumL initial list :=
    list    | Link                      ? initial
            | Link: (_first & _rest)    ? f accumL (initial f first) rest
