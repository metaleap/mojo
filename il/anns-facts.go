package atmoil

type IAnnFact interface {
}

// collection of facts that "all must be true based on analysis",
// if they contradict we'll find out later, in here they're just all collected.
// for the `astNodeBase.facts` usage of `AnnFactAll`, the first is "the one
// indisputable core truth" about the node and others are derived from context
type AnnFactAll struct {
	Core    IAnnFact
	Derived AnnFacts
}

func (me *AnnFactAll) Reset() { me.Core, me.Derived = nil, nil }

// this originates from `AstIdentName`s (due to their `Anns.ResolvesTo`)
// but may as a consequence also show up in the ancestors
type AnnFactAlts struct {
	Possibilities AnnFacts
}

// rune or string or uint64 or float64
type AnnFactLit struct {
	Value interface{}
}

type AnnFactTag struct {
	Value string
}

// invalid node (the corresponding errs are already on record)
type AnnFactBad struct{}

// reference: all facts for `To` are my facts
type AnnFactRef struct {
	To IAstNode
}

type AnnFactCall struct {
	Callee *AnnFactRef
	Arg    *AnnFactRef
}

type AnnFactCallable struct {
	Arg *AnnFactRef
	Ret *AnnFactRef
}

type AnnFacts []IAnnFact

func (me *AnnFacts) Add(facts ...IAnnFact) {
	*me = append(*me, facts...)
}
