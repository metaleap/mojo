package atmoil

const annFactDescIndent = "  "

type IAnnFact interface {
	description(string) string
}

// collection of facts that "all must be true based on analysis",
// if they contradict we'll find out later, in here they're just all collected.
// for the `astNodeBase.facts` usage of `AnnFactAll`, the first is "the one
// indisputable core truth" about the node and others are derived from context
type AnnFactAll struct {
	Core    IAnnFact
	Derived AnnFacts
}

func (me *AnnFactAll) All() (all AnnFacts) {
	all = make(AnnFacts, 1, len(me.Derived)+1)
	all[0] = me.Core
	all = append(all, me.Derived...)
	return
}

func (me *AnnFactAll) Description() string { return me.description("") }

func (me *AnnFactAll) description(p string) (d string) {
	return me.Derived.describe("ALL OF:", p, me.Core)
}

func (me *AnnFactAll) Reset() { me.Core, me.Derived = nil, nil }

// this originates from `AstIdentName`s (due to their `Anns.ResolvesTo`)
// but may as a consequence also show up in the ancestors
type AnnFactAlts struct {
	Possibilities AnnFacts
}

func (me *AnnFactAlts) description(p string) (d string) {
	return me.Possibilities.describe("ONE OF:", p, nil)
}

// rune or string or uint64 or float64
type AnnFactLit struct {
	Value interface{}
	Str   func() string
}

func (me *AnnFactLit) description(p string) string { return p + "always equals: " + me.Str() }

type AnnFactTag struct {
	Value string
}

func (me *AnnFactTag) description(p string) string { return p + "always equals: " + me.Value }

// invalid node (the corresponding errs are already on record)
type AnnFactUndef struct{}

func (me *AnnFactUndef) description(p string) string { return p + "undefined / unreachable / invalid" }

// reference: all facts for `To` are my facts
type AnnFactRef struct {
	To IAstNode
}

func (me *AnnFactRef) description(p string) string {
	return p + "REF:\n" + me.To.Facts().description(p+annFactDescIndent)
}

type AnnFactCall struct {
	Callee *AnnFactRef
	Arg    *AnnFactRef
}

func (me *AnnFactCall) description(p string) (d string) {
	pref := p + annFactDescIndent
	d = p + "call:\n" + me.Callee.description(pref) + "\n" + p + "with:\n" + me.Arg.description(pref)
	return
}

type AnnFactCallable struct {
	Arg *AnnFactRef
	Ret *AnnFactRef
}

func (me *AnnFactCallable) description(p string) (d string) {
	pref := p + annFactDescIndent
	d = p + "callable: returns\n" + me.Ret.description(pref) + "\n" + p + "given arg:\n" + me.Arg.description(pref)
	return
}

type AnnFacts []IAnnFact

func (me *AnnFacts) Add(facts ...IAnnFact) {
	*me = append(*me, facts...)
}

func (me AnnFacts) description(p string) string {
	return me.describe("MANY:", p, nil)
}

func (me AnnFacts) describe(label string, p string, prior IAnnFact) (d string) {
	d = p + label
	pref := p + "->"
	if prior != nil {
		d += "\n" + prior.description(pref)
	}
	for i := range me {
		d += "\n" + me[i].description(pref)
	}
	return
}
