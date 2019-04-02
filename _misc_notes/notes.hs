// -- option type
any Maybe   := No | Ok: any

// -- equivalent to Haskell's Either
// -- second comment line just to have one
dis Or dat  :=  This: dis
            |   That: dat

any OrErr   :=  Ret any
            |   Err: msg: Text

t List      :=  Empty
            |   Link: t & t List

t MinList   := t List, val must != Empty

Txt         := Text, trim, len must > 3

Name        := FirstLast: Txt & Txt

Address     := Addr:    street_HouseNo  : (Txt & Text, trim, len must > 0)
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
    val check cmp arg   ? True  : val
                        | False : msg="must on $T$val not satisfied: $check $cmp $arg" Err
    // -- val check cmp arg && val
    // --                     || Err msg="must on $T$val not satisfied: $check $cmp $arg"



// -- id
any val := any

// -- const
any use drop := any

list rest :=
    list    ? f Link r  : rest
            | Empty     : msg="rest: list must not be Empty" Err


list first , list must != Empty :=
    list ? f Link r : f


x pow y :=
    y < 0   ? True    : 1 / (x pow y.neg)
            | False   : tmp accum 1 y , tmp := x * val // -- x*_   // -- * accumL 1 (y × x)


f accum initial n, n must >= 0 :=
    True    ? n==0  : f accum x y , x := initial f // , some unused := 123
            |    : initial
    y := n - 1


a × b, a must >= 0 :=
    // -- a==0 && Empty || ret
    a == 0  ? True  : Empty
            | False : b ret

    foo ret := foo Link ab
    ab := a-1 × b


f accumR initial list :=
    list    ? Empty           : initial
            | first Link rest : first f (f accumR initial rest)


f accumL initial list :=
    list    ? Empty           : initial
            | first Link rest : f accumL (initial /* foo */ f first) 123.456 /* c1*/ /* c2 */ // c3
