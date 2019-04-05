

Bool                := False | True

// -- option type
Maybe               := No | Ok: _ | No:

// -- equivalent to Haskell's data Either = Left dis | Right dat
Or                  :=  Yay: _ | Nay: __

foo Bar             :=  x

any OrErr           :=  Ret: any
                    |   Err: msg: Text

t List              :=  Empty
                    |   Link: t & t List

t ListInfinite      := Link: t & t List

t ListNonEmpty      := t List, val must != Empty

t Vector n, n > 0   := t List, len must == n

t Tree              :=  Leaf
                    |   Node: & t Tree t & t Tree &

Txt                 := Text , trimmed , len must > 3

TxtBadIdea          := Txt, len must < 3 // -- let's see if we can be smart here later on

Name                := FirstLast: Txt & Txt

Address             := Addr:    street_HouseNo  : (Txt & Text, trimmed, len must > 0)
                            &   zip_City        : (Txt & Txt)   /*
                            &   foo             : bar
                            &   moo             : baz           */
                            &   country         : Txt

PhoneNo             := Txt

Customer            := Cust: Name & Address & PhoneNo

Person              :=  name: Name
                    &   bday: Date
                    &   addr: Address

User                :=  name: Txt
                    &   details: Person
                    &   numLogins: Int, val must >= 0
                    &   avatarPic: Byte List

Circle2D            :=  radius: Float, val must > 0
                    &   Float Vector 'x'...'y'

Sphere3D            :=  pos: Float Vector 3
                    &   extent: Float, val must > 0

/* -- freestanding comment */



check must cmp arg val :=
    val check cmp arg   ? val
                        | Err msg="must on $T$val not satisfied: $check $cmp $arg"


||  := _ ? True | __
not := _ ? False | True
&&  := _ ? (__ ? True | False) | False



||  := _ ? some Ok : some | __
||  := _ ? some Yay : some | __


_[_]	 :=	  Nil



// compose rtl
f2 <. f1 v := v f1 f2

// compose ltr
f1 .> f2 := _ f1 f2

// -- id , the well-known: func id(foo) {return foo}
val := _

// -- const
v only _ := v


list first , list must != Empty :=
    list ? f Link r : f

list rest :=
    list    ? f Link r  : rest
            | Empty     : msg="rest: list must not be Empty" Err
    x foo := (x trim len == 0) && "(none)" || x


x pow y :=
    y < 0   ? True  : 1 / (x pow y.neg)
            | False : x* accum 1 y , tmp := x * _   // -- * accumL 1 (y × x)


f accum initial n , n must >= 0 , x  :=
    True    ? n==0  : f accum x y , , x := initial f , , _ unused := 123
            |       : initial

    y := n - 1


a × b, a must >= 0 :=
    // -- a==0 && Empty || ret
    a == 0  ? True  : Empty
            | False
            | True  : b ret  // -- should catch such

    foo ret := foo Link ab
    ab := a-1 × b


f accumR initial list :=
    list    ? Empty           : initial
            | first Link rest : first f (f accumR initial rest)


f accumL initial list :=
    list    ? Empty           : initial
            | first Link rest : f accumL (initial /* foo */ f first) 123.456 /* c1*/ /* c2 */ // c3
