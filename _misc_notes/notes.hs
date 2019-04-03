

Bool := Nay | Yay

// -- option type
Maybe   := No | Yo: _

// -- equivalent to Haskell's data Either = Left dis | Right dat
Or  :=  Neat: _ | Weak: __

foo Bar :=  x , y

any OrErr   :=  Ret: any
            |   Err: msg: TEXT

t List      :=  Empty
            |   Link: t & t List

t MinList   := t List, val must != Empty

Txt         := TEXT , trim , len must > 3

TxtBadIdea  := Txt, len must < 3 // -- let's see if we can be smart here later on

Name        := FirstLast: Txt & Txt

Address     := Addr:    street_HouseNo  : (Txt & TEXT, trim, len must > 0)
                    &   zip_City        : (Txt & Txt)   /*
                    &   foo             : bar
                    &   moo             : baz           */
                    &   country         : Txt

PhoneNo     := Txt

Customer    := Cust: Name & Address & PhoneNo

Person      :=  name: Name
            &   bday: Date
            &   addr: Address

User        :=  name: Txt
            &   details: Person



/* -- freestanding comment */



check must cmp arg val :=
    val check cmp arg   && val
                        || Err msg="must on $T$val not satisfied: $check $cmp $arg"



check && ifSo || otherwise :=
    check   ? Yay   : ifSo
            |       : otherwise


maybe || other :=
    maybe   ? some Yo   : some
            |           : other


_[_]	 :=	  Nil



// compose rtl
f2 <. f1 v := v f1 f2

// compose ltr
f1 .> f2 := _ f1 f2

// -- id , the well-known: func id(foo) {return foo}
val := _

// -- const
v only _ := v

list rest :=
    list    ? f Link r  : rest
            | Empty     : msg="rest: list must not be Empty" Err
    x foo := (x trim len == 0) && "(none)" || x


list first , list must != Empty :=
    list ? f Link r : f


x pow y :=
    y < 0   ? Yay   : 1 / (x pow y.neg)
            | Nay   : x* accum 1 y , tmp := x * _   // -- * accumL 1 (y × x)


f accum initial n, n must >= 0 :=
    Yay ? n==0  : f accum x y , x := initial f , _ unused := 123
        |       : initial
    y := n - 1


a × b, a must >= 0 :=
    // -- a==0 && Empty || ret
    a == 0  ? Yay   : Empty
            | Nay   : b ret

    foo ret := foo Link ab
    ab := a-1 × b


f accumR initial Yo list :=
    list    ? Empty           : initial
            | first Link rest : first f (f accumR initial rest)


f accumL initial list :=
    list    ? Empty           : initial
            | first Link rest : f accumL (initial /* foo */ f first) 123.456 /* c1*/ /* c2 */ // c3
