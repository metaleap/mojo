// -- option type
any Maybe   := No | Ok: any

// -- equivalent to Haskell's Either
// -- second comment line just to have one
dis Or dat  :=  This: dis
// -- not so nice but must be legal
            |   That: dat
// -- hrm

any OrErr   :=  Ret any
            |   Err: msg: Text

t List      :=  Empty
            |   Link: t & t List

t MinList   := t List, self must != Empty

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





check must cmp arg val :=
    val.check cmp arg   ? True  : val
                        | False : Err msg="must on $T$val not satisfied: $check $cmp $arg"
    // -- (val check) cmp arg && val
    // --                     || Err msg="must on $T$val not satisfied: $check $cmp $arg"





list rest :=
    list    ? Link first rest :   rest
            | Empty           :   Err msg="rest: list must not be Empty"


list first, list must != Empty :=
    list ? Link first rest : first


x pow y :=
    y < 0   ? True    : 1 / (x pow y.neg)
            | False   : tmp accum 1 y, v tmp := x * val // -- x*_   // -- * accumL 1 (y × x)


f accum initial n, n must >= 0 :=
    y := n - 1
    True    ? n==0  : f accum x y, x := initial f
            |       : initial


a × b c, a must >= 0 :=
    // -- a==0 && Empty || ret
    a == 0  ? True  : Empty
            | False : ret
    ret := Link b ab, ab := (a-1 × b)


f accumR initial list :=
    list    ? Empty           : initial
            | Link first rest : first f (f accumR initial rest)


f accumL initial list :=
    list    ? Empty           : initial
            | Link first rest : f accumL (initial f first) rest
