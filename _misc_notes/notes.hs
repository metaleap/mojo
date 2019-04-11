
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
t list          :=  _   ? Link                  : True
                        | Link (_foo & _rest)   : foo t && rest (t list)
*/




check must cmp arg val  :=
    val check cmp arg   ? val
                        | Err msg="must on $T$val not satisfied: $check $cmp $arg"


or  := _ ? True | __
not := _ ? False | True
and := _ ? (__ ? True | False) | False



||  := _ ? some Ok : some | No : __
||  := _ ? some Yay : some | Nay : __


_[_]	 :=	  EoL



// compose rtl
f2 <. f1 v := v f1 f2

// compose ltr
f1 .> f2 := _ f1 f2

// -- id , the well-known: func id(foo) {return foo}
self := _

// -- const
v only _ := v


list first , list must /= Empty :=
    list ? Link (_f & _) : f

list rest :=
    list    ? f Link r  : rest
            | Empty     : msg="rest: list must not be Empty" Err
    x foo   := (x trim len == 0) && "(none)" || x


x pow y :=
    y < 0   ? True  : 1 / (x pow y.neg)
            | False : x* accum 1 y , tmp := x * _   // -- * accumL 1 (y × x)


f accum initial n , n must >= 0 , x  :=
    True    ? n==0  : f accum x y , x := initial f , _ unused := 123
            |       : initial

    y := n - 1


a × b  /* huh1 */, /* huh2 */  a must >= 0   /* huh3 */ :=
    // -- a==0 && Empty || ret
    a == 0  ? True  : Empty
            | False
            | True  : b ret  // -- should catch such

    foo ret := foo Link ab
    ab := a-1 × b
    // yo


f accumR initial list :=
    list    ? Empty                 : initial
            | Link _first & _rest : first f (f accumR initial rest)


f accumL initial list :=
    list    ? Link              : initial
            | Link _first & _   : f accumL (initial f first) rest
